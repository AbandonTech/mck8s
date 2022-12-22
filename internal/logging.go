package internal

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/util/runtime"
	"os"
)

// ConfigureLogging using zerolog
func ConfigureLogging() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	runtime.ErrorHandlers = []func(error){
		func(err error) {
			log.Warn().
				Err(err).
				Msg("[k8s]")
		},
	}

	// Default to InfoLevel, unless level is provided in environment variables
	level := zerolog.InfoLevel
	envLevel := os.Getenv("LOG_LEVEL")
	if envLevel != "" {
		var err error
		level, err = zerolog.ParseLevel(envLevel)
		if err != nil {
			log.Fatal().
				Str("LOG_LEVEL", envLevel).
				Msg("Unable to parse log level string provided")
		}
	}

	zerolog.SetGlobalLevel(level)

	log.Debug().
		Str("LogLevel", level.String()).
		Msg("Logger configured")
}
