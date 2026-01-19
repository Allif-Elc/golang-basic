package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"golang-basic/api/internal/model"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthorizationRepository struct {
	db *pgxpool.Pool
}

func NewAuthorizationRepository(db *pgxpool.Pool) *AuthorizationRepository {
	return &AuthorizationRepository{db: db}
}

func (r *AuthorizationRepository) GetUserRoles(ctx context.Context, userID int64) ([]string, error) {
	// Required indexes:
	//   CREATE INDEX idx_user_attributes_user ON user_attributes(id_user);
	//   CREATE INDEX idx_user_attributes_attribute ON user_attributes(id_attribute);
	//   CREATE INDEX idx_attributes_id ON attributes(id_user_attribute);
	//   CREATE INDEX idx_attributes_name ON attributes(name);
	query := `
		SELECT ua.value
		FROM user_attributes ua
		INNER JOIN attributes a ON ua.id_attribute = a.id_user_attribute
		WHERE ua.id_user = $1 AND a.name = 'role'
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user roles: %w", err)
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}
		roles = append(roles, value)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating roles: %w", err)
	}

	return roles, nil
}

// GetMatchingPolicies fetches policies that match the resource and action
// Required indexes:
//   CREATE INDEX idx_policies_is_active ON policies(is_active);
//   CREATE INDEX idx_policies_rule_resource ON policies USING GIN ((policy_rule->>'resource'));
//   CREATE INDEX idx_policies_rule_action ON policies USING GIN ((policy_rule->>'action'));
func (r *AuthorizationRepository) GetMatchingPolicies(ctx context.Context, resource, action string) ([]model.Policy, error) {
	query := `
		SELECT policy_id, name, policy_rule, is_active, created_at, updated_at
		FROM policies
		WHERE is_active = true
		  AND policy_rule->>'resource' = $1
		  AND policy_rule->>'action' = $2
	`

	rows, err := r.db.Query(ctx, query, resource, action)
	if err != nil {
		return nil, fmt.Errorf("failed to query policies: %w", err)
	}
	defer rows.Close()

	var policies []model.Policy
	for rows.Next() {
		var p model.Policy
		if err := rows.Scan(&p.PolicyId, &p.Name, &p.PolicyRule, &p.IsActive, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan policy: %w", err)
		}
		policies = append(policies, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating policies: %w", err)
	}

	return policies, nil
}

// LogAuditDecision records authorization decision for compliance
// Required indexes:
//   CREATE INDEX idx_audit_logs_user ON audit_logs(user_id);
//   CREATE INDEX idx_audit_logs_resource ON audit_logs(resource);
//   CREATE INDEX idx_audit_logs_action ON audit_logs(action);
//   CREATE INDEX idx_audit_logs_created ON audit_logs(created_at);
func (r *AuthorizationRepository) LogAuditDecision(
	ctx context.Context,
	userID int64,
	resource, action string,
	allowed bool,
	reason string,
) error {
	query := `
		INSERT INTO audit_logs (user_id, resource, action, allowed, reason, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.Exec(ctx, query, userID, resource, action, allowed, reason, time.Now())
	if err != nil {
		return fmt.Errorf("failed to log audit: %w", err)
	}

	return nil
}

// ParsePolicyRule converts JSONB bytes to PolicyRule struct
func ParsePolicyRule(data []byte) (*model.PolicyRule, error) {
	var rule model.PolicyRule
	if err := json.Unmarshal(data, &rule); err != nil {
		return nil, fmt.Errorf("failed to parse policy rule: %w", err)
	}
	return &rule, nil
}

// PrepareStatements is a no-op for pgxpool as it handles statement preparation internally
// This method is kept for compatibility but does nothing
func (r *AuthorizationRepository) PrepareStatements(ctx context.Context) error {
	// pgxpool handles statement preparation internally
	// This method is kept for API compatibility but does nothing
	return nil
}
