-- Down migration for user_policies
DROP TABLE IF EXISTS user_policies CASCADE;
DROP TRIGGER IF EXISTS update_user_policies_updated_at ON user_policies;
