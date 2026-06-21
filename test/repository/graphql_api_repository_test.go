package repository_test

import (
	"context"
	"encoding/json"
	"testing"

	"golang-basic/api/internal/model"
	"golang-basic/api/internal/repository"
)

// TestListGraphQLByProjectIDPaginated_Success tests paginated retrieval of GraphQL APIs
func TestListGraphQLByProjectIDPaginated_Success(t *testing.T) {
	pool, cleanup := SetupPostgres(t)
	defer cleanup()

	ctx := context.Background()
	repo := repository.NewGraphQLAPIRepository(pool)

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

	// Insert test GraphQL APIs
	for i := 1; i <= 25; i++ {
		arguments, _ := json.Marshal([]model.GraphQLArgument{{Name: "id", Type: "ID!"}})
		examples, _ := json.Marshal([]model.GraphQLExample{{Name: "Example", Query: "query { user }"}})

		_, err = pool.Exec(ctx, `
			INSERT INTO graphql_apis (id_graphql_api, id_project, id_user, name, type, description, arguments, return_type, examples, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		`, int64(i), 1, 1, "API "+string(rune('A'+byte(i))), "query", "Description "+string(rune('a'+byte(i))), arguments, "String", examples)
		if err != nil {
			t.Fatalf("Failed to insert GraphQL API: %v", err)
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

	if apis[0].Name != "API A" {
		t.Errorf("Expected first API name 'API A', got '%s'", apis[0].Name)
	}
}

// TestListGraphQLByProjectIDPaginated_Empty tests empty result for project with no APIs
func TestListGraphQLByProjectIDPaginated_Empty(t *testing.T) {
	pool, cleanup := SetupPostgres(t)
	defer cleanup()

	ctx := context.Background()
	repo := repository.NewGraphQLAPIRepository(pool)

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

// TestListGraphQLByProjectIDPaginated_RespectsLimit tests that limit parameter is respected
func TestListGraphQLByProjectIDPaginated_RespectsLimit(t *testing.T) {
	pool, cleanup := SetupPostgres(t)
	defer cleanup()

	ctx := context.Background()
	repo := repository.NewGraphQLAPIRepository(pool)

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

	// Insert 15 test GraphQL APIs
	for i := 1; i <= 15; i++ {
		arguments, _ := json.Marshal([]model.GraphQLArgument{})
		examples, _ := json.Marshal([]model.GraphQLExample{})

		_, err = pool.Exec(ctx, `
			INSERT INTO graphql_apis (id_graphql_api, id_project, id_user, name, type, description, arguments, return_type, examples, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		`, int64(i), 1, 1, "API "+string(rune('A'+byte(i))), "query", "Description "+string(rune('a'+byte(i))), arguments, "String", examples)
		if err != nil {
			t.Fatalf("Failed to insert GraphQL API: %v", err)
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

// TestListGraphQLByProjectIDPaginated_RespectsOffset tests that offset (calculated from page) is respected
func TestListGraphQLByProjectIDPaginated_RespectsOffset(t *testing.T) {
	pool, cleanup := SetupPostgres(t)
	defer cleanup()

	ctx := context.Background()
	repo := repository.NewGraphQLAPIRepository(pool)

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

	// Insert 25 test GraphQL APIs
	for i := 1; i <= 25; i++ {
		arguments, _ := json.Marshal([]model.GraphQLArgument{})
		examples, _ := json.Marshal([]model.GraphQLExample{})

		_, err = pool.Exec(ctx, `
			INSERT INTO graphql_apis (id_graphql_api, id_project, id_user, name, type, description, arguments, return_type, examples, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		`, int64(i), 1, 1, "API "+string(rune('A'+byte(i))), "query", "Description "+string(rune('a'+byte(i))), arguments, "String", examples)
		if err != nil {
			t.Fatalf("Failed to insert GraphQL API: %v", err)
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

	// Verify that first API on page 2 is 11th API (API K)
	if apis[0].Name != "API K" {
		t.Errorf("Expected first API on page 2 to be 'API K', got '%s'", apis[0].Name)
	}
}

// TestListGraphQLByProjectIDPaginated_ProjectNotFound tests querying APIs for non-existent project
func TestListGraphQLByProjectIDPaginated_ProjectNotFound(t *testing.T) {
	pool, cleanup := SetupPostgres(t)
	defer cleanup()

	ctx := context.Background()
	repo := repository.NewGraphQLAPIRepository(pool)

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

// TestListGraphQLByProjectIDPaginated_InvalidPagination tests with invalid pagination parameters
func TestListGraphQLByProjectIDPaginated_InvalidPagination(t *testing.T) {
	pool, cleanup := SetupPostgres(t)
	defer cleanup()

	ctx := context.Background()
	repo := repository.NewGraphQLAPIRepository(pool)

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

	// Insert test GraphQL API
	arguments, _ := json.Marshal([]model.GraphQLArgument{{Name: "id", Type: "ID!"}})
	examples, _ := json.Marshal([]model.GraphQLExample{{Name: "Example", Query: "query { user }"}})
	_, err = pool.Exec(ctx, `
		INSERT INTO graphql_apis (id_graphql_api, id_project, id_user, name, type, description, arguments, return_type, examples, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
	`, 1, 1, 1, "Test API", "query", "Test Description", arguments, "String", examples)
	if err != nil {
		t.Fatalf("Failed to insert GraphQL API: %v", err)
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
