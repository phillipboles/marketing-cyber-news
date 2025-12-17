# Tasks: Article Approval Workflow

**Input**: Design documents from `/specs/003-article-approval-workflow/`
**Prerequisites**: plan.md, spec.md, data-model.md, contracts/approval-api.yaml, research.md, quickstart.md
**Tech Stack**: Go 1.22+ (backend), TypeScript 5.9 (frontend), PostgreSQL 14+, React 19.2, TanStack Query v5

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## User Story Mapping

| ID | Story | Priority | Description |
|----|-------|----------|-------------|
| US1 | Marketing Approval | P1 | Marketing gate approval/rejection |
| US2 | Sequential Gate Progression | P1 | 5-gate sequential workflow |
| US3 | Article Release | P1 | Release fully-approved articles |
| US4 | Rejection from Pipeline | P2 | Reject with reason, remove from queue |
| US5 | Super Admin Override | P2 | Multi-gate approval power |
| US6 | Admin Role Management | P2 | Assign/change user roles |
| US7 | Approval Audit Trail | P3 | View approval history |

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Database migration and domain type setup
**Parallel Agents**: 4 agents can work on T001-T004 simultaneously

### Wave 1.1 - Migrations & Types (Parallel)

- [ ] T001 [P] Create database migration file `aci-backend/migrations/000007_approval_workflow.up.sql`
- [ ] T002 [P] Create rollback migration file `aci-backend/migrations/000007_approval_workflow.down.sql`
- [ ] T003 [P] Create approval domain types in `aci-backend/internal/domain/approval.go`
- [ ] T004 [P] Create approval TypeScript types in `aci-frontend/src/types/approval.ts`

### Wave 1.2 - Migration Execution & Seed Data (Sequential)

- [ ] T005 Run database migration in Kubernetes: `kubectl exec -n aci-backend deploy/postgres -- psql -U aci_user -d aci_db -f /migrations/000007_approval_workflow.up.sql`
- [ ] T005a [P] Create test user seed script for all approver roles in `aci-backend/scripts/seed_approval_users.sql`

### Post-Wave 1 Review & Commit

- [ ] T006-R1 [US1-7] Code review with `code-reviewer` agent - Wave 1 artifacts
- [ ] T006-R2 [US1-7] Security review with `security-reviewer` agent - Wave 1 artifacts
- [ ] T006-GIT Git commit with `git-manager` agent: "feat(approval): add database migrations and domain types for approval workflow"

**Checkpoint**: Phase 1 complete - Database schema and types ready

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story
**Parallel Agents**: 6 agents can work on T007-T012a simultaneously

### Wave 2.1 - Backend Repository & Middleware (Parallel)

- [ ] T007 [P] Create approval repository interface in `aci-backend/internal/repository/approval_repository.go`
- [ ] T008 [P] Implement approval repository in `aci-backend/internal/repository/postgres/approval_repo.go`
- [ ] T009 [P] Create role authorization middleware in `aci-backend/internal/api/middleware/role_auth.go`
- [ ] T010 [P] Create approval DTOs in `aci-backend/internal/api/dto/approval_dto.go`
- [ ] T010a [P] Create unit tests for approval domain types (4-case coverage: happy, fail, null, edge) in `aci-backend/internal/domain/approval_test.go`

### Wave 2.2 - Approval Service (Sequential - depends on 2.1)

- [ ] T011 Create approval service with business logic in `aci-backend/internal/service/approval_service.go`

### Wave 2.3 - Frontend API & Hooks (Parallel)

- [ ] T012 [P] Create approval API service in `aci-frontend/src/services/api/approvals.ts`
- [ ] T013 [P] Create approval TanStack Query hooks in `aci-frontend/src/hooks/useApprovalQueue.ts`
- [ ] T013a [P] Create MSW mock handlers for approval endpoints in `aci-frontend/src/mocks/handlers/approvals.ts`
- [ ] T013b [P] Ensure branding logos from `aci-frontend/public/branding/logos/` are used on login page and site header

