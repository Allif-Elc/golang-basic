package service

import (
	"context"
	"fmt"
	"golang-basic/api/internal/model"
	"golang-basic/api/internal/repository"
	"golang-basic/api/internal/utility"
	"regexp"
	"strings"
	"unicode/utf8"
)

type GrpcAPIService struct {
	repo        *repository.GrpcAPIRepository
	projectRepo *repository.ProjectRepository
}

func NewGrpcAPIService(repo *repository.GrpcAPIRepository, projectRepo *repository.ProjectRepository) *GrpcAPIService {
	return &GrpcAPIService{
		repo:        repo,
		projectRepo: projectRepo,
	}
}

// CreateGrpcAPI creates a new gRPC API documentation
// Per Specs: Authorization handled by ABAC middleware at controller level
func (s *GrpcAPIService) CreateGrpcAPI(ctx context.Context, userID int64, req model.CreateGrpcAPIRequest) (*model.GrpcAPI, error) {
	// Verify project exists and user has access
	_, err := s.projectRepo.FindByID(ctx, req.IDProject)
	if err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	// Sanitize and validate all inputs
	sanitizedServiceName := sanitizeInput(req.ServiceName)
	sanitizedMethodName := sanitizeInput(req.MethodName)
	sanitizedDescription := sanitizeInput(req.Description)
	sanitizedProtoDefinition := req.ProtoDefinition // Don't sanitize proto definition as it's code

	// Validate service name
	if err := ValidateGrpcServiceName(sanitizedServiceName); err != nil {
		return nil, err
	}

	// Validate method name
	if err := ValidateGrpcMethodName(sanitizedMethodName); err != nil {
		return nil, err
	}

	// Validate description
	if err := validateGrpcAPIDescription(sanitizedDescription); err != nil {
		return nil, err
	}

	// Validate proto definition
	if err := validateProtoDefinition(sanitizedProtoDefinition); err != nil {
		return nil, err
	}

	// Validate request message fields
	if len(req.RequestMessage) > 0 {
		if err := validateGrpcFields(req.RequestMessage); err != nil {
			return nil, fmt.Errorf("request message: %w", err)
		}
	}

	// Validate response message fields
	if len(req.ResponseMessage) > 0 {
		if err := validateGrpcFields(req.ResponseMessage); err != nil {
			return nil, fmt.Errorf("response message: %w", err)
		}
	}

	// Validate examples
	if len(req.Examples) > 0 {
		if err := validateGrpcExamples(req.Examples); err != nil {
			return nil, err
		}
	}

	// Update request with sanitized values
	req.ServiceName = sanitizedServiceName
	req.MethodName = sanitizedMethodName
	req.Description = sanitizedDescription
	req.ProtoDefinition = sanitizedProtoDefinition

	// Create gRPC API
	grpcAPI, err := s.repo.Create(ctx, &req, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC API: %w", err)
	}

	return grpcAPI, nil
}

// GetGrpcAPIByID retrieves a gRPC API by its ID
func (s *GrpcAPIService) GetGrpcAPIByID(ctx context.Context, apiID int64) (model.GrpcAPI, error) {
	grpcAPI, err := s.repo.FindByID(ctx, apiID)
	if err != nil {
		return model.GrpcAPI{}, fmt.Errorf("failed to get gRPC API: %w", err)
	}

	return grpcAPI, nil
}

// UpdateGrpcAPI modifies an existing gRPC API documentation
func (s *GrpcAPIService) UpdateGrpcAPI(ctx context.Context, apiID int64, req model.UpdateGrpcAPIRequest) (*model.GrpcAPI, error) {
	// Check if gRPC API exists
	_, err := s.repo.FindByID(ctx, apiID)
	if err != nil {
		return nil, fmt.Errorf("gRPC API not found: %w", err)
	}

	// Sanitize and validate provided fields
	if req.ServiceName != nil {
		sanitizedServiceName := sanitizeInput(*req.ServiceName)
		if err := ValidateGrpcServiceName(sanitizedServiceName); err != nil {
			return nil, err
		}
		req.ServiceName = &sanitizedServiceName
	}

	if req.MethodName != nil {
		sanitizedMethodName := sanitizeInput(*req.MethodName)
		if err := ValidateGrpcMethodName(sanitizedMethodName); err != nil {
			return nil, err
		}
		req.MethodName = &sanitizedMethodName
	}

	if req.Description != nil {
		sanitizedDescription := sanitizeInput(*req.Description)
		if err := validateGrpcAPIDescription(sanitizedDescription); err != nil {
			return nil, err
		}
		req.Description = &sanitizedDescription
	}

	if req.ProtoDefinition != nil {
		if err := validateProtoDefinition(*req.ProtoDefinition); err != nil {
			return nil, err
		}
	}

	if req.RequestMessage != nil {
		if err := validateGrpcFields(*req.RequestMessage); err != nil {
			return nil, fmt.Errorf("request message: %w", err)
		}
	}

	if req.ResponseMessage != nil {
		if err := validateGrpcFields(*req.ResponseMessage); err != nil {
			return nil, fmt.Errorf("response message: %w", err)
		}
	}

	if req.Examples != nil {
		if err := validateGrpcExamples(*req.Examples); err != nil {
			return nil, err
		}
	}

	// Update gRPC API
	grpcAPI, err := s.repo.Update(ctx, apiID, &req)
	if err != nil {
		return nil, fmt.Errorf("failed to update gRPC API: %w", err)
	}

	return grpcAPI, nil
}

