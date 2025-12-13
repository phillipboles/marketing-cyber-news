-- Migration 000001: Initial Schema (Rollback)
-- Description: Remove users and authentication tables
-- Author: Database Developer Agent
-- Date: 2025-12-11

-- Drop triggers
DROP TRIGGER IF EXISTS update_user_preferences_updated_at ON user_preferences;
DROP TRIGGER IF EXISTS update_refresh_tokens_updated_at ON refresh_tokens;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Drop trigger function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables (CASCADE handles foreign key dependencies)
DROP TABLE IF EXISTS user_preferences CASCADE;
DROP TABLE IF EXISTS refresh_tokens CASCADE;
DROP TABLE IF EXISTS users CASCADE;

-- Note: Extensions are NOT dropped as they may be used by other migrations
-- DROP EXTENSION IF EXISTS "vector";
-- DROP EXTENSION IF EXISTS "uuid-ossp";
