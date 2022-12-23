package middleware

import (
	"github.com/rs/zerolog/log"
	"net/http"
)

func LoggingMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debug().
			Str("Method", r.Method).
			Str("Remote", r.RemoteAddr).
			Str("Path", r.URL.Path).
			Msg("Handling request")

		h.ServeHTTP(w, r)
	})
}
