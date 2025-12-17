# Feature Specification: Article Approval Workflow

**Feature Branch**: `003-article-approval-workflow`
**Created**: 2025-12-16
**Status**: Draft
**Input**: Multi-gate approval process for articles with role-based access control

## Overview

The Article Approval Workflow is a multi-gate content approval system that requires articles to pass through 5 sequential approval stages before release. Each stage has a designated approver role, and articles can be approved (gate passed) or rejected (removed from pipeline). The system includes 7 user roles with hierarchical permissions, providing a robust role-based access control (RBAC) foundation for the platform.

### Approval Gates (Sequential Order)

1. **Marketing** - Ensures content aligns with marketing guidelines and messaging
2. **Branding** - Verifies brand consistency and voice
3. **SOC Level 1** - Initial security operations center review
4. **SOC Level 3** - Senior security analyst deep review
5. **CISO** - Chief Information Security Officer final approval

### User Roles (7 Total)

| Role | Permission Level | Description |
|------|------------------|-------------|
| `user` | 1 | Standard user, read-only access to approved content |
| `marketing` | 2 | Can approve/reject at Marketing gate |
| `branding` | 3 | Can approve/reject at Branding gate |
| `soc_level_1` | 4 | Can approve/reject at SOC Level 1 gate |
| `soc_level_3` | 5 | Can approve/reject at SOC Level 3 gate |
| `ciso` | 6 | Can approve/reject at CISO gate |
| `admin` | 7 | Full system administration, all gate access |
| `super_admin` | 8 | Same as CISO power + admin capabilities |

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Marketing Approval (Priority: P1)

A marketing team member wants to review articles awaiting marketing approval and approve or reject them based on content alignment with marketing guidelines.

**Why this priority**: Marketing is the first gate in the pipeline - no articles can proceed without marketing approval.

**Independent Test**: Can be fully tested by logging in as marketing role, viewing pending articles, and approving/rejecting one.

**Acceptance Scenarios**:

1. **Given** an article in "pending_marketing" status, **When** a marketing user views the approval queue, **Then** the article appears in their queue with full content visible.
2. **Given** a marketing user viewing an article, **When** they click "Approve", **Then** the article status changes to "pending_branding" and an approval record is created.
3. **Given** a marketing user viewing an article, **When** they click "Reject" with a reason, **Then** the article is marked as "rejected" with a boolean flag and the reason is recorded.
4. **Given** a user without marketing role, **When** they attempt to approve at marketing gate, **Then** they receive 403 Forbidden.

---

### User Story 2 - Sequential Gate Progression (Priority: P1)

A branding team member wants to review articles that have passed marketing approval and are now awaiting branding review.

**Why this priority**: Sequential flow is fundamental to the workflow - gates must be processed in order.

**Independent Test**: Can be fully tested by completing marketing approval, then verifying the article appears in branding queue.

**Acceptance Scenarios**:

1. **Given** an article approved by marketing, **When** a branding user views their queue, **Then** only articles in "pending_branding" status appear.
2. **Given** an article in "pending_branding", **When** branding user approves it, **Then** status changes to "pending_soc_l1".
3. **Given** an article in "pending_soc_l1", **When** SOC L1 user approves it, **Then** status changes to "pending_soc_l3".
4. **Given** an article in "pending_soc_l3", **When** SOC L3 user approves it, **Then** status changes to "pending_ciso".
5. **Given** an article in "pending_ciso", **When** CISO approves it, **Then** status changes to "approved" and all gates are marked complete.

---

### User Story 3 - Article Release (Priority: P1)

An admin, CISO, or super_admin wants to release an article that has passed all 5 approval gates so it becomes visible to regular users.

**Why this priority**: Release is the final step that delivers value to end users.

**Independent Test**: Can be fully tested by having an article pass all gates, then releasing it and verifying visibility.

**Acceptance Scenarios**:

1. **Given** an article with status "approved" (all gates passed), **When** an admin clicks "Release", **Then** the article becomes published and visible to all users.
2. **Given** an article without all gates passed, **When** a user attempts to release it, **Then** the system prevents release with clear error message.
3. **Given** a released article, **When** any user browses the article list, **Then** the article appears in their feed.
4. **Given** an unreleased article (any status except "released"), **When** a regular user browses articles, **Then** the article does not appear.

