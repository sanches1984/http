package server

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/urfave/negroni"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	reg           = prometheus.NewPedanticRegistry()
	collectorList = []prometheus.Collector{}

	CountError   *prometheus.CounterVec
	CountRequest *prometheus.CounterVec
	ResponseTime *prometheus.HistogramVec
)

func newMetricsMiddleware(prefix string) func(next http.Handler) http.Handler {
	addBasicCollector(prefix)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			lrw := negroni.NewResponseWriter(w)

			CountRequest.WithLabelValues(r.RequestURI, r.Method).Inc()

			next.ServeHTTP(lrw, r)

			if lrw.Status() >= 300 {
				CountError.WithLabelValues(r.RequestURI, r.Method, strconv.Itoa(lrw.Status())).Inc()
			}

			ResponseTime.WithLabelValues(r.RequestURI, r.Method).Observe(time.Since(start).Seconds())
		})
	}
}

// Metrics prometheus metrics
func Metrics() http.Handler {
	reg.MustRegister(collectorList...)
	return promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
}

func addBasicCollector(prefix string) {
	prefix = strings.ReplaceAll(prefix, "-", "_")

	CountError = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: prefix + "_error_count",
		Help: "The total number of request errors",
	}, []string{"handler", "method", "code"})

	CountRequest = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: prefix + "_request_count",
		Help: "The total request count",
	}, []string{"handler", "method"})

	ResponseTime = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    prefix + "_response_time",
		Help:    "Response time in seconds",
		Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
	}, []string{"handler", "method"})

	collectorList = append(collectorList,
		prometheus.NewGoCollector(),
		CountError,
		CountRequest,
		ResponseTime,
	)
}
