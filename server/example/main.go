package main

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/sanches1984/http/server"
	"github.com/sanches1984/http/server/middleware"
	"net/http"
)

func main() {
	srv := server.NewServer("test", ":8089",
		server.WithMiddleware(
			middleware.NewVersionMiddleware("0.0.2"),
			middleware.NewCorsMiddleware(),
		),
	)
	srv.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		// do something
		w.WriteHeader(http.StatusOK)
	})

	err := srv.Start(context.Background())
	if err != nil {
		zerolog.DefaultContextLogger.Fatal().Err(err).Msg("app fatal")
	}
}
