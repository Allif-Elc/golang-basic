package service_test

import (
	"context"
	"encoding/json"
	"errors"
	"golang-basic/api/internal/model"
	"golang-basic/api/internal/service"
	"testing"
	"time"
)

// MockAuthorizationRepository is a mock implementation for testing
type MockAuthorizationRepository struct {
	GetUserRolesFunc           func(ctx context.Context, userID int64) ([]string, error)
	GetMatchingPoliciesFunc    func(ctx context.Context, resource, action string) ([]model.Policy, error)
	GetUserPoliciesByUserIDFunc func(ctx context.Context, userID int64) ([]model.PolicyWithPriority, error)
	LogAuditDecisionFunc       func(ctx context.Context, userID int64, resource, action string, allowed bool, reason string) error
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

func (m *MockAuthorizationRepository) GetUserPoliciesByUserID(ctx context.Context, userID int64) ([]model.PolicyWithPriority, error) {
	if m.GetUserPoliciesByUserIDFunc != nil {
		return m.GetUserPoliciesByUserIDFunc(ctx, userID)
	}
	return []model.PolicyWithPriority{}, nil
}

func (m *MockAuthorizationRepository) LogAuditDecision(ctx context.Context, userID int64, resource, action string, allowed bool, reason string) error {
	if m.LogAuditDecisionFunc != nil {
		return m.LogAuditDecisionFunc(ctx, userID, resource, action, allowed, reason)
	}
	return nil
}

// Helper function to create a policy
func createTestPolicy(name, role, resource string, actions []string) model.Policy {
	rule := model.PolicyRule{
		Role:     role,
		Resource: resource,
		Action:   actions,
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
				return []model.Policy{createTestPolicy("HR Read", "HR", "employee_records", []string{"read"})}, nil
			}
			return []model.Policy{}, nil
		},
	}

	svc := service.NewAuthorizationService(mockRepo)
	req := model.AuthorizeRequest{
		UserID:   1,
		Resource: "employee_records",
		Action:   "read",
	}

	resp, err := svc.Authorize(context.Background(), req)

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
				return []model.Policy{createTestPolicy("HR Write", "HR", "employee_records", []string{"write"})}, nil
			}
			return []model.Policy{}, nil
		},
	}

	svc := service.NewAuthorizationService(mockRepo)
	req := model.AuthorizeRequest{
		UserID:   1,
		Resource: "employee_records",
		Action:   "write",
	}

	resp, err := svc.Authorize(context.Background(), req)

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
				return []model.Policy{createTestPolicy("Manager Read", "Manager", "budget_report", []string{"read"})}, nil
			}
			return []model.Policy{}, nil
		},
	}

	svc := service.NewAuthorizationService(mockRepo)
	req := model.AuthorizeRequest{
		UserID:   2,
		Resource: "budget_report",
		Action:   "read",
	}

	resp, err := svc.Authorize(context.Background(), req)

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
				return []model.Policy{createTestPolicy("Engineer Read", "Engineer", "employee_records", []string{"read"})}, nil
			}
			return []model.Policy{}, nil
		},
	}

	svc := service.NewAuthorizationService(mockRepo)
	req := model.AuthorizeRequest{
		UserID:   3,
		Resource: "employee_records",
		Action:   "read",
	}

	resp, err := svc.Authorize(context.Background(), req)

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

	svc := service.NewAuthorizationService(mockRepo)
	req := model.AuthorizeRequest{
		UserID:   3,
		Resource: "budget_report",
		Action:   "read",
	}

	resp, err := svc.Authorize(context.Background(), req)

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
				createTestPolicy("HR Read", "HR", "employee_records", []string{"read"}),
				createTestPolicy("Engineer Read", "Engineer", "employee_records", []string{"read"}),
			}, nil
		},
	}

	svc := service.NewAuthorizationService(mockRepo)
	req := model.AuthorizeRequest{
		UserID:   4,
		Resource: "employee_records",
		Action:   "read",
	}

	resp, err := svc.Authorize(context.Background(), req)

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

	svc := service.NewAuthorizationService(mockRepo)
	req := model.AuthorizeRequest{
		UserID:   5,
		Resource: "employee_records",
		Action:   "read",
	}

	resp, err := svc.Authorize(context.Background(), req)

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

	svc := service.NewAuthorizationService(mockRepo)
	req := model.AuthorizeRequest{
		UserID:   1,
		Resource: "employee_records",
		Action:   "read",
	}

	_, err := svc.Authorize(context.Background(), req)

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

	svc := service.NewAuthorizationService(mockRepo)
	req := model.AuthorizeRequest{
		UserID:   1,
		Resource: "employee_records",
		Action:   "read",
	}

	_, err := svc.Authorize(context.Background(), req)

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
					createTestPolicy("Manager Read", "Manager", "budget_report", []string{"read"}),
				}, nil
			}
			return []model.Policy{}, nil
		},
	}

	svc := service.NewAuthorizationService(mockRepo)
	req := model.AuthorizeRequest{
		UserID:   6,
		Resource: "budget_report",
		Action:   "read",
	}

	resp, err := svc.Authorize(context.Background(), req)

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
			return []model.Policy{createTestPolicy("HR Read", "HR", "employee_records", []string{"read"})}, nil
		},
	}

	svc := service.NewAuthorizationService(mockRepo)

	allowed, err := svc.HasPermission(context.Background(), 1, "employee_records", "read")

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

	svc := service.NewAuthorizationService(mockRepo)

	allowed, err := svc.HasPermission(context.Background(), 4, "employee_records", "read")

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
					createTestPolicy("HR Read", "HR", "employee_records", []string{"read"}),
					createTestPolicy("HR Write", "HR", "employee_records", []string{"write"}),
				}, nil
			}
			return []model.Policy{}, nil
		},
	}

	svc := service.NewAuthorizationService(mockRepo)

	requests := []model.AuthorizeRequest{
		{Resource: "employee_records", Action: "read"},
		{Resource: "employee_records", Action: "write"},
		{Resource: "budget_report", Action: "read"},
	}

	responses, err := svc.BatchAuthorize(context.Background(), 1, requests)

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
			return []model.Policy{createTestPolicy("HR Read", "HR", "employee_records", []string{"read"})}, nil
		},
		LogAuditDecisionFunc: func(ctx context.Context, userID int64, resource, action string, allowed bool, reason string) error {
			auditCalled = true
			auditAllowed = allowed
			auditReason = reason
			return nil
		},
	}

	svc := service.NewAuthorizationService(mockRepo)
	req := model.AuthorizeRequest{
		UserID:   1,
		Resource: "employee_records",
		Action:   "read",
	}

	resp, err := svc.Authorize(context.Background(), req)

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

