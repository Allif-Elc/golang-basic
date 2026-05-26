-- ============================================================================
-- Wildcard Authorization Performance Verification Script
-- ============================================================================
-- Purpose: Verify wildcard query performance meets <50ms p95 target
-- Usage:   psql -U your_user -d your_database -f verify_wildcard_performance.sql
-- ============================================================================

-- Ensure required indexes exist
-- These are defined in CLAUDE.md and design-document.md

-- Index for active policies (should already exist)
CREATE INDEX IF NOT EXISTS idx_policies_is_active ON policies(is_active);

-- GIN index for action array wildcards (supports @> operator)
CREATE INDEX IF NOT EXISTS idx_policies_rule_action ON policies USING GIN ((policy_rule->'action'));

-- B-tree index for resource lookups
CREATE INDEX IF NOT EXISTS idx_policies_rule_resource ON policies ((policy_rule->>'resource'));

-- ============================================================================
-- Sample test data (for performance testing)
-- ============================================================================

-- Ensure user 999 exists before referencing it in user_attributes
INSERT INTO users (id_user, name, email, password_hash, is_active)
VALUES (999, 'Performance Test Admin', 'perf-test@company.com', '$argon2id$v=19$m=65536,t=3,p=2$pSIkdXLvwu4TBZnviQUd5g$3USJSHpTUpRmjGiKty7PPStQDgvP6NmJeIgIxCnfDHI', true)
ON CONFLICT (id_user) DO NOTHING;

-- Insert sample attribute with full fields (skip if exists)
DO $$
BEGIN
    INSERT INTO attributes (id_attribute, name, description, type, enum_values)
    VALUES (1, 'role', 'User role attribute', 'enum', ARRAY['admin','project_manager','tech_lead','developer','technical_writer','viewer'])
    ON CONFLICT (name) DO NOTHING;
END $$;

-- Insert sample user with admin role (skip if exists)
DO $$
BEGIN
    INSERT INTO user_attributes (id_user, id_attribute, value) VALUES (999, 1, 'admin') ON CONFLICT DO NOTHING;
END $$;

-- Insert sample wildcard policies (skip if exists)
DO $$
BEGIN
    -- ===== Role-based wildcard policies =====

    -- Full wildcard: resource="*", action=["*"]
    INSERT INTO policies (id_policy, name, policy_rule, is_active)
    VALUES (1001, 'Admin All Access', '{"role": "admin", "resource": "*", "action": ["*"]}', true)
    ON CONFLICT (id_policy) DO NOTHING;

    -- Prefix wildcard: resource="api_docs_*", action=["*"]
    INSERT INTO policies (id_policy, name, policy_rule, is_active)
    VALUES (1002, 'API Docs Admin', '{"role": "admin", "resource": "api_docs_*", "action": ["*"]}', true)
    ON CONFLICT (id_policy) DO NOTHING;

    -- Prefix wildcard: resource="employee_*", action=["read", "write"]
    INSERT INTO policies (id_policy, name, policy_rule, is_active)
    VALUES (1003, 'Employee HR', '{"role": "hr", "resource": "employee_*", "action": ["read", "write"]}', true)
    ON CONFLICT (id_policy) DO NOTHING;

    -- ===== Attribute-based wildcard policies (ABAC v3) =====

    -- Full wildcard using attribute_name/attribute_value format
    INSERT INTO policies (id_policy, name, policy_rule, is_active)
    VALUES (1004, 'Admin All Access (ABAC v3)', '{"attribute_name": "role", "attribute_value": "admin", "resource": "*", "action": ["*"]}', true)
    ON CONFLICT (id_policy) DO NOTHING;

    -- Prefix wildcard using department attribute
    INSERT INTO policies (id_policy, name, policy_rule, is_active)
    VALUES (1005, 'Engineering API Docs (ABAC v3)', '{"attribute_name": "department", "attribute_value": "engineering", "resource": "api_docs_*", "action": ["*"]}', true)
    ON CONFLICT (id_policy) DO NOTHING;

    -- Prefix wildcard: department-based employee access
    INSERT INTO policies (id_policy, name, policy_rule, is_active)
    VALUES (1006, 'HR Employee Records (ABAC v3)', '{"attribute_name": "department", "attribute_value": "hr", "resource": "employee_*", "action": ["read", "write"]}', true)
    ON CONFLICT (id_policy) DO NOTHING;
END $$;

-- ============================================================================
-- Performance Test Queries
-- ============================================================================

