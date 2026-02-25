-- ============================================================================
-- User Policies Override Schema
-- Priority-based user-to-policy assignment system
-- Allows direct policy assignment to specific users with configurable priority
-- Updated: 2025-01-17
-- ============================================================================

-- ============================================================================
-- TABLES
-- ============================================================================

-- User policies: direct user-to-policy assignments with priority
-- Allows granting exceptions or additional policies to specific users
-- Example: A manager needs temporary access to a resource their role doesn't normally have
CREATE TABLE IF NOT EXISTS user_policies (
    id_user_policy BIGSERIAL PRIMARY KEY,
    id_user BIGINT NOT NULL,
    id_policy BIGINT NOT NULL,
    priority INTEGER NOT NULL DEFAULT 0,  -- Higher = evaluated first, range: -100 to 100
    is_active BOOLEAN NOT NULL DEFAULT true,
    expires_at TIMESTAMP,                 -- Optional expiration for temporary access
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by BIGINT,                    -- User who granted the policy (nullable)
    UNIQUE(id_user, id_policy),
    FOREIGN KEY (id_user) REFERENCES users(id_user) ON DELETE CASCADE,
    FOREIGN KEY (id_policy) REFERENCES policies(id_policy) ON DELETE CASCADE
);

-- ============================================================================
-- INDEXES (Critical for <50ms Performance)
-- ============================================================================

-- Index for user lookups (most common query pattern)
CREATE INDEX IF NOT EXISTS idx_user_policies_id_user
ON user_policies(id_user);

-- Index for policy reverse lookups (find all users with a specific policy)
CREATE INDEX IF NOT EXISTS idx_user_policies_id_policy
ON user_policies(id_policy);

-- Priority index for sorting policies by evaluation order
-- DESC order because higher priority = evaluated first
CREATE INDEX IF NOT EXISTS idx_user_policies_priority
ON user_policies(priority DESC);

-- Composite index for active user policy lookups (authorization path)
-- Used by: GetUserPoliciesByUserID in authorization flow
CREATE INDEX IF NOT EXISTS idx_user_policies_active
ON user_policies(id_user, is_active)
WHERE is_active = true;

-- Partial index for expiring policies (cleanup and filtering)
CREATE INDEX IF NOT EXISTS idx_user_policies_expires_at
ON user_policies(expires_at)
WHERE expires_at IS NOT NULL;

-- Composite index for user + active + expires (full authorization query)
-- Covers the most common authorization query with all filters
CREATE INDEX IF NOT EXISTS idx_user_policies_user_active_expires
ON user_policies(id_user, is_active, expires_at)
WHERE is_active = true;

-- ============================================================================
-- AUTOVACUUM SETTINGS
-- ============================================================================

-- Analyze table for query planner statistics
ANALYZE user_policies;

-- Set autovacuum for user_policies (moderate write frequency)
ALTER TABLE user_policies SET (autovacuum_vacuum_scale_factor = 0.1);
ALTER TABLE user_policies SET (autovacuum_analyze_scale_factor = 0.05);

-- ============================================================================
-- PERFORMANCE QUERIES (used by Go code)
-- ============================================================================

-- Query 1: Get User Policies with Priority (used by GetUserPoliciesByUserID)
-- EXPLAIN ANALYZE
-- SELECT p.id_policy, p.name, p.policy_rule, p.is_active, p.created_at, p.updated_at, up.priority
-- FROM user_policies up
-- INNER JOIN policies p ON up.id_policy = p.id_policy
-- WHERE up.id_user = $1
--   AND up.is_active = true
--   AND p.is_active = true
--   AND (up.expires_at IS NULL OR up.expires_at > CURRENT_TIMESTAMP)
-- ORDER BY up.priority DESC;
-- Expected: <10ms with idx_user_policies_user_active_expires

-- Query 2: Check for duplicate user-policy assignment
-- EXPLAIN ANALYZE
-- SELECT COUNT(*) FROM user_policies WHERE id_user = $1 AND id_policy = $2;
-- Expected: <5ms with UNIQUE constraint index

-- Query 3: Get User Policy Details with JOIN (list view)
-- EXPLAIN ANALYZE
-- SELECT up.id_user_policy, up.id_user, u.name as user_name, u.email as user_email,
--        up.id_policy, p.name as policy_name, p.policy_rule,
--        up.priority, up.is_active, up.expires_at, up.created_at
-- FROM user_policies up
-- INNER JOIN users u ON up.id_user = u.id_user
-- INNER JOIN policies p ON up.id_policy = p.id_policy;
-- Expected: <20ms with proper join indexes

-- ============================================================================
-- PERFORMANCE NOTES
-- ============================================================================

-- 1. Priority System:
--    - User policies have configurable priority (-100 to 100)
--    - Higher priority = evaluated first in authorization flow
--    - Role-based policies have implicit priority 0
--    - First matching policy wins (allow or deny)

-- 2. Index Strategy:
--    - idx_user_policies_id_user: Fast user lookups
--    - idx_user_policies_priority: Sorted policy evaluation
--    - idx_user_policies_active: Active-only filtering
--    - idx_user_policies_user_active_expires: Covers full authorization query

-- 3. Expiration Handling:
--    - NULL expires_at = permanent assignment
--    - Non-NULL expires_at = temporary access grant
--    - Expired policies are filtered out in authorization queries
--    - Cleanup job can delete expired policies periodically

-- 4. Foreign Keys:
--    - ON DELETE CASCADE on id_user: Clean up when user deleted
--    - ON DELETE CASCADE on id_policy: Clean up when policy deleted

-- 5. Authorization Flow:
--    - User policies evaluated BEFORE role-based policies
--    - User policies with priority > 0 override role-based policies
--    - Negative priority can be used to evaluate after role-based policies

-- ============================================================================
-- SAMPLE DATA (for testing)
-- ============================================================================

-- Note: This sample data assumes the users and policies from abac_schema.sql exist
-- Run this after abac_schema.sql to ensure foreign key constraints are satisfied

-- Grant user 2 (Bob Manager) direct access to employee_records with high priority
-- This overrides his normal role-based permissions for this specific resource
INSERT INTO user_policies (id_user, id_policy, priority, is_active, created_by)
VALUES (2, 1, 10, true, 7)  -- Granted by admin (user 7)
ON CONFLICT (id_user, id_policy) DO NOTHING;

-- Grant user 4 (Dave Staff) temporary HR-like access that expires in 30 days
INSERT INTO user_policies (id_user, id_policy, priority, is_active, expires_at, created_by)
VALUES (4, 1, 5, true, CURRENT_TIMESTAMP + INTERVAL '30 days', 7)
ON CONFLICT (id_user, id_policy) DO NOTHING;

-- ============================================================================
-- MAINTENANCE QUERIES
-- ============================================================================

-- Clean up expired user policies (run periodically via cron/job)
-- DELETE FROM user_policies WHERE expires_at IS NOT NULL AND expires_at < CURRENT_TIMESTAMP;

-- Find all users with expiring policies within next 7 days (for notifications)
-- SELECT id_user, id_policy, expires_at
-- FROM user_policies
-- WHERE expires_at IS NOT NULL
--   AND expires_at BETWEEN CURRENT_TIMESTAMP AND CURRENT_TIMESTAMP + INTERVAL '7 days';

-- Find all active user policies for a specific user (admin view)
-- SELECT up.id_user_policy, p.name as policy_name, up.priority, up.expires_at
-- FROM user_policies up
-- INNER JOIN policies p ON up.id_policy = p.id_policy
-- WHERE up.id_user = $1 AND up.is_active = true;