// DeleteGrpcAPI removes a gRPC API documentation
func (s *GrpcAPIService) DeleteGrpcAPI(ctx context.Context, apiID int64) error {
	rowsAffected, err := s.repo.Delete(ctx, apiID)
	if err != nil {
		return fmt.Errorf("failed to delete gRPC API: %w", err)
	}

	if rowsAffected == 0 {
		return utility.NotFoundError("gRPC API not found")
	}

	return nil
}

// ListGrpcAPIs retrieves gRPC APIs with pagination and filters
func (s *GrpcAPIService) ListGrpcAPIs(ctx context.Context, req model.ListGrpcAPIsRequest) (*model.PageResult[model.GrpcAPI], error) {
	// Sanitize search input to prevent SQL injection
	if req.Search != "" {
		req.Search = sanitizeInput(req.Search)
	}

	// Sanitize service name if provided
	if req.ServiceName != "" {
		req.ServiceName = sanitizeInput(req.ServiceName)
	}

	// Set default values
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 || req.Limit > 100 {
		req.Limit = 20
	}

	// Get gRPC APIs
	result, err := s.repo.List(ctx, &req)
	if err != nil {
		return nil, fmt.Errorf("failed to list gRPC APIs: %w", err)
	}

	return result, nil
}

// ==================== GRPC API SPECIFIC VALIDATION FUNCTIONS ====================

// ValidateGrpcServiceName performs comprehensive validation on gRPC service name
func ValidateGrpcServiceName(serviceName string) error {
	// Check length
	if utf8.RuneCountInString(serviceName) < 1 {
		return fmt.Errorf("service name is required")
	}
	if utf8.RuneCountInString(serviceName) > 255 {
		return fmt.Errorf("service name must not exceed 255 characters")
	}

	// Check for null bytes
	if strings.Contains(serviceName, "\x00") {
		return fmt.Errorf("service name contains null bytes")
	}

	// Check for XSS patterns
	if containsXSSPatterns(serviceName) {
		return fmt.Errorf("service name contains potentially dangerous content")
	}

	// gRPC service names should follow PascalCase convention
	// Must start with an uppercase letter, followed by letters, digits, or underscores
	matched, _ := regexp.MatchString(`^[A-Z][a-zA-Z0-9_]*$`, serviceName)
	if !matched {
		return fmt.Errorf("service name contains invalid characters (must be PascalCase)")
	}

	return nil
}

// ValidateGrpcMethodName performs comprehensive validation on gRPC method name
func ValidateGrpcMethodName(methodName string) error {
	// Check length
	if utf8.RuneCountInString(methodName) < 1 {
		return fmt.Errorf("method name is required")
	}
	if utf8.RuneCountInString(methodName) > 255 {
		return fmt.Errorf("method name must not exceed 255 characters")
	}

	// Check for null bytes
	if strings.Contains(methodName, "\x00") {
		return fmt.Errorf("method name contains null bytes")
	}

	// Check for XSS patterns
	if containsXSSPatterns(methodName) {
		return fmt.Errorf("method name contains potentially dangerous content")
	}

	// gRPC method names should follow PascalCase convention
	matched, _ := regexp.MatchString(`^[A-Z][a-zA-Z0-9_]*$`, methodName)
	if !matched {
		return fmt.Errorf("method name contains invalid characters (must be PascalCase)")
	}

	return nil
}

// validateGrpcAPIDescription performs validation on gRPC API description
func validateGrpcAPIDescription(description string) error {
	// Check length
	if utf8.RuneCountInString(description) > 1000 {
		return fmt.Errorf("description must not exceed 1000 characters")
	}

	// Allow empty description
	if description == "" {
		return nil
	}

	// Check for null bytes
	if strings.Contains(description, "\x00") {
		return fmt.Errorf("description contains null bytes")
	}

	// Check for XSS patterns
	if containsXSSPatterns(description) {
		return fmt.Errorf("description contains potentially dangerous content")
	}

	return nil
}

// validateProtoDefinition validates Protocol Buffers definition
func validateProtoDefinition(protoDefinition string) error {
	// Check length
	if len(protoDefinition) > 10000 {
		return fmt.Errorf("proto definition must not exceed 10000 characters")
	}

	// Allow empty proto definition (it's optional)
	if protoDefinition == "" {
		return nil
	}

	// Check for null bytes
	if strings.Contains(protoDefinition, "\x00") {
		return fmt.Errorf("proto definition contains null bytes")
	}

	// Basic validation: proto definition should contain valid proto syntax
	// We don't do full syntax validation here, just basic sanity checks
	// A more sophisticated validator would parse the proto file

	return nil
}

