package utility

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Test JWTManager methods
func TestJWTManager_GenerateToken(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate test key: %v", err)
	}

	jwtManager := &JWTManager{
		privateKey: privateKey,
		publicKey:  &privateKey.PublicKey,
	}

	tests := []struct {
		name       string
		userID     int64
		email      string
		expiration time.Duration
		wantErr    bool
	}{
		{
			name:       "valid token generation",
			userID:     123,
			email:      "test@example.com",
			expiration: time.Hour,
			wantErr:    false,
		},
		{
			name:       "zero user ID",
			userID:     0,
			email:      "test@example.com",
			expiration: time.Hour,
			wantErr:    false,
		},
		{
			name:       "negative user ID",
			userID:     -1,
			email:      "test@example.com",
			expiration: time.Hour,
			wantErr:    false,
		},
		{
			name:       "empty email",
			userID:     123,
			email:      "",
			expiration: time.Hour,
			wantErr:    false,
		},
		{
			name:       "short expiration",
			userID:     123,
			email:      "test@example.com",
			expiration: time.Millisecond,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := jwtManager.GenerateToken(tt.userID, tt.email, tt.expiration)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && token == "" {
				t.Error("GenerateToken() returned empty token")
			}
		})
	}
}

func TestJWTManager_ValidateToken(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate test key: %v", err)
	}

	jwtManager := &JWTManager{
		privateKey: privateKey,
		publicKey:  &privateKey.PublicKey,
	}

	validToken, _ := jwtManager.GenerateToken(123, "test@example.com", time.Hour)

	tests := []struct {
		name      string
		token     string
		wantErr   bool
		wantUserID int64
		wantEmail string
	}{
		{
			name:      "valid token",
			token:     validToken,
			wantErr:   false,
			wantUserID: 123,
			wantEmail: "test@example.com",
		},
		{
			name:    "empty token",
			token:   "",
			wantErr: true,
		},
		{
			name:    "invalid token format",
			token:   "invalid.token.string",
			wantErr: true,
		},
		{
			name:    "malformed token",
			token:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ",
			wantErr: true,
		},
		{
			name:    "expired token",
			token:   generateExpiredToken(privateKey),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := jwtManager.ValidateToken(tt.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if claims.UserID != tt.wantUserID {
					t.Errorf("ValidateToken() userID = %v, want %v", claims.UserID, tt.wantUserID)
				}
				if claims.Email != tt.wantEmail {
					t.Errorf("ValidateToken() email = %v, want %v", claims.Email, tt.wantEmail)
				}
			}
		})
	}
}

func TestJWTManager_RefreshToken(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate test key: %v", err)
	}

	jwtManager := &JWTManager{
		privateKey: privateKey,
		publicKey:  &privateKey.PublicKey,
	}

	validToken, _ := jwtManager.GenerateToken(123, "test@example.com", time.Hour)

	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "valid token refresh",
			token:   validToken,
			wantErr: false,
		},
		{
			name:    "invalid token refresh",
			token:   "invalid.token",
			wantErr: true,
		},
		{
			name:    "empty token refresh",
			token:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newToken, err := jwtManager.RefreshToken(tt.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("RefreshToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if newToken == "" {
					t.Error("RefreshToken() returned empty token")
				}
				if newToken == tt.token {
					t.Error("RefreshToken() returned same token")
				}
				claims, err := jwtManager.ValidateToken(newToken)
				if err != nil {
					t.Errorf("RefreshToken() produced invalid token: %v", err)
				}
				if claims.UserID != 123 {
					t.Errorf("RefreshToken() userID = %v, want 123", claims.UserID)
				}
			}
		})
	}
}

func TestJWTManager_Integration(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate test key: %v", err)
	}

	jwtManager := &JWTManager{
		privateKey: privateKey,
		publicKey:  &privateKey.PublicKey,
	}

	userID := int64(456)
	email := "integration@example.com"
	expiration := 24 * time.Hour

	token, err := jwtManager.GenerateToken(userID, email, expiration)
	if err != nil {
		t.Fatalf("GenerateToken() failed: %v", err)
	}

	claims, err := jwtManager.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken() failed: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("UserID = %v, want %v", claims.UserID, userID)
	}
	if claims.Email != email {
		t.Errorf("Email = %v, want %v", claims.Email, email)
	}

	refreshedToken, err := jwtManager.RefreshToken(token)
	if err != nil {
		t.Fatalf("RefreshToken() failed: %v", err)
	}

	refreshedClaims, err := jwtManager.ValidateToken(refreshedToken)
	if err != nil {
		t.Fatalf("ValidateToken() on refreshed token failed: %v", err)
	}

	if refreshedClaims.UserID != userID {
		t.Errorf("Refreshed UserID = %v, want %v", refreshedClaims.UserID, userID)
	}
	if refreshedClaims.Email != email {
		t.Errorf("Refreshed Email = %v, want %v", refreshedClaims.Email, email)
	}

	if refreshedToken == token {
		t.Error("Refreshed token is identical to original token")
	}
}

