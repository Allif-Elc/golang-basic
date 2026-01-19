package repository

import (
	"context"
	"golang-basic/api/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PermissionRepository struct {
	db *pgxpool.Pool
}

func NewPermissionRepository(db *pgxpool.Pool) *PermissionRepository {
	return &PermissionRepository{db: db}
}

func (r *PermissionRepository) CreateAttributes(ctx context.Context, attribute *model.CreateAttributesRequest) (*model.Attributes, error) {
	query := `INSERT into attributes (name, description) VALUES ($1, $2) RETURNING id_attribute, created_at`
	var createdAttribute model.Attributes
	err := r.db.QueryRow(ctx, query, attribute.Name, attribute.Description).Scan(&createdAttribute.AttributeID, &createdAttribute.CreateAt)
	if err != nil {
		return nil, err
	}

	createdAttribute.Name = attribute.Name
	createdAttribute.Description = attribute.Description
	return &createdAttribute, nil
}

func (r *PermissionRepository) CreateResoucesRequest(ctx context.Context, resource *model.CreateResoucesRequest) (*model.Resources, error) {
	query := `INSERT into resources (name, description) VALUES ($1, $2) RETURNING id_resource, created_at`
	var createdResouce model.Resources
	err := r.db.QueryRow(ctx, query, resource.Name, resource.Description).Scan(&createdResouce.ResourceID, &createdResouce.CreateAt)
	if err != nil {
		return nil, err
	}

	createdResouce.Name = resource.Name
	createdResouce.Description = resource.Description
	return &createdResouce, nil
}

func (r *PermissionRepository) CreatePermissionsRequest(ctx context.Context, permission *model.CreatePermissionsRequest) (*model.Permissions, error) {
	query := `INSERT into permissions (name, description) VALUES ($1, $2) RETURNING id_permission`
	var createdPermission model.Permissions
	err := r.db.QueryRow(ctx, query, permission.Name, permission.Description).Scan(&createdPermission.PermissionID)
	if err != nil {
		return nil, err
	}

	createdPermission.Name = permission.Name
	createdPermission.Description = permission.Description
	return &createdPermission, nil
}

// GetAttributeByID retrieves an attribute by its ID
func (r *PermissionRepository) GetAttributeByID(ctx context.Context, id int64) (*model.Attributes, error) {
	query := `SELECT id_attribute, name, description, created_at FROM attributes WHERE id_attribute = $1`
	var attribute model.Attributes
	err := r.db.QueryRow(ctx, query, id).Scan(
		&attribute.AttributeID,
		&attribute.Name,
		&attribute.Description,
		&attribute.CreateAt,
	)
	if err != nil {
		return nil, err
	}
	return &attribute, nil
}

// GetAttributeByName retrieves an attribute by its name
// Required indexes:
//   CREATE UNIQUE INDEX idx_attributes_name ON attributes(name);
func (r *PermissionRepository) GetAttributeByName(ctx context.Context, name string) (*model.Attributes, error) {
	query := `SELECT id_attribute, name, description, created_at FROM attributes WHERE name = $1`
	var attribute model.Attributes
	err := r.db.QueryRow(ctx, query, name).Scan(
		&attribute.AttributeID,
		&attribute.Name,
		&attribute.Description,
		&attribute.CreateAt,
	)
	if err != nil {
		return nil, err
	}
	return &attribute, nil
}

// ListAttributes retrieves all attributes
// Required indexes:
//   CREATE INDEX idx_attributes_id ON attributes(id_attribute);
func (r *PermissionRepository) ListAttributes(ctx context.Context) ([]model.Attributes, error) {
	query := `SELECT id_attribute, name, description, created_at FROM attributes ORDER BY id_attribute`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attributes []model.Attributes
	for rows.Next() {
		var attribute model.Attributes
		err := rows.Scan(
			&attribute.AttributeID,
			&attribute.Name,
			&attribute.Description,
			&attribute.CreateAt,
		)
		if err != nil {
			return nil, err
		}
		attributes = append(attributes, attribute)
	}

	return attributes, rows.Err()
}

