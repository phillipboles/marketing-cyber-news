-- Migration 000004: Engagement Schema
-- Description: User engagement tracking (bookmarks, reads, statistics)
-- Author: Database Developer Agent
-- Date: 2025-12-11

-- Bookmarks table (many-to-many with composite primary key)
CREATE TABLE bookmarks (
    user_id UUID NOT NULL,
    article_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (user_id, article_id),

    CONSTRAINT fk_bookmarks_user FOREIGN KEY (user_id)
        REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_bookmarks_article FOREIGN KEY (article_id)
        REFERENCES articles(id) ON DELETE CASCADE
);

-- Article reads tracking
CREATE TABLE article_reads (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    article_id UUID NOT NULL,
    read_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    reading_time_seconds INTEGER NOT NULL DEFAULT 0,

    CONSTRAINT fk_article_reads_user FOREIGN KEY (user_id)
        REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_article_reads_article FOREIGN KEY (article_id)
        REFERENCES articles(id) ON DELETE CASCADE,
    CONSTRAINT chk_reading_time_positive CHECK (reading_time_seconds >= 0)
);

-- Daily statistics aggregation table
CREATE TABLE daily_stats (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    date DATE UNIQUE NOT NULL,
    total_articles INTEGER NOT NULL DEFAULT 0,
    critical_articles INTEGER NOT NULL DEFAULT 0,
    high_articles INTEGER NOT NULL DEFAULT 0,
    articles_by_category JSONB NOT NULL DEFAULT '{}'::JSONB,
    total_views INTEGER NOT NULL DEFAULT 0,
    unique_readers INTEGER NOT NULL DEFAULT 0,
    alert_matches INTEGER NOT NULL DEFAULT 0,
    new_users INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chk_total_articles_positive CHECK (total_articles >= 0),
    CONSTRAINT chk_critical_articles_positive CHECK (critical_articles >= 0),
    CONSTRAINT chk_high_articles_positive CHECK (high_articles >= 0),
    CONSTRAINT chk_total_views_positive CHECK (total_views >= 0),
    CONSTRAINT chk_unique_readers_positive CHECK (unique_readers >= 0),
    CONSTRAINT chk_alert_matches_positive CHECK (alert_matches >= 0),
    CONSTRAINT chk_new_users_positive CHECK (new_users >= 0)
);

-- Indexes for bookmarks table
CREATE INDEX idx_bookmarks_user_id ON bookmarks(user_id);
CREATE INDEX idx_bookmarks_article_id ON bookmarks(article_id);
CREATE INDEX idx_bookmarks_created_at ON bookmarks(created_at DESC);
CREATE INDEX idx_bookmarks_user_created ON bookmarks(user_id, created_at DESC);

-- Indexes for article_reads table
CREATE INDEX idx_article_reads_user_id ON article_reads(user_id);
CREATE INDEX idx_article_reads_article_id ON article_reads(article_id);
CREATE INDEX idx_article_reads_read_at ON article_reads(read_at DESC);
CREATE INDEX idx_article_reads_user_read_at ON article_reads(user_id, read_at DESC);
CREATE INDEX idx_article_reads_article_read_at ON article_reads(article_id, read_at DESC);

-- Indexes for daily_stats table
CREATE INDEX idx_daily_stats_date ON daily_stats(date DESC);
CREATE INDEX idx_daily_stats_total_articles ON daily_stats(total_articles DESC);

-- GIN index for JSONB queries
CREATE INDEX idx_daily_stats_articles_by_category ON daily_stats USING GIN(articles_by_category);

-- Apply updated_at trigger
CREATE TRIGGER update_daily_stats_updated_at
    BEFORE UPDATE ON daily_stats
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Function to record article read
CREATE OR REPLACE FUNCTION record_article_read(
    user_id_param UUID,
    article_id_param UUID,
    reading_time_param INTEGER DEFAULT 0
)
RETURNS UUID AS $$
DECLARE
    read_id UUID;
BEGIN
    -- Insert read record
    INSERT INTO article_reads (user_id, article_id, reading_time_seconds)
    VALUES (user_id_param, article_id_param, reading_time_param)
    RETURNING id INTO read_id;

    -- Increment article view count
    PERFORM increment_article_views(article_id_param);

    RETURN read_id;
END;
$$ LANGUAGE plpgsql;

-- Function to toggle bookmark
CREATE OR REPLACE FUNCTION toggle_bookmark(
    user_id_param UUID,
    article_id_param UUID
)
RETURNS BOOLEAN AS $$
DECLARE
    is_bookmarked BOOLEAN;
BEGIN
    -- Check if bookmark exists
    SELECT EXISTS(
        SELECT 1 FROM bookmarks
        WHERE user_id = user_id_param
          AND article_id = article_id_param
    ) INTO is_bookmarked;

    IF is_bookmarked THEN
        -- Remove bookmark
        DELETE FROM bookmarks
        WHERE user_id = user_id_param
          AND article_id = article_id_param;
        RETURN false;
    ELSE
        -- Add bookmark
        INSERT INTO bookmarks (user_id, article_id)
        VALUES (user_id_param, article_id_param)
        ON CONFLICT (user_id, article_id) DO NOTHING;
        RETURN true;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Function to generate daily statistics
