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

// GetAPIStats retrieves the count of REST, GraphQL, and gRPC APIs for a project
func (r *ProjectRepository) GetAPIStats(ctx context.Context, projectID int64) (map[string]int64, error) {
	query := `
		SELECT
			(SELECT COUNT(*) FROM rest_apis WHERE id_project = $1) as rest_count,
			(SELECT COUNT(*) FROM graphql_apis WHERE id_project = $1) as graphql_count,
			(SELECT COUNT(*) FROM grpc_apis WHERE id_project = $1) as grpc_count
	`
	var restCount, graphqlCount, grpcCount int64
	err := r.db.QueryRow(ctx, query, projectID).Scan(&restCount, &graphqlCount, &grpcCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get API stats: %w", err)
	}
	return map[string]int64{
		"rest":     restCount,
		"graphql": graphqlCount,
		"grpc":     grpcCount,
	}, nil
}

// DeleteAllAPIsByProject explicitly deletes all APIs for a project
func (r *ProjectRepository) DeleteAllAPIsByProject(ctx context.Context, projectID int64) (map[string]int64, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var restCount, graphqlCount, grpcCount int64

	// Delete REST APIs
	tag, err := tx.Exec(ctx, "DELETE FROM rest_apis WHERE id_project = $1", projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete REST APIs: %w", err)
	}
	restCount = tag.RowsAffected()

	// Delete GraphQL APIs
	tag, err = tx.Exec(ctx, "DELETE FROM graphql_apis WHERE id_project = $1", projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete GraphQL APIs: %w", err)
	}
	graphqlCount = tag.RowsAffected()

	// Delete gRPC APIs
	tag, err = tx.Exec(ctx, "DELETE FROM grpc_apis WHERE id_project = $1", projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete gRPC APIs: %w", err)
	}
	grpcCount = tag.RowsAffected()

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return map[string]int64{
		"rest":     restCount,
		"graphql": graphqlCount,
		"grpc":     grpcCount,
	}, nil
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

	var whereBuilder strings.Builder
	whereBuilder.WriteString("WHERE 1=1")
	args := []interface{}{}
	argNum := 1

	if req.IDUser != nil {
		whereBuilder.WriteString(fmt.Sprintf(" AND id_user = $%d", argNum))
		args = append(args, *req.IDUser)
		argNum++
	}

	if req.IsPublic != nil {
		whereBuilder.WriteString(fmt.Sprintf(" AND is_public = $%d", argNum))
		args = append(args, *req.IsPublic)
		argNum++
	}

	if req.Search != "" {
		whereBuilder.WriteString(fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d)", argNum, argNum+1))
		searchPattern := "%" + req.Search + "%"
		args = append(args, searchPattern, searchPattern)
		argNum += 2
	}
	whereClause := whereBuilder.String()

	var orderByBuilder strings.Builder
	orderByBuilder.WriteString("created_at")
	if req.SortBy != "" {
		validSortFields := map[string]string{
			"name":       "name",
			"created_at": "created_at",
			"updated_at": "updated_at",
		}
		if field, ok := validSortFields[req.SortBy]; ok {
			orderByBuilder.Reset()
			orderByBuilder.WriteString(field)
			if req.SortOrder == "asc" || req.SortOrder == "desc" {
				orderByBuilder.WriteString(" ")
				orderByBuilder.WriteString(strings.ToUpper(req.SortOrder))
			} else {
				orderByBuilder.WriteString(" DESC")
			}
		}
	}
	orderBy := orderByBuilder.String()

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
	if argNum > 2 {
		err = r.db.QueryRow(ctx, countQuery, args[:argNum-2]...).Scan(&totalCount)
	} else {
		err = r.db.QueryRow(ctx, countQuery).Scan(&totalCount)
	}
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

// FindAllWithStats retrieves all projects with API statistics in a single query using LATERAL JOIN
// This eliminates N+1 queries by fetching API counts along with project data
func (r *ProjectRepository) FindAllWithStats(ctx context.Context, req model.ListProjectsRequest) (*model.PageResult[model.ProjectWithStats], error) {
	// Validate pagination parameters
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 || req.Limit > 100 {
		req.Limit = 10
	}

	// Calculate offset
	offset := (req.Page - 1) * req.Limit

	var whereBuilder strings.Builder
	whereBuilder.WriteString("WHERE 1=1")
	args := []interface{}{}
	argNum := 1

	if req.IDUser != nil {
		whereBuilder.WriteString(fmt.Sprintf(" AND p.id_user = $%d", argNum))
		args = append(args, *req.IDUser)
		argNum++
	}

	if req.IsPublic != nil {
		whereBuilder.WriteString(fmt.Sprintf(" AND p.is_public = $%d", argNum))
		args = append(args, *req.IsPublic)
		argNum++
	}

	if req.Search != "" {
		whereBuilder.WriteString(fmt.Sprintf(" AND (p.name ILIKE $%d OR p.description ILIKE $%d)", argNum, argNum+1))
		searchPattern := "%" + req.Search + "%"
		args = append(args, searchPattern, searchPattern)
		argNum += 2
	}
	whereClause := whereBuilder.String()

	var orderByBuilder strings.Builder
	orderByBuilder.WriteString("p.created_at")
	if req.SortBy != "" {
		validSortFields := map[string]string{
			"name":       "p.name",
			"created_at": "p.created_at",
			"updated_at": "p.updated_at",
		}
		if field, ok := validSortFields[req.SortBy]; ok {
			orderByBuilder.Reset()
			orderByBuilder.WriteString(field)
			if req.SortOrder == "asc" || req.SortOrder == "desc" {
				orderByBuilder.WriteString(" ")
				orderByBuilder.WriteString(strings.ToUpper(req.SortOrder))
			} else {
				orderByBuilder.WriteString(" DESC")
			}
		}
	}
	orderBy := orderByBuilder.String()

	// Main query with LATERAL JOINs for API stats - single round-trip
	query := fmt.Sprintf(`
		SELECT
			p.id_project, p.id_user, p.name, p.slug, p.description, p.version, p.is_public,
			p.created_at, p.updated_at,
			COALESCE(ra.rest_count, 0) as rest_count,
			COALESCE(ga.graphql_count, 0) as graphql_count,
			COALESCE(gr.grpc_count, 0) as grpc_count
		FROM projects p
		LEFT JOIN LATERAL (
			SELECT COUNT(*) as rest_count
			FROM rest_apis
			WHERE id_project = p.id_project
		) ra ON true
		LEFT JOIN LATERAL (
			SELECT COUNT(*) as graphql_count
			FROM graphql_apis
			WHERE id_project = p.id_project
		) ga ON true
		LEFT JOIN LATERAL (
			SELECT COUNT(*) as grpc_count
			FROM grpc_apis
			WHERE id_project = p.id_project
		) gr ON true
		%s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderBy, argNum, argNum+1)

	args = append(args, req.Limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query projects with stats: %w", err)
	}
	defer rows.Close()

	var projects []model.ProjectWithStats
	for rows.Next() {
		var pws model.ProjectWithStats
		err := rows.Scan(
			&pws.IDProject,
			&pws.IDUser,
			&pws.Name,
			&pws.Slug,
			&pws.Description,
			&pws.Version,
			&pws.IsPublic,
			&pws.CreatedAt,
			&pws.UpdatedAt,
			&pws.RestCount,
			&pws.GraphQLCount,
			&pws.GrpcCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project with stats: %w", err)
		}
		pws.TotalAPICount = pws.RestCount + pws.GraphQLCount + pws.GrpcCount
		projects = append(projects, pws)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("error iterating projects with stats: %w", rows.Err())
	}

	// Get total count for pagination
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM projects p %s", whereClause)
	var totalCount int64
	if argNum > 2 {
		err = r.db.QueryRow(ctx, countQuery, args[:argNum-2]...).Scan(&totalCount)
	} else {
		err = r.db.QueryRow(ctx, countQuery).Scan(&totalCount)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to count projects: %w", err)
	}

	result := &model.PageResult[model.ProjectWithStats]{
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