### Wave 2.4 - Route Registration (Sequential)

- [ ] T014 Register approval routes in `aci-backend/internal/api/router.go`

### Post-Wave 2 Review & Commit

- [ ] T015-R1 [US1-7] Code review with `code-reviewer` agent - Wave 2 artifacts
- [ ] T015-R2 [US1-7] Security review with `security-reviewer` agent - Wave 2 artifacts
- [ ] T015-GIT Git commit with `git-manager` agent: "feat(approval): add approval repository, service, and frontend hooks"

---

## PM-1 Gate (Required before Phase 3)

- [ ] T016 PM-1: Verify spec approval with prioritized backlog
- [ ] T017 PM-1: Confirm success metrics defined and measurable
- [ ] T018 PM-1: Verify gap analysis completed (Critical items addressed)
- [ ] T019 PM-1: Ensure pm-review.md exists in `specs/003-article-approval-workflow/`
- [ ] T020 PM-1: Obtain PM sign-off for user story implementation

**Checkpoint**: Foundation ready AND PM-1 passed - user story implementation can now begin

---

## Phase 3: User Story 1 - Marketing Approval (Priority: P1)

**Goal**: Marketing team can view queue and approve/reject articles at marketing gate
**Independent Test**: Login as marketing role, view pending articles, approve one, verify status changes to pending_branding
**Parallel Agents**: Backend (3 agents), Frontend (4 agents), Tests (3 agents)

### Wave 3.1 - US1 Backend Handlers (Parallel)

- [ ] T021 [P] [US1] Implement GET /api/v1/approvals/queue handler in `aci-backend/internal/api/handlers/approval_handler.go`
- [ ] T022 [P] [US1] Implement POST /api/v1/articles/{id}/approve handler in `aci-backend/internal/api/handlers/approval_handler.go`
- [ ] T023 [P] [US1] Implement POST /api/v1/articles/{id}/reject handler in `aci-backend/internal/api/handlers/approval_handler.go`

### Wave 3.2 - US1 Backend Logic (Sequential - depends on 3.1)

- [ ] T024 [US1] Add approval queue query to repository with status filter in `aci-backend/internal/repository/postgres/approval_repo.go`
- [ ] T025 [US1] Implement role-to-gate validation in approval service `aci-backend/internal/service/approval_service.go`
- [ ] T026 [US1] Add audit logging for approve/reject actions in `aci-backend/internal/service/approval_service.go`

### Wave 3.3 - US1 Frontend Components (Parallel)

- [ ] T027 [P] [US1] Create ApprovalQueue component in `aci-frontend/src/components/approval/ApprovalQueue.tsx`
- [ ] T028 [P] [US1] Create ApprovalCard component in `aci-frontend/src/components/approval/ApprovalCard.tsx`
- [ ] T029 [P] [US1] Create ApproveButton with confirmation dialog in `aci-frontend/src/components/approval/ApproveButton.tsx`
- [ ] T030 [P] [US1] Create RejectButton with reason modal in `aci-frontend/src/components/approval/RejectButton.tsx`

### Wave 3.4 - US1 Page Integration (Sequential - depends on 3.3)

- [ ] T031 [US1] Create ApprovalPage integrating queue and action components in `aci-frontend/src/pages/ApprovalPage.tsx`
- [ ] T032 [US1] Add /approvals route to App.tsx with role-based protection in `aci-frontend/src/App.tsx`

### Wave 3.5 - US1 Playwright E2E Tests (Parallel)

- [ ] T033 [P] [US1] Create Playwright test for marketing login and queue display in `aci-frontend/tests/e2e/approval-marketing.spec.ts`
- [ ] T034 [P] [US1] Create Playwright test for approve action flow in `aci-frontend/tests/e2e/approval-marketing.spec.ts`
- [ ] T035 [P] [US1] Create Playwright test for reject action with reason in `aci-frontend/tests/e2e/approval-marketing.spec.ts`

### Post-Wave 3 Review & Commit

