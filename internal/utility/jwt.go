package utility

import (
	"crypto/rsa"
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
		return nil, fmt.Errorf("JWT_PRIVATE_KEY environment variable not set")
	}

	block, _ := pem.Decode([]byte(keyPEM))
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block containing private key")
	}

	key, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(keyPEM))
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSA private key: %w", err)
	}

	return key, nil
}

func loadPublicKey() (*rsa.PublicKey, error) {
	keyPEM := os.Getenv("JWT_PUBLIC_KEY")
	if keyPEM == "" {
		return nil, fmt.Errorf("JWT_PUBLIC_KEY environment variable not set")
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
