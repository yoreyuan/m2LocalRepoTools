package config

import (
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Config struct {
	Verbose         bool   `mapstructure:"verbose,default=false"`
	LogLevel        string `yaml:"logLevel"`
	LocalRepository string `mapstructure:"localRepository,default=~/.m2/repository"`
	Push            `mapstructure:"push"`
	Clean           RepoClean `mapstructure:"clean"`
	Command         string    `mapstructure:"-"`
}

type Push struct {
	ReleaseUrl     string `mapstructure:"releaseUrl"`
	SnapshotsUrl   string `mapstructure:"snapshotsUrl"`
	MixedUrl       string `mapstructure:"mixedUrl"`
	Username       string `mapstructure:"username"`
	Password       string `mapstructure:"password"`
	Worker         int    `mapstructure:"worker,default=4"`
	ExcludeSuffixs `mapstructure:"excludeSuffixs"`
}

type ExcludeSuffixs []string

type RepoClean struct {
	CleanSuffixs `mapstructure:"suffixs"`
	DirRegex     []string `mapstructure:"dirRegex"`
}

type CleanSuffixs []string

func NewDefaultConfig() *Config {
	return &Config{
		Verbose:         false,
		LogLevel:        "INFO",
		LocalRepository: "",
		Push: Push{
			ReleaseUrl:     "",
			SnapshotsUrl:   "",
			MixedUrl:       "",
			Username:       "",
			Password:       "",
			Worker:         4,
			ExcludeSuffixs: ExcludeSuffixs{},
		},
		Clean: RepoClean{
			CleanSuffixs: CleanSuffixs{},
		},
	}
}

var globalConf = &Config{}

func GetConfig() *Config {
	return globalConf
}

func SetConfig(conf *Config) {
	globalConf = conf
}

func Parse() (string, error) {
	defaultConfigFile, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	var configFileName string
	pflag.StringVar(&configFileName,
		"config",
		filepath.Join(defaultConfigFile, "config.yaml"),
		"explicitly defines the path, name and extension of the config file")
	pflag.Parse()
	return configFileName, err
}

func Load(fileName string) (*Config, error) {
	viper.SetConfigFile(fileName)
	err := viper.ReadInConfig()
	if err != nil {
		log.Warn().Err(err).Msg("viper read in config failed, use env or default")
		return nil, err
	}

	conf := NewDefaultConfig()
	err = viper.Unmarshal(&conf)
	if err != nil {
		return nil, err
	}

	globalConf = conf
	logLevel, err := zerolog.ParseLevel(conf.LogLevel)
	if err != nil {
		return nil, err
	}
	zerolog.SetGlobalLevel(logLevel)

	return conf, nil
}