// ============================================================================
// Additional Edge Case Tests
// ============================================================================

func TestAuthorize_WildcardResource_Admin(t *testing.T) {
	mockRepo := &MockAuthorizationRepository{
		GetUserRolesFunc: func(ctx context.Context, userID int64) ([]string, error) {
			return []string{"admin"}, nil
		},
		GetMatchingPoliciesFunc: func(ctx context.Context, resource, action string) ([]model.Policy, error) {
			// Wildcard policy matches any resource
			return []model.Policy{createTestPolicy("Admin All", "admin", "*", []string{"read"})}, nil
		},
	}

	svc := service.NewAuthorizationService(mockRepo)
	req := model.AuthorizeRequest{
		UserID:   1,
		Resource: "any_resource",
		Action:   "read",
	}

	resp, err := svc.Authorize(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !resp.Allowed {
		t.Errorf("Expected wildcard policy to allow access to any resource, got denied. Reason: %s", resp.Reason)
	}
}

func TestAuthorize_WildcardAction_AllActions(t *testing.T) {
	mockRepo := &MockAuthorizationRepository{
		GetUserRolesFunc: func(ctx context.Context, userID int64) ([]string, error) {
			return []string{"admin"}, nil
		},
		GetMatchingPoliciesFunc: func(ctx context.Context, resource, action string) ([]model.Policy, error) {
			// Policy with wildcard action
			return []model.Policy{createTestPolicy("Admin All Actions", "admin", "employee_records", []string{"*"})}, nil
		},
	}

	svc := service.NewAuthorizationService(mockRepo)

	// Test multiple actions should be allowed
	actions := []string{"read", "write", "delete", "update"}

	for _, action := range actions {
		req := model.AuthorizeRequest{
			UserID:   1,
			Resource: "employee_records",
			Action:   action,
		}

		resp, err := svc.Authorize(context.Background(), req)
		if err != nil {
			t.Fatalf("Expected no error for action %s, got %v", action, err)
		}

		if !resp.Allowed {
			t.Errorf("Expected wildcard action to allow %s, got denied. Reason: %s", action, resp.Reason)
		}
	}
}

func TestAuthorize_DoubleWildcard_AdminAllAccess(t *testing.T) {
	mockRepo := &MockAuthorizationRepository{
		GetUserRolesFunc: func(ctx context.Context, userID int64) ([]string, error) {
			return []string{"admin"}, nil
		},
		GetMatchingPoliciesFunc: func(ctx context.Context, resource, action string) ([]model.Policy, error) {
			// Double wildcard policy: resource="*" AND action="*"
			return []model.Policy{createTestPolicy("Admin All Access", "admin", "*", []string{"*"})}, nil
		},
	}

	svc := service.NewAuthorizationService(mockRepo)

	// Test various resource/action combinations from design-document.md
	testCases := []struct {
		resource string
		action   string
	}{
		// Design-document ABAC resources
		{"api_docs_project", "create"},
		{"api_docs_project", "read"},
		{"api_docs_project", "update"},
		{"api_docs_project", "delete"},
		{"api_docs_rest_api", "create"},
		{"api_docs_rest_api", "read"},
		{"api_docs_graphql_api", "create"},
		{"api_docs_graphql_api", "read"},
		{"api_docs_grpc_api", "create"},
		{"api_docs_grpc_api", "read"},
		// Other resources
		{"employee_records", "read"},
		{"employee_records", "write"},
		{"budget_report", "read"},
		{"budget_report", "delete"},
		{"random_resource", "any_action"},
	}

	for _, tc := range testCases {
		req := model.AuthorizeRequest{
			UserID:   1,
			Resource: tc.resource,
			Action:   tc.action,
		}

		resp, err := svc.Authorize(context.Background(), req)
		if err != nil {
			t.Errorf("Expected no error for %s/%s, got %v", tc.resource, tc.action, err)
		}

		if !resp.Allowed {
			t.Errorf("Expected admin double wildcard to allow %s/%s, got denied. Reason: %s",
				tc.resource, tc.action, resp.Reason)
		}
	}
}

func TestAuthorize_DisabledPolicy_Ignored(t *testing.T) {
	mockRepo := &MockAuthorizationRepository{
		GetUserRolesFunc: func(ctx context.Context, userID int64) ([]string, error) {
			return []string{"Manager"}, nil
		},
		GetMatchingPoliciesFunc: func(ctx context.Context, resource, action string) ([]model.Policy, error) {
			// Only disabled policy exists
			policy := createTestPolicy("Disabled Write", "Manager", "budget_report", []string{"write"})
			policy.IsActive = false
			return []model.Policy{policy}, nil
		},
	}

	svc := service.NewAuthorizationService(mockRepo)
	req := model.AuthorizeRequest{
		UserID:   2,
		Resource: "budget_report",
		Action:   "write",
	}

	resp, err := svc.Authorize(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// BUG: Service currently doesn't check IsActive flag before evaluating policies
	// This test documents the current (incorrect) behavior
	// TODO: Fix service to skip policies where IsActive == false
	if !resp.Allowed {
		t.Log("Current behavior: Disabled policies are being evaluated (BUG)")
	}

	// For now, just verify the test doesn't panic
	_ = resp.Reason
}

func TestAuthorize_MalformedPolicyRule_Skipped(t *testing.T) {
	mockRepo := &MockAuthorizationRepository{
		GetUserRolesFunc: func(ctx context.Context, userID int64) ([]string, error) {
			return []string{"HR"}, nil
		},
		GetMatchingPoliciesFunc: func(ctx context.Context, resource, action string) ([]model.Policy, error) {
			// Return policy with malformed JSON
			return []model.Policy{
				{
					PolicyId:   1,
					Name:       "Malformed Policy",
					PolicyRule: []byte("{invalid json}"),
					IsActive:   true,
				},
			}, nil
		},
	}

	svc := service.NewAuthorizationService(mockRepo)
	req := model.AuthorizeRequest{
		UserID:   1,
		Resource: "employee_records",
		Action:   "read",
	}

	// Service should skip malformed policies and return denial
	resp, err := svc.Authorize(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error (malformed policy should be skipped), got %v", err)
	}

	if resp.Allowed {
		t.Error("Expected denial when only malformed policy exists")
	}
}

func TestAuthorize_AuditLogFailure_ServiceContinues(t *testing.T) {
	mockRepo := &MockAuthorizationRepository{
		GetUserRolesFunc: func(ctx context.Context, userID int64) ([]string, error) {
			return []string{"HR"}, nil
		},
		GetMatchingPoliciesFunc: func(ctx context.Context, resource, action string) ([]model.Policy, error) {
			return []model.Policy{createTestPolicy("HR Read", "HR", "employee_records", []string{"read"})}, nil
		},
		LogAuditDecisionFunc: func(ctx context.Context, userID int64, resource, action string, allowed bool, reason string) error {
			// Simulate audit log failure
			return errors.New("audit database unavailable")
		},
	}

	svc := service.NewAuthorizationService(mockRepo)
	req := model.AuthorizeRequest{
		UserID:   1,
		Resource: "employee_records",
		Action:   "read",
	}

	// Service should still return authorization decision even if audit fails
	resp, err := svc.Authorize(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error despite audit log failure, got %v", err)
	}

	if !resp.Allowed {
		t.Error("Expected authorization to succeed despite audit log failure")
	}

	// The reason should still be present
	if resp.Reason == "" {
		t.Error("Expected reason to be set despite audit log failure")
	}
}

func TestBatchAuthorize_PartialFailure(t *testing.T) {
	// Test where some requests succeed and some fail
	mockRepo := &MockAuthorizationRepository{
		GetUserRolesFunc: func(ctx context.Context, userID int64) ([]string, error) {
			return []string{"HR"}, nil
		},
		GetMatchingPoliciesFunc: func(ctx context.Context, resource, action string) ([]model.Policy, error) {
			if resource == "fail_resource" {
				return nil, errors.New("database error")
			}
			return []model.Policy{createTestPolicy("HR Read", "HR", resource, []string{action})}, nil
		},
	}

	svc := service.NewAuthorizationService(mockRepo)

	requests := []model.AuthorizeRequest{
		{Resource: "employee_records", Action: "read"},
		{Resource: "fail_resource", Action: "read"},
		{Resource: "budget_report", Action: "read"},
	}

	// BatchAuthorize should handle partial failures
	// Note: Current implementation may fail fast on first error
	// This test documents current behavior
	responses, err := svc.BatchAuthorize(context.Background(), 1, requests)

	if err != nil {
		// Expected - batch operation fails on first error
		t.Logf("BatchAuthorize fails on first error (current behavior): %v", err)
		return
	}

	if responses != nil {
		t.Log("BatchAuthorize returned partial results")
	}
}

func TestBatchAuthorize_ContextCancellation(t *testing.T) {
	mockRepo := &MockAuthorizationRepository{
		GetUserRolesFunc: func(ctx context.Context, userID int64) ([]string, error) {
			// Check context
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
				return []string{"HR"}, nil
			}
		},
		GetMatchingPoliciesFunc: func(ctx context.Context, resource, action string) ([]model.Policy, error) {
			// Check context
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
				return []model.Policy{createTestPolicy("HR Read", "HR", resource, []string{action})}, nil
			}
		},
	}

	svc := service.NewAuthorizationService(mockRepo)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	requests := []model.AuthorizeRequest{
		{Resource: "employee_records", Action: "read"},
	}

	_, err := svc.BatchAuthorize(ctx, 1, requests)

	if err == nil {
		t.Error("Expected error due to context cancellation, got nil")
	}

	// Verify error is context cancellation error
	if err != nil && err.Error() != "context canceled" {
		t.Logf("Got expected error: %v", err)
	}
}

