# Data Model: ACI Backend

**Date**: 2025-12-11
**Feature**: 001-aci-backend

---

## Entity Relationship Diagram

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│     users       │     │   categories    │     │    sources      │
├─────────────────┤     ├─────────────────┤     ├─────────────────┤
│ id (PK)         │     │ id (PK)         │     │ id (PK)         │
│ email           │     │ name            │     │ name            │
│ password_hash   │     │ slug            │     │ url             │
│ name            │     │ description     │     │ description     │
│ role            │     │ color           │     │ is_active       │
│ created_at      │     │ icon            │     │ trust_score     │
│ updated_at      │     │ created_at      │     │ created_at      │
└────────┬────────┘     └────────┬────────┘     └────────┬────────┘
         │                       │                       │
         │                       │                       │
    ┌────┴────┐            ┌─────┴─────┐           ┌─────┴─────┐
    │         │            │           │           │           │
    ▼         ▼            ▼           ▼           ▼           │
┌────────┐ ┌────────┐  ┌───────────────────────────────────┐   │
│refresh │ │user_   │  │           articles                │   │
│tokens  │ │prefs   │  ├───────────────────────────────────┤   │
└────────┘ └────────┘  │ id (PK)                           │◄──┘
                       │ title                             │
    ┌──────────────────│ slug                              │
    │                  │ content                           │
    │                  │ summary                           │
    │                  │ category_id (FK) ─────────────────┤
    │                  │ source_id (FK) ───────────────────┤
    │                  │ severity                          │
    │                  │ tags[]                            │
    │                  │ cves[]                            │
    │                  │ vendors[]                         │
    │                  │ embedding (vector)                │
    │                  │ armor_relevance                   │
    │                  │ competitor_score                  │
    │                  │ published_at                      │
    │                  │ created_at                        │
    │                  └───────────────────────────────────┘
    │                              │
    │                              │
    ▼                              ▼
┌────────────────┐         ┌────────────────┐
│    alerts      │         │   bookmarks    │
├────────────────┤         ├────────────────┤
│ id (PK)        │         │ user_id (FK)   │
│ user_id (FK)   │         │ article_id (FK)│
│ name           │         │ created_at     │
│ type           │         └────────────────┘
│ value          │
│ is_active      │         ┌────────────────┐
│ created_at     │         │ article_reads  │
└───────┬────────┘         ├────────────────┤
        │                  │ user_id (FK)   │
        ▼                  │ article_id (FK)│
┌────────────────┐         │ read_at        │
│ alert_matches  │         └────────────────┘
├────────────────┤
│ alert_id (FK)  │
│ article_id (FK)│
│ matched_at     │
│ priority       │
└────────────────┘
```

---

## Entities

### 1. User

**Table**: `users`

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| id | UUID | PK, DEFAULT uuid_generate_v4() | Unique identifier |
| email | VARCHAR(255) | UNIQUE, NOT NULL | User email address |
| password_hash | VARCHAR(255) | NOT NULL | bcrypt hash (cost 12) |
| name | VARCHAR(255) | NOT NULL | Display name |
| role | VARCHAR(50) | DEFAULT 'user' | 'user' or 'admin' |
| created_at | TIMESTAMPTZ | DEFAULT NOW() | Creation timestamp |
| updated_at | TIMESTAMPTZ | DEFAULT NOW() | Last update timestamp |

**Indexes**:
- `idx_users_email` on `email` (UNIQUE)

**Domain Entity** (`internal/domain/user.go`):
```go
type User struct {
    ID           uuid.UUID  `json:"id"`
    Email        string     `json:"email"`
    PasswordHash string     `json:"-"`
    Name         string     `json:"name"`
    Role         UserRole   `json:"role"`
    CreatedAt    time.Time  `json:"created_at"`
    UpdatedAt    time.Time  `json:"updated_at"`
}

