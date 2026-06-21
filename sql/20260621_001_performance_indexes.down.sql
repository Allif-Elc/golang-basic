-- Rollback performance indexes
DROP INDEX IF EXISTS idx_policies_attr_name_value;
DROP INDEX IF EXISTS idx_projects_description_notnull;
DROP INDEX IF EXISTS idx_projects_public_created;
DROP INDEX IF EXISTS idx_rest_apis_project_id_created;
DROP INDEX IF EXISTS idx_graphql_apis_project_id_created;
DROP INDEX IF EXISTS idx_grpc_apis_project_id_created;
DROP INDEX IF EXISTS idx_user_policies_expired;
DROP INDEX IF EXISTS idx_api_tags_composite;
DROP INDEX IF EXISTS idx_users_email_covering;
