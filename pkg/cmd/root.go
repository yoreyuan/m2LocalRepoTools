package cmd

import (
	"github.com/rs/zerolog/log"
	"yoreyuan/m2LocalRepoTools/pkg/mvn"
)

func Execute() {
	err := initArgs()
	if err != nil {
		log.Error().Msg(err.Error())
		return
	}

	mvn.Init()
	mvn.Run()
}