type UserRole string
const (
    RoleUser  UserRole = "user"
    RoleAdmin UserRole = "admin"
)
```

---

### 2. Refresh Token

**Table**: `refresh_tokens`

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| id | UUID | PK | Unique identifier |
| user_id | UUID | FK users(id) ON DELETE CASCADE | Owner user |
| token_hash | VARCHAR(64) | NOT NULL | SHA-256 hash of token |
| expires_at | TIMESTAMPTZ | NOT NULL | Expiration time |
| revoked_at | TIMESTAMPTZ | NULL | Revocation time (if revoked) |
| created_at | TIMESTAMPTZ | DEFAULT NOW() | Creation timestamp |

**Indexes**:
- `idx_refresh_tokens_user_id` on `user_id`
- `idx_refresh_tokens_token_hash` on `token_hash`

---

### 3. User Preferences

**Table**: `user_preferences`

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| user_id | UUID | PK, FK users(id) ON DELETE CASCADE | User reference |
| preferred_categories | UUID[] | DEFAULT '{}' | Preferred category IDs |
| notification_frequency | VARCHAR(50) | DEFAULT 'realtime' | realtime/daily/weekly |
| email_notifications | BOOLEAN | DEFAULT true | Email alerts enabled |
| timezone | VARCHAR(100) | DEFAULT 'UTC' | User timezone |
| updated_at | TIMESTAMPTZ | DEFAULT NOW() | Last update |

---

### 4. Category

**Table**: `categories`

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| id | UUID | PK | Unique identifier |
| name | VARCHAR(100) | NOT NULL | Category name |
| slug | VARCHAR(100) | UNIQUE, NOT NULL | URL-friendly slug |
| description | TEXT | NULL | Category description |
| color | VARCHAR(7) | NOT NULL | Hex color code |
| icon | VARCHAR(50) | NULL | Icon identifier |
| created_at | TIMESTAMPTZ | DEFAULT NOW() | Creation timestamp |

**Seed Data** (8 categories):
```sql
INSERT INTO categories (id, name, slug, description, color) VALUES
('cat-001', 'Vulnerabilities', 'vulnerabilities', 'CVEs, security flaws, patches', '#EA580C'),
('cat-002', 'Ransomware', 'ransomware', 'Ransomware attacks and groups', '#DC2626'),
('cat-003', 'Data Breaches', 'data-breaches', 'Data leaks and exposures', '#7C3AED'),
('cat-004', 'Threat Actors', 'threat-actors', 'APT groups and campaigns', '#0891B2'),
('cat-005', 'Malware', 'malware', 'Trojans, worms, spyware', '#059669'),
('cat-006', 'Phishing', 'phishing', 'Social engineering attacks', '#D97706'),
('cat-007', 'Compliance', 'compliance', 'Regulations and standards', '#4F46E5'),
('cat-008', 'Industry News', 'industry-news', 'Market and vendor updates', '#6B7280');
```

---

### 5. Source

**Table**: `sources`

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| id | UUID | PK | Unique identifier |
| name | VARCHAR(255) | NOT NULL | Source name |
| url | VARCHAR(500) | UNIQUE, NOT NULL | Source base URL |
| description | TEXT | NULL | Source description |
| is_active | BOOLEAN | DEFAULT true | Source enabled |
| trust_score | DECIMAL(3,2) | DEFAULT 1.0 | Reliability score 0-1 |
| last_scraped_at | TIMESTAMPTZ | NULL | Last scrape time |
| created_at | TIMESTAMPTZ | DEFAULT NOW() | Creation timestamp |

**Seed Data** (10 sources):
```sql
INSERT INTO sources (id, name, url, description, trust_score) VALUES
('src-001', 'CISA', 'https://www.cisa.gov', 'US Cybersecurity Agency', 1.0),
('src-002', 'NVD', 'https://nvd.nist.gov', 'National Vulnerability Database', 1.0),
('src-003', 'Krebs on Security', 'https://krebsonsecurity.com', 'Brian Krebs blog', 0.95),
('src-004', 'BleepingComputer', 'https://www.bleepingcomputer.com', 'Tech news site', 0.90),
('src-005', 'The Hacker News', 'https://thehackernews.com', 'Security news', 0.85),
('src-006', 'Dark Reading', 'https://www.darkreading.com', 'Security news', 0.90),
('src-007', 'Threatpost', 'https://threatpost.com', 'Security news', 0.85),
('src-008', 'SecurityWeek', 'https://www.securityweek.com', 'Security news', 0.90),
('src-009', 'US-CERT', 'https://www.us-cert.gov', 'US-CERT alerts', 1.0),
('src-010', 'MITRE ATT&CK', 'https://attack.mitre.org', 'Threat framework', 1.0);
```

---

### 6. Article

**Table**: `articles`

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| id | UUID | PK | Unique identifier |
| title | VARCHAR(500) | NOT NULL | Article title |
| slug | VARCHAR(600) | UNIQUE, NOT NULL | URL-friendly slug |
| content | TEXT | NOT NULL | Full HTML/markdown content |
| summary | TEXT | NULL | Brief summary |
| category_id | UUID | FK categories(id) | Category reference |
| source_id | UUID | FK sources(id) | Source reference |
| source_url | VARCHAR(1000) | UNIQUE, NOT NULL | Original URL |
| severity | VARCHAR(20) | DEFAULT 'medium' | critical/high/medium/low/informational |
| tags | TEXT[] | DEFAULT '{}' | Content tags |
| cves | TEXT[] | DEFAULT '{}' | Related CVE IDs |
| vendors | TEXT[] | DEFAULT '{}' | Affected vendors |
| threat_type | VARCHAR(50) | NULL | AI-detected threat type |
| attack_vector | TEXT | NULL | AI-analyzed attack vector |
| impact_assessment | TEXT | NULL | AI impact analysis |
| recommended_actions | TEXT[] | DEFAULT '{}' | AI recommendations |
| iocs | JSONB | DEFAULT '[]' | Indicators of compromise |
| embedding | vector(1536) | NULL | pgvector embedding |
| search_vector | tsvector | GENERATED | Full-text search vector |
| armor_relevance | DECIMAL(3,2) | DEFAULT 0.0 | Armor.com relevance 0-1 |
| armor_cta | JSONB | NULL | CTA injection data |
| competitor_score | DECIMAL(3,2) | DEFAULT 0.0 | Competitor mention score |
| is_competitor_favorable | BOOLEAN | DEFAULT false | Block flag |
| reading_time_minutes | INTEGER | DEFAULT 5 | Estimated read time |
| view_count | INTEGER | DEFAULT 0 | View counter |
| is_published | BOOLEAN | DEFAULT true | Publication status |
| published_at | TIMESTAMPTZ | NOT NULL | Publication timestamp |
| enriched_at | TIMESTAMPTZ | NULL | AI enrichment timestamp |
| created_at | TIMESTAMPTZ | DEFAULT NOW() | Creation timestamp |
| updated_at | TIMESTAMPTZ | DEFAULT NOW() | Last update |

**Indexes**:
- `idx_articles_category_id` on `category_id`
- `idx_articles_source_id` on `source_id`
- `idx_articles_source_url` on `source_url` (UNIQUE)
- `idx_articles_severity` on `severity`
- `idx_articles_published_at` on `published_at DESC`
- `idx_articles_search_vector` GIN on `search_vector`
- `idx_articles_embedding` HNSW on `embedding vector_cosine_ops`
- `idx_articles_tags` GIN on `tags`
- `idx_articles_cves` GIN on `cves`
- `idx_articles_vendors` GIN on `vendors`

**Generated Column**:
```sql
search_vector tsvector GENERATED ALWAYS AS (
    setweight(to_tsvector('english', coalesce(title, '')), 'A') ||
    setweight(to_tsvector('english', coalesce(summary, '')), 'B') ||
    setweight(to_tsvector('english', coalesce(content, '')), 'C')
) STORED
```

**Domain Entity** (`internal/domain/article.go`):
```go
type Article struct {
    ID                   uuid.UUID         `json:"id"`
    Title                string            `json:"title"`
    Slug                 string            `json:"slug"`
    Content              string            `json:"content"`
    Summary              string            `json:"summary,omitempty"`
    CategoryID           uuid.UUID         `json:"category_id"`
    Category             *Category         `json:"category,omitempty"`
    SourceID             uuid.UUID         `json:"source_id"`
    Source               *Source           `json:"source,omitempty"`
    SourceURL            string            `json:"source_url"`
    Severity             Severity          `json:"severity"`
    Tags                 []string          `json:"tags"`
    CVEs                 []string          `json:"cves"`
    Vendors              []string          `json:"vendors"`
    ThreatType           string            `json:"threat_type,omitempty"`
    AttackVector         string            `json:"attack_vector,omitempty"`
    ImpactAssessment     string            `json:"impact_assessment,omitempty"`
    RecommendedActions   []string          `json:"recommended_actions,omitempty"`
    IOCs                 []IOC             `json:"iocs,omitempty"`
    ArmorRelevance       float64           `json:"armor_relevance"`
    ArmorCTA             *ArmorCTA         `json:"armor_cta,omitempty"`
    CompetitorScore      float64           `json:"-"`
    IsCompetitorFavorable bool             `json:"-"`
    ReadingTimeMinutes   int               `json:"reading_time_minutes"`
    ViewCount            int               `json:"view_count"`
    IsPublished          bool              `json:"is_published"`
    PublishedAt          time.Time         `json:"published_at"`
    EnrichedAt           *time.Time        `json:"enriched_at,omitempty"`
    CreatedAt            time.Time         `json:"created_at"`
    UpdatedAt            time.Time         `json:"updated_at"`
}

