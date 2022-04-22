package main

import (
	"context"
	"fmt"
	log "github.com/sanches1984/gopkg-logger"
	"github.com/sanches1984/http/client"
)

func main() {
	c := client.New("test", "http://localhost:8089/", client.WithLogger(log.For("client")), client.WithTracer())
	code, _, err := c.Get(context.Background(), "/hello")
	if err != nil {
		panic(err)
	}
	fmt.Println(code)
}
