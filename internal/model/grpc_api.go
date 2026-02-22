package model

import (
	"encoding/json"
	"time"
)

// GrpcAPI represents a gRPC API documentation
type GrpcAPI struct {
	IDGrpcAPI        int64              `json:"id_grpc_api"`
	IDProject        int64              `json:"id_project"`
	IDUser           int64              `json:"id_user"`
	ServiceName      string             `json:"service_name"`
	MethodName       string             `json:"method_name"`
	Description      string             `json:"description"`
	RequestMessage   []byte             `json:"request_message"`  // JSONB: Array of field objects
	ResponseMessage  []byte             `json:"response_message"` // JSONB: Array of field objects
	ProtoDefinition  string             `json:"proto_definition"` // TEXT: Actual .proto content
	Examples         []byte             `json:"examples"`         // JSONB: Array of code examples
	CreatedAt        time.Time          `json:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at"`
}

// MarshalJSON implements custom JSON marshaling for GrpcAPI
func (g *GrpcAPI) MarshalJSON() ([]byte, error) {
	type Alias GrpcAPI
	aux := &struct {
		RequestMessage  json.RawMessage `json:"request_message"`
		ResponseMessage json.RawMessage `json:"response_message"`
		Examples        json.RawMessage `json:"examples"`
		*Alias
	}{
		Alias: (*Alias)(g),
	}

	if len(g.RequestMessage) > 0 {
		aux.RequestMessage = json.RawMessage(g.RequestMessage)
	}
	if len(g.ResponseMessage) > 0 {
		aux.ResponseMessage = json.RawMessage(g.ResponseMessage)
	}
	if len(g.Examples) > 0 {
		aux.Examples = json.RawMessage(g.Examples)
	}

	if aux.RequestMessage == nil {
		aux.RequestMessage = json.RawMessage("[]")
	}
	if aux.ResponseMessage == nil {
		aux.ResponseMessage = json.RawMessage("[]")
	}
	if aux.Examples == nil {
		aux.Examples = json.RawMessage("[]")
	}

	return json.Marshal(aux)
}

// CreateGrpcAPIRequest represents a request to create gRPC API documentation
type CreateGrpcAPIRequest struct {
	IDProject       int64            `json:"id_project" validate:"required"`
	ServiceName     string           `json:"service_name" validate:"required,max=255"`
	MethodName      string           `json:"method_name" validate:"required,max=255"`
	Description     string           `json:"description" validate:"max=1000"`
	RequestMessage  []GrpcField      `json:"request_message"`
	ResponseMessage []GrpcField      `json:"response_message"`
	ProtoDefinition string           `json:"proto_definition" validate:"max=10000"` // TEXT field
	Examples        []GrpcExample     `json:"examples"`
}

// UpdateGrpcAPIRequest represents a request to update gRPC API documentation
type UpdateGrpcAPIRequest struct {
	ServiceName     *string       `json:"service_name" validate:"omitempty,max=255"`
	MethodName      *string       `json:"method_name" validate:"omitempty,max=255"`
	Description     *string       `json:"description" validate:"omitempty,max=1000"`
	RequestMessage  *[]GrpcField  `json:"request_message"`
	ResponseMessage *[]GrpcField  `json:"response_message"`
	ProtoDefinition *string       `json:"proto_definition" validate:"omitempty,max=10000"`
	Examples        *[]GrpcExample `json:"examples"`
}

// GrpcField represents a gRPC message field
type GrpcField struct {
	Name        string `json:"name" validate:"required"`
	Type        string `json:"type" validate:"required"` // e.g., "string", "int32", "bool", "message"
	Label       string `json:"label" validate:"omitempty,oneof=optional repeated required"`
	Description string `json:"description"`
}

// GrpcExample represents a code example for gRPC
type GrpcExample struct {
	Language    string      `json:"language" validate:"required"` // e.g., "go", "python", "java"
	Description string      `json:"description"`
	Code        string      `json:"code" validate:"required"`
	Variables   any `json:"variables"`
}

// ListGrpcAPIsRequest represents a request to list gRPC APIs with filters
type ListGrpcAPIsRequest struct {
	Page        int    `json:"page" validate:"min=1"`
	Limit       int    `json:"limit" validate:"min=1,max=100"`
	IDProject   *int64 `json:"id_project"` // Filter by project
	ServiceName string `json:"service_name"` // Filter by service name
	Search      string `json:"search"`
}
