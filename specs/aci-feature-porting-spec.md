# ACI Feature Porting Specification
## From n8n-bloom to n8n-cyber-news

**Version:** 1.0
**Date:** 2024-12-13
**Status:** Ready for Implementation

---

## Executive Summary

This specification documents the feature porting strategy from n8n-bloom (health content system) to n8n-cyber-news (Armor Cyber Intelligence platform). The analysis identified 25+ portable features with the backend at ~85% completion and frontend at ~60% completion.

---

## Part 1: UX Specification

### 1.1 Brand Identity

**Name:** NEXUS by Armor
**Tagline:** Cyber Intelligence, Unified

**Color Palette:**
```css
:root {
  /* Primary - Dark Theme Base */
  --color-bg-primary: #0A1628;      /* Deep navy */
  --color-bg-secondary: #1E293B;    /* Slate card */
  --color-bg-tertiary: #334155;     /* Hover states */

  /* Accent Colors */
  --color-accent-primary: #00D4FF;  /* Cyber blue */
  --color-accent-secondary: #6366F1; /* Indigo */

  /* Severity Colors */
  --color-critical: #FF4444;        /* Red */
  --color-high: #FF8C00;            /* Orange */
  --color-medium: #FACC15;          /* Yellow */
  --color-low: #22C55E;             /* Green */

  /* Text */
  --color-text-primary: #F8FAFC;    /* White */
  --color-text-secondary: #94A3B8;  /* Gray */
  --color-text-muted: #64748B;      /* Muted */

  /* Borders */
  --color-border: #334155;
  --color-border-focus: #00D4FF;
}
```

**Typography:**
- Headings: Inter (600/700 weight)
- Body: Inter (400/500 weight)
- Code/Technical: JetBrains Mono

### 1.2 User Personas

| Persona | Role | Primary Goals | Key Features |
|---------|------|---------------|--------------|
| **Alex** | Security Analyst | Monitor threats, investigate CVEs | Real-time feed, CVE search, bookmarks |
| **Jordan** | SOC Manager | Team oversight, alert config | Alert rules, team dashboards, reports |
| **Morgan** | CISO | Executive summary, compliance | Stats dashboard, export, trends |

### 1.3 Information Architecture

```
NEXUS Dashboard
├── Dashboard (Home)
│   ├── Threat Overview Cards
│   ├── Recent Activity Feed
│   ├── Severity Distribution
│   └── Quick Actions
├── Threats
│   ├── All Threats (filtered list)
│   ├── Threat Detail View
│   └── Search & Advanced Filters
├── Alerts
│   ├── My Alerts
│   ├── Create Alert Rule
│   └── Alert History
├── Bookmarks
│   └── Saved Threats
├── Analytics
│   ├── Threat Trends
│   ├── Category Breakdown
│   └── Source Analysis
├── Admin (role-gated)
│   ├── Content Review Queue
│   ├── User Management
│   └── System Health
└── Settings
    ├── Profile
    ├── Notifications
    └── API Keys
```

### 1.4 Component Specifications

#### Header Component
```
┌─────────────────────────────────────────────────────────────────┐
│ [NEXUS Logo]  Dashboard  Threats  Alerts  Analytics     [User] │
│                                            🔔 [Search]   [Menu] │
└─────────────────────────────────────────────────────────────────┘

Props:
- currentRoute: string
- user: { name, avatar, role }
- notificationCount: number
- onSearch: (query: string) => void
```

#### Threat Card Component
```
┌─────────────────────────────────────────────────────────────────┐
│ [CRITICAL]  New Zero-Day Vulnerability in Apache Log4j         │
│ ─────────────────────────────────────────────────────────────── │
│ CVE-2024-XXXX | Ransomware | CISA | 2 hours ago               │
│                                                                 │
│ Brief summary of the threat (2-3 lines max)...                 │
│                                                                 │
│ [Vulnerability] [Remote Code Execution]      [🔖] [→ Details]  │
└─────────────────────────────────────────────────────────────────┘

Props:
- id: string
- title: string
- severity: 'critical' | 'high' | 'medium' | 'low'
- category: string
- source: string
- publishedAt: Date
- summary: string
- cves: string[]
- tags: string[]
- isBookmarked: boolean
- onBookmark: () => void
- onClick: () => void
```

