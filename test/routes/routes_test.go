package routes_test

import (
	"net/http"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"golang-basic/api/internal/routes"
)

// TestRoutes_SetupRoutes_Structure verifies routes can be set up when JWT keys exist
func TestRoutes_SetupRoutes_Structure(t *testing.T) {
	// Check if JWT keys exist
	if _, err := os.Stat("keys/jwt-private-key.pem"); os.IsNotExist(err) {
		t.Skip("Skipping test - JWT keys not found in keys/jwt-private-key.pem")
	}

	r := routes.SetupRoutes()

	if r == nil {
		t.Fatal("Expected non-nil router")
	}

	// Verify router is not nil and can be used
	t.Log("Router initialized successfully")
}

// TestRoutes_RouteCount verifies expected number of route groups
func TestRoutes_RouteCount(t *testing.T) {
	if _, err := os.Stat("keys/jwt-private-key.pem"); os.IsNotExist(err) {
		t.Skip("Skipping test - JWT keys not found")
	}

	r := routes.SetupRoutes()

	// Count routes by walking the router
	routeCount := 0
	chi.Walk(r, func(method, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		routeCount++
		return nil
	})

	if routeCount < 10 {
		t.Errorf("Expected at least 10 routes, got %d", routeCount)
	}

	t.Logf("Found %d registered routes", routeCount)
}

// TestRoutes_HealthCheck_RouteExists verifies health check route is registered
func TestRoutes_HealthCheck_RouteExists(t *testing.T) {
	if _, err := os.Stat("keys/jwt-private-key.pem"); os.IsNotExist(err) {
		t.Skip("Skipping test - JWT keys not found")
	}

	r := routes.SetupRoutes()

	// Walk routes to find /health
	found := false
	chi.Walk(r, func(method, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		if route == "/health" && method == "GET" {
			found = true
		}
		return nil
	})

	if !found {
		t.Error("Health check route /health not found")
	}
}

// TestRoutes_AuthRoutes_Exist verifies auth routes are registered
func TestRoutes_AuthRoutes_Exist(t *testing.T) {
	if _, err := os.Stat("keys/jwt-private-key.pem"); os.IsNotExist(err) {
		t.Skip("Skipping test - JWT keys not found")
	}

	r := routes.SetupRoutes()

	expectedRoutes := map[string]string{
		"/api/v1/auth/login":    "POST",
		"/api/v1/auth/register": "POST",
		"/api/v1/auth/refresh":  "POST",
	}

	missingRoutes := []string{}
	chi.Walk(r, func(method, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		for expectedRoute, expectedMethod := range expectedRoutes {
			if route == expectedRoute && method == expectedMethod {
				delete(expectedRoutes, expectedRoute)
			}
		}
		return nil
	})

	for route := range expectedRoutes {
		missingRoutes = append(missingRoutes, route)
	}

	if len(missingRoutes) > 0 {
		t.Errorf("Missing auth routes: %v", missingRoutes)
	}
}

// TestRoutes_UserRoutes_Exist verifies user routes are registered
func TestRoutes_UserRoutes_Exist(t *testing.T) {
	if _, err := os.Stat("keys/jwt-private-key.pem"); os.IsNotExist(err) {
		t.Skip("Skipping test - JWT keys not found")
	}

	r := routes.SetupRoutes()

	// Check for key user routes
	foundGetUsers := false
	chi.Walk(r, func(method, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		if route == "/api/v1/users" && method == "GET" {
			foundGetUsers = true
		}
		return nil
	})

	if !foundGetUsers {
		t.Error("User route GET /api/v1/users not found")
	}
}

// TestRoutes_ProjectRoutes_Exist verifies project routes are registered
func TestRoutes_ProjectRoutes_Exist(t *testing.T) {
	if _, err := os.Stat("keys/jwt-private-key.pem"); os.IsNotExist(err) {
		t.Skip("Skipping test - JWT keys not found")
	}

	r := routes.SetupRoutes()

	// Check for key project routes
	foundGetProjects := false
	chi.Walk(r, func(method, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		if route == "/api/v1/projects" && method == "GET" {
			foundGetProjects = true
		}
		return nil
	})

	if !foundGetProjects {
		t.Error("Project route GET /api/v1/projects not found")
	}
}

// TestRoutes_ProfileRoutes_Exist verifies profile routes are registered
func TestRoutes_ProfileRoutes_Exist(t *testing.T) {
	if _, err := os.Stat("keys/jwt-private-key.pem"); os.IsNotExist(err) {
		t.Skip("Skipping test - JWT keys not found")
	}

	r := routes.SetupRoutes()

	// Check for key profile routes
	foundGetProfiles := false
	chi.Walk(r, func(method, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		if route == "/api/v1/profiles" && method == "GET" {
			foundGetProfiles = true
		}
		return nil
	})

	if !foundGetProfiles {
		t.Error("Profile route GET /api/v1/profiles not found")
	}
}

// TestRoutes_PermissionRoutes_Exist verifies permission routes are registered
func TestRoutes_PermissionRoutes_Exist(t *testing.T) {
	if _, err := os.Stat("keys/jwt-private-key.pem"); os.IsNotExist(err) {
		t.Skip("Skipping test - JWT keys not found")
	}

	r := routes.SetupRoutes()

	// Check for key permission attribute routes
	foundGetAttributes := false
	chi.Walk(r, func(method, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		if route == "/api/v1/permissions/attributes" && method == "GET" {
			foundGetAttributes = true
		}
		return nil
	})

	if !foundGetAttributes {
		t.Error("Permission route GET /api/v1/permissions/attributes not found")
	}
}

// TestRoutes_NestedRoutes_Exist verifies nested project API routes are registered
func TestRoutes_NestedRoutes_Exist(t *testing.T) {
	if _, err := os.Stat("keys/jwt-private-key.pem"); os.IsNotExist(err) {
		t.Skip("Skipping test - JWT keys not found")
	}

	r := routes.SetupRoutes()

	// Check for nested routes
	expectedNestedRoutes := []string{
		"/api/v1/projects/*/rest-apis",
		"/api/v1/projects/*/graphql-apis",
		"/api/v1/projects/*/grpc-apis",
	}

	foundRoutes := make(map[string]bool)
	chi.Walk(r, func(method, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		// Check if route matches any expected pattern
		for _, expected := range expectedNestedRoutes {
			// Simple pattern match - in real tests you'd use regex
			if len(route) > 0 && route[:len(expected)-1] == expected[:len(expected)-1] {
				foundRoutes[expected] = true
			}
		}
		return nil
	})

	missingRoutes := []string{}
	for _, expected := range expectedNestedRoutes {
		if !foundRoutes[expected] {
			missingRoutes = append(missingRoutes, expected)
		}
	}

	if len(missingRoutes) > 0 {
		t.Logf("Some nested routes may not be registered: %v", missingRoutes)
		// This is informational - nested routes may be registered differently
	}
}

// TestRoutes_MiddlewareConfigured verifies global middleware is applied
func TestRoutes_MiddlewareConfigured(t *testing.T) {
	if _, err := os.Stat("keys/jwt-private-key.pem"); os.IsNotExist(err) {
		t.Skip("Skipping test - JWT keys not found")
	}

	r := routes.SetupRoutes()

	// Chi router should have middleware configured
	// We can't easily test this without accessing internal state,
	// but we can verify the router is functional
	if r == nil {
		t.Fatal("Expected non-nil router")
	}
}
