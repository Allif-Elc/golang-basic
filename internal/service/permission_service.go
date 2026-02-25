package service

import (
	"context"
	"errors"
	"fmt"
	"golang-basic/api/internal/model"
	"golang-basic/api/internal/repository"
	"regexp"
	"strings"
	"unicode/utf8"
)

var (
	// OWASP Validation Patterns for Permissions
	// Prevents injection attacks by ensuring only valid characters
	permissionNameRegex = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
	// Allows letters, numbers, spaces, and common punctuation for descriptions
	permissionDescriptionRegex = regexp.MustCompile(`^[A-Za-z0-9\s.,!?;:'"()-]+$`)
)

// Validation constants based on OWASP recommendations
const (
	MinNameLength        = 1
	MaxNameLength        = 100
	MaxDescriptionLength = 1000
)

type PermissionService struct {
	repo *repository.PermissionRepository
}

func NewPermissionService(repo *repository.PermissionRepository) *PermissionService {
	return &PermissionService{repo: repo}
}

// ValidationError represents a field validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// validateName validates permission/attribute/resource names according to OWASP standards
func (s *PermissionService) validateName(name string, fieldName string) error {
	nameLen := utf8.RuneCountInString(name)

	// Length validation - prevent buffer overflow and DoS
	if nameLen < MinNameLength || nameLen > MaxNameLength {
		return ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("length must be between %d and %d characters", MinNameLength, MaxNameLength),
		}
	}

	// Character validation - prevent injection attacks
	// Only allow alphanumeric characters, hyphens, and underscores
	if !permissionNameRegex.MatchString(name) {
		return ValidationError{
			Field:   fieldName,
			Message: "only alphanumeric characters, hyphens, and underscores are allowed (A-Z, a-z, 0-9, -, _)",
		}
	}

	// SQL injection prevention - check for dangerous patterns
	dangerousPatterns := []string{
		"SELECT", "INSERT", "UPDATE", "DELETE", "DROP", "UNION",
		"OR", "AND", "--", ";", "/*", "*/", "xp_", "sp_",
	}
	upperName := strings.ToUpper(name)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(upperName, pattern) {
			return ValidationError{
				Field:   fieldName,
				Message: "contains potentially dangerous patterns",
			}
		}
	}

	return nil
}

// validateDescription validates description text according to OWASP standards
func (s *PermissionService) validateDescription(description string, fieldName string) error {
	if description == "" {
		return nil // Optional field
	}

	descLen := utf8.RuneCountInString(description)

	// Length validation - prevent buffer overflow and DoS
	if descLen > MaxDescriptionLength {
		return ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("maximum length is %d characters", MaxDescriptionLength),
		}
	}

	// Character validation - prevent XSS and injection attacks
	if !permissionDescriptionRegex.MatchString(description) {
		return ValidationError{
			Field:   fieldName,
			Message: "contains invalid characters",
		}
	}

	return nil
}

// validateID validates that an ID is positive
func (s *PermissionService) validateID(id int64, fieldName string) error {
	if id <= 0 {
		return ValidationError{
			Field:   fieldName,
			Message: "must be a positive integer",
		}
	}
	return nil
}

// CreateAttribute creates a new attribute with OWASP validation
func (s *PermissionService) CreateAttribute(ctx context.Context, req model.CreateAttributesRequest) (*model.Attributes, error) {
	// Validate input
	if err := s.validateName(req.Name, "name"); err != nil {
		return nil, err
	}
	if err := s.validateDescription(req.Description, "description"); err != nil {
		return nil, err
	}
	if err := s.validateAttributeType(req.Type); err != nil {
		return nil, err
	}
	if err := s.validateEnumValues(req.Type, req.EnumValues); err != nil {
		return nil, err
	}

	// Check if attribute already exists (uniqueness validation)
	existing, err := s.repo.GetAttributeByName(ctx, req.Name)
	if err == nil && existing != nil {
		return nil, ValidationError{
			Field:   "name",
			Message: "attribute with this name already exists",
		}
	}

	// Create attribute
	created, err := s.repo.CreateAttributes(ctx, &req)
	if err != nil {
		return nil, fmt.Errorf("failed to create attribute: %w", err)
	}

	return created, nil
}

// GetAttributeByID retrieves an attribute by ID with validation
func (s *PermissionService) GetAttributeByID(ctx context.Context, id int64) (*model.Attributes, error) {
	if err := s.validateID(id, "id"); err != nil {
		return nil, err
	}

	attribute, err := s.repo.GetAttributeByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve attribute: %w", err)
	}

	return attribute, nil
}

