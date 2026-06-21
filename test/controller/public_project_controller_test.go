package controller_test

import (
	"context"
	"encoding/json"
	"golang-basic/api/internal/controller"
	"golang-basic/api/internal/model"
	"golang-basic/api/internal/utility"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
)

// MockProjectServicePublic is a mock implementation of ProjectServiceInterface for public endpoints
type MockProjectServicePublic struct {
	GetPublicProjectDocumentationFunc func(ctx context.Context, slug string, page, limit int) (*model.PublicProjectDocumentationResponse, error)
}

func (m *MockProjectServicePublic) CreateProject(ctx context.Context, userID int64, req model.CreateProjectRequest) (*model.Project, error) {
	return nil, nil
}

func (m *MockProjectServicePublic) GetProjectByID(ctx context.Context, projectID int64) (model.Project, error) {
	return model.Project{}, nil
}

func (m *MockProjectServicePublic) UpdateProject(ctx context.Context, projectID int64, req model.UpdateProjectRequest) (*model.Project, error) {
	return nil, nil
}

func (m *MockProjectServicePublic) DeleteProject(ctx context.Context, projectID int64) (map[string]int64, error) {
	return nil, nil
}

func (m *MockProjectServicePublic) GetAPIStats(ctx context.Context, projectID int64) (map[string]int64, error) {
	return nil, nil
}

func (m *MockProjectServicePublic) ListProjects(ctx context.Context, req model.ListProjectsRequest) (*model.PageResult[model.Project], error) {
	return nil, nil
}

func (m *MockProjectServicePublic) ListProjectsWithStats(ctx context.Context, req model.ListProjectsRequest) (*model.PageResult[model.ProjectWithStats], error) {
	return nil, nil
}

func (m *MockProjectServicePublic) GetPublicProjectDocumentation(ctx context.Context, slug string, page, limit int) (*model.PublicProjectDocumentationResponse, error) {
	if m.GetPublicProjectDocumentationFunc != nil {
		return m.GetPublicProjectDocumentationFunc(ctx, slug, page, limit)
	}
	return &model.PublicProjectDocumentationResponse{}, nil
}

// setupPublicProjectTestRouter creates a Chi router with public project routes for testing
func setupPublicProjectTestRouter(projectCtrl *controller.ProjectController) *chi.Mux {
	r := chi.NewRouter()
	r.Route("/api/v1/public/projects", func(r chi.Router) {
		r.Get("/{slug}/full", projectCtrl.GetPublicProjectDocumentation)
	})
	return r
}

// ==================== TestGetPublicProjectDocumentation Tests ====================

func TestGetPublicProjectDocumentation_Success(t *testing.T) {
	mockService := &MockProjectServicePublic{
		GetPublicProjectDocumentationFunc: func(ctx context.Context, slug string, page, limit int) (*model.PublicProjectDocumentationResponse, error) {
			return &model.PublicProjectDocumentationResponse{
				Project: model.Project{
					IDProject:   1,
					IDUser:      1,
					Name:        "Test Project",
					Slug:        "test-project",
					Description: "A test project",
					Version:     "1.0.0",
					IsPublic:    true,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				},
				Rest: model.PaginatedAPIResponse[model.RestAPI]{
					Data:  []model.RestAPI{{IDRestAPI: 1, Name: "GET /api/users", Method: "GET", Endpoint: "/api/users"}},
					Page:  1,
					Limit: 20,
					Total: 1,
				},
				GraphQL: model.PaginatedAPIResponse[model.GraphQLAPI]{
					Data:  []model.GraphQLAPI{{IDGraphqlAPI: 1, Name: "getUser", Type: "query", ReturnType: "User"}},
					Page:  1,
					Limit: 20,
					Total: 1,
				},
				Grpc: model.PaginatedAPIResponse[model.GrpcAPI]{
					Data:  []model.GrpcAPI{{IDGrpcAPI: 1, ServiceName: "UserService", MethodName: "GetUser"}},
					Page:  1,
					Limit: 20,
					Total: 1,
				},
			}, nil
		},
	}

	ctrl := controller.NewProjectController(mockService)
	r := setupPublicProjectTestRouter(ctrl)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/public/projects/test-project/full", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	if response["status"] != "success" {
		t.Errorf("Expected status 'success', got '%s'", response["status"])
	}
}

func TestGetPublicProjectDocumentation_WithPagination(t *testing.T) {
	mockService := &MockProjectServicePublic{
		GetPublicProjectDocumentationFunc: func(ctx context.Context, slug string, page, limit int) (*model.PublicProjectDocumentationResponse, error) {
			if page != 2 || limit != 50 {
				t.Errorf("Expected page=2, limit=50, got page=%d, limit=%d", page, limit)
			}
			return &model.PublicProjectDocumentationResponse{
				Project: model.Project{
					IDProject: 1,
					Name:     "Test Project",
					Slug:     "test-project",
					IsPublic: true,
				},
				Rest: model.PaginatedAPIResponse[model.RestAPI]{
					Data:  []model.RestAPI{{IDRestAPI: 21, Name: "API 21", Method: "GET", Endpoint: "/api/21"}},
					Page:  2,
					Limit: 50,
					Total: 100,
				},
				GraphQL: model.PaginatedAPIResponse[model.GraphQLAPI]{Data: []model.GraphQLAPI{}, Page: 2, Limit: 50, Total: 0},
				Grpc:    model.PaginatedAPIResponse[model.GrpcAPI]{Data: []model.GrpcAPI{}, Page: 2, Limit: 50, Total: 0},
			}, nil
		},
	}

	ctrl := controller.NewProjectController(mockService)
	r := setupPublicProjectTestRouter(ctrl)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/public/projects/test-project/full?page=2&limit=50", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	data := response["data"].(map[string]interface{})
	rest := data["rest"].(map[string]interface{})
	if rest["page"].(float64) != 2 {
		t.Errorf("Expected page 2, got %v", rest["page"])
	}
	if rest["limit"].(float64) != 50 {
		t.Errorf("Expected limit 50, got %v", rest["limit"])
	}
}