// DeleteAttribute deletes an attribute by ID
func (r *PermissionRepository) DeleteAttribute(ctx context.Context, id int64) error {
	query := `DELETE FROM attributes WHERE id_attribute = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// GetResourceByID retrieves a resource by its ID
func (r *PermissionRepository) GetResourceByID(ctx context.Context, id int64) (*model.Resources, error) {
	query := `SELECT id_resource, name, description, created_at FROM resources WHERE id_resource = $1`
	var resource model.Resources
	err := r.db.QueryRow(ctx, query, id).Scan(
		&resource.ResourceID,
		&resource.Name,
		&resource.Description,
		&resource.CreateAt,
	)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

// GetResourceByName retrieves a resource by its name
// Required indexes:
//   CREATE UNIQUE INDEX idx_resources_name ON resources(name);
func (r *PermissionRepository) GetResourceByName(ctx context.Context, name string) (*model.Resources, error) {
	query := `SELECT id_resource, name, description, created_at FROM resources WHERE name = $1`
	var resource model.Resources
	err := r.db.QueryRow(ctx, query, name).Scan(
		&resource.ResourceID,
		&resource.Name,
		&resource.Description,
		&resource.CreateAt,
	)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

// ListResources retrieves all resources
// Required indexes:
//   CREATE INDEX idx_resources_id ON resources(id_resource);
func (r *PermissionRepository) ListResources(ctx context.Context) ([]model.Resources, error) {
	query := `SELECT id_resource, name, description, created_at FROM resources ORDER BY id_resource`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resources []model.Resources
	for rows.Next() {
		var resource model.Resources
		err := rows.Scan(
			&resource.ResourceID,
			&resource.Name,
			&resource.Description,
			&resource.CreateAt,
		)
		if err != nil {
			return nil, err
		}
		resources = append(resources, resource)
	}

	return resources, rows.Err()
}

// DeleteResource deletes a resource by ID
func (r *PermissionRepository) DeleteResource(ctx context.Context, id int64) error {
	query := `DELETE FROM resources WHERE id_resource = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// GetPermissionByID retrieves a permission by its ID
func (r *PermissionRepository) GetPermissionByID(ctx context.Context, id int64) (*model.Permissions, error) {
	query := `SELECT id_permission, name, description, created_at FROM permissions WHERE id_permission = $1`
	var permission model.Permissions
	err := r.db.QueryRow(ctx, query, id).Scan(
		&permission.PermissionID,
		&permission.Name,
		&permission.Description,
		&permission.CreateAt,
	)
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

// GetPermissionByName retrieves a permission by its name
// Required indexes:
//   CREATE UNIQUE INDEX idx_permissions_name ON permissions(name);
func (r *PermissionRepository) GetPermissionByName(ctx context.Context, name string) (*model.Permissions, error) {
	query := `SELECT id_permission, name, description, created_at FROM permissions WHERE name = $1`
	var permission model.Permissions
	err := r.db.QueryRow(ctx, query, name).Scan(
		&permission.PermissionID,
		&permission.Name,
		&permission.Description,
		&permission.CreateAt,
	)
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

// ListPermissions retrieves all permissions
// Required indexes:
//   CREATE INDEX idx_permissions_id ON permissions(id_permission);
func (r *PermissionRepository) ListPermissions(ctx context.Context) ([]model.Permissions, error) {
	query := `SELECT id_permission, name, description, created_at FROM permissions ORDER BY id_permission`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []model.Permissions
	for rows.Next() {
		var permission model.Permissions
		err := rows.Scan(
			&permission.PermissionID,
			&permission.Name,
			&permission.Description,
			&permission.CreateAt,
		)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, permission)
	}

	return permissions, rows.Err()
}

// DeletePermission deletes a permission by ID
func (r *PermissionRepository) DeletePermission(ctx context.Context, id int64) error {
	query := `DELETE FROM permissions WHERE id_permission = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return pgx.ErrNoRows
	}

	return nil
}
