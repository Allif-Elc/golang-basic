package controller_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"golang-basic/api/internal/controller"
	"golang-basic/api/internal/model"
	"golang-basic/api/internal/service"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// MockPermissionService is a mock implementation of PermissionServiceInterface for testing
type MockPermissionService struct {
	// Attribute mocks
	CreateAttributeFunc  func(ctx context.Context, req model.CreateAttributesRequest) (*model.Attributes, error)
	GetAttributeByIDFunc func(ctx context.Context, id int64) (*model.Attributes, error)
	ListAttributesFunc   func(ctx context.Context) ([]model.Attributes, error)
	UpdateAttributeFunc  func(ctx context.Context, id int64, req model.UpdateAttributesRequest) (*model.Attributes, error)
	DeleteAttributeFunc  func(ctx context.Context, id int64) error

	// Resource mocks
	CreateResourceFunc  func(ctx context.Context, req model.CreateResoucesRequest) (*model.Resources, error)
	GetResourceByIDFunc func(ctx context.Context, id int64) (*model.Resources, error)
	ListResourcesFunc   func(ctx context.Context) ([]model.Resources, error)
	UpdateResourceFunc  func(ctx context.Context, id int64, req model.UpdateResourcesRequest) (*model.Resources, error)
	DeleteResourceFunc  func(ctx context.Context, id int64) error

	// Permission mocks
	CreatePermissionFunc  func(ctx context.Context, req model.CreatePermissionsRequest) (*model.Permissions, error)
	GetPermissionByIDFunc func(ctx context.Context, id int64) (*model.Permissions, error)
	ListPermissionsFunc   func(ctx context.Context) ([]model.Permissions, error)
	UpdatePermissionFunc  func(ctx context.Context, id int64, req model.UpdatePermissionsRequest) (*model.Permissions, error)
	DeletePermissionFunc  func(ctx context.Context, id int64) error
}

// Attribute interface implementations
func (m *MockPermissionService) CreateAttribute(ctx context.Context, req model.CreateAttributesRequest) (*model.Attributes, error) {
	if m.CreateAttributeFunc != nil {
		return m.CreateAttributeFunc(ctx, req)
	}
	return &model.Attributes{AttributeID: 1, Name: req.Name, Description: req.Description}, nil
}

func (m *MockPermissionService) GetAttributeByID(ctx context.Context, id int64) (*model.Attributes, error) {
	if m.GetAttributeByIDFunc != nil {
		return m.GetAttributeByIDFunc(ctx, id)
	}
	return &model.Attributes{AttributeID: id, Name: "test_role"}, nil
}

func (m *MockPermissionService) ListAttributes(ctx context.Context) ([]model.Attributes, error) {
	if m.ListAttributesFunc != nil {
		return m.ListAttributesFunc(ctx)
	}
	return []model.Attributes{{AttributeID: 1, Name: "role"}}, nil
}

func (m *MockPermissionService) DeleteAttribute(ctx context.Context, id int64) error {
	if m.DeleteAttributeFunc != nil {
		return m.DeleteAttributeFunc(ctx, id)
	}
	return nil
}

func (m *MockPermissionService) UpdateAttribute(ctx context.Context, id int64, req model.UpdateAttributesRequest) (*model.Attributes, error) {
	if m.UpdateAttributeFunc != nil {
		return m.UpdateAttributeFunc(ctx, id, req)
	}
	result := &model.Attributes{AttributeID: id}
	if req.Name != nil {
		result.Name = *req.Name
	}
	if req.Description != nil {
		result.Description = *req.Description
	}
	return result, nil
}

// Resource interface implementations
func (m *MockPermissionService) CreateResource(ctx context.Context, req model.CreateResoucesRequest) (*model.Resources, error) {
	if m.CreateResourceFunc != nil {
		return m.CreateResourceFunc(ctx, req)
	}
	return &model.Resources{ResourceID: 1, Name: req.Name, Description: req.Description}, nil
}

