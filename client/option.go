package client

import (
	"github.com/rs/zerolog"
	"net"
	"net/http"
	"time"
)

type Option func(c *client)

func WithTimeout(timeout time.Duration) Option {
	return func(c *client) {
		c.http.Timeout = timeout
		c.http.Transport = &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   timeout,
				KeepAlive: timeout,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}
	}
}

func WithLogger(logger zerolog.Logger) Option {
	return func(c *client) {
		c.logger = logger
	}
}

func WithTracer() Option {
	return func(c *client) {
		tracer, closer := initTracer(c.appName)
		c.tracer = tracer
		c.closers["tracer"] = closer.Close
	}
}

func WithBasicAuth(login, password string) Option {
	return func(c *client) {
		c.middlewares = append(c.middlewares, newBasicAuthMiddleware(login, password))
	}
}

func WithBearerTokenAuth(token string) Option {
	return func(c *client) {
		c.middlewares = append(c.middlewares, newBearerTokenAuthMiddleware(token))
	}
}

func WithXTokenAuth(ctxTokenKey interface{}) Option {
	return func(c *client) {
		c.middlewares = append(c.middlewares, newXTokenAuthMiddleware(ctxTokenKey))
	}
}

func WithMiddleware(mw ...Middleware) Option {
	return func(c *client) {
		c.middlewares = append(c.middlewares, mw...)
	}
}
