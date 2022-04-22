# HTTP Client

Simple HTTP client with:
* Basic/X-Token/Bearer authorization
* byte response
* logger
* tracer

## Example

```go
package main

import (
	"fmt"
	"context"
	"github.com/sanches1984/http/client"
)

func main() {
	c := client.New("app", "http://localhost:8080/")
	code, data, err := c.Get(context.Background(), "/hello")
	if err != nil {
		panic(err)
	}

	fmt.Println(code, string(data))
}
```
