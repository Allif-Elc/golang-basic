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

// UserAttributeServiceInterface defines the interface for user attribute service operations
type UserAttributeServiceInterface interface {
	ListUserAttributes(ctx context.Context, userID int64) ([]model.UserAttributeDetail, error)
	GetUserAttributeByID(ctx context.Context, id int64) (*model.UserAttributeDetail, error)
	CreateUserAttribute(ctx context.Context, req model.CreateUserAttributeRequest) (*model.UserAttribute, error)
	UpdateUserAttribute(ctx context.Context, id int64, req model.UpdateUserAttributeRequest) (*model.UserAttribute, error)
	DeleteUserAttribute(ctx context.Context, id int64) error
}

type UserAttributeController struct {
	service UserAttributeServiceInterface
}

func NewUserAttributeController(service UserAttributeServiceInterface) *UserAttributeController {
	return &UserAttributeController{service: service}
}

// ListUserAttributes handles GET /user-attributes?user_id={id}
func (c *UserAttributeController) ListUserAttributes(w http.ResponseWriter, r *http.Request) {
	// Parse optional user_id query parameter
	var userID int64
	if userIDStr := r.URL.Query().Get("user_id"); userIDStr != "" {
		id, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			utility.SendError(w, http.StatusBadRequest, "Invalid user_id parameter")
			return
		}
		userID = id
	}

	attributes, err := c.service.ListUserAttributes(r.Context(), userID)
	if err != nil {
		utility.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utility.SendSuccess(w, http.StatusOK, "User attributes retrieved successfully", attributes)
}

// GetUserAttributeByID handles GET /user-attributes/{id}
func (c *UserAttributeController) GetUserAttributeByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utility.SendError(w, http.StatusBadRequest, "Invalid user attribute ID")
		return
	}

	attribute, err := c.service.GetUserAttributeByID(r.Context(), id)
	if err != nil {
		utility.SendError(w, http.StatusNotFound, err.Error())
		return
	}

	utility.SendSuccess(w, http.StatusOK, "User attribute retrieved successfully", attribute)
}

// CreateUserAttribute handles POST /user-attributes
func (c *UserAttributeController) CreateUserAttribute(w http.ResponseWriter, r *http.Request) {
	var req model.CreateUserAttributeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utility.SendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	attribute, err := c.service.CreateUserAttribute(r.Context(), req)
	if err != nil {
		utility.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	utility.SendSuccess(w, http.StatusCreated, "User attribute created successfully", attribute)
}

// UpdateUserAttribute handles PUT /user-attributes/{id}
func (c *UserAttributeController) UpdateUserAttribute(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utility.SendError(w, http.StatusBadRequest, "Invalid user attribute ID")
		return
	}

	var req model.UpdateUserAttributeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utility.SendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	attribute, err := c.service.UpdateUserAttribute(r.Context(), id, req)
	if err != nil {
		utility.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	utility.SendSuccess(w, http.StatusOK, "User attribute updated successfully", attribute)
}

// DeleteUserAttribute handles DELETE /user-attributes/{id}
func (c *UserAttributeController) DeleteUserAttribute(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utility.SendError(w, http.StatusBadRequest, "Invalid user attribute ID")
		return
	}

	err = c.service.DeleteUserAttribute(r.Context(), id)
	if err != nil {
		utility.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	utility.SendSuccess(w, http.StatusOK, "User attribute deleted successfully", nil)
}
