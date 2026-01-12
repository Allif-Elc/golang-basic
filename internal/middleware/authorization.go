package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"golang-basic/internal/model"
	"net/http"
	"strconv"
	"strings"
)

type contextKey string

const (
	UserIDKey    contextKey = "user_id"
	UserRolesKey contextKey = "user_roles"
)

type AuthorizationServiceInterface interface {
	Authorize(ctx context.Context, req model.AuthorizeRequest) (model.AuthorizeResponse, error)
}

type AuthorizationMiddleware struct {
	authService AuthorizationServiceInterface
}

func NewAuthorizationMiddleware(authService AuthorizationServiceInterface) *AuthorizationMiddleware {
	return &AuthorizationMiddleware{authService: authService}
}

func (m *AuthorizationMiddleware) RequirePermission(resource, action string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, err := m.extractUserID(r)
			if err != nil {
				m.sendErrorResponse(w, http.StatusUnauthorized, "Failed to extract user ID", err)
				return
			}

			authReq := model.AuthorizeRequest{
				UserID:   userID,
				Resource: resource,
				Action:   action,
			}

			authResp, err := m.authService.Authorize(r.Context(), authReq)
			if err != nil {
				m.sendErrorResponse(w, http.StatusInternalServerError, "Authorization check failed", err)
				return
			}

			// Check if authorized
			if !authResp.Allowed {
				m.sendErrorResponse(w, http.StatusForbidden, authResp.Reason, nil)
				return
			}

			// Authorization successful, proceed to next handler
			next.ServeHTTP(w, r)
		})
	}
}

func (m *AuthorizationMiddleware) RequirePermissionCustom(
	resourceExtractor, actionExtractor func(r *http.Request) (string, string),
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, err := m.extractUserID(r)
			if err != nil {
				m.sendErrorResponse(w, http.StatusUnauthorized, "Failed to extract user ID", err)
				return
			}

			// Extract resource and action dynamically
			resource, action := resourceExtractor(r)
			resource, action = actionExtractor(r)

			// Perform authorization check
			authReq := model.AuthorizeRequest{
				UserID:   userID,
				Resource: resource,
				Action:   action,
			}

			authResp, err := m.authService.Authorize(r.Context(), authReq)
			if err != nil {
				m.sendErrorResponse(w, http.StatusInternalServerError, "Authorization check failed", err)
				return
			}

			if !authResp.Allowed {
				m.sendErrorResponse(w, http.StatusForbidden, authResp.Reason, nil)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (m *AuthorizationMiddleware) RequireRole(requiredRole string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, err := m.extractUserID(r)
			if err != nil {
				m.sendErrorResponse(w, http.StatusUnauthorized, "Failed to extract user ID", err)
				return
			}

			// Use "any" as resource to check if user has the role
			authReq := model.AuthorizeRequest{
				UserID:   userID,
				Resource: "any",
				Action:   requiredRole,
			}

			authResp, err := m.authService.Authorize(r.Context(), authReq)
			if err != nil {
				m.sendErrorResponse(w, http.StatusInternalServerError, "Role check failed", err)
				return
			}

			if !authResp.Allowed {
				m.sendErrorResponse(w, http.StatusForbidden,
					fmt.Sprintf("Role '%s' required", requiredRole), nil)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (m *AuthorizationMiddleware) extractUserID(r *http.Request) (int64, error) {
	// Try context first (preferred method)
	if ctxUserID := r.Context().Value(UserIDKey); ctxUserID != nil {
		if uid, ok := ctxUserID.(int64); ok {
			return uid, nil
		}
	}

	// Try X-User-ID header (for testing/simple cases)
	if headerUserID := r.Header.Get("X-User-ID"); headerUserID != "" {
		uid, err := strconv.ParseInt(headerUserID, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid X-User-ID header: %w", err)
		}
		return uid, nil
	}

	// Try Authorization header (JWT - placeholder)
	// This would decode JWT and extract user_id claim
	// For now, return error as this is not implemented
	if authHeader := r.Header.Get("Authorization"); authHeader != "" {
		return 0, fmt.Errorf("JWT authentication not yet implemented, use X-User-ID header or context")
	}

	return 0, fmt.Errorf("user ID not found in request")
}

func (m *AuthorizationMiddleware) sendErrorResponse(w http.ResponseWriter, status int, message string, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := map[string]interface{}{
		"status":  "error",
		"message": message,
	}

	if err != nil {
		response["error"] = err.Error()
	}

	json.NewEncoder(w).Encode(response)
}

func ResourceFromPath(r *http.Request) string {
	path := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(path, "/")
	if len(parts) > 0 {
		// Convert kebab-case to snake_case and remove 'api' prefix
		resource := parts[len(parts)-1]
		if resource == "api" && len(parts) > 1 {
			resource = parts[len(parts)-2]
		}
		return strings.ReplaceAll(resource, "-", "_")
	}
	return "unknown"
}

// ActionFromMethod maps HTTP methods to actions
func ActionFromMethod(r *http.Request) string {
	switch r.Method {
	case http.MethodGet:
		return "read"
	case http.MethodPost, http.MethodPut:
		return "write"
	case http.MethodDelete:
		return "delete"
	default:
		return "unknown"
	}
}

func SetUserIDInContext() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userIDStr := r.Header.Get("X-User-ID")
			if userIDStr != "" {
				userID, err := strconv.ParseInt(userIDStr, 10, 64)
				if err == nil {
					ctx := context.WithValue(r.Context(), UserIDKey, userID)
					r = r.WithContext(ctx)
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
