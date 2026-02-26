-- Down migration for api_docs_schema
ALTER TABLE tags DROP CONSTRAINT IF EXISTS chk_tags_color_format;
ALTER TABLE projects DROP CONSTRAINT IF EXISTS chk_projects_version_format;
ALTER TABLE rest_apis DROP CONSTRAINT IF EXISTS chk_rest_apis_endpoint_format;
DROP TABLE IF EXISTS api_tags CASCADE;
DROP TABLE IF EXISTS tags CASCADE;
DROP TABLE IF EXISTS grpc_apis CASCADE;
DROP TABLE IF EXISTS graphql_apis CASCADE;
DROP TABLE IF EXISTS rest_apis CASCADE;
DROP TABLE IF EXISTS projects CASCADE;

DROP TRIGGER IF EXISTS update_projects_updated_at ON projects;
DROP TRIGGER IF EXISTS update_rest_apis_updated_at ON rest_apis;
DROP TRIGGER IF EXISTS update_graphql_apis_updated_at ON graphql_apis;
DROP TRIGGER IF EXISTS update_grpc_apis_updated_at ON grpc_apis;