type Severity string
const (
    SeverityCritical      Severity = "critical"
    SeverityHigh          Severity = "high"
    SeverityMedium        Severity = "medium"
    SeverityLow           Severity = "low"
    SeverityInformational Severity = "informational"
)

type IOC struct {
    Type    string `json:"type"`    // ip, domain, hash, url
    Value   string `json:"value"`
    Context string `json:"context,omitempty"`
}

type ArmorCTA struct {
    Type  string `json:"type"`
    Title string `json:"title"`
    URL   string `json:"url"`
}
```

---

### 7. Alert

**Table**: `alerts`

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| id | UUID | PK | Unique identifier |
| user_id | UUID | FK users(id) ON DELETE CASCADE | Owner user |
| name | VARCHAR(255) | NOT NULL | Alert name |
| type | VARCHAR(50) | NOT NULL | keyword/category/severity/vendor/cve |
| value | VARCHAR(500) | NOT NULL | Match value |
| is_active | BOOLEAN | DEFAULT true | Alert enabled |
| created_at | TIMESTAMPTZ | DEFAULT NOW() | Creation timestamp |
| updated_at | TIMESTAMPTZ | DEFAULT NOW() | Last update |

**Indexes**:
- `idx_alerts_user_id` on `user_id`
- `idx_alerts_type_value` on `type, value`

**Domain Entity**:
```go
type Alert struct {
    ID        uuid.UUID  `json:"id"`
    UserID    uuid.UUID  `json:"user_id"`
    Name      string     `json:"name"`
    Type      AlertType  `json:"type"`
    Value     string     `json:"value"`
    IsActive  bool       `json:"is_active"`
    CreatedAt time.Time  `json:"created_at"`
    UpdatedAt time.Time  `json:"updated_at"`
}

