-- ============================================================================
-- Performance Indexes for Production Readiness
-- Target: <50ms p95 for all queries under 100K+ rows
-- ============================================================================

-- Expression index for ABAC v3 attribute-based policy lookups (MISSING)
-- Covers: SELECT ... WHERE is_active=true AND policy_rule->>'attribute_name'=$1 AND policy_rule->>'attribute_value'=$2
CREATE INDEX IF NOT EXISTS idx_policies_attr_name_value
    ON policies ((policy_rule->>'attribute_name'), (policy_rule->>'attribute_value'))
    WHERE is_active = true;

-- Partial index for non-null descriptions (improves project search queries)
CREATE INDEX IF NOT EXISTS idx_projects_description_notnull
    ON projects(description)
    WHERE description IS NOT NULL AND description != '';

-- Covering index for public project listing with stats (LATERAL JOIN in FindAllWithStats)
CREATE INDEX IF NOT EXISTS idx_projects_public_created
    ON projects(id_user, is_public, created_at DESC)
    INCLUDE (name, slug, description, version)
    WHERE is_public = true;

-- Composite index for sorted API listing within a project
-- Covers ORDER BY created_at DESC with WHERE id_project = $1
CREATE INDEX IF NOT EXISTS idx_rest_apis_project_id_created
    ON rest_apis(id_project, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_graphql_apis_project_id_created
    ON graphql_apis(id_project, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_grpc_apis_project_id_created
    ON grpc_apis(id_project, created_at DESC);

-- Partial index for soft-deleted/long-tail user_policies cleanup queries
CREATE INDEX IF NOT EXISTS idx_user_policies_expired
    ON user_policies(expires_at)
    WHERE expires_at IS NOT NULL AND expires_at < CURRENT_TIMESTAMP;

-- GIN index on api_tags junction for tag-based filtering
CREATE INDEX IF NOT EXISTS idx_api_tags_composite
    ON api_tags(api_type, id_tag);

-- Covering index for user lookup by email (login hot path)
-- Uses unique name to avoid conflicts with existing idx_users_email
-- INCLUDE avoids extra heap lookup when selecting id_user, name, is_active
CREATE INDEX IF NOT EXISTS idx_users_email_covering
    ON users(email)
    INCLUDE (id_user, name, is_active);

COMMENT ON INDEX idx_policies_attr_name_value IS 'ABAC v3: fast attribute-based policy matching';
COMMENT ON INDEX idx_projects_public_created IS 'Covering index for public project listing with stats';
COMMENT ON INDEX idx_api_tags_composite IS 'Composite index for tag-based API filtering';

-- Update statistics
ANALYZE policies;
ANALYZE projects;
ANALYZE rest_apis;
ANALYZE graphql_apis;
ANALYZE grpc_apis;
ANALYZE user_policies;
ANALYZE api_tags;
ANALYZE users;
