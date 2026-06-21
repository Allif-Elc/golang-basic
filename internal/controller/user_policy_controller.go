package controller

import (
	"context"
	"encoding/json"
	"golang-basic/api/internal/model"
	"golang-basic/api/internal/utility"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// UserPolicyServiceInterface defines the interface for user policy service operations
type UserPolicyServiceInterface interface {
	CreateUserPolicy(ctx context.Context, req *model.CreateUserPolicyRequest, createdBy *int64) (*model.UserPolicy, error)
	GetUserPolicyByID(ctx context.Context, id int64) (*model.UserPolicy, error)
	ListUserPolicies(ctx context.Context) ([]model.UserPolicy, error)
	ListUserPoliciesByUserID(ctx context.Context, userID int64) ([]model.UserPolicy, error)
	ListUserPolicyDetails(ctx context.Context) ([]model.UserPolicyDetail, error)
	UpdateUserPolicy(ctx context.Context, id int64, req *model.UpdateUserPolicyRequest) (*model.UserPolicy, error)
	DeleteUserPolicy(ctx context.Context, id int64) error
	ListPolicies(ctx context.Context) ([]model.Policy, error)
}

// UserPolicyController handles user-policy CRUD operations
type UserPolicyController struct {
	service UserPolicyServiceInterface
}

// NewUserPolicyController creates a new user policy controller instance
func NewUserPolicyController(service UserPolicyServiceInterface) *UserPolicyController {
	return &UserPolicyController{service: service}
}

// CreateUserPolicy handles POST /user-policies
// Creates a new user-to-policy assignment with priority
func (c *UserPolicyController) CreateUserPolicy(w http.ResponseWriter, r *http.Request) {
	var req model.CreateUserPolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utility.SendErrorResponse(w, utility.ValidationError("Invalid request body"))
		return
	}

	// Get created_by from context (set by JWT middleware)
	// For now, we'll pass nil since the user context might not be available
	// TODO: Extract user ID from JWT context when available
	userPolicy, err := c.service.CreateUserPolicy(r.Context(), &req, nil)
	if err != nil {
		utility.SendErrorResponse(w, err)
		return
	}

	utility.SendSuccess(w, http.StatusCreated, "User policy created successfully", userPolicy)
}

// GetUserPolicyByID handles GET /user-policies/{id}
func (c *UserPolicyController) GetUserPolicyByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utility.SendErrorResponse(w, utility.ValidationError("Invalid user policy ID"))
		return
	}

	userPolicy, err := c.service.GetUserPolicyByID(r.Context(), id)
	if err != nil {
		utility.SendErrorResponse(w, err)
		return
	}

	utility.SendSuccess(w, http.StatusOK, "User policy retrieved successfully", userPolicy)
}

// ListUserPolicies handles GET /user-policies
// Returns all user policies with basic information
func (c *UserPolicyController) ListUserPolicies(w http.ResponseWriter, r *http.Request) {
	userPolicies, err := c.service.ListUserPolicies(r.Context())
	if err != nil {
		utility.SendErrorResponse(w, utility.InternalError(err.Error()))
		return
	}

	utility.SendSuccess(w, http.StatusOK, "User policies retrieved successfully", userPolicies)
}

// ListUserPolicyDetails handles GET /user-policies/details
// Returns all user policies with related user and policy information
func (c *UserPolicyController) ListUserPolicyDetails(w http.ResponseWriter, r *http.Request) {
	details, err := c.service.ListUserPolicyDetails(r.Context())
	if err != nil {
		utility.SendErrorResponse(w, utility.InternalError(err.Error()))
		return
	}

	utility.SendSuccess(w, http.StatusOK, "User policy details retrieved successfully", details)
}

// GetUserPoliciesByUserID handles GET /user-policies/user/{userId}
// Returns all policies assigned to a specific user
func (c *UserPolicyController) GetUserPoliciesByUserID(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "userId")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		utility.SendErrorResponse(w, utility.ValidationError("Invalid user ID"))
		return
	}

	userPolicies, err := c.service.ListUserPoliciesByUserID(r.Context(), userID)
	if err != nil {
		utility.SendErrorResponse(w, utility.InternalError(err.Error()))
		return
	}

	utility.SendSuccess(w, http.StatusOK, "User policies retrieved successfully", userPolicies)
}

// UpdateUserPolicy handles PUT /user-policies/{id}
// Updates priority, is_active, or expires_at fields
func (c *UserPolicyController) UpdateUserPolicy(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utility.SendErrorResponse(w, utility.ValidationError("Invalid user policy ID"))
		return
	}

	var req model.UpdateUserPolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utility.SendErrorResponse(w, utility.ValidationError("Invalid request body"))
		return
	}

	userPolicy, err := c.service.UpdateUserPolicy(r.Context(), id, &req)
	if err != nil {
		utility.SendErrorResponse(w, err)
		return
	}

	utility.SendSuccess(w, http.StatusOK, "User policy updated successfully", userPolicy)
}

// DeleteUserPolicy handles DELETE /user-policies/{id}
func (c *UserPolicyController) DeleteUserPolicy(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utility.SendErrorResponse(w, utility.ValidationError("Invalid user policy ID"))
		return
	}

	err = c.service.DeleteUserPolicy(r.Context(), id)
	if err != nil {
		utility.SendErrorResponse(w, err)
		return
	}

	utility.SendSuccess(w, http.StatusOK, "User policy deleted successfully", nil)
}

// ListPolicies handles GET /policies
// Returns all available policies for use in user policy assignments
func (c *UserPolicyController) ListPolicies(w http.ResponseWriter, r *http.Request) {
	policies, err := c.service.ListPolicies(r.Context())
	if err != nil {
		utility.SendErrorResponse(w, utility.InternalError(err.Error()))
		return
	}

	utility.SendSuccess(w, http.StatusOK, "Policies retrieved successfully", policies)
}