// Test loadPrivateKey with PKCS1 format
func TestLoadPrivateKey_PKCS1(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate test key: %v", err)
	}

	pkcs1PEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "pkcs1-private.pem")
	if err := os.WriteFile(keyPath, pkcs1PEM, 0600); err != nil {
		t.Fatalf("Failed to write test key file: %v", err)
	}

	t.Setenv("JWT_PRIVATE_KEY_PATH", keyPath)
	t.Setenv("JWT_PUBLIC_KEY_PATH", "") // Use default

	loadedKey, err := loadPrivateKey()
	if err != nil {
		t.Fatalf("loadPrivateKey() failed: %v", err)
	}

	if loadedKey.N.Cmp(privateKey.N) != 0 {
		t.Error("Loaded key modulus doesn't match original key")
	}
	if loadedKey.D.Cmp(privateKey.D) != 0 {
		t.Error("Loaded key private exponent doesn't match original key")
	}
}

// Test loadPrivateKey with PKCS8 format
func TestLoadPrivateKey_PKCS8(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate test key: %v", err)
	}

	pkcs8Bytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		t.Fatalf("Failed to marshal PKCS8 key: %v", err)
	}

	pkcs8PEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: pkcs8Bytes,
	})

	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "pkcs8-private.pem")
	if err := os.WriteFile(keyPath, pkcs8PEM, 0600); err != nil {
		t.Fatalf("Failed to write test key file: %v", err)
	}

	t.Setenv("JWT_PRIVATE_KEY_PATH", keyPath)
	t.Setenv("JWT_PUBLIC_KEY_PATH", "")

	loadedKey, err := loadPrivateKey()
	if err != nil {
		t.Fatalf("loadPrivateKey() failed: %v", err)
	}

	if loadedKey.N.Cmp(privateKey.N) != 0 {
		t.Error("Loaded key modulus doesn't match original key")
	}
	if loadedKey.D.Cmp(privateKey.D) != 0 {
		t.Error("Loaded key private exponent doesn't match original key")
	}
}

// Test loadPrivateKey from environment variable
func TestLoadPrivateKey_FromEnv(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate test key: %v", err)
	}

	pkcs8Bytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		t.Fatalf("Failed to marshal PKCS8 key: %v", err)
	}

	pkcs8PEM := string(pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: pkcs8Bytes,
	}))

	t.Setenv("JWT_PRIVATE_KEY", pkcs8PEM)
	t.Setenv("JWT_PRIVATE_KEY_PATH", "")
	t.Setenv("JWT_PUBLIC_KEY_PATH", "")

	loadedKey, err := loadPrivateKey()
	if err != nil {
		t.Fatalf("loadPrivateKey() failed: %v", err)
	}

	if loadedKey.N.Cmp(privateKey.N) != 0 {
		t.Error("Loaded key modulus doesn't match original key")
	}
}

// Test loadPrivateKey error cases
func TestLoadPrivateKey_Errors(t *testing.T) {
	tests := []struct {
		name       string
		setup      func()
		wantErrMsg string
	}{
		{
			name: "file not found",
			setup: func() {
				t.Setenv("JWT_PRIVATE_KEY", "")
				t.Setenv("JWT_PRIVATE_KEY_PATH", "/nonexistent/key.pem")
			},
			wantErrMsg: "failed to read private key file",
		},
		{
			name: "invalid PEM format",
			setup: func() {
				t.Setenv("JWT_PRIVATE_KEY", "invalid pem content")
				t.Setenv("JWT_PRIVATE_KEY_PATH", "")
			},
			wantErrMsg: "failed to decode PEM block",
		},
		{
			name: "unsupported key type",
			setup: func() {
				ecKey := pem.EncodeToMemory(&pem.Block{
					Type:  "EC PRIVATE KEY",
					Bytes: []byte("not a real EC key"),
				})
				t.Setenv("JWT_PRIVATE_KEY", string(ecKey))
				t.Setenv("JWT_PRIVATE_KEY_PATH", "")
			},
			wantErrMsg: "unsupported key type",
		},
		{
			name: "invalid PKCS8 content",
			setup: func() {
				invalidPEM := pem.EncodeToMemory(&pem.Block{
					Type:  "PRIVATE KEY",
					Bytes: []byte("invalid pkcs8 content"),
				})
				t.Setenv("JWT_PRIVATE_KEY", string(invalidPEM))
				t.Setenv("JWT_PRIVATE_KEY_PATH", "")
			},
			wantErrMsg: "failed to parse PKCS8 private key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			_, err := loadPrivateKey()
			if err == nil {
				t.Error("loadPrivateKey() expected error, got nil")
				return
			}
			if tt.wantErrMsg != "" {
				errMsg := err.Error()
				// Check if error message contains expected substring
				found := false
				for i := 0; i <= len(errMsg)-len(tt.wantErrMsg); i++ {
					if errMsg[i:i+len(tt.wantErrMsg)] == tt.wantErrMsg {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("loadPrivateKey() error = %v, want contains %v", err, tt.wantErrMsg)
				}
			}
		})
	}
}

