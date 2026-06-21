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

// PolicyServiceInterface defines the interface for policy service operations
type PolicyServiceInterface interface {
	CreatePolicy(ctx context.Context, req model.CreatePolicyRequest) (*model.Policy, error)
	GetPolicyByID(ctx context.Context, id int64) (*model.Policy, error)
	ListPolicies(ctx context.Context) ([]model.Policy, error)
	UpdatePolicy(ctx context.Context, id int64, req model.UpdatePolicyRequest) (*model.Policy, error)
	DeletePolicy(ctx context.Context, id int64) error
}

type PolicyController struct {
	service PolicyServiceInterface
}

func NewPolicyController(service PolicyServiceInterface) *PolicyController {
	return &PolicyController{service: service}
}

// CreatePolicy handles POST /policies
func (c *PolicyController) CreatePolicy(w http.ResponseWriter, r *http.Request) {
	var req model.CreatePolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utility.SendErrorResponse(w, utility.ValidationError("Invalid request body"))
		return
	}

	policy, err := c.service.CreatePolicy(r.Context(), req)
	if err != nil {
		utility.SendErrorResponse(w, err)
		return
	}

	utility.SendSuccess(w, http.StatusCreated, "Policy created successfully", policy)
}

// GetPolicyByID handles GET /policies/{id}
func (c *PolicyController) GetPolicyByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utility.SendErrorResponse(w, utility.ValidationError("Invalid policy ID"))
		return
	}

	policy, err := c.service.GetPolicyByID(r.Context(), id)
	if err != nil {
		utility.SendErrorResponse(w, err)
		return
	}

	utility.SendSuccess(w, http.StatusOK, "Policy retrieved successfully", policy)
}

// ListPolicies handles GET /policies
func (c *PolicyController) ListPolicies(w http.ResponseWriter, r *http.Request) {
	policies, err := c.service.ListPolicies(r.Context())
	if err != nil {
		utility.SendErrorResponse(w, utility.InternalError(err.Error()))
		return
	}

	utility.SendSuccess(w, http.StatusOK, "Policies retrieved successfully", policies)
}

// UpdatePolicy handles PUT /policies/{id}
func (c *PolicyController) UpdatePolicy(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utility.SendErrorResponse(w, utility.ValidationError("Invalid policy ID"))
		return
	}

	var req model.UpdatePolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utility.SendErrorResponse(w, utility.ValidationError("Invalid request body"))
		return
	}

	policy, err := c.service.UpdatePolicy(r.Context(), id, req)
	if err != nil {
		utility.SendErrorResponse(w, err)
		return
	}

	utility.SendSuccess(w, http.StatusOK, "Policy updated successfully", policy)
}

// DeletePolicy handles DELETE /policies/{id}
func (c *PolicyController) DeletePolicy(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utility.SendErrorResponse(w, utility.ValidationError("Invalid policy ID"))
		return
	}

	err = c.service.DeletePolicy(r.Context(), id)
	if err != nil {
		utility.SendErrorResponse(w, err)
		return
	}

	utility.SendSuccess(w, http.StatusOK, "Policy deleted successfully", nil)
}
