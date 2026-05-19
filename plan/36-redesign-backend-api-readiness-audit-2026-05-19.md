# Hotline Redesign — Backend/API Readiness Audit

Date: 2026-05-19
Status: Audit Complete
Auditor: dev-backend-1

## 1. Existing Backend Modules (all verified, tests green)

| Module | Routes | Status |
|--------|--------|--------|
| auth | `/v1/auth/*` (login, register, refresh, logout, me) | Complete |
| team | `/v1/teams/*` CRUD | Complete |
| teamplan | `/v1/team-plans/*` CRUD | Complete |
| task | `/v1/tasks/*` CRUD + by-team + by-filter | Complete |
| planningcalendar | `/v1/planning/calendar/:year/:month` | Complete |
| monthlyplan | `/v1/monthly-plans/*` (settings, period, files, upload, download, submission) | Complete |
| largework | `/v1/large-work-items/*` + `/v1/large-work-tasks/*` | Complete |
| dailyreportdraft | `/v1/daily-report-drafts/*` (sources, from-plan) | Complete |
| workreport | `/v1/work-reports` (read-only monthly TaskDaily report) | Complete |
| dashboard | `/v1/dashboard/*` (summary, top-jobs, top-feeders, feeder-matrix, stats) | Existing legacy routes; redesign does not use Dashboard |
| user | `/v1/users/*` CRUD, contacts, passwords | Complete |
| contact-directory | `/v1/contact-directory/*` (list, get-by-id) | Complete |
| upload | `/v1/upload/*` (presigned URL, delete) | Complete |
| station, pea, feeder, operationcenter, jobtype, jobdetail | Master data CRUD | Complete |

## 2. Gap Analysis: Redesign Requirements vs Backend

### GAP-1: Viewer Download Restriction (monthly plan) -- FIXED 2026-05-19
**Severity: Medium** | **Risk: Low** | **Effort: Small** | **Status: FIXED**

`GetDownloadURL` in monthly plan controller did NOT block `viewer` role from downloading files.
- `GetFile` checks team ownership via `CanAccessFile`, but viewer passes that check (viewer can read).
- Per Requirement A, viewer should NOT be able to download monthly plan files.
- **Fix applied**: Added `CanDownloadFile()` method to entity policy (blocks viewer) + check in `GetDownloadURL` controller.
- **Tests**: `TestViewerCannotDownloadMonthlyPlanFiles` (entity) + `TestGetDownloadURLBlocksViewerEvenWithMatchingTeam` (controller).

**Files changed:**
- `internal/feature/monthlyplan/entity/policy.go` — added `CanDownloadFile` method
- `internal/feature/monthlyplan/entity/policy_test.go` — added 5-case policy test
- `internal/feature/monthlyplan/controller/v1.go` — added viewer check in `GetDownloadURL`
- `internal/feature/monthlyplan/controller/v1_test.go` — added controller integration test

### GAP-2: Work Report Module -- FIXED 2026-05-19
**Severity: High** | **Risk: Medium** | **Effort: Large** | **Status: FIXED (read-only endpoint, no schema migration)**

Requirement B specifies `/work-report` (รายงานการปฏิบัติงาน) — a page to view past daily work reports.
- `dailyreportdraft` remains draft/prefill-only; no new daily-report persistence table was added.
- Tasks (`TaskDaily` table) are now exposed through a report view endpoint.
- **Fix applied**: Added `GET /v1/work-reports?month=YYYY-MM&teamId=X` and `GET /v1/work-reports?year=YYYY&month=M&teamId=X`.
- **Contract**: returns `summary.totalTasks`, `summary.uniqueTeamCount`, `summary.teamSummaries`, and mapped TaskDaily report items.
- **Permissions**: viewer is read-only; user/team_lead are scoped to own team; admin/super_admin can read requested/all teams.
- **No schema migration**: source attribution (planning vs monthly plan vs outside plan) still needs schema/product design before adding summary buckets.

**Files changed:**
- `internal/feature/workreport/repository/initiator.go` — read-only adapter to canonical TaskDaily repository
- `internal/feature/workreport/service/service.go` — read-only monthly aggregation over TaskDaily via task repository filter
- `internal/feature/workreport/service/service_test.go` — service validation, scoping, mapping, empty-state, and repository-error coverage
- `internal/feature/workreport/controller/v1.go` — authenticated API controller with query parsing and envelope/error mapping
- `internal/feature/workreport/controller/v1_test.go` — API contract tests for query forms, auth context, errors, and response envelope
- `internal/router/router.go` — registered `GET /v1/work-reports`
- `internal/router/workreport_routes_test.go` — route registration coverage

### GAP-3: Planning Board Endpoint (Unscheduled Tasks)
**Severity: Medium** | **Risk: Medium** | **Effort: Medium**

Requirement B specifies Board tab in `/planning` — shows unscheduled/planned items that haven't been assigned a date yet.
- Currently, team plans always require `start_date` (NOT NULL in schema).
- Tasks always have dates.
- Need a way to represent "unscheduled" or "backlog" items.

**Options:**
1. Allow team plans with a far-future date (e.g., 2999-12-31) as "unscheduled" — query by status='planned' AND start_date > threshold
2. Add a new `backlog` or `unscheduled` flag to team_plans
3. Create a separate backlog entity