#### Severity Badge Component
```typescript
interface SeverityBadgeProps {
  severity: 'critical' | 'high' | 'medium' | 'low';
  size?: 'sm' | 'md' | 'lg';
  showIcon?: boolean;
}

// Visual:
// [🔴 CRITICAL] - Red bg, white text
// [🟠 HIGH]     - Orange bg, white text
// [🟡 MEDIUM]   - Yellow bg, dark text
// [🟢 LOW]      - Green bg, white text
```

#### Filter Panel Component
```
┌─────────────────────────────────────────────────────────────────┐
│ Filters                                          [Clear All]    │
├─────────────────────────────────────────────────────────────────┤
│ Severity:  [×Critical] [×High] [Medium] [Low]                  │
│ Category:  [Dropdown: All Categories ▼]                        │
│ Source:    [Dropdown: All Sources ▼]                           │
│ Date:      [Last 24h ▼]  or  [Custom Range]                    │
│ Search:    [🔍 Search threats, CVEs...]                        │
└─────────────────────────────────────────────────────────────────┘
```

### 1.5 User Flows

#### Flow 1: Threat Investigation
```
Login → Dashboard → Click Threat Card → View Detail
→ [Branch A: Bookmark for later]
→ [Branch B: Create Alert Rule based on this]
→ [Branch C: Share/Export]
```

#### Flow 2: Alert Configuration
```
Alerts → Create New Alert → Select Criteria:
  - Keywords (e.g., "Log4j", "CVE-2024")
  - Severity threshold (Critical/High)
  - Categories (Ransomware, Vulnerabilities)
→ Set Notification Preferences → Save → Confirm
```

#### Flow 3: Admin Content Review
```
Admin → Review Queue → Select Pending Item
→ View AI-generated analysis
→ [Approve] / [Reject] / [Edit & Approve]
→ Item moves to Published/Rejected
→ Next item auto-loads
```

### 1.6 Responsive Breakpoints

| Breakpoint | Width | Layout Changes |
|------------|-------|----------------|
| Mobile | < 640px | Single column, bottom nav, collapsed filters |
| Tablet | 640-1024px | Two column, sidebar collapsed by default |
| Desktop | > 1024px | Full sidebar, three column dashboard |

### 1.7 Accessibility Requirements

- **WCAG 2.1 AA Compliance**
- Color contrast ratio: minimum 4.5:1 for text
- Focus indicators: 2px cyan outline
- Keyboard navigation: Tab order follows visual flow
- Screen reader: ARIA labels on all interactive elements
- Motion: Respect `prefers-reduced-motion`

---

## Part 2: Data Visualization Specification

### 2.1 Dashboard Layout

```
┌─────────────────────────────────────────────────────────────────┐
│                        HEADER                                   │
├───────────────┬─────────────────────────────────────────────────┤
│               │  ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐           │
│   SIDEBAR     │  │ Card │ │ Card │ │ Card │ │ Card │           │
│   (Nav)       │  │Total │ │Crit. │ │Today │ │Alerts│           │
│               │  └──────┘ └──────┘ └──────┘ └──────┘           │
│               ├─────────────────────┬───────────────────────────┤
│               │   THREAT TIMELINE   │   SEVERITY DONUT          │
│               │   (Line Chart)      │   (Donut Chart)           │
│               ├─────────────────────┴───────────────────────────┤
│               │           REAL-TIME ACTIVITY FEED               │
│               │   ┌─────────────────────────────────────────┐   │
│               │   │ [CRIT] New threat detected...  2s ago   │   │
│               │   │ [HIGH] CVE-2024-1234 updated   5m ago   │   │
│               │   │ [MED]  Phishing campaign...    12m ago  │   │
│               │   └─────────────────────────────────────────┘   │
└───────────────┴─────────────────────────────────────────────────┘
```

### 2.2 Chart Specifications

#### 2.2.1 Summary Metric Cards
```typescript
interface MetricCardProps {
  title: string;
  value: number;
  change?: { value: number; direction: 'up' | 'down' };
  icon: ReactNode;
  color: 'default' | 'critical' | 'warning' | 'success';
  sparklineData?: number[];
}

// Example cards:
// - Total Threats: 1,234 (+12 today)
// - Critical: 23 (↑ 5 from yesterday)
// - New Today: 47
// - Active Alerts: 8 matched
```

