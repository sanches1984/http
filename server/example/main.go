package main

import (
	"context"
	logger "github.com/sanches1984/gopkg-logger"
	"github.com/sanches1984/http/server"
	"net/http"
	"time"
)

func main() {
	srv := server.New("test", ":8089",
		server.WithLogger(logger.For("test")),
		server.WithTracer())
	srv.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		// do something
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	})

	err := srv.Start(context.Background())
	if err != nil {
		panic(err)
	}
}
