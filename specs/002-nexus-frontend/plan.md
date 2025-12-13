# Implementation Plan: NEXUS Frontend Dashboard

**Branch**: `002-nexus-frontend` | **Date**: 2024-12-13 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/002-nexus-frontend/spec.md`

## Summary

NEXUS by Armor is the frontend dashboard for the Armor Cyber Intelligence (ACI) platform. This implementation delivers a React 19 + Vite 7 + TypeScript 5.9 cybersecurity threat intelligence dashboard with real-time WebSocket updates, data visualizations using Reviz, and shadcn/ui components. The dashboard enables security analysts to monitor threats, manage alerts, and view analytics with a dark theme cyber aesthetic.

## Technical Context

**Language/Version**: TypeScript 5.9 with React 19.2
**Primary Dependencies**: React 19.2, Vite 7.2, shadcn/ui (Radix UI + Tailwind), Reviz, TanStack Query v5, react-router-dom v7
**Storage**: N/A (frontend-only, backend manages persistence)
**Testing**: Vitest + React Testing Library (unit), Playwright (E2E)
**Target Platform**: Modern browsers (Chrome 90+, Firefox 90+, Safari 15+, Edge 90+)
**Project Type**: Web application (frontend-only, consumes aci-backend API)
**Performance Goals**: < 3s initial load, < 500ms interactions, 60fps animations
**Constraints**: < 200ms API response handling, < 100KB per route bundle
**Scale/Scope**: 500 concurrent users, 8 pages, 20+ components

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. License Compliance | PASS | Apache 2.0 compatible libraries only |
| II. Security First | PASS | HttpOnly cookies, no token in localStorage |
| III. Integration Integrity | PASS | API-only communication with backend |
| V. Validation Gates | PASS | Vitest + Playwright testing |
| VIII. Test-First Development | PASS | TDD with 4-case coverage |
| IX. Clean Code Standards | PASS | TypeScript strict mode, ESLint |
| X. Observable Systems | PASS | OpenTelemetry + SigNoz |
| XI. Parallel-First Orchestration | PASS | Phase/wave structure defined |
| XII. User Experience Excellence | PASS | Dark theme, accessibility, responsive |
| XIII. API-First Design | PASS | Contracts defined in /contracts |
| XIV. Demonstrable Verification | PASS | E2E tests for critical paths |
| XV. Submodule-Centric Development | PASS | Code in aci-frontend/ |
| XVI. Product Manager Ownership | PASS | PM gates defined |
| XVII. Post-Wave Review | PASS | 6-agent review after each wave |

## PM Review Gates

*Per Constitution Principle XVI - Product Manager Ownership*

| Gate | Phase | Status | Reviewer | Date |
|------|-------|--------|----------|------|
| **PM-1** | Pre-Implementation | [ ] Pending | | |
| **PM-2** | Mid-Implementation | [ ] Pending | | |
| **PM-3** | Pre-Release | [ ] Pending | | |

## Post-Wave Review (MANDATORY)

*Per Constitution Principle XVII - Post-Wave Review & Quality Assurance (NON-NEGOTIABLE)*

**After EVERY wave completion, ALL of the following agents MUST review:**

| Reviewer | Agent | Focus Area | Required |
|----------|-------|------------|----------|
| Product Manager | `product-manager-agent` | Business alignment, user value, scope compliance | YES |
| UI Designer | `ui-ux-designer` | Visual design, layout, component consistency | YES |
| UX Designer | `ux-designer` | Usability, user flows, accessibility | YES |
| Visualization | `reviz-visualization` | Charts, graphs, data visualization quality | YES |
| Code Reviewer | `code-reviewer` | Code quality, patterns, maintainability | YES |
| Security Reviewer | `security-reviewer` | Security vulnerabilities, OWASP compliance | YES |

**Requirements:**
- All 6 reviewers must complete review before wave is marked complete
- All task ratings must be ≥ 5 for wave to pass
- Checklist sign-offs required per spec requirements
- Wave summary report created in `specs/002-nexus-frontend/wave-reports/wave-N-report.md`

**PM-1 Deliverables** (Required before Phase 2):
- [ ] Approved spec with prioritized backlog
- [ ] Success metrics defined
- [ ] Gap analysis completed (Critical items addressed)
- [ ] pm-review.md created in `specs/002-nexus-frontend/`

**PM-2 Deliverables** (Required at Phase 3 midpoint):
- [ ] Feature completeness check
- [ ] Scope validation (no creep)
- [ ] Risk assessment updated

**PM-3 Deliverables** (Required before deployment):
- [ ] UAT sign-off
- [ ] Launch readiness confirmed
- [ ] Documentation approval
- [ ] Product verification checklist completed (60+ items)

## Project Structure

### Documentation (this feature)

```text
specs/002-nexus-frontend/
├── spec.md              # Feature specification
├── plan.md              # This file
├── research.md          # Technology decisions
├── data-model.md        # TypeScript type definitions
├── quickstart.md        # Developer onboarding guide
├── contracts/           # API contracts
│   ├── frontend-api-client.md
│   └── websocket-protocol.md
├── checklists/          # Review checklists
│   ├── ux.md
│   ├── ux-api-auth-websockets.md
│   └── pm-ux-verification.md
├── wave-reports/        # Post-wave review summaries
└── tasks.md             # Task breakdown (generated by /speckit.tasks)
```

### Source Code

```text
aci-frontend/
├── src/
│   ├── components/
│   │   ├── ui/              # shadcn/ui components
│   │   ├── charts/          # Reviz chart components
│   │   ├── layout/          # Header, Sidebar, Footer
│   │   └── threat/          # Threat-specific components
│   ├── pages/               # Route page components
│   ├── services/            # API and WebSocket clients
│   │   ├── api/             # REST API client
│   │   └── websocket/       # WebSocket client
│   ├── hooks/               # Custom React hooks
│   ├── stores/              # React Context providers
│   ├── types/               # TypeScript definitions
│   ├── utils/               # Helper functions
│   └── mocks/               # MSW mock handlers
├── tests/
│   ├── unit/                # Component unit tests
│   ├── integration/         # User flow tests
│   └── e2e/                 # Playwright E2E tests
├── public/                  # Static assets
└── package.json
```

**Structure Decision**: Standalone React + Vite frontend application. No monorepo needed - single frontend consuming aci-backend REST API and WebSocket.

## Implementation Phases

### Phase 1: Project Setup & Infrastructure

**Goal**: Bootstrapped React project with core dependencies and tooling

- Vite 7 + React 19 + TypeScript 5.9 project initialization
- Tailwind CSS configuration with dark theme
- shadcn/ui installation and component setup
- ESLint + Prettier configuration
- Vitest + React Testing Library setup
- Basic project structure creation

### Phase 2: Core Foundation

**Goal**: Authentication, routing, and base layout ready

- React Router v7 with protected routes
- Auth context with HttpOnly cookie handling
- Base layout (Header, Sidebar, Footer)
- Error boundary implementation
- Loading states and skeleton components
- TanStack Query client configuration

### Phase 3: User Story Implementation (P1)

**Goal**: Critical P1 user stories functional

- **US1**: Dashboard with metric cards, severity chart, timeline, activity feed
- **US2**: Threat list with filters, search, pagination
- **US3**: Threat detail view with CVEs, bookmarking

### Phase 4: User Story Implementation (P2)

**Goal**: P2 user stories functional

- **US4**: Real-time notifications via WebSocket
- **US5**: Bookmark management page
- **US6**: Alert configuration and management

### Phase 5: User Story Implementation (P3) & Polish

**Goal**: P3 stories and production readiness

- **US7**: Analytics page with trend charts
- **US8**: Admin content review queue
- Performance optimization
- Accessibility audit and fixes
- E2E test coverage
- Documentation finalization

## Key Dependencies

| Dependency | Version | Purpose |
|------------|---------|---------|
| react | 19.2 | UI framework |
| react-dom | 19.2 | DOM rendering |
| vite | 7.2 | Build tool |
| typescript | 5.9 | Type safety |
| tailwindcss | 4.0 | Styling |
| @tanstack/react-query | 5.x | Server state |
| react-router-dom | 7.x | Routing |
| reviz | latest | Charts |
| @radix-ui/* | latest | Accessible primitives |
| vitest | latest | Unit testing |
| @playwright/test | latest | E2E testing |
| @opentelemetry/api | latest | Observability |

## Risk Mitigation

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Backend API delays | Medium | High | MSW mocking for development |
| WebSocket reliability | Medium | Medium | Robust reconnection with UI feedback |
| Bundle size growth | Low | Medium | Size budget alerts in CI |
| Browser compatibility | Low | Low | Playwright cross-browser tests |

## Generated Artifacts

- [research.md](./research.md) - Technology decisions
- [data-model.md](./data-model.md) - TypeScript type definitions
- [quickstart.md](./quickstart.md) - Developer onboarding
- [contracts/frontend-api-client.md](./contracts/frontend-api-client.md) - REST API contract
- [contracts/websocket-protocol.md](./contracts/websocket-protocol.md) - WebSocket protocol

## Next Steps

1. Run `/speckit.tasks` to generate detailed task breakdown
2. Complete PM-1 review gate
3. Begin Phase 1 implementation