---

### User Story 4 - Rejection from Pipeline (Priority: P2)

A SOC Level 3 analyst determines an article contains inaccurate threat intelligence and wants to reject it from the pipeline entirely.

**Why this priority**: Rejection is essential for quality control but less frequent than approvals.

**Independent Test**: Can be fully tested by rejecting an article at any gate and verifying it cannot proceed.

**Acceptance Scenarios**:

1. **Given** an article at any approval gate, **When** an authorized user clicks "Reject", **Then** the article status changes to "rejected" and `rejected = true` is set.
2. **Given** a rejected article, **When** any approver views their queue, **Then** the rejected article does not appear.
3. **Given** a rejected article, **When** an admin views all articles, **Then** the article appears with "rejected" status visible.
4. **Given** a rejected article with rejection reason, **When** viewing article details, **Then** the rejection reason and rejector information is displayed.

---

### User Story 5 - Super Admin Override (Priority: P2)

A super admin needs to approve an urgent article through all remaining gates immediately due to a critical vulnerability disclosure.

**Why this priority**: Super admin capabilities provide emergency override for critical situations.

**Independent Test**: Can be fully tested by logging in as super_admin and approving an article through multiple gates.

**Acceptance Scenarios**:

1. **Given** a super_admin user, **When** they view any approval queue, **Then** they see articles pending at all gates.
2. **Given** an article at any gate, **When** super_admin approves it, **Then** the approval is recorded as if the appropriate role approved it.
3. **Given** a super_admin user, **When** they approve all remaining gates for an article, **Then** the article becomes "approved" status ready for release.
4. **Given** a super_admin user, **When** they attempt admin functions (user management, etc.), **Then** they have full admin access.

---

### User Story 6 - Admin Role Management (Priority: P2)

An admin wants to assign the "soc_level_1" role to a new security analyst so they can participate in article approvals.

**Why this priority**: Role management enables the approval workflow to function with actual users.

**Independent Test**: Can be fully tested by admin assigning a role and verifying the user can access appropriate queues.

**Acceptance Scenarios**:

1. **Given** an admin user, **When** they view user management, **Then** they see all users with their current roles.
2. **Given** an admin assigning "marketing" role to a user, **When** they save changes, **Then** the user can access the marketing approval queue.
3. **Given** a user with "soc_level_1" role, **When** admin changes it to "soc_level_3", **Then** the user's queue access updates accordingly.
4. **Given** a non-admin user, **When** they attempt to change user roles, **Then** they receive 403 Forbidden.

---

### User Story 7 - Approval Audit Trail (Priority: P3)

A compliance officer wants to view the complete approval history for an article to verify proper review processes were followed.

**Why this priority**: Audit trails are important for compliance but not blocking for core workflow.

**Independent Test**: Can be fully tested by completing approvals and viewing the audit trail for an article.

**Acceptance Scenarios**:

1. **Given** an approved article, **When** viewing its approval history, **Then** all 5 gate approvals are listed with timestamps and approvers.
2. **Given** a rejected article, **When** viewing its history, **Then** the rejection details including reason and rejector are shown.
3. **Given** an article in progress, **When** viewing its status, **Then** completed gates show green, current gate shows yellow, pending gates show gray.
4. **Given** any approval or rejection action, **When** completed, **Then** an audit log entry is created with user, action, timestamp, and article reference.

---

### Edge Cases

- What happens when an approver tries to approve an article not at their gate? System returns 400 Bad Request with clear error.
- What happens when a rejected article needs to be re-submitted? Admin can reset article to "pending_marketing" status.
- What happens when an approver is removed from a role mid-approval? Their previous approvals remain valid, but they can no longer approve new articles.
- What happens when multiple users have the same approver role? Any user with the role can approve - first one wins.
- What happens when the CISO role is vacant? Super_admin can fulfill CISO approval duties.
- What happens when an article is modified after partial approval? System should track if content changed and optionally reset to first gate (configurable).

## Requirements *(mandatory)*

### Functional Requirements

