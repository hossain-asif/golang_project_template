package middlewares

import (
	"go_project_structure/common_pkg/json"
	"go_project_structure/common_pkg/logger"
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

// global rate limiter: One aggressive user exhausts the bucket for everyone.
var limiter = rate.NewLimiter(rate.Every(1*time.Second), 5) // 5 request per second

var rateLimitMiddlewareLogger = logger.Log.Scope("", "middleware", "rate_limit_middleware")

func RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := rateLimitMiddlewareLogger.Method("RateLimitMiddleware").WithContext(r.Context())

		if !limiter.Allow() {
			log.Errorf("Too many requests")
			json.WriteJsonErrorResponse(w, http.StatusTooManyRequests, "Too Many Requests", nil)
			return
		}

		next.ServeHTTP(w, r)
	})
}
