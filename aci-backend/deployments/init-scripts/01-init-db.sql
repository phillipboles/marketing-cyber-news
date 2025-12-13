-- ACI Backend Database Initialization Script
-- This script runs automatically when PostgreSQL container starts for the first time

-- Enable pgvector extension for vector embeddings
CREATE EXTENSION IF NOT EXISTS vector;

-- Enable other useful extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";  -- For text search with trigrams
CREATE EXTENSION IF NOT EXISTS "btree_gin"; -- For GIN indexes on multiple column types

-- Create schema for application tables (optional, using public schema by default)
-- CREATE SCHEMA IF NOT EXISTS aci;
-- SET search_path TO aci, public;

-- Grant permissions to aci_user
GRANT ALL PRIVILEGES ON DATABASE aci TO aci_user;
GRANT ALL PRIVILEGES ON SCHEMA public TO aci_user;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO aci_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO aci_user;

-- Ensure future tables/sequences are also granted to aci_user
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO aci_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO aci_user;

-- Example tables (uncomment if you want to initialize schema)
-- Users table
-- CREATE TABLE IF NOT EXISTS users (
--     id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
--     email VARCHAR(255) UNIQUE NOT NULL,
--     password_hash VARCHAR(255) NOT NULL,
--     created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
--     updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
-- );

-- Sessions table
-- CREATE TABLE IF NOT EXISTS sessions (
--     id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
--     user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
--     token_hash VARCHAR(255) NOT NULL,
--     expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
--     created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
--     INDEX idx_sessions_user_id (user_id),
--     INDEX idx_sessions_token_hash (token_hash)
-- );

-- Example vector embeddings table
-- CREATE TABLE IF NOT EXISTS embeddings (
--     id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
--     content TEXT NOT NULL,
--     embedding vector(1536),  -- OpenAI ada-002 dimension
--     metadata JSONB,
--     created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
--     INDEX idx_embeddings_vector (embedding vector_cosine_ops)
-- );

-- Create indexes for performance
-- CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
-- CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);

-- Database statistics
SELECT
    'Database initialized successfully' AS status,
    NOW() AS timestamp,
    version() AS postgres_version;

-- List installed extensions
SELECT
    extname AS extension_name,
    extversion AS version
FROM pg_extension
ORDER BY extname;
