package server

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"strings"
)

var (
	reg           = prometheus.NewPedanticRegistry()
	collectorList = []prometheus.Collector{}

	CountRequest  *prometheus.CounterVec
	CountResponse *prometheus.CounterVec
	ResponseTime  *prometheus.HistogramVec
)

func handlerMetrics() http.Handler {
	reg.MustRegister(collectorList...)
	return promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
}

func addBasicCollector(prefix string) {
	prefix = strings.ReplaceAll(prefix, "-", "_")

	CountResponse = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: prefix + "_response_count",
		Help: "The total number of responses",
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
		CountResponse,
		CountRequest,
		ResponseTime,
	)
}
