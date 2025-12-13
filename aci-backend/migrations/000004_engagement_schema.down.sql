-- Migration 000004: Engagement Schema (Rollback)
-- Description: Remove engagement tracking tables
-- Author: Database Developer Agent
-- Date: 2025-12-11

-- Drop functions
DROP FUNCTION IF EXISTS get_user_reading_stats(UUID);
DROP FUNCTION IF EXISTS generate_daily_stats(DATE);
DROP FUNCTION IF EXISTS toggle_bookmark(UUID, UUID);
DROP FUNCTION IF EXISTS record_article_read(UUID, UUID, INTEGER);

-- Drop trigger
DROP TRIGGER IF EXISTS update_daily_stats_updated_at ON daily_stats;

-- Drop indexes
DROP INDEX IF EXISTS idx_daily_stats_articles_by_category;
DROP INDEX IF EXISTS idx_daily_stats_total_articles;
DROP INDEX IF EXISTS idx_daily_stats_date;
DROP INDEX IF EXISTS idx_article_reads_article_read_at;
DROP INDEX IF EXISTS idx_article_reads_user_read_at;
DROP INDEX IF EXISTS idx_article_reads_read_at;
DROP INDEX IF EXISTS idx_article_reads_article_id;
DROP INDEX IF EXISTS idx_article_reads_user_id;
DROP INDEX IF EXISTS idx_bookmarks_user_created;
DROP INDEX IF EXISTS idx_bookmarks_created_at;
DROP INDEX IF EXISTS idx_bookmarks_article_id;
DROP INDEX IF EXISTS idx_bookmarks_user_id;

-- Drop tables (CASCADE handles foreign key dependencies)
DROP TABLE IF EXISTS daily_stats CASCADE;
DROP TABLE IF EXISTS article_reads CASCADE;
DROP TABLE IF EXISTS bookmarks CASCADE;
