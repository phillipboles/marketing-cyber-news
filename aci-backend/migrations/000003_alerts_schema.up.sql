-- Migration 000003: Alerts Schema
-- Description: User alert configuration and matching system
-- Author: Database Developer Agent
-- Date: 2025-12-11

-- Alerts table (user-defined alert configurations)
CREATE TABLE alerts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    value TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_alerts_user FOREIGN KEY (user_id)
        REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT chk_alert_type_valid CHECK (
        type IN ('keyword', 'cve', 'vendor', 'category', 'severity', 'source')
    ),
    CONSTRAINT chk_name_length CHECK (LENGTH(name) >= 2),
    CONSTRAINT chk_value_not_empty CHECK (LENGTH(value) >= 1)
);

-- Alert matches table (track when articles match user alerts)
CREATE TABLE alert_matches (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    alert_id UUID NOT NULL,
    article_id UUID NOT NULL,
    priority VARCHAR(20) NOT NULL DEFAULT 'medium',
    matched_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    notified_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT fk_alert_matches_alert FOREIGN KEY (alert_id)
        REFERENCES alerts(id) ON DELETE CASCADE,
    CONSTRAINT fk_alert_matches_article FOREIGN KEY (article_id)
        REFERENCES articles(id) ON DELETE CASCADE,
    CONSTRAINT chk_priority_valid CHECK (
        priority IN ('critical', 'high', 'medium', 'low')
    ),
    CONSTRAINT uq_alert_matches_alert_article UNIQUE (alert_id, article_id)
);

-- Indexes for alerts table
CREATE INDEX idx_alerts_user_id ON alerts(user_id);
CREATE INDEX idx_alerts_type ON alerts(type);
CREATE INDEX idx_alerts_is_active ON alerts(is_active)
    WHERE is_active = true;
CREATE INDEX idx_alerts_user_active ON alerts(user_id, is_active)
    WHERE is_active = true;

-- Indexes for alert_matches table
CREATE INDEX idx_alert_matches_alert_id ON alert_matches(alert_id);
CREATE INDEX idx_alert_matches_article_id ON alert_matches(article_id);
CREATE INDEX idx_alert_matches_matched_at ON alert_matches(matched_at DESC);
CREATE INDEX idx_alert_matches_priority ON alert_matches(priority);
CREATE INDEX idx_alert_matches_unnotified ON alert_matches(alert_id, matched_at DESC)
    WHERE notified_at IS NULL;

-- Composite index for user notification queries
CREATE INDEX idx_alert_matches_user_unnotified ON alert_matches(alert_id, priority, matched_at DESC)
    WHERE notified_at IS NULL;

-- Apply updated_at trigger
CREATE TRIGGER update_alerts_updated_at
    BEFORE UPDATE ON alerts
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Function to match article against active alerts
CREATE OR REPLACE FUNCTION match_article_alerts(article_id_param UUID)
RETURNS void AS $$
DECLARE
    article_record RECORD;
    alert_record RECORD;
    match_priority VARCHAR(20);
BEGIN
    -- Get article details
    SELECT a.*, c.slug as category_slug, s.name as source_name
    INTO article_record
    FROM articles a
    JOIN categories c ON a.category_id = c.id
    JOIN sources s ON a.source_id = s.id
    WHERE a.id = article_id_param;

    IF NOT FOUND THEN
        RETURN;
    END IF;

    -- Loop through active alerts and check for matches
    FOR alert_record IN
        SELECT * FROM alerts WHERE is_active = true
    LOOP
        match_priority := NULL;

        -- Keyword alerts (case-insensitive search in title, summary, content)
        IF alert_record.type = 'keyword' AND (
            article_record.title ILIKE '%' || alert_record.value || '%' OR
            article_record.summary ILIKE '%' || alert_record.value || '%' OR
            article_record.content ILIKE '%' || alert_record.value || '%'
        ) THEN
            match_priority := CASE article_record.severity
                WHEN 'critical' THEN 'critical'
                WHEN 'high' THEN 'high'
                ELSE 'medium'
            END;
        END IF;

        -- CVE alerts (exact match in cves array)
        IF alert_record.type = 'cve' AND alert_record.value = ANY(article_record.cves) THEN
            match_priority := 'critical';
        END IF;

        -- Vendor alerts (case-insensitive match in vendors array)
        IF alert_record.type = 'vendor' THEN
            FOR i IN 1..array_length(article_record.vendors, 1) LOOP
                IF article_record.vendors[i] ILIKE alert_record.value THEN
                    match_priority := 'high';
                    EXIT;
                END IF;
            END LOOP;
        END IF;

        -- Category alerts
        IF alert_record.type = 'category' AND alert_record.value = article_record.category_slug THEN
            match_priority := 'medium';
        END IF;

        -- Severity alerts
        IF alert_record.type = 'severity' AND alert_record.value = article_record.severity THEN
            match_priority := alert_record.value::VARCHAR;
        END IF;

        -- Source alerts
        IF alert_record.type = 'source' AND alert_record.value = article_record.source_name THEN
            match_priority := 'medium';
        END IF;

        -- Insert match if found (ON CONFLICT DO NOTHING for idempotency)
        IF match_priority IS NOT NULL THEN
            INSERT INTO alert_matches (alert_id, article_id, priority)
            VALUES (alert_record.id, article_id_param, match_priority)
            ON CONFLICT (alert_id, article_id) DO NOTHING;
        END IF;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- Function to mark alert matches as notified
CREATE OR REPLACE FUNCTION mark_alerts_notified(alert_match_ids UUID[])
RETURNS INTEGER AS $$
DECLARE
    updated_count INTEGER;
BEGIN
    UPDATE alert_matches
    SET notified_at = CURRENT_TIMESTAMP
    WHERE id = ANY(alert_match_ids)
      AND notified_at IS NULL;

    GET DIAGNOSTICS updated_count = ROW_COUNT;
    RETURN updated_count;
END;
$$ LANGUAGE plpgsql;

-- Comments for documentation
COMMENT ON TABLE alerts IS 'User-defined alert configurations for article monitoring';
COMMENT ON TABLE alert_matches IS 'Matched articles for user alerts with notification tracking';
COMMENT ON COLUMN alerts.type IS 'Alert type: keyword, cve, vendor, category, severity, source';
COMMENT ON COLUMN alert_matches.priority IS 'Match priority derived from article severity and alert type';
COMMENT ON FUNCTION match_article_alerts(UUID) IS 'Match a new article against all active user alerts';
COMMENT ON FUNCTION mark_alerts_notified(UUID[]) IS 'Mark alert matches as notified (bulk operation)';
