package model

import (
	"encoding/json"
	"time"
)

// GraphQLAPI represents a GraphQL API documentation
type GraphQLAPI struct {
	IDGraphqlAPI int64              `json:"id_graphql_api"`
	IDProject    int64              `json:"id_project"`
	IDUser       int64              `json:"id_user"`
	Name         string             `json:"name"`
	Type         string             `json:"type"` // query, mutation, subscription
	Description  string             `json:"description"`
	Arguments    []byte             `json:"arguments"`   // JSONB: Array of argument objects
	ReturnType   string             `json:"return_type"`
	Examples     []byte             `json:"examples"`    // JSONB: Array of example queries
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
}

// MarshalJSON implements custom JSON marshaling for GraphQLAPI
func (g *GraphQLAPI) MarshalJSON() ([]byte, error) {
	type Alias GraphQLAPI
	aux := &struct {
		Arguments json.RawMessage `json:"arguments"`
		Examples  json.RawMessage `json:"examples"`
		*Alias
	}{
		Alias: (*Alias)(g),
	}

	if len(g.Arguments) > 0 {
		aux.Arguments = json.RawMessage(g.Arguments)
	}
	if len(g.Examples) > 0 {
		aux.Examples = json.RawMessage(g.Examples)
	}

	if aux.Arguments == nil {
		aux.Arguments = json.RawMessage("[]")
	}
	if aux.Examples == nil {
		aux.Examples = json.RawMessage("[]")
	}

	return json.Marshal(aux)
}

// CreateGraphQLAPIRequest represents a request to create GraphQL API documentation
type CreateGraphQLAPIRequest struct {
	IDProject   int64              `json:"id_project" validate:"required"`
	Name        string             `json:"name" validate:"required,min=3,max=255"`
	Type        string             `json:"type" validate:"required,oneof=query mutation subscription"`
	Description string             `json:"description" validate:"max=1000"`
	Arguments   []GraphQLArgument  `json:"arguments"`
	ReturnType  string             `json:"return_type" validate:"required,max=500"`
	Examples    []GraphQLExample   `json:"examples"`
}

// UpdateGraphQLAPIRequest represents a request to update GraphQL API documentation
type UpdateGraphQLAPIRequest struct {
	Name        *string            `json:"name" validate:"omitempty,min=3,max=255"`
	Type        *string            `json:"type" validate:"omitempty,oneof=query mutation subscription"`
	Description *string            `json:"description" validate:"omitempty,max=1000"`
	Arguments   *[]GraphQLArgument `json:"arguments"`
	ReturnType  *string            `json:"return_type" validate:"omitempty,max=500"`
	Examples    *[]GraphQLExample  `json:"examples"`
}

// GraphQLArgument represents a GraphQL argument
type GraphQLArgument struct {
	Name         string      `json:"name" validate:"required"`
	Type         string      `json:"type" validate:"required"` // e.g., "ID!", "String", "[String!]"
	Description  string      `json:"description"`
	Required     bool        `json:"required"`
	DefaultValue any `json:"default_value"`
}

// GraphQLExample represents a GraphQL query example
type GraphQLExample struct {
	Name        string      `json:"name" validate:"required"`
	Description string      `json:"description"`
	Query       string      `json:"query" validate:"required"`
	Variables   any `json:"variables"`
	Response    any `json:"response"`
}

// ListGraphQLAPIsRequest represents a request to list GraphQL APIs with filters
type ListGraphQLAPIsRequest struct {
	Page      int    `json:"page" validate:"min=1"`
	Limit     int    `json:"limit" validate:"min=1,max=100"`
	IDProject *int64 `json:"id_project"` // Filter by project
	Type      string `json:"type" validate:"omitempty,oneof=query mutation subscription"`
	Search    string `json:"search"`
}
