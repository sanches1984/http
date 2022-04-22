# HTTP Server

Simple HTTP server with:
* Basic/Bearer authorization
* prometheus metrics
* logger
* tracer
* exposed profiler
* openapi.json

## Example

```go
package main

import (
	"context"
	"github.com/sanches1984/http/server"
	"io"
	"net/http"
)

func main() {
	srv := server.New("app", ":8080")
	srv.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		_, err := io.WriteString(w, "hello world")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.WriteHeader(http.StatusOK)
	})
	
	err := srv.Start(context.Background())
	if err != nil {
		panic(err)
	}
}
```
