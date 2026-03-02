package middlewares

import (
	"context"
	"net/http"
	"strings"

	"go_project_structure/common_pkg/json"
	"go_project_structure/common_pkg/logger"
	env "go_project_structure/config/env"
	enums "go_project_structure/utils/enums"

	"github.com/golang-jwt/jwt/v5"
)

var jwtAuthMiddlewareLogger = logger.Log.Scope("", "middleware", "jwt_auth_middleware")

func JwtAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := jwtAuthMiddlewareLogger.Method("JwtAuthMiddleware").WithContext(r.Context())
		
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			log.Errorf("Authorization header missing")
			json.WriteJsonErrorResponse(w, http.StatusUnauthorized, "Authorization header missing", nil)
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			log.Errorf("Invalid authorization header format")
			json.WriteJsonErrorResponse(w, http.StatusUnauthorized, "Invalid authorization header format", nil)
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			log.Errorf("Token missing in authorization header")
			json.WriteJsonErrorResponse(w, http.StatusUnauthorized, "Token missing in authorization header", nil)
			return
		}

		claims := jwt.MapClaims{}

		_, err := jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(env.GetString("JWT_SECRET", "default_secret_key")), nil
		})

		if err != nil {
			log.Errorf("Invalid token: %v", err)
			json.WriteJsonErrorResponse(w, http.StatusUnauthorized, "Invalid token: "+err.Error(), nil)
			return
		}

		userEmail, okEmail := claims["email"].(string)
		if !okEmail {
			log.Errorf("Invalid token claims: email not found")
			json.WriteJsonErrorResponse(w, http.StatusUnauthorized, "Invalid token claims: email not found", nil)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, enums.CtxUserEmail, userEmail)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
