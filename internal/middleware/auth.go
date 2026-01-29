package middleware

import (
	"context"
	"net/http"
	"strings"

	"golang-basic/api/internal/utility"
)

type JWTMiddleware struct {
	jwtManager *utility.JWTManager
}

func NewJWTMiddleware() *JWTMiddleware {
	return &JWTMiddleware{
		jwtManager: utility.GetJWTManager(),
	}
}

// NewJWTMiddlewareWithManager creates a new JWTMiddleware with a custom JWTManager for testing
func NewJWTMiddlewareWithManager(manager *utility.JWTManager) *JWTMiddleware {
	return &JWTMiddleware{
		jwtManager: manager,
	}
}

func (j *JWTMiddleware) Authenticate() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				utility.SendErrorResponse(w, utility.UnauthorizedError("Missing authorization header"))
				return
			}

			if !strings.HasPrefix(authHeader, "Bearer ") {
				utility.SendErrorResponse(w, utility.UnauthorizedError("Invalid authorization header format"))
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := j.jwtManager.ValidateToken(tokenString)
			if err != nil {
				utility.SendErrorResponse(w, utility.UnauthorizedError("Invalid token"))
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, UserEmailKey, claims.Email)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (j *JWTMiddleware) OptionalAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				next.ServeHTTP(w, r)
				return
			}

			if !strings.HasPrefix(authHeader, "Bearer ") {
				next.ServeHTTP(w, r)
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := j.jwtManager.ValidateToken(tokenString)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, UserEmailKey, claims.Email)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