#### 2.2.2 Severity Distribution (Donut Chart)
```typescript
interface SeverityDonutProps {
  data: {
    critical: number;
    high: number;
    medium: number;
    low: number;
  };
  showLegend?: boolean;
  showPercentages?: boolean;
  size?: 'sm' | 'md' | 'lg';
}

// Colors:
// - Critical: #FF4444
// - High: #FF8C00
// - Medium: #FACC15
// - Low: #22C55E

// Interactions:
// - Hover: Show tooltip with count and percentage
// - Click segment: Filter threat list by severity
```

#### 2.2.3 Threat Timeline (Area Chart)
```typescript
interface ThreatTimelineProps {
  data: Array<{
    date: Date;
    total: number;
    critical: number;
    high: number;
    medium: number;
    low: number;
  }>;
  timeRange: '24h' | '7d' | '30d' | '90d';
  onTimeRangeChange: (range: string) => void;
  showBreakdown?: boolean;
}

// Visual:
// - Stacked area chart with severity breakdown
// - X-axis: Time (formatted based on range)
// - Y-axis: Threat count
// - Gradient fill with 30% opacity
// - Hover: Crosshair with tooltip showing all values
```

#### 2.2.4 Category Distribution (Horizontal Bar)
```typescript
interface CategoryBarProps {
  data: Array<{
    category: string;
    count: number;
    change: number;
  }>;
  maxCategories?: number;
  showPercentage?: boolean;
}

// Categories:
// - Vulnerabilities
// - Ransomware
// - Data Breaches
// - Malware
// - Phishing
// - Threat Actors

// Visual:
// - Horizontal bars with gradient fill
// - Category label on left
// - Count on right
// - Subtle animation on load
```

#### 2.2.5 Real-Time Activity Feed
```typescript
interface ActivityFeedProps {
  items: Array<{
    id: string;
    type: 'threat' | 'alert' | 'system';
    severity?: SeverityLevel;
    title: string;
    timestamp: Date;
  }>;
  maxItems?: number;
  onItemClick: (id: string) => void;
  autoRefresh?: boolean;
  refreshInterval?: number;
}

// Visual:
// - Vertically stacked cards
// - New items slide in from top with fade
// - Severity indicator on left edge
// - Relative timestamp (e.g., "2m ago")
// - Pulse animation for items < 30s old
```

#### 2.2.6 Source Distribution (Pie/Treemap)
```typescript
interface SourceDistributionProps {
  data: Array<{
    source: string;
    count: number;
    logo?: string;
  }>;
  variant: 'pie' | 'treemap';
}

// Sources:
// - CISA
// - BleepingComputer
// - HackerNews
// - NVD
// - Custom feeds
```

### 2.3 Library Recommendations

| Library | Use Case | Rationale |
|---------|----------|-----------|
| **Recharts** | Primary charts | React-native, composable, good DX |
| **Framer Motion** | Animations | Smooth transitions, gesture support |
| **date-fns** | Date formatting | Lightweight, tree-shakeable |
| **TanStack Query** | Data fetching | Caching, real-time updates |

### 2.4 Animation Guidelines

```css
/* Standard transitions */
--transition-fast: 150ms ease-out;
--transition-normal: 250ms ease-out;
--transition-slow: 400ms ease-out;

/* Chart animations */
- Bar/line entry: 400ms with stagger
- Donut segments: 600ms with spring
- New feed items: 300ms slide + fade
- Number counters: 500ms with easing
```

### 2.5 Sample Data Structures

```typescript
// Dashboard summary endpoint
interface DashboardSummary {
  totalThreats: number;
  criticalCount: number;
  newToday: number;
  activeAlerts: number;
  severityDistribution: {
    critical: number;
    high: number;
    medium: number;
    low: number;
  };
  categoryDistribution: Array<{
    category: string;
    count: number;
  }>;
  recentActivity: Array<ActivityItem>;
  timelineData: Array<TimelinePoint>;
}
```

---

## Part 3: Product Management Review

### 3.1 Feature Prioritization (RICE Framework)

