package mvn

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/rs/zerolog/log"

	"yoreyuan/m2LocalRepoTools/pkg/config"
	"yoreyuan/m2LocalRepoTools/pkg/constant"
	"yoreyuan/m2LocalRepoTools/pkg/utils"
)

type MvnModel struct {
	File       string
	PomFile    string
	Packaging  string
	Classifier string
}

type PomProject struct {
	GroupId    string `xml:"groupId"`
	ArtifactId string `xml:"artifactId"`
	Version    string `xml:"version"`
	Parent     *struct {
		GroupId string `xml:"groupId"`
		Version string `xml:"version"`
	} `xml:"parent"`
}

var (
	pomDirSet     = utils.NewSet()
	invalidDirSet = utils.NewSet()
	conf          *config.Config
	httpClient    *http.Client
	successCnt    int64
	skipCnt       int64
	failCnt       int64

	errAssetExists = errors.New("asset already exists")
)

func Init() {
	conf = config.GetConfig()
	httpClient = &http.Client{}
}

func parsePom(pomPath string) (*PomProject, error) {
	data, err := os.ReadFile(pomPath)
	if err != nil {
		return nil, err
	}
	var pom PomProject
	if err := xml.Unmarshal(data, &pom); err != nil {
		return nil, err
	}
	if pom.GroupId == "" && pom.Parent != nil {
		pom.GroupId = pom.Parent.GroupId
	}
	if pom.Version == "" && pom.Parent != nil {
		pom.Version = pom.Parent.Version
	}
	return &pom, nil
}

func buildUploadUrl(repoUrl, groupId, artifactId, version, fileName string) string {
	groupPath := strings.ReplaceAll(groupId, ".", "/")
	base := strings.TrimRight(repoUrl, "/")
	return fmt.Sprintf("%s/%s/%s/%s/%s", base, groupPath, artifactId, version, fileName)
}

func uploadFile(filePath, uploadUrl string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open file failed: %w", err)
	}
	defer f.Close()

	req, err := http.NewRequest(http.MethodPut, uploadUrl, f)
	if err != nil {
		return fmt.Errorf("create request failed: %w", err)
	}
	req.SetBasicAuth(conf.Push.Username, conf.Push.Password)
	req.Header.Set("Content-Type", "application/octet-stream")

	fi, err := f.Stat()
	if err == nil {
		req.ContentLength = fi.Size()
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("upload request failed: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode == http.StatusConflict {
		return fmt.Errorf("%w: %s", errAssetExists, resp.Status)
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("upload failed, status: %s", resp.Status)
	}
	return nil
}

func resolveRepoUrl(pom *PomProject) string {
	if strings.HasSuffix(pom.Version, "-SNAPSHOT") {
		if conf.Push.SnapshotsUrl != "" {
			return conf.Push.SnapshotsUrl
		}
		return conf.Push.MixedUrl
	}
	if conf.Push.ReleaseUrl != "" {
		return conf.Push.ReleaseUrl
	}
	return conf.Push.MixedUrl
}

func deployArtifact(model *MvnModel, pom *PomProject) error {
	fileName := filepath.Base(model.File)
	repoUrl := resolveRepoUrl(pom)
	if repoUrl == "" {
		log.Warn().Msgf("skipping %s: version %s but no suitable repository URL configured", model.File, pom.Version)
		atomic.AddInt64(&failCnt, 1)
		return nil
	}
	uploadUrl := buildUploadUrl(repoUrl, pom.GroupId, pom.ArtifactId, pom.Version, fileName)

	if conf.Verbose {
		log.Info().Msgf("uploading %s -> %s", model.File, uploadUrl)
	}

	return uploadFile(model.File, uploadUrl)
}

func handleDeployResult(err error, file string, failedCmds *[]string, mu *sync.Mutex) {
	if err == nil {
		atomic.AddInt64(&successCnt, 1)
		return
	}
	if errors.Is(err, errAssetExists) {
		log.Info().Msgf("skipped (already exists): %s", file)
		atomic.AddInt64(&skipCnt, 1)
		return
	}
	log.Err(err).Msgf("deploy failed: %s", file)
	mu.Lock()
	*failedCmds = append(*failedCmds, file)
	mu.Unlock()
	atomic.AddInt64(&failCnt, 1)
}

