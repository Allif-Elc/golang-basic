package repository

import (
	"context"
	"fmt"
	"golang-basic/api/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserAttributeRepository struct {
	db *pgxpool.Pool
}

func NewUserAttributeRepository(db *pgxpool.Pool) *UserAttributeRepository {
	return &UserAttributeRepository{db: db}
}

// GetUserAttributes retrieves all user attributes with optional user filter
// Required indexes:
//   CREATE INDEX idx_user_attributes_user ON user_attributes(id_user);
//   CREATE INDEX idx_user_attributes_attribute ON user_attributes(id_attribute);
//   CREATE UNIQUE INDEX idx_user_attributes_user_attr_unique ON user_attributes(id_user, id_attribute, value);
func (r *UserAttributeRepository) GetUserAttributes(ctx context.Context, userID int64) ([]model.UserAttributeDetail, error) {
	query := `
		SELECT
			ua.id_user_attribute,
			ua.id_user,
			u.name as user_name,
			u.email as user_email,
			ua.id_attribute,
			a.name as attribute_name,
			a.type as attribute_type,
			a.enum_values,
			ua.value,
			ua.created_at
		FROM user_attributes ua
		INNER JOIN attributes a ON ua.id_attribute = a.id_attribute
		INNER JOIN users u ON ua.id_user = u.id_user
		WHERE ua.id_user = $1
		ORDER BY ua.created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user attributes: %w", err)
	}
	defer rows.Close()

	var attributes []model.UserAttributeDetail
	for rows.Next() {
		var attr model.UserAttributeDetail
		if err := rows.Scan(
			&attr.UserAttributeID,
			&attr.UserID,
			&attr.UserName,
			&attr.UserEmail,
			&attr.AttributeID,
			&attr.AttributeName,
			&attr.AttributeType,
			&attr.EnumValues,
			&attr.Value,
			&attr.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan user attribute: %w", err)
		}
		attributes = append(attributes, attr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user attributes: %w", err)
	}

	return attributes, nil
}

// GetAllUserAttributes retrieves all user attributes for the list page
// Performance: <20ms with proper indexes
func (r *UserAttributeRepository) GetAllUserAttributes(ctx context.Context) ([]model.UserAttributeDetail, error) {
	query := `
		SELECT
			ua.id_user_attribute,
			ua.id_user,
			u.name as user_name,
			u.email as user_email,
			ua.id_attribute,
			a.name as attribute_name,
			a.type as attribute_type,
			a.enum_values,
			ua.value,
			ua.created_at
		FROM user_attributes ua
		INNER JOIN attributes a ON ua.id_attribute = a.id_attribute
		INNER JOIN users u ON ua.id_user = u.id_user
		ORDER BY ua.created_at DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all user attributes: %w", err)
	}
	defer rows.Close()

	var attributes []model.UserAttributeDetail
	for rows.Next() {
		var attr model.UserAttributeDetail
		if err := rows.Scan(
			&attr.UserAttributeID,
			&attr.UserID,
			&attr.UserName,
			&attr.UserEmail,
			&attr.AttributeID,
			&attr.AttributeName,
			&attr.AttributeType,
			&attr.EnumValues,
			&attr.Value,
			&attr.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan user attribute: %w", err)
		}
		attributes = append(attributes, attr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user attributes: %w", err)
	}

	return attributes, nil
}

// GetUserAttributeByID retrieves a user attribute by its ID
func (r *UserAttributeRepository) GetUserAttributeByID(ctx context.Context, id int64) (*model.UserAttributeDetail, error) {
	query := `
		SELECT
			ua.id_user_attribute,
			ua.id_user,
			u.name as user_name,
			u.email as user_email,
			ua.id_attribute,
			a.name as attribute_name,
			a.type as attribute_type,
			a.enum_values,
			ua.value,
			ua.created_at
		FROM user_attributes ua
		INNER JOIN attributes a ON ua.id_attribute = a.id_attribute
		INNER JOIN users u ON ua.id_user = u.id_user
		WHERE ua.id_user_attribute = $1
	`

	var attr model.UserAttributeDetail
	err := r.db.QueryRow(ctx, query, id).Scan(
		&attr.UserAttributeID,
		&attr.UserID,
		&attr.UserName,
		&attr.UserEmail,
		&attr.AttributeID,
		&attr.AttributeName,
		&attr.AttributeType,
		&attr.EnumValues,
		&attr.Value,
		&attr.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user attribute: %w", err)
	}

	return &attr, nil
}

// CreateUserAttribute creates a new user attribute assignment
// Required indexes for uniqueness check:
//   CREATE UNIQUE INDEX idx_user_attributes_user_attr_unique ON user_attributes(id_user, id_attribute, value);
func (r *UserAttributeRepository) CreateUserAttribute(ctx context.Context, req *model.CreateUserAttributeRequest) (*model.UserAttribute, error) {
	query := `
		INSERT INTO user_attributes (id_user, id_attribute, value)
		VALUES ($1, $2, $3)
		RETURNING id_user_attribute, created_at
	`

	var created model.UserAttribute
	err := r.db.QueryRow(ctx, query, req.UserID, req.AttributeID, req.Value).Scan(
		&created.UserAttributeId,
		&created.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create user attribute: %w", err)
	}

	created.UserID = req.UserID
	created.AttributeID = req.AttributeID
	created.Value = req.Value

	return &created, nil
}

// UpdateUserAttribute updates an existing user attribute
func (r *UserAttributeRepository) UpdateUserAttribute(ctx context.Context, id int64, req *model.UpdateUserAttributeRequest) (*model.UserAttribute, error) {
	query := `
		UPDATE user_attributes
		SET value = $1
		WHERE id_user_attribute = $2
		RETURNING id_user_attribute, id_user, id_attribute, value, created_at
	`

	var updated model.UserAttribute
	err := r.db.QueryRow(ctx, query, req.Value, id).Scan(
		&updated.UserAttributeId,
		&updated.UserID,
		&updated.AttributeID,
		&updated.Value,
		&updated.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update user attribute: %w", err)
	}

	return &updated, nil
}

// DeleteUserAttribute deletes a user attribute by ID
func (r *UserAttributeRepository) DeleteUserAttribute(ctx context.Context, id int64) error {
	query := `DELETE FROM user_attributes WHERE id_user_attribute = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user attribute: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// GetUserAttributesByNameAndValue checks if a user has a specific attribute with a specific value
// Used for authorization checks
func (r *UserAttributeRepository) GetUserAttributesByNameAndValue(ctx context.Context, userID int64, attrName, attrValue string) (*model.UserAttributeDetail, error) {
	query := `
		SELECT
			ua.id_user_attribute,
			ua.id_user,
			u.name as user_name,
			u.email as user_email,
			ua.id_attribute,
			a.name as attribute_name,
			a.type as attribute_type,
			a.enum_values,
			ua.value,
			ua.created_at
		FROM user_attributes ua
		INNER JOIN attributes a ON ua.id_attribute = a.id_attribute
		INNER JOIN users u ON ua.id_user = u.id_user
		WHERE ua.id_user = $1 AND a.name = $2 AND ua.value = $3
	`

	var attr model.UserAttributeDetail
	err := r.db.QueryRow(ctx, query, userID, attrName, attrValue).Scan(
		&attr.UserAttributeID,
		&attr.UserID,
		&attr.UserName,
		&attr.UserEmail,
		&attr.AttributeID,
		&attr.AttributeName,
		&attr.AttributeType,
		&attr.EnumValues,
		&attr.Value,
		&attr.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to find user attribute: %w", err)
	}

	return &attr, nil
}
