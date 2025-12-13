-- Migration 000003: Alerts Schema (Rollback)
-- Description: Remove alerts and alert matching tables
-- Author: Database Developer Agent
-- Date: 2025-12-11

-- Drop functions
DROP FUNCTION IF EXISTS mark_alerts_notified(UUID[]);
DROP FUNCTION IF EXISTS match_article_alerts(UUID);

-- Drop trigger
DROP TRIGGER IF EXISTS update_alerts_updated_at ON alerts;

-- Drop indexes
DROP INDEX IF EXISTS idx_alert_matches_user_unnotified;
DROP INDEX IF EXISTS idx_alert_matches_unnotified;
DROP INDEX IF EXISTS idx_alert_matches_priority;
DROP INDEX IF EXISTS idx_alert_matches_matched_at;
DROP INDEX IF EXISTS idx_alert_matches_article_id;
DROP INDEX IF EXISTS idx_alert_matches_alert_id;
DROP INDEX IF EXISTS idx_alerts_user_active;
DROP INDEX IF EXISTS idx_alerts_is_active;
DROP INDEX IF EXISTS idx_alerts_type;
DROP INDEX IF EXISTS idx_alerts_user_id;

-- Drop tables (CASCADE handles foreign key dependencies)
DROP TABLE IF EXISTS alert_matches CASCADE;
DROP TABLE IF EXISTS alerts CASCADE;
