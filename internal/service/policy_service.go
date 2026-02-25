package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"golang-basic/api/internal/model"
	"golang-basic/api/internal/repository"
	"regexp"
	"strings"
	"unicode/utf8"
)

// Policy-specific validation constants
const (
	MinPolicyNameLength = 1
	MaxPolicyNameLength = 255
)

var (
	// Policy name regex - allows alphanumeric, hyphens, underscores
	policyNameRegex = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
	// Resource regex - allows alphanumeric, underscores, and wildcards
	resourceRegex = regexp.MustCompile(`^[A-Za-z0-9_*]+$`)
)

type PolicyService struct {
	repo *repository.PolicyRepository
}

func NewPolicyService(repo *repository.PolicyRepository) *PolicyService {
	return &PolicyService{repo: repo}
}

// validatePolicyName validates policy name according to OWASP standards
func (s *PolicyService) validatePolicyName(name string) error {
	nameLen := utf8.RuneCountInString(name)

	if nameLen < MinPolicyNameLength || nameLen > MaxPolicyNameLength {
		return ValidationError{
			Field:   "name",
			Message: fmt.Sprintf("length must be between %d and %d characters", MinPolicyNameLength, MaxPolicyNameLength),
		}
	}

	if !policyNameRegex.MatchString(name) {
		return ValidationError{
			Field:   "name",
			Message: "only alphanumeric characters, hyphens, and underscores are allowed (A-Z, a-z, 0-9, -, _)",
		}
	}

	// SQL injection prevention
	dangerousPatterns := []string{
		"SELECT", "INSERT", "UPDATE", "DELETE", "DROP", "UNION",
		"OR", "AND", "--", ";", "/*", "*/", "xp_", "sp_",
	}
	upperName := strings.ToUpper(name)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(upperName, pattern) {
			return ValidationError{
				Field:   "name",
				Message: "contains potentially dangerous patterns",
			}
		}
	}

	return nil
}

// validatePolicyRule validates the policy rule JSON structure
func (s *PolicyService) validatePolicyRule(ruleJSON []byte) error {
	var rule model.PolicyRule
	if err := json.Unmarshal(ruleJSON, &rule); err != nil {
		return ValidationError{
			Field:   "policy_rule",
			Message: "invalid JSON format",
		}
	}

	// Validate role
	if rule.Role == "" {
		return ValidationError{
			Field:   "role",
			Message: "role is required",
		}
	}

	// Validate resource
	if rule.Resource == "" {
		return ValidationError{
			Field:   "resource",
			Message: "resource is required",
		}
	}

	// Resource can contain wildcards but must be valid pattern
	if !resourceRegex.MatchString(rule.Resource) {
		return ValidationError{
			Field:   "resource",
			Message: "resource can only contain letters, numbers, underscores, and wildcards (*)",
		}
	}

	// Validate actions
	if len(rule.Action) == 0 {
		return ValidationError{
			Field:   "action",
			Message: "at least one action is required",
		}
	}

	return nil
}

// CreatePolicy creates a new policy with validation
func (s *PolicyService) CreatePolicy(ctx context.Context, req model.CreatePolicyRequest) (*model.Policy, error) {
	// Validate name
	if err := s.validatePolicyName(req.Name); err != nil {
		return nil, err
	}

	// Validate policy rule JSON
	if err := s.validatePolicyRule(req.PolicyRule); err != nil {
		return nil, err
	}

	// Check if policy with same name already exists
	existing, err := s.repo.GetPolicyByName(ctx, req.Name)
	if err == nil && existing != nil {
		return nil, ValidationError{
			Field:   "name",
			Message: "policy with this name already exists",
		}
	}

	// Create policy
	created, err := s.repo.CreatePolicy(ctx, &req)
	if err != nil {
		return nil, fmt.Errorf("failed to create policy: %w", err)
	}

	return created, nil
}

// GetPolicyByID retrieves a policy by ID with validation
func (s *PolicyService) GetPolicyByID(ctx context.Context, id int64) (*model.Policy, error) {
	if id <= 0 {
		return nil, ValidationError{
			Field:   "id",
			Message: "must be a positive integer",
		}
	}

	policy, err := s.repo.GetPolicyByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve policy: %w", err)
	}

	return policy, nil
}

// ListPolicies retrieves all policies
func (s *PolicyService) ListPolicies(ctx context.Context) ([]model.Policy, error) {
	policies, err := s.repo.ListPolicies(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list policies: %w", err)
	}

	return policies, nil
}

// UpdatePolicy updates a policy by ID with validation
func (s *PolicyService) UpdatePolicy(ctx context.Context, id int64, req model.UpdatePolicyRequest) (*model.Policy, error) {
	if id <= 0 {
		return nil, ValidationError{
			Field:   "id",
			Message: "must be a positive integer",
		}
	}

	// Validate name if provided
	if req.Name != nil {
		if err := s.validatePolicyName(*req.Name); err != nil {
			return nil, err
		}
	}

	// Validate policy rule if provided
	if req.PolicyRule != nil {
		if err := s.validatePolicyRule(*req.PolicyRule); err != nil {
			return nil, err
		}
	}

	updated, err := s.repo.UpdatePolicy(ctx, id, &req)
	if err != nil {
		if errors.Is(err, errors.New("no rows in result set")) {
			return nil, ValidationError{
				Field:   "id",
				Message: "policy not found",
			}
		}
		return nil, fmt.Errorf("failed to update policy: %w", err)
	}

	return updated, nil
}

// DeletePolicy deletes a policy by ID with validation
func (s *PolicyService) DeletePolicy(ctx context.Context, id int64) error {
	if id <= 0 {
		return ValidationError{
			Field:   "id",
			Message: "must be a positive integer",
		}
	}

	err := s.repo.DeletePolicy(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete policy: %w", err)
	}

	return nil
}
