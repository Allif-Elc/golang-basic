package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"golang-basic/api/internal/config"
	"golang-basic/api/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RestAPIRepository struct {
	db *pgxpool.Pool
}

func NewRestAPIRepository(db *pgxpool.Pool) *RestAPIRepository {
	return &RestAPIRepository{db: db}
}

func (r *RestAPIRepository) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, config.QueryTimeout)
}

// Create inserts a new REST API documentation into the database
func (r *RestAPIRepository) Create(ctx context.Context, req *model.CreateRestAPIRequest, userID int64) (*model.RestAPI, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	var headersBytes, pathParamsBytes, queryParamsBytes, requestBodyBytes, responsesBytes []byte
	var err error

	if req.Headers != nil {
		headersBytes, err = json.Marshal(req.Headers)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal headers: %w", err)
		}
	}

	if req.PathParams != nil {
		pathParamsBytes, err = json.Marshal(req.PathParams)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal path params: %w", err)
		}
	}

	if req.QueryParams != nil {
		queryParamsBytes, err = json.Marshal(req.QueryParams)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal query params: %w", err)
		}
	}

	if req.RequestBody != nil {
		requestBodyBytes, err = json.Marshal(req.RequestBody)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	if req.Responses != nil {
		responsesBytes, err = json.Marshal(req.Responses)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal responses: %w", err)
		}
	}

	query := `
		INSERT INTO rest_apis (id_project, id_user, name, description, method, endpoint, headers, path_params, query_params, request_body, responses, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW())
		RETURNING id_rest_api, created_at, updated_at
	`

	var restAPI model.RestAPI
	err = r.db.QueryRow(ctx, query,
		req.IDProject,
		userID,
		req.Name,
		req.Description,
		req.Method,
		req.Endpoint,
		headersBytes,
		pathParamsBytes,
		queryParamsBytes,
		requestBodyBytes,
		responsesBytes,
	).Scan(&restAPI.IDRestAPI, &restAPI.CreatedAt, &restAPI.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create REST API: %w", err)
	}

	restAPI.IDProject = req.IDProject
	restAPI.IDUser = userID
	restAPI.Name = req.Name
	restAPI.Description = req.Description
	restAPI.Method = req.Method
	restAPI.Endpoint = req.Endpoint
	restAPI.Headers = headersBytes
	restAPI.PathParams = pathParamsBytes
	restAPI.QueryParams = queryParamsBytes
	restAPI.RequestBody = requestBodyBytes
	restAPI.Responses = responsesBytes

	return &restAPI, nil
}

// FindByID retrieves a REST API by its ID
func (r *RestAPIRepository) FindByID(ctx context.Context, apiID int64) (model.RestAPI, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	query := `
		SELECT id_rest_api, id_project, id_user, name, description, method, endpoint, headers, path_params, query_params, request_body, responses, created_at, updated_at
		FROM rest_apis
		WHERE id_rest_api = $1
	`

	var restAPI model.RestAPI
	err := r.db.QueryRow(ctx, query, apiID).Scan(
		&restAPI.IDRestAPI,
		&restAPI.IDProject,
		&restAPI.IDUser,
		&restAPI.Name,
		&restAPI.Description,
		&restAPI.Method,
		&restAPI.Endpoint,
		&restAPI.Headers,
		&restAPI.PathParams,
		&restAPI.QueryParams,
		&restAPI.RequestBody,
		&restAPI.Responses,
		&restAPI.CreatedAt,
		&restAPI.UpdatedAt,
	)

	if err != nil {
		return model.RestAPI{}, fmt.Errorf("failed to find REST API: %w", err)
	}

	return restAPI, nil
}

