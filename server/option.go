package server

import (
	"github.com/rs/zerolog"
	"time"
)

type Option func(s *server)

func WithMiddleware(mw ...Middleware) Option {
	return func(s *server) {
		s.middlewares = append(s.middlewares, mw...)
	}
}

func WithGracefulShutdown(delay, timeout time.Duration) Option {
	return func(s *server) {
		s.gracefulDelay = delay
		s.gracefulTimeout = timeout
	}
}

func WithHTTPTimeout(timeout time.Duration) Option {
	return func(s *server) {
		s.srv.ReadTimeout = timeout
		s.srv.WriteTimeout = timeout
		s.srv.ReadHeaderTimeout = timeout
	}
}

func WithSwaggerInfo() Option {
	return func(s *server) {
		s.showSwagger = true
	}
}

func WithLogger(logger zerolog.Logger) Option {
	return func(s *server) {
		s.middlewares = append(s.middlewares, newLogMiddleware(logger))
		s.logger = logger
		s.isLoggerSet = true
	}
}
