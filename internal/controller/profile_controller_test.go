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

type MockProfileService struct {
	CreateProfileFunc  func(ctx context.Context, req model.CreateProfileRequest) (*model.Profile, error)
	UpdateProfileFunc  func(ctx context.Context, req model.UpdateProfileRequest) (*model.Profile, error)
	GetAllProfilesFunc func(ctx context.Context, req model.PageRequest, lastCursor int64) (*model.PageResult[model.Profile], error)
	GetProfileByIDFunc func(ctx context.Context, id int64) (model.Profile, error)
	DeleteProfileFunc  func(ctx context.Context, profileId int64) error
}

func (m *MockProfileService) CreateProfile(ctx context.Context, req model.CreateProfileRequest) (*model.Profile, error) {
	return m.CreateProfileFunc(ctx, req)
}

func (m *MockProfileService) UpdateProfile(ctx context.Context, req model.UpdateProfileRequest) (*model.Profile, error) {
	return m.UpdateProfileFunc(ctx, req)
}

func (m *MockProfileService) GetAllProfiles(ctx context.Context, req model.PageRequest, lastCursor int64) (*model.PageResult[model.Profile], error) {
	return m.GetAllProfilesFunc(ctx, req, lastCursor)
}

func (m *MockProfileService) GetProfileByID(ctx context.Context, id int64) (model.Profile, error) {
	return m.GetProfileByIDFunc(ctx, id)
}

func (m *MockProfileService) DeleteProfile(ctx context.Context, profileId int64) error {
	return m.DeleteProfileFunc(ctx, profileId)
}

func TestCreateProfile_Success(t *testing.T) {
	mockService := &MockProfileService{
		CreateProfileFunc: func(ctx context.Context, req model.CreateProfileRequest) (*model.Profile, error) {
			return &model.Profile{
				ProfileId:   1,
				UserID:      req.UserID,
				Age:         req.Age,
				Gender:      req.Gender,
				Bio:         req.Bio,
				PhoneNumber: req.PhoneNumber,
				Website:     req.Website,
			}, nil
		},
	}

	controller := &ProfileController{service: mockService}

	reqBody := model.CreateProfileRequest{
		UserID:      1,
		Age:         25,
		Gender:      "male",
		Bio:         "Software developer",
		PhoneNumber: "+1234567890",
		Website:     "https://example.com",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/profiles", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	controller.CreateProfile(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	if response["status"] != "success" {
		t.Errorf("Expected status 'success', got '%s'", response["status"])
	}
}

