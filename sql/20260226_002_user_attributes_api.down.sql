-- Rollback: User Attributes API Indexes
-- Description: Removes indexes for user attributes CRUD operations
-- Version: 20260226_002

DROP INDEX IF EXISTS idx_user_attributes_user_covering;
DROP INDEX IF EXISTS idx_user_attributes_attribute_covering;
DROP INDEX IF EXISTS idx_user_attributes_user_attr_unique;
DROP INDEX IF EXISTS idx_user_attributes_attribute_id;
