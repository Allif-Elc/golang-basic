package controller

import (
	"context"
	"encoding/json"
	"golang-basic/api/internal/middleware"
	"golang-basic/api/internal/model"
	"golang-basic/api/internal/utility"
	"net/http"
)

type AuthServiceInterface interface {
	Login(ctx context.Context, req model.LoginRequest) (*model.TokenResponse, error)
	Register(ctx context.Context, req model.RegisterRequest) (*model.User, *model.TokenResponse, error)
	RefreshToken(ctx context.Context, req model.RefreshTokenRequest) (*model.TokenResponse, error)
}

type AuthController struct {
	service AuthServiceInterface
}

func NewAuthController(service AuthServiceInterface) *AuthController {
	return &AuthController{service: service}
}

func (c *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utility.SendErrorResponse(w, utility.ValidationError("Invalid request body"))
		return
	}

	tokens, err := c.service.Login(r.Context(), req)
	if err != nil {
		utility.SendErrorResponse(w, err)
		return
	}

	utility.SendSuccess(w, http.StatusOK, "Login successful", tokens)
}

func (c *AuthController) Register(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utility.SendErrorResponse(w, utility.ValidationError("Invalid request body"))
		return
	}

	user, tokens, err := c.service.Register(r.Context(), req)
	if err != nil {
		utility.SendErrorResponse(w, err)
		return
	}

	response := map[string]interface{}{
		"user":   user,
		"tokens": tokens,
	}

	utility.SendSuccess(w, http.StatusCreated, "Registration successful", response)
}

func (c *AuthController) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req model.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utility.SendErrorResponse(w, utility.ValidationError("Invalid request body"))
		return
	}

	tokens, err := c.service.RefreshToken(r.Context(), req)
	if err != nil {
		utility.SendErrorResponse(w, err)
		return
	}

	utility.SendSuccess(w, http.StatusOK, "Token refreshed successfully", tokens)
}

func (c *AuthController) Logout(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		utility.SendErrorResponse(w, utility.UnauthorizedError("User not authenticated"))
		return
	}

	utility.SendSuccess(w, http.StatusOK, "Logout successful", map[string]interface{}{
		"user_id": userID,
		"message": "Token invalidated client-side",
	})
}
