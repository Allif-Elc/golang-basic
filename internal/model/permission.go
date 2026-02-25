package model

import "time"

type Attributes struct {
	AttributeID int64      `json:"id_attribute"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Type        string     `json:"type"`        // string, number, boolean, enum
	EnumValues  []string   `json:"enum_values"` // for enum type
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type Resources struct {
	ResourceID   int64     `json:"id_resource"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	ResourceType string    `json:"resource_type"` // general, api, document, etc.
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Permissions struct {
	PermissionID int64      `json:"id_permission"`
	Name         string     `json:"name"`
	Description  string     `json:"description"`
	Effect       string     `json:"effect"`       // allow, deny
	Actions      []string   `json:"actions"`      // array of actions
	Condition    *string    `json:"condition"`    // conditional logic (nullable)
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type CreateAttributesRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Type        string   `json:"type"`        // string, number, boolean, enum
	EnumValues  []string `json:"enum_values"` // for enum type
}

type CreateResoucesRequest struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	ResourceType string `json:"resource_type"` // general, api, document, etc.
}

type CreatePermissionsRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Effect      string   `json:"effect"`       // allow, deny
	Actions     []string `json:"actions"`      // array of actions
	Condition   *string  `json:"condition"`    // conditional logic (nullable)
}

type UpdateAttributesRequest struct {
	Name        *string  `json:"name"`
	Description *string  `json:"description"`
	Type        *string  `json:"type"`
	EnumValues  []string `json:"enum_values"`
}

type UpdateResourcesRequest struct {
	Name         *string `json:"name"`
	Description  *string `json:"description"`
	ResourceType *string `json:"resource_type"`
}

type UpdatePermissionsRequest struct {
	Name        *string  `json:"name"`
	Description *string  `json:"description"`
	Effect      *string `json:"effect"`
	Actions     []string `json:"actions"`
	Condition   *string `json:"condition"`
}
