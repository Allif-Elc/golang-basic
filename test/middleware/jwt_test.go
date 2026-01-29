package middleware_test

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"golang-basic/api/internal/middleware"
	"golang-basic/api/internal/utility"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TestJWTMiddleware_Authenticate_Success tests successful authentication
func TestJWTMiddleware_Authenticate_Success(t *testing.T) {
	privateKey, err := generateTestRSAKey()
	if err != nil {
		t.Fatalf("Failed to generate test key: %v", err)
	}

	jwtManager := utility.NewJWTManager(privateKey, &privateKey.PublicKey)

	token, err := jwtManager.GenerateToken(123, "test@example.com", time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	mid := middleware.NewJWTMiddlewareWithManager(jwtManager)
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(middleware.UserIDKey)
		email := r.Context().Value(middleware.UserEmailKey)

		if userID == nil || email == nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"user_id": userID,
			"email":   email,
		})
	})

	authHandler := mid.Authenticate()(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	authHandler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	if response["user_id"] != float64(123) {
		t.Errorf("Expected user_id 123, got %v", response["user_id"])
	}

	if response["email"] != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got %v", response["email"])
	}
}

// TestJWTMiddleware_Authenticate_MissingHeader tests missing authorization header
func TestJWTMiddleware_Authenticate_MissingHeader(t *testing.T) {
	privateKey, err := generateTestRSAKey()
	if err != nil {
		t.Fatalf("Failed to generate test key: %v", err)
	}

	jwtManager := utility.NewJWTManager(privateKey, &privateKey.PublicKey)

	mid := middleware.NewJWTMiddlewareWithManager(jwtManager)
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	authHandler := mid.Authenticate()(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	w := httptest.NewRecorder()

	authHandler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	if response["status"] != "error" {
		t.Errorf("Expected status 'error', got %v", response["status"])
	}
}

// TestJWTMiddleware_Authenticate_InvalidFormat tests invalid authorization header format
func TestJWTMiddleware_Authenticate_InvalidFormat(t *testing.T) {
	privateKey, err := generateTestRSAKey()
	if err != nil {
		t.Fatalf("Failed to generate test key: %v", err)
	}

	jwtManager := utility.NewJWTManager(privateKey, &privateKey.PublicKey)

	mid := middleware.NewJWTMiddlewareWithManager(jwtManager)
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	authHandler := mid.Authenticate()(testHandler)

	tests := []struct {
		name   string
		header string
	}{
		{
			name:   "no Bearer prefix",
			header: "invalid-token",
		},
		{
			name:   "missing token",
			header: "Bearer ",
		},
		{
			name:   "wrong prefix case",
			header: "bearer token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
			req.Header.Set("Authorization", tt.header)
			w := httptest.NewRecorder()

			authHandler.ServeHTTP(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
			}
		})
	}
}

// TestJWTMiddleware_Authenticate_InvalidToken tests invalid token
func TestJWTMiddleware_Authenticate_InvalidToken(t *testing.T) {
	privateKey, err := generateTestRSAKey()
	if err != nil {
		t.Fatalf("Failed to generate test key: %v", err)
	}

	jwtManager := utility.NewJWTManager(privateKey, &privateKey.PublicKey)

	mid := middleware.NewJWTMiddlewareWithManager(jwtManager)
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	authHandler := mid.Authenticate()(testHandler)

	tests := []struct {
		name   string
		token  string
	}{
		{
			name:  "malformed token",
			token: "not.a.valid.token",
		},
		{
			name:  "expired token",
			token: generateExpiredTokenJWT(privateKey),
		},
		{
			name:  "random string",
			token: "randomstring",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
			req.Header.Set("Authorization", "Bearer "+tt.token)
			w := httptest.NewRecorder()

			authHandler.ServeHTTP(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
			}
		})
	}
}

// TestJWTMiddleware_OptionalAuth_WithToken tests optional auth with valid token
func TestJWTMiddleware_OptionalAuth_WithToken(t *testing.T) {
	privateKey, err := generateTestRSAKey()
	if err != nil {
		t.Fatalf("Failed to generate test key: %v", err)
	}

	jwtManager := utility.NewJWTManager(privateKey, &privateKey.PublicKey)

	token, err := jwtManager.GenerateToken(456, "optional@example.com", time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	mid := middleware.NewJWTMiddlewareWithManager(jwtManager)
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(middleware.UserIDKey)
		email := r.Context().Value(middleware.UserEmailKey)

		if userID == nil {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"status": "no_user"})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"user_id": userID,
			"email":   email,
		})
	})

	authHandler := mid.OptionalAuth()(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	authHandler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	if response["user_id"] != float64(456) {
		t.Errorf("Expected user_id 456, got %v", response["user_id"])
	}
}

// TestJWTMiddleware_OptionalAuth_WithoutToken tests optional auth without token
func TestJWTMiddleware_OptionalAuth_WithoutToken(t *testing.T) {
	privateKey, err := generateTestRSAKey()
	if err != nil {
		t.Fatalf("Failed to generate test key: %v", err)
	}

	jwtManager := utility.NewJWTManager(privateKey, &privateKey.PublicKey)

	mid := middleware.NewJWTMiddlewareWithManager(jwtManager)
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(middleware.UserIDKey)
		email := r.Context().Value(middleware.UserEmailKey)

		if userID == nil && email == nil {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"status": "no_user"})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"user_id": userID,
			"email":   email,
		})
	})

	authHandler := mid.OptionalAuth()(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	w := httptest.NewRecorder()

	authHandler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	if response["status"] != "no_user" {
		t.Errorf("Expected status 'no_user', got %v", response["status"])
	}
}