- [ ] T036-R1 [US1] Code review with `code-reviewer` agent - US1 complete implementation
- [ ] T036-R2 [US1] Security review with `security-reviewer` agent - US1 complete implementation
- [ ] T036-GIT Git commit with `git-manager` agent: "feat(approval): implement marketing approval gate (US1)"

**Checkpoint**: User Story 1 complete - Marketing approval workflow functional

---

## Phase 4: User Story 2 - Sequential Gate Progression (Priority: P1)

**Goal**: Articles progress through all 5 gates in sequence: Marketing → Branding → SOC L1 → SOC L3 → CISO
**Independent Test**: Complete marketing approval, login as branding, approve, verify status changes to pending_soc_l1
**Parallel Agents**: Backend (3 agents), Frontend (3 agents), Tests (2 agents)

### Wave 4.1 - US2 Backend (Parallel for state transition, Sequential for validation)

- [ ] T037 [P] [US2] Implement status transition logic in approval service `aci-backend/internal/service/approval_service.go`
- [ ] T038 [P] [US2] Add gate order validation (prevent skipping gates) in `aci-backend/internal/service/approval_service.go`
- [ ] T039 [US2] Create article_approvals record on each gate approval in `aci-backend/internal/repository/postgres/approval_repo.go`

### Wave 4.2 - US2 Frontend (Parallel)

- [ ] T040 [P] [US2] Create ApprovalProgress component (green/yellow/gray gates) in `aci-frontend/src/components/approval/ApprovalProgress.tsx`
- [ ] T041 [P] [US2] Update ApprovalCard to show progress indicator in `aci-frontend/src/components/approval/ApprovalCard.tsx`
- [ ] T042 [US2] Update queue hook to filter by user's role gate in `aci-frontend/src/hooks/useApprovalQueue.ts`

### Wave 4.3 - US2 Playwright E2E Tests (Parallel)

- [ ] T043 [P] [US2] Create Playwright test for full 5-gate progression in `aci-frontend/tests/e2e/approval-progression.spec.ts`
- [ ] T044 [P] [US2] Create Playwright test for gate skip prevention (403 error) in `aci-frontend/tests/e2e/approval-progression.spec.ts`

### Post-Wave 4 Review & Commit

- [ ] T045-R1 [US2] Code review with `code-reviewer` agent - US2 complete implementation
- [ ] T045-R2 [US2] Security review with `security-reviewer` agent - US2 complete implementation
- [ ] T045-GIT Git commit with `git-manager` agent: "feat(approval): implement sequential gate progression (US2)"

**Checkpoint**: User Story 2 complete - Full sequential gate progression working

---

## Phase 5: User Story 3 - Article Release (Priority: P1)

**Goal**: Admin/CISO/super_admin can release fully-approved articles for public viewing
**Independent Test**: Complete all 5 gates, login as admin, click Release, verify article visible to regular users
**Parallel Agents**: Backend (3 agents), Frontend (3 agents), Tests (2 agents)

### Wave 5.1 - US3 Backend (Parallel)

- [ ] T046 [P] [US3] Implement POST /api/v1/articles/{id}/release handler in `aci-backend/internal/api/handlers/approval_handler.go`
- [ ] T047 [P] [US3] Add release validation (must be in 'approved' status) in `aci-backend/internal/service/approval_service.go`
- [ ] T048 [P] [US3] Update article query to filter by released status for public feed in `aci-backend/internal/repository/postgres/article_repo.go`

### Wave 5.2 - US3 Frontend (Parallel then Sequential)

- [ ] T049 [P] [US3] Create ReleaseButton component in `aci-frontend/src/components/approval/ReleaseButton.tsx`
- [ ] T050 [US3] Add release action to approved articles in ApprovalPage in `aci-frontend/src/pages/ApprovalPage.tsx`
- [ ] T051 [US3] Update threat list to only show released articles for non-admin users in `aci-frontend/src/pages/ThreatsPage.tsx`

### Wave 5.3 - US3 Playwright E2E Tests (Parallel)

