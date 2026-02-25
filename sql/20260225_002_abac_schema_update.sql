-- ============================================================================
-- ABAC Schema Update - Add Missing Fields for Frontend Compatibility
-- Updated: 2025-02-24
-- ============================================================================

-- ============================================================================
-- ATTRIBUTES TABLE UPDATES
-- ============================================================================

-- Add type column for attribute type (string, number, boolean, enum)
ALTER TABLE attributes ADD COLUMN IF NOT EXISTS type VARCHAR(20) DEFAULT 'string' CHECK (type IN ('string', 'number', 'boolean', 'enum'));

-- Add enum_values array for enum type attributes
ALTER TABLE attributes ADD COLUMN IF NOT EXISTS enum_values TEXT[];

-- Add updated_at timestamp
ALTER TABLE attributes ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

-- ============================================================================
-- RESOURCES TABLE UPDATES
-- ============================================================================

-- Add resource_type column
ALTER TABLE resources ADD COLUMN IF NOT EXISTS resource_type VARCHAR(100) DEFAULT 'general';

-- Add updated_at timestamp
ALTER TABLE resources ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

-- ============================================================================
-- PERMISSIONS TABLE UPDATES
-- ============================================================================

-- Add effect column (allow/deny)
ALTER TABLE permissions ADD COLUMN IF NOT EXISTS effect VARCHAR(10) DEFAULT 'allow' CHECK (effect IN ('allow', 'deny'));

-- Add actions array for multiple actions
ALTER TABLE permissions ADD COLUMN IF NOT EXISTS actions TEXT[] DEFAULT '{}';

-- Add condition column for conditional logic
ALTER TABLE permissions ADD COLUMN IF NOT EXISTS condition TEXT;

-- Add updated_at timestamp
ALTER TABLE permissions ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

-- ============================================================================
-- INDEXES FOR NEW FIELDS
-- ============================================================================

-- Index for attributes type
CREATE INDEX IF NOT EXISTS idx_attributes_type ON attributes(type);

-- Index for resources resource_type
CREATE INDEX IF NOT EXISTS idx_resources_resource_type ON resources(resource_type);

-- Index for permissions effect
CREATE INDEX IF NOT EXISTS idx_permissions_effect ON permissions(effect);

-- GIN index for permissions actions (for array queries)
CREATE INDEX IF NOT EXISTS idx_permissions_actions ON permissions USING GIN (actions);

-- ============================================================================
-- TRIGGERS FOR UPDATED_AT
-- ============================================================================

-- Ensure the update_updated_at_column function exists
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Trigger for attributes table
DROP TRIGGER IF EXISTS update_attributes_updated_at ON attributes;
CREATE TRIGGER update_attributes_updated_at
BEFORE UPDATE ON attributes
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Trigger for resources table
DROP TRIGGER IF EXISTS update_resources_updated_at ON resources;
CREATE TRIGGER update_resources_updated_at
BEFORE UPDATE ON resources
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Trigger for permissions table
DROP TRIGGER IF EXISTS update_permissions_updated_at ON permissions;
CREATE TRIGGER update_permissions_updated_at
BEFORE UPDATE ON permissions
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- ANALYZE TABLES FOR QUERY PLANNER STATISTICS
-- ============================================================================

ANALYZE attributes;
ANALYZE resources;
ANALYZE permissions;

-- ============================================================================
-- NOTES
-- ============================================================================
-- 1. The type column in attributes allows specifying the data type:
--    - string: Text values
--    - number: Numeric values
--    - boolean: True/false values
--    - enum: Predefined set of values (stored in enum_values)
--
-- 2. The resource_type column allows categorizing resources
--
-- 3. The effect column in permissions specifies allow/deny
--
-- 4. The actions array allows multiple actions per permission
--    Example: '{create, read, update}'
--
-- 5. The condition column stores conditional logic as text
--    Example: "resource.owner == user.id"