CREATE OR REPLACE FUNCTION generate_daily_stats(target_date DATE DEFAULT CURRENT_DATE)
RETURNS UUID AS $$
DECLARE
    stats_id UUID;
    category_stats JSONB;
BEGIN
    -- Calculate articles by category
    SELECT jsonb_object_agg(c.name, article_count)
    INTO category_stats
    FROM (
        SELECT c.name, COUNT(a.id) as article_count
        FROM categories c
        LEFT JOIN articles a ON a.category_id = c.id
            AND DATE(a.created_at) = target_date
            AND a.is_published = true
        GROUP BY c.name
    ) c;

    -- Insert or update daily stats
    INSERT INTO daily_stats (
        date,
        total_articles,
        critical_articles,
        high_articles,
        articles_by_category,
        total_views,
        unique_readers,
        alert_matches,
        new_users
    )
    VALUES (
        target_date,
        -- Total articles published today
        (SELECT COUNT(*) FROM articles
         WHERE DATE(created_at) = target_date AND is_published = true),
        -- Critical articles
        (SELECT COUNT(*) FROM articles
         WHERE DATE(created_at) = target_date AND severity = 'critical' AND is_published = true),
        -- High severity articles
        (SELECT COUNT(*) FROM articles
         WHERE DATE(created_at) = target_date AND severity = 'high' AND is_published = true),
        -- Articles by category
        category_stats,
        -- Total views (reads) today
        (SELECT COUNT(*) FROM article_reads
         WHERE DATE(read_at) = target_date),
        -- Unique readers today
        (SELECT COUNT(DISTINCT user_id) FROM article_reads
         WHERE DATE(read_at) = target_date),
        -- Alert matches today
        (SELECT COUNT(*) FROM alert_matches
         WHERE DATE(matched_at) = target_date),
        -- New users registered today
        (SELECT COUNT(*) FROM users
         WHERE DATE(created_at) = target_date)
    )
    ON CONFLICT (date) DO UPDATE SET
        total_articles = EXCLUDED.total_articles,
        critical_articles = EXCLUDED.critical_articles,
        high_articles = EXCLUDED.high_articles,
        articles_by_category = EXCLUDED.articles_by_category,
        total_views = EXCLUDED.total_views,
        unique_readers = EXCLUDED.unique_readers,
        alert_matches = EXCLUDED.alert_matches,
        new_users = EXCLUDED.new_users,
        updated_at = CURRENT_TIMESTAMP
    RETURNING id INTO stats_id;

    RETURN stats_id;
END;
$$ LANGUAGE plpgsql;

-- Function to get user reading statistics
CREATE OR REPLACE FUNCTION get_user_reading_stats(user_id_param UUID)
RETURNS TABLE (
    total_reads BIGINT,
    total_bookmarks BIGINT,
    total_reading_time_seconds BIGINT,
    avg_reading_time_seconds NUMERIC,
    favorite_category VARCHAR,
    articles_this_week BIGINT,
    articles_this_month BIGINT
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        -- Total articles read
        (SELECT COUNT(*) FROM article_reads WHERE user_id = user_id_param),
        -- Total bookmarks
        (SELECT COUNT(*) FROM bookmarks WHERE user_id = user_id_param),
        -- Total reading time
        (SELECT COALESCE(SUM(reading_time_seconds), 0) FROM article_reads WHERE user_id = user_id_param),
        -- Average reading time
        (SELECT COALESCE(AVG(reading_time_seconds), 0) FROM article_reads WHERE user_id = user_id_param),
        -- Favorite category (most read)
        (SELECT c.name
         FROM article_reads ar
         JOIN articles a ON ar.article_id = a.id
         JOIN categories c ON a.category_id = c.id
         WHERE ar.user_id = user_id_param
         GROUP BY c.name
         ORDER BY COUNT(*) DESC
         LIMIT 1),
        -- Articles read this week
        (SELECT COUNT(*) FROM article_reads
         WHERE user_id = user_id_param
           AND read_at >= CURRENT_DATE - INTERVAL '7 days'),
        -- Articles read this month
        (SELECT COUNT(*) FROM article_reads
         WHERE user_id = user_id_param
           AND read_at >= CURRENT_DATE - INTERVAL '30 days');
END;
$$ LANGUAGE plpgsql;

-- Comments for documentation
COMMENT ON TABLE bookmarks IS 'User article bookmarks (saved for later)';
COMMENT ON TABLE article_reads IS 'Article read tracking with reading time metrics';
COMMENT ON TABLE daily_stats IS 'Daily aggregated statistics for analytics dashboard';
COMMENT ON COLUMN daily_stats.articles_by_category IS 'JSONB object mapping category names to article counts';
COMMENT ON FUNCTION record_article_read(UUID, UUID, INTEGER) IS 'Record article read and increment view count';
COMMENT ON FUNCTION toggle_bookmark(UUID, UUID) IS 'Add or remove bookmark (returns true if added, false if removed)';
COMMENT ON FUNCTION generate_daily_stats(DATE) IS 'Generate or update daily statistics for given date';
COMMENT ON FUNCTION get_user_reading_stats(UUID) IS 'Get comprehensive reading statistics for a user';
