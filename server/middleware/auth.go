package middleware

import (
	"net/http"
	"strings"
)

func NewBasicAuthMiddleware(login, password string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, p, ok := r.BasicAuth()
			if !ok {
				w.WriteHeader(401)
				return
			}
			if u != login || p != password {
				w.WriteHeader(403)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func NewBearerAuthMiddleware(token string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				w.WriteHeader(401)
				return
			}

			values := strings.Split(header, " ")
			if len(values) < 2 || values[0] != "Bearer" || values[1] != token {
				w.WriteHeader(403)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