#### Approval Workflow
- **FR-001**: System MUST support 5 sequential approval gates: Marketing, Branding, SOC L1, SOC L3, CISO
- **FR-002**: System MUST enforce gate order - articles cannot skip gates
- **FR-003**: System MUST allow rejection at any gate with mandatory reason field
- **FR-004**: System MUST mark rejected articles with a boolean `rejected = true` flag
- **FR-005**: System MUST track approval status as enum: pending_marketing, pending_branding, pending_soc_l1, pending_soc_l3, pending_ciso, approved, rejected, released
- **FR-006**: System MUST record approver identity and timestamp for each gate approval
- **FR-007**: System MUST support "Release" action only for fully approved articles (all 5 gates passed)

#### Role-Based Access Control
- **FR-008**: System MUST support 7 user roles: user, marketing, branding, soc_level_1, soc_level_3, ciso, admin, super_admin
- **FR-009**: System MUST restrict gate approval to users with matching role (marketing role approves marketing gate, etc.)
- **FR-010**: System MUST grant super_admin same approval powers as ciso
- **FR-011**: System MUST grant admin full system management capabilities
- **FR-012**: System MUST allow admin and super_admin to manage user roles
- **FR-013**: System MUST allow admin to reset rejected articles to initial state

#### Approval Queue Management
- **FR-014**: System MUST provide role-specific approval queues (marketing user sees only marketing-pending articles)
- **FR-015**: System MUST support pagination for approval queues (default 20, max 100)
- **FR-016**: System MUST support sorting approval queues by created_at, severity, category
- **FR-017**: System MUST display approval progress indicator showing completed/pending gates
- **FR-018**: System MUST support filtering approval queue by category, severity, date range

#### Audit & Compliance
- **FR-019**: System MUST create audit log entry for every approval/rejection action
- **FR-020**: System MUST store rejection reasons with rejector identity
- **FR-021**: System MUST provide approval history endpoint for any article
- **FR-022**: System MUST track gate completion timestamps in article record

#### API Endpoints
- **FR-023**: System MUST provide GET /api/v1/approvals/queue endpoint with role-based filtering
- **FR-024**: System MUST provide POST /api/v1/articles/{id}/approve endpoint
- **FR-025**: System MUST provide POST /api/v1/articles/{id}/reject endpoint
- **FR-026**: System MUST provide POST /api/v1/articles/{id}/release endpoint
- **FR-027**: System MUST provide GET /api/v1/articles/{id}/approval-history endpoint
- **FR-028**: System MUST provide PUT /api/v1/users/{id}/role endpoint (admin only)

### Key Entities

- **ApprovalStatus** (enum): pending_marketing, pending_branding, pending_soc_l1, pending_soc_l3, pending_ciso, approved, rejected, released
- **UserRole** (enum): user, marketing, branding, soc_level_1, soc_level_3, ciso, admin, super_admin
- **ArticleApproval**: Junction table linking article to approval gate with approver_id, approved_at timestamp
- **ApprovalGate** (enum): marketing, branding, soc_l1, soc_l3, ciso

### Database Schema Changes

