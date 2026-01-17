package model

import "time"

// RestAPI represents a REST API documentation
type RestAPI struct {
	IDRestAPI   int64              `json:"id_rest_api"`
	IDProject   int64              `json:"id_project"`
	IDUser      int64              `json:"id_user"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Method      string             `json:"method"` // GET, POST, PUT, DELETE, PATCH
	Endpoint    string             `json:"endpoint"`
	Headers     []byte             `json:"headers"`     // JSONB: Array of header objects
	PathParams  []byte             `json:"path_params"`  // JSONB: Array of parameter objects
	QueryParams []byte             `json:"query_params"` // JSONB: Array of parameter objects
	RequestBody []byte             `json:"request_body"` // JSONB: JSON schema
	Responses   []byte             `json:"responses"`    // JSONB: Response examples by status
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

// CreateRestAPIRequest represents a request to create REST API documentation
type CreateRestAPIRequest struct {
	IDProject   int64              `json:"id_project" validate:"required"`
	Name        string             `json:"name" validate:"required,min=3,max=255"`
	Description string             `json:"description" validate:"max=1000"`
	Method      string             `json:"method" validate:"required,oneof=GET POST PUT DELETE PATCH"`
	Endpoint    string             `json:"endpoint" validate:"required,max=500"`
	Headers     []Header           `json:"headers"`
	PathParams  []Parameter        `json:"path_params"`
	QueryParams []Parameter        `json:"query_params"`
	RequestBody *JSONSchema        `json:"request_body"`
	Responses   map[int]ResponseExample `json:"responses"`
}

// UpdateRestAPIRequest represents a request to update REST API documentation
type UpdateRestAPIRequest struct {
	Name        *string            `json:"name" validate:"omitempty,min=3,max=255"`
	Description *string            `json:"description" validate:"omitempty,max=1000"`
	Method      *string            `json:"method" validate:"omitempty,oneof=GET POST PUT DELETE PATCH"`
	Endpoint    *string            `json:"endpoint" validate:"omitempty,max=500"`
	Headers     *[]Header          `json:"headers"`
	PathParams  *[]Parameter       `json:"path_params"`
	QueryParams *[]Parameter       `json:"query_params"`
	RequestBody *JSONSchema        `json:"request_body"`
	Responses   *map[int]ResponseExample `json:"responses"`
}

// Header represents an HTTP header
type Header struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
	Example     string `json:"example"`
}

// Parameter represents a path or query parameter
type Parameter struct {
	Name        string      `json:"name" validate:"required"`
	Type        string      `json:"type" validate:"required,oneof=string integer boolean number"`
	Description string      `json:"description"`
	Required    bool        `json:"required"`
	Default     interface{} `json:"default"`
}

// JSONSchema represents a JSON schema for request body
type JSONSchema struct {
	Type       string                 `json:"type" validate:"required,oneof=object array string number boolean null"`
	Properties map[string]Property    `json:"properties"`
	Required   []string               `json:"required"`
	Items      *JSONSchema            `json:"items"` // For array types
}

// Property represents a JSON schema property
type Property struct {
	Type        string       `json:"type" validate:"required,oneof=string integer boolean number array object"`
	Description string       `json:"description"`
	Format      string       `json:"format"` // e.g., "email", "date-time"
	Minimum     *float64     `json:"minimum"`
	Maximum     *float64     `json:"maximum"`
	Enum        []string     `json:"enum"`
	Required    []string     `json:"required"`
	Properties  map[string]Property `json:"properties"` // For object types
	Items       *JSONSchema  `json:"items"` // For array types
}

// ResponseExample represents a response example
type ResponseExample struct {
	StatusCode  int         `json:"status_code"`
	Description string      `json:"description"`
	Body        interface{} `json:"body"`
}

// ListRestAPIsRequest represents a request to list REST APIs with filters
type ListRestAPIsRequest struct {
	Page      int    `json:"page" validate:"min=1"`
	Limit     int    `json:"limit" validate:"min=1,max=100"`
	IDProject *int64 `json:"id_project"` // Filter by project
	Method    string `json:"method" validate:"omitempty,oneof=GET POST PUT DELETE PATCH"`
	Search    string `json:"search"`
}
