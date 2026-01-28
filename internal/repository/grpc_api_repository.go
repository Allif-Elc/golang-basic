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

type GrpcAPIRepository struct {
	db *pgxpool.Pool
}

func NewGrpcAPIRepository(db *pgxpool.Pool) *GrpcAPIRepository {
	return &GrpcAPIRepository{db: db}
}

func (r *GrpcAPIRepository) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, config.QueryTimeout)
}

// Create inserts a new gRPC API documentation into the database
// Per Specs: explicit columns (no SELECT *), context timeout, defer Close()
func (r *GrpcAPIRepository) Create(ctx context.Context, req *model.CreateGrpcAPIRequest, userID int64) (*model.GrpcAPI, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	var requestMessageBytes, responseMessageBytes, examplesBytes []byte
	var err error

	if req.RequestMessage != nil {
		requestMessageBytes, err = json.Marshal(req.RequestMessage)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request message: %w", err)
		}
	}

	if req.ResponseMessage != nil {
		responseMessageBytes, err = json.Marshal(req.ResponseMessage)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal response message: %w", err)
		}
	}

	if req.Examples != nil {
		examplesBytes, err = json.Marshal(req.Examples)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal examples: %w", err)
		}
	}

	query := `
		INSERT INTO grpc_apis (id_project, id_user, service_name, method_name, description, request_message, response_message, proto_definition, examples, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		RETURNING id_grpc_api, created_at, updated_at
	`

	var grpcAPI model.GrpcAPI
	err = r.db.QueryRow(ctx, query,
		req.IDProject,
		userID,
		req.ServiceName,
		req.MethodName,
		req.Description,
		requestMessageBytes,
		responseMessageBytes,
		req.ProtoDefinition,
		examplesBytes,
	).Scan(&grpcAPI.IDGrpcAPI, &grpcAPI.CreatedAt, &grpcAPI.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC API: %w", err)
	}

	grpcAPI.IDProject = req.IDProject
	grpcAPI.IDUser = userID
	grpcAPI.ServiceName = req.ServiceName
	grpcAPI.MethodName = req.MethodName
	grpcAPI.Description = req.Description
	grpcAPI.RequestMessage = requestMessageBytes
	grpcAPI.ResponseMessage = responseMessageBytes
	grpcAPI.ProtoDefinition = req.ProtoDefinition
	grpcAPI.Examples = examplesBytes

	return &grpcAPI, nil
}

// FindByID retrieves a gRPC API by its ID
// Per Specs: explicit columns, QueryRow for single row
func (r *GrpcAPIRepository) FindByID(ctx context.Context, apiID int64) (model.GrpcAPI, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	query := `
		SELECT id_grpc_api, id_project, id_user, service_name, method_name, description, request_message, response_message, proto_definition, examples, created_at, updated_at
		FROM grpc_apis
		WHERE id_grpc_api = $1
	`

	var grpcAPI model.GrpcAPI
	err := r.db.QueryRow(ctx, query, apiID).Scan(
		&grpcAPI.IDGrpcAPI,
		&grpcAPI.IDProject,
		&grpcAPI.IDUser,
		&grpcAPI.ServiceName,
		&grpcAPI.MethodName,
		&grpcAPI.Description,
		&grpcAPI.RequestMessage,
		&grpcAPI.ResponseMessage,
		&grpcAPI.ProtoDefinition,
		&grpcAPI.Examples,
		&grpcAPI.CreatedAt,
		&grpcAPI.UpdatedAt,
	)

	if err != nil {
		return model.GrpcAPI{}, fmt.Errorf("failed to find gRPC API: %w", err)
	}

	return grpcAPI, nil
}

// Update modifies an existing gRPC API documentation
func (r *GrpcAPIRepository) Update(ctx context.Context, apiID int64, req *model.UpdateGrpcAPIRequest) (*model.GrpcAPI, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	var serviceName interface{} = nil
	var methodName interface{} = nil
	var description interface{} = nil
	var protoDefinition interface{} = nil
	var requestMessageBytes []byte = nil
	var responseMessageBytes []byte = nil
	var examplesBytes []byte = nil
	var err error

	if req.ServiceName != nil {
		serviceName = *req.ServiceName
	}
	if req.MethodName != nil {
		methodName = *req.MethodName
	}
	if req.Description != nil {
		description = *req.Description
	}
	if req.ProtoDefinition != nil {
		protoDefinition = *req.ProtoDefinition
	}
	if req.RequestMessage != nil {
		requestMessageBytes, err = json.Marshal(*req.RequestMessage)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request message: %w", err)
		}
	}
	if req.ResponseMessage != nil {
		responseMessageBytes, err = json.Marshal(*req.ResponseMessage)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal response message: %w", err)
		}
	}
	if req.Examples != nil {
		examplesBytes, err = json.Marshal(*req.Examples)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal examples: %w", err)
		}
	}

	query := `
		UPDATE grpc_apis
		SET service_name = COALESCE($1, service_name),
		    method_name = COALESCE($2, method_name),
		    description = COALESCE($3, description),
		    request_message = COALESCE($4, request_message),
		    response_message = COALESCE($5, response_message),
		    proto_definition = COALESCE($6, proto_definition),
		    examples = COALESCE($7, examples),
		    updated_at = NOW()
		WHERE id_grpc_api = $8
		RETURNING id_grpc_api, id_project, id_user, service_name, method_name, description, request_message, response_message, proto_definition, examples, created_at, updated_at
	`

	var grpcAPI model.GrpcAPI
	err = r.db.QueryRow(ctx, query,
		serviceName, methodName, description,
		requestMessageBytes, responseMessageBytes, protoDefinition, examplesBytes,
		apiID,
	).Scan(
		&grpcAPI.IDGrpcAPI,
		&grpcAPI.IDProject,
		&grpcAPI.IDUser,
		&grpcAPI.ServiceName,
		&grpcAPI.MethodName,
		&grpcAPI.Description,
		&grpcAPI.RequestMessage,
		&grpcAPI.ResponseMessage,
		&grpcAPI.ProtoDefinition,
		&grpcAPI.Examples,
		&grpcAPI.CreatedAt,
		&grpcAPI.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update gRPC API: %w", err)
	}

	return &grpcAPI, nil
}

