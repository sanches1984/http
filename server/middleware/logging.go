package middleware

import (
	"github.com/rs/zerolog"
	"github.com/sanches1984/gopkg-logger"
	"net/http"
	"time"
)

var (
	exceptionsList = []string{"/metrics", "/health"}
	uriExceptions  map[string]struct{}
)

func init() {
	uriExceptions = make(map[string]struct{}, len(exceptionsList))
	for _, uri := range exceptionsList {
		uriExceptions[uri] = struct{}{}
	}
}

func NewLogMiddleware() func(next http.Handler) http.Handler {

	logger := log.WithContext(r.Context(), zerolog.Logger{})
	logger.Debug().
		Str("source", "http-server").
		Str("address", addr).
		Str("method", r.Method).
		Str("uri", r.RequestURI).
		Str("proto", r.Proto).
		Int("status", wp.statusCode).
		Int64("bytes", wp.written).
		Float64("latency_ms", float64(time.Since(start).Nanoseconds())/(1000*1000)).Msg("request")
}
