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

// PermissionServiceInterface defines the interface for permission service operations
type PermissionServiceInterface interface {
	// Attribute operations
	CreateAttribute(ctx context.Context, req model.CreateAttributesRequest) (*model.Attributes, error)
	GetAttributeByID(ctx context.Context, id int64) (*model.Attributes, error)
	ListAttributes(ctx context.Context) ([]model.Attributes, error)
	UpdateAttribute(ctx context.Context, id int64, req model.UpdateAttributesRequest) (*model.Attributes, error)
	DeleteAttribute(ctx context.Context, id int64) error

	// Resource operations
	CreateResource(ctx context.Context, req model.CreateResoucesRequest) (*model.Resources, error)
	GetResourceByID(ctx context.Context, id int64) (*model.Resources, error)
	ListResources(ctx context.Context) ([]model.Resources, error)
	UpdateResource(ctx context.Context, id int64, req model.UpdateResourcesRequest) (*model.Resources, error)
	DeleteResource(ctx context.Context, id int64) error

	// Permission operations
	CreatePermission(ctx context.Context, req model.CreatePermissionsRequest) (*model.Permissions, error)
	GetPermissionByID(ctx context.Context, id int64) (*model.Permissions, error)
	ListPermissions(ctx context.Context) ([]model.Permissions, error)
	UpdatePermission(ctx context.Context, id int64, req model.UpdatePermissionsRequest) (*model.Permissions, error)
	DeletePermission(ctx context.Context, id int64) error
}

type PermissionController struct {
	service PermissionServiceInterface
}

func NewPermissionController(service PermissionServiceInterface) *PermissionController {
	return &PermissionController{service: service}
}

// ==================== Attribute Handlers ====================

// CreateAttribute handles POST /attributes
func (c *PermissionController) CreateAttribute(w http.ResponseWriter, r *http.Request) {
	var req model.CreateAttributesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utility.SendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	attribute, err := c.service.CreateAttribute(r.Context(), req)
	if err != nil {
		utility.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	utility.SendSuccess(w, http.StatusCreated, "Attribute created successfully", attribute)
}

// GetAttributeByID handles GET /attributes/{id}
func (c *PermissionController) GetAttributeByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utility.SendError(w, http.StatusBadRequest, "Invalid attribute ID")
		return
	}

	attribute, err := c.service.GetAttributeByID(r.Context(), id)
	if err != nil {
		utility.SendError(w, http.StatusNotFound, err.Error())
		return
	}

	utility.SendSuccess(w, http.StatusOK, "Attribute retrieved successfully", attribute)
}

// ListAttributes handles GET /attributes
func (c *PermissionController) ListAttributes(w http.ResponseWriter, r *http.Request) {
	attributes, err := c.service.ListAttributes(r.Context())
	if err != nil {
		utility.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utility.SendSuccess(w, http.StatusOK, "Attributes retrieved successfully", attributes)
}

// DeleteAttribute handles DELETE /attributes/{id}
func (c *PermissionController) DeleteAttribute(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utility.SendError(w, http.StatusBadRequest, "Invalid attribute ID")
		return
	}

	err = c.service.DeleteAttribute(r.Context(), id)
	if err != nil {
		utility.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	utility.SendSuccess(w, http.StatusOK, "Attribute deleted successfully", nil)
}

// UpdateAttribute handles PUT /attributes/{id}
func (c *PermissionController) UpdateAttribute(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utility.SendError(w, http.StatusBadRequest, "Invalid attribute ID")
		return
	}

	var req model.UpdateAttributesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utility.SendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	attribute, err := c.service.UpdateAttribute(r.Context(), id, req)
	if err != nil {
		utility.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	utility.SendSuccess(w, http.StatusOK, "Attribute updated successfully", attribute)
}

// ==================== Resource Handlers ====================

// CreateResource handles POST /resources
func (c *PermissionController) CreateResource(w http.ResponseWriter, r *http.Request) {
	var req model.CreateResoucesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utility.SendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	resource, err := c.service.CreateResource(r.Context(), req)
	if err != nil {
		utility.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	utility.SendSuccess(w, http.StatusCreated, "Resource created successfully", resource)
}

// GetResourceByID handles GET /resources/{id}
func (c *PermissionController) GetResourceByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utility.SendError(w, http.StatusBadRequest, "Invalid resource ID")
		return
	}

	resource, err := c.service.GetResourceByID(r.Context(), id)
	if err != nil {
		utility.SendError(w, http.StatusNotFound, err.Error())
		return
	}

	utility.SendSuccess(w, http.StatusOK, "Resource retrieved successfully", resource)
}

