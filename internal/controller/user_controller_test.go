package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"golang-basic/internal/model"
	"net/http"
	"net/http/httptest"
	"testing"
)

type MockUserService struct {
	CreateUserFunc  func(ctx context.Context, req model.CreateUserRequest) (*model.User, error)
	UpdateUserFunc  func(ctx context.Context, req model.UpdateUserRequest) error
	GetAllUsersFunc func(ctx context.Context, req model.PageRequest, lastCursor int64) (*model.PageResult[model.User], error)
	GetUserByIDFunc func(ctx context.Context, id int64) (model.User, error)
}

func (m *MockUserService) CreateUser(ctx context.Context, req model.CreateUserRequest) (*model.User, error) {
	return m.CreateUserFunc(ctx, req)
}

func (m *MockUserService) UpdateUser(ctx context.Context, req model.UpdateUserRequest) error {
	return m.UpdateUserFunc(ctx, req)
}

func (m *MockUserService) GetAllUsers(ctx context.Context, req model.PageRequest, lastCursor int64) (*model.PageResult[model.User], error) {
	return m.GetAllUsersFunc(ctx, req, lastCursor)
}

func (m *MockUserService) GetUserByID(ctx context.Context, id int64) (model.User, error) {
	return m.GetUserByIDFunc(ctx, id)
}

