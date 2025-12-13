# Feature Specification: ACI Backend Service

**Feature Branch**: `001-aci-backend`
**Created**: 2025-12-11
**Status**: Draft
**Input**: Armor Cyber Intelligence (ACI) backend service for aggregating, enriching, and distributing cybersecurity news content

## Overview

The ACI Backend is a cybersecurity news aggregation and intelligence platform that collects articles from trusted sources (CISA, NVD, etc.) via n8n workflows, enriches them with AI analysis, filters competitor content, and delivers real-time updates to security professionals through REST APIs and WebSockets.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - User Authentication (Priority: P1)

A security analyst wants to create an account and log in to access personalized cybersecurity news feeds, so they can stay informed about threats relevant to their organization.

**Why this priority**: Authentication is foundational - no other features work without it. Users cannot access articles, create alerts, or receive personalized content without first authenticating.

**Independent Test**: Can be fully tested by registering a new user, logging in, receiving JWT tokens, and accessing a protected endpoint.

**Acceptance Scenarios**:

1. **Given** a new user with valid email and password, **When** they submit the registration form, **Then** they receive access and refresh tokens and can access protected resources.
2. **Given** a registered user with correct credentials, **When** they submit login, **Then** they receive valid JWT tokens within 2 seconds.
3. **Given** an authenticated user with an expiring access token, **When** they submit a refresh request with valid refresh token, **Then** they receive new tokens without re-entering credentials.
4. **Given** an authenticated user, **When** they log out, **Then** their refresh token is invalidated and cannot be reused.

---

### User Story 2 - Browse Cybersecurity Articles (Priority: P1)

A security analyst wants to browse and search cybersecurity news articles filtered by category, severity, and keywords so they can quickly find relevant threat intelligence.

**Why this priority**: Content browsing is the core value proposition - users come to the platform specifically to consume cybersecurity news.

**Independent Test**: Can be fully tested by querying the articles endpoint with various filters and verifying correct results are returned.

**Acceptance Scenarios**:

1. **Given** the system has articles, **When** a user requests the article list with pagination, **Then** they receive a paginated list of articles sorted by publish date (newest first).
2. **Given** articles exist in multiple categories, **When** a user filters by category "vulnerabilities", **Then** only vulnerability articles are returned.
3. **Given** a critical CVE article exists, **When** a user searches for that CVE ID, **Then** the matching article appears in search results.
4. **Given** a user requests a specific article by slug, **When** the article exists, **Then** the full article content is returned with enrichment data.

---

### User Story 3 - Receive Real-time Updates (Priority: P2)

A security analyst wants to receive instant notifications when critical vulnerabilities or threats matching their interests are published, so they can respond quickly to emerging threats.

**Why this priority**: Real-time updates differentiate this platform from static news sites but depend on P1 content infrastructure being in place.

**Independent Test**: Can be fully tested by connecting to WebSocket, subscribing to channels, and verifying messages arrive when new articles are created.

**Acceptance Scenarios**:

1. **Given** an authenticated user connected via WebSocket, **When** they subscribe to "articles:critical", **Then** they receive notifications for all new critical-severity articles.
2. **Given** a user subscribed to "articles:category:ransomware", **When** a ransomware article is published, **Then** they receive the article notification within 5 seconds.
3. **Given** a user with an expiring JWT, **When** the server sends a token_expiring warning, **Then** the user can refresh and re-authenticate without disconnecting.

---

### User Story 4 - Create and Manage Alert Subscriptions (Priority: P2)

A security analyst wants to create custom alerts for specific keywords, vendors, or CVEs, so they are notified immediately when matching content is published.

**Why this priority**: Personalized alerts add significant value but require both authentication (P1) and real-time infrastructure (P2) to function.

**Independent Test**: Can be fully tested by creating an alert, publishing a matching article via webhook, and verifying the alert match notification is delivered.

**Acceptance Scenarios**:

1. **Given** an authenticated user, **When** they create an alert for vendor "VMware", **Then** the alert is saved and active.
2. **Given** a user has an active VMware alert, **When** an article mentioning VMware is published, **Then** the user receives an alert_match notification.
3. **Given** a user has multiple alerts, **When** they list their alerts, **Then** all alerts are returned with match statistics.
4. **Given** a user no longer wants an alert, **When** they delete it, **Then** the alert stops matching and is removed.

---

### User Story 5 - n8n Webhook Content Ingestion (Priority: P1)

The n8n workflow system wants to push scraped cybersecurity articles to the ACI backend via webhooks, so content is automatically ingested and available to users.

**Why this priority**: Without content ingestion, there are no articles to browse - this is the content pipeline that feeds the entire system.

**Independent Test**: Can be fully tested by sending a properly signed webhook with article data and verifying the article appears in the database and is queryable via API.

**Acceptance Scenarios**:

1. **Given** a valid webhook payload with HMAC signature, **When** the n8n workflow sends an article.created event, **Then** the article is stored and assigned a slug.
2. **Given** a bulk.import event with 50 articles, **When** submitted via webhook, **Then** all non-duplicate articles are queued for processing.
3. **Given** an invalid HMAC signature, **When** a webhook is received, **Then** it is rejected with 401 Unauthorized.
4. **Given** an article.created event without skip_enrichment flag, **When** processed, **Then** the article is queued for AI enrichment.

---

### User Story 6 - AI Content Enrichment (Priority: P3)

The system wants to automatically enrich articles with AI-generated threat analysis, so users receive actionable intelligence beyond raw news.

**Why this priority**: AI enrichment adds significant value but is not blocking for basic functionality - articles can be displayed without enrichment.

**Independent Test**: Can be fully tested by triggering enrichment for an article and verifying the enrichment fields (threat_type, recommended_actions, etc.) are populated.

**Acceptance Scenarios**:

1. **Given** an article pending enrichment, **When** the AI enrichment workflow completes, **Then** the article is updated with threat analysis fields.
2. **Given** an enrichment.complete webhook, **When** received with valid signature, **Then** the article's enrichment data is stored.
3. **Given** AI enrichment fails, **When** the system detects the failure, **Then** the article remains available without enrichment data.

---

### User Story 7 - Bookmarks and Read Tracking (Priority: P3)

A security analyst wants to bookmark articles and track their reading history, so they can reference important content later and see what they've already read.

**Why this priority**: Nice-to-have engagement features that enhance UX but are not core to the threat intelligence mission.

**Independent Test**: Can be fully tested by bookmarking an article, retrieving bookmarks list, and marking articles as read.

**Acceptance Scenarios**:

1. **Given** an authenticated user viewing an article, **When** they bookmark it, **Then** the article appears in their bookmarks list.
2. **Given** a user has bookmarks, **When** they request their bookmark list, **Then** all bookmarked articles are returned with bookmark timestamps.
3. **Given** a user reads an article, **When** they mark it as read, **Then** the read status is tracked for analytics.

---

### User Story 8 - Admin Content Management (Priority: P3)

An admin user wants to manage articles, sources, and users through administrative endpoints, so they can maintain platform quality and moderate content.

**Why this priority**: Administrative functions are needed for operations but not for initial user value delivery.

**Independent Test**: Can be fully tested by an admin user creating/updating/deleting articles and managing sources.

**Acceptance Scenarios**:

1. **Given** an admin user, **When** they update an article's severity, **Then** the change is persisted and reflected in APIs.
2. **Given** an admin user, **When** they disable a source, **Then** no new articles from that source are ingested.
3. **Given** a non-admin user, **When** they attempt admin actions, **Then** they receive 403 Forbidden.

---

### Edge Cases