// TestJWTMiddleware_OptionalAuth_InvalidToken tests optional auth with invalid token
func TestJWTMiddleware_OptionalAuth_InvalidToken(t *testing.T) {
	privateKey, err := generateTestRSAKey()
	if err != nil {
		t.Fatalf("Failed to generate test key: %v", err)
	}

	jwtManager := utility.NewJWTManager(privateKey, &privateKey.PublicKey)

	mid := middleware.NewJWTMiddlewareWithManager(jwtManager)
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(middleware.UserIDKey)
		email := r.Context().Value(middleware.UserEmailKey)

		if userID == nil && email == nil {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"status": "no_user"})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"user_id": userID,
			"email":   email,
		})
	})

	authHandler := mid.OptionalAuth()(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	w := httptest.NewRecorder()

	authHandler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	if response["status"] != "no_user" {
		t.Errorf("Expected status 'no_user', got %v", response["status"])
	}
}

// TestJWTMiddleware_ContextValues tests that context values are properly set
func TestJWTMiddleware_ContextValues(t *testing.T) {
	privateKey, err := generateTestRSAKey()
	if err != nil {
		t.Fatalf("Failed to generate test key: %v", err)
	}

	jwtManager := utility.NewJWTManager(privateKey, &privateKey.PublicKey)

	testUserID := int64(789)
	testEmail := "context@example.com"
	token, err := jwtManager.GenerateToken(testUserID, testEmail, time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	mid := middleware.NewJWTMiddlewareWithManager(jwtManager)
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(middleware.UserIDKey)
		email := r.Context().Value(middleware.UserEmailKey)

		if userID == nil || email == nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "context values missing"})
			return
		}

		userIDInt, ok := userID.(int64)
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid user ID type"})
			return
		}

		emailStr, ok := email.(string)
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid email type"})
			return
		}

		if userIDInt != testUserID {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "user ID mismatch"})
			return
		}

		if emailStr != testEmail {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "email mismatch"})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	authHandler := mid.Authenticate()(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	authHandler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)

		var response map[string]interface{}
		json.NewDecoder(w.Body).Decode(&response)
		if errMsg, ok := response["error"].(string); ok {
			t.Errorf("Handler error: %s", errMsg)
		}
	}
}

// TestJWTMiddleware_Chained tests JWT middleware chained with other middleware
func TestJWTMiddleware_Chained(t *testing.T) {
	privateKey, err := generateTestRSAKey()
	if err != nil {
		t.Fatalf("Failed to generate test key: %v", err)
	}

	jwtManager := utility.NewJWTManager(privateKey, &privateKey.PublicKey)

	token, err := jwtManager.GenerateToken(999, "chained@example.com", time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	mid := middleware.NewJWTMiddlewareWithManager(jwtManager)

	// Custom middleware that adds a header
	customMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Custom-Header", "test-value")
			next.ServeHTTP(w, r)
		})
	}

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(middleware.UserIDKey)
		if userID == nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"user_id": userID,
			"custom_header": w.Header().Get("X-Custom-Header"),
		})
	})

	// Chain: custom -> JWT auth -> handler
	authHandler := mid.Authenticate()(testHandler)
	chainedHandler := customMiddleware(authHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	chainedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	if w.Header().Get("X-Custom-Header") != "test-value" {
		t.Error("Custom header not set by middleware")
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	if response["user_id"] != float64(999) {
		t.Errorf("Expected user_id 999, got %v", response["user_id"])
	}
}

// Benchmark tests
func BenchmarkJWTMiddleware_Authenticate(b *testing.B) {
	privateKey, _ := generateTestRSAKey()
	jwtManager := utility.NewJWTManager(privateKey, &privateKey.PublicKey)
	token, _ := jwtManager.GenerateToken(123, "test@example.com", time.Hour)

	mid := middleware.NewJWTMiddlewareWithManager(jwtManager)
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	authHandler := mid.Authenticate()(testHandler)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		authHandler.ServeHTTP(w, req)
	}
}

func BenchmarkJWTMiddleware_OptionalAuth(b *testing.B) {
	privateKey, _ := generateTestRSAKey()
	jwtManager := utility.NewJWTManager(privateKey, &privateKey.PublicKey)
	token, _ := jwtManager.GenerateToken(123, "test@example.com", time.Hour)

	mid := middleware.NewJWTMiddlewareWithManager(jwtManager)
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	authHandler := mid.OptionalAuth()(testHandler)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		authHandler.ServeHTTP(w, req)
	}
}

// Helper functions
func generateTestRSAKey() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, 2048)
}

func generateExpiredTokenJWT(privateKey *rsa.PrivateKey) string {
	claims := &utility.Claims{
		UserID: 123,
		Email:  "test@example.com",
	}
	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(-time.Hour))
	claims.IssuedAt = jwt.NewNumericDate(time.Now().Add(-2 * time.Hour))
	claims.NotBefore = jwt.NewNumericDate(time.Now().Add(-2 * time.Hour))

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, _ := token.SignedString(privateKey)
	return tokenString
}
