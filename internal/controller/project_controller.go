package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"golang-basic/api/internal/model"
	"golang-basic/api/internal/utility"
	"golang-basic/api/internal/middleware"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type ProjectServiceInterface interface {
	CreateProject(ctx context.Context, userID int64, req model.CreateProjectRequest) (*model.Project, error)
	GetProjectByID(ctx context.Context, projectID int64) (model.Project, error)
	UpdateProject(ctx context.Context, projectID int64, req model.UpdateProjectRequest) (*model.Project, error)
	DeleteProject(ctx context.Context, projectID int64) (map[string]int64, error)
	GetAPIStats(ctx context.Context, projectID int64) (map[string]int64, error)
	ListProjects(ctx context.Context, req model.ListProjectsRequest) (*model.PageResult[model.Project], error)
	ListProjectsWithStats(ctx context.Context, req model.ListProjectsRequest) (*model.PageResult[model.ProjectWithStats], error)
	GetPublicProjectDocumentation(ctx context.Context, slug string, page, limit int) (*model.PublicProjectDocumentationResponse, error)
}

type ProjectController struct {
	service ProjectServiceInterface
}

func NewProjectController(service ProjectServiceInterface) *ProjectController {
	return &ProjectController{service: service}
}

// CreateProject handles project creation
func (c *ProjectController) CreateProject(w http.ResponseWriter, r *http.Request) {
	// Extract userID from context
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		utility.SendErrorResponse(w, utility.UnauthorizedError("User ID not found in context"))
		return
	}

	// Parse request body
	var req model.CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utility.SendErrorResponse(w, utility.ValidationError("Invalid request body"))
		return
	}

	// Create project
	project, err := c.service.CreateProject(r.Context(), userID, req)
	if err != nil {
		utility.SendErrorResponse(w, err)
		return
	}

	utility.SendSuccess(w, http.StatusCreated, "Project created successfully", project)
}

// GetProjectByID retrieves a project by ID
func (c *ProjectController) GetProjectByID(w http.ResponseWriter, r *http.Request) {
	// Extract project ID from URL
	projectIDStr := chi.URLParam(r, "id")
	projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
	if err != nil {
		utility.SendErrorResponse(w, utility.ValidationError("Invalid project ID"))
		return
	}

	// Get project
	project, err := c.service.GetProjectByID(r.Context(), projectID)
	if err != nil {
		utility.SendErrorResponse(w, err)
		return
	}

	utility.SendSuccess(w, http.StatusOK, "Project retrieved successfully", project)
}

// UpdateProject handles project updates
func (c *ProjectController) UpdateProject(w http.ResponseWriter, r *http.Request) {
	// Extract project ID from URL
	projectIDStr := chi.URLParam(r, "id")
	projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
	if err != nil {
		utility.SendErrorResponse(w, utility.ValidationError("Invalid project ID"))
		return
	}

	// Parse request body
	var req model.UpdateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utility.SendErrorResponse(w, utility.ValidationError("Invalid request body"))
		return
	}

	// Update project
	project, err := c.service.UpdateProject(r.Context(), projectID, req)
	if err != nil {
		utility.SendErrorResponse(w, err)
		return
	}

	utility.SendSuccess(w, http.StatusOK, "Project updated successfully", project)
}

// DeleteProject handles project deletion
func (c *ProjectController) DeleteProject(w http.ResponseWriter, r *http.Request) {
	// Extract project ID from URL
	projectIDStr := chi.URLParam(r, "id")
	projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
	if err != nil {
		utility.SendErrorResponse(w, utility.ValidationError("Invalid project ID"))
		return
	}

	// Delete project and get deleted counts
	deletedCounts, err := c.service.DeleteProject(r.Context(), projectID)
	if err != nil {
		utility.SendErrorResponse(w, err)
		return
	}

	utility.SendSuccess(w, http.StatusOK,
		fmt.Sprintf("Project deleted successfully with %d REST, %d GraphQL, %d gRPC APIs",
			deletedCounts["rest"], deletedCounts["graphql"], deletedCounts["grpc"]),
		deletedCounts)
}