- [ ] T052 [P] [US3] Create Playwright test for release action on fully-approved article in `aci-frontend/tests/e2e/approval-release.spec.ts`
- [ ] T053 [P] [US3] Create Playwright test verifying released article visible to regular user in `aci-frontend/tests/e2e/approval-release.spec.ts`

### Post-Wave 5 Review & Commit

- [ ] T054-R1 [US3] Code review with `code-reviewer` agent - US3 complete implementation
- [ ] T054-R2 [US3] Security review with `security-reviewer` agent - US3 complete implementation
- [ ] T054-GIT Git commit with `git-manager` agent: "feat(approval): implement article release (US3) - MVP complete"

**Checkpoint**: User Story 3 complete - P1 stories all functional, MVP ready for demo

---

## Phase 6: User Story 4 - Rejection from Pipeline (Priority: P2)

**Goal**: Any approver can reject articles with mandatory reason, removing them from all queues
**Independent Test**: Login as SOC L3, reject an article, verify it disappears from all queues and shows rejection reason
**Parallel Agents**: Backend (3 agents), Frontend (2 agents), Tests (2 agents)

### Wave 6.1 - US4 Backend (Parallel)

- [ ] T055 [P] [US4] Implement rejection with mandatory reason validation in `aci-backend/internal/service/approval_service.go`
- [ ] T056 [P] [US4] Update queue queries to exclude rejected articles in `aci-backend/internal/repository/postgres/approval_repo.go`
- [ ] T057 [P] [US4] Store rejection metadata (reason, rejector, timestamp) in `aci-backend/internal/repository/postgres/approval_repo.go`

### Wave 6.2 - US4 Frontend (Parallel)

- [ ] T058 [P] [US4] Update RejectButton to require reason (min 10 chars) in `aci-frontend/src/components/approval/RejectButton.tsx`
- [ ] T059 [P] [US4] Add rejection details display to article view in `aci-frontend/src/components/approval/ApprovalCard.tsx`

### Wave 6.3 - US4 Playwright E2E Tests (Parallel)

- [ ] T060 [P] [US4] Create Playwright test for rejection with reason in `aci-frontend/tests/e2e/approval-rejection.spec.ts`
- [ ] T061 [P] [US4] Create Playwright test verifying rejected article removed from queues in `aci-frontend/tests/e2e/approval-rejection.spec.ts`

### Post-Wave 6 Review & Commit

- [ ] T062-R1 [US4] Code review with `code-reviewer` agent - US4 complete implementation
- [ ] T062-R2 [US4] Security review with `security-reviewer` agent - US4 complete implementation
- [ ] T062-GIT Git commit with `git-manager` agent: "feat(approval): implement rejection from pipeline (US4)"

**Checkpoint**: User Story 4 complete - Rejection workflow functional

---

## Phase 7: User Story 5 - Super Admin Override (Priority: P2)

**Goal**: Super_admin can approve at any gate and has CISO release power
**Independent Test**: Login as super_admin, view all pending articles across gates, approve through multiple gates sequentially
**Parallel Agents**: Backend (2 agents), Frontend (2 agents), Tests (1 agent)

### Wave 7.1 - US5 Backend (Parallel)

- [ ] T063 [P] [US5] Update role authorization to grant super_admin all gate access in `aci-backend/internal/api/middleware/role_auth.go`
- [ ] T064 [P] [US5] Update queue endpoint to show all gates for admin/super_admin in `aci-backend/internal/api/handlers/approval_handler.go`

### Wave 7.2 - US5 Frontend (Parallel)

- [ ] T065 [P] [US5] Update queue filtering to show all pending articles for super_admin in `aci-frontend/src/hooks/useApprovalQueue.ts`
- [ ] T066 [P] [US5] Add gate selector for admin/super_admin in ApprovalPage in `aci-frontend/src/pages/ApprovalPage.tsx`

### Wave 7.3 - US5 Playwright E2E Tests

- [ ] T067 [US5] Create Playwright test for super_admin multi-gate approval in `aci-frontend/tests/e2e/approval-superadmin.spec.ts`