// ListAttributes retrieves all attributes
func (s *PermissionService) ListAttributes(ctx context.Context) ([]model.Attributes, error) {
	attributes, err := s.repo.ListAttributes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list attributes: %w", err)
	}

	return attributes, nil
}

// DeleteAttribute deletes an attribute by ID with validation
func (s *PermissionService) DeleteAttribute(ctx context.Context, id int64) error {
	if err := s.validateID(id, "id"); err != nil {
		return fmt.Errorf("invalid ID: %w", err)
	}

	err := s.repo.DeleteAttribute(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete attribute: %w", err)
	}

	return nil
}

// CreateResource creates a new resource with OWASP validation
func (s *PermissionService) CreateResource(ctx context.Context, req model.CreateResoucesRequest) (*model.Resources, error) {
	// Validate input
	if err := s.validateName(req.Name, "name"); err != nil {
		return nil, err
	}
	if err := s.validateDescription(req.Description, "description"); err != nil {
		return nil, err
	}
	if err := s.validateResourceType(req.ResourceType); err != nil {
		return nil, err
	}

	// Check if resource already exists (uniqueness validation)
	existing, err := s.repo.GetResourceByName(ctx, req.Name)
	if err == nil && existing != nil {
		return nil, ValidationError{
			Field:   "name",
			Message: "resource with this name already exists",
		}
	}

	// Create resource
	created, err := s.repo.CreateResoucesRequest(ctx, &req)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	return created, nil
}

// GetResourceByID retrieves a resource by ID with validation
func (s *PermissionService) GetResourceByID(ctx context.Context, id int64) (*model.Resources, error) {
	if err := s.validateID(id, "id"); err != nil {
		return nil, err
	}

	resource, err := s.repo.GetResourceByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve resource: %w", err)
	}

	return resource, nil
}

// ListResources retrieves all resources
func (s *PermissionService) ListResources(ctx context.Context) ([]model.Resources, error) {
	resources, err := s.repo.ListResources(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list resources: %w", err)
	}

	return resources, nil
}

// DeleteResource deletes a resource by ID with validation
func (s *PermissionService) DeleteResource(ctx context.Context, id int64) error {
	if err := s.validateID(id, "id"); err != nil {
		return fmt.Errorf("invalid ID: %w", err)
	}

	err := s.repo.DeleteResource(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete resource: %w", err)
	}

	return nil
}

// CreatePermission creates a new permission with OWASP validation
func (s *PermissionService) CreatePermission(ctx context.Context, req model.CreatePermissionsRequest) (*model.Permissions, error) {
	// Validate input
	if err := s.validateName(req.Name, "name"); err != nil {
		return nil, err
	}
	if err := s.validateDescription(req.Description, "description"); err != nil {
		return nil, err
	}
	if err := s.validateEffect(req.Effect); err != nil {
		return nil, err
	}
	if err := s.validateActions(req.Actions); err != nil {
		return nil, err
	}

	// Check if permission already exists (uniqueness validation)
	existing, err := s.repo.GetPermissionByName(ctx, req.Name)
	if err == nil && existing != nil {
		return nil, ValidationError{
			Field:   "name",
			Message: "permission with this name already exists",
		}
	}

	// Create permission
	created, err := s.repo.CreatePermissionsRequest(ctx, &req)
	if err != nil {
		return nil, fmt.Errorf("failed to create permission: %w", err)
	}

	return created, nil
}

// GetPermissionByID retrieves a permission by ID with validation
func (s *PermissionService) GetPermissionByID(ctx context.Context, id int64) (*model.Permissions, error) {
	if err := s.validateID(id, "id"); err != nil {
		return nil, err
	}

	permission, err := s.repo.GetPermissionByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve permission: %w", err)
	}

	return permission, nil
}

// ListPermissions retrieves all permissions
func (s *PermissionService) ListPermissions(ctx context.Context) ([]model.Permissions, error) {
	permissions, err := s.repo.ListPermissions(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list permissions: %w", err)
	}

	return permissions, nil
}

// DeletePermission deletes a permission by ID with validation
func (s *PermissionService) DeletePermission(ctx context.Context, id int64) error {
	if err := s.validateID(id, "id"); err != nil {
		return fmt.Errorf("invalid ID: %w", err)
	}

	err := s.repo.DeletePermission(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete permission: %w", err)
	}

	return nil
}

