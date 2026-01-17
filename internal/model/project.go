package model

import "time"

// Project represents a documentation project
type Project struct {
	IDProject   int64     `json:"id_project"`
	IDUser      int64     `json:"id_user"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	Version     string    `json:"version"`
	IsPublic    bool      `json:"is_public"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateProjectRequest represents a request to create a project
type CreateProjectRequest struct {
	Name        string `json:"name" validate:"required,min=3,max=255"`
	Description string `json:"description" validate:"max=1000"`
	Version     string `json:"version" validate:"max=50"`
	IsPublic    bool   `json:"is_public"`
}

// UpdateProjectRequest represents a request to update a project
type UpdateProjectRequest struct {
	Name        *string `json:"name" validate:"omitempty,min=3,max=255"`
	Description *string `json:"description" validate:"omitempty,max=1000"`
	Version     *string `json:"version" validate:"omitempty,max=50"`
	IsPublic    *bool   `json:"is_public"`
}

// ListProjectsRequest represents a request to list projects with filters
type ListProjectsRequest struct {
	Page      int    `json:"page" validate:"min=1"`
	Limit     int    `json:"limit" validate:"min=1,max=100"`
	Search    string `json:"search"`
	IsPublic  *bool  `json:"is_public"`
	IDUser    *int64 `json:"id_user"` // Filter by owner
	SortBy    string `json:"sort_by" validate:"omitempty,oneof=name created_at updated_at"`
	SortOrder string `json:"sort_order" validate:"omitempty,oneof=asc desc"`
}
