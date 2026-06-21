package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"golang-basic/api/internal/config"
	"golang-basic/api/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GraphQLAPIRepository struct {
	db *pgxpool.Pool
}

func NewGraphQLAPIRepository(db *pgxpool.Pool) *GraphQLAPIRepository {
	return &GraphQLAPIRepository{db: db}
}

func (r *GraphQLAPIRepository) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, config.QueryTimeout)
}

// Create inserts a new GraphQL API documentation into the database
// Per Specs: explicit columns (no SELECT *), context timeout, defer Close()
func (r *GraphQLAPIRepository) Create(ctx context.Context, req *model.CreateGraphQLAPIRequest, userID int64) (*model.GraphQLAPI, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	var argumentsBytes, examplesBytes []byte
	var err error

	if req.Arguments != nil {
		argumentsBytes, err = json.Marshal(req.Arguments)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal arguments: %w", err)
		}
	}

	if req.Examples != nil {
		examplesBytes, err = json.Marshal(req.Examples)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal examples: %w", err)
		}
	}

	query := `
		INSERT INTO graphql_apis (id_project, id_user, name, type, description, arguments, return_type, examples, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		RETURNING id_graphql_api, created_at, updated_at
	`

	var graphqlAPI model.GraphQLAPI
	err = r.db.QueryRow(ctx, query,
		req.IDProject,
		userID,
		req.Name,
		req.Type,
		req.Description,
		argumentsBytes,
		req.ReturnType,
		examplesBytes,
	).Scan(&graphqlAPI.IDGraphqlAPI, &graphqlAPI.CreatedAt, &graphqlAPI.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create GraphQL API: %w", err)
	}

	graphqlAPI.IDProject = req.IDProject
	graphqlAPI.IDUser = userID
	graphqlAPI.Name = req.Name
	graphqlAPI.Type = req.Type
	graphqlAPI.Description = req.Description
	graphqlAPI.Arguments = argumentsBytes
	graphqlAPI.ReturnType = req.ReturnType
	graphqlAPI.Examples = examplesBytes

	return &graphqlAPI, nil
}

// FindByID retrieves a GraphQL API by its ID
// Per Specs: explicit columns, QueryRow for single row
func (r *GraphQLAPIRepository) FindByID(ctx context.Context, apiID int64) (model.GraphQLAPI, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	query := `
		SELECT id_graphql_api, id_project, id_user, name, type, description, arguments, return_type, examples, created_at, updated_at
		FROM graphql_apis
		WHERE id_graphql_api = $1
	`

	var graphqlAPI model.GraphQLAPI
	err := r.db.QueryRow(ctx, query, apiID).Scan(
		&graphqlAPI.IDGraphqlAPI,
		&graphqlAPI.IDProject,
		&graphqlAPI.IDUser,
		&graphqlAPI.Name,
		&graphqlAPI.Type,
		&graphqlAPI.Description,
		&graphqlAPI.Arguments,
		&graphqlAPI.ReturnType,
		&graphqlAPI.Examples,
		&graphqlAPI.CreatedAt,
		&graphqlAPI.UpdatedAt,
	)

	if err != nil {
		return model.GraphQLAPI{}, fmt.Errorf("failed to find GraphQL API: %w", err)
	}

	return graphqlAPI, nil
}

