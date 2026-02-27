package service

import (
	"context"
	"fmt"
	"golang-basic/api/internal/model"
	"golang-basic/api/internal/repository"
	"strconv"
	"strings"
)

// UserAttributeService handles business logic for user attribute assignments
type UserAttributeService struct {
	userAttrRepo *repository.UserAttributeRepository
	attrRepo     *repository.PermissionRepository
	userRepo     *repository.UserRepository
}

// NewUserAttributeService creates a new UserAttributeService
func NewUserAttributeService(
	userAttrRepo *repository.UserAttributeRepository,
	attrRepo *repository.PermissionRepository,
	userRepo *repository.UserRepository,
) *UserAttributeService {
	return &UserAttributeService{
		userAttrRepo: userAttrRepo,
		attrRepo:     attrRepo,
		userRepo:     userRepo,
	}
}

// ListUserAttributes retrieves all user attributes with optional user filter
func (s *UserAttributeService) ListUserAttributes(ctx context.Context, userID int64) ([]model.UserAttributeDetail, error) {
	if userID > 0 {
		return s.userAttrRepo.GetUserAttributes(ctx, userID)
	}
	return s.userAttrRepo.GetAllUserAttributes(ctx)
}

// GetUserAttributeByID retrieves a user attribute by ID
func (s *UserAttributeService) GetUserAttributeByID(ctx context.Context, id int64) (*model.UserAttributeDetail, error) {
	if err := s.validateID(id, "id"); err != nil {
		return nil, err
	}

	attr, err := s.userAttrRepo.GetUserAttributeByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve user attribute: %w", err)
	}

	return attr, nil
}

// CreateUserAttribute creates a new user attribute assignment with validation
func (s *UserAttributeService) CreateUserAttribute(ctx context.Context, req model.CreateUserAttributeRequest) (*model.UserAttribute, error) {
	// Validate user ID
	if err := s.validateID(req.UserID, "id_user"); err != nil {
		return nil, err
	}

	// Validate attribute ID
	if err := s.validateID(req.AttributeID, "id_attribute"); err != nil {
		return nil, err
	}

	// Validate value is not empty
	if strings.TrimSpace(req.Value) == "" {
		return nil, ValidationError{
			Field:   "value",
			Message: "cannot be empty",
		}
	}

	// Check if user exists
	user, err := s.userRepo.FindByID(ctx, req.UserID)
	if err != nil || user.UserId == 0 {
		return nil, ValidationError{
			Field:   "id_user",
			Message: "user not found",
		}
	}

	// Check if attribute exists
	attribute, err := s.attrRepo.GetAttributeByID(ctx, req.AttributeID)
	if err != nil {
		return nil, ValidationError{
			Field:   "id_attribute",
			Message: "attribute not found",
		}
	}

	// Validate value matches attribute type
	if err := s.validateAttributeValue(req.Value, attribute.Type, attribute.EnumValues); err != nil {
		return nil, err
	}

	// Check for duplicate (user, attribute, value) combination
	existing, _ := s.userAttrRepo.GetUserAttributesByNameAndValue(ctx, req.UserID, attribute.Name, req.Value)
	if existing != nil {
		return nil, ValidationError{
			Field:   "value",
			Message: "user already has this attribute with this value",
		}
	}

	// Create user attribute
	created, err := s.userAttrRepo.CreateUserAttribute(ctx, &req)
	if err != nil {
		return nil, fmt.Errorf("failed to create user attribute: %w", err)
	}

	return created, nil
}

// UpdateUserAttribute updates an existing user attribute with validation
func (s *UserAttributeService) UpdateUserAttribute(ctx context.Context, id int64, req model.UpdateUserAttributeRequest) (*model.UserAttribute, error) {
	// Validate ID
	if err := s.validateID(id, "id"); err != nil {
		return nil, err
	}

	// Get existing user attribute to retrieve attribute info
	existing, err := s.userAttrRepo.GetUserAttributeByID(ctx, id)
	if err != nil {
		return nil, ValidationError{
			Field:   "id",
			Message: "user attribute not found",
		}
	}

	// Validate value is not empty
	if req.Value == nil || strings.TrimSpace(*req.Value) == "" {
		return nil, ValidationError{
			Field:   "value",
			Message: "cannot be empty",
		}
	}

	// Get attribute details for validation
	attribute, err := s.attrRepo.GetAttributeByID(ctx, existing.AttributeID)
	if err != nil {
		return nil, ValidationError{
			Field:   "id_attribute",
			Message: "attribute not found",
		}
	}

	// Validate value matches attribute type
	if err := s.validateAttributeValue(*req.Value, attribute.Type, attribute.EnumValues); err != nil {
		return nil, err
	}

	// Check for duplicate if value is different
	if *req.Value != existing.Value {
		duplicate, _ := s.userAttrRepo.GetUserAttributesByNameAndValue(ctx, existing.UserID, attribute.Name, *req.Value)
		if duplicate != nil && duplicate.UserAttributeID != id {
			return nil, ValidationError{
				Field:   "value",
				Message: "user already has this attribute with this value",
			}
		}
	}

	// Update user attribute
	updated, err := s.userAttrRepo.UpdateUserAttribute(ctx, id, &req)
	if err != nil {
		return nil, fmt.Errorf("failed to update user attribute: %w", err)
	}

	return updated, nil
}

// DeleteUserAttribute deletes a user attribute by ID
func (s *UserAttributeService) DeleteUserAttribute(ctx context.Context, id int64) error {
	if err := s.validateID(id, "id"); err != nil {
		return err
	}

	err := s.userAttrRepo.DeleteUserAttribute(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete user attribute: %w", err)
	}

	return nil
}

// validateAttributeValue validates that the value matches the attribute type
func (s *UserAttributeService) validateAttributeValue(value string, attrType string, enumValues []string) error {
	switch attrType {
	case "string":
		if len(value) > 1000 {
			return ValidationError{
				Field:   "value",
				Message: "string value too long (max 1000 characters)",
			}
		}
		return nil

	case "number":
		if _, err := strconv.ParseFloat(value, 64); err != nil {
			return ValidationError{
				Field:   "value",
				Message: "must be a valid number",
			}
		}
		return nil

	case "boolean":
		if value != "true" && value != "false" {
			return ValidationError{
				Field:   "value",
				Message: "must be 'true' or 'false'",
			}
		}
		return nil

	case "enum":
		if len(enumValues) == 0 {
			return ValidationError{
				Field:   "value",
				Message: "attribute has no enum values defined",
			}
		}
		for _, enumVal := range enumValues {
			if value == enumVal {
				return nil
			}
		}
		return ValidationError{
			Field:   "value",
			Message: fmt.Sprintf("must be one of: %s", strings.Join(enumValues, ", ")),
		}

	default:
		return ValidationError{
			Field:   "type",
			Message: "unknown attribute type",
		}
	}
}

// validateID validates that an ID is positive
func (s *UserAttributeService) validateID(id int64, fieldName string) error {
	if id <= 0 {
		return ValidationError{
			Field:   fieldName,
			Message: "must be a positive integer",
		}
	}
	return nil
}
