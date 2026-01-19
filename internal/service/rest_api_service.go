package service

import (
	"context"
	"encoding/json"
	"fmt"
	"golang-basic/api/internal/model"
	"golang-basic/api/internal/repository"
	"golang-basic/api/internal/utility"
	"regexp"
	"strings"
	"unicode/utf8"
)

type RestAPIService struct {
	repo       *repository.RestAPIRepository
	projectRepo *repository.ProjectRepository
}

func NewRestAPIService(repo *repository.RestAPIRepository, projectRepo *repository.ProjectRepository) *RestAPIService {
	return &RestAPIService{
		repo:        repo,
		projectRepo: projectRepo,
	}
}

// CreateRestAPI creates a new REST API documentation with OWASP-compliant validation
func (s *RestAPIService) CreateRestAPI(ctx context.Context, userID int64, req model.CreateRestAPIRequest) (*model.RestAPI, error) {
	// Verify project exists and user has access
	_, err := s.projectRepo.FindByID(ctx, req.IDProject)
	if err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	// Note: Authorization is handled by ABAC middleware at controller level

	// Sanitize and validate all inputs
	sanitizedName := sanitizeInput(req.Name)
	sanitizedDescription := sanitizeInput(req.Description)
	sanitizedEndpoint := sanitizeInput(req.Endpoint)

	// Validate name
	if err := ValidateRestAPIName(sanitizedName); err != nil {
		return nil, err
	}

	// Validate description
	if err := validateRestAPIDescription(sanitizedDescription); err != nil {
		return nil, err
	}

	// Validate HTTP method
	if err := ValidateRESTMethod(req.Method); err != nil {
		return nil, err
	}

	// Validate endpoint
	if err := ValidateEndpoint(sanitizedEndpoint); err != nil {
		return nil, err
	}

	// Validate headers
	if len(req.Headers) > 0 {
		if err := validateHeaders(req.Headers); err != nil {
			return nil, err
		}
	}

	// Validate path parameters
	if len(req.PathParams) > 0 {
		if err := validateParameters(req.PathParams); err != nil {
			return nil, err
		}
	}

	// Validate query parameters
	if len(req.QueryParams) > 0 {
		if err := validateParameters(req.QueryParams); err != nil {
			return nil, err
		}
	}

	// Validate request body schema
	if req.RequestBody != nil {
		if err := validateJSONSchema(req.RequestBody); err != nil {
			return nil, err
		}
	}

	// Validate responses
	if len(req.Responses) > 0 {
		if err := validateResponses(req.Responses); err != nil {
			return nil, err
		}
	}

	// Update request with sanitized values
	req.Name = sanitizedName
	req.Description = sanitizedDescription
	req.Endpoint = sanitizedEndpoint

	// Create REST API
	restAPI, err := s.repo.Create(ctx, &req, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to create REST API: %w", err)
	}

	return restAPI, nil
}

// GetRestAPIByID retrieves a REST API by its ID
func (s *RestAPIService) GetRestAPIByID(ctx context.Context, apiID int64) (model.RestAPI, error) {
	restAPI, err := s.repo.FindByID(ctx, apiID)
	if err != nil {
		return model.RestAPI{}, fmt.Errorf("failed to get REST API: %w", err)
	}

	return restAPI, nil
}

// UpdateRestAPI modifies an existing REST API documentation with OWASP-compliant validation
func (s *RestAPIService) UpdateRestAPI(ctx context.Context, apiID int64, req model.UpdateRestAPIRequest) (*model.RestAPI, error) {
	// Check if REST API exists
	_, err := s.repo.FindByID(ctx, apiID)
	if err != nil {
		return nil, fmt.Errorf("REST API not found: %w", err)
	}

	// Note: Authorization is handled by ABAC middleware at controller level

	// Sanitize and validate provided fields
	if req.Name != nil {
		sanitizedName := sanitizeInput(*req.Name)
		if err := ValidateRestAPIName(sanitizedName); err != nil {
			return nil, err
		}
		req.Name = &sanitizedName
	}

	if req.Description != nil {
		sanitizedDescription := sanitizeInput(*req.Description)
		if err := validateRestAPIDescription(sanitizedDescription); err != nil {
			return nil, err
		}
		req.Description = &sanitizedDescription
	}

	if req.Method != nil {
		if err := ValidateRESTMethod(*req.Method); err != nil {
			return nil, err
		}
	}

	if req.Endpoint != nil {
		sanitizedEndpoint := sanitizeInput(*req.Endpoint)
		if err := ValidateEndpoint(sanitizedEndpoint); err != nil {
			return nil, err
		}
		req.Endpoint = &sanitizedEndpoint
	}

	if req.Headers != nil {
		if err := validateHeaders(*req.Headers); err != nil {
			return nil, err
		}
	}

	if req.PathParams != nil {
		if err := validateParameters(*req.PathParams); err != nil {
			return nil, err
		}
	}

	if req.QueryParams != nil {
		if err := validateParameters(*req.QueryParams); err != nil {
			return nil, err
		}
	}

	if req.RequestBody != nil {
		if err := validateJSONSchema(req.RequestBody); err != nil {
			return nil, err
		}
	}

	if req.Responses != nil {
		if err := validateResponses(*req.Responses); err != nil {
			return nil, err
		}
	}

	// Update REST API
	restAPI, err := s.repo.Update(ctx, apiID, &req)
	if err != nil {
		return nil, fmt.Errorf("failed to update REST API: %w", err)
	}

	return restAPI, nil
}

