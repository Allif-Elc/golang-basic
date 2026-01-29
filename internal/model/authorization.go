package model

import "time"

// Policy represents an access control policy
type Policy struct {
	PolicyId   int64     `json:"id_policy"`
	Name       string    `json:"name"`
	PolicyRule []byte    `json:"policy_rule"` // JSONB stored as []byte
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
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
