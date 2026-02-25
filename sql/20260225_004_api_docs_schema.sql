-- ============================================================================
-- API Documentation Platform Schema
-- Add to existing ABAC database
-- PostgreSQL 17+
-- ============================================================================

-- ============================================================================
-- Projects table
-- ============================================================================
CREATE TABLE IF NOT EXISTS projects (
    id_project BIGSERIAL PRIMARY KEY,
    id_user BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    version VARCHAR(50) DEFAULT '1.0',
    is_public BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (id_user) REFERENCES users(id_user) ON DELETE CASCADE
);

-- ============================================================================
-- REST APIs table
-- ============================================================================
CREATE TABLE IF NOT EXISTS rest_apis (
    id_rest_api BIGSERIAL PRIMARY KEY,
    id_project BIGINT NOT NULL,
    id_user BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    method VARCHAR(10) NOT NULL CHECK (method IN ('GET', 'POST', 'PUT', 'DELETE', 'PATCH')),
    endpoint VARCHAR(500) NOT NULL,
    headers JSONB,
    path_params JSONB,
    query_params JSONB,
    request_body JSONB,
    responses JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (id_project) REFERENCES projects(id_project) ON DELETE CASCADE,
    FOREIGN KEY (id_user) REFERENCES users(id_user) ON DELETE SET NULL
);

-- ============================================================================
-- GraphQL APIs table
-- ============================================================================
CREATE TABLE IF NOT EXISTS graphql_apis (
    id_graphql_api BIGSERIAL PRIMARY KEY,
    id_project BIGINT NOT NULL,
    id_user BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('query', 'mutation', 'subscription')),
    description TEXT,
    arguments JSONB,
    return_type TEXT,
    examples JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (id_project) REFERENCES projects(id_project) ON DELETE CASCADE,
    FOREIGN KEY (id_user) REFERENCES users(id_user) ON DELETE SET NULL
);

-- ============================================================================
-- gRPC APIs table
-- ============================================================================
CREATE TABLE IF NOT EXISTS grpc_apis (
    id_grpc_api BIGSERIAL PRIMARY KEY,
    id_project BIGINT NOT NULL,
    id_user BIGINT NOT NULL,
    service_name VARCHAR(255) NOT NULL,
    method_name VARCHAR(255) NOT NULL,
    description TEXT,
    request_message JSONB,
    response_message JSONB,
    proto_definition TEXT,
    examples JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (id_project) REFERENCES projects(id_project) ON DELETE CASCADE,
    FOREIGN KEY (id_user) REFERENCES users(id_user) ON DELETE SET NULL
);