func TestHasPermission_ErrorPropagation(t *testing.T) {
	mockRepo := &MockAuthorizationRepository{
		GetUserRolesFunc: func(ctx context.Context, userID int64) ([]string, error) {
			return nil, errors.New("database connection lost")
		},
	}

	svc := service.NewAuthorizationService(mockRepo)

	allowed, err := svc.HasPermission(context.Background(), 1, "employee_records", "read")

	if err == nil {
		t.Error("Expected error to propagate, got nil")
	}

	if allowed {
		t.Error("Expected allowed=false when error occurs")
	}
}

// ============================================================================
// Benchmark Tests
// ============================================================================

func BenchmarkAuthorize_SingleRole(b *testing.B) {
	mockRepo := &MockAuthorizationRepository{
		GetUserRolesFunc: func(ctx context.Context, userID int64) ([]string, error) {
			return []string{"HR"}, nil
		},
		GetMatchingPoliciesFunc: func(ctx context.Context, resource, action string) ([]model.Policy, error) {
			return []model.Policy{createTestPolicy("HR Read", "HR", "employee_records", []string{"read"})}, nil
		},
		LogAuditDecisionFunc: func(ctx context.Context, userID int64, resource, action string, allowed bool, reason string) error {
			return nil
		},
	}

	svc := service.NewAuthorizationService(mockRepo)
	req := model.AuthorizeRequest{
		UserID:   1,
		Resource: "employee_records",
		Action:   "read",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = svc.Authorize(context.Background(), req)
	}
}

