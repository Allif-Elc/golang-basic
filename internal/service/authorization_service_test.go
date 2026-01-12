package service

import (
	"context"
	"encoding/json"
	"errors"
	"golang-basic/internal/model"
	"testing"
	"time"
)

// MockAuthorizationRepository is a mock implementation for testing
type MockAuthorizationRepository struct {
	GetUserRolesFunc        func(ctx context.Context, userID int64) ([]string, error)
	GetMatchingPoliciesFunc func(ctx context.Context, resource, action string) ([]model.Policy, error)
	LogAuditDecisionFunc    func(ctx context.Context, userID int64, resource, action string, allowed bool, reason string) error
}

func (m *MockAuthorizationRepository) GetUserRoles(ctx context.Context, userID int64) ([]string, error) {
	if m.GetUserRolesFunc != nil {
		return m.GetUserRolesFunc(ctx, userID)
	}
	return []string{}, nil
}

func (m *MockAuthorizationRepository) GetMatchingPolicies(ctx context.Context, resource, action string) ([]model.Policy, error) {
	if m.GetMatchingPoliciesFunc != nil {
		return m.GetMatchingPoliciesFunc(ctx, resource, action)
	}
	return []model.Policy{}, nil
}

func (m *MockAuthorizationRepository) LogAuditDecision(ctx context.Context, userID int64, resource, action string, allowed bool, reason string) error {
	if m.LogAuditDecisionFunc != nil {
		return m.LogAuditDecisionFunc(ctx, userID, resource, action, allowed, reason)
	}
	return nil
}

