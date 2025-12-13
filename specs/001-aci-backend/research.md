# Research: ACI Backend

**Date**: 2025-12-11
**Feature**: 001-aci-backend

## Overview

This document consolidates research findings from the existing specification documents. All technical decisions have been pre-determined in the specification files.

---

## 1. Technology Stack Decisions

### 1.1 Programming Language

**Decision**: Go 1.25+

**Rationale**:
- Strong concurrency model (goroutines, channels) ideal for WebSocket hub
- Excellent performance for high-throughput API services
- Static typing catches errors at compile time
- Rich ecosystem for web services
- Easy deployment (single binary)

**Alternatives Considered**:
- Node.js: Higher throughput for I/O but less type safety
- Rust: Better performance but steeper learning curve
- Python: Slower runtime performance

### 1.2 HTTP Framework

**Decision**: Chi v5.0.12

**Rationale**:
- Lightweight, idiomatic Go router
- Chi v5 leverages Go 1.22+ standard library mux
- Excellent middleware support
- Active maintenance

**Alternatives Considered**:
- Gin: More features but heavier
- Standard library net/http: Too low-level
- Fiber: Non-standard Go patterns

### 1.3 Database

**Decision**: PostgreSQL 18+ with pgvector 0.8.1

**Rationale**:
- pgvector enables semantic search with HNSW indexes
- PostgreSQL 18 includes performance improvements
- ts_vector for full-text search built-in
- Mature, reliable, ACID compliant
- JSON/JSONB support for flexible schemas

**Alternatives Considered**:
- MongoDB: Good for documents but weaker for relational queries
- SQLite: Not suitable for concurrent writes
- Pinecone: Additional service dependency for vectors

### 1.4 AI Integration

**Decision**: Claude claude-4.5-sonnet via anthropic-sdk-go v0.2.0-beta.3

**Rationale**:
- Official Anthropic Go SDK
- High-quality content analysis
- Streaming support for long responses
- Good rate limits for production

**Alternatives Considered**:
- OpenAI GPT-4: Similar capability, different pricing
- Local LLM: Insufficient quality for threat analysis
- Amazon Bedrock: Additional AWS dependency

---

## 2. Authentication Design

### 2.1 Token Strategy

**Decision**: JWT RS256 with refresh token rotation

**Specifications**:
- Access token: 15 minutes, RS256 signed
- Refresh token: 7 days, opaque (32-byte random), SHA-256 hashed in DB
- Token rotation: Each refresh generates new refresh token

**Security Measures**:
- Algorithm restriction: RS256 only (reject HS256)
- Key rotation: Every 90 days with 24-hour overlap
- Brute force protection: Progressive delays, account lockout after 10 attempts

### 2.2 Rate Limiting

| Endpoint | Limit | Window |
|----------|-------|--------|
| POST /auth/login | 5 requests | 1 minute per IP |
| POST /auth/refresh | 10 requests | 1 minute per user |
| POST /auth/register | 3 requests | 1 minute per IP |

---

## 3. WebSocket Protocol

### 3.1 Message Format

```json
{
  "type": "string",
  "id": "uuid",
  "timestamp": "ISO8601",
  "payload": {}
}
```

### 3.2 Channels

| Channel | Description |
|---------|-------------|
| articles:all | All new/updated articles |
| articles:critical | Critical severity only |
| articles:high | High severity and above |
| articles:category:{slug} | Category-specific |
| articles:vendor:{name} | Vendor-specific |
| alerts:user | User's alert matches |
| system | System announcements |

### 3.3 Connection Limits

| Limit | Value |
|-------|-------|
| Max subscriptions per connection | 50 |
| Max connections per user | 5 |
| Max message size | 64 KB |
| Messages per minute | 60 |
| Ping interval | 30 seconds |
| Reconnect grace period | 5 minutes |

---

## 4. n8n Integration

### 4.1 Webhook Authentication

**Decision**: HMAC-SHA256 signatures

**Header**: `X-N8N-Signature: sha256=<hex-signature>`