```sql
-- Extend user role enum
ALTER TYPE user_role ADD VALUE IF NOT EXISTS 'marketing';
ALTER TYPE user_role ADD VALUE IF NOT EXISTS 'branding';
ALTER TYPE user_role ADD VALUE IF NOT EXISTS 'soc_level_1';
ALTER TYPE user_role ADD VALUE IF NOT EXISTS 'soc_level_3';
ALTER TYPE user_role ADD VALUE IF NOT EXISTS 'ciso';
ALTER TYPE user_role ADD VALUE IF NOT EXISTS 'super_admin';

-- Add approval status enum
CREATE TYPE approval_status AS ENUM (
  'pending_marketing',
  'pending_branding',
  'pending_soc_l1',
  'pending_soc_l3',
  'pending_ciso',
  'approved',
  'rejected',
  'released'
);

-- Add approval gate enum
CREATE TYPE approval_gate AS ENUM (
  'marketing',
  'branding',
  'soc_l1',
  'soc_l3',
  'ciso'
);

-- Extend articles table
ALTER TABLE articles ADD COLUMN approval_status approval_status DEFAULT 'pending_marketing';
ALTER TABLE articles ADD COLUMN rejected BOOLEAN DEFAULT false;
ALTER TABLE articles ADD COLUMN rejection_reason TEXT;
ALTER TABLE articles ADD COLUMN rejected_by UUID REFERENCES users(id);
ALTER TABLE articles ADD COLUMN rejected_at TIMESTAMPTZ;
ALTER TABLE articles ADD COLUMN released_at TIMESTAMPTZ;
ALTER TABLE articles ADD COLUMN released_by UUID REFERENCES users(id);

-- Article approvals junction table
CREATE TABLE article_approvals (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  article_id UUID NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
  gate approval_gate NOT NULL,
  approved_by UUID NOT NULL REFERENCES users(id),
  approved_at TIMESTAMPTZ DEFAULT NOW(),
  notes TEXT,
  UNIQUE(article_id, gate)
);

-- Indexes for approval queries
CREATE INDEX idx_articles_approval_status ON articles(approval_status);
CREATE INDEX idx_articles_rejected ON articles(rejected) WHERE rejected = true;
CREATE INDEX idx_article_approvals_article ON article_approvals(article_id);
CREATE INDEX idx_article_approvals_gate ON article_approvals(gate);
```

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Approvers can view and action their queue with p95 latency under 2 seconds
- **SC-002**: Approval/rejection actions complete in under 1 second
- **SC-003**: System correctly enforces role-based access 100% of the time
- **SC-004**: All approval actions create audit log entries within 100ms
- **SC-005**: Approval history retrieval completes in under 500ms for articles with full approval chain
- **SC-006**: System handles 100 concurrent approval operations without degradation
- **SC-007**: Zero unauthorized approvals (proper role enforcement)
- **SC-008**: 100% of rejected articles properly flagged and removed from queues
- **SC-009**: Release action only succeeds for fully approved articles
- **SC-010**: 80% or higher test coverage for approval workflow modules

## PM Acceptance Criteria *(mandatory)*

*Per Constitution Principle XVI - Product Manager Ownership*

### PM-1 Gate: Pre-Implementation Approval

- [ ] All user stories have clear acceptance scenarios
- [ ] Priorities (P1, P2, P3) are assigned and justified
- [ ] Edge cases are identified and documented
- [ ] Success metrics are measurable and achievable
- [ ] Out-of-scope items are explicitly declared (see below)
- [ ] Role hierarchy and permissions are clearly defined
- [ ] Approval flow diagram approved by stakeholders

### PM-2 Gate: Mid-Implementation Alignment

- [ ] Feature implementation aligns with original scope
- [ ] No scope creep has occurred (or changes are documented/approved)
- [ ] P1 user stories are functional and testable
- [ ] Role-based access control is properly enforced
- [ ] Risks identified during implementation are tracked

### PM-3 Gate: Pre-Release Verification

- [ ] All acceptance scenarios pass
- [ ] User journeys validated end-to-end
- [ ] Documentation is complete and accurate
- [ ] Performance targets met
- [ ] Security requirements validated (role enforcement, audit logging)
- [ ] Product verification checklist completed

## Out of Scope

The following are explicitly **NOT** included in this feature:

- Email notifications for approval status changes (future enhancement)
- Slack/Teams integration for approval notifications (future enhancement)
- Approval delegation (one user approving on behalf of another)
- Batch approval operations (approve multiple articles at once)
- Approval SLA tracking and alerts (future enhancement)
- Automatic escalation when approval is delayed
- Approval workflow configuration UI (fixed 5-gate flow for MVP)
- Article editing by approvers (read-only review)
- Comments/discussion threads on approvals
- Mobile-optimized approval interface

## Assumptions

- User authentication and JWT tokens are already implemented (001-aci-backend)
- Articles are created via n8n webhook and start in pending_marketing status
- Each user has exactly one role (no multi-role support for MVP)
- Role changes take effect immediately
- Audit log infrastructure exists (000005_audit_schema migration)
- Frontend will be built to support approval workflow UI

## Dependencies

- **001-aci-backend**: Authentication, JWT, existing User and Article models
- **002-nexus-frontend**: UI components for approval workflow
- **PostgreSQL 14+**: Database with enum support
- **Existing audit_logs table**: For compliance tracking

## Risks

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Role assignment errors grant unauthorized access | High | Medium | Comprehensive role validation, audit logging |
| Approval bottlenecks if single approver | Medium | Medium | Allow multiple users per role |
| Articles stuck in rejected state | Low | Low | Admin reset capability |
| Database migration breaks existing users | High | Low | Careful migration with role defaults |
| Concurrent approval race conditions | Medium | Low | Database unique constraints on article_approvals |
