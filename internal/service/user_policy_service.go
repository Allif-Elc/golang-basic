package service

import (
	"context"
	"fmt"
	"time"

	"golang-basic/api/internal/model"
	"golang-basic/api/internal/repository"
)

// UserPolicyService handles user-policy business logic
// Priority-based policy override system for ABAC
type UserPolicyService struct {
	userPolicyRepo      *repository.UserPolicyRepository
	authorizationRepo   *repository.AuthorizationRepository
}

// NewUserPolicyService creates a new user policy service instance
func NewUserPolicyService(userPolicyRepo *repository.UserPolicyRepository, authorizationRepo *repository.AuthorizationRepository) *UserPolicyService {
	return &UserPolicyService{
		userPolicyRepo:    userPolicyRepo,
		authorizationRepo: authorizationRepo,
	}
}

// CreateUserPolicy creates a new user-to-policy assignment with validation
// Validates: user exists, policy exists, no duplicate assignment, priority range
func (s *UserPolicyService) CreateUserPolicy(ctx context.Context, req *model.CreateUserPolicyRequest, createdBy *int64) (*model.UserPolicy, error) {
	// Validate priority range
	if req.Priority < -100 || req.Priority > 100 {
		return nil, ValidationError{
			Field:   "priority",
			Message: "priority must be between -100 and 100",
		}
	}

	// Validate user exists
	userExists, err := s.userPolicyRepo.CheckUserExists(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}
	if !userExists {
		return nil, ValidationError{
			Field:   "id_user",
			Message: "user does not exist",
		}
	}

	// Validate policy exists
	policyExists, err := s.userPolicyRepo.CheckPolicyExists(ctx, req.PolicyID)
	if err != nil {
		return nil, fmt.Errorf("failed to check policy existence: %w", err)
	}
	if !policyExists {
		return nil, ValidationError{
			Field:   "id_policy",
			Message: "policy does not exist",
		}
	}

	// Validate no duplicate assignment
	duplicate, err := s.userPolicyRepo.CheckDuplicateUserPolicy(ctx, req.UserID, req.PolicyID)
	if err != nil {
		return nil, fmt.Errorf("failed to check duplicate assignment: %w", err)
	}
	if duplicate {
		return nil, ValidationError{
			Field:   "assignment",
			Message: "user already has this policy assigned",
		}
	}

	// Validate expiration date is in the future (if provided)
	if req.ExpiresAt != nil && req.ExpiresAt.Before(time.Now()) {
		return nil, ValidationError{
			Field:   "expires_at",
			Message: "expiration date must be in the future",
		}
	}

	// Create user policy
	created, err := s.userPolicyRepo.CreateUserPolicy(ctx, req, createdBy)
	if err != nil {
		return nil, fmt.Errorf("failed to create user policy: %w", err)
	}

	return created, nil
}

// GetUserPolicyByID retrieves a user policy by ID with validation
func (s *UserPolicyService) GetUserPolicyByID(ctx context.Context, id int64) (*model.UserPolicy, error) {
	if id <= 0 {
		return nil, ValidationError{
			Field:   "id",
			Message: "must be a positive integer",
		}
	}

	userPolicy, err := s.userPolicyRepo.GetUserPolicyByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve user policy: %w", err)
	}

	return userPolicy, nil
}

// ListUserPolicies retrieves all user policies
func (s *UserPolicyService) ListUserPolicies(ctx context.Context) ([]model.UserPolicy, error) {
	userPolicies, err := s.userPolicyRepo.ListUserPolicies(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list user policies: %w", err)
	}

	return userPolicies, nil
}

// ListUserPoliciesByUserID retrieves all policies for a specific user
func (s *UserPolicyService) ListUserPoliciesByUserID(ctx context.Context, userID int64) ([]model.UserPolicy, error) {
	if userID <= 0 {
		return nil, ValidationError{
			Field:   "id_user",
			Message: "must be a positive integer",
		}
	}

	userPolicies, err := s.userPolicyRepo.ListUserPoliciesByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list user policies: %w", err)
	}

	return userPolicies, nil
}

// ListUserPolicyDetails retrieves all user policies with related information
// Uses JOIN for optimal performance (N+1 prevention)
func (s *UserPolicyService) ListUserPolicyDetails(ctx context.Context) ([]model.UserPolicyDetail, error) {
	details, err := s.userPolicyRepo.ListUserPolicyDetails(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list user policy details: %w", err)
	}

	return details, nil
}

// UpdateUserPolicy updates an existing user policy with validation
func (s *UserPolicyService) UpdateUserPolicy(ctx context.Context, id int64, req *model.UpdateUserPolicyRequest) (*model.UserPolicy, error) {
	if id <= 0 {
		return nil, ValidationError{
			Field:   "id",
			Message: "must be a positive integer",
		}
	}

	// Validate priority range if provided
	if req.Priority != nil && (*req.Priority < -100 || *req.Priority > 100) {
		return nil, ValidationError{
			Field:   "priority",
			Message: "priority must be between -100 and 100",
		}
	}

	// Validate expiration date is in the future if provided
	if req.ExpiresAt != nil && req.ExpiresAt.Before(time.Now()) {
		return nil, ValidationError{
			Field:   "expires_at",
			Message: "expiration date must be in the future",
		}
	}

	// Update user policy
	updated, err := s.userPolicyRepo.UpdateUserPolicy(ctx, id, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update user policy: %w", err)
	}

	return updated, nil
}

// DeleteUserPolicy deletes a user policy by ID with validation
func (s *UserPolicyService) DeleteUserPolicy(ctx context.Context, id int64) error {
	if id <= 0 {
		return ValidationError{
			Field:   "id",
			Message: "must be a positive integer",
		}
	}

	err := s.userPolicyRepo.DeleteUserPolicy(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete user policy: %w", err)
	}

	return nil
}

// CleanupExpiredPolicies removes expired user policies
// This should be called periodically (e.g., via cron job)
// Returns the number of policies deleted
func (s *UserPolicyService) CleanupExpiredPolicies(ctx context.Context) (int64, error) {
	count, err := s.userPolicyRepo.CleanupExpiredPolicies(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired policies: %w", err)
	}
	return count, nil
}

// ListPolicies retrieves all available policies for use in user policy assignments
func (s *UserPolicyService) ListPolicies(ctx context.Context) ([]model.Policy, error) {
	policies, err := s.authorizationRepo.ListPolicies(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list policies: %w", err)
	}
	return policies, nil
}
