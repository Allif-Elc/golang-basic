package utility

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

// NewJWTManager creates a new JWTManager with the provided RSA keys.
// This is primarily intended for testing purposes.
func NewJWTManager(privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey) *JWTManager {
	return &JWTManager{
		privateKey: privateKey,
		publicKey:  publicKey,
	}
}

var (
	jwtManager *JWTManager
	jwtOnce    sync.Once
)

func GetJWTManager() *JWTManager {
	jwtOnce.Do(func() {
		privateKey, err := loadPrivateKey()
		if err != nil {
			panic(fmt.Sprintf("failed to load JWT private key: %v", err))
		}

		publicKey, err := loadPublicKey()
		if err != nil {
			panic(fmt.Sprintf("failed to load JWT public key: %v", err))
		}

		jwtManager = &JWTManager{
			privateKey: privateKey,
			publicKey:  publicKey,
		}
	})
	return jwtManager
}

func loadPrivateKey() (*rsa.PrivateKey, error) {
	keyPEM := os.Getenv("JWT_PRIVATE_KEY")
	if keyPEM == "" {
		keyPath := os.Getenv("JWT_PRIVATE_KEY_PATH")
		if keyPath == "" {
			keyPath = "keys/jwt-private-key.pem"
		}
		data, err := os.ReadFile(keyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read private key file %s: %w", keyPath, err)
		}
		keyPEM = string(data)
	}

	block, _ := pem.Decode([]byte(keyPEM))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	var key *rsa.PrivateKey
	var err error

	if block.Type == "RSA PRIVATE KEY" {
		key, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse PKCS1 private key: %w", err)
		}
	} else if block.Type == "PRIVATE KEY" {
		parsedKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse PKCS8 private key: %w", err)
		}
		var ok bool
		key, ok = parsedKey.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("not an RSA private key")
		}
	} else {
		return nil, fmt.Errorf("unsupported key type: %s", block.Type)
	}

	return key, nil
}

func loadPublicKey() (*rsa.PublicKey, error) {
	keyPEM := os.Getenv("JWT_PUBLIC_KEY")
	if keyPEM == "" {
		keyPath := os.Getenv("JWT_PUBLIC_KEY_PATH")
		if keyPath == "" {
			keyPath = "keys/jwt-public-key.pem"
		}
		data, err := os.ReadFile(keyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read public key file %s: %w", keyPath, err)
		}
		keyPEM = string(data)
	}

	key, err := jwt.ParseRSAPublicKeyFromPEM([]byte(keyPEM))
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSA public key: %w", err)
	}

	return key, nil
}

func (j *JWTManager) GenerateToken(userID int64, email string, expiration time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(expiration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(j.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

func (j *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func (j *JWTManager) RefreshToken(tokenString string) (string, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return "", fmt.Errorf("invalid token: %w", err)
	}

	newToken, err := j.GenerateToken(claims.UserID, claims.Email, 7*24*time.Hour)
	if err != nil {
		return "", fmt.Errorf("failed to generate new token: %w", err)
	}

	return newToken, nil
}
