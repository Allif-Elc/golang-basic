package repository_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"golang-basic/api/internal/model"
	"golang-basic/api/internal/repository"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// SetupPostgres starts a test PostgreSQL container with Podman
// Returns the connection pool and a cleanup function
func SetupPostgres(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()

	ctx := context.Background()

	// Configure Podman as the container provider
	// This allows testcontainers to use Podman instead of Docker
	testcontainersURL := os.Getenv("TESTCONTAINERS_PODMAN_SOCKET")
	if testcontainersURL == "" {
		// Default to unix:///run/podman/podman.sock if not set
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

	// Configure PostgreSQL container with Podman
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
		t.Skipf("Skipping test - Podman not available or not running: %v", err)
		return nil, func() {}
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

	// Cleanup function
	cleanup := func() {
		pool.Close()
		if err := container.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate container: %v", err)
		}
	}

	return pool, cleanup
}

// Helper to create a test policy
// Note: Action is stored as a string in Go model but as an array in JSONB
func createTestPolicy(id int64, name, role, resource string, action []string, isActive bool) model.Policy {
	rule := model.PolicyRule{
		Role:     role,
		Resource: resource,
		Action:   action,
	}
	ruleJSON, _ := json.Marshal(rule)

	return model.Policy{
		PolicyId:   id,
		Name:       name,
		PolicyRule: ruleJSON,
		IsActive:   isActive,
	}
}

// Helper to create a test policy with action array (for JSONB compatibility)
func createTestPolicyWithActions(id int64, name, role, resource string, actions []string, isActive bool) model.Policy {
	type PolicyRuleWithArray struct {
		Role     string   `json:"role"`
		Resource string   `json:"resource"`
		Action   []string `json:"action"`
	}
	rule := PolicyRuleWithArray{
		Role:     role,
		Resource: resource,
		Action:   actions,
	}
	ruleJSON, _ := json.Marshal(rule)

	return model.Policy{
		PolicyId:   id,
		Name:       name,
		PolicyRule: ruleJSON,
		IsActive:   isActive,
	}
}

// ============================================================================
// GetUserRoles Tests
// ============================================================================

func TestGetUserRoles_Success(t *testing.T) {
	pool, cleanup := SetupPostgres(t)
	defer cleanup()

	ctx := context.Background()
	repo := repository.NewAuthorizationRepository(pool)

	// Insert test data
	// Create attribute
	_, err := pool.Exec(ctx, "INSERT INTO attributes (id_attribute, name) VALUES ($1, $2)", 1, "role")
	if err != nil {
		t.Fatalf("Failed to insert attribute: %v", err)
	}

	// Insert user
	_, err = pool.Exec(ctx, "INSERT INTO users (id_user, name, email, password_hash) VALUES ($1, $2, $3, $4)", 1, "Test User", "test@example.com", "hash")
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}

	// Assign roles to user
	_, err = pool.Exec(ctx, "INSERT INTO user_attributes (id_user, id_attribute, value) VALUES ($1, $2, $3)", 1, 1, "HR")
	if err != nil {
		t.Fatalf("Failed to assign role: %v", err)
	}

	// Test
	roles, err := repo.GetUserRoles(ctx, 1)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(roles) != 1 {
		t.Fatalf("Expected 1 role, got %d", len(roles))
	}

	if roles[0] != "HR" {
		t.Errorf("Expected role 'HR', got '%s'", roles[0])
	}
}

func TestGetUserRoles_NoRoles(t *testing.T) {
	pool, cleanup := SetupPostgres(t)
	defer cleanup()

	ctx := context.Background()
	repo := repository.NewAuthorizationRepository(pool)

	// Insert user without roles
	_, err := pool.Exec(ctx, "INSERT INTO users (id_user, name, email, password_hash) VALUES ($1, $2, $3, $4)", 1, "Test User", "test@example.com", "hash")
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}

	// Test
	roles, err := repo.GetUserRoles(ctx, 1)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(roles) != 0 {
		t.Fatalf("Expected 0 roles, got %d", len(roles))
	}
}

