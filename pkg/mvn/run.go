package mvn

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
	"strings"
)

func Run() {
	//err := filepath.Walk(constant.LocalRepository, visit)
	err := listFilesRecursively(conf.LocalRepository)
	if err != nil {
		log.Err(err).Msg("Exception when traversing files")
	}

	switch conf.Command {
	case "clean":
		err = doClean(*pomDirSet.GetSet())
		if err != nil {
			log.Err(err).Msg("Exception during clean operation")
		}
	case "push":
		err = executeMvn(*pomDirSet.GetSet())
		if err != nil {
			log.Err(err).Msg("Exception during push operation")
		}
	default:
		log.Error().Msgf("unknown command: %s", conf.Command)
	}
}

// visit is a function to be called for each file or directory found
func visit(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	// Check if this is a directory, if not, it's a file
	if !info.IsDir() {
		fmt.Println(path)
	}

	return nil
}

// listFilesRecursively
// Is a recursive function used to list all files in the specified directory
func listFilesRecursively(dir string) error {
	// Read the contents of the specified directory
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		// Get the full path of an entry
		fullPath := filepath.Join(dir, entry.Name())

		// If the entry is a directory, it calls itself recursively.
		if entry.IsDir() {
			if matchDirRegex(entry.Name()) {
				invalidDirSet.Add(fullPath)
				continue
			}
			err := listFilesRecursively(fullPath)
			if err != nil {
				return err
			}
		} else {
			if strings.HasSuffix(entry.Name(), ".pom") {
				pomDirSet.Add(dir)
			}
		}
	}

	return nil
}
