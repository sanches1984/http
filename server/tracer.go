package server

import (
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"io"
)

func initTracer(serviceName string) (opentracing.Tracer, io.Closer) {
	appTracer, closer := jaeger.NewTracer(
		serviceName,
		jaeger.NewConstSampler(true),
		jaeger.NewLoggingReporter(jaeger.StdLogger),
	)
	opentracing.InitGlobalTracer(appTracer)
	return appTracer, closer
}