-- ============================================================================
-- Tags table
-- ============================================================================
CREATE TABLE IF NOT EXISTS tags (
    id_tag BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    color VARCHAR(7) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- API Tags junction table (many-to-many)
-- ============================================================================
CREATE TABLE IF NOT EXISTS api_tags (
    id_api BIGINT NOT NULL,
    api_type VARCHAR(20) NOT NULL CHECK (api_type IN ('rest', 'graphql', 'grpc')),
    id_tag BIGINT NOT NULL,
    PRIMARY KEY (id_api, api_type, id_tag),
    FOREIGN KEY (id_tag) REFERENCES tags(id_tag) ON DELETE CASCADE
);

-- ============================================================================
-- Performance Indexes (CRITICAL for performance)
-- ============================================================================

-- Projects indexes
CREATE INDEX IF NOT EXISTS idx_projects_id_user ON projects(id_user);
CREATE INDEX IF NOT EXISTS idx_projects_slug ON projects(slug);
CREATE INDEX IF NOT EXISTS idx_projects_is_public ON projects(is_public);
CREATE INDEX IF NOT EXISTS idx_projects_created_at ON projects(created_at DESC);

-- REST APIs indexes
CREATE INDEX IF NOT EXISTS idx_rest_apis_id_project ON rest_apis(id_project);
CREATE INDEX IF NOT EXISTS idx_rest_apis_method ON rest_apis(method);
-- Composite index for project + method filtering (order: selective → equality → range)
CREATE INDEX IF NOT EXISTS idx_rest_apis_project_method ON rest_apis(id_project, method);
CREATE INDEX IF NOT EXISTS idx_rest_apis_headers ON rest_apis USING GIN (headers);
CREATE INDEX IF NOT EXISTS idx_rest_apis_path_params ON rest_apis USING GIN (path_params);
CREATE INDEX IF NOT EXISTS idx_rest_apis_query_params ON rest_apis USING GIN (query_params);
CREATE INDEX IF NOT EXISTS idx_rest_apis_request_body ON rest_apis USING GIN (request_body);
CREATE INDEX IF NOT EXISTS idx_rest_apis_responses ON rest_apis USING GIN (responses);

-- GraphQL APIs indexes
CREATE INDEX IF NOT EXISTS idx_graphql_apis_id_project ON graphql_apis(id_project);
CREATE INDEX IF NOT EXISTS idx_graphql_apis_type ON graphql_apis(type);
-- Composite index for project + type filtering
CREATE INDEX IF NOT EXISTS idx_graphql_apis_project_type ON graphql_apis(id_project, type);
CREATE INDEX IF NOT EXISTS idx_graphql_apis_arguments ON graphql_apis USING GIN (arguments);
CREATE INDEX IF NOT EXISTS idx_graphql_apis_examples ON graphql_apis USING GIN (examples);

-- gRPC APIs indexes
CREATE INDEX IF NOT EXISTS idx_grpc_apis_id_project ON grpc_apis(id_project);
CREATE INDEX IF NOT EXISTS idx_grpc_apis_service_name ON grpc_apis(service_name);
-- Composite index for project + service filtering
CREATE INDEX IF NOT EXISTS idx_grpc_apis_project_service ON grpc_apis(id_project, service_name);
CREATE INDEX IF NOT EXISTS idx_grpc_apis_request_message ON grpc_apis USING GIN (request_message);
CREATE INDEX IF NOT EXISTS idx_grpc_apis_response_message ON grpc_apis USING GIN (response_message);
CREATE INDEX IF NOT EXISTS idx_grpc_apis_examples ON grpc_apis USING GIN (examples);

-- API tags indexes
CREATE INDEX IF NOT EXISTS idx_api_tags_id_api ON api_tags(id_api);
CREATE INDEX IF NOT EXISTS idx_api_tags_api_type ON api_tags(api_type);
CREATE INDEX IF NOT EXISTS idx_api_tags_id_tag ON api_tags(id_tag);

-- Full-text search indexes (for searching documentation)
CREATE INDEX IF NOT EXISTS idx_rest_apis_fts ON rest_apis USING GIN (to_tsvector('english', name || ' ' || COALESCE(description, '')));
CREATE INDEX IF NOT EXISTS idx_graphql_apis_fts ON graphql_apis USING GIN (to_tsvector('english', name || ' ' || COALESCE(description, '')));
CREATE INDEX IF NOT EXISTS idx_grpc_apis_fts ON grpc_apis USING GIN (to_tsvector('english', service_name || ' ' || method_name || ' ' || COALESCE(description, '')));
CREATE INDEX IF NOT EXISTS idx_projects_fts ON projects USING GIN (to_tsvector('english', name || ' ' || COALESCE(description, '')));

-- ============================================================================
-- Triggers for updated_at timestamp
-- ============================================================================
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_projects_updated_at BEFORE UPDATE ON projects
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_rest_apis_updated_at BEFORE UPDATE ON rest_apis
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_graphql_apis_updated_at BEFORE UPDATE ON graphql_apis
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_grpc_apis_updated_at BEFORE UPDATE ON grpc_apis
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- ABAC Policies for API Documentation
-- ============================================================================

-- Policy: Users can manage their own projects
INSERT INTO policies (name, policy_rule, is_active, created_at) VALUES
('api_docs_own_project', '{
    "resource": "api_docs_project",
    "action": ["create", "read", "update", "delete"],
    "condition": "id_user == project_owner_id"
}', true, NOW());

-- Policy: Developers can create and edit any documentation
INSERT INTO policies (name, policy_rule, is_active, created_at) VALUES
('api_docs_developer', '{
    "role": "developer",
    "resource": "api_docs_*",
    "action": ["create", "read", "update"]
}', true, NOW());

-- Policy: Technical writers can read and update documentation
INSERT INTO policies (name, policy_rule, is_active, created_at) VALUES
('api_docs_technical_writer', '{
    "role": "technical_writer",
    "resource": "api_docs_*",
    "action": ["read", "update"]
}', true, NOW());

-- Policy: Viewers can only read documentation
INSERT INTO policies (name, policy_rule, is_active, created_at) VALUES
('api_docs_viewer', '{
    "role": "viewer",
    "resource": "api_docs_*",
    "action": ["read"]
}', true, NOW());

-- Policy: Public projects can be viewed by anyone
INSERT INTO policies (name, policy_rule, is_active, created_at) VALUES
('api_docs_public_view', '{
    "resource": "api_docs_project",
    "action": "read",
    "condition": "project.is_public == true"
}', true, NOW());

-- Policy: Admins can do everything
INSERT INTO policies (name, policy_rule, is_active, created_at) VALUES
('api_docs_admin', '{
    "role": "admin",
    "resource": "api_docs_*",
    "action": ["*"]
}', true, NOW());

-- ============================================================================
-- Sample Data (for testing)
-- ============================================================================

-- Sample tags
INSERT INTO tags (name, color) VALUES
('Authentication', '#FF5733'),
('User Management', '#33FF57'),
('Data Processing', '#3357FF'),
('Reporting', '#F333FF'),
('Utilities', '#FF33A8');

-- Sample project (if user with id_user = 1 exists)
-- Uncomment to insert sample data
/*
INSERT INTO projects (id_project, id_user, name, slug, description, version, is_public) VALUES
(1, 1, 'User Service API', 'user-service-api', 'API for managing users and authentication', '1.0', true);

-- Sample REST API
INSERT INTO rest_apis (id_rest_api, id_project, id_user, name, description, method, endpoint, headers, path_params, query_params, request_body, responses) VALUES
(1, 1, 1, 'Create User', 'Create a new user account', 'POST', '/api/users',
    '[{"name": "Content-Type", "description": "Request content type", "required": true, "example": "application/json"}]'::jsonb,
    '[]'::jsonb,
    '[]'::jsonb,
    '{"type": "object", "properties": {"name": {"type": "string"}, "email": {"type": "string", "format": "email"}}, "required": ["name", "email"]}'::jsonb,
    '{"200": {"status_code": 200, "description": "User created successfully", "body": {"id_user": 1, "name": "John Doe", "email": "john@example.com"}}}'::jsonb
);
*/