- What happens when a user submits a duplicate article (same source_url)? System skips duplicate, returns success.
- How does the system handle webhook payloads exceeding size limits? Returns 400 Bad Request with clear error.
- What happens when WebSocket connection limit (5 per user) is exceeded? Oldest connection is gracefully closed.
- How does the system handle n8n being unavailable? Graceful degradation - content still viewable, new ingestion queued.
- What happens when a JWT expires mid-WebSocket session? Server sends token_expiring warning 60s before, client can refresh.
- How does the system handle malformed webhook signatures? Returns 401 without processing, logs attempt.
- What happens when pgvector semantic search fails? Falls back to full-text search.

## Requirements *(mandatory)*

### Functional Requirements

#### Authentication & Authorization
- **FR-001**: System MUST allow users to register with email, password, and display name
- **FR-002**: System MUST authenticate users via JWT with RS256 signing (15-minute access tokens, 7-day refresh tokens)
- **FR-003**: System MUST implement refresh token rotation (each refresh invalidates the previous token)
- **FR-004**: System MUST support two roles: "user" and "admin" with appropriate permission boundaries
- **FR-005**: System MUST enforce rate limiting on auth endpoints (5 login attempts per minute per IP)

#### Content Management
- **FR-006**: System MUST store articles with title, content, summary, category, severity, tags, source_url, and metadata
- **FR-007**: System MUST automatically generate URL-friendly slugs for articles
- **FR-008**: System MUST support article filtering by category, severity, tags, source, date range, and CVE
- **FR-009**: System MUST support full-text search across article titles, content, and summaries
- **FR-010**: System MUST support semantic search using vector embeddings when enabled
- **FR-011**: System MUST paginate article results (default 20, max 100 per page)

#### Content Ingestion
- **FR-012**: System MUST accept webhooks from n8n with HMAC-SHA256 signature verification
- **FR-013**: System MUST support article.created, article.updated, article.deleted, bulk.import, and enrichment.complete events
- **FR-014**: System MUST detect and filter competitor-promotional content based on configurable keyword lists
- **FR-015**: System MUST calculate Armor.com relevance scores for CTA injection opportunities
- **FR-016**: System MUST deduplicate articles by source_url

#### Real-time Updates
- **FR-017**: System MUST provide WebSocket connections for real-time article and alert notifications
- **FR-018**: System MUST support channel subscriptions: articles:all, articles:critical, articles:high, articles:category:{slug}, articles:vendor:{name}, alerts:user, system
- **FR-019**: System MUST limit WebSocket subscriptions to 50 channels per connection and 5 connections per user
- **FR-020**: System MUST implement ping/pong heartbeat (30-second interval, 60-second timeout)
- **FR-021**: System MUST send token_expiring warnings 60 seconds before JWT expiration

#### Alerts
- **FR-022**: System MUST allow users to create alert subscriptions for keywords, vendors, CVEs, categories, and severity levels
- **FR-023**: System MUST match new articles against active alerts and notify users in real-time
- **FR-024**: System MUST track alert match history for analytics

#### User Engagement
- **FR-025**: System MUST allow users to bookmark articles
- **FR-026**: System MUST track article read status per user
- **FR-027**: System MUST provide user preference storage

#### AI Integration
- **FR-028**: System MUST integrate with Claude API for content enrichment
- **FR-029**: System MUST generate threat analysis including threat_type, attack_vector, impact_assessment, recommended_actions
- **FR-030**: System MUST extract and store indicators of compromise (IOCs) from enriched content

#### Administration
- **FR-031**: System MUST provide admin endpoints for article CRUD operations
- **FR-032**: System MUST provide admin endpoints for source management
- **FR-033**: System MUST maintain audit logs for admin actions

#### Observability
- **FR-034**: System MUST expose /health and /ready endpoints
- **FR-035**: System MUST use structured JSON logging with correlation IDs
- **FR-036**: System MUST emit Prometheus-compatible metrics

### Key Entities