func TestCreateUser_Success(t *testing.T) {
	mockService := &MockUserService{
		CreateUserFunc: func(ctx context.Context, req model.CreateUserRequest) (*model.User, error) {
			return &model.User{
				Name:  req.Name,
				Email: req.Email,
			}, nil
		},
	}

	controller := &UserController{service: mockService}

	reqBody := model.CreateUserRequest{
		Name:  "John Doe",
		Email: "john.doe@example.com",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	controller.CreateUser(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	if response["status"] != "success" {
		t.Errorf("Expected status 'success', got '%s'", response["status"])
	}
}

func TestCreateUser_InvalidMethod(t *testing.T) {
	mockService := &MockUserService{}
	controller := &UserController{service: mockService}

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	w := httptest.NewRecorder()

	controller.CreateUser(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status code %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestCreateUser_InvalidRequestBody(t *testing.T) {
	mockService := &MockUserService{}
	controller := &UserController{service: mockService}

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer([]byte("invalid json")))
	w := httptest.NewRecorder()

	controller.CreateUser(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	if response["message"] != "Invalid request body" {
		t.Errorf("Expected message 'Invalid request body', got '%s'", response["message"])
	}
}

func TestCreateUser_ServiceError(t *testing.T) {
	mockService := &MockUserService{
		CreateUserFunc: func(ctx context.Context, req model.CreateUserRequest) (*model.User, error) {
			return nil, errors.New("validation error: invalid email")
		},
	}

	controller := &UserController{service: mockService}

	reqBody := model.CreateUserRequest{
		Name:  "John Doe",
		Email: "invalid-email",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	controller.CreateUser(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	if response["status"] != "error" {
		t.Errorf("Expected status 'error', got '%s'", response["status"])
	}
}

func TestUpdateUser_Success(t *testing.T) {
	mockService := &MockUserService{
		UpdateUserFunc: func(ctx context.Context, req model.UpdateUserRequest) error {
			return nil
		},
	}

	controller := &UserController{service: mockService}

	reqBody := model.UpdateUserRequest{
		UserId:   1,
		Name:     "John Updated",
		Email:    "john.updated@example.com",
		IsActive: true,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/users", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	controller.UpdateUser(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	if response["status"] != "success" {
		t.Errorf("Expected status 'success', got '%s'", response["status"])
	}
}

func TestUpdateUser_InvalidMethod(t *testing.T) {
	mockService := &MockUserService{}
	controller := &UserController{service: mockService}

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	w := httptest.NewRecorder()

	controller.UpdateUser(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status code %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestUpdateUser_InvalidRequestBody(t *testing.T) {
	mockService := &MockUserService{}
	controller := &UserController{service: mockService}

	req := httptest.NewRequest(http.MethodPut, "/users", bytes.NewBuffer([]byte("{invalid}")))
	w := httptest.NewRecorder()

	controller.UpdateUser(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestUpdateUser_ServiceError(t *testing.T) {
	mockService := &MockUserService{
		UpdateUserFunc: func(ctx context.Context, req model.UpdateUserRequest) error {
			return errors.New("user not found")
		},
	}

	controller := &UserController{service: mockService}

	reqBody := model.UpdateUserRequest{
		UserId: 999,
		Name:   "Non Existent",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/users", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	controller.UpdateUser(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGetAllUsers_Success(t *testing.T) {
	mockUsers := []model.User{
		{UserId: 1, Name: "John Doe", Email: "john@example.com", IsActive: true},
		{UserId: 2, Name: "Jane Smith", Email: "jane@example.com", IsActive: false},
	}

	mockService := &MockUserService{
		GetAllUsersFunc: func(ctx context.Context, req model.PageRequest, lastCursor int64) (*model.PageResult[model.User], error) {
			return &model.PageResult[model.User]{
				Data: mockUsers,
			}, nil
		},
	}

	controller := &UserController{service: mockService}

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	w := httptest.NewRecorder()

	controller.GetAllUsers(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	if response["status"] != "success" {
		t.Errorf("Expected status 'success', got '%s'", response["status"])
	}
}

func TestGetAllUsers_InvalidMethod(t *testing.T) {
	mockService := &MockUserService{}
	controller := &UserController{service: mockService}

	req := httptest.NewRequest(http.MethodPost, "/users", nil)
	w := httptest.NewRecorder()

	controller.GetAllUsers(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status code %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestGetAllUsers_EmptyList(t *testing.T) {
	mockService := &MockUserService{
		GetAllUsersFunc: func(ctx context.Context, req model.PageRequest, lastCursor int64) (*model.PageResult[model.User], error) {
			return &model.PageResult[model.User]{
				Data: []model.User{},
			}, nil
		},
	}

	controller := &UserController{service: mockService}

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	w := httptest.NewRecorder()

	controller.GetAllUsers(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

func TestGetUserByID_Success(t *testing.T) {
	mockService := &MockUserService{
		GetUserByIDFunc: func(ctx context.Context, id int64) (model.User, error) {
			return model.User{
				UserId:   id,
				Name:     "John Doe",
				Email:    "john@example.com",
				IsActive: true,
			}, nil
		},
	}

	controller := &UserController{service: mockService}

	req := httptest.NewRequest(http.MethodGet, "/users/1", nil)
	w := httptest.NewRecorder()

	controller.GetUserByID(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	if response["status"] != "success" {
		t.Errorf("Expected status 'success', got '%s'", response["status"])
	}
}

func TestGetUserByID_InvalidMethod(t *testing.T) {
	mockService := &MockUserService{}
	controller := &UserController{service: mockService}

	req := httptest.NewRequest(http.MethodPost, "/users/1", nil)
	w := httptest.NewRecorder()

	controller.GetUserByID(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status code %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestGetUserByID_InvalidID(t *testing.T) {
	mockService := &MockUserService{}
	controller := &UserController{service: mockService}

	req := httptest.NewRequest(http.MethodGet, "/users/invalid", nil)
	w := httptest.NewRecorder()

	controller.GetUserByID(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	if response["message"] != "Invalid user ID" {
		t.Errorf("Expected message 'Invalid user ID', got '%s'", response["message"])
	}
}

func TestGetUserByID_UserNotFound(t *testing.T) {
	mockService := &MockUserService{
		GetUserByIDFunc: func(ctx context.Context, id int64) (model.User, error) {
			return model.User{}, errors.New("user not found")
		},
	}

	controller := &UserController{service: mockService}

	req := httptest.NewRequest(http.MethodGet, "/users/999", nil)
	w := httptest.NewRecorder()

	controller.GetUserByID(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	if response["status"] != "error" {
		t.Errorf("Expected status 'error', got '%s'", response["status"])
	}
}

func TestHandleUserRoutes_CreateUser(t *testing.T) {
	mockService := &MockUserService{
		CreateUserFunc: func(ctx context.Context, req model.CreateUserRequest) (*model.User, error) {
			return &model.User{UserId: 1, Name: req.Name}, nil
		},
	}

	controller := &UserController{service: mockService}

	reqBody := model.CreateUserRequest{
		Name:  "John Doe",
		Email: "john@example.com",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	controller.HandleUserRoutes(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestHandleUserRoutes_GetAllUsers(t *testing.T) {
	mockService := &MockUserService{
		GetAllUsersFunc: func(ctx context.Context, req model.PageRequest, lastCursor int64) (*model.PageResult[model.User], error) {
			return &model.PageResult[model.User]{
				Data: []model.User{{UserId: 1, Name: "John Doe"}},
			}, nil
		},
	}

	controller := &UserController{service: mockService}

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	w := httptest.NewRecorder()

	controller.HandleUserRoutes(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHandleUserRoutes_UpdateUser(t *testing.T) {
	mockService := &MockUserService{
		UpdateUserFunc: func(ctx context.Context, req model.UpdateUserRequest) error {
			return nil
		},
	}

	controller := &UserController{service: mockService}

	reqBody := model.UpdateUserRequest{UserId: 1, Name: "Updated Name"}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/users", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	controller.HandleUserRoutes(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHandleUserRoutes_GetUserByID(t *testing.T) {
	mockService := &MockUserService{
		GetUserByIDFunc: func(ctx context.Context, id int64) (model.User, error) {
			return model.User{UserId: id, Name: "John Doe"}, nil
		},
	}

	controller := &UserController{service: mockService}

	req := httptest.NewRequest(http.MethodGet, "/users/1", nil)
	w := httptest.NewRecorder()

	controller.HandleUserRoutes(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHandleUserRoutes_InvalidRoute(t *testing.T) {
	mockService := &MockUserService{}
	controller := &UserController{service: mockService}

	req := httptest.NewRequest(http.MethodGet, "/invalid/route", nil)
	w := httptest.NewRecorder()

	controller.HandleUserRoutes(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestHandleUserRoutes_MethodNotAllowed(t *testing.T) {
	mockService := &MockUserService{}
	controller := &UserController{service: mockService}

	req := httptest.NewRequest(http.MethodDelete, "/users", nil)
	w := httptest.NewRecorder()

	controller.HandleUserRoutes(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status code %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}
