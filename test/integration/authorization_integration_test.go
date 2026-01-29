package integration_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"golang-basic/api/internal/model"
	"golang-basic/api/internal/repository"
	"golang-basic/api/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// SetupIntegration initializes the database, repository, and service for integration tests
// Returns the authorization service and a cleanup function
func SetupIntegration(t *testing.T) (*service.AuthorizationService, *pgxpool.Pool, func()) {
	t.Helper()

	ctx := context.Background()

	// Configure Podman
	testcontainersURL := os.Getenv("TESTCONTAINERS_PODMAN_SOCKET")
	if testcontainersURL == "" {
		os.Setenv("TESTCONTAINERS_DOCKER_SOCKET", "unix:///run/podman/podman.sock")
	} else {
		os.Setenv("TESTCONTAINERS_DOCKER_SOCKET", testcontainersURL)
	}

	// Get the SQL schema path
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	schemaPath := filepath.Join(wd, "../../sql/abac_schema.sql")
	schemaPath = filepath.Clean(schemaPath)

	// Start PostgreSQL container
	container, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:17-alpine"),
		postgres.WithDatabase("test_db"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		t.Skipf("Skipping test - Podman not available: %v", err)
		return nil, nil, func() {}
	}

	// Read schema file
	schemaSQL, err := os.ReadFile(schemaPath)
	if err != nil {
		t.Fatalf("Failed to read schema file: %v", err)
	}

	// Get connection string
	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("Failed to get connection string: %v", err)
	}

	// Create connection pool
	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		t.Fatalf("Failed to parse pool config: %v", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		t.Fatalf("Failed to create connection pool: %v", err)
	}

	// Execute schema
	_, err = pool.Exec(ctx, string(schemaSQL))
	if err != nil {
		t.Fatalf("Failed to execute schema: %v", err)
	}

	// Create repository and service
	repo := repository.NewAuthorizationRepository(pool)
	authService := service.NewAuthorizationService(repo)

	// Cleanup function
	cleanup := func() {
		pool.Close()
		if err := container.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate container: %v", err)
		}
	}

	return authService, pool, cleanup
}

// ============================================================================
// End-to-End Authorization Tests
// ============================================================================

