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

		if header := r.Header.Get("Request-Id"); header != "" {
			reqId, _ = uuid.Parse(header)
		} else {
			reqId = uuid.New()
		}

		// Add the UUID to the context
		ctx := r.Context()
		ctx = context.WithValue(ctx, CtxRequestID, reqId.String())
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
