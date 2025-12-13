# Implementation Plan: ACI Backend (Armor Cyber Intelligence)

**Branch**: `001-aci-backend` | **Date**: 2025-12-11 | **Spec**: /specs/001-aci-backend/
**Input**: Feature specifications from `/specs/*.md` (authentication.md, websocket-protocol.md, n8n-integration.md, project-structure.md, implementation-roadmap.md)

## Summary

Build a Go-based backend service for Armor Cyber Intelligence (ACI) that aggregates cybersecurity news, enriches content with Claude AI, provides real-time updates via WebSocket, and integrates with n8n for workflow automation. The system filters competitor content, scores Armor.com relevance, and delivers personalized alerts to users.

## Technical Context

**Language/Version**: Go 1.25+ (Latest: 1.25.5, December 2025)
**Primary Dependencies**:
- HTTP Router: Chi v5.0.12
- WebSocket: gorilla/websocket v1.5.1
- Database Driver: pgx v5.7+
- JWT: golang-jwt v5
- AI: anthropic-sdk-go v0.2.0-beta.3
- Logging: zerolog v1.33+
- Validation: go-playground/validator v10.28.0
- Migrations: golang-migrate v4.19.1

**Storage**: PostgreSQL 18+ with pgvector 0.8.1 (vector search, HNSW indexes)
**Testing**: Go testing + testify v1.9.0, testcontainers-go v0.34.0
**Target Platform**: Linux server (Docker/Kubernetes)
**Project Type**: Web application (REST API + WebSocket backend)
**Performance Goals**: 1000 req/s sustained, 500 concurrent WebSocket connections
**Constraints**: <200ms p95 API latency, <100MB memory per instance
**Scale/Scope**: ~10k users, ~100k articles, 8 categories, 10+ sources

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Gate 1: Automated Validation
- [x] Pre-commit hooks configured (secrets, sensitive files)
- [x] Configuration validation via `make validate`
- [x] Automated tests required (80%+ coverage per phase)

### Gate 2: License Compliance
- [x] All dependencies Apache 2.0 / MIT compatible
- [x] No GPL code in repository
- [x] anthropic-sdk-go license verified (Apache 2.0)

### Gate 3: Security Review
- [x] JWT RS256 signing (no HS256 algorithm confusion)
- [x] HMAC-SHA256 webhook signature verification
- [x] bcrypt password hashing (cost 12)
- [x] Rate limiting on auth endpoints
- [x] Brute force protection with progressive delays
- [x] No hardcoded secrets (all via environment variables)
- [x] OWASP Top 10 compliance required

### Gate 4: Code Review
- [x] Clean code standards enforced (no nested ifs, no hardcoded values)
- [x] Type safety with Go's static typing
- [x] Conventional commit format encouraged

### Gate 5: Architecture Review
- [x] Clean Architecture layers defined (domain, repository, service, api)
- [x] API-First design with OpenAPI contracts
- [x] Rollback capability via database migrations

### Gate 6: UX Review
- [x] REST API responses follow consistent format
- [x] Error messages are actionable (error code + message + request_id)
- [x] WebSocket protocol documented with all message types

### Gate 7: Verification Evidence
- [x] Health endpoints required (/health, /ready)
- [x] Integration tests with testcontainers
- [x] Load testing with k6

### Gate 8: Test Coverage
- [x] Four-case testing mandate acknowledged
- [x] 80%+ coverage required per phase
- [x] Critical paths require 90% coverage

## Project Structure

### Documentation (this feature)

```text
specs/001-aci-backend/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output (OpenAPI specs)
│   ├── auth.yaml
│   ├── articles.yaml
│   ├── alerts.yaml
│   ├── webhooks.yaml
│   └── websocket.yaml
└── tasks.md             # Phase 2 output (/speckit.tasks command)
```

### Source Code (repository root)

