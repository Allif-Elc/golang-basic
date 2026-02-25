package repository

import (
	"context"
	"golang-basic/api/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PolicyRepository struct {
	db *pgxpool.Pool
}

func NewPolicyRepository(db *pgxpool.Pool) *PolicyRepository {
	return &PolicyRepository{db: db}
}

// CreatePolicy creates a new policy
// Required indexes:
//   CREATE UNIQUE INDEX idx_policies_name ON policies(name);
func (r *PolicyRepository) CreatePolicy(ctx context.Context, req *model.CreatePolicyRequest) (*model.Policy, error) {
	query := `INSERT INTO policies (name, policy_rule, is_active) VALUES ($1, $2, $3) RETURNING id_policy, created_at, updated_at`
	var createdPolicy model.Policy
	err := r.db.QueryRow(ctx, query, req.Name, req.PolicyRule, req.IsActive).Scan(
		&createdPolicy.PolicyId,
		&createdPolicy.CreatedAt,
		&createdPolicy.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	createdPolicy.Name = req.Name
	createdPolicy.PolicyRule = req.PolicyRule
	createdPolicy.IsActive = req.IsActive
	return &createdPolicy, nil
}

// GetPolicyByID retrieves a policy by its ID
func (r *PolicyRepository) GetPolicyByID(ctx context.Context, id int64) (*model.Policy, error) {
	query := `SELECT id_policy, name, policy_rule, is_active, created_at, updated_at FROM policies WHERE id_policy = $1`
	var policy model.Policy
	err := r.db.QueryRow(ctx, query, id).Scan(
		&policy.PolicyId,
		&policy.Name,
		&policy.PolicyRule,
		&policy.IsActive,
		&policy.CreatedAt,
		&policy.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &policy, nil
}

// GetPolicyByName retrieves a policy by its name
func (r *PolicyRepository) GetPolicyByName(ctx context.Context, name string) (*model.Policy, error) {
	query := `SELECT id_policy, name, policy_rule, is_active, created_at, updated_at FROM policies WHERE name = $1`
	var policy model.Policy
	err := r.db.QueryRow(ctx, query, name).Scan(
		&policy.PolicyId,
		&policy.Name,
		&policy.PolicyRule,
		&policy.IsActive,
		&policy.CreatedAt,
		&policy.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &policy, nil
}

// ListPolicies retrieves all policies
// Required indexes:
//   CREATE INDEX idx_policies_policy_rule_gin ON policies USING GIN (policy_rule);
func (r *PolicyRepository) ListPolicies(ctx context.Context) ([]model.Policy, error) {
	query := `SELECT id_policy, name, policy_rule, is_active, created_at, updated_at FROM policies ORDER BY id_policy`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var policies []model.Policy
	for rows.Next() {
		var policy model.Policy
		err := rows.Scan(
			&policy.PolicyId,
			&policy.Name,
			&policy.PolicyRule,
			&policy.IsActive,
			&policy.CreatedAt,
			&policy.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		policies = append(policies, policy)
	}

	return policies, rows.Err()
}

// UpdatePolicy updates a policy by ID
func (r *PolicyRepository) UpdatePolicy(ctx context.Context, id int64, req *model.UpdatePolicyRequest) (*model.Policy, error) {
	query := `
		UPDATE policies
		SET name = COALESCE($1, name),
		    policy_rule = COALESCE($2, policy_rule),
		    is_active = COALESCE($3, is_active),
		    updated_at = CURRENT_TIMESTAMP
		WHERE id_policy = $4
		RETURNING id_policy, name, policy_rule, is_active, created_at, updated_at
	`
	var updated model.Policy
	err := r.db.QueryRow(ctx, query, req.Name, req.PolicyRule, req.IsActive, id).Scan(
		&updated.PolicyId,
		&updated.Name,
		&updated.PolicyRule,
		&updated.IsActive,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &updated, nil
}

// DeletePolicy deletes a policy by ID
func (r *PolicyRepository) DeletePolicy(ctx context.Context, id int64) error {
	query := `DELETE FROM policies WHERE id_policy = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return pgx.ErrNoRows
	}

	return nil
}