func (m *MockPermissionService) GetResourceByID(ctx context.Context, id int64) (*model.Resources, error) {
	if m.GetResourceByIDFunc != nil {
		return m.GetResourceByIDFunc(ctx, id)
	}
	return &model.Resources{ResourceID: id, Name: "test_resource"}, nil
}

func (m *MockPermissionService) ListResources(ctx context.Context) ([]model.Resources, error) {
	if m.ListResourcesFunc != nil {
		return m.ListResourcesFunc(ctx)
	}
	return []model.Resources{{ResourceID: 1, Name: "employee_records"}}, nil
}

func (m *MockPermissionService) DeleteResource(ctx context.Context, id int64) error {
	if m.DeleteResourceFunc != nil {
		return m.DeleteResourceFunc(ctx, id)
	}
	return nil
}

func (m *MockPermissionService) UpdateResource(ctx context.Context, id int64, req model.UpdateResourcesRequest) (*model.Resources, error) {
	if m.UpdateResourceFunc != nil {
		return m.UpdateResourceFunc(ctx, id, req)
	}
	result := &model.Resources{ResourceID: id}
	if req.Name != nil {
		result.Name = *req.Name
	}
	if req.Description != nil {
		result.Description = *req.Description
	}
	return result, nil
}

// Permission interface implementations
func (m *MockPermissionService) CreatePermission(ctx context.Context, req model.CreatePermissionsRequest) (*model.Permissions, error) {
	if m.CreatePermissionFunc != nil {
		return m.CreatePermissionFunc(ctx, req)
	}
	return &model.Permissions{PermissionID: 1, Name: req.Name, Description: req.Description}, nil
}

func (m *MockPermissionService) GetPermissionByID(ctx context.Context, id int64) (*model.Permissions, error) {
	if m.GetPermissionByIDFunc != nil {
		return m.GetPermissionByIDFunc(ctx, id)
	}
	return &model.Permissions{PermissionID: id, Name: "test_permission"}, nil
}

func (m *MockPermissionService) ListPermissions(ctx context.Context) ([]model.Permissions, error) {
	if m.ListPermissionsFunc != nil {
		return m.ListPermissionsFunc(ctx)
	}
	return []model.Permissions{{PermissionID: 1, Name: "read_permission"}}, nil
}

func (m *MockPermissionService) DeletePermission(ctx context.Context, id int64) error {
	if m.DeletePermissionFunc != nil {
		return m.DeletePermissionFunc(ctx, id)
	}
	return nil
}

func (m *MockPermissionService) UpdatePermission(ctx context.Context, id int64, req model.UpdatePermissionsRequest) (*model.Permissions, error) {
	if m.UpdatePermissionFunc != nil {
		return m.UpdatePermissionFunc(ctx, id, req)
	}
	result := &model.Permissions{PermissionID: id}
	if req.Name != nil {
		result.Name = *req.Name
	}
	if req.Description != nil {
		result.Description = *req.Description
	}
	return result, nil
}

// ==================== ATTRIBUTE TESTS - POSITIVE CASES ====================

func TestCreateAttribute_Success(t *testing.T) {
	mockService := &MockPermissionService{}
	ctrl := controller.NewPermissionController(mockService)

	reqBody := model.CreateAttributesRequest{
		Name:        "test_role",
		Description: "Test role attribute",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/attributes", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	ctrl.CreateAttribute(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	if response["status"] != "success" {
		t.Errorf("Expected status 'success', got '%s'", response["status"])
	}
}

