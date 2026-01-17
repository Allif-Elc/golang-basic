package model

import "time"

// Tag represents a categorization tag
type Tag struct {
	IDTag      int64     `json:"id_tag"`
	Name       string    `json:"name"`
	Color      string    `json:"color"` // Hex color code, e.g., "#FF5733"
	CreatedAt  time.Time `json:"created_at"`
}

// CreateTagRequest represents a request to create a tag
type CreateTagRequest struct {
	Name  string `json:"name" validate:"required,min=1,max=100"`
	Color string `json:"color" validate:"required,hexcolor|min=7|max=7"`
}

// UpdateTagRequest represents a request to update a tag
type UpdateTagRequest struct {
	Name  *string `json:"name" validate:"omitempty,min=1,max=100"`
	Color *string `json:"color" validate:"omitempty,hexcolor|min=7|max=7"`
}

// ApiTag represents a relationship between an API and a tag
type ApiTag struct {
	IDApi      int64  `json:"id_api"`
	ApiType    string `json:"api_type"` // 'rest', 'graphql', 'grpc'
	IDTag      int64  `json:"id_tag"`
}

// AddTagsToAPIRequest represents a request to add tags to an API
type AddTagsToAPIRequest struct {
	ApiType string `json:"api_type" validate:"required,oneof=rest graphql grpc"`
	IDTags  []int64 `json:"id_tags" validate:"required,min=1"`
}
