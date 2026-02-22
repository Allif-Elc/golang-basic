package controller_test

import (
	"bytes"
	"context"
	"encoding/json"
	"golang-basic/api/internal/controller"
	appmiddleware "golang-basic/api/internal/middleware"
	"golang-basic/api/internal/model"
	"golang-basic/api/internal/utility"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
)

// MockRestAPIService is a mock implementation of RestAPIServiceInterface
type MockRestAPIService struct {
	CreateRestAPIFunc    func(ctx context.Context, userID int64, req model.CreateRestAPIRequest) (*model.RestAPI, error)
	GetRestAPIByIDFunc   func(ctx context.Context, apiID int64) (model.RestAPI, error)
	UpdateRestAPIFunc    func(ctx context.Context, apiID int64, req model.UpdateRestAPIRequest) (*model.RestAPI, error)
	DeleteRestAPIFunc    func(ctx context.Context, apiID int64) error
	ListRestAPIsFunc     func(ctx context.Context, req model.ListRestAPIsRequest) (*model.PageResult[model.RestAPI], error)
}

func (m *MockRestAPIService) CreateRestAPI(ctx context.Context, userID int64, req model.CreateRestAPIRequest) (*model.RestAPI, error) {
	if m.CreateRestAPIFunc != nil {
		return m.CreateRestAPIFunc(ctx, userID, req)
	}
	return &model.RestAPI{}, nil
}

func (m *MockRestAPIService) GetRestAPIByID(ctx context.Context, apiID int64) (model.RestAPI, error) {
	if m.GetRestAPIByIDFunc != nil {
		return m.GetRestAPIByIDFunc(ctx, apiID)
	}
	return model.RestAPI{}, nil
}

func (m *MockRestAPIService) UpdateRestAPI(ctx context.Context, apiID int64, req model.UpdateRestAPIRequest) (*model.RestAPI, error) {
	if m.UpdateRestAPIFunc != nil {
		return m.UpdateRestAPIFunc(ctx, apiID, req)
	}
	return &model.RestAPI{}, nil
}

func (m *MockRestAPIService) DeleteRestAPI(ctx context.Context, apiID int64) error {
	if m.DeleteRestAPIFunc != nil {
		return m.DeleteRestAPIFunc(ctx, apiID)
	}
	return nil
}

func (m *MockRestAPIService) ListRestAPIs(ctx context.Context, req model.ListRestAPIsRequest) (*model.PageResult[model.RestAPI], error) {
	if m.ListRestAPIsFunc != nil {
		return m.ListRestAPIsFunc(ctx, req)
	}
	return &model.PageResult[model.RestAPI]{}, nil
}

// setupTestRouter creates a Chi router with REST API routes for testing
func setupRestAPITestRouter(restCtrl *controller.RestAPIController) *chi.Mux {
	r := chi.NewRouter()
	r.Route("/api/v1/projects", func(r chi.Router) {
		r.Route("/{projectID}", func(r chi.Router) {
			r.Post("/rest-apis", restCtrl.CreateRestAPI)
			r.Get("/rest-apis", restCtrl.ListRestAPIsByProject)
		})
	})
	r.Route("/api/v1/rest-apis", func(r chi.Router) {
		r.Get("/{id}", restCtrl.GetRestAPIByID)
		r.Put("/{id}", restCtrl.UpdateRestAPI)
		r.Delete("/{id}", restCtrl.DeleteRestAPI)
		r.Get("/", restCtrl.ListRestAPIs)
	})
	return r
}

// Helper to set authenticated user context
func withAuthenticatedUser(r *http.Request, userID int64) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), appmiddleware.UserIDKey, userID))
}

// ==================== CreateRestAPI Tests ====================

func TestCreateRestAPI_Success(t *testing.T) {
	mockService := &MockRestAPIService{
		CreateRestAPIFunc: func(ctx context.Context, userID int64, req model.CreateRestAPIRequest) (*model.RestAPI, error) {
			return &model.RestAPI{
				IDRestAPI: 1,
				IDProject: req.IDProject,
				IDUser:    userID,
				Name:      req.Name,
				Method:    req.Method,
				Endpoint:  req.Endpoint,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}, nil
		},
	}

	ctrl := controller.NewRestAPIController(mockService)
	r := setupRestAPITestRouter(ctrl)

	reqBody := model.CreateRestAPIRequest{
		Name:     "Get User Profile",
		Method:   "GET",
		Endpoint: "/api/users/{id}",
		Headers: []model.Header{
			{Name: "Authorization", Required: true},
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/1/rest-apis", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req = withAuthenticatedUser(req, 1)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	if response["status"] != "success" {
		t.Errorf("Expected status 'success', got '%s'", response["status"])
	}

	if response["message"] != "REST API created successfully" {
		t.Errorf("Expected message 'REST API created successfully', got '%s'", response["message"])
	}
}

func TestCreateRestAPI_InvalidProjectID(t *testing.T) {
	mockService := &MockRestAPIService{}
	ctrl := controller.NewRestAPIController(mockService)
	r := setupRestAPITestRouter(ctrl)

	reqBody := model.CreateRestAPIRequest{
		Name:     "Test API",
		Method:   "GET",
		Endpoint: "/test",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/invalid/rest-apis", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req = withAuthenticatedUser(req, 1)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	if response["message"] != "Invalid project ID" {
		t.Errorf("Expected message 'Invalid project ID', got '%s'", response["message"])
	}
}

func TestCreateRestAPI_InvalidRequestBody(t *testing.T) {
	mockService := &MockRestAPIService{}
	ctrl := controller.NewRestAPIController(mockService)
	r := setupRestAPITestRouter(ctrl)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/1/rest-apis", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	req = withAuthenticatedUser(req, 1)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	if response["message"] != "Invalid request body" {
		t.Errorf("Expected message 'Invalid request body', got '%s'", response["message"])
	}
}