| Feature | Reach | Impact | Confidence | Effort | RICE Score | Priority |
|---------|-------|--------|------------|--------|------------|----------|
| Fix Router Wiring | 100% | 3 | 100% | 0.5d | **600** | P0 |
| Fix Integration Tests | 100% | 3 | 90% | 1d | **270** | P0 |
| Frontend-Backend Integration | 100% | 3 | 85% | 2d | **127** | P0 |
| Dashboard Implementation | 100% | 3 | 80% | 3d | **80** | P1 |
| Vector Search (pgvector) | 70% | 2 | 70% | 2d | **49** | P1 |
| Content Export Service | 40% | 2 | 80% | 2d | **32** | P2 |
| Claude AI Enrichment | 60% | 2 | 75% | 3d | **30** | P1 |
| CDN Sync API | 30% | 2 | 70% | 2d | **21** | P2 |
| OpenTelemetry | 50% | 1 | 90% | 1d | **45** | P2 |
| Review UI | 20% | 2 | 80% | 3d | **11** | P3 |
| Translation Service | 10% | 1 | 60% | 3d | **2** | P4 |

### 3.2 MVP Scope Definition

#### MVP (Phase 1) - Must Have
- [ ] Working API endpoints (router wiring)
- [ ] User authentication flow
- [ ] Threat listing with filters
- [ ] Threat detail view
- [ ] Real-time WebSocket updates
- [ ] Basic alert subscriptions
- [ ] Bookmark functionality

#### Phase 2 - Should Have
- [ ] AI-powered threat enrichment
- [ ] Vector semantic search
- [ ] Dashboard visualizations
- [ ] Advanced filtering

#### Phase 3 - Nice to Have
- [ ] Content export/bundling
- [ ] CDN sync for mobile
- [ ] OpenTelemetry observability
- [ ] Admin review UI

#### Future Roadmap
- [ ] Multi-language support
- [ ] Email digest notifications
- [ ] Slack/Teams integration
- [ ] Custom RSS feed generation

### 3.3 User Stories

#### P0 - Critical (Blocking Launch)

**US-001: API Accessibility**
> As a frontend developer, I want all API endpoints properly routed so that the frontend can communicate with the backend.

**Acceptance Criteria:**
- All handlers wired in router.go
- Health check endpoint returns 200
- CORS configured correctly
- OpenAPI spec matches implementation

---

**US-002: User Authentication**
> As a user, I want to log in securely so that I can access personalized features.

**Acceptance Criteria:**
- Registration with email/password
- Login returns JWT token
- Token refresh works
- Logout invalidates session

---

**US-003: Threat Browsing**
> As a security analyst, I want to browse threats with filters so that I can find relevant information quickly.

**Acceptance Criteria:**
- List view with pagination (20 per page)
- Filter by severity, category, source, date
- Search by title/content/CVE
- Sort by date (newest first default)

---

#### P1 - High Priority

**US-004: Real-time Updates**
> As a user, I want to receive real-time threat notifications so that I stay informed of critical issues.

**Acceptance Criteria:**
- WebSocket connection established on login
- New threats appear without refresh
- Notification badge updates
- Connection recovery on disconnect

---

**US-005: Alert Configuration**
> As a user, I want to create alert rules so that I'm notified of threats matching my criteria.

**Acceptance Criteria:**
- Create alert with keywords, severity, categories
- Edit/delete existing alerts
- View alert match history
- Real-time notification when alert triggers

---

**US-006: Dashboard Visualizations**
> As a CISO, I want a visual dashboard so that I can quickly assess the threat landscape.

**Acceptance Criteria:**
- Summary metric cards
- Severity distribution chart
- Threat timeline (7-day default)
- Recent activity feed

---

### 3.4 Success Metrics (KPIs)

| Metric | Target | Measurement Method |
|--------|--------|-------------------|
| API Response Time | < 200ms p95 | Application metrics |
| WebSocket Latency | < 500ms | Message timestamp delta |
| Test Coverage | > 80% | CI coverage report |
| Error Rate | < 0.1% | Error tracking (Sentry) |
| User Engagement | > 3 sessions/week | Analytics |
| Alert Accuracy | > 90% relevance | User feedback |

### 3.5 Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Integration test failures block CI | High | High | Fix type mismatches immediately |
| API rate limits from news sources | Medium | Medium | Implement caching, respect rate limits |
| AI enrichment costs exceed budget | Medium | High | Token budget manager, tiered processing |
| WebSocket scalability issues | Low | High | Load test before launch, horizontal scaling |
| Security vulnerabilities | Low | Critical | Security review gate, OWASP compliance |

### 3.6 Release Plan

