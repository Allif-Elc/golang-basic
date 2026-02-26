-- Down migration for verify_wildcard_performance
DROP INDEX IF EXISTS idx_policies_is_active;
DROP INDEX IF EXISTS idx_policies_rule_action;
DROP INDEX IF EXISTS idx_policies_rule_resource;

DELETE FROM policies WHERE id_policy IN (1001, 1002, 1003);
DELETE FROM attributes WHERE id_attribute = 1;