// GetProjectAPIStats retrieves API statistics for a project
func (c *ProjectController) GetProjectAPIStats(w http.ResponseWriter, r *http.Request) {
	// Extract project ID from URL
	projectIDStr := chi.URLParam(r, "id")
	projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
	if err != nil {
		utility.SendErrorResponse(w, utility.ValidationError("Invalid project ID"))
		return
	}

	// Get API stats
	stats, err := c.service.GetAPIStats(r.Context(), projectID)
	if err != nil {
		utility.SendErrorResponse(w, utility.NotFoundError("Project not found"))
		return
	}

	utility.SendSuccess(w, http.StatusOK, "API stats retrieved successfully", stats)
}

// ListProjects handles project listing with pagination and filters
func (c *ProjectController) ListProjects(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	req := model.ListProjectsRequest{
		Page: parseIntQuery(r, "page", 1),
		Limit: parseIntQuery(r, "limit", 10),
		Search: r.URL.Query().Get("search"),
		SortBy: r.URL.Query().Get("sort_by"),
		SortOrder: r.URL.Query().Get("sort_order"),
	}

	// Parse optional boolean parameter
	if isPublicStr := r.URL.Query().Get("is_public"); isPublicStr != "" {
		if isPublic, err := strconv.ParseBool(isPublicStr); err == nil {
			req.IsPublic = &isPublic
		}
	}

	// Parse optional user ID parameter
	if userIDStr := r.URL.Query().Get("id_user"); userIDStr != "" {
		if userID, err := strconv.ParseInt(userIDStr, 10, 64); err == nil {
			req.IDUser = &userID
		}
	}

	// List projects
	result, err := c.service.ListProjects(r.Context(), req)
	if err != nil {
		utility.SendErrorResponse(w, utility.InternalError("Failed to retrieve projects"))
		return
	}

	utility.SendSuccess(w, http.StatusOK, "Projects retrieved successfully", result)
}

// ListProjectsWithStats handles project listing with API statistics in a single optimized query
// This endpoint should be used instead of ListProjects for better performance
func (c *ProjectController) ListProjectsWithStats(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	req := model.ListProjectsRequest{
		Page:      parseIntQuery(r, "page", 1),
		Limit:     parseIntQuery(r, "limit", 10),
		Search:    r.URL.Query().Get("search"),
		SortBy:    r.URL.Query().Get("sort_by"),
		SortOrder: r.URL.Query().Get("sort_order"),
	}

	// Parse optional boolean parameter
	if isPublicStr := r.URL.Query().Get("is_public"); isPublicStr != "" {
		if isPublic, err := strconv.ParseBool(isPublicStr); err == nil {
			req.IsPublic = &isPublic
		}
	}

	// Parse optional user ID parameter
	if userIDStr := r.URL.Query().Get("id_user"); userIDStr != "" {
		if userID, err := strconv.ParseInt(userIDStr, 10, 64); err == nil {
			req.IDUser = &userID
		}
	}

	// List projects with stats in single query
	result, err := c.service.ListProjectsWithStats(r.Context(), req)
	if err != nil {
		utility.SendErrorResponse(w, utility.InternalError("Failed to retrieve projects with stats"))
		return
	}

	utility.SendSuccess(w, http.StatusOK, "Projects with stats retrieved successfully", result)
}

// parseIntQuery parses an integer query parameter with a default value
func parseIntQuery(r *http.Request, key string, defaultValue int) int {
	if valueStr := r.URL.Query().Get(key); valueStr != "" {
		if value, err := strconv.Atoi(valueStr); err == nil {
			return value
		}
	}
	return defaultValue
}

// GetPublicProjectDocumentation retrieves full documentation for a public project
// This is a public endpoint (no authentication required)
func (c *ProjectController) GetPublicProjectDocumentation(w http.ResponseWriter, r *http.Request) {
	// Extract slug from URL
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		utility.SendErrorResponse(w, utility.ValidationError("Invalid slug"))
		return
	}

	// Parse pagination parameters
	page := parseIntQuery(r, "page", 1)
	limit := parseIntQuery(r, "limit", 20)

	// Enforce max limit for performance
	if limit > 100 {
		limit = 100
	}

	// Get public project documentation
	response, err := c.service.GetPublicProjectDocumentation(r.Context(), slug, page, limit)
	if err != nil {
		// Check if it's a not found error
		if utility.IsNotFoundError(err) {
			utility.SendErrorResponse(w, utility.NotFoundError("Project not found"))
			return
		}
		utility.SendErrorResponse(w, utility.InternalError("Failed to retrieve documentation"))
		return
	}

	utility.SendSuccess(w, http.StatusOK, "Documentation retrieved", response)
}
