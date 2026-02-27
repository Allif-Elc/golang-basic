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
	// Deprecated: Use GetUserAttributes for full attribute-based authorization
	// Required indexes:
	//   CREATE INDEX idx_user_attributes_user ON user_attributes(id_user);
	//   CREATE INDEX idx_user_attributes_attribute ON user_attributes(id_attribute);
	//   CREATE INDEX idx_attributes_id ON attributes(id_user_attribute);
	//   CREATE INDEX idx_attributes_name ON attributes(name);
	query := `
		SELECT ua.value
		FROM user_attributes ua
		INNER JOIN attributes a ON ua.id_attribute = a.id_attribute
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

// GetUserAttributes retrieves all attributes for a user
// Returns detailed attribute information including name, type, and value
// Used for attribute-based authorization (new ABAC model)
// Required indexes:
//   CREATE INDEX idx_user_attributes_user ON user_attributes(id_user);
//   CREATE INDEX idx_user_attributes_attribute ON user_attributes(id_attribute);
//   CREATE INDEX idx_attributes_name ON attributes(name);
func (r *AuthorizationRepository) GetUserAttributes(ctx context.Context, userID int64) ([]model.UserAttributeDetail, error) {
	query := `
		SELECT
			ua.id_user_attribute,
			ua.id_user,
			u.name as user_name,
			u.email as user_email,
			ua.id_attribute,
			a.name as attribute_name,
			a.type as attribute_type,
			a.enum_values,
			ua.value,
			ua.created_at
		FROM user_attributes ua
		INNER JOIN attributes a ON ua.id_attribute = a.id_attribute
		INNER JOIN users u ON ua.id_user = u.id_user
		WHERE ua.id_user = $1
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user attributes: %w", err)
	}
	defer rows.Close()

	var attributes []model.UserAttributeDetail
	for rows.Next() {
		var attr model.UserAttributeDetail
		if err := rows.Scan(
			&attr.UserAttributeID,
			&attr.UserID,
			&attr.UserName,
			&attr.UserEmail,
			&attr.AttributeID,
			&attr.AttributeName,
			&attr.AttributeType,
			&attr.EnumValues,
			&attr.Value,
			&attr.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan user attribute: %w", err)
		}
		attributes = append(attributes, attr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user attributes: %w", err)
	}

	return attributes, nil
}

// GetMatchingPolicies fetches policies that match the resource and action
// Supports wildcard matching:
//   - Full wildcard "*" matches all resources
//   - Prefix wildcard "prefix_*" matches resources starting with "prefix_"
//   - Action wildcard ["*"] matches all actions
// Required indexes:
//
//	CREATE INDEX idx_policies_is_active ON policies(is_active);
//	CREATE INDEX idx_policies_rule_resource ON policies ((policy_rule->>'resource'));
//	CREATE INDEX idx_policies_rule_action ON policies USING GIN ((policy_rule->'action'));
func (r *AuthorizationRepository) GetMatchingPolicies(ctx context.Context, resource, action string) ([]model.Policy, error) {
	query := `
		SELECT id_policy, name, policy_rule, is_active, created_at, updated_at
		FROM policies
		WHERE is_active = true
		  AND (
		    -- Exact match (original behavior)
		    policy_rule->>'resource' = $1
		    -- Full wildcard "*" matches everything
		    OR policy_rule->>'resource' = '*'
		    -- Prefix wildcard: "prefix_*" matches any resource starting with "prefix_"
		    OR (
		      policy_rule->>'resource' LIKE '%_\*'
		      AND substr(policy_rule->>'resource', 1, length(policy_rule->>'resource') - 2) = substr($1, 1, length(policy_rule->>'resource') - 2)
		      AND substr($1, length(policy_rule->>'resource') - 1, 1) = '_'
		      AND length($1) > length(policy_rule->>'resource') - 1
		    )
		  )
		  AND (
		    -- Action array contains requested action (original behavior)
		    policy_rule->'action' @> to_jsonb(ARRAY[$2])
		    -- Or action array contains wildcard "*"
		    OR policy_rule->'action' @> to_jsonb(ARRAY['*'])
		  )
	`

	rows, err := r.db.Query(ctx, query, resource, action)
	if err != nil {
		return nil, fmt.Errorf("failed to query policies: %w", err)
	}
	defer rows.Close()

	policies := make([]model.Policy, 0)
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
//
//	CREATE INDEX idx_audit_logs_user ON audit_logs(user_id);
//	CREATE INDEX idx_audit_logs_resource ON audit_logs(resource);
//	CREATE INDEX idx_audit_logs_action ON audit_logs(action);
//	CREATE INDEX idx_audit_logs_created ON audit_logs(created_at);
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

// GetUserPoliciesByUserID fetches user-specific policies with priority
// Returns policies that are: active, not expired, and assigned to the user
// Used by authorization service to get higher-priority user policies before role-based policies
// Performance: <10ms with proper indexes
// Required indexes:
//
//	CREATE INDEX idx_user_policies_user_active_expires ON user_policies(id_user, is_active, expires_at)
//	    WHERE is_active = true;
//	CREATE INDEX idx_policies_active ON policies USING GIN (policy_rule)
//	    WHERE is_active = true;
func (r *AuthorizationRepository) GetUserPoliciesByUserID(ctx context.Context, userID int64) ([]model.PolicyWithPriority, error) {
	query := `
		SELECT p.id_policy, p.name, p.policy_rule, p.is_active, p.created_at, p.updated_at,
		       up.priority
		FROM user_policies up
		INNER JOIN policies p ON up.id_policy = p.id_policy
		WHERE up.id_user = $1
		  AND up.is_active = true
		  AND p.is_active = true
		  AND (up.expires_at IS NULL OR up.expires_at > CURRENT_TIMESTAMP)
		ORDER BY up.priority DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user policies: %w", err)
	}
	defer rows.Close()

	result := make([]model.PolicyWithPriority, 0)
	for rows.Next() {
		var p model.Policy
		var priority int
		if err := rows.Scan(
			&p.PolicyId,
			&p.Name,
			&p.PolicyRule,
			&p.IsActive,
			&p.CreatedAt,
			&p.UpdatedAt,
			&priority,
		); err != nil {
			return nil, fmt.Errorf("failed to scan user policy: %w", err)
		}
		result = append(result, model.PolicyWithPriority{Policy: p, Priority: priority})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user policies: %w", err)
	}

	return result, nil
}

// ListPolicies retrieves all policies from the database
// Required indexes:
//
//	CREATE INDEX idx_policies_active ON policies(is_active);
func (r *AuthorizationRepository) ListPolicies(ctx context.Context) ([]model.Policy, error) {
	query := `
		SELECT id_policy, name, policy_rule, is_active, created_at, updated_at
		FROM policies
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query policies: %w", err)
	}
	defer rows.Close()

	policies := make([]model.Policy, 0)
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
