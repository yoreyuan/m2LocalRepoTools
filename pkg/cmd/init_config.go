package cmd

import (
	"yoreyuan/m2LocalRepoTools/pkg/config"

	"github.com/rs/zerolog/log"
)

func initConfig() error {
	configFile, err := config.Parse()
	if err != nil {
		log.Error().Err(err).Msg("get config file failed")
		return err
	}
	log.Info().Msgf("config file is %s", configFile)

	_, err = config.Load(configFile)
	if err != nil {
		log.Error().Err(err).Msg("load config form file failed")
		return err
	}

	return nil
}
