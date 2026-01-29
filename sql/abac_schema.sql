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
INSERT INTO attributes (id_attribute, name, description) VALUES
(1, 'role', 'User role attribute'),
(2, 'department', 'Department attribute')
ON CONFLICT (name) DO NOTHING;

-- Insert sample users
INSERT INTO users (id_user, name, email, password_hash, is_active) VALUES
(1, 'Alice HR', 'alice@company.com', '$argon2id$v=19$m=65536,t=3,p=2$EXAMPLE_HASH_1', true),
(2, 'Bob Manager', 'bob@company.com', '$argon2id$v=19$m=65536,t=3,p=2$EXAMPLE_HASH_2', true),
(3, 'Charlie Engineer', 'charlie@company.com', '$argon2id$v=19$m=65536,t=3,p=2$EXAMPLE_HASH_3', true),
(4, 'Dave Staff', 'dave@company.com', '$argon2id$v=19$m=65536,t=3,p=2$EXAMPLE_HASH_4', true),
(5, 'Eve Developer', 'eve@company.com', '$argon2id$v=19$m=65536,t=3,p=2$EXAMPLE_HASH_5', true),
(6, 'Frank TechWriter', 'frank@company.com', '$argon2id$v=19$m=65536,t=3,p=2$EXAMPLE_HASH_6', true),
(7, 'Grace Admin', 'grace@company.com', '$argon2id$v=19$m=65536,t=3,p=2$EXAMPLE_HASH_7', true),
(8, 'Henry Viewer', 'henry@company.com', '$argon2id$v=19$m=65536,t=3,p=2$EXAMPLE_HASH_8', true)
ON CONFLICT (email) DO NOTHING;

-- Insert sample profiles
INSERT INTO profiles (id_profile, id_user, age, gender, bio) VALUES
(1, 1, 30, 'female', 'Alice the HR Manager'),
(2, 2, 35, 'male', 'Bob the Budget Manager'),
(3, 3, 28, 'male', 'Charlie the Engineer'),
(4, 4, 25, 'male', 'Dave the Staff member'),
(5, 5, 27, 'female', 'Eve the Developer'),
(6, 6, 32, 'male', 'Frank the Technical Writer'),
(7, 7, 40, 'female', 'Grace the Admin'),
(8, 8, 26, 'male', 'Henry the Viewer')
ON CONFLICT (id_user) DO NOTHING;

-- Assign roles to users
INSERT INTO user_attributes (id_user, id_attribute, value) VALUES
(1, 1, 'HR'),              -- alice has HR role
(2, 1, 'Manager'),         -- bob has Manager role
(3, 1, 'Engineer'),        -- charlie has Engineer role
(4, 1, 'Staff'),           -- dave has Staff role
(5, 1, 'developer'),       -- eve has developer role
(6, 1, 'technical_writer'), -- frank has technical_writer role
(7, 1, 'admin'),           -- grace has admin role
(8, 1, 'viewer')           -- henry has viewer role
ON CONFLICT (id_user, id_attribute, value) DO NOTHING;

-- Insert resources
INSERT INTO resources (id_resource, name, description) VALUES
(1, 'employee_records', 'Employee HR records'),
(2, 'budget_report', 'Budget reports'),
(3, 'api_docs_project', 'API Documentation Projects'),
(4, 'api_docs_rest_api', 'REST API Documentation'),
(5, 'api_docs_graphql_api', 'GraphQL API Documentation'),
(6, 'api_docs_grpc_api', 'gRPC API Documentation')
ON CONFLICT (name) DO NOTHING;

-- Insert permissions
INSERT INTO permissions (id_permission, name, description) VALUES
(1, 'create', 'Create new resources'),
(2, 'read', 'View resources'),
(3, 'update', 'Edit resources'),
(4, 'delete', 'Delete resources'),
(5, 'publish', 'Make resources public')
ON CONFLICT (name) DO NOTHING;

-- Insert policies
INSERT INTO policies (id_policy, name, policy_rule, is_active) VALUES
(1, 'HR - Employee Records Read/Write',
'{"role": "HR", "resource": "employee_records", "action": ["read", "write"]}'::jsonb,
true),
(2, 'Manager - Budget Report Read',
'{"role": "Manager", "resource": "budget_report", "action": ["read"]}'::jsonb,
true),
(3, 'Engineer - Employee Records Read',
'{"role": "Engineer", "resource": "employee_records", "action": ["read"]}'::jsonb,
true),
(4, 'Manager - Budget Report Write',
'{"role": "Manager", "resource": "budget_report", "action": ["write"]}'::jsonb,
false),  -- Disabled policy
(5, 'Developer - API Docs Create',
'{"role": "developer", "resource": "api_docs_*", "action": ["create", "read", "update"]}'::jsonb,
true),
(6, 'Technical Writer - API Docs Read/Update',
'{"role": "technical_writer", "resource": "api_docs_*", "action": ["read", "update"]}'::jsonb,
true),
(7, 'Viewer - API Docs Read Only',
'{"role": "viewer", "resource": "api_docs_*", "action": ["read"]}'::jsonb,
true),
(8, 'Admin - All Access',
'{"role": "admin", "resource": "*", "action": ["*"]}'::jsonb,
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
