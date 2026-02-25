package controller_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"golang-basic/api/internal/controller"
	"golang-basic/api/internal/model"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
)

type MockProfileService struct {
	CreateProfileFunc    func(ctx context.Context, req model.CreateProfileRequest) (*model.Profile, error)
	UpdateProfileFunc    func(ctx context.Context, req model.UpdateProfileRequest) (*model.Profile, error)
	GetAllProfilesFunc   func(ctx context.Context, req model.PageRequest, lastCursor int64) (*model.PageResult[model.Profile], error)
	GetProfileByIDFunc   func(ctx context.Context, id int64) (model.Profile, error)
	GetProfileByUserIDFunc func(ctx context.Context, userID int64) (model.Profile, error)
	DeleteProfileFunc    func(ctx context.Context, profileId int64) error
}

func (m *MockProfileService) CreateProfile(ctx context.Context, req model.CreateProfileRequest) (*model.Profile, error) {
	if m.CreateProfileFunc == nil {
		return nil, errors.New("CreateProfileFunc not implemented")
	}
	return m.CreateProfileFunc(ctx, req)
}

func (m *MockProfileService) UpdateProfile(ctx context.Context, req model.UpdateProfileRequest) (*model.Profile, error) {
	if m.UpdateProfileFunc == nil {
		return nil, errors.New("UpdateProfileFunc not implemented")
	}
	return m.UpdateProfileFunc(ctx, req)
}

func (m *MockProfileService) GetAllProfiles(ctx context.Context, req model.PageRequest, lastCursor int64) (*model.PageResult[model.Profile], error) {
	if m.GetAllProfilesFunc == nil {
		return nil, errors.New("GetAllProfilesFunc not implemented")
	}
	return m.GetAllProfilesFunc(ctx, req, lastCursor)
}

func (m *MockProfileService) GetProfileByID(ctx context.Context, id int64) (model.Profile, error) {
	if m.GetProfileByIDFunc == nil {
		return model.Profile{}, errors.New("GetProfileByIDFunc not implemented")
	}
	return m.GetProfileByIDFunc(ctx, id)
}

func (m *MockProfileService) GetProfileByUserID(ctx context.Context, userID int64) (model.Profile, error) {
	if m.GetProfileByUserIDFunc == nil {
		return model.Profile{}, errors.New("GetProfileByUserIDFunc not implemented")
	}
	return m.GetProfileByUserIDFunc(ctx, userID)
}

func (m *MockProfileService) DeleteProfile(ctx context.Context, profileId int64) error {
	if m.DeleteProfileFunc == nil {
		return errors.New("DeleteProfileFunc not implemented")
	}
	return m.DeleteProfileFunc(ctx, profileId)
}

// helper to convert *int8 to pgtype.Int8 for tests
func pgTypeInt8(val *int8) pgtype.Int8 {
	if val == nil {
		return pgtype.Int8{Valid: false}
	}
	return pgtype.Int8{Int64: int64(*val), Valid: true}
}

// helper to convert *string to pgtype.Text for tests
func pgTypeText(val *string) pgtype.Text {
	if val == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *val, Valid: true}
}

// helper to convert int8 to pgtype.Int8 for tests
func pgTypeInt8FromInt8(val int8) pgtype.Int8 {
	return pgtype.Int8{Int64: int64(val), Valid: true}
}

// helper to convert string to pgtype.Text for tests
func pgTypeTextFromString(val string) pgtype.Text {
	return pgtype.Text{String: val, Valid: true}
}