// Helper function to create a policy
func createTestPolicy(name, role, resource, action string) model.Policy {
	rule := model.PolicyRule{
		Role:     role,
		Resource: resource,
		Action:   action,
	}
	ruleJSON, _ := json.Marshal(rule)

	return model.Policy{
		PolicyId:   1,
		Name:       name,
		PolicyRule: ruleJSON,
		IsActive:   true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

func TestAuthorize_HR_AllowedReadEmployeeRecords(t *testing.T) {
	mockRepo := &MockAuthorizationRepository{
		GetUserRolesFunc: func(ctx context.Context, userID int64) ([]string, error) {
			return []string{"HR"}, nil
		},
		GetMatchingPoliciesFunc: func(ctx context.Context, resource, action string) ([]model.Policy, error) {
			if resource == "employee_records" && action == "read" {
				return []model.Policy{createTestPolicy("HR Read", "HR", "employee_records", "read")}, nil
			}
			return []model.Policy{}, nil
		},
	}

	service := NewAuthorizationService(mockRepo)
	req := model.AuthorizeRequest{
		UserID:   1,
		Resource: "employee_records",
		Action:   "read",
	}

	resp, err := service.Authorize(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !resp.Allowed {
		t.Errorf("Expected allowed=true, got false. Reason: %s", resp.Reason)
	}
}

func TestAuthorize_HR_AllowedWriteEmployeeRecords(t *testing.T) {
	mockRepo := &MockAuthorizationRepository{
		GetUserRolesFunc: func(ctx context.Context, userID int64) ([]string, error) {
			return []string{"HR"}, nil
		},
		GetMatchingPoliciesFunc: func(ctx context.Context, resource, action string) ([]model.Policy, error) {
			if resource == "employee_records" && action == "write" {
				return []model.Policy{createTestPolicy("HR Write", "HR", "employee_records", "write")}, nil
			}
			return []model.Policy{}, nil
		},
	}

	service := NewAuthorizationService(mockRepo)
	req := model.AuthorizeRequest{
		UserID:   1,
		Resource: "employee_records",
		Action:   "write",
	}

	resp, err := service.Authorize(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !resp.Allowed {
		t.Errorf("Expected allowed=true, got false. Reason: %s", resp.Reason)
	}
}

func TestAuthorize_Manager_AllowedReadBudgetReport(t *testing.T) {
	mockRepo := &MockAuthorizationRepository{
		GetUserRolesFunc: func(ctx context.Context, userID int64) ([]string, error) {
			return []string{"Manager"}, nil
		},
		GetMatchingPoliciesFunc: func(ctx context.Context, resource, action string) ([]model.Policy, error) {
			if resource == "budget_report" && action == "read" {
				return []model.Policy{createTestPolicy("Manager Read", "Manager", "budget_report", "read")}, nil
			}
			return []model.Policy{}, nil
		},
	}

	service := NewAuthorizationService(mockRepo)
	req := model.AuthorizeRequest{
		UserID:   2,
		Resource: "budget_report",
		Action:   "read",
	}

	resp, err := service.Authorize(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !resp.Allowed {
		t.Errorf("Expected allowed=true, got false. Reason: %s", resp.Reason)
	}
}

func TestAuthorize_Engineer_AllowedReadEmployeeRecords(t *testing.T) {
	mockRepo := &MockAuthorizationRepository{
		GetUserRolesFunc: func(ctx context.Context, userID int64) ([]string, error) {
			return []string{"Engineer"}, nil
		},
		GetMatchingPoliciesFunc: func(ctx context.Context, resource, action string) ([]model.Policy, error) {
			if resource == "employee_records" && action == "read" {
				return []model.Policy{createTestPolicy("Engineer Read", "Engineer", "employee_records", "read")}, nil
			}
			return []model.Policy{}, nil
		},
	}

	service := NewAuthorizationService(mockRepo)
	req := model.AuthorizeRequest{
		UserID:   3,
		Resource: "employee_records",
		Action:   "read",
	}

	resp, err := service.Authorize(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !resp.Allowed {
		t.Errorf("Expected allowed=true, got false. Reason: %s", resp.Reason)
	}
}

func TestAuthorize_Engineer_DeniedBudgetReport(t *testing.T) {
	mockRepo := &MockAuthorizationRepository{
		GetUserRolesFunc: func(ctx context.Context, userID int64) ([]string, error) {
			return []string{"Engineer"}, nil
		},
		GetMatchingPoliciesFunc: func(ctx context.Context, resource, action string) ([]model.Policy, error) {
			// No policies for Engineer accessing budget_report
			return []model.Policy{}, nil
		},
	}

	service := NewAuthorizationService(mockRepo)
	req := model.AuthorizeRequest{
		UserID:   3,
		Resource: "budget_report",
		Action:   "read",
	}

	resp, err := service.Authorize(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.Allowed {
		t.Errorf("Expected allowed=false, got true")
	}

	expectedReason := "no policies found for resource/action"
	if resp.Reason != expectedReason {
		t.Errorf("Expected reason '%s', got '%s'", expectedReason, resp.Reason)
	}
}

func TestAuthorize_Staff_DeniedEmployeeRecords(t *testing.T) {
	mockRepo := &MockAuthorizationRepository{
		GetUserRolesFunc: func(ctx context.Context, userID int64) ([]string, error) {
			return []string{"Staff"}, nil
		},
		GetMatchingPoliciesFunc: func(ctx context.Context, resource, action string) ([]model.Policy, error) {
			// Policies exist but none for Staff role
			return []model.Policy{
				createTestPolicy("HR Read", "HR", "employee_records", "read"),
				createTestPolicy("Engineer Read", "Engineer", "employee_records", "read"),
			}, nil
		},
	}

	service := NewAuthorizationService(mockRepo)
	req := model.AuthorizeRequest{
		UserID:   4,
		Resource: "employee_records",
		Action:   "read",
	}

	resp, err := service.Authorize(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.Allowed {
		t.Errorf("Expected allowed=false, got true")
	}
}

func TestAuthorize_NoRolesAssigned(t *testing.T) {
	mockRepo := &MockAuthorizationRepository{
		GetUserRolesFunc: func(ctx context.Context, userID int64) ([]string, error) {
			return []string{}, nil // No roles
		},
		GetMatchingPoliciesFunc: func(ctx context.Context, resource, action string) ([]model.Policy, error) {
			return []model.Policy{}, nil
		},
	}

	service := NewAuthorizationService(mockRepo)
	req := model.AuthorizeRequest{
		UserID:   5,
		Resource: "employee_records",
		Action:   "read",
	}

	resp, err := service.Authorize(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.Allowed {
		t.Errorf("Expected allowed=false, got true")
	}

	expectedReason := "no roles assigned to user"
	if resp.Reason != expectedReason {
		t.Errorf("Expected reason '%s', got '%s'", expectedReason, resp.Reason)
	}
}

func TestAuthorize_UserRolesError(t *testing.T) {
	mockRepo := &MockAuthorizationRepository{
		GetUserRolesFunc: func(ctx context.Context, userID int64) ([]string, error) {
			return nil, errors.New("database connection failed")
		},
	}

	service := NewAuthorizationService(mockRepo)
	req := model.AuthorizeRequest{
		UserID:   1,
		Resource: "employee_records",
		Action:   "read",
	}

	_, err := service.Authorize(context.Background(), req)

	if err == nil {
		t.Fatalf("Expected error, got nil")
	}

	expectedError := "failed to get user roles"
	if !containsString(err.Error(), expectedError) {
		t.Errorf("Expected error to contain '%s', got '%s'", expectedError, err.Error())
	}
}

func TestAuthorize_PoliciesError(t *testing.T) {
	mockRepo := &MockAuthorizationRepository{
		GetUserRolesFunc: func(ctx context.Context, userID int64) ([]string, error) {
			return []string{"HR"}, nil
		},
		GetMatchingPoliciesFunc: func(ctx context.Context, resource, action string) ([]model.Policy, error) {
			return nil, errors.New("failed to query policies")
		},
	}

	service := NewAuthorizationService(mockRepo)
	req := model.AuthorizeRequest{
		UserID:   1,
		Resource: "employee_records",
		Action:   "read",
	}

	_, err := service.Authorize(context.Background(), req)

	if err == nil {
		t.Fatalf("Expected error, got nil")
	}

	expectedError := "failed to fetch policies"
	if !containsString(err.Error(), expectedError) {
		t.Errorf("Expected error to contain '%s', got '%s'", expectedError, err.Error())
	}
}

func TestAuthorize_MultipleRoles_Allowed(t *testing.T) {
	mockRepo := &MockAuthorizationRepository{
		GetUserRolesFunc: func(ctx context.Context, userID int64) ([]string, error) {
			return []string{"Engineer", "Manager"}, nil
		},
		GetMatchingPoliciesFunc: func(ctx context.Context, resource, action string) ([]model.Policy, error) {
			if resource == "budget_report" && action == "read" {
				return []model.Policy{
					createTestPolicy("Manager Read", "Manager", "budget_report", "read"),
				}, nil
			}
			return []model.Policy{}, nil
		},
	}

	service := NewAuthorizationService(mockRepo)
	req := model.AuthorizeRequest{
		UserID:   6,
		Resource: "budget_report",
		Action:   "read",
	}

	resp, err := service.Authorize(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !resp.Allowed {
		t.Errorf("Expected allowed=true, got false. Reason: %s", resp.Reason)
	}
}

func TestHasPermission_Allowed(t *testing.T) {
	mockRepo := &MockAuthorizationRepository{
		GetUserRolesFunc: func(ctx context.Context, userID int64) ([]string, error) {
			return []string{"HR"}, nil
		},
		GetMatchingPoliciesFunc: func(ctx context.Context, resource, action string) ([]model.Policy, error) {
			return []model.Policy{createTestPolicy("HR Read", "HR", "employee_records", "read")}, nil
		},
	}

	service := NewAuthorizationService(mockRepo)

	allowed, err := service.HasPermission(context.Background(), 1, "employee_records", "read")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !allowed {
		t.Errorf("Expected allowed=true, got false")
	}
}

func TestHasPermission_Denied(t *testing.T) {
	mockRepo := &MockAuthorizationRepository{
		GetUserRolesFunc: func(ctx context.Context, userID int64) ([]string, error) {
			return []string{"Staff"}, nil
		},
		GetMatchingPoliciesFunc: func(ctx context.Context, resource, action string) ([]model.Policy, error) {
			return []model.Policy{}, nil
		},
	}

	service := NewAuthorizationService(mockRepo)

	allowed, err := service.HasPermission(context.Background(), 4, "employee_records", "read")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if allowed {
		t.Errorf("Expected allowed=false, got true")
	}
}

func TestBatchAuthorize_MultipleRequests(t *testing.T) {
	mockRepo := &MockAuthorizationRepository{
		GetUserRolesFunc: func(ctx context.Context, userID int64) ([]string, error) {
			return []string{"HR"}, nil
		},
		GetMatchingPoliciesFunc: func(ctx context.Context, resource, action string) ([]model.Policy, error) {
			if resource == "employee_records" {
				return []model.Policy{
					createTestPolicy("HR Read", "HR", "employee_records", "read"),
					createTestPolicy("HR Write", "HR", "employee_records", "write"),
				}, nil
			}
			return []model.Policy{}, nil
		},
	}

	service := NewAuthorizationService(mockRepo)

	requests := []model.AuthorizeRequest{
		{Resource: "employee_records", Action: "read"},
		{Resource: "employee_records", Action: "write"},
		{Resource: "budget_report", Action: "read"},
	}

	responses, err := service.BatchAuthorize(context.Background(), 1, requests)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(responses) != len(requests) {
		t.Fatalf("Expected %d responses, got %d", len(requests), len(responses))
	}

	if !responses[0].Allowed {
		t.Errorf("Expected first request to be allowed")
	}

	if !responses[1].Allowed {
		t.Errorf("Expected second request to be allowed")
	}

	if responses[2].Allowed {
		t.Errorf("Expected third request to be denied")
	}
}

func TestAuthorize_AuditLogCalled(t *testing.T) {
	auditCalled := false
	var auditAllowed bool
	var auditReason string

	mockRepo := &MockAuthorizationRepository{
		GetUserRolesFunc: func(ctx context.Context, userID int64) ([]string, error) {
			return []string{"HR"}, nil
		},
		GetMatchingPoliciesFunc: func(ctx context.Context, resource, action string) ([]model.Policy, error) {
			return []model.Policy{createTestPolicy("HR Read", "HR", "employee_records", "read")}, nil
		},
		LogAuditDecisionFunc: func(ctx context.Context, userID int64, resource, action string, allowed bool, reason string) error {
			auditCalled = true
			auditAllowed = allowed
			auditReason = reason
			return nil
		},
	}

	service := NewAuthorizationService(mockRepo)
	req := model.AuthorizeRequest{
		UserID:   1,
		Resource: "employee_records",
		Action:   "read",
	}

	resp, err := service.Authorize(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !auditCalled {
		t.Errorf("Expected audit log to be called")
	}

	if auditAllowed != resp.Allowed {
		t.Errorf("Expected audit allowed %v, got %v", resp.Allowed, auditAllowed)
	}

	if auditReason != resp.Reason {
		t.Errorf("Expected audit reason '%s', got '%s'", resp.Reason, auditReason)
	}
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && contains(s, substr))
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
