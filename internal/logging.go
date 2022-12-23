package internal

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/util/runtime"
	"os"
)

// ConfigureLogging using zerolog
func ConfigureLogging(verbose bool) {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	runtime.ErrorHandlers = []func(error){
		func(err error) {
			log.Warn().
				Err(err).
				Msg("[k8s]")
		},
	}

	var level zerolog.Level
	if verbose {
		level = zerolog.DebugLevel
	} else {
		level = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(level)

	log.Debug().
		Str("LogLevel", level.String()).
		Msg("Logger configured")
}