// Delete removes a gRPC API from the database and returns rows affected
func (r *GrpcAPIRepository) Delete(ctx context.Context, apiID int64) (int64, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	query := `DELETE FROM grpc_apis WHERE id_grpc_api = $1`

	result, err := r.db.Exec(ctx, query, apiID)
	if err != nil {
		return 0, fmt.Errorf("failed to delete gRPC API: %w", err)
	}

	rowsAffected := result.RowsAffected()
	return rowsAffected, nil
}

// List retrieves gRPC APIs with pagination and filters
// Per Specs: explicit columns, indexed fields (id_project, service_name)
func (r *GrpcAPIRepository) List(ctx context.Context, req *model.ListGrpcAPIsRequest) (*model.PageResult[model.GrpcAPI], error) {
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

	if req.ServiceName != "" {
		whereClause += fmt.Sprintf(" AND service_name = $%d", argNum)
		countArgs = append(countArgs, req.ServiceName)
		queryArgs = append(queryArgs, req.ServiceName)
		argNum++
	}

	if req.Search != "" {
		whereClause += fmt.Sprintf(" AND (service_name ILIKE $%d OR method_name ILIKE $%d OR description ILIKE $%d)", argNum, argNum+1, argNum+2)
		searchPattern := "%" + req.Search + "%"
		countArgs = append(countArgs, searchPattern, searchPattern, searchPattern)
		queryArgs = append(queryArgs, searchPattern, searchPattern, searchPattern)
		argNum += 3
	}

	orderClause := "ORDER BY created_at DESC"

	queryArgs = append(queryArgs, req.Limit, offset)
	query := fmt.Sprintf(`
		SELECT id_grpc_api, id_project, id_user, service_name, method_name, description, request_message, response_message, proto_definition, examples, created_at, updated_at
		FROM grpc_apis
		%s
		%s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderClause, argNum, argNum+1)

	rows, err := r.db.Query(ctx, query, queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to query gRPC APIs: %w", err)
	}
	defer rows.Close()

	grpcAPIs := make([]model.GrpcAPI, 0, req.Limit)
	for rows.Next() {
		var grpcAPI model.GrpcAPI
		err := rows.Scan(
			&grpcAPI.IDGrpcAPI,
			&grpcAPI.IDProject,
			&grpcAPI.IDUser,
			&grpcAPI.ServiceName,
			&grpcAPI.MethodName,
			&grpcAPI.Description,
			&grpcAPI.RequestMessage,
			&grpcAPI.ResponseMessage,
			&grpcAPI.ProtoDefinition,
			&grpcAPI.Examples,
			&grpcAPI.CreatedAt,
			&grpcAPI.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan gRPC API: %w", err)
		}
		grpcAPIs = append(grpcAPIs, grpcAPI)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("error iterating gRPC APIs: %w", rows.Err())
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM grpc_apis %s", whereClause)
	var totalCount int64
	err = r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count gRPC APIs: %w", err)
	}

	result := &model.PageResult[model.GrpcAPI]{
		Data:     grpcAPIs,
		Page:     req.Page,
		Size:     req.Limit,
		StartRow: offset + 1,
		EndRow:   offset + len(grpcAPIs),
	}

	return result, nil
}

// FindByProjectID retrieves all gRPC APIs for a specific project
func (r *GrpcAPIRepository) FindByProjectID(ctx context.Context, projectID int64, req *model.ListGrpcAPIsRequest) (*model.PageResult[model.GrpcAPI], error) {
	req.IDProject = &projectID
	return r.List(ctx, req)
}

// BatchGetGrpcAPIs retrieves multiple gRPC APIs by their IDs using pgx.Batch
// Per Specs: use batch queries to prevent N+1 queries
func (r *GrpcAPIRepository) BatchGetGrpcAPIs(ctx context.Context, ids []int64) ([]*model.GrpcAPI, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	if len(ids) == 0 {
		return []*model.GrpcAPI{}, nil
	}

	// Use pgx.Batch for efficient bulk queries
	// Per Specs: batch queries prevent N+1 query problems
	batch := &pgx.Batch{}
	for _, id := range ids {
		batch.Queue(`
			SELECT id_grpc_api, id_project, id_user, service_name, method_name, description, request_message, response_message, proto_definition, examples, created_at, updated_at
			FROM grpc_apis
			WHERE id_grpc_api = $1
		`, id)
	}

	results := r.db.SendBatch(ctx, batch)
	defer results.Close()

	apis := make([]*model.GrpcAPI, 0, len(ids))
	for i := 0; i < len(ids); i++ {
		var api model.GrpcAPI
		if err := results.QueryRow().Scan(
			&api.IDGrpcAPI,
			&api.IDProject,
			&api.IDUser,
			&api.ServiceName,
			&api.MethodName,
			&api.Description,
			&api.RequestMessage,
			&api.ResponseMessage,
			&api.ProtoDefinition,
			&api.Examples,
			&api.CreatedAt,
			&api.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan gRPC API batch result: %w", err)
		}
		apis = append(apis, &api)
	}

	return apis, nil
}
