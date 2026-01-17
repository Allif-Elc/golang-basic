package middleware_test

import (
	"context"
	"encoding/json"
	"errors"
	"golang-basic/api/internal/middleware"
	"golang-basic/api/internal/model"
	"net/http"
	"net/http/httptest"
	"testing"
)

// MockAuthAuthorizationService is a mock for testing middleware
type MockAuthAuthorizationService struct {
	AuthorizeFunc func(ctx context.Context, req model.AuthorizeRequest) (model.AuthorizeResponse, error)
}

func (m *MockAuthAuthorizationService) Authorize(ctx context.Context, req model.AuthorizeRequest) (model.AuthorizeResponse, error) {
	if m.AuthorizeFunc != nil {
		return m.AuthorizeFunc(ctx, req)
	}
	return model.AuthorizeResponse{Allowed: true}, nil
}

func (m *MockAuthAuthorizationService) BatchAuthorize(ctx context.Context, userID int64, requests []model.AuthorizeRequest) ([]model.AuthorizeResponse, error) {
	return nil, nil
}

func (m *MockAuthAuthorizationService) HasPermission(ctx context.Context, userID int64, resource, action string) (bool, error) {
	return true, nil
}

func TestRequirePermission_Allowed(t *testing.T) {
	mockService := &MockAuthAuthorizationService{
		AuthorizeFunc: func(ctx context.Context, req model.AuthorizeRequest) (model.AuthorizeResponse, error) {
			return model.AuthorizeResponse{
				Allowed: true,
				Reason:  "access granted",
			}, nil
		},
	}

	mid := middleware.NewAuthorizationMiddleware(mockService)
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "success"})
	})

	protectedHandler := mid.RequirePermission("employee_records", "read")(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/employee-records", nil)
	req.Header.Set("X-User-ID", "1")
	w := httptest.NewRecorder()

	protectedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	if response["message"] != "success" {
		t.Errorf("Expected message 'success', got '%v'", response["message"])
	}
}

func TestRequirePermission_Denied(t *testing.T) {
	mockService := &MockAuthAuthorizationService{
		AuthorizeFunc: func(ctx context.Context, req model.AuthorizeRequest) (model.AuthorizeResponse, error) {
			return model.AuthorizeResponse{
				Allowed: false,
				Reason:  "access denied: insufficient permissions",
			}, nil
		},
	}

	mid := middleware.NewAuthorizationMiddleware(mockService)
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "success"})
	})

	protectedHandler := mid.RequirePermission("employee_records", "read")(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/employee-records", nil)
	req.Header.Set("X-User-ID", "1")
	w := httptest.NewRecorder()

	protectedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	if response["status"] != "error" {
		t.Errorf("Expected status 'error', got '%v'", response["status"])
	}
}

func TestRequirePermission_MissingUserID(t *testing.T) {
	mockService := &MockAuthAuthorizationService{}

	mid := middleware.NewAuthorizationMiddleware(mockService)
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	protectedHandler := mid.RequirePermission("employee_records", "read")(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/employee-records", nil)
	// No X-User-ID header
	w := httptest.NewRecorder()

	protectedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	if response["message"] != "Failed to extract user ID" {
		t.Errorf("Expected message about missing user ID, got '%v'", response["message"])
	}
}

func TestRequirePermission_InvalidUserID(t *testing.T) {
	mockService := &MockAuthAuthorizationService{}

	mid := middleware.NewAuthorizationMiddleware(mockService)
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	protectedHandler := mid.RequirePermission("employee_records", "read")(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/employee-records", nil)
	req.Header.Set("X-User-ID", "invalid")
	w := httptest.NewRecorder()

	protectedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestRequirePermission_ServiceError(t *testing.T) {
	mockService := &MockAuthAuthorizationService{
		AuthorizeFunc: func(ctx context.Context, req model.AuthorizeRequest) (model.AuthorizeResponse, error) {
			return model.AuthorizeResponse{}, errors.New("authorization service failed")
		},
	}

	mid := middleware.NewAuthorizationMiddleware(mockService)
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	protectedHandler := mid.RequirePermission("employee_records", "read")(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/employee-records", nil)
	req.Header.Set("X-User-ID", "1")
	w := httptest.NewRecorder()

	protectedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestRequireRole_Allowed(t *testing.T) {
	mockService := &MockAuthAuthorizationService{
		AuthorizeFunc: func(ctx context.Context, req model.AuthorizeRequest) (model.AuthorizeResponse, error) {
			// Check if action matches the required role
			if req.Action == "HR" {
				return model.AuthorizeResponse{Allowed: true, Reason: "has HR role"}, nil
			}
			return model.AuthorizeResponse{Allowed: false, Reason: "missing role"}, nil
		},
	}

	mid := middleware.NewAuthorizationMiddleware(mockService)
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "success"})
	})

	protectedHandler := mid.RequireRole("HR")(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/protected", nil)
	req.Header.Set("X-User-ID", "1")
	w := httptest.NewRecorder()

	protectedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestRequireRole_Denied(t *testing.T) {
	mockService := &MockAuthAuthorizationService{
		AuthorizeFunc: func(ctx context.Context, req model.AuthorizeRequest) (model.AuthorizeResponse, error) {
			return model.AuthorizeResponse{
				Allowed: false,
				Reason:  "Role 'Manager' required",
			}, nil
		},
	}

	mid := middleware.NewAuthorizationMiddleware(mockService)
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	protectedHandler := mid.RequireRole("Manager")(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/protected", nil)
	req.Header.Set("X-User-ID", "1")
	w := httptest.NewRecorder()

	protectedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	if response["message"] != "Role 'Manager' required" {
		t.Errorf("Expected message about required role, got '%v'", response["message"])
	}
}

func TestSetUserIDInContext_ValidUserID(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(middleware.UserIDKey)
		if userID == nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "user ID not found in context"})
			return
		}

		userIDInt, ok := userID.(int64)
		if !ok || userIDInt != 123 {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid user ID type"})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"user_id": userIDInt})
	})

	setUserMiddleware := middleware.SetUserIDInContext()
	protectedHandler := setUserMiddleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("X-User-ID", "123")
	w := httptest.NewRecorder()

	protectedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	if response["user_id"] != float64(123) { // JSON numbers are float64
		t.Errorf("Expected user_id 123, got '%v'", response["user_id"])
	}
}