// DeleteRestAPI removes a REST API documentation
func (s *RestAPIService) DeleteRestAPI(ctx context.Context, apiID int64) error {
	// Delete REST API directly (repository handles data access)
	rowsAffected, err := s.repo.Delete(ctx, apiID)
	if err != nil {
		return fmt.Errorf("failed to delete REST API: %w", err)
	}

	// Check if any row was actually deleted
	if rowsAffected == 0 {
		return utility.NotFoundError("REST API not found")
	}

	return nil
}

// ListRestAPIs retrieves REST APIs with pagination and filters
func (s *RestAPIService) ListRestAPIs(ctx context.Context, req model.ListRestAPIsRequest) (*model.PageResult[model.RestAPI], error) {
	// Sanitize search input to prevent SQL injection
	if req.Search != "" {
		req.Search = sanitizeInput(req.Search)
	}

	// Validate method if provided
	if req.Method != "" {
		if err := ValidateRESTMethod(req.Method); err != nil {
			return nil, err
		}
	}

	// Set default values
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 || req.Limit > 100 {
		req.Limit = 20
	}

	// Get REST APIs
	result, err := s.repo.List(ctx, &req)
	if err != nil {
		return nil, fmt.Errorf("failed to list REST APIs: %w", err)
	}

	return result, nil
}

// ==================== REST API SPECIFIC VALIDATION FUNCTIONS ====================

// ValidateRestAPIName performs comprehensive validation on REST API name
// Exported for testing purposes
func ValidateRestAPIName(name string) error {
	// Check length
	if utf8.RuneCountInString(name) < 3 {
		return fmt.Errorf("name must be at least 3 characters")
	}
	if utf8.RuneCountInString(name) > 255 {
		return fmt.Errorf("name must not exceed 255 characters")
	}

	// Check for null bytes
	if strings.Contains(name, "\x00") {
		return fmt.Errorf("name contains null bytes")
	}

	// Check for XSS patterns
	if containsXSSPatterns(name) {
		return fmt.Errorf("name contains potentially dangerous content")
	}

	// Check for SQL injection patterns
	if containsSQLInjectionPatterns(name) {
		return fmt.Errorf("name contains potentially dangerous SQL patterns")
	}

	// Check for path traversal patterns
	if containsPathTraversalPatterns(name) {
		return fmt.Errorf("name contains potentially dangerous path patterns")
	}

	// Allow most characters for API names (more permissive than project names)
	// Using unicode character classes for internationalization support
	matched, _ := regexp.MatchString(`^[\p{L}\p{N}\s\\\-._/:]+$`, name)
	if !matched {
		return fmt.Errorf("name contains invalid characters")
	}

	return nil
}

// validateRestAPIDescription performs validation on REST API description
func validateRestAPIDescription(description string) error {
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

	// Check for SQL injection patterns
	if containsSQLInjectionPatterns(description) {
		return fmt.Errorf("description contains potentially dangerous SQL patterns")
	}

	return nil
}

// ValidateRESTMethod validates HTTP method
// Exported for testing purposes
func ValidateRESTMethod(method string) error {
	// Whitelist of allowed methods
	validMethods := map[string]bool{
		"GET":    true,
		"POST":   true,
		"PUT":    true,
		"DELETE": true,
		"PATCH":  true,
	}

	if !validMethods[method] {
		return fmt.Errorf("invalid HTTP method: %s (must be GET, POST, PUT, DELETE, or PATCH)", method)
	}

	return nil
}

