-- Migration 000005: Audit Schema (Rollback)
-- Description: Remove webhook logging and audit trail tables
-- Author: Database Developer Agent
-- Date: 2025-12-11

-- Drop functions
DROP FUNCTION IF EXISTS cleanup_old_webhook_logs();
DROP FUNCTION IF EXISTS get_resource_audit_trail(VARCHAR, UUID, INTEGER);
DROP FUNCTION IF EXISTS get_user_audit_trail(UUID, INTEGER, INTEGER);
DROP FUNCTION IF EXISTS get_webhook_retry_queue(INTEGER);
DROP FUNCTION IF EXISTS log_audit_event(UUID, VARCHAR, VARCHAR, UUID, JSONB, JSONB, INET, TEXT);
DROP FUNCTION IF EXISTS update_webhook_status(UUID, VARCHAR, TEXT);
DROP FUNCTION IF EXISTS log_webhook_event(VARCHAR, VARCHAR, VARCHAR, JSONB);

-- Drop trigger
DROP TRIGGER IF EXISTS update_webhook_logs_updated_at ON webhook_logs;

-- Drop indexes
DROP INDEX IF EXISTS idx_audit_logs_new_value;
DROP INDEX IF EXISTS idx_audit_logs_old_value;
DROP INDEX IF EXISTS idx_audit_logs_resource;
DROP INDEX IF EXISTS idx_audit_logs_user_created;
DROP INDEX IF EXISTS idx_audit_logs_created_at;
DROP INDEX IF EXISTS idx_audit_logs_resource_id;
DROP INDEX IF EXISTS idx_audit_logs_resource_type;
DROP INDEX IF EXISTS idx_audit_logs_action;
DROP INDEX IF EXISTS idx_audit_logs_user_id;
DROP INDEX IF EXISTS idx_webhook_logs_payload;
DROP INDEX IF EXISTS idx_webhook_logs_retry;
DROP INDEX IF EXISTS idx_webhook_logs_processed_at;
DROP INDEX IF EXISTS idx_webhook_logs_created_at;
DROP INDEX IF EXISTS idx_webhook_logs_status;
DROP INDEX IF EXISTS idx_webhook_logs_execution_id;
DROP INDEX IF EXISTS idx_webhook_logs_workflow_id;
DROP INDEX IF EXISTS idx_webhook_logs_event_type;

-- Drop tables (CASCADE handles foreign key dependencies)
DROP TABLE IF EXISTS audit_logs CASCADE;
DROP TABLE IF EXISTS webhook_logs CASCADE;