```text
aci-backend/
├── cmd/
│   └── server/
│       └── main.go                 # Application entry point
│
├── internal/
│   ├── config/
│   │   └── config.go               # Configuration loading
│   │
│   ├── domain/
│   │   ├── article.go              # Article entity
│   │   ├── category.go             # Category entity
│   │   ├── user.go                 # User entity
│   │   ├── alert.go                # Alert entity
│   │   ├── source.go               # Source entity
│   │   └── errors.go               # Domain errors
│   │
│   ├── repository/
│   │   ├── interfaces.go           # Repository interfaces
│   │   └── postgres/
│   │       ├── article_repo.go
│   │       ├── category_repo.go
│   │       ├── user_repo.go
│   │       ├── alert_repo.go
│   │       ├── source_repo.go
│   │       └── db.go
│   │
│   ├── service/
│   │   ├── interfaces.go
│   │   ├── article_service.go
│   │   ├── auth_service.go
│   │   ├── alert_service.go
│   │   ├── search_service.go
│   │   └── notification_service.go
│   │
│   ├── api/
│   │   ├── server.go
│   │   ├── router.go
│   │   ├── handlers/
│   │   │   ├── article_handler.go
│   │   │   ├── auth_handler.go
│   │   │   ├── category_handler.go
│   │   │   ├── alert_handler.go
│   │   │   ├── user_handler.go
│   │   │   ├── stats_handler.go
│   │   │   ├── webhook_handler.go
│   │   │   └── health_handler.go
│   │   ├── middleware/
│   │   │   ├── auth.go
│   │   │   ├── cors.go
│   │   │   ├── ratelimit.go
│   │   │   ├── logging.go
│   │   │   ├── recovery.go
│   │   │   └── requestid.go
│   │   └── response/
│   │       ├── response.go
│   │       └── errors.go
│   │
│   ├── websocket/
│   │   ├── hub.go
│   │   ├── client.go
│   │   ├── message.go
│   │   └── handler.go
│   │
│   ├── ai/
│   │   ├── client.go
│   │   ├── enrichment.go
│   │   ├── embeddings.go
│   │   └── prompts.go
│   │
│   └── pkg/
│       ├── jwt/
│       │   └── jwt.go
│       ├── validator/
│       │   └── validator.go
│       ├── slug/
│       │   └── slug.go
│       └── sanitizer/
│           └── sanitizer.go
│
├── migrations/
│   ├── 001_initial_schema.sql
│   ├── 002_add_articles.sql
│   └── ...
│
├── scripts/
│   ├── migrate.sh
│   ├── seed.sh
│   └── generate.sh
│
├── deployments/
│   ├── Dockerfile
│   ├── docker-compose.yml
│   └── k8s/
│       ├── deployment.yaml
│       ├── service.yaml
│       └── configmap.yaml
│
├── docs/
│   └── api/
│       └── openapi.yaml
│
├── tests/
│   ├── integration/
│   │   ├── article_test.go
│   │   ├── auth_test.go
│   │   └── setup_test.go
│   └── mocks/
│       └── ...
│
├── go.mod
├── go.sum
├── Makefile
├── .env.example
└── README.md
```

**Structure Decision**: Web application (backend-only) following Go Clean Architecture. Frontend is separate concern (React, not in scope for this plan).

## Complexity Tracking

> **No constitution violations identified. Standard Clean Architecture pattern justified.**

| Pattern | Justification |
|---------|---------------|
| Repository Pattern | Required for testability and database abstraction per Clean Architecture |
| Service Layer | Business logic isolation, enables unit testing without DB |
| WebSocket Hub | Standard pattern for managing concurrent connections |

---

## Phase 0: Research Summary

All technical decisions are resolved from the existing specifications:

| Question | Decision | Source |
|----------|----------|--------|
| Database | PostgreSQL 18+ with pgvector | project-structure.md |
| Auth method | JWT RS256 (15min access, 7d refresh) | authentication.md |
| Webhook auth | HMAC-SHA256 signatures | n8n-integration.md |
| WebSocket protocol | JSON envelope with type/id/timestamp/payload | websocket-protocol.md |
| AI integration | Claude claude-4.5-sonnet via anthropic-sdk-go | implementation-roadmap.md |
| HTTP router | Chi v5.0.12 | project-structure.md |

## Phase 1-4: Implementation Phases

### Phase 1: Foundation (Weeks 1-2)

**Goal**: Project setup, database schema, authentication system

#### Wave 1.1: Project Bootstrap [P]
- [P] Initialize Go module and directory structure
- [P] Configure development environment (Docker, Make)
- [P] Set up PostgreSQL with pgvector extension
- [P] Implement configuration loading (envconfig)
- [P] Set up zerolog with structured logging

#### Wave 1.2: Database Foundation
- Create migration framework (golang-migrate)
- Implement users table migration
- Implement refresh_tokens table
- Implement user_preferences table
- Create database connection pool (pgx)
- Implement repository base patterns

#### Wave 1.3: Authentication [P]
- [P] Implement JWT service (RS256 signing)
- [P] Create password hashing service (bcrypt)
- Create access token generation/validation
- Create refresh token rotation logic
- Implement user repository

#### Wave 1.4: Auth Endpoints
- POST /v1/auth/register endpoint
- POST /v1/auth/login endpoint
- POST /v1/auth/refresh endpoint
- POST /v1/auth/logout endpoint
- Auth middleware implementation
- GET /health and /ready endpoints

