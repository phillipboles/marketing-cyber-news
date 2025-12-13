-- Migration 000002: Content Schema (Categories, Sources, Articles)
-- Description: Core content management with semantic search and AI enrichment
-- Author: Database Developer Agent
-- Date: 2025-12-11

-- Categories table
CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) UNIQUE NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    color VARCHAR(7) NOT NULL DEFAULT '#6366f1',
    icon VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chk_slug_format CHECK (slug ~* '^[a-z0-9-]+$'),
    CONSTRAINT chk_color_hex CHECK (color ~* '^#[0-9a-f]{6}$'),
    CONSTRAINT chk_name_length CHECK (LENGTH(name) >= 2)
);

-- Sources table
CREATE TABLE sources (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) UNIQUE NOT NULL,
    url VARCHAR(500) UNIQUE NOT NULL,
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    trust_score DECIMAL(3,2) NOT NULL DEFAULT 0.70,
    last_scraped_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chk_url_format CHECK (url ~* '^https?://'),
    CONSTRAINT chk_trust_score_range CHECK (trust_score >= 0 AND trust_score <= 1),
    CONSTRAINT chk_name_length CHECK (LENGTH(name) >= 2)
);

-- Articles table with AI enrichment and semantic search
CREATE TABLE articles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(500) NOT NULL,
    slug VARCHAR(600) UNIQUE NOT NULL,
    content TEXT NOT NULL,
    summary TEXT,

    -- Foreign keys
    category_id UUID NOT NULL,
    source_id UUID NOT NULL,
    source_url VARCHAR(1000) UNIQUE NOT NULL,

    -- Classification
    severity VARCHAR(20) NOT NULL DEFAULT 'informational',
    tags TEXT[] NOT NULL DEFAULT '{}',
    cves TEXT[] NOT NULL DEFAULT '{}',
    vendors TEXT[] NOT NULL DEFAULT '{}',

    -- AI enrichment fields
    threat_type VARCHAR(100),
    attack_vector VARCHAR(100),
    impact_assessment TEXT,
    recommended_actions TEXT[] NOT NULL DEFAULT '{}',
    iocs JSONB NOT NULL DEFAULT '{}'::JSONB,

    -- Semantic search (OpenAI embeddings: 1536 dimensions)
    embedding vector(1536),
    search_vector tsvector GENERATED ALWAYS AS (
        setweight(to_tsvector('english', COALESCE(title, '')), 'A') ||
        setweight(to_tsvector('english', COALESCE(summary, '')), 'B') ||
        setweight(to_tsvector('english', COALESCE(content, '')), 'C')
    ) STORED,

    -- Armor Cybersecurity specific fields
    armor_relevance DECIMAL(3,2) NOT NULL DEFAULT 0.00,
    armor_cta JSONB,
    competitor_score DECIMAL(3,2) NOT NULL DEFAULT 0.00,
    is_competitor_favorable BOOLEAN NOT NULL DEFAULT false,

    -- Metrics
    reading_time_minutes INTEGER NOT NULL DEFAULT 0,
    view_count INTEGER NOT NULL DEFAULT 0,

    -- Publishing
    is_published BOOLEAN NOT NULL DEFAULT true,
    published_at TIMESTAMP WITH TIME ZONE,
    enriched_at TIMESTAMP WITH TIME ZONE,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Constraints
    CONSTRAINT fk_articles_category FOREIGN KEY (category_id)
        REFERENCES categories(id) ON DELETE RESTRICT,
    CONSTRAINT fk_articles_source FOREIGN KEY (source_id)
        REFERENCES sources(id) ON DELETE RESTRICT,
    CONSTRAINT chk_severity_valid CHECK (
        severity IN ('critical', 'high', 'medium', 'low', 'informational')
    ),
    CONSTRAINT chk_armor_relevance_range CHECK (
        armor_relevance >= 0 AND armor_relevance <= 1
    ),
    CONSTRAINT chk_competitor_score_range CHECK (
        competitor_score >= 0 AND competitor_score <= 1
    ),
    CONSTRAINT chk_reading_time_positive CHECK (reading_time_minutes >= 0),
    CONSTRAINT chk_view_count_positive CHECK (view_count >= 0),
    CONSTRAINT chk_title_length CHECK (LENGTH(title) >= 10),
    CONSTRAINT chk_slug_format CHECK (slug ~* '^[a-z0-9-]+$')
);