func TestCreateProfile_Success(t *testing.T) {
	mockService := &MockProfileService{
		CreateProfileFunc: func(ctx context.Context, req model.CreateProfileRequest) (*model.Profile, error) {
			// Convert pointer fields to pgtype for the response
			return &model.Profile{
				ProfileId:   1,
				UserID:      req.UserID,
				Age:         pgTypeInt8(req.Age),
				Gender:      pgTypeText(req.Gender),
				Bio:         pgTypeText(req.Bio),
				PhoneNumber: pgTypeText(req.PhoneNumber),
				Website:     pgTypeText(req.Website),
			}, nil
		},
	}

	ctrl := controller.NewProfileController(mockService)

	age := int8(25)
	gender := "male"
	bio := "Software developer"
	phoneNumber := "+1234567890"
	website := "https://example.com"

	reqBody := model.CreateProfileRequest{
		UserID:      1,
		Age:         &age,
		Gender:      &gender,
		Bio:         &bio,
		PhoneNumber: &phoneNumber,
		Website:     &website,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/profiles", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	ctrl.CreateProfile(w, req)

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
	ctrl := controller.NewProfileController(mockService)

	req := httptest.NewRequest(http.MethodGet, "/profiles", nil)
	w := httptest.NewRecorder()

	ctrl.CreateProfile(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status code %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestCreateProfile_InvalidRequestBody(t *testing.T) {
	mockService := &MockProfileService{}
	ctrl := controller.NewProfileController(mockService)

	req := httptest.NewRequest(http.MethodPost, "/profiles", bytes.NewBuffer([]byte("invalid json")))
	w := httptest.NewRecorder()

	ctrl.CreateProfile(w, req)

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

	ctrl := controller.NewProfileController(mockService)

	bio := "Test bio"
	reqBody := model.CreateProfileRequest{
		UserID: 999,
		Bio:    &bio,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/profiles", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	ctrl.CreateProfile(w, req)

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

	ctrl := controller.NewProfileController(mockService)

	age := int8(25)
	gender := "invalid"
	bio := "Test bio"

	reqBody := model.CreateProfileRequest{
		UserID: 1,
		Age:    &age,
		Gender: &gender,
		Bio:    &bio,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/profiles", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	ctrl.CreateProfile(w, req)

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

	ctrl := controller.NewProfileController(mockService)

	age := int8(121)
	gender := "male"
	bio := "Test bio"

	reqBody := model.CreateProfileRequest{
		UserID: 1,
		Age:    &age,
		Gender: &gender,
		Bio:    &bio,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/profiles", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	ctrl.CreateProfile(w, req)

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
				Age:         pgTypeInt8(req.Age),
				Gender:      pgTypeText(req.Gender),
				Bio:         pgTypeText(req.Bio),
				PhoneNumber: pgTypeText(req.PhoneNumber),
				Website:     pgTypeText(req.Website),
			}, nil
		},
	}

	ctrl := controller.NewProfileController(mockService)

	age := int8(30)
	gender := "female"
	bio := "Updated bio"
	phoneNumber := "+9876543210"
	website := "https://newsite.com"

	reqBody := model.UpdateProfileRequest{
		ProfileId:   1,
		Age:         &age,
		Gender:      &gender,
		Bio:         &bio,
		PhoneNumber: &phoneNumber,
		Website:     &website,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/profiles", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	ctrl.UpdateProfile(w, req)

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
	ctrl := controller.NewProfileController(mockService)

	req := httptest.NewRequest(http.MethodGet, "/profiles", nil)
	w := httptest.NewRecorder()

	ctrl.UpdateProfile(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status code %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestUpdateProfile_InvalidRequestBody(t *testing.T) {
	mockService := &MockProfileService{}
	ctrl := controller.NewProfileController(mockService)

	req := httptest.NewRequest(http.MethodPut, "/profiles", bytes.NewBuffer([]byte("{invalid}")))
	w := httptest.NewRecorder()

	ctrl.UpdateProfile(w, req)

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

	ctrl := controller.NewProfileController(mockService)

	bio := "Updated bio"
	reqBody := model.UpdateProfileRequest{
		ProfileId: 999,
		Bio:       &bio,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/profiles", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	ctrl.UpdateProfile(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGetAllProfiles_Success(t *testing.T) {
	mockProfiles := []model.Profile{
		{
			ProfileId:   1,
			UserID:      1,
			Age:         pgTypeInt8FromInt8(25),
			Gender:      pgTypeTextFromString("male"),
			Bio:         pgTypeTextFromString("Developer"),
			PhoneNumber: pgTypeTextFromString("+1234567890"),
			User:        &model.User{UserId: 1, Name: "John Doe", Email: "john@example.com"},
		},
		{
			ProfileId:   2,
			UserID:      2,
			Age:         pgTypeInt8FromInt8(28),
			Gender:      pgTypeTextFromString("female"),
			Bio:         pgTypeTextFromString("Designer"),
			PhoneNumber: pgTypeTextFromString("+9876543210"),
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

	ctrl := controller.NewProfileController(mockService)

	req := httptest.NewRequest(http.MethodGet, "/profiles?page=1&size=10", nil)
	w := httptest.NewRecorder()

	ctrl.GetAllProfiles(w, req)

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
	ctrl := controller.NewProfileController(mockService)

	req := httptest.NewRequest(http.MethodPost, "/profiles", nil)
	w := httptest.NewRecorder()

	ctrl.GetAllProfiles(w, req)

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
				Age:         pgTypeInt8FromInt8(25),
				Gender:      pgTypeTextFromString("male"),
				Bio:         pgTypeTextFromString("Developer"),
				PhoneNumber: pgTypeTextFromString("+1234567890"),
				User:        &model.User{UserId: 1, Name: "John Doe"},
			}, nil
		},
	}

	ctrl := controller.NewProfileController(mockService)

	req := httptest.NewRequest(http.MethodGet, "/profiles/1", nil)
	w := httptest.NewRecorder()

	ctrl.GetProfileByID(w, req)

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
	ctrl := controller.NewProfileController(mockService)

	req := httptest.NewRequest(http.MethodPost, "/profiles/1", nil)
	w := httptest.NewRecorder()

	ctrl.GetProfileByID(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status code %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestGetProfileByID_InvalidID(t *testing.T) {
	mockService := &MockProfileService{}
	ctrl := controller.NewProfileController(mockService)

	req := httptest.NewRequest(http.MethodGet, "/profiles/invalid", nil)
	w := httptest.NewRecorder()

	ctrl.GetProfileByID(w, req)

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

	ctrl := controller.NewProfileController(mockService)

	req := httptest.NewRequest(http.MethodGet, "/profiles/999", nil)
	w := httptest.NewRecorder()

	ctrl.GetProfileByID(w, req)

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

	ctrl := controller.NewProfileController(mockService)

	req := httptest.NewRequest(http.MethodDelete, "/profiles/1", nil)
	w := httptest.NewRecorder()

	ctrl.DeleteProfile(w, req)

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
	ctrl := controller.NewProfileController(mockService)

	req := httptest.NewRequest(http.MethodGet, "/profiles/1", nil)
	w := httptest.NewRecorder()

	ctrl.DeleteProfile(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status code %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestDeleteProfile_InvalidID(t *testing.T) {
	mockService := &MockProfileService{}
	ctrl := controller.NewProfileController(mockService)

	req := httptest.NewRequest(http.MethodDelete, "/profiles/invalid", nil)
	w := httptest.NewRecorder()

	ctrl.DeleteProfile(w, req)

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

	ctrl := controller.NewProfileController(mockService)

	req := httptest.NewRequest(http.MethodDelete, "/profiles/999", nil)
	w := httptest.NewRecorder()

	ctrl.DeleteProfile(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	if response["status"] != "error" {
		t.Errorf("Expected status 'error', got '%s'", response["status"])
	}
}