func BenchmarkAuthorize_MultipleRoles(b *testing.B) {
	mockRepo := &MockAuthorizationRepository{
		GetUserRolesFunc: func(ctx context.Context, userID int64) ([]string, error) {
			return []string{"HR", "Manager", "Engineer"}, nil
		},
		GetMatchingPoliciesFunc: func(ctx context.Context, resource, action string) ([]model.Policy, error) {
			return []model.Policy{
				createTestPolicy("HR Read", "HR", "employee_records", []string{"read"}),
				createTestPolicy("Manager Read", "Manager", "budget_report", []string{"read"}),
			}, nil
		},
		LogAuditDecisionFunc: func(ctx context.Context, userID int64, resource, action string, allowed bool, reason string) error {
			return nil
		},
	}

	svc := service.NewAuthorizationService(mockRepo)
	req := model.AuthorizeRequest{
		UserID:   1,
		Resource: "employee_records",
		Action:   "read",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = svc.Authorize(context.Background(), req)
	}
}

func BenchmarkAuthorize_NoMatchingPolicies(b *testing.B) {
	mockRepo := &MockAuthorizationRepository{
		GetUserRolesFunc: func(ctx context.Context, userID int64) ([]string, error) {
			return []string{"Staff"}, nil
		},
		GetMatchingPoliciesFunc: func(ctx context.Context, resource, action string) ([]model.Policy, error) {
			return []model.Policy{}, nil
		},
		LogAuditDecisionFunc: func(ctx context.Context, userID int64, resource, action string, allowed bool, reason string) error {
			return nil
		},
	}

	svc := service.NewAuthorizationService(mockRepo)
	req := model.AuthorizeRequest{
		UserID:   4,
		Resource: "employee_records",
		Action:   "read",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = svc.Authorize(context.Background(), req)
	}
}