-- Indexes for categories table
CREATE INDEX idx_categories_slug ON categories(slug);
CREATE INDEX idx_categories_created_at ON categories(created_at DESC);

-- Indexes for sources table
CREATE INDEX idx_sources_name ON sources(name);
CREATE INDEX idx_sources_is_active ON sources(is_active);
CREATE INDEX idx_sources_trust_score ON sources(trust_score DESC);
CREATE INDEX idx_sources_last_scraped_at ON sources(last_scraped_at DESC NULLS LAST);

-- Indexes for articles table (performance critical)
CREATE INDEX idx_articles_category_id ON articles(category_id);
CREATE INDEX idx_articles_source_id ON articles(source_id);
CREATE INDEX idx_articles_severity ON articles(severity);
CREATE INDEX idx_articles_published_at ON articles(published_at DESC NULLS LAST);
CREATE INDEX idx_articles_created_at ON articles(created_at DESC);
CREATE INDEX idx_articles_is_published ON articles(is_published)
    WHERE is_published = true;

-- Composite indexes for common queries
CREATE INDEX idx_articles_category_published ON articles(category_id, published_at DESC)
    WHERE is_published = true;
CREATE INDEX idx_articles_severity_published ON articles(severity, published_at DESC)
    WHERE is_published = true;
CREATE INDEX idx_articles_armor_relevance ON articles(armor_relevance DESC)
    WHERE armor_relevance > 0;

-- GIN indexes for array fields (fast containment queries)
CREATE INDEX idx_articles_tags ON articles USING GIN(tags);
CREATE INDEX idx_articles_cves ON articles USING GIN(cves);
CREATE INDEX idx_articles_vendors ON articles USING GIN(vendors);
CREATE INDEX idx_articles_recommended_actions ON articles USING GIN(recommended_actions);

-- GIN index for JSONB fields
CREATE INDEX idx_articles_iocs ON articles USING GIN(iocs);
CREATE INDEX idx_articles_armor_cta ON articles USING GIN(armor_cta);

-- Full-text search index
CREATE INDEX idx_articles_search_vector ON articles USING GIN(search_vector);

-- Vector similarity search index (HNSW for approximate nearest neighbor)
-- m=16 (connections per layer), ef_construction=64 (search width during construction)
CREATE INDEX idx_articles_embedding ON articles USING hnsw(embedding vector_cosine_ops)
    WITH (m = 16, ef_construction = 64);

-- Apply updated_at triggers
CREATE TRIGGER update_categories_updated_at
    BEFORE UPDATE ON categories
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_sources_updated_at
    BEFORE UPDATE ON sources
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_articles_updated_at
    BEFORE UPDATE ON articles
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Function to update view count atomically
CREATE OR REPLACE FUNCTION increment_article_views(article_id UUID)
RETURNS void AS $$
BEGIN
    UPDATE articles
    SET view_count = view_count + 1
    WHERE id = article_id;
END;
$$ LANGUAGE plpgsql;

-- Comments for documentation
COMMENT ON TABLE categories IS 'Article categories (Vulnerabilities, Ransomware, etc.)';
COMMENT ON TABLE sources IS 'News sources with trust scoring';
COMMENT ON TABLE articles IS 'Main content table with AI enrichment and semantic search';
COMMENT ON COLUMN articles.embedding IS 'OpenAI text-embedding-3-small vector (1536 dimensions)';
COMMENT ON COLUMN articles.search_vector IS 'Full-text search vector (auto-generated from title/summary/content)';
COMMENT ON COLUMN articles.severity IS 'Threat severity: critical, high, medium, low, informational';
COMMENT ON COLUMN articles.iocs IS 'Indicators of Compromise (JSON): IPs, domains, hashes, etc.';
COMMENT ON COLUMN articles.armor_relevance IS 'Relevance score to Armor Cybersecurity (0.00-1.00)';
COMMENT ON COLUMN articles.competitor_score IS 'Competitor mention score (0.00-1.00)';
COMMENT ON INDEX idx_articles_embedding IS 'HNSW index for fast semantic similarity search';