```
PHASE 1: Fix Blockers (3-4 days)
├── Day 1-2: Router wiring + integration test fixes
├── Day 2-3: Frontend-backend integration testing
└── Day 3-4: n8n workflow signature fix

PHASE 2: Complete Core (5-7 days)
├── Day 1-3: Dashboard page implementation
├── Day 3-5: Vector search integration
├── Day 5-6: AI enrichment service
└── Day 6-7: End-to-end testing

PHASE 3: Production Readiness (3-4 days)
├── Day 1-2: Kubernetes deployment
├── Day 2-3: Monitoring/observability
├── Day 3: Load testing
└── Day 4: Documentation & runbooks

PHASE 4: Launch + Iteration
├── Week 1: Soft launch to beta users
├── Week 2-3: Feedback collection + fixes
└── Week 4: General availability
```

---

## Part 4: Implementation Tasks

### 4.1 Immediate Blockers (P0)

```markdown
### Task 1: Wire Router Endpoints
- File: aci-backend/internal/api/router.go
- Action: Connect all handlers to routes
- Estimate: 2-4 hours
- Owner: go-dev agent

### Task 2: Fix Integration Tests
- File: aci-backend/tests/integration/setup_test.go
- Action: Align database driver types (pgxpool vs database/sql)
- Estimate: 4-6 hours
- Owner: go-dev agent

### Task 3: Frontend API Integration
- Files: aci-frontend/src/services/*.ts
- Action: Test all API calls, fix CORS, add error handling
- Estimate: 8-12 hours
- Owner: ts-dev agent

### Task 4: n8n Workflow Signature
- File: n8n-cyber-news-workflow.json
- Action: Implement HMAC-SHA256 signature in Code node
- Estimate: 1-2 hours
- Owner: n8n-workflow agent
```

### 4.2 Core Features (P1)

```markdown
### Task 5: Dashboard Page
- Files: aci-frontend/src/pages/Dashboard.tsx, components/charts/*
- Action: Implement full dashboard with visualizations
- Estimate: 12-16 hours
- Owner: ts-dev + frontend-developer agents

### Task 6: Vector Search
- Files: aci-backend/internal/service/search_service.go
- Action: Integrate pgvector for semantic search
- Estimate: 8-10 hours
- Owner: go-dev + database-dev agents

### Task 7: AI Enrichment Pipeline
- Files: aci-backend/internal/ai/*
- Action: Port enrichment patterns from n8n-bloom
- Estimate: 10-12 hours
- Owner: go-dev + ai-engineer agents
```

### 4.3 Production Readiness (P2)

```markdown
### Task 8: Kubernetes Manifests
- Files: aci-backend/deployments/k8s/*
- Action: Complete K8s config with probes, limits, secrets
- Estimate: 4-8 hours
- Owner: k8s-deployment-engineer agent

### Task 9: Observability Setup
- Files: New files in internal/pkg/telemetry/
- Action: Port OpenTelemetry setup from n8n-bloom
- Estimate: 6-10 hours
- Owner: devops-eng agent

### Task 10: Load Testing
- Files: New k6 test scripts
- Action: Create load tests for 1000 req/s target
- Estimate: 4-6 hours
- Owner: test-automator agent
```

---

## Part 5: Domain Mapping Reference

| n8n-bloom (Health) | n8n-cyber-news (Security) |
|-------------------|---------------------------|
| articles | threat_reports |
| symptoms | threat_indicators / IOCs |
| life_stages | attack_vectors |
| age_ranges | severity_levels |
| enrichment | threat_assessment |
| sources (NIH, CDC) | sources (CISA, NVD) |
| categories | threat_categories |
| screening_tools | mitigation_guides |

---

## Appendix A: File Reference

### Key Files to Modify

**Backend (Go):**
- `aci-backend/internal/api/router.go` - Route wiring
- `aci-backend/internal/service/search_service.go` - Vector search
- `aci-backend/tests/integration/setup_test.go` - Test fixes

**Frontend (React):**
- `aci-frontend/src/pages/Dashboard.tsx` - Dashboard
- `aci-frontend/src/components/charts/` - New directory
- `aci-frontend/src/services/api.ts` - API client

**Infrastructure:**
- `n8n-cyber-news-workflow.json` - Workflow fixes
- `aci-backend/deployments/k8s/` - K8s manifests

---

## Approval

- [ ] UX Review Complete
- [ ] Technical Review Complete
- [ ] PM Sign-off
- [ ] Ready for Implementation

---

*Document generated by orchestration analysis on 2024-12-13*
