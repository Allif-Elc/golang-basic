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

type RestAPIServiceInterface interface {
	CreateRestAPI(ctx context.Context, userID int64, req model.CreateRestAPIRequest) (*model.RestAPI, error)
	GetRestAPIByID(ctx context.Context, apiID int64) (model.RestAPI, error)
	UpdateRestAPI(ctx context.Context, apiID int64, req model.UpdateRestAPIRequest) (*model.RestAPI, error)
	DeleteRestAPI(ctx context.Context, apiID int64) error
	ListRestAPIs(ctx context.Context, req model.ListRestAPIsRequest) (*model.PageResult[model.RestAPI], error)
}

type RestAPIController struct {
	service RestAPIServiceInterface
}

func NewRestAPIController(service RestAPIServiceInterface) *RestAPIController {
	return &RestAPIController{service: service}
}

// CreateRestAPI handles REST API documentation creation
func (c *RestAPIController) CreateRestAPI(w http.ResponseWriter, r *http.Request) {
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
	var req model.CreateRestAPIRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utility.SendErrorResponse(w, utility.ValidationError("Invalid request body"))
		return
	}

	// Set project ID from URL
	req.IDProject = projectID

	// Create REST API
	restAPI, err := c.service.CreateRestAPI(r.Context(), userID, req)
	if err != nil {
		utility.SendErrorResponse(w, err)
		return
	}

	utility.SendSuccess(w, http.StatusCreated, "REST API created successfully", restAPI)
}

// GetRestAPIByID retrieves a REST API by ID
func (c *RestAPIController) GetRestAPIByID(w http.ResponseWriter, r *http.Request) {
	// Extract API ID from URL
	apiIDStr := chi.URLParam(r, "id")
	apiID, err := strconv.ParseInt(apiIDStr, 10, 64)
	if err != nil {
		utility.SendErrorResponse(w, utility.ValidationError("Invalid REST API ID"))
		return
	}

	// Get REST API
	restAPI, err := c.service.GetRestAPIByID(r.Context(), apiID)
	if err != nil {
		utility.SendErrorResponse(w, err)
		return
	}

	utility.SendSuccess(w, http.StatusOK, "REST API retrieved successfully", restAPI)
}

// UpdateRestAPI handles REST API updates
func (c *RestAPIController) UpdateRestAPI(w http.ResponseWriter, r *http.Request) {
	// Extract API ID from URL
	apiIDStr := chi.URLParam(r, "id")
	apiID, err := strconv.ParseInt(apiIDStr, 10, 64)
	if err != nil {
		utility.SendErrorResponse(w, utility.ValidationError("Invalid REST API ID"))
		return
	}

	// Parse request body
	var req model.UpdateRestAPIRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utility.SendErrorResponse(w, utility.ValidationError("Invalid request body"))
		return
	}

	// Update REST API
	restAPI, err := c.service.UpdateRestAPI(r.Context(), apiID, req)
	if err != nil {
		utility.SendErrorResponse(w, err)
		return
	}

	utility.SendSuccess(w, http.StatusOK, "REST API updated successfully", restAPI)
}

// DeleteRestAPI handles REST API deletion
func (c *RestAPIController) DeleteRestAPI(w http.ResponseWriter, r *http.Request) {
	// Extract API ID from URL
	apiIDStr := chi.URLParam(r, "id")
	apiID, err := strconv.ParseInt(apiIDStr, 10, 64)
	if err != nil {
		utility.SendErrorResponse(w, utility.ValidationError("Invalid REST API ID"))
		return
	}

	// Delete REST API
	if err := c.service.DeleteRestAPI(r.Context(), apiID); err != nil {
		utility.SendErrorResponse(w, err)
		return
	}

	utility.SendSuccess(w, http.StatusOK, "REST API deleted successfully", nil)
}

// ListRestAPIs handles listing REST APIs with filters
func (c *RestAPIController) ListRestAPIs(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	req, err := c.parseListRequest(r)
	if err != nil {
		utility.SendErrorResponse(w, err)
		return
	}

	// Get REST APIs
	result, err := c.service.ListRestAPIs(r.Context(), *req)
	if err != nil {
		utility.SendErrorResponse(w, err)
		return
	}

	utility.SendSuccess(w, http.StatusOK, "REST APIs retrieved successfully", result)
}

// ListRestAPIsByProject handles listing REST APIs for a specific project
func (c *RestAPIController) ListRestAPIsByProject(w http.ResponseWriter, r *http.Request) {
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

	// Get REST APIs
	result, err := c.service.ListRestAPIs(r.Context(), *req)
	if err != nil {
		utility.SendErrorResponse(w, err)
		return
	}

	utility.SendSuccess(w, http.StatusOK, "REST APIs retrieved successfully", result)
}

// parseListRequest parses query parameters for list requests
func (c *RestAPIController) parseListRequest(r *http.Request) (*model.ListRestAPIsRequest, error) {
	req := &model.ListRestAPIsRequest{
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

	// Parse method filter
	req.Method = r.URL.Query().Get("method")

	// Parse search
	req.Search = r.URL.Query().Get("search")

	return req, nil
}
