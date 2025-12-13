-- Migration 000005: Audit Schema
-- Description: Webhook logging and audit trail for compliance
-- Author: Database Developer Agent
-- Date: 2025-12-11

-- Webhook logs table (n8n integration tracking)
CREATE TABLE webhook_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_type VARCHAR(100) NOT NULL,
    workflow_id VARCHAR(255),
    execution_id VARCHAR(255),
    payload JSONB NOT NULL DEFAULT '{}'::JSONB,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    error_message TEXT,
    processed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chk_event_type_not_empty CHECK (LENGTH(event_type) >= 1),
    CONSTRAINT chk_status_valid CHECK (
        status IN ('pending', 'processing', 'completed', 'failed', 'retrying')
    )
);

-- Audit logs table (user action tracking for compliance)
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID,
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(100) NOT NULL,
    resource_id UUID,
    old_value JSONB,
    new_value JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_audit_logs_user FOREIGN KEY (user_id)
        REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT chk_action_not_empty CHECK (LENGTH(action) >= 1),
    CONSTRAINT chk_resource_type_not_empty CHECK (LENGTH(resource_type) >= 1)
);

-- Indexes for webhook_logs table
CREATE INDEX idx_webhook_logs_event_type ON webhook_logs(event_type);
CREATE INDEX idx_webhook_logs_workflow_id ON webhook_logs(workflow_id);
CREATE INDEX idx_webhook_logs_execution_id ON webhook_logs(execution_id);
CREATE INDEX idx_webhook_logs_status ON webhook_logs(status);
CREATE INDEX idx_webhook_logs_created_at ON webhook_logs(created_at DESC);
CREATE INDEX idx_webhook_logs_processed_at ON webhook_logs(processed_at DESC NULLS LAST);

-- Composite index for webhook retry queries
CREATE INDEX idx_webhook_logs_retry ON webhook_logs(status, created_at)
    WHERE status IN ('failed', 'retrying');

-- GIN index for payload searches
CREATE INDEX idx_webhook_logs_payload ON webhook_logs USING GIN(payload);

-- Indexes for audit_logs table
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_resource_type ON audit_logs(resource_type);
CREATE INDEX idx_audit_logs_resource_id ON audit_logs(resource_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at DESC);

-- Composite index for user activity queries
CREATE INDEX idx_audit_logs_user_created ON audit_logs(user_id, created_at DESC);

-- Composite index for resource audit trail
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource_type, resource_id, created_at DESC);

-- GIN indexes for JSONB value changes
CREATE INDEX idx_audit_logs_old_value ON audit_logs USING GIN(old_value);
CREATE INDEX idx_audit_logs_new_value ON audit_logs USING GIN(new_value);

-- Apply updated_at trigger
CREATE TRIGGER update_webhook_logs_updated_at
    BEFORE UPDATE ON webhook_logs
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Function to log webhook events
CREATE OR REPLACE FUNCTION log_webhook_event(
    event_type_param VARCHAR,
    workflow_id_param VARCHAR DEFAULT NULL,
    execution_id_param VARCHAR DEFAULT NULL,
    payload_param JSONB DEFAULT '{}'::JSONB
)
RETURNS UUID AS $$
DECLARE
    log_id UUID;
BEGIN
    INSERT INTO webhook_logs (event_type, workflow_id, execution_id, payload)
    VALUES (event_type_param, workflow_id_param, execution_id_param, payload_param)
    RETURNING id INTO log_id;

    RETURN log_id;
END;
$$ LANGUAGE plpgsql;

-- Function to update webhook status
CREATE OR REPLACE FUNCTION update_webhook_status(
    log_id_param UUID,
    status_param VARCHAR,
    error_message_param TEXT DEFAULT NULL
)
RETURNS void AS $$
BEGIN
    UPDATE webhook_logs
    SET
        status = status_param,
        error_message = error_message_param,
        processed_at = CASE
            WHEN status_param IN ('completed', 'failed') THEN CURRENT_TIMESTAMP
            ELSE processed_at
        END,
        updated_at = CURRENT_TIMESTAMP
    WHERE id = log_id_param;
END;
$$ LANGUAGE plpgsql;

-- Function to log audit events
CREATE OR REPLACE FUNCTION log_audit_event(
    user_id_param UUID,
    action_param VARCHAR,
    resource_type_param VARCHAR,
    resource_id_param UUID DEFAULT NULL,
    old_value_param JSONB DEFAULT NULL,
    new_value_param JSONB DEFAULT NULL,
    ip_address_param INET DEFAULT NULL,
    user_agent_param TEXT DEFAULT NULL
)
RETURNS UUID AS $$
DECLARE
    audit_id UUID;
BEGIN
    INSERT INTO audit_logs (
        user_id,
        action,
        resource_type,
        resource_id,
        old_value,
        new_value,
        ip_address,
        user_agent
    )
    VALUES (
        user_id_param,
        action_param,
        resource_type_param,
        resource_id_param,
        old_value_param,
        new_value_param,
        ip_address_param,
        user_agent_param
    )
    RETURNING id INTO audit_id;

    RETURN audit_id;
