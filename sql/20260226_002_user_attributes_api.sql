-- Migration: User Attributes API Indexes
-- Description: Adds indexes for user attributes CRUD operations and attribute-based authorization
-- Version: 20260226_002

-- Covering index for user attributes lookup by user
-- Includes frequently accessed columns for index-only scans
-- Performance: <10ms for user attribute queries
CREATE INDEX IF NOT EXISTS idx_user_attributes_user_covering
    ON user_attributes(id_user, id_attribute, value)
    INCLUDE (created_at);

-- Covering index for user attributes lookup by attribute
-- Useful for finding all users with a specific attribute
CREATE INDEX IF NOT EXISTS idx_user_attributes_attribute_covering
    ON user_attributes(id_attribute, value)
    INCLUDE (id_user, created_at);

-- Unique index for user+attribute+value combination
-- Prevents duplicate attribute assignments to the same user
-- Required by CreateUserAttribute business logic
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_attributes_user_attr_unique
    ON user_attributes(id_user, id_attribute, value);

-- Index for user attributes with joins to attributes table
-- Improves performance of GetUserAttributes with attribute details
CREATE INDEX IF NOT EXISTS idx_user_attributes_attribute_id
    ON user_attributes(id_attribute);

-- Comments for documentation
COMMENT ON INDEX idx_user_attributes_user_covering IS 'Covering index for user attribute lookups by user ID';
COMMENT ON INDEX idx_user_attributes_attribute_covering IS 'Covering index for finding users by attribute';
COMMENT ON INDEX idx_user_attributes_user_attr_unique IS 'Prevents duplicate user+attribute+value combinations';
COMMENT ON INDEX idx_user_attributes_attribute_id IS 'Index for joining user_attributes with attributes table';
