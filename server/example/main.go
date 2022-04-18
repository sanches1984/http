package main

import (
	"context"
	"github.com/sanches1984/http/server"
	"io"
	"log"
	"net/http"
)

func main() {
	srv := server.NewServer()
	srv.HandleFunc("/hello", func(writer http.ResponseWriter, request *http.Request) {
		_, err := io.WriteString(writer, "hello world")
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
		}
	})

	ctx := context.Background()
	err := srv.Start(ctx)
	if err != nil {
		log.Fatal(err)
	}
}
