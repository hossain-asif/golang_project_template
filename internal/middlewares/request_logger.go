package middlewares

import (
	"context"
	"go_project_structure/common_pkg/logger"
	enums "go_project_structure/utils/enums"
	"net/http"
	"time"

	"github.com/google/uuid"
)

var requestLoggerMiddlewareLogger = logger.Log.Scope("", "middleware", "request_logger_middleware")

func RequestLoggerMiddleware(next http.Handler) http.Handler {
	log := requestLoggerMiddlewareLogger.Method("RequestLoggerMiddleware")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		log.Infof("Request received at: %s at time: %s", r.URL.Path, time.Now())

		var reqId uuid.UUID

		if header := r.Header.Get("Request-Id"); header != "" {
			reqId, _ = uuid.Parse(header)
		} else {
			reqId = uuid.New()
		}

		userSlug := r.Header.Get("X-User-Slug")

		// Add the UUID to the context
		ctx := r.Context()
		ctx = context.WithValue(ctx, enums.CtxRequestID, reqId.String())
		ctx = context.WithValue(ctx, enums.CtxUserSlug, userSlug)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