-- Test 1: Exact match (baseline - should be fastest)
EXPLAIN (ANALYZE, BUFFERS, TIMING)
SELECT id_policy, name, policy_rule, is_active, created_at, updated_at
FROM policies
WHERE is_active = true
  AND (
    policy_rule->>'resource' = 'employee_records'
    OR policy_rule->>'resource' = '*'
    OR (
      policy_rule->>'resource' LIKE '%_\*'
      AND substr(policy_rule->>'resource', 1, length(policy_rule->>'resource') - 2) = substr('employee_records', 1, length(policy_rule->>'resource') - 2)
      AND substr('employee_records', length(policy_rule->>'resource') - 1, 1) = '_'
      AND length('employee_records') > length(policy_rule->>'resource') - 1
    )
  )
  AND (
    policy_rule->'action' @> to_jsonb(ARRAY['read'])
    OR policy_rule->'action' @> to_jsonb(ARRAY['*'])
  );

-- Test 2: Full wildcard "*" match
EXPLAIN (ANALYZE, BUFFERS, TIMING)
SELECT id_policy, name, policy_rule, is_active, created_at, updated_at
FROM policies
WHERE is_active = true
  AND (
    policy_rule->>'resource' = 'any_resource'
    OR policy_rule->>'resource' = '*'
    OR (
      policy_rule->>'resource' LIKE '%_\*'
      AND substr(policy_rule->>'resource', 1, length(policy_rule->>'resource') - 2) = substr('any_resource', 1, length(policy_rule->>'resource') - 2)
      AND substr('any_resource', length(policy_rule->>'resource') - 1, 1) = '_'
      AND length('any_resource') > length(policy_rule->>'resource') - 1
    )
  )
  AND (
    policy_rule->'action' @> to_jsonb(ARRAY['any_action'])
    OR policy_rule->'action' @> to_jsonb(ARRAY['*'])
  );

-- Test 3: Prefix wildcard "api_docs_*" match
EXPLAIN (ANALYZE, BUFFERS, TIMING)
SELECT id_policy, name, policy_rule, is_active, created_at, updated_at
FROM policies
WHERE is_active = true
  AND (
    policy_rule->>'resource' = 'api_docs_project'
    OR policy_rule->>'resource' = '*'
    OR (
      policy_rule->>'resource' LIKE '%_\*'
      AND substr(policy_rule->>'resource', 1, length(policy_rule->>'resource') - 2) = substr('api_docs_project', 1, length(policy_rule->>'resource') - 2)
      AND substr('api_docs_project', length(policy_rule->>'resource') - 1, 1) = '_'
      AND length('api_docs_project') > length(policy_rule->>'resource') - 1
    )
  )
  AND (
    policy_rule->'action' @> to_jsonb(ARRAY['read'])
    OR policy_rule->'action' @> to_jsonb(ARRAY['*'])
  );

-- Test 4: Action wildcard ["*"] match
EXPLAIN (ANALYZE, BUFFERS, TIMING)
SELECT id_policy, name, policy_rule, is_active, created_at, updated_at
FROM policies
WHERE is_active = true
  AND (
    policy_rule->>'resource' = 'api_docs_project'
    OR policy_rule->>'resource' = '*'
    OR (
      policy_rule->>'resource' LIKE '%_\*'
      AND substr(policy_rule->>'resource', 1, length(policy_rule->>'resource') - 2) = substr('api_docs_project', 1, length(policy_rule->>'resource') - 2)
      AND substr('api_docs_project', length(policy_rule->>'resource') - 1, 1) = '_'
      AND length('api_docs_project') > length(policy_rule->>'resource') - 1
    )
  )
  AND (
    policy_rule->'action' @> to_jsonb(ARRAY['delete'])
    OR policy_rule->'action' @> to_jsonb(ARRAY['*'])
  );

-- ============================================================================
-- Performance Verification Summary
-- ============================================================================
-- Expected results for each query (with proper indexes):
--
-- 1. Execution Time: <10ms (well under 50ms p95 target)
-- 2. Planning Time: <1ms
-- 3. Buffer Usage: Minimal shared hits (cold cache)
-- 4. Execution Plan:
--    - Index Scan using idx_policies_is_active or idx_policies_rule_resource
--    - Bitmap Index Scan using idx_policies_rule_action (GIN)
--    - No Seq Scan on policies table
--
-- If queries show Seq Scan or execution time >50ms:
--   1. Verify indexes are created (check with \d policies)
--   2. Run ANALYZE policies to update statistics
--   3. Check for missing GIN indexes on JSONB columns
--
-- View actual indexes:
-- \d policies
--
-- Update statistics:
-- ANALYZE policies;
--
-- ============================================================================