// Test loadPublicKey
func TestLoadPublicKey(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate test key: %v", err)
	}

	pubKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(&privateKey.PublicKey),
	})

	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "public.pem")
	if err := os.WriteFile(keyPath, pubKeyPEM, 0644); err != nil {
		t.Fatalf("Failed to write test key file: %v", err)
	}

	t.Setenv("JWT_PUBLIC_KEY_PATH", keyPath)
	t.Setenv("JWT_PRIVATE_KEY_PATH", "")

	loadedKey, err := loadPublicKey()
	if err != nil {
		t.Fatalf("loadPublicKey() failed: %v", err)
	}

	if loadedKey.N.Cmp(privateKey.N) != 0 {
		t.Error("Loaded key modulus doesn't match original key")
	}
	if loadedKey.E != privateKey.E {
		t.Error("Loaded key exponent doesn't match original key")
	}
}

// Helper function to generate an expired token for testing
func generateExpiredToken(privateKey *rsa.PrivateKey) string {
	claims := &Claims{
		UserID: 123,
		Email:  "test@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			NotBefore: jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, _ := token.SignedString(privateKey)
	return tokenString
}

// Benchmark tests
func BenchmarkJWTManager_GenerateToken(b *testing.B) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		b.Fatalf("Failed to generate test key: %v", err)
	}

	jwtManager := &JWTManager{
		privateKey: privateKey,
		publicKey:  &privateKey.PublicKey,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = jwtManager.GenerateToken(123, "test@example.com", time.Hour)
	}
}

func BenchmarkJWTManager_ValidateToken(b *testing.B) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		b.Fatalf("Failed to generate test key: %v", err)
	}

	jwtManager := &JWTManager{
		privateKey: privateKey,
		publicKey:  &privateKey.PublicKey,
	}

	token, _ := jwtManager.GenerateToken(123, "test@example.com", time.Hour)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = jwtManager.ValidateToken(token)
	}
}

func BenchmarkLoadPrivateKey_PKCS1(b *testing.B) {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	pkcs1PEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	tmpDir := b.TempDir()
	keyPath := filepath.Join(tmpDir, "bench-key.pem")
	_ = os.WriteFile(keyPath, pkcs1PEM, 0600)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.Setenv("JWT_PRIVATE_KEY_PATH", keyPath)
		_, _ = loadPrivateKey()
	}
}

func BenchmarkLoadPrivateKey_PKCS8(b *testing.B) {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	pkcs8Bytes, _ := x509.MarshalPKCS8PrivateKey(privateKey)
	pkcs8PEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: pkcs8Bytes,
	})

	tmpDir := b.TempDir()
	keyPath := filepath.Join(tmpDir, "bench-key.pem")
	_ = os.WriteFile(keyPath, pkcs8PEM, 0600)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.Setenv("JWT_PRIVATE_KEY_PATH", keyPath)
		_, _ = loadPrivateKey()
	}
}

// Generate deterministic test key for consistency
func generateTestKey() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, 2048)
}

// Test edge case: non-RSA key in PKCS8 wrapper
func TestLoadPrivateKey_NonRSAKeyInPKCS8(t *testing.T) {
	// Create a mock PKCS8 PEM that parses but isn't RSA
	mockPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: []byte("mock data"),
	})

	t.Setenv("JWT_PRIVATE_KEY", string(mockPEM))
	t.Setenv("JWT_PRIVATE_KEY_PATH", "")

	_, err := loadPrivateKey()
	if err == nil {
		t.Error("loadPrivateKey() expected error for non-RSA key, got nil")
	}
}

// Test concurrent access to JWTManager
func TestJWTManager_Concurrent(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate test key: %v", err)
	}

	jwtManager := &JWTManager{
		privateKey: privateKey,
		publicKey:  &privateKey.PublicKey,
	}

	token, _ := jwtManager.GenerateToken(123, "test@example.com", time.Hour)

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_, _ = jwtManager.GenerateToken(int64(j), "test@example.com", time.Hour)
				_, _ = jwtManager.ValidateToken(token)
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}
