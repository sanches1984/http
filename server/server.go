package server

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/opentracing/opentracing-go"
	"github.com/rs/zerolog"
	"net/http"
	"net/http/pprof"
	"time"
)

const (
	defaultGracefulDelay   = 3 * time.Second
	defaultGracefulTimeout = 5 * time.Second
	defaultHTTPTimeout     = 30 * time.Second
)

type HTTPServer interface {
	HandleFunc(path string, handleFunc func(http.ResponseWriter, *http.Request)) *mux.Route
	Handle(path string, handler http.Handler) *mux.Route
	Start(ctx context.Context) error
}

type server struct {
	srv         *http.Server
	router      *mux.Router
	middlewares []Middleware
	closers     map[string]func() error

	appName         string
	gracefulDelay   time.Duration
	gracefulTimeout time.Duration
	logger          zerolog.Logger
	tracer          opentracing.Tracer
}

func New(appName, addr string, options ...Option) HTTPServer {
	srv := &server{
		appName: appName,
		router:  mux.NewRouter(),
		srv: &http.Server{
			Addr:              addr,
			ReadTimeout:       defaultHTTPTimeout,
			WriteTimeout:      defaultHTTPTimeout,
			ReadHeaderTimeout: defaultHTTPTimeout,
		},
		middlewares:     make([]Middleware, 0, 2),
		closers:         make(map[string]func() error, 1),
		gracefulDelay:   defaultGracefulDelay,
		gracefulTimeout: defaultGracefulTimeout,
	}
	for _, o := range options {
		o(srv)
	}

	if srv.tracer != nil {
		srv.middlewares = append(srv.middlewares, newTracingMiddleware(srv.tracer))
	}
	srv.middlewares = append(srv.middlewares, newMetricsMiddleware(appName))
	srv.middlewares = append(srv.middlewares, newRequestIdMiddleware())
	return srv
}

func (s *server) HandleFunc(path string, handleFunc func(http.ResponseWriter, *http.Request)) *mux.Route {
	handler := http.HandlerFunc(handleFunc)
	return s.Handle(path, handler)
}

func (s *server) Handle(path string, handler http.Handler) *mux.Route {
	wrappedHandler := handler
	for _, mw := range s.middlewares {
		wrappedHandler = mw(wrappedHandler)
	}

	return s.router.HandleFunc(path, wrappedHandler.ServeHTTP)
}

func (s *server) Start(ctx context.Context) error {
	defer s.close()

	// health & metrics
	s.router.HandleFunc("/openapi.json", s.handlerOpenAPI)
	s.router.HandleFunc("/health", s.handlerHealth)
	s.router.Handle("/metrics", handlerMetrics())

	// pprof
	s.router.HandleFunc("/debug/pprof/", pprof.Index)
	s.router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	s.router.HandleFunc("/debug/pprof/profile", pprof.Profile)
	s.router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	s.router.HandleFunc("/debug/pprof/trace", pprof.Trace)
	s.router.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	s.router.Handle("/debug/pprof/heap", pprof.Handler("heap"))
	s.router.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
	s.router.Handle("/debug/pprof/block", pprof.Handler("block"))

	http.Handle("/", s.router)

	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), s.gracefulTimeout)
		defer cancel()

		time.Sleep(s.gracefulDelay)
		s.srv.SetKeepAlivesEnabled(false)

		s.logger.Info().Msg("http-server is shutting down...")
		if err := s.srv.Shutdown(ctx); err != nil {
			s.logger.Error().Err(err).Msg("The http-server stopped with error")
		} else {
			s.logger.Info().Msg("http-server was successfully stopped")
		}
	}()
	s.printRoutes()

	s.logger.Info().Str("addr", s.srv.Addr).Msg("start http-server")
	return s.srv.ListenAndServe()
}

func (s *server) close() {
	for name, closer := range s.closers {
		if err := closer(); err != nil {
			s.logger.Warn().Err(err).Msgf("can't close %s", name)
		}
	}
}

func (s *server) printRoutes() {
	err := s.router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		t, err := route.GetPathTemplate()
		if err != nil {
			return err
		}
		s.logger.Debug().Str("path", t).Msg("")
		return nil
	})
	if err != nil {
		s.logger.Error().Err(err).Msg("printRoutes")
	}
}
