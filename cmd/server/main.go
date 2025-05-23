package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var config Config

func main() {

	if err := loadConfig(&config); err != nil {
		log.Fatal().Err(err).Msg("error in initial configuration")
	}

	if err := validateConfig(&config); err != nil {
		log.Fatal().Err(err).Msg("invalid configuration")
	}

	initLogger(config.Logging.Level)

	fmt.Printf("%+v", config)
}

func initLogger(levelStr string) {
	levelStr = strings.ToLower(levelStr)

	level, err := zerolog.ParseLevel(levelStr)
	if err != nil {
		level = zerolog.InfoLevel
		log.Warn().Str("module", "logging").Msgf("invalid log level '%s' specified, falling back to 'info'", levelStr)
	}

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(level)

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

}
