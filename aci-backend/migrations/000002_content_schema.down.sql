-- Migration 000002: Content Schema (Rollback)
-- Description: Remove content management tables
-- Author: Database Developer Agent
-- Date: 2025-12-11

-- Drop function
DROP FUNCTION IF EXISTS increment_article_views(UUID);

-- Drop triggers
DROP TRIGGER IF EXISTS update_articles_updated_at ON articles;
DROP TRIGGER IF EXISTS update_sources_updated_at ON sources;
DROP TRIGGER IF EXISTS update_categories_updated_at ON categories;

-- Drop indexes (explicit drops for clarity, though CASCADE handles this)
DROP INDEX IF EXISTS idx_articles_embedding;
DROP INDEX IF EXISTS idx_articles_search_vector;
DROP INDEX IF EXISTS idx_articles_armor_cta;
DROP INDEX IF EXISTS idx_articles_iocs;
DROP INDEX IF EXISTS idx_articles_recommended_actions;
DROP INDEX IF EXISTS idx_articles_vendors;
DROP INDEX IF EXISTS idx_articles_cves;
DROP INDEX IF EXISTS idx_articles_tags;
DROP INDEX IF EXISTS idx_articles_armor_relevance;
DROP INDEX IF EXISTS idx_articles_severity_published;
DROP INDEX IF EXISTS idx_articles_category_published;
DROP INDEX IF EXISTS idx_articles_is_published;
DROP INDEX IF EXISTS idx_articles_created_at;
DROP INDEX IF EXISTS idx_articles_published_at;
DROP INDEX IF EXISTS idx_articles_severity;
DROP INDEX IF EXISTS idx_articles_source_id;
DROP INDEX IF EXISTS idx_articles_category_id;
DROP INDEX IF EXISTS idx_sources_last_scraped_at;
DROP INDEX IF EXISTS idx_sources_trust_score;
DROP INDEX IF EXISTS idx_sources_is_active;
DROP INDEX IF EXISTS idx_sources_name;
DROP INDEX IF EXISTS idx_categories_created_at;
DROP INDEX IF EXISTS idx_categories_slug;

-- Drop tables (CASCADE handles foreign key dependencies)
DROP TABLE IF EXISTS articles CASCADE;
DROP TABLE IF EXISTS sources CASCADE;
DROP TABLE IF EXISTS categories CASCADE;