func TestGetUserRoles_MultipleRoles(t *testing.T) {
	pool, cleanup := SetupPostgres(t)
	defer cleanup()

	ctx := context.Background()
	repo := repository.NewAuthorizationRepository(pool)

	// Insert test data
	_, err := pool.Exec(ctx, "INSERT INTO attributes (id_attribute, name) VALUES ($1, $2)", 1, "role")
	if err != nil {
		t.Fatalf("Failed to insert attribute: %v", err)
	}

	_, err = pool.Exec(ctx, "INSERT INTO users (id_user, name, email, password_hash) VALUES ($1, $2, $3, $4)", 1, "Test User", "test@example.com", "hash")
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}

	// Assign multiple roles
	_, err = pool.Exec(ctx, "INSERT INTO user_attributes (id_user, id_attribute, value) VALUES ($1, $2, $3), ($1, $2, $4)", 1, 1, "HR", "Manager")
	if err != nil {
		t.Fatalf("Failed to assign roles: %v", err)
	}

	// Test
	roles, err := repo.GetUserRoles(ctx, 1)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(roles) != 2 {
		t.Fatalf("Expected 2 roles, got %d", len(roles))
	}

	// Check both roles exist
	roleMap := make(map[string]bool)
	for _, role := range roles {
		roleMap[role] = true
	}

	if !roleMap["HR"] || !roleMap["Manager"] {
		t.Errorf("Expected roles [HR, Manager], got %v", roles)
	}
}

func TestGetUserRoles_UserNotFound(t *testing.T) {
	pool, cleanup := SetupPostgres(t)
	defer cleanup()

	ctx := context.Background()
	repo := repository.NewAuthorizationRepository(pool)

	// Test with non-existent user
	roles, err := repo.GetUserRoles(ctx, 999)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(roles) != 0 {
		t.Fatalf("Expected 0 roles for non-existent user, got %d", len(roles))
	}
}

// ============================================================================
// GetMatchingPolicies Tests
// ============================================================================

func TestGetMatchingPolicies_SingleMatch(t *testing.T) {
	pool, cleanup := SetupPostgres(t)
	defer cleanup()

	ctx := context.Background()
	repo := repository.NewAuthorizationRepository(pool)

	// Insert test policy
	policy := createTestPolicyWithActions(1, "HR Policy", "HR", "employee_records", []string{"read"}, true)
	_, err := pool.Exec(ctx, `
		INSERT INTO policies (id_policy, name, policy_rule, is_active)
		VALUES ($1, $2, $3, $4)
	`, policy.PolicyId, policy.Name, policy.PolicyRule, policy.IsActive)
	if err != nil {
		t.Fatalf("Failed to insert policy: %v", err)
	}

	// Test
	policies, err := repo.GetMatchingPolicies(ctx, "employee_records", "read")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(policies) != 1 {
		t.Fatalf("Expected 1 policy, got %d", len(policies))
	}

	if policies[0].Name != "HR Policy" {
		t.Errorf("Expected policy name 'HR Policy', got '%s'", policies[0].Name)
	}
}

func TestGetMatchingPolicies_NoMatches(t *testing.T) {
	pool, cleanup := SetupPostgres(t)
	defer cleanup()

	ctx := context.Background()
	repo := repository.NewAuthorizationRepository(pool)

	// Test with no policies
	policies, err := repo.GetMatchingPolicies(ctx, "nonexistent", "read")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(policies) != 0 {
		t.Fatalf("Expected 0 policies, got %d", len(policies))
	}
}

