package server

import (
	"context"
	mw "github.com/go-chi/chi/middleware"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/rs/zerolog"
	log "github.com/sanches1984/gopkg-logger"
	"github.com/urfave/negroni"
	"net/http"
	"strconv"
	"time"
)

type Middleware func(next http.Handler) http.Handler

func newRequestIdMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if key, ok := r.Context().Value(mw.RequestIDKey).(string); ok {
				w.Header().Set(requestIDHeaderName, key)
			}
			next.ServeHTTP(w, r)
		}
		return mw.RequestID(http.HandlerFunc(fn))
	}
}

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

func newMetricsMiddleware(prefix string) func(next http.Handler) http.Handler {
	addBasicCollector(prefix)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			lrw := negroni.NewResponseWriter(w)

			CountRequest.WithLabelValues(r.RequestURI, r.Method).Inc()

			next.ServeHTTP(lrw, r)

			CountResponse.WithLabelValues(r.RequestURI, r.Method, strconv.Itoa(lrw.Status())).Inc()
			ResponseTime.WithLabelValues(r.RequestURI, r.Method).Observe(time.Since(start).Seconds())
		})
	}
}

func newTracingMiddleware(tracer opentracing.Tracer) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
			span := tracer.StartSpan(r.RequestURI, ext.RPCServerOption(spanCtx))
			defer span.Finish()
			ctx := opentracing.ContextWithSpan(context.Background(), span)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}