// ValidateEndpoint validates API endpoint path
// Exported for testing purposes
func ValidateEndpoint(endpoint string) error {
	// Check length
	if len(endpoint) == 0 {
		return fmt.Errorf("endpoint is required")
	}
	if len(endpoint) > 500 {
		return fmt.Errorf("endpoint must not exceed 500 characters")
	}

	// Must start with /
	if !strings.HasPrefix(endpoint, "/") {
		return fmt.Errorf("endpoint must start with /")
	}

	// Check for path traversal patterns
	if containsPathTraversalPatterns(endpoint) {
		return fmt.Errorf("endpoint contains potentially dangerous path patterns")
	}

	// Check for null bytes
	if strings.Contains(endpoint, "\x00") {
		return fmt.Errorf("endpoint contains null bytes")
	}

	// Validate endpoint format (allow path parameters like {id}, :id, etc.)
	// Pattern: /path/{param}/path/:param/path
	matched, _ := regexp.MatchString(`^/[a-zA-Z0-9\-_/:{}]*$`, endpoint)
	if !matched {
		return fmt.Errorf("endpoint contains invalid characters")
	}

	return nil
}

// validateHeaders validates header definitions
func validateHeaders(headers []model.Header) error {
	// Check number of headers
	if len(headers) > 100 {
		return fmt.Errorf("too many headers (maximum 100)")
	}

	headerNames := make(map[string]bool)

	for i, header := range headers {
		// Validate header name
		if err := validateHeaderName(header.Name); err != nil {
			return fmt.Errorf("header %d: %w", i+1, err)
		}

		// Check for duplicate header names
		if headerNames[header.Name] {
			return fmt.Errorf("duplicate header name: %s", header.Name)
		}
		headerNames[header.Name] = true

		// Validate description if provided
		if header.Description != "" {
			if err := validateRestAPIDescription(header.Description); err != nil {
				return fmt.Errorf("header %s: %w", header.Name, err)
			}
		}

		// Validate example if provided
		if header.Example != "" {
			if len(header.Example) > 500 {
				return fmt.Errorf("header %s: example too long (maximum 500 characters)", header.Name)
			}
			if containsXSSPatterns(header.Example) {
				return fmt.Errorf("header %s: example contains potentially dangerous content", header.Name)
			}
		}
	}

	return nil
}

// validateHeaderName validates a single header name
func validateHeaderName(name string) error {
	if name == "" {
		return fmt.Errorf("header name is required")
	}

	if len(name) > 100 {
		return fmt.Errorf("header name too long (maximum 100 characters)")
	}

	// HTTP header names should be alphanumeric with hyphens
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9\-]+$`, name)
	if !matched {
		return fmt.Errorf("header name contains invalid characters")
	}

	return nil
}

// validateParameters validates parameter definitions (path or query params)
func validateParameters(params []model.Parameter) error {
	// Check number of parameters
	if len(params) > 50 {
		return fmt.Errorf("too many parameters (maximum 50)")
	}

	paramNames := make(map[string]bool)

	for i, param := range params {
		// Validate parameter name
		if err := validateParameterName(param.Name); err != nil {
			return fmt.Errorf("parameter %d: %w", i+1, err)
		}

		// Check for duplicate parameter names
		if paramNames[param.Name] {
			return fmt.Errorf("duplicate parameter name: %s", param.Name)
		}
		paramNames[param.Name] = true

		// Validate parameter type
		if err := validateParameterType(param.Type); err != nil {
			return fmt.Errorf("parameter %s: %w", param.Name, err)
		}

		// Validate description if provided
		if param.Description != "" {
			if err := validateRestAPIDescription(param.Description); err != nil {
				return fmt.Errorf("parameter %s: %w", param.Name, err)
			}
		}

		// Validate default value if provided
		if param.Default != nil {
			if err := validateDefaultValue(param.Default, param.Type); err != nil {
				return fmt.Errorf("parameter %s: %w", param.Name, err)
			}
		}
	}

	return nil
}

// validateParameterName validates a single parameter name
func validateParameterName(name string) error {
	if name == "" {
		return fmt.Errorf("parameter name is required")
	}

	if len(name) > 100 {
		return fmt.Errorf("parameter name too long (maximum 100 characters)")
	}

	// Parameter names should be alphanumeric with underscores
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_]+$`, name)
	if !matched {
		return fmt.Errorf("parameter name contains invalid characters")
	}

	return nil
}

// validateParameterType validates parameter data type
func validateParameterType(paramType string) error {
	validTypes := map[string]bool{
		"string":  true,
		"integer": true,
		"boolean": true,
		"number":  true,
	}

	if !validTypes[paramType] {
		return fmt.Errorf("invalid parameter type: %s (must be string, integer, boolean, or number)", paramType)
	}

	return nil
}

