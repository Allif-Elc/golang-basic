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

type GraphQLAPIService struct {
	repo        *repository.GraphQLAPIRepository
	projectRepo *repository.ProjectRepository
}

func NewGraphQLAPIService(repo *repository.GraphQLAPIRepository, projectRepo *repository.ProjectRepository) *GraphQLAPIService {
	return &GraphQLAPIService{
		repo:        repo,
		projectRepo: projectRepo,
	}
}

// CreateGraphQLAPI creates a new GraphQL API documentation
// Per Specs: Authorization handled by ABAC middleware at controller level
func (s *GraphQLAPIService) CreateGraphQLAPI(ctx context.Context, userID int64, req model.CreateGraphQLAPIRequest) (*model.GraphQLAPI, error) {
	// Verify project exists and user has access
	_, err := s.projectRepo.FindByID(ctx, req.IDProject)
	if err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	// Sanitize and validate all inputs
	sanitizedName := sanitizeInput(req.Name)
	sanitizedDescription := sanitizeInput(req.Description)
	sanitizedReturnType := sanitizeInput(req.ReturnType)

	// Validate name
	if err := ValidateGraphQLAPIName(sanitizedName); err != nil {
		return nil, err
	}

	// Validate description
	if err := validateGraphQLAPIDescription(sanitizedDescription); err != nil {
		return nil, err
	}

	// Validate GraphQL operation type
	if err := ValidateGraphQLOperationType(req.Type); err != nil {
		return nil, err
	}

	// Validate return type
	if err := validateGraphQLReturnType(sanitizedReturnType); err != nil {
		return nil, err
	}

	// Validate arguments
	if len(req.Arguments) > 0 {
		if err := validateGraphQLArguments(req.Arguments); err != nil {
			return nil, err
		}
	}

	// Validate examples
	if len(req.Examples) > 0 {
		if err := validateGraphQLExamples(req.Examples); err != nil {
			return nil, err
		}
	}

	// Update request with sanitized values
	req.Name = sanitizedName
	req.Description = sanitizedDescription
	req.ReturnType = sanitizedReturnType

	// Create GraphQL API
	graphqlAPI, err := s.repo.Create(ctx, &req, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to create GraphQL API: %w", err)
	}

	return graphqlAPI, nil
}

// GetGraphQLAPIByID retrieves a GraphQL API by its ID
func (s *GraphQLAPIService) GetGraphQLAPIByID(ctx context.Context, apiID int64) (model.GraphQLAPI, error) {
	graphqlAPI, err := s.repo.FindByID(ctx, apiID)
	if err != nil {
		return model.GraphQLAPI{}, fmt.Errorf("failed to get GraphQL API: %w", err)
	}

	return graphqlAPI, nil
}

// UpdateGraphQLAPI modifies an existing GraphQL API documentation
func (s *GraphQLAPIService) UpdateGraphQLAPI(ctx context.Context, apiID int64, req model.UpdateGraphQLAPIRequest) (*model.GraphQLAPI, error) {
	// Check if GraphQL API exists
	_, err := s.repo.FindByID(ctx, apiID)
	if err != nil {
		return nil, fmt.Errorf("GraphQL API not found: %w", err)
	}

	// Sanitize and validate provided fields
	if req.Name != nil {
		sanitizedName := sanitizeInput(*req.Name)
		if err := ValidateGraphQLAPIName(sanitizedName); err != nil {
			return nil, err
		}
		req.Name = &sanitizedName
	}

	if req.Description != nil {
		sanitizedDescription := sanitizeInput(*req.Description)
		if err := validateGraphQLAPIDescription(sanitizedDescription); err != nil {
			return nil, err
		}
		req.Description = &sanitizedDescription
	}

	if req.Type != nil {
		if err := ValidateGraphQLOperationType(*req.Type); err != nil {
			return nil, err
		}
	}

	if req.ReturnType != nil {
		sanitizedReturnType := sanitizeInput(*req.ReturnType)
		if err := validateGraphQLReturnType(sanitizedReturnType); err != nil {
			return nil, err
		}
		req.ReturnType = &sanitizedReturnType
	}

	if req.Arguments != nil {
		if err := validateGraphQLArguments(*req.Arguments); err != nil {
			return nil, err
		}
	}

	if req.Examples != nil {
		if err := validateGraphQLExamples(*req.Examples); err != nil {
			return nil, err
		}
	}

	// Update GraphQL API
	graphqlAPI, err := s.repo.Update(ctx, apiID, &req)
	if err != nil {
		return nil, fmt.Errorf("failed to update GraphQL API: %w", err)
	}

	return graphqlAPI, nil
}

// DeleteGraphQLAPI removes a GraphQL API documentation
func (s *GraphQLAPIService) DeleteGraphQLAPI(ctx context.Context, apiID int64) error {
	rowsAffected, err := s.repo.Delete(ctx, apiID)
	if err != nil {
		return fmt.Errorf("failed to delete GraphQL API: %w", err)
	}

	if rowsAffected == 0 {
		return utility.NotFoundError("GraphQL API not found")
	}

	return nil
}

