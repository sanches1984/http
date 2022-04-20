package server

import (
	"github.com/rs/zerolog"
	log "github.com/sanches1984/gopkg-logger"
	"github.com/urfave/negroni"
	"net/http"
	"time"
)

func newLogMiddleware(logger zerolog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			lrw := negroni.NewResponseWriter(w)

			next.ServeHTTP(lrw, r)

			log.WithContext(r.Context(), logger).Debug().
				Str("method", r.Method).
				Str("uri", r.RequestURI).
				Str("proto", r.Proto).
				Int("status", lrw.Status()).
				Float64("latency_ms", float64(time.Since(start).Nanoseconds())/(1000*1000)).
				Msg("request")
		})
	}
}
