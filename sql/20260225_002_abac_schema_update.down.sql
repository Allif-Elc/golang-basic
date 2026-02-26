-- Down migration for abac_schema_update
ALTER TABLE attributes DROP COLUMN IF EXISTS type;
ALTER TABLE attributes DROP COLUMN IF EXISTS enum_values;
ALTER TABLE attributes DROP COLUMN IF EXISTS updated_at;

ALTER TABLE resources DROP COLUMN IF EXISTS resource_type;
ALTER TABLE resources DROP COLUMN IF EXISTS updated_at;

ALTER TABLE permissions DROP COLUMN IF EXISTS effect;
ALTER TABLE permissions DROP COLUMN IF EXISTS actions;
ALTER TABLE permissions DROP COLUMN IF EXISTS condition;
ALTER TABLE permissions DROP COLUMN IF EXISTS updated_at;

DROP INDEX IF EXISTS idx_attributes_type;
DROP INDEX IF EXISTS idx_resources_resource_type;
DROP INDEX IF EXISTS idx_permissions_effect;
DROP INDEX IF EXISTS idx_permissions_actions;

DROP TRIGGER IF EXISTS update_attributes_updated_at ON attributes;
DROP TRIGGER IF EXISTS update_resources_updated_at ON resources;
DROP TRIGGER IF EXISTS update_permissions_updated_at ON permissions;