func TestGetPublicProjectDocumentation_UnknownSlug(t *testing.T) {
	mockService := &MockProjectServicePublic{
		GetPublicProjectDocumentationFunc: func(ctx context.Context, slug string, page, limit int) (*model.PublicProjectDocumentationResponse, error) {
			return nil, utility.NotFoundError("project not found")
		},
	}

	ctrl := controller.NewProjectController(mockService)
	r := setupPublicProjectTestRouter(ctrl)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/public/projects/unknown-slug/full", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	if response["message"] != "Project not found" {
		t.Errorf("Expected message 'Project not found', got '%s'", response["message"])
	}
}

func TestGetPublicProjectDocumentation_PrivateSlug(t *testing.T) {
	mockService := &MockProjectServicePublic{
		GetPublicProjectDocumentationFunc: func(ctx context.Context, slug string, page, limit int) (*model.PublicProjectDocumentationResponse, error) {
			return nil, utility.NotFoundError("project not found")
		},
	}

	ctrl := controller.NewProjectController(mockService)
	r := setupPublicProjectTestRouter(ctrl)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/public/projects/private-project/full", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestGetPublicProjectDocumentation_JSONBSerialization(t *testing.T) {
	mockService := &MockProjectServicePublic{
		GetPublicProjectDocumentationFunc: func(ctx context.Context, slug string, page, limit int) (*model.PublicProjectDocumentationResponse, error) {
			// Create API with JSONB fields (headers, params, etc.)
			headersJSON := []byte(`[{"name":"Authorization","required":true}]`)
			return &model.PublicProjectDocumentationResponse{
				Project: model.Project{
					IDProject: 1,
					Name:     "Test Project",
					Slug:     "test-project",
					IsPublic: true,
				},
				Rest: model.PaginatedAPIResponse[model.RestAPI]{
					Data: []model.RestAPI{
						{
							IDRestAPI:  1,
							Name:       "GET /api/users",
							Method:     "GET",
							Endpoint:   "/api/users",
							Headers:    headersJSON,
							PathParams:  []byte(`[]`),
							QueryParams: []byte(`[]`),
							RequestBody: []byte(`null`),
							Responses:  []byte(`{"200":{"status_code":200,"description":"Success"}}`),
						},
					},
					Page:  1,
					Limit: 20,
					Total: 1,
				},
				GraphQL: model.PaginatedAPIResponse[model.GraphQLAPI]{Data: []model.GraphQLAPI{}, Page: 1, Limit: 20, Total: 0},
				Grpc:    model.PaginatedAPIResponse[model.GrpcAPI]{Data: []model.GrpcAPI{}, Page: 1, Limit: 20, Total: 0},
			}, nil
		},
	}

	ctrl := controller.NewProjectController(mockService)
	r := setupPublicProjectTestRouter(ctrl)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/public/projects/test-project/full", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	data := response["data"].(map[string]interface{})
	rest := data["rest"].(map[string]interface{})
	restData := rest["data"].([]interface{})

	if len(restData) == 0 {
		t.Fatal("Expected at least one REST API")
	}

	api := restData[0].(map[string]interface{})
	headers := api["headers"]
	// Verify headers is a JSON array, not a base64 string
	if _, ok := headers.([]interface{}); !ok {
		t.Errorf("Expected headers to be a JSON array, got %T: %v", headers, headers)
	}
}

func TestGetPublicProjectDocumentation_PaginationMetadata(t *testing.T) {
	mockService := &MockProjectServicePublic{
		GetPublicProjectDocumentationFunc: func(ctx context.Context, slug string, page, limit int) (*model.PublicProjectDocumentationResponse, error) {
			return &model.PublicProjectDocumentationResponse{
				Project: model.Project{
					IDProject: 1,
					Name:     "Test Project",
					Slug:     "test-project",
					IsPublic: true,
				},
				Rest: model.PaginatedAPIResponse[model.RestAPI]{
					Data:  make([]model.RestAPI, 20),
					Page:  1,
					Limit: 20,
					Total: 55,
				},
				GraphQL: model.PaginatedAPIResponse[model.GraphQLAPI]{Data: []model.GraphQLAPI{}, Page: 1, Limit: 20, Total: 0},
				Grpc:    model.PaginatedAPIResponse[model.GrpcAPI]{Data: []model.GrpcAPI{}, Page: 1, Limit: 20, Total: 0},
			}, nil
		},
	}

	ctrl := controller.NewProjectController(mockService)
	r := setupPublicProjectTestRouter(ctrl)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/public/projects/test-project/full", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	data := response["data"].(map[string]interface{})
	rest := data["rest"].(map[string]interface{})

	// Verify pagination metadata
	if rest["page"].(float64) != 1 {
		t.Errorf("Expected page 1, got %v", rest["page"])
	}
	if rest["limit"].(float64) != 20 {
		t.Errorf("Expected limit 20, got %v", rest["limit"])
	}
	if rest["total"].(float64) != 55 {
		t.Errorf("Expected total 55, got %v", rest["total"])
	}
	restData := rest["data"].([]interface{})
	if len(restData) != 20 {
		t.Errorf("Expected 20 REST APIs in data array, got %d", len(restData))
	}
}