func TestIntegration_HR_ReadEmployeeRecords_Allowed(t *testing.T) {
	authService, pool, cleanup := SetupIntegration(t)
	defer cleanup()

	ctx := context.Background()

	// Setup: User with HR role
	_, err := pool.Exec(ctx, `
		INSERT INTO user_attributes (id_user, id_attribute, value)
		VALUES ($1, $2, $3)
	`, 1, 1, "HR") // Using id_attribute=1 (role) from schema
	if err != nil {
		t.Fatalf("Failed to setup test data: %v", err)
	}

	// Setup: HR policy
	type PolicyRuleWithArray struct {
		Role     string   `json:"role"`
		Resource string   `json:"resource"`
		Action   []string `json:"action"`
	}
	rule := PolicyRuleWithArray{Role: "HR", Resource: "employee_records", Action: []string{"read", "write"}}
	ruleJSON, _ := json.Marshal(rule)
	_, err = pool.Exec(ctx, `
		INSERT INTO policies (name, policy_rule, is_active)
		VALUES ($1, $2, $3)
	`, "HR Policy", ruleJSON, true)
	if err != nil {
		t.Fatalf("Failed to insert policy: %v", err)
	}

	// Test
	req := model.AuthorizeRequest{
		UserID:   1,
		Resource: "employee_records",
		Action:   "read",
	}

	resp, err := authService.Authorize(ctx, req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !resp.Allowed {
		t.Errorf("Expected allowed=true, got false. Reason: %s", resp.Reason)
	}
}

func TestIntegration_Staff_ReadEmployeeRecords_Denied(t *testing.T) {
	authService, pool, cleanup := SetupIntegration(t)
	defer cleanup()

	ctx := context.Background()

	// Setup: User with Staff role (no matching policy)
	_, err := pool.Exec(ctx, `
		INSERT INTO user_attributes (id_user, id_attribute, value)
		VALUES ($1, $2, $3)
	`, 1, 1, "Staff")
	if err != nil {
		t.Fatalf("Failed to setup test data: %v", err)
	}

	// Setup: HR policy exists (doesn't match Staff)
	type PolicyRuleWithArray struct {
		Role     string   `json:"role"`
		Resource string   `json:"resource"`
		Action   []string `json:"action"`
	}
	rule := PolicyRuleWithArray{Role: "HR", Resource: "employee_records", Action: []string{"read"}}
	ruleJSON, _ := json.Marshal(rule)
	_, err = pool.Exec(ctx, `
		INSERT INTO policies (name, policy_rule, is_active)
		VALUES ($1, $2, $3)
	`, "HR Policy", ruleJSON, true)
	if err != nil {
		t.Fatalf("Failed to insert policy: %v", err)
	}

	// Test
	req := model.AuthorizeRequest{
		UserID:   1,
		Resource: "employee_records",
		Action:   "read",
	}

	resp, err := authService.Authorize(ctx, req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.Allowed {
		t.Error("Expected allowed=false, got true")
	}

	expectedReason := "no policies found for resource/action"
	if resp.Reason != expectedReason {
		t.Errorf("Expected reason '%s', got '%s'", expectedReason, resp.Reason)
	}
}

func TestIntegration_Manager_ReadBudgetReport_Allowed(t *testing.T) {
	authService, pool, cleanup := SetupIntegration(t)
	defer cleanup()

	ctx := context.Background()

	// Setup: Manager user
	_, err := pool.Exec(ctx, `
		INSERT INTO user_attributes (id_user, id_attribute, value)
		VALUES ($1, $2, $3)
	`, 2, 1, "Manager")
	if err != nil {
		t.Fatalf("Failed to setup test data: %v", err)
	}

	// Setup: Manager budget policy
	type PolicyRuleWithArray struct {
		Role     string   `json:"role"`
		Resource string   `json:"resource"`
		Action   []string `json:"action"`
	}
	rule := PolicyRuleWithArray{Role: "Manager", Resource: "budget_report", Action: []string{"read"}}
	ruleJSON, _ := json.Marshal(rule)
	_, err = pool.Exec(ctx, `
		INSERT INTO policies (name, policy_rule, is_active)
		VALUES ($1, $2, $3)
	`, "Manager Budget Policy", ruleJSON, true)
	if err != nil {
		t.Fatalf("Failed to insert policy: %v", err)
	}

	// Test
	req := model.AuthorizeRequest{
		UserID:   2,
		Resource: "budget_report",
		Action:   "read",
	}

	resp, err := authService.Authorize(ctx, req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !resp.Allowed {
		t.Errorf("Expected allowed=true, got false. Reason: %s", resp.Reason)
	}
}

func TestIntegration_Engineer_ReadBudgetReport_Denied(t *testing.T) {
	authService, pool, cleanup := SetupIntegration(t)
	defer cleanup()

	ctx := context.Background()

	// Setup: Engineer user
	_, err := pool.Exec(ctx, `
		INSERT INTO user_attributes (id_user, id_attribute, value)
		VALUES ($1, $2, $3)
	`, 3, 1, "Engineer")
	if err != nil {
		t.Fatalf("Failed to setup test data: %v", err)
	}

	// Setup: Only Manager policy exists
	type PolicyRuleWithArray struct {
		Role     string   `json:"role"`
		Resource string   `json:"resource"`
		Action   []string `json:"action"`
	}
	rule := PolicyRuleWithArray{Role: "Manager", Resource: "budget_report", Action: []string{"read"}}
	ruleJSON, _ := json.Marshal(rule)
	_, err = pool.Exec(ctx, `
		INSERT INTO policies (name, policy_rule, is_active)
		VALUES ($1, $2, $3)
	`, "Manager Budget Policy", ruleJSON, true)
	if err != nil {
		t.Fatalf("Failed to insert policy: %v", err)
	}

	// Test
	req := model.AuthorizeRequest{
		UserID:   3,
		Resource: "budget_report",
		Action:   "read",
	}

	resp, err := authService.Authorize(ctx, req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.Allowed {
		t.Error("Expected allowed=false, got true")
	}
}

func TestIntegration_NoRolesAssigned_Denied(t *testing.T) {
	authService, _, cleanup := SetupIntegration(t)
	defer cleanup()

	ctx := context.Background()

	// Test with user who has no roles assigned
	req := model.AuthorizeRequest{
		UserID:   999,
		Resource: "employee_records",
		Action:   "read",
	}

	resp, err := authService.Authorize(ctx, req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.Allowed {
		t.Error("Expected allowed=false, got true")
	}

	expectedReason := "no roles assigned to user"
	if resp.Reason != expectedReason {
		t.Errorf("Expected reason '%s', got '%s'", expectedReason, resp.Reason)
	}
}

// ============================================================================
// Audit Log Integration Tests
// ============================================================================

func TestIntegration_AuditLog_AllowedDecision(t *testing.T) {
	authService, pool, cleanup := SetupIntegration(t)
	defer cleanup()

	ctx := context.Background()

	// Setup: User with role and policy
	_, err := pool.Exec(ctx, `
		INSERT INTO user_attributes (id_user, id_attribute, value)
		VALUES ($1, $2, $3)
	`, 1, 1, "HR")
	if err != nil {
		t.Fatalf("Failed to setup test data: %v", err)
	}

	type PolicyRuleWithArray struct {
		Role     string   `json:"role"`
		Resource string   `json:"resource"`
		Action   []string `json:"action"`
	}
	rule := PolicyRuleWithArray{Role: "HR", Resource: "employee_records", Action: []string{"read"}}
	ruleJSON, _ := json.Marshal(rule)
	_, err = pool.Exec(ctx, `
		INSERT INTO policies (name, policy_rule, is_active)
		VALUES ($1, $2, $3)
	`, "HR Policy", ruleJSON, true)
	if err != nil {
		t.Fatalf("Failed to insert policy: %v", err)
	}

	// Test
	req := model.AuthorizeRequest{
		UserID:   1,
		Resource: "employee_records",
		Action:   "read",
	}

	resp, err := authService.Authorize(ctx, req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify audit log was created
	var count int
	err = pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM audit_logs
		WHERE id_user = $1 AND resource = $2 AND action = $3 AND allowed = $4
	`, 1, "employee_records", "read", true).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query audit log: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 audit log entry, got %d", count)
	}

	// Verify audit log details match response
	var allowed bool
	var reason string
	err = pool.QueryRow(ctx, `
		SELECT allowed, reason FROM audit_logs
		WHERE id_user = $1 AND resource = $2 AND action = $3
	`, 1, "employee_records", "read").Scan(&allowed, &reason)
	if err != nil {
		t.Fatalf("Failed to query audit log: %v", err)
	}

	if allowed != resp.Allowed {
		t.Errorf("Audit log allowed %v doesn't match response allowed %v", allowed, resp.Allowed)
	}

	if reason != resp.Reason {
		t.Errorf("Audit log reason '%s' doesn't match response reason '%s'", reason, resp.Reason)
	}
}

func TestIntegration_AuditLog_DeniedDecision(t *testing.T) {
	authService, pool, cleanup := SetupIntegration(t)
	defer cleanup()

	ctx := context.Background()

	// Setup: User without matching policy
	_, err := pool.Exec(ctx, `
		INSERT INTO user_attributes (id_user, id_attribute, value)
		VALUES ($1, $2, $3)
	`, 1, 1, "Staff")
	if err != nil {
		t.Fatalf("Failed to setup test data: %v", err)
	}

	// Test
	req := model.AuthorizeRequest{
		UserID:   1,
		Resource: "employee_records",
		Action:   "read",
	}

	_, err = authService.Authorize(ctx, req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify denial was logged
	var allowed bool
	err = pool.QueryRow(ctx, `
		SELECT allowed FROM audit_logs WHERE id_user = $1
	`, 1).Scan(&allowed)
	if err != nil {
		t.Fatalf("Failed to query audit log: %v", err)
	}

	if allowed {
		t.Error("Expected audit log to show allowed=false, got true")
	}
}

// ============================================================================
// Batch Authorization Tests
// ============================================================================

func TestIntegration_BatchAuthorize_MultipleRequests(t *testing.T) {
	authService, pool, cleanup := SetupIntegration(t)
	defer cleanup()

	ctx := context.Background()

	// Setup: User with HR role
	_, err := pool.Exec(ctx, `
		INSERT INTO user_attributes (id_user, id_attribute, value)
		VALUES ($1, $2, $3)
	`, 1, 1, "HR")
	if err != nil {
		t.Fatalf("Failed to setup test data: %v", err)
	}

	// Setup: Multiple policies
	type PolicyRuleWithArray struct {
		Role     string   `json:"role"`
		Resource string   `json:"resource"`
		Action   []string `json:"action"`
	}

	policies := []struct {
		name     string
		rule     PolicyRuleWithArray
	}{
		{"HR Employee Records", PolicyRuleWithArray{Role: "HR", Resource: "employee_records", Action: []string{"read", "write"}}},
		{"HR Budget Report", PolicyRuleWithArray{Role: "HR", Resource: "budget_report", Action: []string{"read"}}},
	}

	for _, p := range policies {
		ruleJSON, _ := json.Marshal(p.rule)
		_, err = pool.Exec(ctx, `
			INSERT INTO policies (name, policy_rule, is_active)
			VALUES ($1, $2, $3)
		`, p.name, ruleJSON, true)
		if err != nil {
			t.Fatalf("Failed to insert policy: %v", err)
		}
	}

	// Test batch authorization
	requests := []model.AuthorizeRequest{
		{Resource: "employee_records", Action: "read"},
		{Resource: "employee_records", Action: "write"},
		{Resource: "budget_report", Action: "read"},
		{Resource: "budget_report", Action: "write"}, // Should be denied
	}

	responses, err := authService.BatchAuthorize(ctx, 1, requests)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(responses) != len(requests) {
		t.Fatalf("Expected %d responses, got %d", len(requests), len(responses))
	}

	// Verify first three are allowed, last one is denied
	for i := 0; i < 3; i++ {
		if !responses[i].Allowed {
			t.Errorf("Expected request %d to be allowed, got denied. Reason: %s", i, responses[i].Reason)
		}
	}

	if responses[3].Allowed {
		t.Error("Expected last request (write budget_report) to be denied, got allowed")
	}
}

// ============================================================================
// Context Cancellation Tests
// ============================================================================

func TestIntegration_ContextCancellation_DuringRoleFetch(t *testing.T) {
	authService, _, cleanup := SetupIntegration(t)
	defer cleanup()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	req := model.AuthorizeRequest{
		UserID:   1,
		Resource: "employee_records",
		Action:   "read",
	}

	_, err := authService.Authorize(ctx, req)
	if err == nil {
		t.Error("Expected error due to context cancellation, got nil")
	}

	// Context cancellation should result in an error
	if err != nil && err.Error() != "context canceled" && err.Error() != "failed to get user roles: context canceled" {
		t.Logf("Got expected error (may vary): %v", err)
	}
}

// ============================================================================
// Concurrency Tests
// ============================================================================

func TestIntegration_ConcurrentRequests_ThreadSafe(t *testing.T) {
	authService, pool, cleanup := SetupIntegration(t)
	defer cleanup()

	ctx := context.Background()

	// Setup: User with role
	_, err := pool.Exec(ctx, `
		INSERT INTO user_attributes (id_user, id_attribute, value)
		VALUES ($1, $2, $3)
	`, 1, 1, "HR")
	if err != nil {
		t.Fatalf("Failed to setup test data: %v", err)
	}

	// Setup: Policy
	type PolicyRuleWithArray struct {
		Role     string   `json:"role"`
		Resource string   `json:"resource"`
		Action   []string `json:"action"`
	}
	rule := PolicyRuleWithArray{Role: "HR", Resource: "employee_records", Action: []string{"read"}}
	ruleJSON, _ := json.Marshal(rule)
	_, err = pool.Exec(ctx, `
		INSERT INTO policies (name, policy_rule, is_active)
		VALUES ($1, $2, $3)
	`, "HR Policy", ruleJSON, true)
	if err != nil {
		t.Fatalf("Failed to insert policy: %v", err)
	}

	// Test concurrent requests
	req := model.AuthorizeRequest{
		UserID:   1,
		Resource: "employee_records",
		Action:   "read",
	}

	concurrent := 10
	results := make(chan bool, concurrent)

	for i := 0; i < concurrent; i++ {
		go func() {
			resp, err := authService.Authorize(ctx, req)
			if err != nil {
				t.Errorf("Concurrent request failed: %v", err)
				results <- false
				return
			}
			results <- resp.Allowed
		}()
	}

	// Verify all requests succeeded
	allowedCount := 0
	for i := 0; i < concurrent; i++ {
		if <-results {
			allowedCount++
		}
	}

	if allowedCount != concurrent {
		t.Errorf("Expected %d allowed results, got %d", concurrent, allowedCount)
	}
}

// ============================================================================
// Wildcard Authorization Integration Tests
// ============================================================================

func TestIntegration_Admin_DoubleWildcard_AllAPIResources(t *testing.T) {
	authService, pool, cleanup := SetupIntegration(t)
	defer cleanup()

	ctx := context.Background()

	// Setup: Admin user
	_, err := pool.Exec(ctx, `
		INSERT INTO user_attributes (id_user, id_attribute, value)
		VALUES ($1, $2, $3)
	`, 1, 1, "admin")
	if err != nil {
		t.Fatalf("Failed to setup admin user: %v", err)
	}

	// Setup: Double wildcard policy from api_docs_schema.sql pattern
	// {"role": "admin", "resource": "api_docs_*", "action": ["*"]}
	type PolicyRuleWithArray struct {
		Role     string   `json:"role"`
		Resource string   `json:"resource"`
		Action   []string `json:"action"`
	}
	rule := PolicyRuleWithArray{Role: "admin", Resource: "api_docs_*", Action: []string{"*"}}
	ruleJSON, _ := json.Marshal(rule)
	_, err = pool.Exec(ctx, `
		INSERT INTO policies (name, policy_rule, is_active)
		VALUES ($1, $2, $3)
	`, "API Docs Admin Policy", ruleJSON, true)
	if err != nil {
		t.Fatalf("Failed to insert policy: %v", err)
	}

	// Test various api_docs resources and actions
	testCases := []struct {
		resource string
		action   string
		expected bool
	}{
		// api_docs_* resources should be allowed
		{"api_docs_project", "create", true},
		{"api_docs_project", "read", true},
		{"api_docs_project", "update", true},
		{"api_docs_project", "delete", true},
		{"api_docs_rest_api", "create", true},
		{"api_docs_rest_api", "read", true},
		{"api_docs_graphql_api", "create", true},
		{"api_docs_graphql_api", "read", true},
		{"api_docs_grpc_api", "create", true},
		{"api_docs_grpc_api", "read", true},
		// Different resource prefix should be denied
		{"employee_records", "read", false},
		{"budget_report", "read", false},
	}

	for _, tc := range testCases {
		req := model.AuthorizeRequest{
			UserID:   1,
			Resource: tc.resource,
			Action:   tc.action,
		}

		resp, err := authService.Authorize(ctx, req)
		if err != nil {
			t.Fatalf("Expected no error for %s/%s, got %v", tc.resource, tc.action, err)
		}

		if resp.Allowed != tc.expected {
			t.Errorf("Expected %s/%s to be %v, got %v. Reason: %s",
				tc.resource, tc.action, tc.expected, resp.Allowed, resp.Reason)
		}
	}
}

func TestIntegration_MultiplePrefixWildcards_DifferentPrefixes(t *testing.T) {
	authService, pool, cleanup := SetupIntegration(t)
	defer cleanup()

	ctx := context.Background()

	// Setup: Admin user
	_, err := pool.Exec(ctx, `
		INSERT INTO user_attributes (id_user, id_attribute, value)
		VALUES ($1, $2, $3)
	`, 1, 1, "admin")
	if err != nil {
		t.Fatalf("Failed to setup admin user: %v", err)
	}

	// Setup: Multiple prefix wildcard policies
	type PolicyRuleWithArray struct {
		Role     string   `json:"role"`
		Resource string   `json:"resource"`
		Action   []string `json:"action"`
	}

	policies := []struct {
		name     string
		resource string
		actions  []string
	}{
		{"API Docs Policy", "api_docs_*", []string{"create", "read", "update"}},
		{"Employee Policy", "employee_*", []string{"read", "write"}},
	}

	for i, p := range policies {
		rule := PolicyRuleWithArray{Role: "admin", Resource: p.resource, Action: p.actions}
		ruleJSON, _ := json.Marshal(rule)
		_, err = pool.Exec(ctx, `
			INSERT INTO policies (id_policy, name, policy_rule, is_active)
			VALUES ($1, $2, $3, $4)
		`, int64(i+10), p.name, ruleJSON, true)
		if err != nil {
			t.Fatalf("Failed to insert policy %s: %v", p.name, err)
		}
	}

	// Test: api_docs_* should match api_docs resources
	apiDocsTests := []struct {
		resource string
		action   string
	}{
		{"api_docs_project", "create"},
		{"api_docs_project", "read"},
		{"api_docs_rest_api", "read"},
	}

	for _, tc := range apiDocsTests {
		req := model.AuthorizeRequest{
			UserID:   1,
			Resource: tc.resource,
			Action:   tc.action,
		}

		resp, err := authService.Authorize(ctx, req)
		if err != nil {
			t.Errorf("Expected no error for %s/%s, got %v", tc.resource, tc.action, err)
		}

		if !resp.Allowed {
			t.Errorf("Expected api_docs_* policy to allow %s/%s, got denied. Reason: %s",
				tc.resource, tc.action, resp.Reason)
		}
	}

	// Test: employee_* should match employee resources
	employeeTests := []struct {
		resource string
		action   string
	}{
		{"employee_records", "read"},
		{"employee_records", "write"},
		{"employee_salary", "read"},
	}

	for _, tc := range employeeTests {
		req := model.AuthorizeRequest{
			UserID:   1,
			Resource: tc.resource,
			Action:   tc.action,
		}

		resp, err := authService.Authorize(ctx, req)
		if err != nil {
			t.Errorf("Expected no error for %s/%s, got %v", tc.resource, tc.action, err)
		}

		if !resp.Allowed {
			t.Errorf("Expected employee_* policy to allow %s/%s, got denied. Reason: %s",
				tc.resource, tc.action, resp.Reason)
		}
	}

	// Test: api_docs_* should NOT match employee_records
	req := model.AuthorizeRequest{
		UserID:   1,
		Resource: "employee_records",
		Action:   "create",
	}

	resp, err := authService.Authorize(ctx, req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.Allowed {
		t.Error("api_docs_* policy should not allow employee_records/create")
	}
}