func TestGetMatchingPolicies_OnlyActivePolicies(t *testing.T) {
	pool, cleanup := SetupPostgres(t)
	defer cleanup()

	ctx := context.Background()
	repo := repository.NewAuthorizationRepository(pool)

	// Insert active policy
	activePolicy := createTestPolicyWithActions(1, "Active Policy", "HR", "employee_records", []string{"read"}, true)
	_, err := pool.Exec(ctx, `
		INSERT INTO policies (id_policy, name, policy_rule, is_active)
		VALUES ($1, $2, $3, $4)
	`, activePolicy.PolicyId, activePolicy.Name, activePolicy.PolicyRule, activePolicy.IsActive)
	if err != nil {
		t.Fatalf("Failed to insert active policy: %v", err)
	}

	// Insert inactive policy
	inactivePolicy := createTestPolicyWithActions(2, "Inactive Policy", "HR", "employee_records", []string{"read"}, false)
	_, err = pool.Exec(ctx, `
		INSERT INTO policies (id_policy, name, policy_rule, is_active)
		VALUES ($1, $2, $3, $4)
	`, inactivePolicy.PolicyId, inactivePolicy.Name, inactivePolicy.PolicyRule, inactivePolicy.IsActive)
	if err != nil {
		t.Fatalf("Failed to insert inactive policy: %v", err)
	}

	// Test - should only return active policy
	policies, err := repo.GetMatchingPolicies(ctx, "employee_records", "read")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(policies) != 1 {
		t.Fatalf("Expected 1 active policy, got %d", len(policies))
	}

	if policies[0].Name != "Active Policy" {
		t.Errorf("Expected active policy, got '%s'", policies[0].Name)
	}
}

func TestGetMatchingPolicies_MultipleMatches(t *testing.T) {
	pool, cleanup := SetupPostgres(t)
	defer cleanup()

	ctx := context.Background()
	repo := repository.NewAuthorizationRepository(pool)

	// Insert multiple policies
	policy1 := createTestPolicyWithActions(1, "HR Read", "HR", "employee_records", []string{"read"}, true)
	policy2 := createTestPolicyWithActions(2, "HR Write", "HR", "employee_records", []string{"write"}, true)

	for _, p := range []model.Policy{policy1, policy2} {
		_, err := pool.Exec(ctx, `
			INSERT INTO policies (id_policy, name, policy_rule, is_active)
			VALUES ($1, $2, $3, $4)
		`, p.PolicyId, p.Name, p.PolicyRule, p.IsActive)
		if err != nil {
			t.Fatalf("Failed to insert policy: %v", err)
		}
	}

	// Test - query for read should return at least one policy
	policies, err := repo.GetMatchingPolicies(ctx, "employee_records", "read")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(policies) == 0 {
		t.Fatal("Expected at least 1 policy, got 0")
	}
}

// ============================================================================
// LogAuditDecision Tests
// ============================================================================

func TestLogAuditDecision_Allowed(t *testing.T) {
	pool, cleanup := SetupPostgres(t)
	defer cleanup()

	ctx := context.Background()
	repo := repository.NewAuthorizationRepository(pool)

	// Insert user
	_, err := pool.Exec(ctx, "INSERT INTO users (id_user, name, email, password_hash) VALUES ($1, $2, $3, $4)", 1, "Test User", "test@example.com", "hash")
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}

	// Test
	err = repo.LogAuditDecision(ctx, 1, "employee_records", "read", true, "Allowed by HR role")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify log was created
	var count int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM audit_logs WHERE id_user = $1", 1).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query audit log: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 audit log, got %d", count)
	}
}

func TestLogAuditDecision_Denied(t *testing.T) {
	pool, cleanup := SetupPostgres(t)
	defer cleanup()

	ctx := context.Background()
	repo := repository.NewAuthorizationRepository(pool)

	// Insert user
	_, err := pool.Exec(ctx, "INSERT INTO users (id_user, name, email, password_hash) VALUES ($1, $2, $3, $4)", 1, "Test User", "test@example.com", "hash")
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}

	// Test
	err = repo.LogAuditDecision(ctx, 1, "budget_report", "read", false, "No policies found")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify log was created
	var allowed bool
	var reason string
	err = pool.QueryRow(ctx, "SELECT allowed, reason FROM audit_logs WHERE id_user = $1", 1).Scan(&allowed, &reason)
	if err != nil {
		t.Fatalf("Failed to query audit log: %v", err)
	}

	if allowed {
		t.Error("Expected allowed=false, got true")
	}

	if reason != "No policies found" {
		t.Errorf("Expected reason 'No policies found', got '%s'", reason)
	}
}

// ============================================================================
// ParsePolicyRule Tests
// ============================================================================

