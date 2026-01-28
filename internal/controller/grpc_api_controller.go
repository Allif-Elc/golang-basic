package controller

import (
	"context"
	"encoding/json"
	"golang-basic/api/internal/middleware"
	"golang-basic/api/internal/model"
	"golang-basic/api/internal/utility"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type GrpcAPIServiceInterface interface {
	CreateGrpcAPI(ctx context.Context, userID int64, req model.CreateGrpcAPIRequest) (*model.GrpcAPI, error)
	GetGrpcAPIByID(ctx context.Context, apiID int64) (model.GrpcAPI, error)
	UpdateGrpcAPI(ctx context.Context, apiID int64, req model.UpdateGrpcAPIRequest) (*model.GrpcAPI, error)
	DeleteGrpcAPI(ctx context.Context, apiID int64) error
	ListGrpcAPIs(ctx context.Context, req model.ListGrpcAPIsRequest) (*model.PageResult[model.GrpcAPI], error)
}

type GrpcAPIController struct {
	service GrpcAPIServiceInterface
}

func NewGrpcAPIController(service GrpcAPIServiceInterface) *GrpcAPIController {
	return &GrpcAPIController{service: service}
}

// CreateGrpcAPI handles gRPC API documentation creation
// Per Specs: Authorization handled by ABAC middleware at router level
func (c *GrpcAPIController) CreateGrpcAPI(w http.ResponseWriter, r *http.Request) {
	// Extract userID from context
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		utility.SendErrorResponse(w, utility.UnauthorizedError("User ID not found in context"))
		return
	}

	// Extract projectID from URL
	projectIDStr := chi.URLParam(r, "projectID")
	projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
	if err != nil {
		utility.SendErrorResponse(w, utility.ValidationError("Invalid project ID"))
		return
	}

	// Parse request body
	var req model.CreateGrpcAPIRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utility.SendErrorResponse(w, utility.ValidationError("Invalid request body"))
		return
	}

	// Set project ID from URL
	req.IDProject = projectID

	// Create gRPC API
	grpcAPI, err := c.service.CreateGrpcAPI(r.Context(), userID, req)
	if err != nil {
		utility.SendErrorResponse(w, err)
		return
	}

	utility.SendSuccess(w, http.StatusCreated, "gRPC API created successfully", grpcAPI)
}

// GetGrpcAPIByID retrieves a gRPC API by ID
func (c *GrpcAPIController) GetGrpcAPIByID(w http.ResponseWriter, r *http.Request) {
	// Extract API ID from URL
	apiIDStr := chi.URLParam(r, "id")
	apiID, err := strconv.ParseInt(apiIDStr, 10, 64)
	if err != nil {
		utility.SendErrorResponse(w, utility.ValidationError("Invalid gRPC API ID"))
		return
	}

	// Get gRPC API
	grpcAPI, err := c.service.GetGrpcAPIByID(r.Context(), apiID)
	if err != nil {
		utility.SendErrorResponse(w, err)
		return
	}

	utility.SendSuccess(w, http.StatusOK, "gRPC API retrieved successfully", grpcAPI)
}

// UpdateGrpcAPI handles gRPC API updates
func (c *GrpcAPIController) UpdateGrpcAPI(w http.ResponseWriter, r *http.Request) {
	// Extract API ID from URL
	apiIDStr := chi.URLParam(r, "id")
	apiID, err := strconv.ParseInt(apiIDStr, 10, 64)
	if err != nil {
		utility.SendErrorResponse(w, utility.ValidationError("Invalid gRPC API ID"))
		return
	}

	// Parse request body
	var req model.UpdateGrpcAPIRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utility.SendErrorResponse(w, utility.ValidationError("Invalid request body"))
		return
	}

	// Update gRPC API
	grpcAPI, err := c.service.UpdateGrpcAPI(r.Context(), apiID, req)
	if err != nil {
		utility.SendErrorResponse(w, err)
		return
	}

	utility.SendSuccess(w, http.StatusOK, "gRPC API updated successfully", grpcAPI)
}

// DeleteGrpcAPI handles gRPC API deletion
func (c *GrpcAPIController) DeleteGrpcAPI(w http.ResponseWriter, r *http.Request) {
	// Extract API ID from URL
	apiIDStr := chi.URLParam(r, "id")
	apiID, err := strconv.ParseInt(apiIDStr, 10, 64)
	if err != nil {
		utility.SendErrorResponse(w, utility.ValidationError("Invalid gRPC API ID"))
		return
	}

	// Delete gRPC API
	if err := c.service.DeleteGrpcAPI(r.Context(), apiID); err != nil {
		utility.SendErrorResponse(w, err)
		return
	}

	utility.SendSuccess(w, http.StatusOK, "gRPC API deleted successfully", nil)
}

// ListGrpcAPIs handles listing gRPC APIs with filters
func (c *GrpcAPIController) ListGrpcAPIs(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	req, err := c.parseListRequest(r)
	if err != nil {
		utility.SendErrorResponse(w, err)
		return
	}

	// Get gRPC APIs
	result, err := c.service.ListGrpcAPIs(r.Context(), *req)
	if err != nil {
		utility.SendErrorResponse(w, err)
		return
	}

	utility.SendSuccess(w, http.StatusOK, "gRPC APIs retrieved successfully", result)
}

// ListGrpcAPIsByProject handles listing gRPC APIs for a specific project
func (c *GrpcAPIController) ListGrpcAPIsByProject(w http.ResponseWriter, r *http.Request) {
	// Extract projectID from URL
	projectIDStr := chi.URLParam(r, "projectID")
	projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
	if err != nil {
		utility.SendErrorResponse(w, utility.ValidationError("Invalid project ID"))
		return
	}

	// Parse query parameters
	req, err := c.parseListRequest(r)
	if err != nil {
		utility.SendErrorResponse(w, err)
		return
	}

	// Set project filter
	req.IDProject = &projectID

	// Get gRPC APIs
	result, err := c.service.ListGrpcAPIs(r.Context(), *req)
	if err != nil {
		utility.SendErrorResponse(w, err)
		return
	}

	utility.SendSuccess(w, http.StatusOK, "gRPC APIs retrieved successfully", result)
}

// parseListRequest parses query parameters for list requests
func (c *GrpcAPIController) parseListRequest(r *http.Request) (*model.ListGrpcAPIsRequest, error) {
	req := &model.ListGrpcAPIsRequest{
		Page:  parseIntQuery(r, "page", 1),
		Limit: parseIntQuery(r, "limit", 20),
	}

	// Parse project filter
	if projectIDStr := r.URL.Query().Get("id_project"); projectIDStr != "" {
		projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
		if err != nil {
			return nil, utility.ValidationError("Invalid project ID")
		}
		req.IDProject = &projectID
	}

	// Parse service name filter
	req.ServiceName = r.URL.Query().Get("service_name")

	// Parse search
	req.Search = r.URL.Query().Get("search")

	return req, nil
}