// validateDefaultValue validates default value against type
func validateDefaultValue(value interface{}, paramType string) error {
	if value == nil {
		return nil
	}

	switch paramType {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("default value must be a string")
		}
	case "integer":
		if _, ok := value.(int); !ok {
			if _, ok := value.(float64); !ok {
				return fmt.Errorf("default value must be an integer")
			}
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("default value must be a boolean")
		}
	case "number":
		if _, ok := value.(float64); !ok {
			if _, ok := value.(int); !ok {
				return fmt.Errorf("default value must be a number")
			}
		}
	}

	return nil
}

// validateJSONSchema recursively validates JSON schema
func validateJSONSchema(schema *model.JSONSchema) error {
	if schema == nil {
		return nil
	}

	// Validate type
	if err := validateJSONSchemaType(schema.Type); err != nil {
		return fmt.Errorf("invalid JSON schema type: %w", err)
	}

	// Validate properties for object type
	if schema.Type == "object" {
		if len(schema.Properties) > 100 {
			return fmt.Errorf("too many properties in JSON schema (maximum 100)")
		}

		for propName, prop := range schema.Properties {
			// Validate property name
			if len(propName) > 100 {
				return fmt.Errorf("property name too long: %s", propName)
			}

			// Recursively validate property
			if err := validateProperty(&prop); err != nil {
				return fmt.Errorf("property %s: %w", propName, err)
			}
		}
	}

	// Validate items for array type
	if schema.Type == "array" && schema.Items != nil {
		if err := validateJSONSchema(schema.Items); err != nil {
			return fmt.Errorf("array items: %w", err)
		}
	}

	return nil
}

// validateProperty validates a JSON schema property
func validateProperty(prop *model.Property) error {
	// Validate type
	if err := validateJSONSchemaType(prop.Type); err != nil {
		return fmt.Errorf("invalid property type: %w", err)
	}

	// Validate description if provided
	if prop.Description != "" {
		if utf8.RuneCountInString(prop.Description) > 500 {
			return fmt.Errorf("description too long (maximum 500 characters)")
		}
		if containsXSSPatterns(prop.Description) {
			return fmt.Errorf("description contains potentially dangerous content")
		}
	}

	// Validate nested properties for object type
	if prop.Type == "object" && len(prop.Properties) > 0 {
		if len(prop.Properties) > 50 {
			return fmt.Errorf("too many nested properties (maximum 50)")
		}
		for propName, nestedProp := range prop.Properties {
			if err := validateProperty(&nestedProp); err != nil {
				return fmt.Errorf("nested property %s: %w", propName, err)
			}
		}
	}

	// Validate items for array type
	if prop.Type == "array" && prop.Items != nil {
		if err := validateJSONSchema(prop.Items); err != nil {
			return fmt.Errorf("array items: %w", err)
		}
	}

	// Validate enum if provided
	if len(prop.Enum) > 0 {
		if len(prop.Enum) > 50 {
			return fmt.Errorf("too many enum values (maximum 50)")
		}
		for _, enumVal := range prop.Enum {
			if utf8.RuneCountInString(enumVal) > 100 {
				return fmt.Errorf("enum value too long")
			}
		}
	}

	return nil
}

// validateJSONSchemaType validates JSON schema type
func validateJSONSchemaType(schemaType string) error {
	validTypes := map[string]bool{
		"object":  true,
		"array":   true,
		"string":  true,
		"number":  true,
		"integer": true,
		"boolean": true,
		"null":    true,
	}

	if !validTypes[schemaType] {
		return fmt.Errorf("invalid JSON schema type: %s", schemaType)
	}

	return nil
}

// validateResponses validates response examples
func validateResponses(responses map[int]model.ResponseExample) error {
	// Check number of responses
	if len(responses) > 20 {
		return fmt.Errorf("too many response examples (maximum 20)")
	}

	// Validate status codes
	for statusCode, resp := range responses {
		if statusCode < 100 || statusCode > 599 {
			return fmt.Errorf("invalid status code: %d", statusCode)
		}

		// Validate description
		if resp.Description != "" {
			if utf8.RuneCountInString(resp.Description) > 500 {
				return fmt.Errorf("response %d: description too long", statusCode)
			}
			if containsXSSPatterns(resp.Description) {
				return fmt.Errorf("response %d: description contains potentially dangerous content", statusCode)
			}
		}

		// Validate body (basic check - just ensure it's JSON-serializable)
		if resp.Body != nil {
			// Body can be any JSON-serializable type, so we just check for size
			bodyJSON, err := json.Marshal(resp.Body)
			if err != nil {
				return fmt.Errorf("response %d: body is not JSON-serializable: %w", statusCode, err)
			}
			if len(bodyJSON) > 10000 { // 10KB limit per response
				return fmt.Errorf("response %d: body too large (maximum 10KB)", statusCode)
			}
		}
	}

	return nil
}