// validateGrpcFields validates gRPC message field definitions
func validateGrpcFields(fields []model.GrpcField) error {
	// Check number of fields
	if len(fields) > 100 {
		return fmt.Errorf("too many fields (maximum 100)")
	}

	fieldNames := make(map[string]bool)

	for i, field := range fields {
		// Validate field name
		if err := validateGrpcFieldName(field.Name); err != nil {
			return fmt.Errorf("field %d: %w", i+1, err)
		}

		// Check for duplicate field names
		if fieldNames[field.Name] {
			return fmt.Errorf("duplicate field name: %s", field.Name)
		}
		fieldNames[field.Name] = true

		// Validate field type
		if err := validateGrpcFieldType(field.Type); err != nil {
			return fmt.Errorf("field %s: %w", field.Name, err)
		}

		// Validate description if provided
		if field.Description != "" {
			if utf8.RuneCountInString(field.Description) > 500 {
				return fmt.Errorf("field %s: description too long (maximum 500 characters)", field.Name)
			}
			if containsXSSPatterns(field.Description) {
				return fmt.Errorf("field %s: description contains potentially dangerous content", field.Name)
			}
		}

		// Validate label if provided
		if field.Label != "" {
			validLabels := map[string]bool{
				"optional": true,
				"repeated": true,
				"required": true,
			}
			if !validLabels[field.Label] {
				return fmt.Errorf("field %s: invalid label '%s' (must be optional, repeated, or required)", field.Name, field.Label)
			}
		}
	}

	return nil
}

// validateGrpcFieldName validates a single gRPC field name
func validateGrpcFieldName(name string) error {
	if name == "" {
		return fmt.Errorf("field name is required")
	}

	if len(name) > 100 {
		return fmt.Errorf("field name too long (maximum 100 characters)")
	}

	// gRPC field names should follow snake_case convention
	matched, _ := regexp.MatchString(`^[a-z][a-z0-9_]*$`, name)
	if !matched {
		return fmt.Errorf("field name contains invalid characters (must be snake_case)")
	}

	return nil
}

// validateGrpcFieldType validates gRPC field type
func validateGrpcFieldType(fieldType string) error {
	if fieldType == "" {
		return fmt.Errorf("field type is required")
	}

	if len(fieldType) > 100 {
		return fmt.Errorf("field type too long (maximum 100 characters)")
	}

	// Common gRPC/Protobuf types
	scalarTypes := map[string]bool{
		"double":   true,
		"float":    true,
		"int32":    true,
		"int64":    true,
		"uint32":   true,
		"uint64":   true,
		"sint32":   true,
		"sint64":   true,
		"fixed32":  true,
		"fixed64":  true,
		"sfixed32": true,
		"sfixed64": true,
		"bool":     true,
		"string":   true,
		"bytes":    true,
	}

	// Allow scalar types or custom message types
	if !scalarTypes[fieldType] {
		// Custom message type validation (PascalCase)
		matched, _ := regexp.MatchString(`^[A-Z][a-zA-Z0-9_]*$`, fieldType)
		if !matched {
			return fmt.Errorf("field type contains invalid characters")
		}
	}

	return nil
}

// validateGrpcExamples validates gRPC code examples
func validateGrpcExamples(examples []model.GrpcExample) error {
	// Check number of examples
	if len(examples) > 20 {
		return fmt.Errorf("too many examples (maximum 20)")
	}

	languages := make(map[string]bool)

	for i, example := range examples {
		// Validate language
		if example.Language == "" {
			return fmt.Errorf("example %d: language is required", i+1)
		}
		if len(example.Language) > 50 {
			return fmt.Errorf("example %d: language too long (maximum 50 characters)", i+1)
		}

		// Track language combinations
		key := fmt.Sprintf("%s:%d", example.Language, i)
		if languages[key] {
			return fmt.Errorf("duplicate example for language: %s", example.Language)
		}
		languages[key] = true

		// Validate code
		if example.Code == "" {
			return fmt.Errorf("example %d: code is required", i+1)
		}
		if len(example.Code) > 10000 {
			return fmt.Errorf("example %d: code too long (maximum 10000 characters)", i+1)
		}

		// Validate description if provided
		if example.Description != "" {
			if utf8.RuneCountInString(example.Description) > 500 {
				return fmt.Errorf("example %d: description too long", i+1)
			}
			if containsXSSPatterns(example.Description) {
				return fmt.Errorf("example %d: description contains potentially dangerous content", i+1)
			}
		}

		// Validate variables if provided (should be valid JSON)
		if example.Variables != nil {
			// Basic validation - just check it's not a string (to avoid injection)
			if _, ok := example.Variables.(string); ok {
				return fmt.Errorf("example %d: variables must be a JSON object, not a string", i+1)
			}
		}
	}

	return nil
}
