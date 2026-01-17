package repository

import (
	"context"
	"fmt"
	"golang-basic/api/internal/model"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ProjectRepository struct {
	db *pgxpool.Pool
}

func NewProjectRepository(db *pgxpool.Pool) *ProjectRepository {
	return &ProjectRepository{db: db}
}

// Create inserts a new project into the database
func (r *ProjectRepository) Create(ctx context.Context, req *model.CreateProjectRequest, userID int64) (*model.Project, error) {
	// Generate slug from name
	slug := generateSlug(req.Name)

	query := `
		INSERT INTO projects (id_user, name, slug, description, version, is_public, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING id_project, created_at, updated_at
	`

	var project model.Project
	err := r.db.QueryRow(ctx, query,
		userID,
		req.Name,
		slug,
		req.Description,
		req.Version,
		req.IsPublic,
	).Scan(&project.IDProject, &project.CreatedAt, &project.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	project.IDUser = userID
	project.Name = req.Name
	project.Slug = slug
	project.Description = req.Description
	project.Version = req.Version
	project.IsPublic = req.IsPublic

	return &project, nil
}

// FindByID retrieves a project by its ID
func (r *ProjectRepository) FindByID(ctx context.Context, projectID int64) (model.Project, error) {
	query := `
		SELECT id_project, id_user, name, slug, description, version, is_public, created_at, updated_at
		FROM projects
		WHERE id_project = $1
	`

	var project model.Project
	err := r.db.QueryRow(ctx, query, projectID).Scan(
		&project.IDProject,
		&project.IDUser,
		&project.Name,
		&project.Slug,
		&project.Description,
		&project.Version,
		&project.IsPublic,
		&project.CreatedAt,
		&project.UpdatedAt,
	)

	if err != nil {
		return model.Project{}, fmt.Errorf("failed to find project: %w", err)
	}

	return project, nil
}

// FindBySlug retrieves a project by its slug and user ID
func (r *ProjectRepository) FindBySlug(ctx context.Context, slug string, userID int64) (model.Project, error) {
	query := `
		SELECT id_project, id_user, name, slug, description, version, is_public, created_at, updated_at
		FROM projects
		WHERE slug = $1 AND id_user = $2
	`

	var project model.Project
	err := r.db.QueryRow(ctx, query, slug, userID).Scan(
		&project.IDProject,
		&project.IDUser,
		&project.Name,
		&project.Slug,
		&project.Description,
		&project.Version,
		&project.IsPublic,
		&project.CreatedAt,
		&project.UpdatedAt,
	)

	if err != nil {
		return model.Project{}, fmt.Errorf("failed to find project by slug: %w", err)
	}

	return project, nil
}

// Update modifies an existing project
func (r *ProjectRepository) Update(ctx context.Context, projectID int64, req *model.UpdateProjectRequest) (*model.Project, error) {
	// First, get the existing project
	_, err := r.FindByID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	// Build the update query dynamically based on provided fields
	query := `
		UPDATE projects
		SET name = COALESCE($1, name),
		    description = COALESCE($2, description),
		    version = COALESCE($3, version),
		    is_public = COALESCE($4, is_public),
		    updated_at = NOW()
		WHERE id_project = $5
		RETURNING id_project, id_user, name, slug, description, version, is_public, created_at, updated_at
	`

	var project model.Project
	err = r.db.QueryRow(ctx, query,
		req.Name,
		req.Description,
		req.Version,
		req.IsPublic,
		projectID,
	).Scan(
		&project.IDProject,
		&project.IDUser,
		&project.Name,
		&project.Slug,
		&project.Description,
		&project.Version,
		&project.IsPublic,
		&project.CreatedAt,
		&project.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	return &project, nil
}

// Delete removes a project from the database
func (r *ProjectRepository) Delete(ctx context.Context, projectID int64) error {
	query := `DELETE FROM projects WHERE id_project = $1`

	_, err := r.db.Exec(ctx, query, projectID)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	return nil
}

// FindAll retrieves all projects with pagination and filters
func (r *ProjectRepository) FindAll(ctx context.Context, req model.ListProjectsRequest) (*model.PageResult[model.Project], error) {
	// Validate pagination parameters
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 || req.Limit > 100 {
		req.Limit = 10
	}

	// Calculate offset
	offset := (req.Page - 1) * req.Limit

	// Build WHERE clause dynamically
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argNum := 1

	if req.IDUser != nil {
		whereClause += fmt.Sprintf(" AND id_user = $%d", argNum)
		args = append(args, *req.IDUser)
		argNum++
	}

	if req.IsPublic != nil {
		whereClause += fmt.Sprintf(" AND is_public = $%d", argNum)
		args = append(args, *req.IsPublic)
		argNum++
	}

	if req.Search != "" {
		whereClause += fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d)", argNum, argNum+1)
		searchPattern := "%" + req.Search + "%"
		args = append(args, searchPattern, searchPattern)
		argNum += 2
	}

	// Build ORDER BY clause
	orderBy := "created_at DESC"
	if req.SortBy != "" {
		validSortFields := map[string]string{
			"name":       "name",
			"created_at": "created_at",
			"updated_at": "updated_at",
		}
		if field, ok := validSortFields[req.SortBy]; ok {
			orderBy = field
			if req.SortOrder == "asc" || req.SortOrder == "desc" {
				orderBy += " " + strings.ToUpper(req.SortOrder)
			} else {
				orderBy += " DESC"
			}
		}
	}

	// Main query
	query := fmt.Sprintf(`
		SELECT id_project, id_user, name, slug, description, version, is_public, created_at, updated_at
		FROM projects
		%s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderBy, argNum, argNum+1)

	args = append(args, req.Limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query projects: %w", err)
	}
	defer rows.Close()

	var projects []model.Project
	for rows.Next() {
		var project model.Project
		err := rows.Scan(
			&project.IDProject,
			&project.IDUser,
			&project.Name,
			&project.Slug,
			&project.Description,
			&project.Version,
			&project.IsPublic,
			&project.CreatedAt,
			&project.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}
		projects = append(projects, project)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("error iterating projects: %w", rows.Err())
	}

	// Get total count for pagination
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM projects %s", whereClause)
	var totalCount int64
	err = r.db.QueryRow(ctx, countQuery, args[:argNum-2]...).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count projects: %w", err)
	}

	result := &model.PageResult[model.Project]{
		Data:     projects,
		Page:     req.Page,
		Size:     req.Limit,
		StartRow: offset + 1,
		EndRow:   offset + len(projects),
	}

	return result, nil
}

// generateSlug creates a URL-friendly slug from a project name
func generateSlug(name string) string {
	// Convert to lowercase
	slug := strings.ToLower(name)

	// Replace spaces and special characters with hyphens
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	slug = reg.ReplaceAllString(slug, "-")

	// Remove leading/trailing hyphens
	slug = strings.Trim(slug, "-")

	// Limit length to 100 characters
	if len(slug) > 100 {
		slug = slug[:100]
	}

	return slug
}