### Post-Wave 7 Review & Commit

- [ ] T068-R1 [US5] Code review with `code-reviewer` agent - US5 complete implementation
- [ ] T068-R2 [US5] Security review with `security-reviewer` agent - US5 complete implementation
- [ ] T068-GIT Git commit with `git-manager` agent: "feat(approval): implement super admin override (US5)"

**Checkpoint**: User Story 5 complete - Super admin override working

---

## Phase 8: User Story 6 - Admin Role Management (Priority: P2)

**Goal**: Admin can assign/change user roles for approval workflow
**Independent Test**: Login as admin, change user from 'user' to 'marketing', logout, login as that user, verify marketing queue access
**Parallel Agents**: Backend (3 agents), Frontend (2 agents), Tests (1 agent)

### Wave 8.1 - US6 Backend (Parallel)

- [ ] T069 [P] [US6] Implement PUT /api/v1/users/{id}/role handler in `aci-backend/internal/api/handlers/admin_handler.go`
- [ ] T070 [P] [US6] Add admin-only authorization check for role change in `aci-backend/internal/api/middleware/role_auth.go`
- [ ] T071 [P] [US6] Add audit logging for role changes in `aci-backend/internal/service/user_service.go`

### Wave 8.2 - US6 Frontend (Parallel)

- [ ] T072 [P] [US6] Create RoleSelector component in `aci-frontend/src/components/admin/RoleSelector.tsx`
- [ ] T073 [US6] Add role management section to AdminPage in `aci-frontend/src/pages/AdminPage.tsx`

### Wave 8.3 - US6 Playwright E2E Tests

- [ ] T074 [US6] Create Playwright test for role assignment and verification in `aci-frontend/tests/e2e/admin-role-management.spec.ts`

### Post-Wave 8 Review & Commit

- [ ] T075-R1 [US6] Code review with `code-reviewer` agent - US6 complete implementation
- [ ] T075-R2 [US6] Security review with `security-reviewer` agent - US6 complete implementation
- [ ] T075-GIT Git commit with `git-manager` agent: "feat(approval): implement admin role management (US6)"

**Checkpoint**: User Story 6 complete - Role management functional

---

## Phase 9: User Story 7 - Approval Audit Trail (Priority: P3)

**Goal**: View complete approval history for any article with timestamps and approvers
**Independent Test**: View approved article, open history modal, verify all 5 gate approvals listed with correct approvers
**Parallel Agents**: Backend (2 agents), Frontend (3 agents), Tests (1 agent)

### Wave 9.1 - US7 Backend (Parallel)

- [ ] T076 [P] [US7] Implement GET /api/v1/articles/{id}/approval-history handler in `aci-backend/internal/api/handlers/approval_handler.go`
- [ ] T077 [P] [US7] Create approval history query with approver names in `aci-backend/internal/repository/postgres/approval_repo.go`

### Wave 9.2 - US7 Frontend (Parallel)

- [ ] T078 [P] [US7] Create ApprovalHistoryModal component in `aci-frontend/src/components/approval/ApprovalHistoryModal.tsx`
- [ ] T079 [P] [US7] Create useApprovalHistory hook in `aci-frontend/src/hooks/useApprovalHistory.ts`
- [ ] T080 [US7] Add history button to ApprovalCard in `aci-frontend/src/components/approval/ApprovalCard.tsx`

### Wave 9.3 - US7 Playwright E2E Tests

- [ ] T081 [US7] Create Playwright test for approval history display in `aci-frontend/tests/e2e/approval-history.spec.ts`

### Post-Wave 9 Review & Commit

- [ ] T082-R1 [US7] Code review with `code-reviewer` agent - US7 complete implementation
- [ ] T082-R2 [US7] Security review with `security-reviewer` agent - US7 complete implementation
- [ ] T082-GIT Git commit with `git-manager` agent: "feat(approval): implement approval audit trail (US7)"

**Checkpoint**: User Story 7 complete - All user stories implemented

---

## Phase 10: PM-2 Gate Review

