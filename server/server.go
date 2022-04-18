package server

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/sanches1984/gopkg-logger"
	"github.com/sanches1984/http/server/middleware"
	"github.com/urfave/negroni"
	"net/http"
	"net/http/pprof"
	"strconv"
	"time"
)

type Server struct {
	Router                 *mux.Router
	MiddlewaresHttpHandler []middleware.HttpMiddleware
	Middlewares            *negroni.Negroni
	logger                 zerolog.Logger
	srv                    *http.Server
}

func init() {
	logger := log.For("http-server")
	if err := godotenv.Load(); err != nil {
		logger.Error().Err(err).Msg("missing envs")
	}
	if err := envconfig.Process("", &specs); err != nil {
		logger.Error().Err(err).Msg("can't read envs")
	}
}

func NewServer() *Server {
	router := mux.NewRouter()

	mw := negroni.New()
	mw.Use(negroni.HandlerFunc(middleware.RequestMetadataMiddleware))
	mw.Use(negroni.HandlerFunc(middleware.LogMiddleware))

	srv := &http.Server{
		Addr: specs.Host + ":" + strconv.Itoa(specs.Port),
	}

	httpHandlerMiddlewares := make([]middleware.HttpMiddleware, 0, 2)
	prometheusMiddleware := middleware.NewDefault()
	httpHandlerMiddlewares = append(httpHandlerMiddlewares, prometheusMiddleware.Handler)

	return &Server{Router: router,
		Middlewares:            mw,
		MiddlewaresHttpHandler: httpHandlerMiddlewares,
		logger:                 log.For("http-server"),
		srv:                    srv,
	}
}

func middlewareOptions() (options []nethttp.MWOption) {
	opt := nethttp.OperationNameFunc(func(r *http.Request) string {
		return r.Method + " " + r.URL.Path
	})
	options = append(options, opt)
	return options
}

func (s *Server) HandleFunc(path string, handleFunc func(http.ResponseWriter, *http.Request)) *mux.Route {
	handler := http.HandlerFunc(handleFunc)

	return s.Handle(path, handler)
}
func (s *Server) Handle(path string, handler http.Handler) *mux.Route {
	wrappedHandler := handler

	for _, mw := range s.MiddlewaresHttpHandler {
		wrappedHandler = mw(path, wrappedHandler)
	}

	return s.Router.HandleFunc(path, wrappedHandler.ServeHTTP)
}

func (s *Server) Server() *http.Server {
	return s.srv
}

func (s *Server) SetAddr(host string, port int) {
	s.srv.Addr = host + ":" + strconv.Itoa(port)
}

func (s *Server) Start(ctx context.Context) error {
	s.Router.HandleFunc("/health", s.handlerHealth)
	s.Router.HandleFunc("/metrics", promhttp.Handler().ServeHTTP)
	s.Router.HandleFunc("/openapi.json", s.handlerOpenAPI)

	// pprof
	s.Router.HandleFunc("/debug/pprof/", pprof.Index)
	s.Router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	s.Router.HandleFunc("/debug/pprof/profile", pprof.Profile)
	s.Router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	s.Router.HandleFunc("/debug/pprof/trace", pprof.Trace)
	s.Router.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	s.Router.Handle("/debug/pprof/heap", pprof.Handler("heap"))
	s.Router.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
	s.Router.Handle("/debug/pprof/block", pprof.Handler("block"))

	http.Handle("/", s.Router)

	s.Middlewares.UseHandler(nethttp.Middleware(opentracing.GlobalTracer(), sentry.Recoverer(http.DefaultServeMux), middlewareOptions()...))

	s.srv.Handler = s.Middlewares

	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), specs.GracefulTimeout)
		defer cancel()

		time.Sleep(specs.GracefulDelay)
		s.srv.SetKeepAlivesEnabled(false)

		s.logger.Info().Msg("http-server is shutting down...")
		if err := s.srv.Shutdown(ctx); err != nil {
			s.logger.Error().Err(err).Msg("The http-server stopped with error")
		} else {
			s.logger.Info().Msg("http-server was successfully stopped")
		}
	}()
	s.printRoutes()

	s.logger.Info().Str("addr", s.srv.Addr).Msgf("start http-server at %s", s.srv.Addr)
	return s.srv.ListenAndServe()
}

func (s *Server) printRoutes() {
	err := s.Router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		t, err := route.GetPathTemplate()
		if err != nil {
			return err
		}
		s.logger.Info().Str("path", t).Msg("")
		return nil
	})
	if err != nil {
		s.logger.Error().Err(err).Msg("printRoutes")
	}
}