type AlertType string
const (
    AlertTypeKeyword  AlertType = "keyword"
    AlertTypeCategory AlertType = "category"
    AlertTypeSeverity AlertType = "severity"
    AlertTypeVendor   AlertType = "vendor"
    AlertTypeCVE      AlertType = "cve"
)
```

---

### 8. Alert Match

**Table**: `alert_matches`

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| id | UUID | PK | Unique identifier |
| alert_id | UUID | FK alerts(id) ON DELETE CASCADE | Alert reference |
| article_id | UUID | FK articles(id) ON DELETE CASCADE | Article reference |
| priority | VARCHAR(20) | DEFAULT 'normal' | critical/high/normal |
| matched_at | TIMESTAMPTZ | DEFAULT NOW() | Match timestamp |
| notified_at | TIMESTAMPTZ | NULL | Notification sent timestamp |

**Indexes**:
- `idx_alert_matches_alert_id` on `alert_id`
- `idx_alert_matches_article_id` on `article_id`
- UNIQUE constraint on `(alert_id, article_id)`

---

### 9. Bookmark

**Table**: `bookmarks`

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| user_id | UUID | FK users(id) ON DELETE CASCADE | User reference |
| article_id | UUID | FK articles(id) ON DELETE CASCADE | Article reference |
| created_at | TIMESTAMPTZ | DEFAULT NOW() | Bookmark timestamp |

**Primary Key**: `(user_id, article_id)`

---

### 10. Article Read

**Table**: `article_reads`

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| id | UUID | PK | Unique identifier |
| user_id | UUID | FK users(id) ON DELETE CASCADE | User reference |
| article_id | UUID | FK articles(id) ON DELETE CASCADE | Article reference |
| read_at | TIMESTAMPTZ | DEFAULT NOW() | Read timestamp |
| reading_time_seconds | INTEGER | NULL | Actual reading time |

**Indexes**:
- `idx_article_reads_user_id` on `user_id`
- `idx_article_reads_article_id` on `article_id`
- `idx_article_reads_read_at` on `read_at DESC`

---

### 11. Daily Stats

**Table**: `daily_stats`

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| id | UUID | PK | Unique identifier |
| date | DATE | UNIQUE, NOT NULL | Stats date |
| total_articles | INTEGER | DEFAULT 0 | Articles published |
| critical_articles | INTEGER | DEFAULT 0 | Critical severity count |
| high_articles | INTEGER | DEFAULT 0 | High severity count |
| articles_by_category | JSONB | DEFAULT '{}' | {category_slug: count} |
| total_views | INTEGER | DEFAULT 0 | Total article views |
| unique_readers | INTEGER | DEFAULT 0 | Unique users who read |
| alert_matches | INTEGER | DEFAULT 0 | Total alert matches |
| new_users | INTEGER | DEFAULT 0 | New registrations |
| created_at | TIMESTAMPTZ | DEFAULT NOW() | Creation timestamp |

---

### 12. Webhook Log

**Table**: `webhook_logs`

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| id | UUID | PK | Unique identifier |
| event_type | VARCHAR(50) | NOT NULL | Event type |
| workflow_id | VARCHAR(100) | NULL | n8n workflow ID |
| execution_id | VARCHAR(100) | NULL | n8n execution ID |
| payload | JSONB | NOT NULL | Request payload |
| status | VARCHAR(20) | NOT NULL | pending/processing/success/failed |
| error_message | TEXT | NULL | Error details if failed |
| processed_at | TIMESTAMPTZ | NULL | Processing completion time |
| created_at | TIMESTAMPTZ | DEFAULT NOW() | Receipt timestamp |

**Indexes**:
- `idx_webhook_logs_event_type` on `event_type`
- `idx_webhook_logs_status` on `status`
- `idx_webhook_logs_created_at` on `created_at DESC`

---

### 13. Audit Log

**Table**: `audit_logs`

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| id | UUID | PK | Unique identifier |
| user_id | UUID | FK users(id) ON DELETE SET NULL | Actor user (if known) |
| action | VARCHAR(100) | NOT NULL | Action performed |
| resource_type | VARCHAR(50) | NOT NULL | Target resource type |
| resource_id | UUID | NULL | Target resource ID |
| old_value | JSONB | NULL | Previous state |
| new_value | JSONB | NULL | New state |
| ip_address | INET | NULL | Client IP |
| user_agent | TEXT | NULL | Client user agent |
| created_at | TIMESTAMPTZ | DEFAULT NOW() | Action timestamp |

**Indexes**:
- `idx_audit_logs_user_id` on `user_id`
- `idx_audit_logs_action` on `action`
- `idx_audit_logs_resource` on `resource_type, resource_id`
- `idx_audit_logs_created_at` on `created_at DESC`

---

## Validation Rules

### User
- email: required, valid email format, max 255 chars
- password: required, min 8 chars, max 128 chars, 1 upper, 1 lower, 1 digit
- name: required, min 2 chars, max 255 chars
- role: enum (user, admin)

### Article
- title: required, max 500 chars
- content: required
- source_url: required, valid URL, unique
- severity: enum (critical, high, medium, low, informational)
- category_id: required, must exist
- source_id: required, must exist
- tags: array of strings, each max 50 chars
- cves: array of strings, CVE-YYYY-NNNNN format

### Alert
- name: required, max 255 chars
- type: enum (keyword, category, severity, vendor, cve)
- value: required, max 500 chars

---

## Migration Files

### 001_initial_schema.sql
- Enable extensions: uuid-ossp, pgvector
- Create users table
- Create refresh_tokens table
- Create user_preferences table

### 002_content_schema.sql
- Create categories table
- Create sources table
- Create articles table with all indexes
- Seed categories and sources

### 003_alerts_schema.sql
- Create alerts table
- Create alert_matches table

### 004_engagement_schema.sql
- Create bookmarks table
- Create article_reads table
- Create daily_stats table

### 005_audit_schema.sql
- Create webhook_logs table
- Create audit_logs table