**Purpose**: Mid-implementation PM alignment check

- [ ] T083 Feature completeness check - verify P1 stories functional (US1, US2, US3)
- [ ] T084 Scope validation - confirm no scope creep
- [ ] T085 Risk assessment - document implementation risks
- [ ] T086 PM-2 sign-off obtained (document in pm-review.md)

**Checkpoint**: PM-2 gate passed

---

## Phase 11: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories
**Parallel Agents**: 6 agents can work simultaneously

### Wave 11.1 - Polish Tasks (Parallel)

- [ ] T087 [P] Implement POST /api/v1/articles/{id}/reset handler for admin article reset in `aci-backend/internal/api/handlers/approval_handler.go`
- [ ] T088 [P] Add toast notifications for all approval actions in `aci-frontend/src/components/approval/`
- [ ] T089 [P] Add loading skeletons to ApprovalQueue in `aci-frontend/src/components/approval/ApprovalQueue.tsx`
- [ ] T090 [P] Add empty state UI for queues with no pending articles in `aci-frontend/src/components/approval/ApprovalQueue.tsx`
- [ ] T091 [P] Update API documentation in `aci-backend/docs/API.md`

### Wave 11.2 - Quality Assurance (Sequential)

- [ ] T092 Code cleanup - ensure no nested ifs, guard clauses used throughout
- [ ] T093 Verify all config values externalized (no hardcoded values)
- [ ] T094 Run quickstart.md validation against K8s deployment

### Post-Wave 11 Review & Commit

- [ ] T095-R1 Code review with `code-reviewer` agent - Polish artifacts
- [ ] T095-R2 Security review with `security-reviewer` agent - Polish artifacts
- [ ] T095-GIT Git commit with `git-manager` agent: "chore(approval): polish and cross-cutting improvements"

---

## Phase 12: PM-3 Gate & Release Verification

**Purpose**: Final PM verification before deployment

### Wave 12.1 - Final Testing (Parallel)

- [ ] T096 [P] UAT sign-off - all acceptance scenarios pass via Playwright
- [ ] T097 [P] User journey validation - end-to-end testing complete with real data
- [ ] T098 [P] Documentation approval - README, API docs, quickstart complete

### Wave 12.2 - Performance & Security (Parallel)

- [ ] T099 [P] Performance verification - <2s queue load, <1s approval action
- [ ] T100 [P] Security validation - RBAC enforcement, audit logging verified

### Wave 12.3 - Final Regression (Sequential)

- [ ] T101 Playwright E2E full regression - all tests pass against K8s
- [ ] T102 Product verification checklist completed (60+ items)
- [ ] T103 PM-3 sign-off obtained (document in pm-review.md)

### Final Commit

- [ ] T104-GIT Git commit with `git-manager` agent: "feat(approval): article approval workflow feature complete"

**Checkpoint**: PM-3 gate passed - ready for production deployment

---

## Dependencies & Execution Order

### Phase Dependencies

```
Phase 1 (Setup) → Phase 2 (Foundational) → PM-1 Gate
                                               ↓
                                    [User Stories Can Begin]
                                               ↓
Phase 3 (US1) ─┬─→ Phase 4 (US2) ─→ Phase 5 (US3) ─→ [MVP Complete]
               │
               └─→ Phase 6 (US4) ─┬─→ Phase 7 (US5)
                                  │
                                  └─→ Phase 8 (US6) ─→ Phase 9 (US7)
                                               ↓
                                    Phase 10 (PM-2) → Phase 11 (Polish) → Phase 12 (PM-3)
```

### User Story Dependencies

| Story | Can Start After | Dependencies |
|-------|-----------------|--------------|
| US1 | Phase 2 complete | None - first gate |
| US2 | US1 complete | Needs marketing gate working |
| US3 | US2 complete | Needs all gates working |
| US4 | Phase 2 complete | None - independent |
| US5 | US1 complete | Needs at least one gate |
| US6 | Phase 2 complete | None - independent |
| US7 | US2 complete | Needs approval records |