// Update modifies an existing GraphQL API documentation
func (r *GraphQLAPIRepository) Update(ctx context.Context, apiID int64, req *model.UpdateGraphQLAPIRequest) (*model.GraphQLAPI, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	var name interface{} = nil
	var apiType interface{} = nil
	var description interface{} = nil
	var returnType interface{} = nil
	var argumentsBytes []byte = nil
	var examplesBytes []byte = nil
	var err error

	if req.Name != nil {
		name = *req.Name
	}
	if req.Type != nil {
		apiType = *req.Type
	}
	if req.Description != nil {
		description = *req.Description
	}
	if req.ReturnType != nil {
		returnType = *req.ReturnType
	}
	if req.Arguments != nil {
		argumentsBytes, err = json.Marshal(*req.Arguments)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal arguments: %w", err)
		}
	}
	if req.Examples != nil {
		examplesBytes, err = json.Marshal(*req.Examples)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal examples: %w", err)
		}
	}

	query := `
		UPDATE graphql_apis
		SET name = COALESCE($1, name),
		    type = COALESCE($2, type),
		    description = COALESCE($3, description),
		    arguments = COALESCE($4, arguments),
		    return_type = COALESCE($5, return_type),
		    examples = COALESCE($6, examples),
		    updated_at = NOW()
		WHERE id_graphql_api = $7
		RETURNING id_graphql_api, id_project, id_user, name, type, description, arguments, return_type, examples, created_at, updated_at
	`

	var graphqlAPI model.GraphQLAPI
	err = r.db.QueryRow(ctx, query,
		name, apiType, description, returnType, argumentsBytes, examplesBytes, apiID,
	).Scan(
		&graphqlAPI.IDGraphqlAPI,
		&graphqlAPI.IDProject,
		&graphqlAPI.IDUser,
		&graphqlAPI.Name,
		&graphqlAPI.Type,
		&graphqlAPI.Description,
		&graphqlAPI.Arguments,
		&graphqlAPI.ReturnType,
		&graphqlAPI.Examples,
		&graphqlAPI.CreatedAt,
		&graphqlAPI.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update GraphQL API: %w", err)
	}

	return &graphqlAPI, nil
}

// Delete removes a GraphQL API from the database and returns rows affected
func (r *GraphQLAPIRepository) Delete(ctx context.Context, apiID int64) (int64, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	query := `DELETE FROM graphql_apis WHERE id_graphql_api = $1`

	result, err := r.db.Exec(ctx, query, apiID)
	if err != nil {
		return 0, fmt.Errorf("failed to delete GraphQL API: %w", err)
	}

	rowsAffected := result.RowsAffected()
	return rowsAffected, nil
}

