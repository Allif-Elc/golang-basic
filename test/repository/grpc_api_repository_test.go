package repository_test

import (
	"context"
	"encoding/json"
	"testing"

	"golang-basic/api/internal/model"
	"golang-basic/api/internal/repository"
)

// TestListGrpcByProjectIDPaginated_Success tests paginated retrieval of gRPC APIs
func TestListGrpcByProjectIDPaginated_Success(t *testing.T) {
	pool, cleanup := SetupPostgres(t)
	defer cleanup()

	ctx := context.Background()
	repo := repository.NewGrpcAPIRepository(pool)

	// Insert test user
	_, err := pool.Exec(ctx, "INSERT INTO users (id_user, name, email, password_hash) VALUES ($1, $2, $3, $4)", 1, "Test User", "test@example.com", "hash")
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}

	// Insert test project
	_, err = pool.Exec(ctx, "INSERT INTO projects (id_project, id_user, name, slug) VALUES ($1, $2, $3, $4)", 1, 1, "Test Project", "test-project")
	if err != nil {
		t.Fatalf("Failed to insert project: %v", err)
	}

	// Insert test gRPC APIs
	for i := 1; i <= 25; i++ {
		requestMsg, _ := json.Marshal([]model.GrpcField{{Name: "id", Type: "int32"}})
		responseMsg, _ := json.Marshal([]model.GrpcField{{Name: "result", Type: "string"}})
		examples, _ := json.Marshal([]model.GrpcExample{{Language: "go", Code: "example"}})

		_, err = pool.Exec(ctx, `
			INSERT INTO grpc_apis (id_grpc_api, id_project, id_user, service_name, method_name, description, request_message, response_message, proto_definition, examples, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
		`, int64(i), 1, 1, "Service"+string(rune('A'+byte(i))), "Method"+string(rune('A'+byte(i))), "Description "+string(rune('a'+byte(i))), requestMsg, responseMsg, "proto definition", examples)
		if err != nil {
			t.Fatalf("Failed to insert gRPC API: %v", err)
		}
	}

	// Test - first page with limit 20
	apis, total, err := repo.ListByProjectIDPaginated(ctx, 1, 1, 20)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if total != 25 {
		t.Errorf("Expected total 25, got %d", total)
	}

	if len(apis) != 20 {
		t.Errorf("Expected 20 APIs on first page, got %d", len(apis))
	}

	if apis[0].ServiceName != "ServiceA" {
		t.Errorf("Expected first service name 'ServiceA', got '%s'", apis[0].ServiceName)
	}
}

// TestListGrpcByProjectIDPaginated_Empty tests empty result for project with no APIs
func TestListGrpcByProjectIDPaginated_Empty(t *testing.T) {
	pool, cleanup := SetupPostgres(t)
	defer cleanup()

	ctx := context.Background()
	repo := repository.NewGrpcAPIRepository(pool)

	// Insert test user
	_, err := pool.Exec(ctx, "INSERT INTO users (id_user, name, email, password_hash) VALUES ($1, $2, $3, $4)", 1, "Test User", "test@example.com", "hash")
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}

	// Insert test project without any APIs
	_, err = pool.Exec(ctx, "INSERT INTO projects (id_project, id_user, name, slug) VALUES ($1, $2, $3, $4)", 1, 1, "Empty Project", "empty-project")
	if err != nil {
		t.Fatalf("Failed to insert project: %v", err)
	}

	// Test - should return empty array
	apis, total, err := repo.ListByProjectIDPaginated(ctx, 1, 1, 20)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(apis) != 0 {
		t.Errorf("Expected empty array for project with no APIs, got %d", len(apis))
	}

	if total != 0 {
		t.Errorf("Expected 0 total for project with no APIs, got %d", total)
	}
}

// TestListGrpcByProjectIDPaginated_RespectsLimit tests that limit parameter is respected
func TestListGrpcByProjectIDPaginated_RespectsLimit(t *testing.T) {
	pool, cleanup := SetupPostgres(t)
	defer cleanup()

	ctx := context.Background()
	repo := repository.NewGrpcAPIRepository(pool)

	// Insert test user
	_, err := pool.Exec(ctx, "INSERT INTO users (id_user, name, email, password_hash) VALUES ($1, $2, $3, $4)", 1, "Test User", "test@example.com", "hash")
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}

	// Insert test project
	_, err = pool.Exec(ctx, "INSERT INTO projects (id_project, id_user, name, slug) VALUES ($1, $2, $3, $4)", 1, 1, "Test Project", "test-project")
	if err != nil {
		t.Fatalf("Failed to insert project: %v", err)
	}

	// Insert 15 test gRPC APIs
	for i := 1; i <= 15; i++ {
		requestMsg, _ := json.Marshal([]model.GrpcField{})
		responseMsg, _ := json.Marshal([]model.GrpcField{})
		examples, _ := json.Marshal([]model.GrpcExample{})

		_, err = pool.Exec(ctx, `
			INSERT INTO grpc_apis (id_grpc_api, id_project, id_user, service_name, method_name, description, request_message, response_message, proto_definition, examples, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
		`, int64(i), 1, 1, "Service"+string(rune('A'+byte(i))), "Method"+string(rune('A'+byte(i))), "Description "+string(rune('a'+byte(i))), requestMsg, responseMsg, "proto", examples)
		if err != nil {
			t.Fatalf("Failed to insert gRPC API: %v", err)
		}
	}

	// Test - limit of 10 should return only 10 APIs
	apis, total, err := repo.ListByProjectIDPaginated(ctx, 1, 1, 10)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if total != 15 {
		t.Errorf("Expected total 15, got %d", total)
	}

	if len(apis) != 10 {
		t.Errorf("Expected limit to be respected, got %d APIs instead of 10", len(apis))
	}
}