// ListGraphQLAPIs retrieves GraphQL APIs with pagination and filters
func (s *GraphQLAPIService) ListGraphQLAPIs(ctx context.Context, req model.ListGraphQLAPIsRequest) (*model.PageResult[model.GraphQLAPI], error) {
	// Sanitize search input to prevent SQL injection
	if req.Search != "" {
		req.Search = sanitizeInput(req.Search)
	}

	// Validate type if provided
	if req.Type != "" {
		if err := ValidateGraphQLOperationType(req.Type); err != nil {
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

	// Get GraphQL APIs
	result, err := s.repo.List(ctx, &req)
	if err != nil {
		return nil, fmt.Errorf("failed to list GraphQL APIs: %w", err)
	}

	return result, nil
}

// ==================== GRAPHQL API SPECIFIC VALIDATION FUNCTIONS ====================

// ValidateGraphQLAPIName performs comprehensive validation on GraphQL API name
func ValidateGraphQLAPIName(name string) error {
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

	// GraphQL field names should follow GraphQL naming conventions
	// Must start with a letter or underscore, followed by letters, digits, or underscores
	matched, _ := regexp.MatchString(`^[a-zA-Z_][a-zA-Z0-9_]*$`, name)
	if !matched {
		return fmt.Errorf("name contains invalid characters (must be a valid GraphQL field name)")
	}

	return nil
}

// validateGraphQLAPIDescription performs validation on GraphQL API description
func validateGraphQLAPIDescription(description string) error {
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

// ValidateGraphQLOperationType validates GraphQL operation type (query, mutation, subscription)
func ValidateGraphQLOperationType(operationType string) error {
	validTypes := map[string]bool{
		"query":        true,
		"mutation":     true,
		"subscription": true,
	}

	if !validTypes[operationType] {
		return fmt.Errorf("invalid GraphQL operation type: %s (must be query, mutation, or subscription)", operationType)
	}

	return nil
}

// validateGraphQLReturnType validates GraphQL return type
func validateGraphQLReturnType(returnType string) error {
	if returnType == "" {
		return fmt.Errorf("return type is required")
	}

	if utf8.RuneCountInString(returnType) > 500 {
		return fmt.Errorf("return type must not exceed 500 characters")
	}

	// Basic validation for GraphQL type format
	// Allows: String, Int, [String], CustomType, CustomType!, etc.
	matched, _ := regexp.MatchString(`^[a-zA-Z][a-zA-Z0-9_![\]<>{}(),\s]*$`, returnType)
	if !matched {
		return fmt.Errorf("return type contains invalid characters")
	}

	return nil
}

// validateGraphQLArguments validates GraphQL argument definitions
func validateGraphQLArguments(args []model.GraphQLArgument) error {
	// Check number of arguments
	if len(args) > 50 {
		return fmt.Errorf("too many arguments (maximum 50)")
	}

	argNames := make(map[string]bool)

	for i, arg := range args {
		// Validate argument name
		if err := validateGraphQLArgumentName(arg.Name); err != nil {
			return fmt.Errorf("argument %d: %w", i+1, err)
		}

		// Check for duplicate argument names
		if argNames[arg.Name] {
			return fmt.Errorf("duplicate argument name: %s", arg.Name)
		}
		argNames[arg.Name] = true

		// Validate argument type
		if err := validateGraphQLType(arg.Type); err != nil {
			return fmt.Errorf("argument %s: %w", arg.Name, err)
		}

		// Validate description if provided
		if arg.Description != "" {
			if utf8.RuneCountInString(arg.Description) > 500 {
				return fmt.Errorf("argument %s: description too long (maximum 500 characters)", arg.Name)
			}
			if containsXSSPatterns(arg.Description) {
				return fmt.Errorf("argument %s: description contains potentially dangerous content", arg.Name)
			}
		}

		// Validate default value if provided
		if arg.DefaultValue != nil {
			// Default values can be any valid JSON value
			// We'll do basic validation here
			if arg.Required && arg.DefaultValue != nil {
				return fmt.Errorf("argument %s: required arguments cannot have default values", arg.Name)
			}
		}
	}

	return nil
}

// validateGraphQLArgumentName validates a single GraphQL argument name
func validateGraphQLArgumentName(name string) error {
	if name == "" {
		return fmt.Errorf("argument name is required")
	}

	if len(name) > 100 {
		return fmt.Errorf("argument name too long (maximum 100 characters)")
	}

	// GraphQL argument names should follow GraphQL naming conventions
	matched, _ := regexp.MatchString(`^[a-zA-Z_][a-zA-Z0-9_]*$`, name)
	if !matched {
		return fmt.Errorf("argument name contains invalid characters")
	}

	return nil
}

// validateGraphQLType validates a GraphQL type string
func validateGraphQLType(graphqlType string) error {
	if graphqlType == "" {
		return fmt.Errorf("type is required")
	}

	if len(graphqlType) > 100 {
		return fmt.Errorf("type too long (maximum 100 characters)")
	}

	// Basic validation for GraphQL type format
	// Allows: String, String!, [String], [String]!, CustomType, etc.
	matched, _ := regexp.MatchString(`^[a-zA-Z][a-zA-Z0-9_![\]<>]*$`, graphqlType)
	if !matched {
		return fmt.Errorf("type contains invalid characters")
	}

	return nil
}

// validateGraphQLExamples validates GraphQL query examples
func validateGraphQLExamples(examples []model.GraphQLExample) error {
	// Check number of examples
	if len(examples) > 20 {
		return fmt.Errorf("too many examples (maximum 20)")
	}

	for i, example := range examples {
		// Validate example name
		if example.Name == "" {
			return fmt.Errorf("example %d: name is required", i+1)
		}
		if utf8.RuneCountInString(example.Name) > 100 {
			return fmt.Errorf("example %d: name too long (maximum 100 characters)", i+1)
		}

		// Validate query
		if example.Query == "" {
			return fmt.Errorf("example %d: query is required", i+1)
		}
		if len(example.Query) > 10000 {
			return fmt.Errorf("example %d: query too long (maximum 10000 characters)", i+1)
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

		// Validate response if provided
		if example.Response != nil {
			// Response can be any JSON-serializable type
			// Just check it's not a string (to avoid injection)
			if _, ok := example.Response.(string); ok {
				return fmt.Errorf("example %d: response must be a JSON object, not a string", i+1)
			}
		}
	}

	return nil
}
