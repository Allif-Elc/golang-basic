package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang-basic/api/internal/model"
)

// UserPolicyRepository handles user-policy CRUD operations
// Performance target: <50ms for all queries with proper indexes
type UserPolicyRepository struct {
	db *pgxpool.Pool
}

// NewUserPolicyRepository creates a new user policy repository instance
func NewUserPolicyRepository(db *pgxpool.Pool) *UserPolicyRepository {
	return &UserPolicyRepository{db: db}
}

// CreateUserPolicy creates a new user-to-policy assignment
// Required indexes:
//   UNIQUE(id_user, id_policy) - prevents duplicate assignments
//   idx_user_policies_id_user - user lookups
func (r *UserPolicyRepository) CreateUserPolicy(ctx context.Context, req *model.CreateUserPolicyRequest, createdBy *int64) (*model.UserPolicy, error) {
	query := `
		INSERT INTO user_policies (id_user, id_policy, priority, is_active, expires_at, created_by)
		VALUES ($1, $2, $3, true, $4, $5)
		RETURNING id_user_policy, id_user, id_policy, priority, is_active, expires_at, created_at, created_by
	`

	var created model.UserPolicy
	err := r.db.QueryRow(ctx, query, req.UserID, req.PolicyID, req.Priority, req.ExpiresAt, createdBy).Scan(
		&created.UserPolicyId,
		&created.UserID,
		&created.PolicyID,
		&created.Priority,
		&created.IsActive,
		&created.ExpiresAt,
		&created.CreatedAt,
		&created.CreatedBy,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create user policy: %w", err)
	}

	return &created, nil
}

// GetUserPolicyByID retrieves a user policy by its ID
// Required indexes:
//   PRIMARY KEY on id_user_policy
func (r *UserPolicyRepository) GetUserPolicyByID(ctx context.Context, id int64) (*model.UserPolicy, error) {
	query := `
		SELECT id_user_policy, id_user, id_policy, priority, is_active, expires_at, created_at, created_by
		FROM user_policies
		WHERE id_user_policy = $1
	`

	var userPolicy model.UserPolicy
	err := r.db.QueryRow(ctx, query, id).Scan(
		&userPolicy.UserPolicyId,
		&userPolicy.UserID,
		&userPolicy.PolicyID,
		&userPolicy.Priority,
		&userPolicy.IsActive,
		&userPolicy.ExpiresAt,
		&userPolicy.CreatedAt,
		&userPolicy.CreatedBy,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user policy not found: %d", id)
		}
		return nil, fmt.Errorf("failed to get user policy: %w", err)
	}

	return &userPolicy, nil
}

// ListUserPolicies retrieves all user policies
// Required indexes:
//   idx_user_policies_id_user for user lookups
func (r *UserPolicyRepository) ListUserPolicies(ctx context.Context) ([]model.UserPolicy, error) {
	query := `
		SELECT id_user_policy, id_user, id_policy, priority, is_active, expires_at, created_at, created_by
		FROM user_policies
		ORDER BY priority DESC, created_at DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list user policies: %w", err)
	}
	defer rows.Close()

	userPolicies := make([]model.UserPolicy, 0)
	for rows.Next() {
		var up model.UserPolicy
		err := rows.Scan(
			&up.UserPolicyId,
			&up.UserID,
			&up.PolicyID,
			&up.Priority,
			&up.IsActive,
			&up.ExpiresAt,
			&up.CreatedAt,
			&up.CreatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user policy: %w", err)
		}
		userPolicies = append(userPolicies, up)
	}

	return userPolicies, rows.Err()
}

// ListUserPoliciesByUserID retrieves all policies assigned to a specific user
// Required indexes:
//   idx_user_policies_user_active_expires - covers full authorization query
func (r *UserPolicyRepository) ListUserPoliciesByUserID(ctx context.Context, userID int64) ([]model.UserPolicy, error) {
	query := `
		SELECT id_user_policy, id_user, id_policy, priority, is_active, expires_at, created_at, created_by
		FROM user_policies
		WHERE id_user = $1
		ORDER BY priority DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list user policies for user %d: %w", userID, err)
	}
	defer rows.Close()

	userPolicies := make([]model.UserPolicy, 0)
	for rows.Next() {
		var up model.UserPolicy
		err := rows.Scan(
			&up.UserPolicyId,
			&up.UserID,
			&up.PolicyID,
			&up.Priority,
			&up.IsActive,
			&up.ExpiresAt,
			&up.CreatedAt,
			&up.CreatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user policy: %w", err)
		}
		userPolicies = append(userPolicies, up)
	}

	return userPolicies, rows.Err()
}

