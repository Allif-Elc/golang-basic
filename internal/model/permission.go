package model

import "time"

type Attributes struct {
	AttributeID int64     `json:"id_attribute"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreateAt    time.Time `json:"created_at"`
}

type Resources struct {
	ResourceID  int64     `json:"id_resource"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreateAt    time.Time `json:"created_at"`
}

type Permissions struct {
	PermissionID int64     `json:"id_permission"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	CreateAt     time.Time `json:"created_at"`
}

type CreateAttributesRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CreateResoucesRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CreatePermissionsRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