func executeMvn(pomDir map[string]bool) error {
	c := len(pomDir)
	var index int32
	var wg sync.WaitGroup
	sem := make(chan struct{}, conf.Push.Worker)

	var mu sync.Mutex
	var failedCmds []string

	for item := range pomDir {
		if conf.Verbose {
			log.Info().Msgf("Start processing packages in folder %s", item)
		}
		entries, err := os.ReadDir(item)
		if err != nil {
			log.Error().Err(err).Msgf("read dir failed: %s", item)
			continue
		}

		pomFile := ""
		itemSplit := strings.Split(item, constant.Separator)
		if len(itemSplit) <= 1 {
			log.Error().Msg("the maven repository path format is incorrect")
			continue
		}
		parentDirName := itemSplit[len(itemSplit)-1]

		for _, entry := range entries {
			if strings.HasSuffix(entry.Name(), ".pom") {
				pomFile = filepath.Join(item, entry.Name())
				break
			}
		}
		if pomFile == "" {
			log.Error().Msgf("maven pom file not found in folder: %s", item)
			continue
		}

		pom, err := parsePom(pomFile)
		if err != nil {
			log.Error().Err(err).Msgf("parse pom failed: %s", pomFile)
			continue
		}

		var itemSet = utils.NewSet()
		for _, entry := range entries {
			fileName := entry.Name()
			//if strings.HasSuffix(fileName, ".pom") {
			//	continue
			//}
			//if strings.Contains(fileName, ".pom.") {
			//	continue
			//}

			skip := false
			for _, s := range conf.Push.ExcludeSuffixs {
				if strings.HasSuffix(fileName, s) {
					skip = true
					break
				}
			}
			if skip {
				continue
			}

			if !strings.Contains(fileName, "-"+parentDirName) {
				continue
			}

			itemSet.Add(fileName)
		}

		wg.Add(1)
		sem <- struct{}{}
		go func(pomFile string, pom *PomProject) {
			defer wg.Done()
			defer func() { <-sem }()

			model := &MvnModel{PomFile: pomFile, File: pomFile, Packaging: "pom"}
			handleDeployResult(deployArtifact(model, pom), pomFile, &failedCmds, &mu)
		}(pomFile, pom)

		for file := range *itemSet.GetSet() {
			arr1 := strings.Split(file, parentDirName+".")
			if len(arr1) < 1 {
				log.Error().Msgf("incorrect file format: %s", file)
				continue
			}

			model := &MvnModel{PomFile: pomFile, File: filepath.Join(item, file)}
			if len(arr1) == 1 {
				arr2 := strings.Split(file, parentDirName+"-")
				s2 := arr2[len(arr2)-1]
				arr3 := strings.Split(s2, ".")
				if len(arr3) < 2 {
					log.Error().Msgf("incorrect file format: %s", file)
					continue
				}
				model.Packaging = arr3[len(arr3)-1]
				model.Classifier = strings.TrimRight(s2, "."+model.Packaging)
			} else {
				model.Packaging = arr1[len(arr1)-1]
			}

			wg.Add(1)
			sem <- struct{}{}
			go func(model *MvnModel, pom *PomProject) {
				defer wg.Done()
				defer func() { <-sem }()

				handleDeployResult(deployArtifact(model, pom), model.File, &failedCmds, &mu)
			}(model, pom)
		}

		idx := int(atomic.AddInt32(&index, 1))
		log.Info().Msgf("progress is %.2f%%", float32(idx)/float32(c)*100)
	}

	wg.Wait()

	if len(failedCmds) > 0 {
		log.Warn().Msgf("An exception occurred when publishing %d artifacts. Failed files:", len(failedCmds))
		for _, s := range failedCmds {
			log.Warn().Msgf("  %s", s)
		}
	}

	log.Info().Msgf("upload completed: success=%d, skipped=%d, failed=%d, total=%d", successCnt, skipCnt, failCnt, successCnt+skipCnt+failCnt)

	return nil
}
