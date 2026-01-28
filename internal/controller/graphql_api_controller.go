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

type GraphQLAPIServiceInterface interface {
	CreateGraphQLAPI(ctx context.Context, userID int64, req model.CreateGraphQLAPIRequest) (*model.GraphQLAPI, error)
	GetGraphQLAPIByID(ctx context.Context, apiID int64) (model.GraphQLAPI, error)
	UpdateGraphQLAPI(ctx context.Context, apiID int64, req model.UpdateGraphQLAPIRequest) (*model.GraphQLAPI, error)
	DeleteGraphQLAPI(ctx context.Context, apiID int64) error
	ListGraphQLAPIs(ctx context.Context, req model.ListGraphQLAPIsRequest) (*model.PageResult[model.GraphQLAPI], error)
}

type GraphQLAPIController struct {
	service GraphQLAPIServiceInterface
}

func NewGraphQLAPIController(service GraphQLAPIServiceInterface) *GraphQLAPIController {
	return &GraphQLAPIController{service: service}
}

// CreateGraphQLAPI handles GraphQL API documentation creation
// Per Specs: Authorization handled by ABAC middleware at router level
func (c *GraphQLAPIController) CreateGraphQLAPI(w http.ResponseWriter, r *http.Request) {
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
	var req model.CreateGraphQLAPIRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utility.SendErrorResponse(w, utility.ValidationError("Invalid request body"))
		return
	}

	// Set project ID from URL
	req.IDProject = projectID

	// Create GraphQL API
	graphqlAPI, err := c.service.CreateGraphQLAPI(r.Context(), userID, req)
	if err != nil {
		utility.SendErrorResponse(w, err)
		return
	}

	utility.SendSuccess(w, http.StatusCreated, "GraphQL API created successfully", graphqlAPI)
}

// GetGraphQLAPIByID retrieves a GraphQL API by ID
func (c *GraphQLAPIController) GetGraphQLAPIByID(w http.ResponseWriter, r *http.Request) {
	// Extract API ID from URL
	apiIDStr := chi.URLParam(r, "id")
	apiID, err := strconv.ParseInt(apiIDStr, 10, 64)
	if err != nil {
		utility.SendErrorResponse(w, utility.ValidationError("Invalid GraphQL API ID"))
		return
	}

	// Get GraphQL API
	graphqlAPI, err := c.service.GetGraphQLAPIByID(r.Context(), apiID)
	if err != nil {
		utility.SendErrorResponse(w, err)
		return
	}

	utility.SendSuccess(w, http.StatusOK, "GraphQL API retrieved successfully", graphqlAPI)
}

// UpdateGraphQLAPI handles GraphQL API updates
func (c *GraphQLAPIController) UpdateGraphQLAPI(w http.ResponseWriter, r *http.Request) {
	// Extract API ID from URL
	apiIDStr := chi.URLParam(r, "id")
	apiID, err := strconv.ParseInt(apiIDStr, 10, 64)
	if err != nil {
		utility.SendErrorResponse(w, utility.ValidationError("Invalid GraphQL API ID"))
		return
	}

	// Parse request body
	var req model.UpdateGraphQLAPIRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utility.SendErrorResponse(w, utility.ValidationError("Invalid request body"))
		return
	}

	// Update GraphQL API
	graphqlAPI, err := c.service.UpdateGraphQLAPI(r.Context(), apiID, req)
	if err != nil {
		utility.SendErrorResponse(w, err)
		return
	}

	utility.SendSuccess(w, http.StatusOK, "GraphQL API updated successfully", graphqlAPI)
}

// DeleteGraphQLAPI handles GraphQL API deletion
func (c *GraphQLAPIController) DeleteGraphQLAPI(w http.ResponseWriter, r *http.Request) {
	// Extract API ID from URL
	apiIDStr := chi.URLParam(r, "id")
	apiID, err := strconv.ParseInt(apiIDStr, 10, 64)
	if err != nil {
		utility.SendErrorResponse(w, utility.ValidationError("Invalid GraphQL API ID"))
		return
	}

	// Delete GraphQL API
	if err := c.service.DeleteGraphQLAPI(r.Context(), apiID); err != nil {
		utility.SendErrorResponse(w, err)
		return
	}

	utility.SendSuccess(w, http.StatusOK, "GraphQL API deleted successfully", nil)
}

// ListGraphQLAPIs handles listing GraphQL APIs with filters
func (c *GraphQLAPIController) ListGraphQLAPIs(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	req, err := c.parseListRequest(r)
	if err != nil {
		utility.SendErrorResponse(w, err)
		return
	}

	// Get GraphQL APIs
	result, err := c.service.ListGraphQLAPIs(r.Context(), *req)
	if err != nil {
		utility.SendErrorResponse(w, err)
		return
	}

	utility.SendSuccess(w, http.StatusOK, "GraphQL APIs retrieved successfully", result)
}

// ListGraphQLAPIsByProject handles listing GraphQL APIs for a specific project
func (c *GraphQLAPIController) ListGraphQLAPIsByProject(w http.ResponseWriter, r *http.Request) {
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

	// Get GraphQL APIs
	result, err := c.service.ListGraphQLAPIs(r.Context(), *req)
	if err != nil {
		utility.SendErrorResponse(w, err)
		return
	}

	utility.SendSuccess(w, http.StatusOK, "GraphQL APIs retrieved successfully", result)
}

// parseListRequest parses query parameters for list requests
func (c *GraphQLAPIController) parseListRequest(r *http.Request) (*model.ListGraphQLAPIsRequest, error) {
	req := &model.ListGraphQLAPIsRequest{
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

	// Parse type filter (query, mutation, subscription)
	req.Type = r.URL.Query().Get("type")

	// Parse search
	req.Search = r.URL.Query().Get("search")

	return req, nil
}