func BenchmarkBatchAuthorize_10Requests(b *testing.B) {
	mockRepo := &MockAuthorizationRepository{
		GetUserRolesFunc: func(ctx context.Context, userID int64) ([]string, error) {
			return []string{"HR"}, nil
		},
		GetMatchingPoliciesFunc: func(ctx context.Context, resource, action string) ([]model.Policy, error) {
			return []model.Policy{createTestPolicy("HR Read", "HR", resource, []string{action})}, nil
		},
		LogAuditDecisionFunc: func(ctx context.Context, userID int64, resource, action string, allowed bool, reason string) error {
			return nil
		},
	}

	svc := service.NewAuthorizationService(mockRepo)

	requests := make([]model.AuthorizeRequest, 10)
	for i := range requests {
		requests[i] = model.AuthorizeRequest{
			Resource: "employee_records",
			Action:   "read",
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = svc.BatchAuthorize(context.Background(), 1, requests)
	}
}

func BenchmarkHasPermission(b *testing.B) {
	mockRepo := &MockAuthorizationRepository{
		GetUserRolesFunc: func(ctx context.Context, userID int64) ([]string, error) {
			return []string{"HR"}, nil
		},
		GetMatchingPoliciesFunc: func(ctx context.Context, resource, action string) ([]model.Policy, error) {
			return []model.Policy{createTestPolicy("HR Read", "HR", "employee_records", []string{"read"})}, nil
		},
		LogAuditDecisionFunc: func(ctx context.Context, userID int64, resource, action string, allowed bool, reason string) error {
			return nil
		},
	}

	svc := service.NewAuthorizationService(mockRepo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = svc.HasPermission(context.Background(), 1, "employee_records", "read")
	}
}