**Deliverables**: Running app with auth flow, 80%+ test coverage on auth module

---

### Phase 2: Core Content (Weeks 3-5)

**Goal**: Article management, n8n integration, content filtering

#### Wave 2.1: Content Schema [P]
- [P] Implement sources table migration
- [P] Implement categories table migration
- [P] Create source repository
- [P] Create category repository
- Seed initial categories (8 cyber categories)
- Seed initial sources (10 trusted sources)

#### Wave 2.2: Articles System
- Implement articles table migration
- Create article repository (CRUD operations)
- Implement full-text search with ts_vector
- Implement semantic search with pgvector

#### Wave 2.3: Article Endpoints [P]
- [P] GET /v1/articles (list with pagination)
- [P] GET /v1/articles/:id (single article)
- [P] GET /v1/articles/search endpoint
- GET /v1/articles/feed (personalized feed)
- Implement response DTOs and transformers

#### Wave 2.4: n8n Integration
- Implement HMAC-SHA256 signature verification
- POST /v1/webhooks/n8n endpoint
- Handle article.created event
- Handle article.updated event
- Handle bulk.import event
- Implement webhook_logs table and logging

#### Wave 2.5: Content Filtering [P]
- [P] Implement competitor detection service
- [P] Create competitor keyword dictionary
- [P] Implement Armor.com relevance scoring
- Build CTA injection service
- Implement content moderation flags

**Deliverables**: Full article system, n8n webhooks, content filtering, 80%+ coverage

---

### Phase 3: Real-time & Alerts (Weeks 6-8)

**Goal**: WebSocket infrastructure, alert system, user engagement

#### Wave 3.1: WebSocket Hub [P]
- [P] Implement WebSocket hub (connection manager)
- [P] Create client connection handler
- Implement JWT authentication for WS
- Build message router
- Implement ping/pong heartbeat

#### Wave 3.2: Subscriptions
- Implement channel subscription system
- articles:all channel
- articles:critical channel
- articles:category:{slug} channels
- Broadcast service for new articles

#### Wave 3.3: Alert Management [P]
- [P] Implement alerts table migration
- [P] Create alert repository
- [P] POST /v1/alerts endpoint
- [P] GET /v1/alerts endpoint
- PATCH /v1/alerts/:id endpoint
- DELETE /v1/alerts/:id endpoint

#### Wave 3.4: Alert Matching
- Implement alert matching service
- Keyword matching logic
- Category/severity matching
- Implement alert_matches table
- Real-time alert notifications via WS
- alerts:user channel implementation

#### Wave 3.5: User Engagement [P]
- [P] Implement bookmarks table migration
- [P] POST /v1/articles/:id/bookmark endpoint
- [P] DELETE /v1/articles/:id/bookmark endpoint
- GET /v1/users/me/bookmarks endpoint
- Implement article_reads table
- POST /v1/articles/:id/read endpoint

#### Wave 3.6: Statistics
- Implement daily_stats table migration
- Create stats aggregation service
- GET /v1/stats/dashboard endpoint
- GET /v1/stats/categories endpoint
- GET /v1/stats/trends endpoint

**Deliverables**: WebSocket infrastructure, alerts, engagement features, 80%+ coverage

---

### Phase 4: AI & Production (Weeks 9-10)

**Goal**: Claude AI integration, admin endpoints, production hardening

#### Wave 4.1: Claude Integration [P]
- [P] Implement Anthropic client wrapper
- [P] Create AI enrichment service
- Implement threat analysis generation
- Implement Armor CTA generation
- Handle enrichment.complete webhook event

#### Wave 4.2: AI Processing Pipeline
- Build article enrichment queue
- Implement rate limiting for Claude API
- Create embedding generation for semantic search
- Implement AI analysis caching

#### Wave 4.3: Admin & Security [P]
- [P] Admin article management endpoints
- [P] Admin user management endpoints
- [P] Admin source management endpoints
- Implement audit_logs table and logging
- Rate limiting middleware
- Request ID and correlation tracking

#### Wave 4.4: Production Hardening
- Implement graceful shutdown
- Database connection pool tuning
- Add Prometheus metrics endpoint
- Create Kubernetes manifests
- Load testing and optimization
- Security audit and fixes

**Deliverables**: AI enrichment, admin endpoints, production-ready deployment, 80%+ overall coverage

---

## Dependencies & Execution Order