### Parallel Execution Guide

```bash
# Phase 1 - 4 agents
Agent 1: T001 (up.sql migration)
Agent 2: T002 (down.sql migration)
Agent 3: T003 (Go domain types)
Agent 4: T004 (TS types)

# Phase 2 - Wave 2.1: 5 agents
Agent 1: T007 (repository interface)
Agent 2: T008 (repository impl)
Agent 3: T009 (role middleware)
Agent 4: T010 (DTOs)
Agent 5: T010a (unit tests)

# Phase 2 - Wave 2.3: 4 agents
Agent 1: T012 (API service)
Agent 2: T013 (TanStack hooks)
Agent 3: T013a (MSW mocks)
Agent 4: T013b (branding)

# US1 Backend - 3 agents
Agent 1: T021 (GET queue handler)
Agent 2: T022 (POST approve handler)
Agent 3: T023 (POST reject handler)

# US1 Frontend - 4 agents
Agent 1: T027 (ApprovalQueue)
Agent 2: T028 (ApprovalCard)
Agent 3: T029 (ApproveButton)
Agent 4: T030 (RejectButton)

# US1 E2E Tests - 3 agents
Agent 1: T033 (login/queue test)
Agent 2: T034 (approve test)
Agent 3: T035 (reject test)

# Independent User Stories (after foundation)
Stream A: US1 → US2 → US3 (P1 MVP path)
Stream B: US4 (can start after Phase 2)
Stream C: US6 (can start after Phase 2)
Stream D: US5 (can start after US1)
Stream E: US7 (can start after US2)
```

---

## Summary

| Phase | Tasks | Review/Git | Stories | Description |
|-------|-------|------------|---------|-------------|
| 1 | 6 | 3 | - | Setup & Migrations |
| 2 | 9 | 3 | - | Foundational Infrastructure |
| PM-1 | 5 | - | - | Pre-Implementation Gate |
| 3 | 15 | 3 | US1 | Marketing Approval |
| 4 | 8 | 3 | US2 | Sequential Progression |
| 5 | 8 | 3 | US3 | Article Release |
| 6 | 7 | 3 | US4 | Rejection |
| 7 | 5 | 3 | US5 | Super Admin Override |
| 8 | 6 | 3 | US6 | Role Management |
| 9 | 6 | 3 | US7 | Audit Trail |
| PM-2 | 4 | - | - | Mid-Implementation Gate |
| 11 | 8 | 3 | - | Polish |
| PM-3 | 8 | 1 | - | Release Gate |
| **Total** | **95** | **31** | **7** | **126 total** |

### Task Counts by Story

| Story | Backend | Frontend | E2E Tests | Review | Git | Total |
|-------|---------|----------|-----------|--------|-----|-------|
| US1 | 6 | 6 | 3 | 2 | 1 | 18 |
| US2 | 3 | 3 | 2 | 2 | 1 | 11 |
| US3 | 3 | 3 | 2 | 2 | 1 | 11 |
| US4 | 3 | 2 | 2 | 2 | 1 | 10 |
| US5 | 2 | 2 | 1 | 2 | 1 | 8 |
| US6 | 3 | 2 | 1 | 2 | 1 | 9 |
| US7 | 2 | 3 | 1 | 2 | 1 | 9 |

### MVP Scope (P1 Stories Only)

Complete Phases 1-5 for minimum viable product:
- Marketing approval working
- Full 5-gate progression
- Article release for approved content
- **~50 tasks to MVP** (includes review, git, test/mock/branding tasks)

### Agent Utilization Summary

| Phase | Max Parallel Agents | Notes |
|-------|---------------------|-------|
| Phase 1 | 4 | T001-T004 all parallel |
| Phase 2 | 6 | Waves 2.1 + 2.3 have parallel tasks |
| Phase 3 | 4 | Backend handlers + frontend components |
| Phase 4-9 | 2-3 | Smaller focused waves |
| Phase 11 | 6 | Polish tasks all parallel |
| Phase 12 | 3 | Final testing parallel |