- **User**: Account holder who consumes content and creates alerts. Has email, password hash, name, role, preferences.
- **Article**: Cybersecurity news content from trusted sources. Has title, content, summary, category, severity, tags, CVEs, vendors, enrichment data.
- **Category**: Content classification (vulnerabilities, ransomware, data-breaches, etc.). Has name, slug, color, icon.
- **Source**: Trusted content origin (CISA, NVD, etc.). Has name, URL, trust level, active status.
- **Alert**: User-defined notification subscription. Has name, type (keyword/vendor/cve/category), value, enabled status.
- **AlertMatch**: Record of alert matching an article. Links alert, article, and match timestamp.
- **Bookmark**: User's saved article reference. Links user and article with created timestamp.
- **RefreshToken**: Token for obtaining new access tokens. Has token hash, expiry, user reference.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can complete registration and login in under 10 seconds
- **SC-002**: Article search returns results in under 1 second for queries up to 1000 articles
- **SC-003**: WebSocket notifications arrive within 5 seconds of article publication
- **SC-004**: System handles 1000 concurrent REST API requests without degradation
- **SC-005**: System supports 500 concurrent WebSocket connections
- **SC-006**: n8n webhooks are processed within 2 seconds (202 Accepted response)
- **SC-007**: 99% of articles have AI enrichment completed within 5 minutes of ingestion
- **SC-008**: Alert matching correctly identifies 100% of articles matching user-defined criteria
- **SC-009**: Zero security vulnerabilities in OWASP Top 10 categories
- **SC-010**: 80% or higher test coverage across all modules

## PM Acceptance Criteria *(mandatory)*

*Per Constitution Principle XVI - Product Manager Ownership*

### PM-1 Gate: Pre-Implementation Approval

- [x] All user stories have clear acceptance scenarios
- [x] Priorities (P1, P2, P3) are assigned and justified
- [x] Edge cases are identified and documented
- [x] Success metrics are measurable and achievable
- [x] Out-of-scope items are explicitly declared (see below)
- [ ] Gap analysis from PM review has been addressed (Critical items resolved)

### PM-2 Gate: Mid-Implementation Alignment

- [ ] Feature implementation aligns with original scope
- [ ] No scope creep has occurred (or changes are documented/approved)
- [ ] P1 user stories are functional and testable
- [ ] Risks identified during implementation are tracked

### PM-3 Gate: Pre-Release Verification

- [ ] All acceptance scenarios pass
- [ ] User journeys validated end-to-end
- [ ] Documentation is complete and accurate
- [ ] Performance targets met
- [ ] Security requirements validated
- [ ] Product verification checklist completed

## Out of Scope

The following are explicitly **NOT** included in this feature:

- Frontend/UI implementation (separate project)
- Email notifications (future enhancement)
- Mobile push notifications (future enhancement)
- Multi-tenancy/organization support (future enhancement)
- SSO/SAML integration (future enhancement)
- Article comments/discussion features
- Content recommendation engine beyond relevance scoring
- Historical data migration from existing systems

## Assumptions

- n8n is pre-configured and running with scraper workflows
- PostgreSQL 18+ with pgvector extension is available
- Anthropic Claude API access is provisioned
- JWT RSA key pairs are generated and accessible
- Redis is optional (caching enhancement, not required)
- Single-tenant deployment (one organization per instance)

## Dependencies

- **n8n 2.0.0+**: Workflow automation for content scraping
- **PostgreSQL 18+**: Primary database with pgvector 0.8.1 extension
- **Claude API**: AI enrichment (claude-3-haiku model)
- **External Sources**: CISA, NVD, security news feeds via n8n scrapers

## Risks

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Claude API rate limits | Medium | Medium | Queue system with rate limiting, caching |
| pgvector performance at scale | High | Medium | Index tuning, fallback to full-text |
| WebSocket connection limits | High | Low | Horizontal scaling, connection pooling |
| n8n webhook reliability | Medium | Low | Retry logic, dead letter queue |
| Competitor detection false positives | Low | Medium | Manual review flag, configurable thresholds |