func TestListAttributes_Success(t *testing.T) {
	mockService := &MockPermissionService{
		ListAttributesFunc: func(ctx context.Context) ([]model.Attributes, error) {
			return []model.Attributes{
				{AttributeID: 1, Name: "role", Description: "User role", CreatedAt: time.Now()},
				{AttributeID: 2, Name: "department", Description: "Department", CreatedAt: time.Now()},
			}, nil
		},
	}

	ctrl := controller.NewPermissionController(mockService)
	req := httptest.NewRequest(http.MethodGet, "/attributes", nil)
	w := httptest.NewRecorder()

	ctrl.ListAttributes(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	if response["status"] != "success" {
		t.Errorf("Expected status 'success', got '%s'", response["status"])
	}
}

func TestGetAttributeByID_Success(t *testing.T) {
	mockService := &MockPermissionService{}
	ctrl := controller.NewPermissionController(mockService)

	req := httptest.NewRequest(http.MethodGet, "/attributes/1", nil)
	w := httptest.NewRecorder()

	ctrl.GetAttributeByID(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

func TestDeleteAttribute_Success(t *testing.T) {
	mockService := &MockPermissionService{}
	ctrl := controller.NewPermissionController(mockService)

	req := httptest.NewRequest(http.MethodDelete, "/attributes/1", nil)
	w := httptest.NewRecorder()

	ctrl.DeleteAttribute(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

// ==================== ATTRIBUTE TESTS - NEGATIVE CASES ====================

func TestCreateAttribute_InvalidMethod(t *testing.T) {
	mockService := &MockPermissionService{}
	ctrl := controller.NewPermissionController(mockService)

	req := httptest.NewRequest(http.MethodGet, "/attributes", nil)
	w := httptest.NewRecorder()

	ctrl.CreateAttribute(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status code %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestCreateAttribute_InvalidRequestBody(t *testing.T) {
	mockService := &MockPermissionService{}
	ctrl := controller.NewPermissionController(mockService)

	req := httptest.NewRequest(http.MethodPost, "/attributes", bytes.NewBuffer([]byte("invalid json")))
	w := httptest.NewRecorder()

	ctrl.CreateAttribute(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	if response["message"] != "Invalid request body" {
		t.Errorf("Expected message 'Invalid request body', got '%s'", response["message"])
	}
}

func TestCreateAttribute_ValidationError_EmptyName(t *testing.T) {
	mockService := &MockPermissionService{
		CreateAttributeFunc: func(ctx context.Context, req model.CreateAttributesRequest) (*model.Attributes, error) {
			return nil, service.ValidationError{Field: "name", Message: "length must be between 1 and 100 characters"}
		},
	}

	ctrl := controller.NewPermissionController(mockService)

	reqBody := model.CreateAttributesRequest{
		Name:        "",
		Description: "Test attribute",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/attributes", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	ctrl.CreateAttribute(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestCreateAttribute_ValidationError_SpecialCharacters(t *testing.T) {
	mockService := &MockPermissionService{
		CreateAttributeFunc: func(ctx context.Context, req model.CreateAttributesRequest) (*model.Attributes, error) {
			return nil, service.ValidationError{Field: "name", Message: "only alphanumeric characters, hyphens, and underscores are allowed"}
		},
	}

	ctrl := controller.NewPermissionController(mockService)

	reqBody := model.CreateAttributesRequest{
		Name:        "test@role!",
		Description: "Test attribute",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/attributes", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	ctrl.CreateAttribute(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestCreateAttribute_SQLInjectionAttempt(t *testing.T) {
	mockService := &MockPermissionService{
		CreateAttributeFunc: func(ctx context.Context, req model.CreateAttributesRequest) (*model.Attributes, error) {
			return nil, service.ValidationError{Field: "name", Message: "contains potentially dangerous patterns"}
		},
	}

	ctrl := controller.NewPermissionController(mockService)

	reqBody := model.CreateAttributesRequest{
		Name:        "role'; DROP TABLE attributes; --",
		Description: "Malicious attribute",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/attributes", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	ctrl.CreateAttribute(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestCreateAttribute_DuplicateName(t *testing.T) {
	mockService := &MockPermissionService{
		CreateAttributeFunc: func(ctx context.Context, req model.CreateAttributesRequest) (*model.Attributes, error) {
			return nil, service.ValidationError{Field: "name", Message: "attribute with this name already exists"}
		},
	}

	ctrl := controller.NewPermissionController(mockService)

	reqBody := model.CreateAttributesRequest{
		Name:        "role",
		Description: "Duplicate role",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/attributes", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	ctrl.CreateAttribute(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGetAttributeByID_InvalidID(t *testing.T) {
	mockService := &MockPermissionService{}
	ctrl := controller.NewPermissionController(mockService)

	req := httptest.NewRequest(http.MethodGet, "/attributes/invalid", nil)
	w := httptest.NewRecorder()

	ctrl.GetAttributeByID(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	if response["message"] != "Invalid attribute ID" {
		t.Errorf("Expected message 'Invalid attribute ID', got '%s'", response["message"])
	}
}

func TestGetAttributeByID_NotFound(t *testing.T) {
	mockService := &MockPermissionService{
		GetAttributeByIDFunc: func(ctx context.Context, id int64) (*model.Attributes, error) {
			return nil, errors.New("attribute not found")
		},
	}

	ctrl := controller.NewPermissionController(mockService)
	req := httptest.NewRequest(http.MethodGet, "/attributes/999", nil)
	w := httptest.NewRecorder()

	ctrl.GetAttributeByID(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestGetAttributeByID_InvalidMethod(t *testing.T) {
	mockService := &MockPermissionService{}
	ctrl := controller.NewPermissionController(mockService)

	req := httptest.NewRequest(http.MethodPost, "/attributes/1", nil)
	w := httptest.NewRecorder()

	ctrl.GetAttributeByID(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status code %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestDeleteAttribute_InvalidID(t *testing.T) {
	mockService := &MockPermissionService{}
	ctrl := controller.NewPermissionController(mockService)

	req := httptest.NewRequest(http.MethodDelete, "/attributes/abc", nil)
	w := httptest.NewRecorder()

	ctrl.DeleteAttribute(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestDeleteAttribute_NotFound(t *testing.T) {
	mockService := &MockPermissionService{
		DeleteAttributeFunc: func(ctx context.Context, id int64) error {
			return errors.New("attribute not found")
		},
	}

	ctrl := controller.NewPermissionController(mockService)
	req := httptest.NewRequest(http.MethodDelete, "/attributes/999", nil)
	w := httptest.NewRecorder()

	ctrl.DeleteAttribute(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// ==================== RESOURCE TESTS - POSITIVE CASES ====================

func TestCreateResource_Success(t *testing.T) {
	mockService := &MockPermissionService{}
	ctrl := controller.NewPermissionController(mockService)

	reqBody := model.CreateResoucesRequest{
		Name:        "employee_records",
		Description: "Employee records resource",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/resources", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	ctrl.CreateResource(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestListResources_Success(t *testing.T) {
	mockService := &MockPermissionService{
		ListResourcesFunc: func(ctx context.Context) ([]model.Resources, error) {
			return []model.Resources{
				{ResourceID: 1, Name: "employee_records", Description: "Employee data"},
				{ResourceID: 2, Name: "budget_report", Description: "Financial data"},
			}, nil
		},
	}

	ctrl := controller.NewPermissionController(mockService)
	req := httptest.NewRequest(http.MethodGet, "/resources", nil)
	w := httptest.NewRecorder()

	ctrl.ListResources(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

func TestGetResourceByID_Success(t *testing.T) {
	mockService := &MockPermissionService{}
	ctrl := controller.NewPermissionController(mockService)

	req := httptest.NewRequest(http.MethodGet, "/resources/1", nil)
	w := httptest.NewRecorder()

	ctrl.GetResourceByID(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

func TestDeleteResource_Success(t *testing.T) {
	mockService := &MockPermissionService{}
	ctrl := controller.NewPermissionController(mockService)

	req := httptest.NewRequest(http.MethodDelete, "/resources/1", nil)
	w := httptest.NewRecorder()

	ctrl.DeleteResource(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

// ==================== RESOURCE TESTS - NEGATIVE CASES ====================

func TestCreateResource_InvalidRequestBody(t *testing.T) {
	mockService := &MockPermissionService{}
	ctrl := controller.NewPermissionController(mockService)

	req := httptest.NewRequest(http.MethodPost, "/resources", bytes.NewBuffer([]byte("{invalid}")))
	w := httptest.NewRecorder()

	ctrl.CreateResource(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestCreateResource_NameTooLong(t *testing.T) {
	mockService := &MockPermissionService{
		CreateResourceFunc: func(ctx context.Context, req model.CreateResoucesRequest) (*model.Resources, error) {
			return nil, service.ValidationError{Field: "name", Message: "length must be between 1 and 100 characters"}
		},
	}

	ctrl := controller.NewPermissionController(mockService)

	longName := string(make([]byte, 101))
	for i := range longName {
		longName = longName[:i] + "a" + longName[i+1:]
	}

	reqBody := model.CreateResoucesRequest{
		Name:        longName,
		Description: "Test",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/resources", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	ctrl.CreateResource(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGetResourceByID_InvalidID(t *testing.T) {
	mockService := &MockPermissionService{}
	ctrl := controller.NewPermissionController(mockService)

	req := httptest.NewRequest(http.MethodGet, "/resources/not-a-number", nil)
	w := httptest.NewRecorder()

	ctrl.GetResourceByID(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestDeleteResource_NotFound(t *testing.T) {
	mockService := &MockPermissionService{
		DeleteResourceFunc: func(ctx context.Context, id int64) error {
			return errors.New("resource not found")
		},
	}

	ctrl := controller.NewPermissionController(mockService)
	req := httptest.NewRequest(http.MethodDelete, "/resources/999", nil)
	w := httptest.NewRecorder()

	ctrl.DeleteResource(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// ==================== PERMISSION TESTS - POSITIVE CASES ====================

func TestCreatePermission_Success(t *testing.T) {
	mockService := &MockPermissionService{}
	ctrl := controller.NewPermissionController(mockService)

	reqBody := model.CreatePermissionsRequest{
		Name:        "read_employee_records",
		Description: "Read access to employee records",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/permissions", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	ctrl.CreatePermission(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	if response["status"] != "success" {
		t.Errorf("Expected status 'success', got '%s'", response["status"])
	}
}

func TestListPermissions_Success(t *testing.T) {
	mockService := &MockPermissionService{
		ListPermissionsFunc: func(ctx context.Context) ([]model.Permissions, error) {
			return []model.Permissions{
				{PermissionID: 1, Name: "read", Description: "Read access"},
				{PermissionID: 2, Name: "write", Description: "Write access"},
				{PermissionID: 3, Name: "delete", Description: "Delete access"},
			}, nil
		},
	}

	ctrl := controller.NewPermissionController(mockService)
	req := httptest.NewRequest(http.MethodGet, "/permissions", nil)
	w := httptest.NewRecorder()

	ctrl.ListPermissions(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	if response["status"] != "success" {
		t.Errorf("Expected status 'success', got '%s'", response["status"])
	}
}

func TestGetPermissionByID_Success(t *testing.T) {
	mockService := &MockPermissionService{}
	ctrl := controller.NewPermissionController(mockService)

	req := httptest.NewRequest(http.MethodGet, "/permissions/1", nil)
	w := httptest.NewRecorder()

	ctrl.GetPermissionByID(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

func TestDeletePermission_Success(t *testing.T) {
	mockService := &MockPermissionService{}
	ctrl := controller.NewPermissionController(mockService)

	req := httptest.NewRequest(http.MethodDelete, "/permissions/1", nil)
	w := httptest.NewRecorder()

	ctrl.DeletePermission(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	if response["status"] != "success" {
		t.Errorf("Expected status 'success', got '%s'", response["status"])
	}
}

// ==================== PERMISSION TESTS - NEGATIVE CASES ====================

func TestCreatePermission_InvalidMethod(t *testing.T) {
	mockService := &MockPermissionService{}
	ctrl := controller.NewPermissionController(mockService)

	req := httptest.NewRequest(http.MethodGet, "/permissions", nil)
	w := httptest.NewRecorder()

	ctrl.CreatePermission(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status code %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestCreatePermission_MissingName(t *testing.T) {
	mockService := &MockPermissionService{
		CreatePermissionFunc: func(ctx context.Context, req model.CreatePermissionsRequest) (*model.Permissions, error) {
			return nil, service.ValidationError{Field: "name", Message: "length must be between 1 and 100 characters"}
		},
	}

	ctrl := controller.NewPermissionController(mockService)

	reqBody := model.CreatePermissionsRequest{
		Name:        "",
		Description: "Test permission",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/permissions", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	ctrl.CreatePermission(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestCreatePermission_InvalidCharacters(t *testing.T) {
	mockService := &MockPermissionService{
		CreatePermissionFunc: func(ctx context.Context, req model.CreatePermissionsRequest) (*model.Permissions, error) {
			return nil, service.ValidationError{Field: "name", Message: "only alphanumeric characters, hyphens, and underscores are allowed"}
		},
	}

	ctrl := controller.NewPermissionController(mockService)

	reqBody := model.CreatePermissionsRequest{
		Name:        "permission with spaces!",
		Description: "Invalid permission name",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/permissions", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	ctrl.CreatePermission(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestCreatePermission_DuplicateName(t *testing.T) {
	mockService := &MockPermissionService{
		CreatePermissionFunc: func(ctx context.Context, req model.CreatePermissionsRequest) (*model.Permissions, error) {
			return nil, service.ValidationError{Field: "name", Message: "permission with this name already exists"}
		},
	}

	ctrl := controller.NewPermissionController(mockService)

	reqBody := model.CreatePermissionsRequest{
		Name:        "read",
		Description: "Duplicate read permission",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/permissions", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	ctrl.CreatePermission(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGetPermissionByID_InvalidID(t *testing.T) {
	mockService := &MockPermissionService{}
	ctrl := controller.NewPermissionController(mockService)

	req := httptest.NewRequest(http.MethodGet, "/permissions/abc123", nil)
	w := httptest.NewRecorder()

	ctrl.GetPermissionByID(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	if response["message"] != "Invalid permission ID" {
		t.Errorf("Expected message 'Invalid permission ID', got '%s'", response["message"])
	}
}

func TestGetPermissionByID_NotFound(t *testing.T) {
	mockService := &MockPermissionService{
		GetPermissionByIDFunc: func(ctx context.Context, id int64) (*model.Permissions, error) {
			return nil, errors.New("permission not found")
		},
	}

	ctrl := controller.NewPermissionController(mockService)
	req := httptest.NewRequest(http.MethodGet, "/permissions/999", nil)
	w := httptest.NewRecorder()

	ctrl.GetPermissionByID(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	if response["status"] != "error" {
		t.Errorf("Expected status 'error', got '%s'", response["status"])
	}
}

func TestDeletePermission_InvalidID(t *testing.T) {
	mockService := &MockPermissionService{}
	ctrl := controller.NewPermissionController(mockService)

	req := httptest.NewRequest(http.MethodDelete, "/permissions/invalid-id", nil)
	w := httptest.NewRecorder()

	ctrl.DeletePermission(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestDeletePermission_NotFound(t *testing.T) {
	mockService := &MockPermissionService{
		DeletePermissionFunc: func(ctx context.Context, id int64) error {
			return errors.New("permission not found")
		},
	}

	ctrl := controller.NewPermissionController(mockService)
	req := httptest.NewRequest(http.MethodDelete, "/permissions/999", nil)
	w := httptest.NewRecorder()

	ctrl.DeletePermission(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestListPermissions_InvalidMethod(t *testing.T) {
	mockService := &MockPermissionService{}
	ctrl := controller.NewPermissionController(mockService)

	req := httptest.NewRequest(http.MethodPost, "/permissions", nil)
	w := httptest.NewRecorder()

	ctrl.ListPermissions(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status code %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

// ==================== ROUTE HANDLER TESTS ====================