// ValidateMultipleErrors checks if an error is a ValidationError
func IsValidationError(err error) bool {
	_, ok := err.(ValidationError)
	return ok
}

// AggregateErrors aggregates multiple validation errors into a single error
func AggregateErrors(errs []error) error {
	if len(errs) == 0 {
		return nil
	}

	var messages []string
	for _, err := range errs {
		messages = append(messages, err.Error())
	}

	return errors.New(strings.Join(messages, "; "))
}

// validateAttributeType validates the attribute type
func (s *PermissionService) validateAttributeType(attrType string) error {
	validTypes := map[string]bool{"string": true, "number": true, "boolean": true, "enum": true}
	if !validTypes[attrType] {
		return ValidationError{
			Field:   "type",
			Message: "must be one of: string, number, boolean, enum",
		}
	}
	return nil
}

// validateEnumValues validates enum values when type is enum
func (s *PermissionService) validateEnumValues(attrType string, enumValues []string) error {
	if attrType == "enum" && len(enumValues) == 0 {
		return ValidationError{
			Field:   "enum_values",
			Message: "required when type is enum",
		}
	}
	return nil
}

// validateEffect validates the permission effect
func (s *PermissionService) validateEffect(effect string) error {
	validEffects := map[string]bool{"allow": true, "deny": true}
	if !validEffects[effect] {
		return ValidationError{
			Field:   "effect",
			Message: "must be either 'allow' or 'deny'",
		}
	}
	return nil
}

// validateActions validates the permission actions
func (s *PermissionService) validateActions(actions []string) error {
	if len(actions) == 0 {
		return ValidationError{
			Field:   "actions",
			Message: "at least one action is required",
		}
	}
	return nil
}

// validateResourceType validates the resource type
func (s *PermissionService) validateResourceType(resourceType string) error {
	if resourceType == "" {
		return nil // Optional field
	}
	validTypes := map[string]bool{"general": true, "api": true, "document": true, "data": true}
	if !validTypes[resourceType] {
		return ValidationError{
			Field:   "resource_type",
			Message: "must be one of: general, api, document, data",
		}
	}
	return nil
}

// UpdateAttribute updates an attribute by ID with validation
func (s *PermissionService) UpdateAttribute(ctx context.Context, id int64, req model.UpdateAttributesRequest) (*model.Attributes, error) {
	if err := s.validateID(id, "id"); err != nil {
		return nil, err
	}

	if req.Name != nil {
		if err := s.validateName(*req.Name, "name"); err != nil {
			return nil, err
		}
	}

	if req.Type != nil {
		if err := s.validateAttributeType(*req.Type); err != nil {
			return nil, err
		}
		if *req.Type == "enum" {
			if err := s.validateEnumValues(*req.Type, req.EnumValues); err != nil {
				return nil, err
			}
		}
	}

	updated, err := s.repo.UpdateAttribute(ctx, id, &req)
	if err != nil {
		return nil, fmt.Errorf("failed to update attribute: %w", err)
	}

	return updated, nil
}

// UpdateResource updates a resource by ID with validation
func (s *PermissionService) UpdateResource(ctx context.Context, id int64, req model.UpdateResourcesRequest) (*model.Resources, error) {
	if err := s.validateID(id, "id"); err != nil {
		return nil, err
	}

	if req.Name != nil {
		if err := s.validateName(*req.Name, "name"); err != nil {
			return nil, err
		}
	}

	if req.ResourceType != nil {
		if err := s.validateResourceType(*req.ResourceType); err != nil {
			return nil, err
		}
	}

	updated, err := s.repo.UpdateResource(ctx, id, &req)
	if err != nil {
		return nil, fmt.Errorf("failed to update resource: %w", err)
	}

	return updated, nil
}

// UpdatePermission updates a permission by ID with validation
func (s *PermissionService) UpdatePermission(ctx context.Context, id int64, req model.UpdatePermissionsRequest) (*model.Permissions, error) {
	if err := s.validateID(id, "id"); err != nil {
		return nil, err
	}

	if req.Name != nil {
		if err := s.validateName(*req.Name, "name"); err != nil {
			return nil, err
		}
	}

	if req.Effect != nil {
		if err := s.validateEffect(*req.Effect); err != nil {
			return nil, err
		}
	}

	if len(req.Actions) > 0 {
		if err := s.validateActions(req.Actions); err != nil {
			return nil, err
		}
	}

	updated, err := s.repo.UpdatePermission(ctx, id, &req)
	if err != nil {
		return nil, fmt.Errorf("failed to update permission: %w", err)
	}

	return updated, nil
}
