package internal

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/util/runtime"
	"os"
	"time"
)

// ConfigureLogging using zerolog
func ConfigureLogging(verbose bool) {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.Stamp})
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
		Str("log-level", level.String()).
		Msg("logger configured")
}

// DisableLogging from zerolog
func DisableLogging() {
	zerolog.SetGlobalLevel(zerolog.Disabled)
}