func TestCreateRestAPI_UnauthorizedUser(t *testing.T) {
	mockService := &MockRestAPIService{}
	ctrl := controller.NewRestAPIController(mockService)
	r := setupRestAPITestRouter(ctrl)

	reqBody := model.CreateRestAPIRequest{
		Name:     "Test API",
		Method:   "GET",
		Endpoint: "/test",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/1/rest-apis", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	// No user ID set in context

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestCreateRestAPI_ServiceValidationError(t *testing.T) {
	mockService := &MockRestAPIService{
		CreateRestAPIFunc: func(ctx context.Context, userID int64, req model.CreateRestAPIRequest) (*model.RestAPI, error) {
			return nil, utility.ValidationError("name must be at least 3 characters")
		},
	}

	ctrl := controller.NewRestAPIController(mockService)
	r := setupRestAPITestRouter(ctrl)

	reqBody := model.CreateRestAPIRequest{
		Name:     "AB",
		Method:   "GET",
		Endpoint: "/test",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/1/rest-apis", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req = withAuthenticatedUser(req, 1)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	if response["status"] != "error" {
		t.Errorf("Expected status 'error', got '%s'", response["status"])
	}
}

func TestCreateRestAPI_ServiceNotFoundError(t *testing.T) {
	mockService := &MockRestAPIService{
		CreateRestAPIFunc: func(ctx context.Context, userID int64, req model.CreateRestAPIRequest) (*model.RestAPI, error) {
			return nil, utility.NotFoundError("project not found")
		},
	}

	ctrl := controller.NewRestAPIController(mockService)
	r := setupRestAPITestRouter(ctrl)

	reqBody := model.CreateRestAPIRequest{
		Name:     "Test API",
		Method:   "GET",
		Endpoint: "/test",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/999/rest-apis", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req = withAuthenticatedUser(req, 1)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestCreateRestAPI_WithRequestBodySchema(t *testing.T) {
	mockService := &MockRestAPIService{
		CreateRestAPIFunc: func(ctx context.Context, userID int64, req model.CreateRestAPIRequest) (*model.RestAPI, error) {
			// Marshal request body to JSON for storage
			requestBodyBytes := []byte("{}")
			if req.RequestBody != nil {
				requestBodyBytes, _ = json.Marshal(req.RequestBody)
			}

			return &model.RestAPI{
				IDRestAPI:   1,
				IDProject:   req.IDProject,
				IDUser:      userID,
				Name:        req.Name,
				Method:      req.Method,
				Endpoint:    req.Endpoint,
				RequestBody: requestBodyBytes,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}, nil
		},
	}

	ctrl := controller.NewRestAPIController(mockService)
	r := setupRestAPITestRouter(ctrl)

	requestBodySchema := &model.JSONSchema{
		Type: "object",
		Properties: map[string]model.Property{
			"email": {
				Type: "string",
			},
			"password": {
				Type: "string",
			},
		},
		Required: []string{"email", "password"},
	}

	reqBody := model.CreateRestAPIRequest{
		Name:        "Create User",
		Method:      "POST",
		Endpoint:    "/users",
		RequestBody: requestBodySchema,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/1/rest-apis", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req = withAuthenticatedUser(req, 1)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestCreateRestAPI_WithHeadersAndParameters(t *testing.T) {
	mockService := &MockRestAPIService{
		CreateRestAPIFunc: func(ctx context.Context, userID int64, req model.CreateRestAPIRequest) (*model.RestAPI, error) {
			return &model.RestAPI{
				IDRestAPI: 1,
				IDProject: req.IDProject,
				IDUser:    userID,
				Name:      req.Name,
				Method:    req.Method,
				Endpoint:  req.Endpoint,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}, nil
		},
	}

	ctrl := controller.NewRestAPIController(mockService)
	r := setupRestAPITestRouter(ctrl)

	reqBody := model.CreateRestAPIRequest{
		Name:     "Update User",
		Method:   "PUT",
		Endpoint: "/users/{id}",
		Headers: []model.Header{
			{Name: "Authorization", Required: true, Description: "Bearer token"},
			{Name: "Content-Type", Required: false, Example: "application/json"},
		},
		PathParams: []model.Parameter{
			{Name: "id", Type: "integer", Required: true, Description: "User ID"},
		},
		QueryParams: []model.Parameter{
			{Name: "include", Type: "string", Required: false},
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/1/rest-apis", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req = withAuthenticatedUser(req, 1)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestCreateRestAPI_WithResponses(t *testing.T) {
	mockService := &MockRestAPIService{
		CreateRestAPIFunc: func(ctx context.Context, userID int64, req model.CreateRestAPIRequest) (*model.RestAPI, error) {
			return &model.RestAPI{
				IDRestAPI: 1,
				IDProject: req.IDProject,
				IDUser:    userID,
				Name:      req.Name,
				Method:    req.Method,
				Endpoint:  req.Endpoint,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}, nil
		},
	}

	ctrl := controller.NewRestAPIController(mockService)
	r := setupRestAPITestRouter(ctrl)

	responses := map[int]model.ResponseExample{
		200: {
			StatusCode:  200,
			Description: "Success",
			Body:        map[string]interface{}{"id": 1, "name": "John Doe"},
		},
		404: {
			StatusCode:  404,
			Description: "User not found",
			Body:        map[string]interface{}{"error": "User not found"},
		},
	}

	reqBody := model.CreateRestAPIRequest{
		Name:      "Get User",
		Method:    "GET",
		Endpoint:  "/users/{id}",
		Responses: responses,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/1/rest-apis", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req = withAuthenticatedUser(req, 1)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}
}

// ==================== GetRestAPIByID Tests ====================

func TestGetRestAPIByID_Success(t *testing.T) {
	mockService := &MockRestAPIService{
		GetRestAPIByIDFunc: func(ctx context.Context, apiID int64) (model.RestAPI, error) {
			return model.RestAPI{
				IDRestAPI: apiID,
				IDProject: 1,
				IDUser:    1,
				Name:      "Test API",
				Method:    "GET",
				Endpoint:  "/test",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}, nil
		},
	}

	ctrl := controller.NewRestAPIController(mockService)
	r := setupRestAPITestRouter(ctrl)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rest-apis/1", nil)
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

func TestGetRestAPIByID_InvalidID(t *testing.T) {
	mockService := &MockRestAPIService{}
	ctrl := controller.NewRestAPIController(mockService)
	r := setupRestAPITestRouter(ctrl)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rest-apis/invalid", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	if response["message"] != "Invalid REST API ID" {
		t.Errorf("Expected message 'Invalid REST API ID', got '%s'", response["message"])
	}
}

func TestGetRestAPIByID_NotFound(t *testing.T) {
	mockService := &MockRestAPIService{
		GetRestAPIByIDFunc: func(ctx context.Context, apiID int64) (model.RestAPI, error) {
			return model.RestAPI{}, utility.NotFoundError("REST API not found")
		},
	}

	ctrl := controller.NewRestAPIController(mockService)
	r := setupRestAPITestRouter(ctrl)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rest-apis/999", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

// ==================== UpdateRestAPI Tests ====================

func TestUpdateRestAPI_Success(t *testing.T) {
	mockService := &MockRestAPIService{
		UpdateRestAPIFunc: func(ctx context.Context, apiID int64, req model.UpdateRestAPIRequest) (*model.RestAPI, error) {
			updatedName := "Updated API Name"
			return &model.RestAPI{
				IDRestAPI: apiID,
				Name:      updatedName,
				Method:    "GET",
				Endpoint:  "/updated-endpoint",
				UpdatedAt: time.Now(),
			}, nil
		},
	}

	ctrl := controller.NewRestAPIController(mockService)
	r := setupRestAPITestRouter(ctrl)

	newName := "Updated API Name"
	reqBody := model.UpdateRestAPIRequest{
		Name:     &newName,
		Endpoint: strPtr("/updated-endpoint"),
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/rest-apis/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

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

func TestUpdateRestAPI_InvalidID(t *testing.T) {
	mockService := &MockRestAPIService{}
	ctrl := controller.NewRestAPIController(mockService)
	r := setupRestAPITestRouter(ctrl)

	reqBody := model.UpdateRestAPIRequest{
		Name: strPtr("Updated Name"),
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/rest-apis/invalid", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestUpdateRestAPI_InvalidRequestBody(t *testing.T) {
	mockService := &MockRestAPIService{}
	ctrl := controller.NewRestAPIController(mockService)
	r := setupRestAPITestRouter(ctrl)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/rest-apis/1", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestUpdateRestAPI_NotFound(t *testing.T) {
	mockService := &MockRestAPIService{
		UpdateRestAPIFunc: func(ctx context.Context, apiID int64, req model.UpdateRestAPIRequest) (*model.RestAPI, error) {
			return nil, utility.NotFoundError("REST API not found")
		},
	}

	ctrl := controller.NewRestAPIController(mockService)
	r := setupRestAPITestRouter(ctrl)

	newName := "Updated Name"
	reqBody := model.UpdateRestAPIRequest{
		Name: &newName,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/rest-apis/999", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

// ==================== DeleteRestAPI Tests ====================

func TestDeleteRestAPI_Success(t *testing.T) {
	mockService := &MockRestAPIService{
		DeleteRestAPIFunc: func(ctx context.Context, apiID int64) error {
			return nil
		},
	}

	ctrl := controller.NewRestAPIController(mockService)
	r := setupRestAPITestRouter(ctrl)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/rest-apis/1", nil)
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

	if response["message"] != "REST API deleted successfully" {
		t.Errorf("Expected message 'REST API deleted successfully', got '%s'", response["message"])
	}
}

func TestDeleteRestAPI_InvalidID(t *testing.T) {
	mockService := &MockRestAPIService{}
	ctrl := controller.NewRestAPIController(mockService)
	r := setupRestAPITestRouter(ctrl)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/rest-apis/invalid", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestDeleteRestAPI_NotFound(t *testing.T) {
	mockService := &MockRestAPIService{
		DeleteRestAPIFunc: func(ctx context.Context, apiID int64) error {
			return utility.NotFoundError("REST API not found")
		},
	}

	ctrl := controller.NewRestAPIController(mockService)
	r := setupRestAPITestRouter(ctrl)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/rest-apis/999", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

// ==================== ListRestAPIs Tests ====================

func TestListRestAPIs_Success(t *testing.T) {
	mockAPIs := []model.RestAPI{
		{IDRestAPI: 1, Name: "API 1", Method: "GET", Endpoint: "/api1"},
		{IDRestAPI: 2, Name: "API 2", Method: "POST", Endpoint: "/api2"},
	}

	mockService := &MockRestAPIService{
		ListRestAPIsFunc: func(ctx context.Context, req model.ListRestAPIsRequest) (*model.PageResult[model.RestAPI], error) {
			return &model.PageResult[model.RestAPI]{
				Data:     mockAPIs,
				Page:     1,
				Size:     20,
				StartRow: 0,
				EndRow:   2,
			}, nil
		},
	}

	ctrl := controller.NewRestAPIController(mockService)
	r := setupRestAPITestRouter(ctrl)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rest-apis?page=1&limit=20", nil)
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

func TestListRestAPIsByProject_Success(t *testing.T) {
	mockAPIs := []model.RestAPI{
		{IDRestAPI: 1, IDProject: 1, Name: "Project API 1", Method: "GET", Endpoint: "/p1/api1"},
	}

	mockService := &MockRestAPIService{
		ListRestAPIsFunc: func(ctx context.Context, req model.ListRestAPIsRequest) (*model.PageResult[model.RestAPI], error) {
			return &model.PageResult[model.RestAPI]{
				Data:     mockAPIs,
				Page:     1,
				Size:     20,
				StartRow: 0,
				EndRow:   1,
			}, nil
		},
	}

	ctrl := controller.NewRestAPIController(mockService)
	r := setupRestAPITestRouter(ctrl)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/1/rest-apis", nil)
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

func TestListRestAPIsByProject_InvalidProjectID(t *testing.T) {
	mockService := &MockRestAPIService{}
	ctrl := controller.NewRestAPIController(mockService)
	r := setupRestAPITestRouter(ctrl)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/invalid/rest-apis", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestListRestAPIs_WithFilters(t *testing.T) {
	mockService := &MockRestAPIService{
		ListRestAPIsFunc: func(ctx context.Context, req model.ListRestAPIsRequest) (*model.PageResult[model.RestAPI], error) {
			// Verify filters were applied
			if req.Method != "GET" {
				t.Errorf("Expected method filter 'GET', got '%s'", req.Method)
			}
			if req.Search != "user" {
				t.Errorf("Expected search 'user', got '%s'", req.Search)
			}
			return &model.PageResult[model.RestAPI]{Data: []model.RestAPI{}}, nil
		},
	}

	ctrl := controller.NewRestAPIController(mockService)
	r := setupRestAPITestRouter(ctrl)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rest-apis?method=GET&search=user", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

// ==================== Helper Functions ====================

// strPtr returns a pointer to a string
func strPtr(s string) *string {
	return &s
}
