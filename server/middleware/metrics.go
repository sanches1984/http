package middleware

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"strings"
)

var (
	reg           = prometheus.NewPedanticRegistry()
	collectorList = []prometheus.Collector{}

	CountError   *prometheus.CounterVec
	CountRequest *prometheus.CounterVec
	ResponseTime *prometheus.HistogramVec
)

func NewMetricsMiddleware() func(next http.Handler) http.Handler {
	// todo
}

func AddBasicCollector(prefix string) {
	prefix = strings.ReplaceAll(prefix, "-", "_")

	CountError = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: prefix + "_error_count",
		Help: "The total number of request errors",
	}, []string{"handler", "method", "code"})

	CountRequest = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: prefix + "_request_count",
		Help: "The total request count",
	}, []string{"handler", "method", "code"})

	ResponseTime = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    prefix + "_response_time",
		Help:    "Response time in seconds",
		Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
	}, []string{"handler", "method", "code"})

	collectorList = append(collectorList,
		prometheus.NewGoCollector(),
		CountError,
		CountRequest,
		ResponseTime,
	)
}

func AddCollector(collector ...prometheus.Collector) {
	collectorList = append(collectorList, collector...)
}

// Metrics prometheus metrics
func Metrics() http.Handler {
	reg.MustRegister(collectorList...)
	return promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
}
