package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"golang-basic/api/internal/model"
)

// AuthorizationRepositoryInterface defines the contract for authorization data operations
// This allows for both real repository and mock implementations
type AuthorizationRepositoryInterface interface {
	GetUserRoles(ctx context.Context, userID int64) ([]string, error)
	GetMatchingPolicies(ctx context.Context, resource, action string) ([]model.Policy, error)
	GetUserPoliciesByUserID(ctx context.Context, userID int64) ([]model.PolicyWithPriority, error)
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
// Authorization Flow (Priority-Based):
// 1. Fetch user roles (~5-10ms)
// 2. Fetch user-specific policies with priority (~5-10ms)
// 3. Fetch matching role-based policies (~10-20ms)
// 4. Merge and sort by priority (in-memory)
// 5. Evaluate policies in priority order (~1-5ms)
// 6. Log audit decision (~5-10ms)
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

	// Step 2: Fetch user-specific policies with priority (NEW - higher priority)
	userPolicies, err := s.repo.GetUserPoliciesByUserID(ctx, req.UserID)
	if err != nil {
		return model.AuthorizeResponse{}, fmt.Errorf("failed to fetch user policies: %w", err)
	}

	// Step 3: Fetch matching role-based policies (existing)
	policies, err := s.repo.GetMatchingPolicies(ctx, req.Resource, req.Action)
	if err != nil {
		return model.AuthorizeResponse{}, fmt.Errorf("failed to fetch policies: %w", err)
	}

	// Step 4: Merge and sort by priority (NEW)
	// User policies have their assigned priority, role-based policies have priority 0
	allPolicies := s.mergeAndSortPolicies(userPolicies, policies)

	if len(allPolicies) == 0 {
		resp := model.AuthorizeResponse{
			Allowed: false,
			Reason:  "no policies found for resource/action",
		}
		s.logAudit(ctx, req, resp)
		return resp, nil
	}

	// Step 5: Evaluate policies with priority (MODIFIED)
	allowed, reason := s.evaluatePoliciesWithPriority(ctx, roles, allPolicies, req.Resource, req.Action)

	resp := model.AuthorizeResponse{
		Allowed: allowed,
		Reason:  reason,
	}

	// Step 6: Log audit decision (async insert: ~5-10ms)
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

// mergeAndSortPolicies combines user policies and role-based policies, sorting by priority
// User policies have their assigned priority, role-based policies have priority 0
func (s *AuthorizationService) mergeAndSortPolicies(
	userPolicies []model.PolicyWithPriority,
	rolePolicies []model.Policy,
) []model.PolicyWithPriority {
	result := make([]model.PolicyWithPriority, 0, len(userPolicies)+len(rolePolicies))

	// Add user policies with their priority
	result = append(result, userPolicies...)

	// Add role-based policies with priority 0 (lower than user policies)
	for _, p := range rolePolicies {
		result = append(result, model.PolicyWithPriority{Policy: p, Priority: 0})
	}

	// Sort by priority descending (higher priority first)
	sort.Slice(result, func(i, j int) bool {
		return result[i].Priority > result[j].Priority
	})

	return result
}

// evaluatePoliciesWithPriority checks policies in priority order
// First matching policy wins (regardless of allow/deny)
// User policies (priority > 0) are evaluated before role-based policies (priority = 0)
func (s *AuthorizationService) evaluatePoliciesWithPriority(
	ctx context.Context,
	userRoles []string,
	policiesWithPriority []model.PolicyWithPriority,
	resource, action string,
) (bool, string) {
	// Build role set for O(1) lookup
	roleSet := make(map[string]struct{})
	for _, role := range userRoles {
		roleSet[role] = struct{}{}
	}

	// Check each policy in priority order
	for _, pwp := range policiesWithPriority {
		rule, err := parsePolicyRule(pwp.Policy.PolicyRule)
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

		// Validate action matches policy's action array
		if !matchesAction(rule.Action, action) {
			continue
		}

		// Found matching policy
		source := "role-based policy"
		if pwp.Priority > 0 {
			source = fmt.Sprintf("user policy (priority: %d)", pwp.Priority)
		} else if pwp.Priority < 0 {
			source = fmt.Sprintf("user policy (priority: %d, low priority)", pwp.Priority)
		}
		return true, fmt.Sprintf("access granted via %s '%s' with role '%s'", source, pwp.Policy.Name, rule.Role)
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
