package common_middlewares

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func RequestLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Request received at:", r.URL.Path, "at time:", time.Now())

		var reqId uuid.UUID

		// If X-Request-Id header is present, use it
		if r.Header.Get("Request-Id") != "" {
			reqId, _ = uuid.Parse(r.Header.Get("Request-Id"))
		}

		// Check if X-Request-Id header is present
		if r.Header.Get("Request-Id") == "" {
			reqId = uuid.New()
		}

		// Add the UUID to the context
		ctx := r.Context()
		ctx = context.WithValue(ctx, "requestId", reqId.String())
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