// List retrieves GraphQL APIs with pagination and filters
// Per Specs: explicit columns, indexed fields (id_project, type), use EXISTS for subqueries
func (r *GraphQLAPIRepository) List(ctx context.Context, req *model.ListGraphQLAPIsRequest) (*model.PageResult[model.GraphQLAPI], error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 || req.Limit > 100 {
		req.Limit = 20
	}

	offset := (req.Page - 1) * req.Limit

	whereClause := "WHERE 1=1"
	countArgs := []interface{}{}
	queryArgs := []interface{}{}
	argNum := 1

	if req.IDProject != nil {
		whereClause += fmt.Sprintf(" AND id_project = $%d", argNum)
		countArgs = append(countArgs, *req.IDProject)
		queryArgs = append(queryArgs, *req.IDProject)
		argNum++
	}

	if req.Type != "" {
		whereClause += fmt.Sprintf(" AND type = $%d", argNum)
		countArgs = append(countArgs, req.Type)
		queryArgs = append(queryArgs, req.Type)
		argNum++
	}

	if req.Search != "" {
		whereClause += fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d)", argNum, argNum+1)
		searchPattern := "%" + req.Search + "%"
		countArgs = append(countArgs, searchPattern, searchPattern)
		queryArgs = append(queryArgs, searchPattern, searchPattern)
		argNum += 2
	}

	orderClause := "ORDER BY created_at DESC"

	queryArgs = append(queryArgs, req.Limit, offset)
	query := fmt.Sprintf(`
		SELECT id_graphql_api, id_project, id_user, name, type, description, arguments, return_type, examples, created_at, updated_at
		FROM graphql_apis
		%s
		%s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderClause, argNum, argNum+1)

	rows, err := r.db.Query(ctx, query, queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to query GraphQL APIs: %w", err)
	}
	defer rows.Close()

	graphqlAPIs := make([]model.GraphQLAPI, 0, req.Limit)
	for rows.Next() {
		var graphqlAPI model.GraphQLAPI
		err := rows.Scan(
			&graphqlAPI.IDGraphqlAPI,
			&graphqlAPI.IDProject,
			&graphqlAPI.IDUser,
			&graphqlAPI.Name,
			&graphqlAPI.Type,
			&graphqlAPI.Description,
			&graphqlAPI.Arguments,
			&graphqlAPI.ReturnType,
			&graphqlAPI.Examples,
			&graphqlAPI.CreatedAt,
			&graphqlAPI.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan GraphQL API: %w", err)
		}
		graphqlAPIs = append(graphqlAPIs, graphqlAPI)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("error iterating GraphQL APIs: %w", rows.Err())
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM graphql_apis %s", whereClause)
	var totalCount int64
	err = r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count GraphQL APIs: %w", err)
	}

	result := &model.PageResult[model.GraphQLAPI]{
		Data:     graphqlAPIs,
		Page:     req.Page,
		Size:     req.Limit,
		StartRow: offset + 1,
		EndRow:   offset + len(graphqlAPIs),
	}

	return result, nil
}

// FindByProjectID retrieves all GraphQL APIs for a specific project
func (r *GraphQLAPIRepository) FindByProjectID(ctx context.Context, projectID int64, req *model.ListGraphQLAPIsRequest) (*model.PageResult[model.GraphQLAPI], error) {
	req.IDProject = &projectID
	return r.List(ctx, req)
}

// ListByProjectIDPaginated retrieves GraphQL APIs for a specific project with pagination
// Returns APIs, total count, and error
func (r *GraphQLAPIRepository) ListByProjectIDPaginated(ctx context.Context, projectID int64, page, limit int) ([]model.GraphQLAPI, int, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Main query with explicit columns for performance
	query := `
		SELECT id_graphql_api, id_project, id_user, name, type, description, arguments, return_type, examples, created_at, updated_at
		FROM graphql_apis
		WHERE id_project = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, projectID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query GraphQL APIs: %w", err)
	}
	defer rows.Close()

	graphqlAPIs := make([]model.GraphQLAPI, 0, limit)
	for rows.Next() {
		var graphqlAPI model.GraphQLAPI
		err := rows.Scan(
			&graphqlAPI.IDGraphqlAPI,
			&graphqlAPI.IDProject,
			&graphqlAPI.IDUser,
			&graphqlAPI.Name,
			&graphqlAPI.Type,
			&graphqlAPI.Description,
			&graphqlAPI.Arguments,
			&graphqlAPI.ReturnType,
			&graphqlAPI.Examples,
			&graphqlAPI.CreatedAt,
			&graphqlAPI.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan GraphQL API: %w", err)
		}
		graphqlAPIs = append(graphqlAPIs, graphqlAPI)
	}

	if rows.Err() != nil {
		return nil, 0, fmt.Errorf("error iterating GraphQL APIs: %w", rows.Err())
	}

	// Get total count for pagination metadata
	countQuery := `SELECT COUNT(*) FROM graphql_apis WHERE id_project = $1`
	var totalCount int
	err = r.db.QueryRow(ctx, countQuery, projectID).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count GraphQL APIs: %w", err)
	}

	return graphqlAPIs, totalCount, nil
}

// BatchGetGraphQLAPIs retrieves multiple GraphQL APIs by their IDs using pgx.Batch
// Per Specs: use batch queries to prevent N+1 queries
func (r *GraphQLAPIRepository) BatchGetGraphQLAPIs(ctx context.Context, ids []int64) ([]*model.GraphQLAPI, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	if len(ids) == 0 {
		return []*model.GraphQLAPI{}, nil
	}

	// Use pgx.Batch for efficient bulk queries
	// Per Specs: batch queries prevent N+1 query problems
	batch := &pgx.Batch{}
	for _, id := range ids {
		batch.Queue(`
			SELECT id_graphql_api, id_project, id_user, name, type, description, arguments, return_type, examples, created_at, updated_at
			FROM graphql_apis
			WHERE id_graphql_api = $1
		`, id)
	}

	results := r.db.SendBatch(ctx, batch)
	defer results.Close()

	apis := make([]*model.GraphQLAPI, 0, len(ids))
	for i := 0; i < len(ids); i++ {
		var api model.GraphQLAPI
		if err := results.QueryRow().Scan(
			&api.IDGraphqlAPI,
			&api.IDProject,
			&api.IDUser,
			&api.Name,
			&api.Type,
			&api.Description,
			&api.Arguments,
			&api.ReturnType,
			&api.Examples,
			&api.CreatedAt,
			&api.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan GraphQL API batch result: %w", err)
		}
		apis = append(apis, &api)
	}

	return apis, nil
}
