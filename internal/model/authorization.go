package model

import (
	"encoding/json"
	"time"
)

// Policy represents an access control policy
type Policy struct {
	PolicyId   int64     `json:"id_policy"`
	Name       string    `json:"name"`
	PolicyRule []byte    `json:"policy_rule"` // JSONB stored as []byte
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// MarshalJSON implements custom JSON marshaling for Policy
// Handles PolicyRule []byte as JSON instead of base64
func (p *Policy) MarshalJSON() ([]byte, error) {
	type Alias Policy
	temp := struct {
		Alias
		PolicyRule interface{} `json:"policy_rule"`
	}{
		Alias:      (Alias)(*p),
		PolicyRule: json.RawMessage(p.PolicyRule),
	}
	return json.Marshal(temp)
}

// UnmarshalJSON implements custom JSON unmarshaling for Policy
// Handles PolicyRule []byte as JSON instead of base64
func (p *Policy) UnmarshalJSON(data []byte) error {
	type Alias Policy
	temp := struct {
		Alias
		PolicyRule json.RawMessage `json:"policy_rule"`
	}{}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	*p = Policy(temp.Alias)
	p.PolicyRule = temp.PolicyRule
	return nil
}

// PolicyWithPriority represents a policy with its evaluation priority
// Used to merge user policies (higher priority) with role-based policies (priority 0)
type PolicyWithPriority struct {
	Policy   Policy
	Priority int
}

// UserAttribute represents a user's attribute value (role, department, etc.)
type UserAttribute struct {
	UserAttributeId int64     `json:"id_user_attribute"`
	UserID          int64     `json:"id_user"`
	AttributeID     int64     `json:"id_attribute"`
	Value           string    `json:"value"` // HR, Manager, Engineering, etc.
	CreatedAt       time.Time `json:"created_at"`
}

// AuditLog represents authorization decision logging
type AuditLog struct {
	ID        int64     `json:"id_audit_log"`
	UserID    int64     `json:"id_user"`
	Resource  string    `json:"resource"`
	Action    string    `json:"action"`
	Allowed   bool      `json:"allowed"`
	Reason    string    `json:"reason"`
	CreatedAt time.Time `json:"created_at"`
}

// PolicyRule represents the parsed JSONB policy structure
type PolicyRule struct {
	Role     string   `json:"role"`
	Resource string   `json:"resource"`
	Action   []string `json:"action"` // Action array: ["read", "write"] or ["*"]
}

// AuthorizeRequest represents an authorization check request
type AuthorizeRequest struct {
	UserID   int64  `json:"id_user"`
	Resource string `json:"resource"` // employee_records, budget_report, etc.
	Action   string `json:"action"`   // read, write, delete, etc.
}

// AuthorizeResponse represents the authorization decision
type AuthorizeResponse struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason"`
}

// UserPolicy represents a direct user-to-policy assignment with priority
type UserPolicy struct {
	UserPolicyId int64       `json:"id_user_policy"`
	UserID       int64       `json:"id_user"`
	PolicyID     int64       `json:"id_policy"`
	Priority     int         `json:"priority"`
	IsActive     bool        `json:"is_active"`
	ExpiresAt    *time.Time  `json:"expires_at,omitempty"`
	CreatedAt    time.Time   `json:"created_at"`
	CreatedBy    *int64      `json:"created_by,omitempty"`
}

// CreateUserPolicyRequest creates a new user-policy assignment
type CreateUserPolicyRequest struct {
	UserID    int64       `json:"id_user" validate:"required"`
	PolicyID  int64       `json:"id_policy" validate:"required"`
	Priority  int         `json:"priority" validate:"min=-100,max=100"`
	ExpiresAt *time.Time  `json:"expires_at,omitempty"`
}

// UpdateUserPolicyRequest updates an existing user-policy assignment
type UpdateUserPolicyRequest struct {
	Priority  *int        `json:"priority,omitempty" validate:"omitempty,min=-100,max=100"`
	IsActive  *bool       `json:"is_active,omitempty"`
	ExpiresAt *time.Time  `json:"expires_at,omitempty"`
}

// UserPolicyDetail includes related policy and user information
type UserPolicyDetail struct {
	UserPolicyId int64       `json:"id_user_policy"`
	UserID       int64       `json:"id_user"`
	UserName     string      `json:"user_name"`
	UserEmail    string      `json:"user_email"`
	PolicyID     int64       `json:"id_policy"`
	PolicyName   string      `json:"policy_name"`
	PolicyRule   []byte      `json:"policy_rule"`
	Priority     int         `json:"priority"`
	IsActive     bool        `json:"is_active"`
	ExpiresAt    *time.Time  `json:"expires_at,omitempty"`
	CreatedAt    time.Time   `json:"created_at"`
}

// MarshalJSON implements custom JSON marshaling for UserPolicyDetail
// Handles PolicyRule []byte as JSON instead of base64
func (u *UserPolicyDetail) MarshalJSON() ([]byte, error) {
	type Alias UserPolicyDetail
	temp := struct {
		Alias
		PolicyRule interface{} `json:"policy_rule"`
	}{
		Alias:      (Alias)(*u),
		PolicyRule: json.RawMessage(u.PolicyRule),
	}
	return json.Marshal(temp)
}

// UnmarshalJSON implements custom JSON unmarshaling for UserPolicyDetail
// Handles PolicyRule []byte as JSON instead of base64
func (u *UserPolicyDetail) UnmarshalJSON(data []byte) error {
	type Alias UserPolicyDetail
	temp := struct {
		Alias
		PolicyRule json.RawMessage `json:"policy_rule"`
	}{}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	*u = UserPolicyDetail(temp.Alias)
	u.PolicyRule = temp.PolicyRule
	return nil
}

// CreatePolicyRequest creates a new access control policy
type CreatePolicyRequest struct {
	Name       string          `json:"name" validate:"required,min=1,max=255"`
	PolicyRule json.RawMessage `json:"policy_rule" validate:"required"`
	IsActive   bool            `json:"is_active"`
}

// UpdatePolicyRequest updates an existing policy
type UpdatePolicyRequest struct {
	Name       *string         `json:"name,omitempty"`
	PolicyRule *json.RawMessage `json:"policy_rule,omitempty"`
	IsActive   *bool           `json:"is_active,omitempty"`
}