// Update modifies an existing REST API documentation
func (r *RestAPIRepository) Update(ctx context.Context, apiID int64, req *model.UpdateRestAPIRequest) (*model.RestAPI, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	var name interface{} = nil
	var description interface{} = nil
	var method interface{} = nil
	var endpoint interface{} = nil
	var headersBytes []byte = nil
	var pathParamsBytes []byte = nil
	var queryParamsBytes []byte = nil
	var requestBodyBytes []byte = nil
	var responsesBytes []byte = nil
	var err error

	if req.Name != nil {
		name = *req.Name
	}
	if req.Description != nil {
		description = *req.Description
	}
	if req.Method != nil {
		method = *req.Method
	}
	if req.Endpoint != nil {
		endpoint = *req.Endpoint
	}
	if req.Headers != nil {
		headersBytes, err = json.Marshal(*req.Headers)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal headers: %w", err)
		}
	}
	if req.PathParams != nil {
		pathParamsBytes, err = json.Marshal(*req.PathParams)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal path params: %w", err)
		}
	}
	if req.QueryParams != nil {
		queryParamsBytes, err = json.Marshal(*req.QueryParams)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal query params: %w", err)
		}
	}
	if req.RequestBody != nil {
		requestBodyBytes, err = json.Marshal(*req.RequestBody)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
	}
	if req.Responses != nil {
		responsesBytes, err = json.Marshal(*req.Responses)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal responses: %w", err)
		}
	}

	query := `
		UPDATE rest_apis
		SET name = COALESCE($1, name),
		    description = COALESCE($2, description),
		    method = COALESCE($3, method),
		    endpoint = COALESCE($4, endpoint),
		    headers = COALESCE($5, headers),
		    path_params = COALESCE($6, path_params),
		    query_params = COALESCE($7, query_params),
		    request_body = COALESCE($8, request_body),
		    responses = COALESCE($9, responses),
		    updated_at = NOW()
		WHERE id_rest_api = $10
		RETURNING id_rest_api, id_project, id_user, name, description, method, endpoint, headers, path_params, query_params, request_body, responses, created_at, updated_at
	`

	var restAPI model.RestAPI
	err = r.db.QueryRow(ctx, query,
		name, description, method, endpoint,
		headersBytes, pathParamsBytes, queryParamsBytes, requestBodyBytes, responsesBytes,
		apiID,
	).Scan(
		&restAPI.IDRestAPI,
		&restAPI.IDProject,
		&restAPI.IDUser,
		&restAPI.Name,
		&restAPI.Description,
		&restAPI.Method,
		&restAPI.Endpoint,
		&restAPI.Headers,
		&restAPI.PathParams,
		&restAPI.QueryParams,
		&restAPI.RequestBody,
		&restAPI.Responses,
		&restAPI.CreatedAt,
		&restAPI.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update REST API: %w", err)
	}

	return &restAPI, nil
}

// Delete removes a REST API from the database and returns rows affected
func (r *RestAPIRepository) Delete(ctx context.Context, apiID int64) (int64, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	query := `DELETE FROM rest_apis WHERE id_rest_api = $1`

	result, err := r.db.Exec(ctx, query, apiID)
	if err != nil {
		return 0, fmt.Errorf("failed to delete REST API: %w", err)
	}

	rowsAffected := result.RowsAffected()
	return rowsAffected, nil
}

// List retrieves REST APIs with pagination and filters
// Required indexes:
//   CREATE INDEX idx_rest_apis_project ON rest_apis(id_project);
//   CREATE INDEX idx_rest_apis_method ON rest_apis(method);
//   CREATE INDEX idx_rest_apis_name ON rest_apis USING GIN (to_tsvector('english', name));
//   CREATE INDEX idx_rest_apis_description ON rest_apis USING GIN (to_tsvector('english', description));
//   CREATE INDEX idx_rest_apis_created_at ON rest_apis(created_at);
func (r *RestAPIRepository) List(ctx context.Context, req *model.ListRestAPIsRequest) (*model.PageResult[model.RestAPI], error) {
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

	if req.Method != "" {
		whereClause += fmt.Sprintf(" AND method = $%d", argNum)
		countArgs = append(countArgs, req.Method)
		queryArgs = append(queryArgs, req.Method)
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
		SELECT id_rest_api, id_project, id_user, name, description, method, endpoint, headers, path_params, query_params, request_body, responses, created_at, updated_at
		FROM rest_apis
		%s
		%s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderClause, argNum, argNum+1)

	rows, err := r.db.Query(ctx, query, queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to query REST APIs: %w", err)
	}
	defer rows.Close()

	restAPIs := make([]model.RestAPI, 0, req.Limit)
	for rows.Next() {
		var restAPI model.RestAPI
		err := rows.Scan(
			&restAPI.IDRestAPI,
			&restAPI.IDProject,
			&restAPI.IDUser,
			&restAPI.Name,
			&restAPI.Description,
			&restAPI.Method,
			&restAPI.Endpoint,
			&restAPI.Headers,
			&restAPI.PathParams,
			&restAPI.QueryParams,
			&restAPI.RequestBody,
			&restAPI.Responses,
			&restAPI.CreatedAt,
			&restAPI.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan REST API: %w", err)
		}
		restAPIs = append(restAPIs, restAPI)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("error iterating REST APIs: %w", rows.Err())
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM rest_apis %s", whereClause)
	var totalCount int64
	err = r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count REST APIs: %w", err)
	}

	result := &model.PageResult[model.RestAPI]{
		Data:     restAPIs,
		Page:     req.Page,
		Size:     req.Limit,
		StartRow: offset + 1,
		EndRow:   offset + len(restAPIs),
	}

	return result, nil
}

// FindByProjectID retrieves all REST APIs for a specific project
func (r *RestAPIRepository) FindByProjectID(ctx context.Context, projectID int64, req *model.ListRestAPIsRequest) (*model.PageResult[model.RestAPI], error) {
	// Set the project filter
	req.IDProject = &projectID
	return r.List(ctx, req)
}