END;
$$ LANGUAGE plpgsql;

-- Function to get webhook retry queue
CREATE OR REPLACE FUNCTION get_webhook_retry_queue(limit_param INTEGER DEFAULT 100)
RETURNS TABLE (
    id UUID,
    event_type VARCHAR,
    workflow_id VARCHAR,
    execution_id VARCHAR,
    payload JSONB,
    created_at TIMESTAMP WITH TIME ZONE,
    retry_count BIGINT
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        wl.id,
        wl.event_type,
        wl.workflow_id,
        wl.execution_id,
        wl.payload,
        wl.created_at,
        COALESCE((wl.payload->>'retry_count')::BIGINT, 0) as retry_count
    FROM webhook_logs wl
    WHERE wl.status IN ('failed', 'retrying')
      AND wl.created_at >= CURRENT_TIMESTAMP - INTERVAL '24 hours'
      AND COALESCE((wl.payload->>'retry_count')::BIGINT, 0) < 3
    ORDER BY wl.created_at ASC
    LIMIT limit_param;
END;
$$ LANGUAGE plpgsql;

-- Function to get user audit trail
CREATE OR REPLACE FUNCTION get_user_audit_trail(
    user_id_param UUID,
    limit_param INTEGER DEFAULT 100,
    offset_param INTEGER DEFAULT 0
)
RETURNS TABLE (
    id UUID,
    action VARCHAR,
    resource_type VARCHAR,
    resource_id UUID,
    old_value JSONB,
    new_value JSONB,
    ip_address INET,
    created_at TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        al.id,
        al.action,
        al.resource_type,
        al.resource_id,
        al.old_value,
        al.new_value,
        al.ip_address,
        al.created_at
    FROM audit_logs al
    WHERE al.user_id = user_id_param
    ORDER BY al.created_at DESC
    LIMIT limit_param
    OFFSET offset_param;
END;
$$ LANGUAGE plpgsql;

-- Function to get resource audit trail
CREATE OR REPLACE FUNCTION get_resource_audit_trail(
    resource_type_param VARCHAR,
    resource_id_param UUID,
    limit_param INTEGER DEFAULT 100
)
RETURNS TABLE (
    id UUID,
    user_id UUID,
    action VARCHAR,
    old_value JSONB,
    new_value JSONB,
    created_at TIMESTAMP WITH TIME ZONE,
    user_email VARCHAR
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        al.id,
        al.user_id,
        al.action,
        al.old_value,
        al.new_value,
        al.created_at,
        u.email
    FROM audit_logs al
    LEFT JOIN users u ON al.user_id = u.id
    WHERE al.resource_type = resource_type_param
      AND al.resource_id = resource_id_param
    ORDER BY al.created_at DESC
    LIMIT limit_param;
END;
$$ LANGUAGE plpgsql;

-- Automated cleanup function for old webhook logs (keep 90 days)
CREATE OR REPLACE FUNCTION cleanup_old_webhook_logs()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM webhook_logs
    WHERE created_at < CURRENT_TIMESTAMP - INTERVAL '90 days'
      AND status IN ('completed', 'failed');

    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Comments for documentation
COMMENT ON TABLE webhook_logs IS 'n8n webhook event logging with retry tracking';
COMMENT ON TABLE audit_logs IS 'User action audit trail for compliance and security';
COMMENT ON COLUMN webhook_logs.status IS 'Processing status: pending, processing, completed, failed, retrying';
COMMENT ON COLUMN audit_logs.old_value IS 'Previous state of the resource (JSONB)';
COMMENT ON COLUMN audit_logs.new_value IS 'New state of the resource (JSONB)';
COMMENT ON FUNCTION log_webhook_event(VARCHAR, VARCHAR, VARCHAR, JSONB) IS 'Log incoming webhook event from n8n';
COMMENT ON FUNCTION update_webhook_status(UUID, VARCHAR, TEXT) IS 'Update webhook processing status and error message';
COMMENT ON FUNCTION log_audit_event(UUID, VARCHAR, VARCHAR, UUID, JSONB, JSONB, INET, TEXT) IS 'Log user action for audit trail';
COMMENT ON FUNCTION get_webhook_retry_queue(INTEGER) IS 'Get failed webhooks eligible for retry (< 3 attempts, < 24h old)';
COMMENT ON FUNCTION get_user_audit_trail(UUID, INTEGER, INTEGER) IS 'Get paginated audit trail for a specific user';
COMMENT ON FUNCTION get_resource_audit_trail(VARCHAR, UUID, INTEGER) IS 'Get audit trail for a specific resource';
COMMENT ON FUNCTION cleanup_old_webhook_logs() IS 'Delete webhook logs older than 90 days (completed/failed only)';
