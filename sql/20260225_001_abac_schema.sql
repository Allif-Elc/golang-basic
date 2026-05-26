-- ============================================================================
-- ABAC PostgreSQL 17 Schema with Performance Optimizations
-- Target: <50ms authorization checks with 1M+ rows
-- Updated: 2025-01-17 - Consistent with model folder naming conventions
-- ============================================================================

-- ============================================================================
-- TABLES
-- ============================================================================

-- Attributes table: defines attribute types (role, department, etc.)
CREATE TABLE IF NOT EXISTS attributes (
    id_attribute BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Users table: basic user information
CREATE TABLE IF NOT EXISTS users (
    id_user BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Profiles table: user profile information
CREATE TABLE IF NOT EXISTS profiles (
    id_profile BIGSERIAL PRIMARY KEY,
    id_user BIGINT NOT NULL UNIQUE,
    age SMALLINT,
    gender VARCHAR(6),
    bio TEXT,
    phonenumber VARCHAR(100),
    website TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (id_user) REFERENCES users(id_user) ON DELETE CASCADE
);

-- User attributes: many-to-many relationship between users and attributes
-- Example: (id_user: 1, id_attribute: role_id_attribute, value: "HR")
CREATE TABLE IF NOT EXISTS user_attributes (
    id_user_attribute BIGSERIAL PRIMARY KEY,
    id_user BIGINT NOT NULL,
    id_attribute BIGINT NOT NULL,
    value VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(id_user, id_attribute, value),
    FOREIGN KEY (id_user) REFERENCES users(id_user) ON DELETE CASCADE,
    FOREIGN KEY (id_attribute) REFERENCES attributes(id_attribute) ON DELETE CASCADE
);

-- Resources: protected resources in the system
CREATE TABLE IF NOT EXISTS resources (
    id_resource BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Permissions: granular permissions (optional, can be derived from policies)
CREATE TABLE IF NOT EXISTS permissions (
    id_permission BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Policies: access control rules stored as JSONB
-- Example: {"role": "HR", "resource": "employee_records", "action": "read"}
CREATE TABLE IF NOT EXISTS policies (
    id_policy BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    policy_rule JSONB NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Audit logs: track all authorization decisions for compliance
CREATE TABLE IF NOT EXISTS audit_logs (
    id_audit_log BIGSERIAL PRIMARY KEY,
    id_user BIGINT NOT NULL,
    resource VARCHAR(255) NOT NULL,
    action VARCHAR(100) NOT NULL,
    allowed BOOLEAN NOT NULL,
    reason TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (id_user) REFERENCES users(id_user) ON DELETE SET NULL
);

-- ============================================================================
-- INDEXES (Critical for <50ms Performance)
-- ============================================================================

-- Users table indexes
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_name ON users(name);
CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active) WHERE is_active = false;

-- Profiles table indexes
CREATE INDEX IF NOT EXISTS idx_profiles_id_user ON profiles(id_user);
CREATE INDEX IF NOT EXISTS idx_profiles_gender ON profiles(gender) WHERE gender IS NOT NULL;

-- User attributes indexes (CRITICAL for authorization performance)
-- Used by: GetUserRoles()
CREATE INDEX IF NOT EXISTS idx_user_attributes_id_user
ON user_attributes(id_user);

CREATE INDEX IF NOT EXISTS idx_user_attributes_id_attribute
ON user_attributes(id_attribute);

-- Composite index for role lookup (covers most common query)
-- Used by: SELECT ua.value FROM user_attributes ua JOIN attributes a ON ua.id_attribute = a.id WHERE ua.id_user = ? AND a.name = 'role'
CREATE INDEX IF NOT EXISTS idx_user_attributes_user_attr_value
ON user_attributes(id_user, id_attribute, value);

-- Covering index that includes value (eliminates table lookup for sorting/filtering)
CREATE INDEX IF NOT EXISTS idx_user_attributes_attr_value
ON user_attributes(id_attribute, value);

-- GIN index for JSONB queries on policies (CRITICAL)
-- This enables fast policy_rule->>'resource' queries
CREATE INDEX IF NOT EXISTS idx_policies_policy_rule_gin
ON policies USING GIN (policy_rule);

-- Partial index for active policies only (reduces index size, improves query performance)
CREATE INDEX IF NOT EXISTS idx_policies_active_gin
ON policies USING GIN (policy_rule)
WHERE is_active = true;

-- Regular index for policy lookups
CREATE INDEX IF NOT EXISTS idx_policies_name
ON policies(name) WHERE is_active = true;

CREATE INDEX IF NOT EXISTS idx_policies_is_active
ON policies(is_active) WHERE is_active = false;

-- JSONB expression indexes for authorization queries
CREATE INDEX IF NOT EXISTS idx_policies_rule_resource_exact
ON policies ((policy_rule->>'resource'));

CREATE INDEX IF NOT EXISTS idx_policies_rule_role
ON policies ((policy_rule->>'role')) WHERE is_active = true;

-- Index for attribute lookups by name
CREATE INDEX IF NOT EXISTS idx_attributes_name
ON attributes(name);

-- Resources table indexes
CREATE INDEX IF NOT EXISTS idx_resources_name
ON resources(name);

-- Permissions table indexes
CREATE INDEX IF NOT EXISTS idx_permissions_name
ON permissions(name);

-- Audit log indexes for compliance queries (CRITICAL)
CREATE INDEX IF NOT EXISTS idx_audit_logs_id_user
ON audit_logs(id_user);

CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at
ON audit_logs(created_at DESC);

-- Composite index for audit queries (user + time range)
-- Used by compliance reports
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_time
ON audit_logs(id_user, created_at DESC);

-- Composite index for audit filtering (resource + time)
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource_time
ON audit_logs(resource, created_at DESC);

-- Composite index for audit filtering (action + time)
CREATE INDEX IF NOT EXISTS idx_audit_logs_action_time
ON audit_logs(action, created_at DESC);

-- Composite index for audit filtering (allowed + time)
CREATE INDEX IF NOT EXISTS idx_audit_logs_allowed_time
ON audit_logs(allowed, created_at DESC);

-- Full-text search index on audit_logs for searching reasons
CREATE INDEX IF NOT EXISTS idx_audit_logs_reason_fts
ON audit_logs USING GIN (to_tsvector('english', COALESCE(reason, '')));

-- ============================================================================
-- TRIGGERS (for automatic timestamp updates)
-- ============================================================================

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Trigger for users table
CREATE TRIGGER update_users_updated_at
BEFORE UPDATE ON users
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Trigger for policies table
CREATE TRIGGER update_policies_updated_at
BEFORE UPDATE ON policies
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Trigger for profiles table
CREATE TRIGGER update_profiles_updated_at
BEFORE UPDATE ON profiles
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- AUTOVACUUM SETTINGS (Optimal for High-Write Tables)
-- ============================================================================

-- Analyze tables for query planner statistics
ANALYZE attributes;
ANALYZE users;
ANALYZE profiles;
ANALYZE user_attributes;
ANALYZE resources;
ANALYZE permissions;
ANALYZE policies;
ANALYZE audit_logs;

-- Set autovacuum for audit_logs (high-write table)
-- Scale factor of 0.1 means autovacuum runs when 10% of rows are dead/modified
ALTER TABLE audit_logs SET (autovacuum_vacuum_scale_factor = 0.1);
ALTER TABLE audit_logs SET (autovacuum_analyze_scale_factor = 0.05);

-- Set autovacuum for user_attributes (frequent updates)
ALTER TABLE user_attributes SET (autovacuum_vacuum_scale_factor = 0.05);
ALTER TABLE user_attributes SET (autovacuum_analyze_scale_factor = 0.02);

-- Set autovovacuum for policies (occasional updates)
ALTER TABLE policies SET (autovacuum_vacuum_scale_factor = 0.1);
ALTER TABLE policies SET (autovacuum_analyze_scale_factor = 0.05);

-- ============================================================================
-- SAMPLE DATA (for testing)
-- ============================================================================

-- Insert attribute types
INSERT INTO attributes (id_attribute, name, description, type, enum_values) VALUES
(1, 'role', 'User role attribute', 'enum', ARRAY['admin','project_manager','tech_lead','developer','technical_writer','viewer']),
(2, 'department', 'Department attribute', 'enum', ARRAY['engineering','product','hr','finance'])
ON CONFLICT (name) DO NOTHING;

-- Insert sample users
INSERT INTO users (id_user, name, email, password_hash, is_active) VALUES
(1, 'Alice Project Manager', 'alice@company.com', '$argon2id$v=19$m=65536,t=3,p=2$pSIkdXLvwu4TBZnviQUd5g$3USJSHpTUpRmjGiKty7PPStQDgvP6NmJeIgIxCnfDHI', true),
(2, 'Bob Tech Lead', 'bob@company.com', '$argon2id$v=19$m=65536,t=3,p=2$pSIkdXLvwu4TBZnviQUd5g$3USJSHpTUpRmjGiKty7PPStQDgvP6NmJeIgIxCnfDHI', true),
(3, 'Charlie Developer', 'charlie@company.com', '$argon2id$v=19$m=65536,t=3,p=2$pSIkdXLvwu4TBZnviQUd5g$3USJSHpTUpRmjGiKty7PPStQDgvP6NmJeIgIxCnfDHI', true),
(4, 'David Developer', 'david@company.com', '$argon2id$v=19$m=65536,t=3,p=2$pSIkdXLvwu4TBZnviQUd5g$3USJSHpTUpRmjGiKty7PPStQDgvP6NmJeIgIxCnfDHI', true),
(5, 'Eve Technical Writer', 'eve@company.com', '$argon2id$v=19$m=65536,t=3,p=2$pSIkdXLvwu4TBZnviQUd5g$3USJSHpTUpRmjGiKty7PPStQDgvP6NmJeIgIxCnfDHI', true),
(6, 'Frank Viewer', 'frank@company.com', '$argon2id$v=19$m=65536,t=3,p=2$pSIkdXLvwu4TBZnviQUd5g$3USJSHpTUpRmjGiKty7PPStQDgvP6NmJeIgIxCnfDHI', true),
(7, 'Grace Admin', 'grace@company.com', '$argon2id$v=19$m=65536,t=3,p=2$pSIkdXLvwu4TBZnviQUd5g$3USJSHpTUpRmjGiKty7PPStQDgvP6NmJeIgIxCnfDHI', true),
(8, 'Henry Guest', 'henry@company.com', '$argon2id$v=19$m=65536,t=3,p=2$pSIkdXLvwu4TBZnviQUd5g$3USJSHpTUpRmjGiKty7PPStQDgvP6NmJeIgIxCnfDHI', true)
ON CONFLICT (email) DO NOTHING;

-- Insert sample profiles (now with phonenumber and website)
INSERT INTO profiles (id_profile, id_user, age, gender, bio, phonenumber, website) VALUES
(1, 1, 35, 'female', 'Alice the Project Manager', '+62-812-3456-7890', 'https://alice.dev'),
(2, 2, 32, 'male', 'Bob the Tech Lead', '+62-813-3456-7890', 'https://bob.dev'),
(3, 3, 28, 'male', 'Charlie the Backend Developer', '+62-814-3456-7890', 'https://charlie.dev'),
(4, 4, 26, 'male', 'David the Frontend Developer', '+62-815-3456-7890', 'https://david.dev'),
(5, 5, 30, 'female', 'Eve the Technical Writer', '+62-816-3456-7890', 'https://eve.dev'),
(6, 6, 25, 'male', 'Frank the Read-only Viewer', '+62-817-3456-7890', 'https://frank.dev'),
(7, 7, 40, 'female', 'Grace the System Admin', '+62-818-3456-7890', 'https://grace.dev'),
(8, 8, 24, 'male', 'Henry the Guest User', '+62-819-3456-7890', 'https://henry.dev')
ON CONFLICT (id_user) DO NOTHING;

-- Assign role attributes to users
INSERT INTO user_attributes (id_user, id_attribute, value) VALUES
(1, 1, 'project_manager'),
(2, 1, 'tech_lead'),
(3, 1, 'developer'),
(4, 1, 'developer'),
(5, 1, 'technical_writer'),
(6, 1, 'viewer'),
(7, 1, 'admin'),
(8, 1, 'viewer')
ON CONFLICT (id_user, id_attribute, value) DO NOTHING;

-- Assign department attributes to users (ABAC v3 multi-attribute)
INSERT INTO user_attributes (id_user, id_attribute, value) VALUES
(1, 2, 'product'),          -- alice: product dept
(2, 2, 'engineering'),      -- bob: engineering
(3, 2, 'engineering'),      -- charlie: engineering
(4, 2, 'engineering'),      -- david: engineering
(5, 2, 'product'),          -- eve: product
(6, 2, 'engineering'),      -- frank: engineering
(7, 2, 'engineering'),      -- grace: engineering
(8, 2, 'hr')                -- henry: hr
ON CONFLICT (id_user, id_attribute, value) DO NOTHING;

-- Insert resources (with resource_type)
INSERT INTO resources (id_resource, name, description, resource_type) VALUES
(1, 'projects', 'Project management resources', 'api'),
(2, 'api_docs_project', 'API Documentation Projects', 'api'),
(3, 'api_docs_rest_api', 'REST API Documentation', 'api'),
(4, 'api_docs_graphql_api', 'GraphQL API Documentation', 'api'),
(5, 'api_docs_grpc_api', 'gRPC API Documentation', 'api')
ON CONFLICT (name) DO NOTHING;

-- Insert permissions (with effect, actions, condition)
INSERT INTO permissions (id_permission, name, description, effect, actions, condition) VALUES
(1, 'create', 'Create new resources', 'allow', ARRAY['create'], NULL),
(2, 'read', 'View resources', 'allow', ARRAY['read'], NULL),
(3, 'update', 'Edit resources', 'allow', ARRAY['update'], NULL),
(4, 'delete', 'Delete resources', 'allow', ARRAY['delete'], NULL),
(5, 'publish', 'Make resources public', 'allow', ARRAY['publish'], 'is_owner')
ON CONFLICT (name) DO NOTHING;

-- Insert policies (role-based v1 + attribute-based v3 ABAC)
INSERT INTO policies (id_policy, name, policy_rule, is_active) VALUES
-- ===== Role-based policies (v1 — backward compatible) =====

-- 1: Admin - Full wildcard access (role-based)
(1, 'Admin - Full Access',
 '{"role": "admin", "resource": "*", "action": ["*"]}'::jsonb,
 true),

-- 2: Project Manager - Full project CRUD (role-based)
(2, 'Project Manager - Projects CRUD',
 '{"role": "project_manager", "resource": "projects", "action": ["create","read","update","delete"]}'::jsonb,
 true),

-- 3: Tech Lead - Full API docs access (role-based)
(3, 'Tech Lead - API Docs Full Access',
 '{"role": "tech_lead", "resource": "api_docs_*", "action": ["create","read","update","delete"]}'::jsonb,
 true),

-- 4: Developer - Create and edit API docs (role-based)
(4, 'Developer - API Docs CRU',
 '{"role": "developer", "resource": "api_docs_*", "action": ["create","read","update"]}'::jsonb,
 true),

-- 5: Technical Writer - Read and update API docs (role-based)
(5, 'Technical Writer - API Docs RU',
 '{"role": "technical_writer", "resource": "api_docs_*", "action": ["read","update"]}'::jsonb,
 true),

-- 6: Viewer - Read-only access (role-based)
(6, 'Viewer - Read Only',
 '{"role": "viewer", "resource": "*", "action": ["read"]}'::jsonb,
 true),

-- ===== Attribute-based policies (v3 ABAC — new format) =====

-- 7: Engineering dept - Full API docs access (attribute-based)
(7, 'Engineering - API Docs Full Access',
 '{"attribute_name": "department", "attribute_value": "engineering", "resource": "api_docs_*", "action": ["create","read","update","delete"]}'::jsonb,
 true),

-- 8: Product dept - Read & update docs (attribute-based)
(8, 'Product - API Docs Read & Update',
 '{"attribute_name": "department", "attribute_value": "product", "resource": "api_docs_*", "action": ["read","update"]}'::jsonb,
 true),

-- 9: Admin role (attribute-based v3 — alternative to policy #1)
(9, 'Admin - Full Access (ABAC v3)',
 '{"attribute_name": "role", "attribute_value": "admin", "resource": "*", "action": ["*"]}'::jsonb,
 true)
ON CONFLICT (name) DO NOTHING;

-- ============================================================================
-- PERFORMANCE QUERIES (used by Go code)
-- ============================================================================

-- Query 1: Get User Roles (used by GetUserRoles)
-- EXPLAIN ANALYZE
-- SELECT ua.value
-- FROM user_attributes ua
-- INNER JOIN attributes a ON ua.id_attribute = a.id
-- WHERE ua.id_user = $1 AND a.name = 'role';
-- Expected: <10ms with idx_user_attributes_user_attr_value

-- Query 2: Get Matching Policies (used by GetMatchingPolicies)
-- EXPLAIN ANALYZE
-- SELECT id_policy, name, policy_rule, is_active, created_at, updated_at
-- FROM policies
-- WHERE is_active = true
--   AND policy_rule->>'resource' = $1
--   AND policy_rule->>'action' = $2;
-- Expected: <20ms with idx_policies_active_gin (GIN index)

-- Query 3: Insert Audit Log
-- EXPLAIN ANALYZE
-- INSERT INTO audit_logs (id_audit_log, id_user, resource, action, allowed, reason, created_at)
-- VALUES ($1, $2, $3, $4, $5, $6);
-- Expected: <5ms (indexed on id_user, created_at)

-- Query 4: Get User Profile
-- EXPLAIN ANALYZE
-- SELECT * FROM profiles WHERE id_user = $1;
-- Expected: <5ms with idx_profiles_id_user

-- ============================================================================
-- PERFORMANCE NOTES
-- ============================================================================

-- 1. GIN Indexes on JSONB:
--    - Allow fast queries like: policy_rule->>'resource'
--    - Only active on policies table (reduces size)
--    - Critical for <50ms authorization checks

-- 2. Composite Indexes:
--    - idx_user_attributes_user_attr_value: Covers most common authorization query
--    - idx_audit_logs_user_time: Covers compliance report queries
--    - Covering indexes eliminate table lookups, reducing I/O

-- 3. Partial Indexes:
--    - idx_users_is_active WHERE is_active = false: Small, fast index for filtering active users
--    - idx_policies_active_gin: Only indexes active policies, reducing size

-- 4. Autovacuum Settings:
--    - Prevents table bloat in high-write tables (audit_logs)
--    - 0.1 scale factor = autovacuum when 10% of rows are dead
--    - 0.05 scale factor = analyze when 5% of data changes

-- 5. Foreign Keys:
--    - ON DELETE CASCADE on user_attributes: Automatically cleans up when user deleted
--    - ON DELETE CASCADE on profiles: Automatically cleans up when user deleted
--    - ON DELETE SET NULL on audit_logs: Preserves audit log but removes user reference

-- ============================================================================
-- INDEX MAINTENANCE
-- ============================================================================

-- Periodically run to maintain optimal performance:
-- ANALYZE; -- Updates statistics for query planner
-- REINDEX TABLE CONCURRENTLY policies; -- Rebuilds fragmented GIN indexes
-- VACUUM ANALYZE user_attributes; -- Removes dead rows from table