func TestParsePolicyRule_Valid(t *testing.T) {
	ruleJSON := []byte(`{"role": "HR", "resource": "employee_records", "action": "read"}`)

	rule, err := repository.ParsePolicyRule(ruleJSON)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if rule.Role != "HR" {
		t.Errorf("Expected role 'HR', got '%s'", rule.Role)
	}

	if rule.Resource != "employee_records" {
		t.Errorf("Expected resource 'employee_records', got '%s'", rule.Resource)
	}

	if !slices.Contains(rule.Action, "read") {
		t.Errorf("Expected action 'read', got '%s'", rule.Action)
	}
}

func TestParsePolicyRule_MalformedJSON(t *testing.T) {
	ruleJSON := []byte(`{invalid json}`)

	_, err := repository.ParsePolicyRule(ruleJSON)
	if err == nil {
		t.Fatal("Expected error for malformed JSON, got nil")
	}
}

// ============================================================================
// Wildcard Matching Tests
// ============================================================================

func TestGetMatchingPolicies_WildcardResource_All(t *testing.T) {
	pool, cleanup := SetupPostgres(t)
	defer cleanup()

	ctx := context.Background()
	repo := repository.NewAuthorizationRepository(pool)

	// Insert full wildcard policy
	wildcardPolicy := createTestPolicyWithActions(1, "Full Wildcard", "admin", "*", []string{"*"}, true)
	_, err := pool.Exec(ctx, `
		INSERT INTO policies (id_policy, name, policy_rule, is_active)
		VALUES ($1, $2, $3, $4)
	`, wildcardPolicy.PolicyId, wildcardPolicy.Name, wildcardPolicy.PolicyRule, wildcardPolicy.IsActive)
	if err != nil {
		t.Fatalf("Failed to insert wildcard policy: %v", err)
	}

	// Test: wildcard should match any resource
	testResources := []string{"employee_records", "budget_report", "api_docs_project", "random_resource"}
	for _, resource := range testResources {
		policies, err := repo.GetMatchingPolicies(ctx, resource, "read")
		if err != nil {
			t.Errorf("Expected no error for resource %s, got %v", resource, err)
		}

		if len(policies) != 1 {
			t.Errorf("Expected wildcard policy to match resource %s, got %d policies", resource, len(policies))
		}
	}
}

func TestGetMatchingPolicies_PrefixWildcard_MultiplePrefixes(t *testing.T) {
	pool, cleanup := SetupPostgres(t)
	defer cleanup()

	ctx := context.Background()
	repo := repository.NewAuthorizationRepository(pool)

	// Insert policies with different prefix wildcards
	policyDefs := []struct {
		name     string
		role     string
		resource string
		actions  []string
	}{
		{"API Docs Policy", "developer", "api_docs_*", []string{"create", "read", "update"}},
		{"Employee Policy", "hr", "employee_*", []string{"read", "write"}},
		{"Projects Policy", "manager", "projects_*", []string{"read"}},
	}

	for i, pd := range policyDefs {
		policy := createTestPolicyWithActions(int64(i+1), pd.name, pd.role, pd.resource, pd.actions, true)
		_, err := pool.Exec(ctx, `
			INSERT INTO policies (id_policy, name, policy_rule, is_active)
			VALUES ($1, $2, $3, $4)
		`, policy.PolicyId, policy.Name, policy.PolicyRule, policy.IsActive)
		if err != nil {
			t.Fatalf("Failed to insert policy %s: %v", pd.name, err)
		}
	}

	// Test: api_docs_* should match api_docs resources
	apiDocsResources := []string{"api_docs_project", "api_docs_rest_api", "api_docs_graphql_api"}
	for _, resource := range apiDocsResources {
		policies, err := repo.GetMatchingPolicies(ctx, resource, "read")
		if err != nil {
			t.Errorf("Expected no error for %s, got %v", resource, err)
		}

		// Should match api_docs_* policy
		found := false
		for _, p := range policies {
			if p.Name == "API Docs Policy" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected api_docs_* policy to match %s", resource)
		}
	}

	// Test: employee_* should match employee resources
	employeeResources := []string{"employee_records", "employee_salary"}
	for _, resource := range employeeResources {
		policies, err := repo.GetMatchingPolicies(ctx, resource, "read")
		if err != nil {
			t.Errorf("Expected no error for %s, got %v", resource, err)
		}

		// Should match employee_* policy
		found := false
		for _, p := range policies {
			if p.Name == "Employee Policy" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected employee_* policy to match %s", resource)
		}
	}

	// Test: api_docs_* should NOT match employee_records
	policies, err := repo.GetMatchingPolicies(ctx, "employee_records", "read")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	for _, p := range policies {
		if p.Name == "API Docs Policy" {
			t.Error("api_docs_* policy should not match employee_records")
		}
	}
}