**Decision needed:** Product decision on how to represent unscheduled items.

### GAP-4: Capability Model (can_upload_approved_monthly_plan, can_manage_own_team_monthly_plan)
**Severity: Medium** | **Risk: High** | **Effort: Large**

Requirement A specifies granular capabilities:
- `can_manage_own_team_monthly_plan`: team_lead can manage own team's monthly plan
- `can_upload_approved_monthly_plan`: specific users can upload approved/master plans
- These are per-user capabilities granted by super_admin, NOT just roles

**Current state:**
- Monthly plan uses role-based access only
- `CanUploadMasterPlan()` checks role === super_admin
- `CanUploadForTeam()` checks team ownership
- No `user_capabilities` table or service exists

**What's needed:**
- New DB migration: `user_capabilities` table (user_id, capability, granted_by, created_at)
- New service methods to check/grant/revoke capabilities
- Integration into monthly plan upload flow
- Admin UI for managing capabilities (deferred to admin phase)

**Note:** This is a significant schema and service change. Should be done as a dedicated phase.

### GAP-5: Viewer Enforcement Gaps -- ALREADY COVERED (false alarm)
**Severity: N/A** | **Risk: N/A** | **Effort: N/A** | **Status: Already covered**

Re-audit confirmed that `CanUploadForTeam` already blocks viewer via `IsTeamSubmitter()`:
- `IsTeamSubmitter()` returns true only for admin, team_lead, user — NOT viewer
- Existing test case `viewer cannot upload own team plan` already covers this
- `PresignUpload` and `ConfirmUpload` both call `CanUploadForPeriod` → `CanUploadForTeam` → blocked for viewer
- No code changes needed

### GAP-6: Contact Directory (Minor Enhancement)
**Severity: Low** | **Risk: Low** | **Effort: Small**

Current state:
- `GET /v1/contact-directory` — list contacts with team filter, search, pagination ✓
- `GET /v1/contact-directory/:userId` — get single contact ✓
- `PATCH /v1/users/me/contact` — update own contact info ✓
- `PATCH /v1/users/:id/contact` — admin updates any user's contact ✓

**Missing per requirements:**
- Team lead cannot add/edit contacts for their own team members (only admin can via user update)
- This may be acceptable if "contacts" is read-only directory with self-service update only
- **Decision needed:** Does redesign require team_lead to edit team members' contact info?

### GAP-7: Admin Module (DEFERRED)
**Severity: N/A** | **Risk: N/A** | **Effort: N/A**

Requirements say "ค่อยว่ากัน" (will discuss later). Current admin routes:
- User CRUD (super_admin only) ✓
- Master data CRUD (super_admin only) ✓
- Monthly plan settings (super_admin only) ✓
- Dashboard stats ✓

## 3. Role/Permission Matrix Audit

| Role | Planning (Calendar/Board) | Monthly Plan | Daily Report | Work Report | Contacts |
|------|--------------------------|-------------|-------------|-------------|----------|
| super_admin | Read + Manage all | Read + Upload all + Settings | Read + Write all | Read all | Read all + Manage |
| admin | Read + Manage all | Read + Upload all | Read + Write all | Read all | Read all |
| team_lead | Read own team + Create | Read + Upload own team | Read + Write own team | Read own team | Read all + Own contact |
| user | Read own team + Create | Read own team | Read + Write own team | Read own team | Read all + Own contact |
| viewer | **Read only** | **Read only** (NO download!) | **Read only** | **Read only** | Read only |

## 4. Recommended Implementation Priority

### Phase 1: Quick Fixes (complete)
1. **GAP-5**: Re-audit confirmed viewer upload is already blocked by `IsTeamSubmitter()`
2. **GAP-1**: Added viewer download restriction in monthly plan + tests

### Phase 2: Work Report Endpoint (complete)
3. **GAP-2**: Added `GET /v1/work-reports` endpoint as a thin `workreport` module over TaskDaily
   - Supports `month=YYYY-MM` and `year=YYYY&month=M` query contracts
   - Reuses existing task repository filtering; no schema migration

### Phase 3: Planning Board (needs product decision)
4. **GAP-3**: Planning board endpoint — requires deciding how to represent unscheduled items

### Phase 4: Capability Model (significant effort)
5. **GAP-4**: User capabilities table + service + integration into monthly plan flow

## 5. Test Coverage Summary

All existing tests pass (verified 2026-05-19):
- `internal/feature/teamplan/controller` ✓
- `internal/feature/teamplan/entity` ✓
- `internal/feature/teamplan/repository` ✓
- `internal/feature/teamplan/service` ✓
- `internal/feature/monthlyplan/entity` ✓
- `internal/feature/monthlyplan/service` ✓
- `internal/feature/largework/entity` ✓
- `internal/feature/largework/service` ✓
- `internal/feature/dailyreportdraft/service` ✓
- `internal/feature/dailyreportdraft/controller` ✓
- `internal/feature/workreport/service` ✓
- `internal/feature/workreport/controller` ✓
- `internal/feature/dashboard/service` ✓
- `internal/feature/user/bootstrap` ✓
- `internal/feature/user/controller` ✓
- `internal/feature/user/service` ✓
- `internal/router` (RBAC + redesign route tests) ✓
- `internal/middleware` ✓
