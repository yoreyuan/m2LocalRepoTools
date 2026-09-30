package cmd

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"yoreyuan/m2LocalRepoTools/pkg/config"
	"yoreyuan/m2LocalRepoTools/pkg/constant"
)

func extractCommand() string {
	args := os.Args[1:]
	for i := len(args) - 1; i >= 0; i-- {
		arg := args[i]
		if !strings.HasPrefix(arg, "-") {
			if arg == "push" || arg == "clean" {
				os.Args = append(os.Args[:i+1], os.Args[i+2:]...)
				return arg
			}
		}
	}
	return ""
}

func initArgs() error {
	command := extractCommand()

	configFile := flag.String("config", "", "Configuration file path")

	repo := flag.String("repo", "", "The path to the local repository maven will use to store artifacts")
	releaseUrl := flag.String("releaseUrl", "", "The URL of the nexus releases repository (releases only)")
	snapshotsUrl := flag.String("snapshotsUrl", "", "The URL of the nexus snapshots repository (snapshots only)")
	mixedUrl := flag.String("mixedUrl", "", "The URL of the nexus mixed repository (releases and snapshots)")
	username := flag.String("username", "", "Nexus username")
	password := flag.String("password", "", "Nexus password")
	worker := flag.Int("worker", 4, "Number of parallel upload workers")

	verbose := flag.Bool("verbose", false, "Enable app verbose mode")
	help := flag.Bool("help", false, "Display help information")

	flag.Parse()

	if *help {
		printUsage()
		os.Exit(0)
	}

	if command == "" {
		printUsage()
		return errors.New("subcommand is required: push or clean")
	}

	var conf *config.Config

	if *configFile != "" {
		err := initConfig()
		if err != nil {
			return err
		}
		conf = config.GetConfig()
	} else {
		conf = config.NewDefaultConfig()

		if *verbose {
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
			conf.Verbose = true
		}

		homeDir := getUserHomeDir()
		conf.LocalRepository = *repo
		if conf.LocalRepository == "" {
			conf.LocalRepository = fmt.Sprintf("%s/.m2/repository", homeDir)
		}
		log.Info().Msgf("local repository is %s", conf.LocalRepository)

		if *releaseUrl != "" {
			conf.Push.ReleaseUrl = *releaseUrl
		}
		if conf.Push.ReleaseUrl != "" {
			log.Info().Msgf("releases repository url is %s", conf.Push.ReleaseUrl)
		}

		if *snapshotsUrl != "" {
			conf.Push.SnapshotsUrl = *snapshotsUrl
		}
		if conf.Push.SnapshotsUrl != "" {
			log.Info().Msgf("snapshots repository url is %s", conf.Push.SnapshotsUrl)
		}

		if *mixedUrl != "" {
			conf.Push.MixedUrl = *mixedUrl
		}
		if conf.Push.MixedUrl != "" {
			log.Info().Msgf("mixed repository url is %s", conf.Push.MixedUrl)
		}

		if *username != "" {
			conf.Push.Username = *username
		}
		if *password != "" {
			conf.Push.Password = *password
		}
		if *worker > 0 {
			conf.Push.Worker = *worker
		}
		log.Info().Msgf("upload workers: %d", conf.Push.Worker)

		config.SetConfig(conf)
	}

	conf.Command = command

	if command == "push" {
		if conf.Push.ReleaseUrl == "" && conf.Push.SnapshotsUrl == "" && conf.Push.MixedUrl == "" {
			return errors.New("at least one of -releaseUrl, -snapshotsUrl or -mixedUrl must be configured")
		}
	}

	if strings.HasSuffix(conf.LocalRepository, constant.Separator) {
		conf.LocalRepository = strings.TrimRight(conf.LocalRepository, constant.Separator)
	}
	if strings.HasPrefix(conf.LocalRepository, "~") {
		newLocalRepository := strings.Replace(conf.LocalRepository, "~", getUserHomeDir(), 1)
		conf.LocalRepository = newLocalRepository
	}

	log.Info().Interface("config", conf).Msg("print config")

	return nil
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: m2LocalRepoTools [OPTIONS] <command>\n\n")
	fmt.Fprintf(os.Stderr, "Commands:\n")
	fmt.Fprintf(os.Stderr, "  push    Upload local Maven artifacts to Nexus repository\n")
	fmt.Fprintf(os.Stderr, "  clean   Clean files with specified suffixes in local repository\n\n")
	fmt.Fprintf(os.Stderr, "OPTIONS:\n")
	flag.PrintDefaults()
}

func getUserHomeDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Err(err).Msg("The home directory is obtained abnormally. The home directory will be set to the root directory")
		homeDir = "/root"
	}
	return homeDir
}