func TestGetMatchingPolicies_DoubleWildcard_MatchesAll(t *testing.T) {
	pool, cleanup := SetupPostgres(t)
	defer cleanup()

	ctx := context.Background()
	repo := repository.NewAuthorizationRepository(pool)

	// Insert double wildcard policy: resource="*" AND action=["*"]
	doubleWildcard := createTestPolicyWithActions(1, "Admin All Access", "admin", "*", []string{"*"}, true)
	_, err := pool.Exec(ctx, `
		INSERT INTO policies (id_policy, name, policy_rule, is_active)
		VALUES ($1, $2, $3, $4)
	`, doubleWildcard.PolicyId, doubleWildcard.Name, doubleWildcard.PolicyRule, doubleWildcard.IsActive)
	if err != nil {
		t.Fatalf("Failed to insert double wildcard policy: %v", err)
	}

	// Test: double wildcard should match any resource and action
	testCases := []struct {
		resource string
		action   string
	}{
		{"any_resource", "any_action"},
		{"employee_records", "read"},
		{"api_docs_project", "create"},
		{"budget_report", "delete"},
	}

	for _, tc := range testCases {
		policies, err := repo.GetMatchingPolicies(ctx, tc.resource, tc.action)
		if err != nil {
			t.Errorf("Expected no error for %s/%s, got %v", tc.resource, tc.action, err)
		}

		if len(policies) != 1 {
			t.Errorf("Expected double wildcard to match %s/%s, got %d policies", tc.resource, tc.action, len(policies))
		}
	}
}

func TestGetMatchingPolicies_ActionWildcard_AllActions(t *testing.T) {
	pool, cleanup := SetupPostgres(t)
	defer cleanup()

	ctx := context.Background()
	repo := repository.NewAuthorizationRepository(pool)

	// Insert policy with action wildcard
	actionWildcardPolicy := createTestPolicyWithActions(1, "Read All Resources", "viewer", "employee_*", []string{"*"}, true)
	_, err := pool.Exec(ctx, `
		INSERT INTO policies (id_policy, name, policy_rule, is_active)
		VALUES ($1, $2, $3, $4)
	`, actionWildcardPolicy.PolicyId, actionWildcardPolicy.Name, actionWildcardPolicy.PolicyRule, actionWildcardPolicy.IsActive)
	if err != nil {
		t.Fatalf("Failed to insert action wildcard policy: %v", err)
	}

	// Test: action wildcard ["*"] should match any action
	actions := []string{"read", "write", "delete", "create", "update"}
	for _, action := range actions {
		policies, err := repo.GetMatchingPolicies(ctx, "employee_records", action)
		if err != nil {
			t.Errorf("Expected no error for action %s, got %v", action, err)
		}

		if len(policies) != 1 {
			t.Errorf("Expected action wildcard to match %s, got %d policies", action, len(policies))
		}
	}
}

// ============================================================================
// Benchmarks
// ============================================================================

func BenchmarkGetUserRoles(b *testing.B) {
	// Note: This benchmark requires a container setup
	// For proper benchmarking, consider setting up the container once in BenchmarkSetup
	b.Skip("Skipping benchmark - requires persistent container setup")
}

func BenchmarkGetMatchingPolicies(b *testing.B) {
	b.Skip("Skipping benchmark - requires persistent container setup")
}

func BenchmarkLogAuditDecision(b *testing.B) {
	b.Skip("Skipping benchmark - requires persistent container setup")
}
