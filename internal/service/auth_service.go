package service

import (
	"context"
	"fmt"
	"time"

	"golang-basic/api/internal/model"
	"golang-basic/api/internal/repository"
	"golang-basic/api/internal/utility"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo   *repository.UserRepository
	jwtManager *utility.JWTManager
}

func NewAuthService(userRepo *repository.UserRepository) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		jwtManager: utility.GetJWTManager(),
	}
}

func (s *AuthService) Login(ctx context.Context, req model.LoginRequest) (*model.TokenResponse, error) {
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, utility.UnauthorizedError("Invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, utility.UnauthorizedError("Invalid credentials")
	}

	accessToken, err := s.jwtManager.GenerateToken(user.UserId, user.Email, 15*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.jwtManager.GenerateToken(user.UserId, user.Email, 7*24*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &model.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    900,
		TokenType:    "Bearer",
	}, nil
}

func (s *AuthService) Register(ctx context.Context, req model.RegisterRequest) (*model.User, *model.TokenResponse, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to hash password: %w", err)
	}

	createReq := model.CreateUserRequest{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	user, err := s.userRepo.Create(ctx, &createReq)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create user: %w", err)
	}

	accessToken, err := s.jwtManager.GenerateToken(user.UserId, user.Email, 15*time.Minute)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.jwtManager.GenerateToken(user.UserId, user.Email, 7*24*time.Hour)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	tokenResponse := &model.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    900,
		TokenType:    "Bearer",
	}

	return user, tokenResponse, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, req model.RefreshTokenRequest) (*model.TokenResponse, error) {
	claims, err := s.jwtManager.ValidateToken(req.RefreshToken)
	if err != nil {
		return nil, utility.UnauthorizedError("Invalid refresh token")
	}

	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, utility.UnauthorizedError("User not found")
	}

	if !user.IsActive {
		return nil, utility.UnauthorizedError("User account is inactive")
	}

	accessToken, err := s.jwtManager.GenerateToken(user.UserId, user.Email, 15*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, err := s.jwtManager.GenerateToken(user.UserId, user.Email, 7*24*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &model.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    900,
		TokenType:    "Bearer",
	}, nil
}

func (s *AuthService) ValidateToken(ctx context.Context, tokenString string) (*utility.Claims, error) {
	claims, err := s.jwtManager.ValidateToken(tokenString)
	if err != nil {
		return nil, utility.UnauthorizedError("Invalid token")
	}

	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, utility.UnauthorizedError("User not found")
	}

	if !user.IsActive {
		return nil, utility.UnauthorizedError("User account is inactive")
	}

	return claims, nil
}