// ListResources handles GET /resources
func (c *PermissionController) ListResources(w http.ResponseWriter, r *http.Request) {
	resources, err := c.service.ListResources(r.Context())
	if err != nil {
		utility.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utility.SendSuccess(w, http.StatusOK, "Resources retrieved successfully", resources)
}

// DeleteResource handles DELETE /resources/{id}
func (c *PermissionController) DeleteResource(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utility.SendError(w, http.StatusBadRequest, "Invalid resource ID")
		return
	}

	err = c.service.DeleteResource(r.Context(), id)
	if err != nil {
		utility.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	utility.SendSuccess(w, http.StatusOK, "Resource deleted successfully", nil)
}

// UpdateResource handles PUT /resources/{id}
func (c *PermissionController) UpdateResource(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utility.SendError(w, http.StatusBadRequest, "Invalid resource ID")
		return
	}

	var req model.UpdateResourcesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utility.SendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	resource, err := c.service.UpdateResource(r.Context(), id, req)
	if err != nil {
		utility.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	utility.SendSuccess(w, http.StatusOK, "Resource updated successfully", resource)
}

// ==================== Permission Handlers ====================

// CreatePermission handles POST /permissions
func (c *PermissionController) CreatePermission(w http.ResponseWriter, r *http.Request) {
	var req model.CreatePermissionsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utility.SendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	permission, err := c.service.CreatePermission(r.Context(), req)
	if err != nil {
		utility.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	utility.SendSuccess(w, http.StatusCreated, "Permission created successfully", permission)
}

// GetPermissionByID handles GET /permissions/{id}
func (c *PermissionController) GetPermissionByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utility.SendError(w, http.StatusBadRequest, "Invalid permission ID")
		return
	}

	permission, err := c.service.GetPermissionByID(r.Context(), id)
	if err != nil {
		utility.SendError(w, http.StatusNotFound, err.Error())
		return
	}

	utility.SendSuccess(w, http.StatusOK, "Permission retrieved successfully", permission)
}

// ListPermissions handles GET /permissions
func (c *PermissionController) ListPermissions(w http.ResponseWriter, r *http.Request) {
	permissions, err := c.service.ListPermissions(r.Context())
	if err != nil {
		utility.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utility.SendSuccess(w, http.StatusOK, "Permissions retrieved successfully", permissions)
}

// DeletePermission handles DELETE /permissions/{id}
func (c *PermissionController) DeletePermission(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utility.SendError(w, http.StatusBadRequest, "Invalid permission ID")
		return
	}

	err = c.service.DeletePermission(r.Context(), id)
	if err != nil {
		utility.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	utility.SendSuccess(w, http.StatusOK, "Permission deleted successfully", nil)
}

// UpdatePermission handles PUT /permissions/{id}
func (c *PermissionController) UpdatePermission(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utility.SendError(w, http.StatusBadRequest, "Invalid permission ID")
		return
	}

	var req model.UpdatePermissionsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utility.SendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	permission, err := c.service.UpdatePermission(r.Context(), id, req)
	if err != nil {
		utility.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	utility.SendSuccess(w, http.StatusOK, "Permission updated successfully", permission)
}
