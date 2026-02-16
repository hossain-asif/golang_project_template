package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	env "go_project_structure/config/env"

	"github.com/golang-jwt/jwt/v5"
)

func JwtAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header missing", http.StatusUnauthorized)
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			http.Error(w, "Token missing in authorization header", http.StatusUnauthorized)
			return
		}

		fmt.Println("jwt token: ", token)

		claims := jwt.MapClaims{}

		_, err := jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(env.GetString("JWT_SECRET", "default_secret_key")), nil
		})

		if err != nil {
			http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		fmt.Println("claims: ", claims)

		userEmail, okEmail := claims["email"].(string)
		if !okEmail {
			http.Error(w, "Invalid token claims: email not found", http.StatusUnauthorized)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, "email", userEmail)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