// ListUserPolicyDetails retrieves all user policies with related user and policy information
// Uses JOIN to avoid N+1 queries
// Required indexes:
//   idx_user_policies_id_user
//   idx_user_policies_id_policy
//   PRIMARY KEY on users.id_user
//   PRIMARY KEY on policies.id_policy
func (r *UserPolicyRepository) ListUserPolicyDetails(ctx context.Context) ([]model.UserPolicyDetail, error) {
	query := `
		SELECT
			up.id_user_policy,
			up.id_user,
			u.name as user_name,
			u.email as user_email,
			up.id_policy,
			p.name as policy_name,
			p.policy_rule,
			up.priority,
			up.is_active,
			up.expires_at,
			up.created_at
		FROM user_policies up
		INNER JOIN users u ON up.id_user = u.id_user
		INNER JOIN policies p ON up.id_policy = p.id_policy
		ORDER BY up.priority DESC, up.created_at DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list user policy details: %w", err)
	}
	defer rows.Close()

	details := make([]model.UserPolicyDetail, 0)
	for rows.Next() {
		var ud model.UserPolicyDetail
		err := rows.Scan(
			&ud.UserPolicyId,
			&ud.UserID,
			&ud.UserName,
			&ud.UserEmail,
			&ud.PolicyID,
			&ud.PolicyName,
			&ud.PolicyRule,
			&ud.Priority,
			&ud.IsActive,
			&ud.ExpiresAt,
			&ud.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user policy detail: %w", err)
		}
		details = append(details, ud)
	}

	return details, rows.Err()
}

// UpdateUserPolicy updates an existing user policy
// Supports partial updates using COALESCE
func (r *UserPolicyRepository) UpdateUserPolicy(ctx context.Context, id int64, req *model.UpdateUserPolicyRequest) (*model.UserPolicy, error) {
	query := `
		UPDATE user_policies
		SET
			priority = COALESCE($1, priority),
			is_active = COALESCE($2, is_active),
			expires_at = COALESCE($3, expires_at)
		WHERE id_user_policy = $4
		RETURNING id_user_policy, id_user, id_policy, priority, is_active, expires_at, created_at, created_by
	`

	var updated model.UserPolicy
	err := r.db.QueryRow(ctx, query, req.Priority, req.IsActive, req.ExpiresAt, id).Scan(
		&updated.UserPolicyId,
		&updated.UserID,
		&updated.PolicyID,
		&updated.Priority,
		&updated.IsActive,
		&updated.ExpiresAt,
		&updated.CreatedAt,
		&updated.CreatedBy,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user policy not found: %d", id)
		}
		return nil, fmt.Errorf("failed to update user policy: %w", err)
	}

	return &updated, nil
}

// DeleteUserPolicy deletes a user policy by ID
func (r *UserPolicyRepository) DeleteUserPolicy(ctx context.Context, id int64) error {
	query := `DELETE FROM user_policies WHERE id_user_policy = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user policy: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// CheckUserExists verifies if a user exists (for validation)
// Required indexes:
//   PRIMARY KEY on users.id_user
func (r *UserPolicyRepository) CheckUserExists(ctx context.Context, userID int64) (bool, error) {
	query := `SELECT 1 FROM users WHERE id_user = $1 LIMIT 1`

	var exists int
	err := r.db.QueryRow(ctx, query, userID).Scan(&exists)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}

	return true, nil
}

// CheckPolicyExists verifies if a policy exists (for validation)
// Required indexes:
//   PRIMARY KEY on policies.id_policy
func (r *UserPolicyRepository) CheckPolicyExists(ctx context.Context, policyID int64) (bool, error) {
	query := `SELECT 1 FROM policies WHERE id_policy = $1 LIMIT 1`

	var exists int
	err := r.db.QueryRow(ctx, query, policyID).Scan(&exists)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("failed to check policy existence: %w", err)
	}

	return true, nil
}

// CheckDuplicateUserPolicy checks if a user-policy assignment already exists
// Required indexes:
//   UNIQUE constraint on (id_user, id_policy)
func (r *UserPolicyRepository) CheckDuplicateUserPolicy(ctx context.Context, userID, policyID int64) (bool, error) {
	query := `SELECT 1 FROM user_policies WHERE id_user = $1 AND id_policy = $2 LIMIT 1`

	var exists int
	err := r.db.QueryRow(ctx, query, userID, policyID).Scan(&exists)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("failed to check duplicate user policy: %w", err)
	}

	return true, nil
}

// CleanupExpiredPolicies removes expired user policies from the database
// Returns the number of policies deleted
// Required indexes:
//   idx_user_policies_expires_at - partial index on expires_at WHERE NOT NULL
func (r *UserPolicyRepository) CleanupExpiredPolicies(ctx context.Context) (int64, error) {
	query := `DELETE FROM user_policies WHERE expires_at IS NOT NULL AND expires_at < CURRENT_TIMESTAMP`

	result, err := r.db.Exec(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired policies: %w", err)
	}

	return result.RowsAffected(), nil
}