func TestCreateProfile_InvalidMethod(t *testing.T) {
	mockService := &MockProfileService{}
	controller := &ProfileController{service: mockService}

	req := httptest.NewRequest(http.MethodGet, "/profiles", nil)
	w := httptest.NewRecorder()

	controller.CreateProfile(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status code %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestCreateProfile_InvalidRequestBody(t *testing.T) {
	mockService := &MockProfileService{}
	controller := &ProfileController{service: mockService}

	req := httptest.NewRequest(http.MethodPost, "/profiles", bytes.NewBuffer([]byte("invalid json")))
	w := httptest.NewRecorder()

	controller.CreateProfile(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	if response["message"] != "Invalid request body" {
		t.Errorf("Expected message 'Invalid request body', got '%s'", response["message"])
	}
}

func TestCreateProfile_ServiceError(t *testing.T) {
	mockService := &MockProfileService{
		CreateProfileFunc: func(ctx context.Context, req model.CreateProfileRequest) (*model.Profile, error) {
			return nil, errors.New("user not found")
		},
	}

	controller := &ProfileController{service: mockService}

	reqBody := model.CreateProfileRequest{
		UserID: 999,
		Bio:    "Test bio",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/profiles", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	controller.CreateProfile(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	if response["status"] != "error" {
		t.Errorf("Expected status 'error', got '%s'", response["status"])
	}
}

func TestCreateProfile_InvalidGender(t *testing.T) {
	mockService := &MockProfileService{
		CreateProfileFunc: func(ctx context.Context, req model.CreateProfileRequest) (*model.Profile, error) {
			return nil, errors.New("Gender: must be 'male', 'female', or 'other'")
		},
	}

	controller := &ProfileController{service: mockService}

	reqBody := model.CreateProfileRequest{
		UserID:  1,
		Age:     25,
		Gender:  "invalid",
		Bio:     "Test bio",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/profiles", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	controller.CreateProfile(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestCreateProfile_InvalidAge(t *testing.T) {
	mockService := &MockProfileService{
		CreateProfileFunc: func(ctx context.Context, req model.CreateProfileRequest) (*model.Profile, error) {
			return nil, errors.New("Age: must be between 0 and 120")
		},
	}

	controller := &ProfileController{service: mockService}

	reqBody := model.CreateProfileRequest{
		UserID: 1,
		Age:    150,
		Gender: "male",
		Bio:    "Test bio",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/profiles", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	controller.CreateProfile(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestUpdateProfile_Success(t *testing.T) {
	mockService := &MockProfileService{
		UpdateProfileFunc: func(ctx context.Context, req model.UpdateProfileRequest) (*model.Profile, error) {
			return &model.Profile{
				ProfileId:   req.ProfileId,
				UserID:      1,
				Age:         req.Age,
				Gender:      req.Gender,
				Bio:         req.Bio,
				PhoneNumber: req.PhoneNumber,
				Website:     req.Website,
			}, nil
		},
	}

	controller := &ProfileController{service: mockService}

	reqBody := model.UpdateProfileRequest{
		ProfileId:   1,
		Age:         30,
		Gender:      "female",
		Bio:         "Updated bio",
		PhoneNumber: "+9876543210",
		Website:     "https://newsite.com",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/profiles", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	controller.UpdateProfile(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	if response["status"] != "success" {
		t.Errorf("Expected status 'success', got '%s'", response["status"])
	}
}

func TestUpdateProfile_InvalidMethod(t *testing.T) {
	mockService := &MockProfileService{}
	controller := &ProfileController{service: mockService}

	req := httptest.NewRequest(http.MethodGet, "/profiles", nil)
	w := httptest.NewRecorder()

	controller.UpdateProfile(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status code %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestUpdateProfile_InvalidRequestBody(t *testing.T) {
	mockService := &MockProfileService{}
	controller := &ProfileController{service: mockService}

	req := httptest.NewRequest(http.MethodPut, "/profiles", bytes.NewBuffer([]byte("{invalid}")))
	w := httptest.NewRecorder()

	controller.UpdateProfile(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestUpdateProfile_ServiceError(t *testing.T) {
	mockService := &MockProfileService{
		UpdateProfileFunc: func(ctx context.Context, req model.UpdateProfileRequest) (*model.Profile, error) {
			return nil, errors.New("profile not found")
		},
	}

	controller := &ProfileController{service: mockService}

	reqBody := model.UpdateProfileRequest{
		ProfileId: 999,
		Bio:       "Updated bio",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/profiles", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	controller.UpdateProfile(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGetAllProfiles_Success(t *testing.T) {
	mockProfiles := []model.Profile{
		{
			ProfileId:   1,
			UserID:      1,
			Age:         25,
			Gender:      "male",
			Bio:         "Developer",
			PhoneNumber: "+1234567890",
			User:        &model.User{UserId: 1, Name: "John Doe", Email: "john@example.com"},
		},
		{
			ProfileId:   2,
			UserID:      2,
			Age:         28,
			Gender:      "female",
			Bio:         "Designer",
			PhoneNumber: "+9876543210",
			User:        &model.User{UserId: 2, Name: "Jane Smith", Email: "jane@example.com"},
		},
	}

	mockService := &MockProfileService{
		GetAllProfilesFunc: func(ctx context.Context, req model.PageRequest, lastCursor int64) (*model.PageResult[model.Profile], error) {
			return &model.PageResult[model.Profile]{
				Data:       mockProfiles,
				Page:       1,
				Size:       10,
				StartRow:   1,
				EndRow:     2,
				NextCursor: 2,
			}, nil
		},
	}

	controller := &ProfileController{service: mockService}

	req := httptest.NewRequest(http.MethodGet, "/profiles?page=1&size=10", nil)
	w := httptest.NewRecorder()

	controller.GetAllProfiles(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	if response["status"] != "success" {
		t.Errorf("Expected status 'success', got '%s'", response["status"])
	}
}

func TestGetAllProfiles_InvalidMethod(t *testing.T) {
	mockService := &MockProfileService{}
	controller := &ProfileController{service: mockService}

	req := httptest.NewRequest(http.MethodPost, "/profiles", nil)
	w := httptest.NewRecorder()

	controller.GetAllProfiles(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status code %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestGetProfileByID_Success(t *testing.T) {
	mockService := &MockProfileService{
		GetProfileByIDFunc: func(ctx context.Context, id int64) (model.Profile, error) {
			return model.Profile{
				ProfileId:   id,
				UserID:      1,
				Age:         25,
				Gender:      "male",
				Bio:         "Developer",
				PhoneNumber: "+1234567890",
				User:        &model.User{UserId: 1, Name: "John Doe"},
			}, nil
		},
	}

	controller := &ProfileController{service: mockService}

	req := httptest.NewRequest(http.MethodGet, "/profiles/1", nil)
	w := httptest.NewRecorder()

	controller.GetProfileByID(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	if response["status"] != "success" {
		t.Errorf("Expected status 'success', got '%s'", response["status"])
	}
}

func TestGetProfileByID_InvalidMethod(t *testing.T) {
	mockService := &MockProfileService{}
	controller := &ProfileController{service: mockService}

	req := httptest.NewRequest(http.MethodPost, "/profiles/1", nil)
	w := httptest.NewRecorder()

	controller.GetProfileByID(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status code %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestGetProfileByID_InvalidID(t *testing.T) {
	mockService := &MockProfileService{}
	controller := &ProfileController{service: mockService}

	req := httptest.NewRequest(http.MethodGet, "/profiles/invalid", nil)
	w := httptest.NewRecorder()

	controller.GetProfileByID(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	if response["message"] != "Invalid profile ID" {
		t.Errorf("Expected message 'Invalid profile ID', got '%s'", response["message"])
	}
}

func TestGetProfileByID_ProfileNotFound(t *testing.T) {
	mockService := &MockProfileService{
		GetProfileByIDFunc: func(ctx context.Context, id int64) (model.Profile, error) {
			return model.Profile{}, errors.New("profile not found")
		},
	}

	controller := &ProfileController{service: mockService}

	req := httptest.NewRequest(http.MethodGet, "/profiles/999", nil)
	w := httptest.NewRecorder()

	controller.GetProfileByID(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	if response["status"] != "error" {
		t.Errorf("Expected status 'error', got '%s'", response["status"])
	}
}

func TestDeleteProfile_Success(t *testing.T) {
	mockService := &MockProfileService{
		DeleteProfileFunc: func(ctx context.Context, profileId int64) error {
			return nil
		},
	}

	controller := &ProfileController{service: mockService}

	req := httptest.NewRequest(http.MethodDelete, "/profiles/1", nil)
	w := httptest.NewRecorder()

	controller.DeleteProfile(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	if response["status"] != "success" {
		t.Errorf("Expected status 'success', got '%s'", response["status"])
	}
}

func TestDeleteProfile_InvalidMethod(t *testing.T) {
	mockService := &MockProfileService{}
	controller := &ProfileController{service: mockService}

	req := httptest.NewRequest(http.MethodGet, "/profiles/1", nil)
	w := httptest.NewRecorder()

	controller.DeleteProfile(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status code %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestDeleteProfile_InvalidID(t *testing.T) {
	mockService := &MockProfileService{}
	controller := &ProfileController{service: mockService}

	req := httptest.NewRequest(http.MethodDelete, "/profiles/invalid", nil)
	w := httptest.NewRecorder()

	controller.DeleteProfile(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	if response["message"] != "Invalid profile ID" {
		t.Errorf("Expected message 'Invalid profile ID', got '%s'", response["message"])
	}
}

func TestDeleteProfile_ProfileNotFound(t *testing.T) {
	mockService := &MockProfileService{
		DeleteProfileFunc: func(ctx context.Context, profileId int64) error {
			return errors.New("profile not found")
		},
	}

	controller := &ProfileController{service: mockService}

	req := httptest.NewRequest(http.MethodDelete, "/profiles/999", nil)
	w := httptest.NewRecorder()

	controller.DeleteProfile(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	if response["status"] != "error" {
		t.Errorf("Expected status 'error', got '%s'", response["status"])
	}
}

func TestHandleProfileRoutes_CreateProfile(t *testing.T) {
	mockService := &MockProfileService{
		CreateProfileFunc: func(ctx context.Context, req model.CreateProfileRequest) (*model.Profile, error) {
			return &model.Profile{ProfileId: 1, UserID: req.UserID, Bio: req.Bio, Age: req.Age, Gender: req.Gender}, nil
		},
	}

	controller := &ProfileController{service: mockService}

	reqBody := model.CreateProfileRequest{
		UserID:  1,
		Age:     25,
		Gender:  "male",
		Bio:     "Developer",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/profiles", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	controller.HandleProfileRoutes(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestHandleProfileRoutes_GetAllProfiles(t *testing.T) {
	mockService := &MockProfileService{
		GetAllProfilesFunc: func(ctx context.Context, req model.PageRequest, lastCursor int64) (*model.PageResult[model.Profile], error) {
			return &model.PageResult[model.Profile]{
				Data: []model.Profile{{ProfileId: 1, UserID: 1, Bio: "Developer", Age: 25, Gender: "male"}},
			}, nil
		},
	}

	controller := &ProfileController{service: mockService}

	req := httptest.NewRequest(http.MethodGet, "/profiles", nil)
	w := httptest.NewRecorder()

	controller.HandleProfileRoutes(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHandleProfileRoutes_UpdateProfile(t *testing.T) {
	mockService := &MockProfileService{
		UpdateProfileFunc: func(ctx context.Context, req model.UpdateProfileRequest) (*model.Profile, error) {
			return &model.Profile{ProfileId: req.ProfileId, Bio: req.Bio, Age: req.Age, Gender: req.Gender}, nil
		},
	}

	controller := &ProfileController{service: mockService}

	reqBody := model.UpdateProfileRequest{ProfileId: 1, Bio: "Updated bio", Age: 30, Gender: "female"}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/profiles", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	controller.HandleProfileRoutes(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHandleProfileRoutes_GetProfileByID(t *testing.T) {
	mockService := &MockProfileService{
		GetProfileByIDFunc: func(ctx context.Context, id int64) (model.Profile, error) {
			return model.Profile{ProfileId: id, UserID: 1, Bio: "Developer", Age: 25, Gender: "male"}, nil
		},
	}

	controller := &ProfileController{service: mockService}

	req := httptest.NewRequest(http.MethodGet, "/profiles/1", nil)
	w := httptest.NewRecorder()

	controller.HandleProfileRoutes(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHandleProfileRoutes_DeleteProfile(t *testing.T) {
	mockService := &MockProfileService{
		DeleteProfileFunc: func(ctx context.Context, profileId int64) error {
			return nil
		},
	}

	controller := &ProfileController{service: mockService}

	req := httptest.NewRequest(http.MethodDelete, "/profiles/1", nil)
	w := httptest.NewRecorder()

	controller.HandleProfileRoutes(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHandleProfileRoutes_InvalidRoute(t *testing.T) {
	mockService := &MockProfileService{}
	controller := &ProfileController{service: mockService}

	req := httptest.NewRequest(http.MethodGet, "/invalid/route", nil)
	w := httptest.NewRecorder()

	controller.HandleProfileRoutes(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestHandleProfileRoutes_MethodNotAllowed(t *testing.T) {
	mockService := &MockProfileService{}
	controller := &ProfileController{service: mockService}

	req := httptest.NewRequest(http.MethodPatch, "/profiles", nil)
	w := httptest.NewRecorder()

	controller.HandleProfileRoutes(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status code %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}
