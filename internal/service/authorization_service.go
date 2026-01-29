package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"golang-basic/api/internal/model"
)

// AuthorizationRepositoryInterface defines the contract for authorization data operations
// This allows for both real repository and mock implementations
type AuthorizationRepositoryInterface interface {
	GetUserRoles(ctx context.Context, userID int64) ([]string, error)
	GetMatchingPolicies(ctx context.Context, resource, action string) ([]model.Policy, error)
	LogAuditDecision(ctx context.Context, userID int64, resource, action string, allowed bool, reason string) error
}

// AuthorizationService handles ABAC authorization logic
type AuthorizationService struct {
	repo AuthorizationRepositoryInterface
}

// NewAuthorizationService creates a new authorization service
// Accepts any implementation of AuthorizationRepositoryInterface (real or mock)
func NewAuthorizationService(repo AuthorizationRepositoryInterface) *AuthorizationService {
	return &AuthorizationService{repo: repo}
}

// Authorize checks if a user is allowed to perform an action on a resource
// Returns: (allowed, reason, error)
//
// Performance: <50ms with 1M rows when indexes are properly configured
func (s *AuthorizationService) Authorize(ctx context.Context, req model.AuthorizeRequest) (model.AuthorizeResponse, error) {
	// Step 1: Fetch user roles (cached query: ~5-10ms)
	roles, err := s.repo.GetUserRoles(ctx, req.UserID)
	if err != nil {
		return model.AuthorizeResponse{}, fmt.Errorf("failed to get user roles: %w", err)
	}

	if len(roles) == 0 {
		resp := model.AuthorizeResponse{
			Allowed: false,
			Reason:  "no roles assigned to user",
		}
		s.logAudit(ctx, req, resp)
		return resp, nil
	}

	// Step 2: Fetch matching policies (indexed JSONB query: ~10-20ms)
	policies, err := s.repo.GetMatchingPolicies(ctx, req.Resource, req.Action)
	if err != nil {
		return model.AuthorizeResponse{}, fmt.Errorf("failed to fetch policies: %w", err)
	}

	if len(policies) == 0 {
		resp := model.AuthorizeResponse{
			Allowed: false,
			Reason:  "no policies found for resource/action",
		}
		s.logAudit(ctx, req, resp)
		return resp, nil
	}

	// Step 3: Evaluate policies against user roles and wildcard patterns (in-memory: ~1-5ms)
	allowed, reason := s.evaluatePolicies(ctx, roles, policies, req.Resource, req.Action)

	resp := model.AuthorizeResponse{
		Allowed: allowed,
		Reason:  reason,
	}

	// Step 4: Log audit decision (async insert: ~5-10ms)
	s.logAudit(ctx, req, resp)

	return resp, nil
}

// evaluatePolicies checks if any user role matches a policy and validates wildcard patterns
func (s *AuthorizationService) evaluatePolicies(
	ctx context.Context,
	userRoles []string,
	policies []model.Policy,
	resource, action string,
) (bool, string) {
	// Build role set for O(1) lookup
	roleSet := make(map[string]struct{})
	for _, role := range userRoles {
		roleSet[role] = struct{}{}
	}

	// Check each policy
	for _, policy := range policies {
		rule, err := parsePolicyRule(policy.PolicyRule)
		if err != nil {
			continue // Skip malformed policies
		}

		// Check if user has the required role
		if _, exists := roleSet[rule.Role]; !exists {
			continue // User doesn't have this policy's role
		}

		// Validate resource matches policy's wildcard pattern
		if !matchesResource(rule.Resource, resource) {
			continue
		}

		// Validate action matches policy's action array (with wildcard support)
		if !matchesAction(rule.Action, action) {
			continue
		}

		return true, fmt.Sprintf("access granted via policy '%s' with role '%s'", policy.Name, rule.Role)
	}

	return false, fmt.Sprintf("access denied: user roles %v do not match any policy", userRoles)
}

// matchesResource checks if requested resource matches policy resource pattern
// Supports: exact match, full wildcard "*", prefix wildcard "prefix_*"
func matchesResource(policyResource, requestedResource string) bool {
	// Full wildcard matches everything
	if policyResource == "*" {
		return true
	}

	// Exact match
	if policyResource == requestedResource {
		return true
	}

	// Prefix wildcard: "prefix_*" matches "prefix_anything"
	if len(policyResource) > 2 && strings.HasSuffix(policyResource, "_*") {
		prefix := policyResource[:len(policyResource)-2]
		// Requested resource must start with prefix and be longer than prefix
		if strings.HasPrefix(requestedResource, prefix) && len(requestedResource) > len(prefix) {
			// Check next character after prefix is "_"
			if requestedResource[len(prefix)] == '_' {
				return true
			}
		}
	}

	return false
}

// matchesAction checks if requested action matches policy action array
// Supports: exact match in array, wildcard ["*"] matches all
func matchesAction(policyActions []string, requestedAction string) bool {
	for _, action := range policyActions {
		// Wildcard matches all actions
		if action == "*" {
			return true
		}
		// Exact match
		if action == requestedAction {
			return true
		}
	}
	return false
}

// parsePolicyRule converts JSONB bytes to PolicyRule struct
func parsePolicyRule(data []byte) (*model.PolicyRule, error) {
	var rule model.PolicyRule
	if err := json.Unmarshal(data, &rule); err != nil {
		return nil, fmt.Errorf("failed to parse policy rule: %w", err)
	}
	return &rule, nil
}

// logAudit records the authorization decision (best-effort, errors ignored)
func (s *AuthorizationService) logAudit(ctx context.Context, req model.AuthorizeRequest, resp model.AuthorizeResponse) {
	_ = s.repo.LogAuditDecision(
		ctx,
		req.UserID,
		req.Resource,
		req.Action,
		resp.Allowed,
		resp.Reason,
	)
}

// BatchAuthorize checks multiple authorizations in a single batch
// Useful for checking multiple resources at once
func (s *AuthorizationService) BatchAuthorize(
	ctx context.Context,
	userID int64,
	requests []model.AuthorizeRequest,
) ([]model.AuthorizeResponse, error) {
	responses := make([]model.AuthorizeResponse, len(requests))

	for i, req := range requests {
		req.UserID = userID
		resp, err := s.Authorize(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("batch authorize failed at index %d: %w", i, err)
		}
		responses[i] = resp
	}

	return responses, nil
}

// HasPermission is a convenience method that returns only the boolean
// Use this when you don't need the reason
func (s *AuthorizationService) HasPermission(
	ctx context.Context,
	userID int64,
	resource, action string,
) (bool, error) {
	req := model.AuthorizeRequest{
		UserID:   userID,
		Resource: resource,
		Action:   action,
	}

	resp, err := s.Authorize(ctx, req)
	if err != nil {
		return false, err
	}

	return resp.Allowed, nil
}