**Computation**:
```go
mac := hmac.New(sha256.New, []byte(secret))
mac.Write(body)
signature := "sha256=" + hex.EncodeToString(mac.Sum(nil))
```

### 4.2 Event Types

| Event | Description |
|-------|-------------|
| article.created | New article from scraper |
| article.updated | Update existing article |
| article.deleted | Soft delete article |
| bulk.import | Import multiple articles (max 100) |
| enrichment.complete | AI enrichment results |

### 4.3 Processing Flow

1. Receive webhook
2. Verify HMAC signature
3. Validate payload schema
4. Check for duplicates (by source_url)
5. Generate slug
6. Insert article
7. Log webhook
8. Queue AI enrichment (if not skipped)
9. Check alert subscriptions
10. Broadcast to WebSocket hub

---

## 5. Content Filtering

### 5.1 Competitor Detection

**Competitors to filter**:
- CrowdStrike Falcon
- Palo Alto Prisma
- SentinelOne
- Microsoft Defender
- Carbon Black
- And others...

**Scoring**:
- Promotional mention: +0.3
- Neutral mention: +0.1
- Block if score > 0.5

### 5.2 Armor Relevance Scoring

**Keywords and weights**:
| Keyword | Weight |
|---------|--------|
| managed detection | 0.3 |
| mdr | 0.3 |
| cloud security | 0.2 |
| compliance | 0.2 |
| vulnerability management | 0.2 |
| incident response | 0.2 |
| threat intelligence | 0.2 |
| soc | 0.2 |

**Severity boost**:
- Critical: +0.2
- High: +0.1

---

## 6. Database Schema Summary

### Core Tables

1. **users** - User accounts with bcrypt passwords
2. **refresh_tokens** - Hashed refresh tokens
3. **user_preferences** - User settings
4. **categories** - 8 cybersecurity categories
5. **sources** - Content sources (CISA, NVD, etc.)
6. **articles** - Main content with embeddings
7. **alerts** - User alert subscriptions
8. **alert_matches** - Alert-article matches
9. **bookmarks** - User bookmarks
10. **article_reads** - Read tracking
11. **daily_stats** - Aggregated statistics
12. **webhook_logs** - Webhook audit trail
13. **audit_logs** - Admin action audit

---

## 7. Performance Targets

| Metric | Target |
|--------|--------|
| API latency p95 | <200ms |
| Throughput | 1000 req/s |
| Concurrent WebSocket | 500 connections |
| Memory per instance | <100MB |
| Database connections | 25 max |

---

## 8. Open Questions (Resolved)

All questions from initial analysis have been resolved by the specification documents:

| Question | Resolution | Source |
|----------|------------|--------|
| Which AI model? | Claude claude-4.5-sonnet | implementation-roadmap.md |
| Vector dimensions? | pgvector default (1536 for OpenAI-compatible) | project-structure.md |
| Webhook retry policy? | Exponential backoff, 3 retries, 5s initial | n8n-integration.md |
| Session storage? | Stateless JWT, no server sessions | authentication.md |

---

## 9. Dependencies Summary

```go
// go.mod core dependencies
github.com/go-chi/chi/v5 v5.0.12
github.com/gorilla/websocket v1.5.1
github.com/jackc/pgx/v5 v5.7.2
github.com/golang-migrate/migrate/v4 v4.19.1
github.com/rs/zerolog v1.33.0
github.com/go-playground/validator/v10 v10.28.0
github.com/golang-jwt/jwt/v5 v5.2.1
github.com/anthropics/anthropic-sdk-go v0.2.0-beta.3
github.com/google/uuid v1.6.0
github.com/stretchr/testify v1.9.0
github.com/testcontainers/testcontainers-go v0.34.0
golang.org/x/crypto v0.45.0
```

---

## 10. Conclusion

All technical decisions are resolved from the specification documents. The implementation can proceed directly to Phase 1 without additional research.

**Next Steps**:
1. Create data-model.md with entity definitions
2. Generate OpenAPI contracts in /contracts/
3. Create quickstart.md with setup instructions
4. Generate tasks.md via /speckit.tasks command
