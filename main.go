package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"yoreyuan/m2LocalRepoTools/pkg/cmd"
)

func init() {
	// Optimize print output
	//log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	log.Logger = log.With().Caller().Logger()
	log.Logger = log.Output(
		zerolog.ConsoleWriter{
			NoColor:    true,
			Out:        os.Stdout,
			TimeFormat: "2006/01/02 15:04:05",
			FormatLevel: func(i interface{}) string {
				return strings.ToUpper(fmt.Sprintf("| %-6s|", i))
			}})
}

func main() {
	cmd.Execute()
}