// TestListGrpcByProjectIDPaginated_RespectsOffset tests that offset (calculated from page) is respected
func TestListGrpcByProjectIDPaginated_RespectsOffset(t *testing.T) {
	pool, cleanup := SetupPostgres(t)
	defer cleanup()

	ctx := context.Background()
	repo := repository.NewGrpcAPIRepository(pool)

	// Insert test user
	_, err := pool.Exec(ctx, "INSERT INTO users (id_user, name, email, password_hash) VALUES ($1, $2, $3, $4)", 1, "Test User", "test@example.com", "hash")
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}

	// Insert test project
	_, err = pool.Exec(ctx, "INSERT INTO projects (id_project, id_user, name, slug) VALUES ($1, $2, $3, $4)", 1, 1, "Test Project", "test-project")
	if err != nil {
		t.Fatalf("Failed to insert project: %v", err)
	}

	// Insert 25 test gRPC APIs
	for i := 1; i <= 25; i++ {
		requestMsg, _ := json.Marshal([]model.GrpcField{})
		responseMsg, _ := json.Marshal([]model.GrpcField{})
		examples, _ := json.Marshal([]model.GrpcExample{})

		_, err = pool.Exec(ctx, `
			INSERT INTO grpc_apis (id_grpc_api, id_project, id_user, service_name, method_name, description, request_message, response_message, proto_definition, examples, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
		`, int64(i), 1, 1, "Service"+string(rune('A'+byte(i))), "Method"+string(rune('A'+byte(i))), "Description "+string(rune('a'+byte(i))), requestMsg, responseMsg, "proto", examples)
		if err != nil {
			t.Fatalf("Failed to insert gRPC API: %v", err)
		}
	}

	// Test - page 2 with limit 10 should skip first 10 and return next 10
	apis, total, err := repo.ListByProjectIDPaginated(ctx, 1, 2, 10)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if total != 25 {
		t.Errorf("Expected total 25, got %d", total)
	}

	if len(apis) != 10 {
		t.Errorf("Expected 10 APIs on page 2, got %d", len(apis))
	}

	// Verify that first API on page 2 is 11th API (ServiceK)
	if apis[0].ServiceName != "ServiceK" {
		t.Errorf("Expected first service on page 2 to be 'ServiceK', got '%s'", apis[0].ServiceName)
	}
}

// TestListGrpcByProjectIDPaginated_ProjectNotFound tests querying APIs for non-existent project
func TestListGrpcByProjectIDPaginated_ProjectNotFound(t *testing.T) {
	pool, cleanup := SetupPostgres(t)
	defer cleanup()

	ctx := context.Background()
	repo := repository.NewGrpcAPIRepository(pool)

	// Test - should return empty array for non-existent project
	apis, total, err := repo.ListByProjectIDPaginated(ctx, 999, 1, 20)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(apis) != 0 {
		t.Errorf("Expected empty array for non-existent project, got %d", len(apis))
	}

	if total != 0 {
		t.Errorf("Expected 0 total for non-existent project, got %d", total)
	}
}

// TestListGrpcByProjectIDPaginated_InvalidPagination tests with invalid pagination parameters
func TestListGrpcByProjectIDPaginated_InvalidPagination(t *testing.T) {
	pool, cleanup := SetupPostgres(t)
	defer cleanup()

	ctx := context.Background()
	repo := repository.NewGrpcAPIRepository(pool)

	// Insert test user
	_, err := pool.Exec(ctx, "INSERT INTO users (id_user, name, email, password_hash) VALUES ($1, $2, $3, $4)", 1, "Test User", "test@example.com", "hash")
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}

	// Insert test project
	_, err = pool.Exec(ctx, "INSERT INTO projects (id_project, id_user, name, slug) VALUES ($1, $2, $3, $4)", 1, 1, "Test Project", "test-project")
	if err != nil {
		t.Fatalf("Failed to insert project: %v", err)
	}

	// Insert test gRPC API
	requestMsg, _ := json.Marshal([]model.GrpcField{{Name: "id", Type: "int32"}})
	responseMsg, _ := json.Marshal([]model.GrpcField{{Name: "result", Type: "string"}})
	examples, _ := json.Marshal([]model.GrpcExample{{Language: "go", Code: "example"}})
	_, err = pool.Exec(ctx, `
		INSERT INTO grpc_apis (id_grpc_api, id_project, id_user, service_name, method_name, description, request_message, response_message, proto_definition, examples, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
	`, 1, 1, 1, "TestService", "TestMethod", "Test Description", requestMsg, responseMsg, "proto", examples)
	if err != nil {
		t.Fatalf("Failed to insert gRPC API: %v", err)
	}

	// Test - page 0 should be treated as page 1
	apis, total, err := repo.ListByProjectIDPaginated(ctx, 1, 0, 20)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(apis) != 1 {
		t.Errorf("Expected 1 API with page 0 (treated as 1), got %d", len(apis))
	}

	if total != 1 {
		t.Errorf("Expected total 1, got %d", total)
	}

	// Test - limit 0 should be treated as default 20
	apis, _, err = repo.ListByProjectIDPaginated(ctx, 1, 1, 0)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(apis) != 1 {
		t.Errorf("Expected 1 API with limit 0 (treated as default 20), got %d", len(apis))
	}

	// Test - limit > 100 should be capped at 100
	apis, _, err = repo.ListByProjectIDPaginated(ctx, 1, 1, 200)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(apis) != 1 {
		t.Errorf("Expected 1 API with limit 200 (capped at 100), got %d", len(apis))
	}
}