```
Phase 1 (Foundation)
├── Wave 1.1: Project Bootstrap [P] ────────────────┐
├── Wave 1.2: Database Foundation ──────────────────┤
├── Wave 1.3: Authentication [P] (depends: 1.2) ────┤
└── Wave 1.4: Auth Endpoints (depends: 1.3) ────────┘

Phase 2 (Core Content) - BLOCKS on Phase 1
├── Wave 2.1: Content Schema [P] ───────────────────┐
├── Wave 2.2: Articles System (depends: 2.1) ───────┤
├── Wave 2.3: Article Endpoints [P] (depends: 2.2) ─┤
├── Wave 2.4: n8n Integration (depends: 2.2) ───────┤
└── Wave 2.5: Content Filtering [P] (depends: 2.2) ─┘

Phase 3 (Real-time) - BLOCKS on Phase 2
├── Wave 3.1: WebSocket Hub [P] ────────────────────┐
├── Wave 3.2: Subscriptions (depends: 3.1) ─────────┤
├── Wave 3.3: Alert Management [P] ─────────────────┤
├── Wave 3.4: Alert Matching (depends: 3.3) ────────┤
├── Wave 3.5: User Engagement [P] ──────────────────┤
└── Wave 3.6: Statistics (depends: 3.5) ────────────┘

Phase 4 (AI & Production) - BLOCKS on Phase 3
├── Wave 4.1: Claude Integration [P] ───────────────┐
├── Wave 4.2: AI Pipeline (depends: 4.1) ───────────┤
├── Wave 4.3: Admin & Security [P] ─────────────────┤
└── Wave 4.4: Production Hardening (depends: all) ──┘
```

## Team Distribution

| Phase | Developer A | Developer B | Developer C |
|-------|-------------|-------------|-------------|
| 1 | Bootstrap, Auth | Database | Testing |
| 2 | Articles API | n8n Webhooks | Content Filtering |
| 3 | WebSocket Hub | Alerts | Engagement/Stats |
| 4 | AI Integration | Admin Endpoints | Production |

---

## Risk Mitigation

| Risk | Impact | Mitigation |
|------|--------|------------|
| pgvector performance at scale | High | Index tuning, fallback to full-text |
| WebSocket connection limits | High | Horizontal scaling, connection pooling |
| Claude API rate limits | Medium | Queue system, caching, rate limiting |
| n8n webhook reliability | Medium | Retry logic, dead letter queue |

---

## Product Manager Review Gates

The product-manager-agent participates at key milestones:

### Gate PM-1: Pre-Implementation Review (Before Phase 1)
- [ ] Plan reviewed for completeness
- [ ] User stories validated
- [ ] Acceptance criteria clear
- [ ] Business value confirmed
- [ ] Gaps from PM review addressed (see `pm-review.md`)
- **Blocker**: Implementation MUST NOT start until PM approves

### Gate PM-2: Mid-Implementation Review (After Phase 2)
- [ ] Core content features demonstrated
- [ ] n8n integration verified working
- [ ] User journey walkthrough completed
- [ ] Feedback incorporated into Phase 3-4 plan
- **Checkpoint**: No-go if critical issues found

### Gate PM-3: Pre-Release Review (After Phase 4)
- [ ] All features verified against acceptance criteria
- [ ] Performance targets validated (1000 req/s, <200ms p95)
- [ ] Security audit findings addressed
- [ ] Documentation complete and accurate
- [ ] Demo to stakeholders completed
- [ ] 60+ point verification checklist passed (see `pm-review.md`)
- **Go/No-Go**: Final release decision

### PM Verification Protocol
For each deliverable, PM verifies:
1. **Functional**: Does it work as specified?
2. **Usable**: Is the API intuitive and well-documented?
3. **Documented**: Is usage clear with examples?
4. **Tested**: Are all four test cases covered (happy/fail/null/edge)?
5. **Observable**: Are metrics, logs, and health checks in place?

### PM Contact Points
- Phase gate reviews (mandatory)
- User story clarification (on-demand)
- Acceptance criteria validation (per feature)
- Final delivery sign-off (required)

**Reference**: See `specs/001-aci-backend/pm-review.md` for detailed PM analysis, gap analysis, RICE-scored suggestions, and 10 actionable improvement tasks.

---

## Definition of Done (Per Phase)

- [ ] All features implemented and compiling
- [ ] Unit tests written and passing (80%+ coverage)
- [ ] Integration tests where applicable
- [ ] Four-case testing mandate satisfied
- [ ] No critical security issues (gosec)
- [ ] Documentation updated
- [ ] Health endpoints verified
- [ ] Load testing passed (1000 req/s target)
- [ ] PM gate review passed (PM-1, PM-2, or PM-3 as applicable)
