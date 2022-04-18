package middleware

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	mw "github.com/go-chi/chi/middleware"
	"net/http"
	"os"
	"strings"
)

const requestIDHeaderName = "X-Request-Id"

var prefix string

func NewRequestIdMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if key, ok := r.Context().Value(mw.RequestIDKey).(string); ok {
				w.Header().Set(requestIDHeaderName, key)
			}
			next.ServeHTTP(w, r)
		}
		return mw.RequestID(http.HandlerFunc(fn))
	}
}

func SetRequestId(ctx context.Context) context.Context {
	id := mw.NextRequestID()
	return context.WithValue(ctx, mw.RequestIDKey, fmt.Sprintf("%s-%06d", prefix, id))
}

func GetRequestId(ctx context.Context) string {
	return mw.GetReqID(ctx)
}

func init() {
	hostname, err := os.Hostname()
	if hostname == "" || err != nil {
		hostname = "localhost"
	}
	var buf [12]byte
	var b64 string
	for len(b64) < 10 {
		_, _ = rand.Read(buf[:])
		b64 = base64.StdEncoding.EncodeToString(buf[:])
		b64 = strings.NewReplacer("+", "", "/", "").Replace(b64)
	}

	prefix = fmt.Sprintf("%s/%s", hostname, b64[0:10])
}
