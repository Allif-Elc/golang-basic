package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"golang-basic/api/internal/model"
	"time"
)

type AuthorizationRepository struct {
	db *sql.DB
}

func NewAuthorizationRepository(db *sql.DB) *AuthorizationRepository {
	return &AuthorizationRepository{db: db}
}

func (r *AuthorizationRepository) GetUserRoles(ctx context.Context, userID int64) ([]string, error) {
	query := `
		SELECT ua.value
		FROM user_attributes ua
		INNER JOIN attributes a ON ua.id_attribute = a.id_user_attribute
		WHERE ua.id_user = $1 AND a.name = 'role'
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
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
// Uses GIN index on policy_rule JSONB for efficient JSON queries
func (r *AuthorizationRepository) GetMatchingPolicies(ctx context.Context, resource, action string) ([]model.Policy, error) {
	query := `
		SELECT policy_id, name, policy_rule, is_active, created_at, updated_at
		FROM policies
		WHERE is_active = true
		  AND policy_rule->>'resource' = $1
		  AND policy_rule->>'action' = $2
	`

	rows, err := r.db.QueryContext(ctx, query, resource, action)
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

	_, err := r.db.ExecContext(ctx, query, userID, resource, action, allowed, reason, time.Now())
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

// PrepareStatements prepares frequently used queries for better performance
func (r *AuthorizationRepository) PrepareStatements(ctx context.Context) error {
	// Prepare user roles query
	_, err := r.db.PrepareContext(ctx, `
		SELECT ua.value
		FROM user_attributes ua
		INNER JOIN attributes a ON ua.id_attribute = a.id_user_attribute
		WHERE ua.id_user = $1 AND a.name = 'role'
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare user roles statement: %w", err)
	}

	// Prepare policies query
	_, err = r.db.PrepareContext(ctx, `
		SELECT policy_id, name, policy_rule, is_active, created_at, updated_at
		FROM policies
		WHERE is_active = true
		  AND policy_rule->>'resource' = $1
		  AND policy_rule->>'action' = $2
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare policies statement: %w", err)
	}

	// Prepare audit log insert
	_, err = r.db.PrepareContext(ctx, `
		INSERT INTO audit_logs (id_user, resource, action, allowed, reason, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare audit statement: %w", err)
	}

	return nil
}
