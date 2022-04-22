package client

import (
	"context"
	"fmt"
	"net/http"
)

type Middleware func(ctx context.Context, request *http.Request)

func newXTokenAuthMiddleware(tokenKey interface{}) Middleware {
	return func(ctx context.Context, request *http.Request) {
		token := fmt.Sprintf("%v", ctx.Value(tokenKey))
		request.Header.Add("X-Token", token)
	}
}

func newBearerTokenAuthMiddleware(token string) Middleware {
	return func(ctx context.Context, request *http.Request) {
		request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	}
}

func newBasicAuthMiddleware(login, password string) Middleware {
	return func(ctx context.Context, request *http.Request) {
		request.SetBasicAuth(login, password)
	}
}