func TestSetUserIDInContext_NoUserIDHeader(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(middleware.UserIDKey)
		if userID != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "user ID should not be in context"})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "no user ID in context"})
	})

	setUserMiddleware := middleware.SetUserIDInContext()
	protectedHandler := setUserMiddleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	// No X-User-ID header
	w := httptest.NewRecorder()

	protectedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestResourceFromPath(t *testing.T) {
	tests := []struct {
		path           string
		expectedResult string
	}{
		{"/api/employee-records", "employee_records"},
		{"/api/budget-report", "budget_report"},
		{"/employee-records", "employee_records"},
		{"/api/v1/user-profiles", "user_profiles"},
		{"/", "unknown"},
	}

	for _, tt := range tests {
		req := httptest.NewRequest(http.MethodGet, tt.path, nil)
		result := middleware.ResourceFromPath(req)
		if result != tt.expectedResult {
			t.Errorf("ResourceFromPath(%s) = %s, expected %s", tt.path, result, tt.expectedResult)
		}
	}

	// Test empty path separately
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.URL.Path = ""
	result := middleware.ResourceFromPath(req)
	if result != "unknown" {
		t.Errorf("ResourceFromPath(empty) = %s, expected unknown", result)
	}
}

func TestActionFromMethod(t *testing.T) {
	tests := []struct {
		method         string
		expectedResult string
	}{
		{http.MethodGet, "read"},
		{http.MethodPost, "write"},
		{http.MethodPut, "write"},
		{http.MethodDelete, "delete"},
		{http.MethodPatch, "unknown"},
	}

	for _, tt := range tests {
		req := httptest.NewRequest(tt.method, "/api/test", nil)
		result := middleware.ActionFromMethod(req)
		if result != tt.expectedResult {
			t.Errorf("ActionFromMethod(%s) = %s, expected %s", tt.method, result, tt.expectedResult)
		}
	}
}

func TestRequirePermission_ChainedMiddleware(t *testing.T) {
	mockService := &MockAuthAuthorizationService{
		AuthorizeFunc: func(ctx context.Context, req model.AuthorizeRequest) (model.AuthorizeResponse, error) {
			return model.AuthorizeResponse{Allowed: true, Reason: "access granted"}, nil
		},
	}

	mid := middleware.NewAuthorizationMiddleware(mockService)
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(middleware.UserIDKey)
		if userID == nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "success",
			"user_id": userID,
		})
	})

	// Chain middlewares: SetUserIDInContext -> RequirePermission
	protectedHandler := middleware.SetUserIDInContext()(
		mid.RequirePermission("employee_records", "read")(testHandler),
	)

	req := httptest.NewRequest(http.MethodGet, "/api/employee-records", nil)
	req.Header.Set("X-User-ID", "42")
	w := httptest.NewRecorder()

	protectedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	if response["message"] != "success" {
		t.Errorf("Expected message 'success', got '%v'", response["message"])
	}

	if response["user_id"] != float64(42) {
		t.Errorf("Expected user_id 42, got '%v'", response["user_id"])
	}
}
