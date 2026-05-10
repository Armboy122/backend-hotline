# Hotline Replan — Performance First, RBAC Split, Yearly Monthly Plan

Date: 2026-05-09

## Source

User feedback after first live browser smoke against local dev:

1. Frontend feels noticeably slow even in test/dev mode.
2. `super_admin` still behaves too much like `admin`; scopes must be separated.
3. Monthly plan should show the whole Buddhist year 2569 / Gregorian 2026, not require fixed month-only operation.
4. Monthly-plan is not removed for teams; the first live operating period starts with the June 2026 plan. Upload/submit actions are role-limited: super_admin/admin can upload/manage broadly; team_lead and user can upload/download only their own team plan while viewing all team rows for awareness.
5. Future month upload rule: June 2026 plan can be uploaded until 23 May 2026, matching `MonthlyPlanSetting.lockDay = 23` already in DB.

## Decisions Answered By User

### D1 — Performance must be checked before other feature work

Decision: Yes. Create a performance baseline/audit lane first and block RBAC/monthly-plan implementation behind it.

Rationale: If frontend is already slow in local test, adding more role/monthly-plan work without measuring can create critical UX issues later.

### D2 — Admin is not a broad system administrator

Decision: `admin` is closer to an operational monthly-plan manager, not a full database/system administrator.

Admin scope:

- Login and use normal user workflows as applicable.
- Manage monthly-plan operational data only.
- Upload/download/replace monthly master plan files for teams.
- View monthly plan status needed for operations.
- Should not manage users, roles, passwords, master data, dashboard/global reports, or unrestricted CRUD unless explicitly approved later.

### D3 — Super admin is the real unrestricted administrator

Decision: `super_admin` can do everything, conceptually equivalent to direct DB-backed full CRUD access through the application.

Super admin scope:

- Exactly one active `super_admin`.
- Full user CRUD and role customization.
- Create/update/deactivate admins/users/team_leads/viewers.
- Reset passwords.
- Read any data.
- Full CRUD over master data and operational data.
- Access all dashboard/report data.
- Upload/manage monthly plans even when locked.
- Bootstrap remains one-time local CLI against real DB env.

### D4 — Monthly-plan starts with June 2026 and upload is admin-controlled

Decision: Do not remove monthly-plan for teams. The live monthly-plan workflow starts with the June 2026 plan, and submissions/uploads are allowed according to lock settings. The upload actor is restricted by role and team: super_admin/admin can upload/manage broadly; team_lead and user can upload/download only their own team plan while all authenticated roles can view team rows for awareness.

New monthly-plan flow:

1. Admin prepares/uploads each month’s master plan into the system, starting with the June 2026 operating cycle.
2. Teams/team_leads/users open/read the monthly plan list in the system.
3. Team_lead/user can see plans across teams for awareness. Team_lead/user can upload/download only their own team’s plan.
4. Admin/super_admin can upload/replace monthly plan files according to the lock policy.
5. No approval workflow.
6. The system should show the whole year 2569 / 2026, with each month’s status.

### D5 — Yearly monthly-plan page

Decision: Monthly plan page should render a yearly view for 2569 / 2026 instead of forcing a single fixed month context.

Expected behavior:

- Show all 12 months for year 2569 / 2026.
- Each month card/row shows whether plan files exist, locked/unlocked state, and download/upload actions according to role.
- Admin/super_admin can upload a future month before its lock deadline.
- Team_lead/user can see all teams for awareness. Team_lead/user can upload/download only their own team’s plan; user upload is no longer read-only under the latest override.

### D6 — Lock rule

Decision: Use DB settings already present.

Current DB setting known from Neon check:

- `MonthlyPlanSetting.reminderStartDay = 20`
- `MonthlyPlanSetting.lockDay = 23`

Interpretation:

- For the plan of target month M, admin/team upload window should close on day `lockDay` of the previous month.
- Example: June 2026 plan upload is allowed until 23 May 2026.
- Admin can still upload/manage monthly plans according to the DB setting. Team_lead and user can upload/download only their own team plan. Super_admin may override/manage any locked period.

## Performance Plan

### Goal

Make the frontend feel fast enough for real operator use before adding more feature changes.

### Baseline measurements required

Workers must capture objective numbers before changing code:

- Next dev page load timing for `/login`, `/`, `/list`, `/monthly-plan`, `/admin/dashboard`.
- Browser console errors/warnings.
- Network waterfall summary: duplicated API calls, slow endpoints, heavy bundles, repeated auth/session fetches.
- React render symptoms: repeated loading states, unnecessary refetch, unstable query keys.
- Backend endpoint latency from smoke/curl for the APIs used by slow pages.
- Compare `npm run dev` behavior with `npm run build && npm start` if feasible, because Turbopack dev can feel slower than production.

### Performance acceptance criteria

- No page should stay on loading/skeleton if direct API requests succeed.
- No obvious duplicate API bursts on first render.
- Monthly-plan yearly view should not make 12 serial month calls if a batched/year API or parallel query strategy is possible.
- Production build must pass: `npm run build`.
- Frontend typecheck must pass: `npx tsc --noEmit`.
- Backend gates must pass after API changes: `go test ./... && go vet ./... && go build -o /tmp/hotlines-api main.go`.

## RBAC Replan

### Old behavior to remove

- Treating `super_admin` and `admin` as the same privileged bucket for most features.
- `canAccessAdminConsole(super_admin/admin)` implying both can use the same management menu.
- Admin ability to manage users/team_leads or broad master data.

### New capability matrix

| Capability | super_admin | admin | team_lead | user | viewer |
|---|---:|---:|---:|---:|---:|
| Login | yes | yes | yes | yes | yes |
| Daily task create/update in own scope | yes | optional/read as configured | yes/user scope if applicable | yes own team | no |
| View own/team monthly plan | yes | yes | yes | yes | yes/read |
| Upload monthly plan | yes | yes | own team only | own team only | no |
| Upload/manage monthly plan after lock | yes/override | yes, per user decision | own team only if settings allow | own team only if settings allow | no |
| Monthly plan settings | yes | yes | no | no | no |
| User CRUD/role customization | yes | no | no | no | no |
| Password reset for others | yes | no | no | no | no |
| Master data CRUD | yes | no by default | no | no | no |
| Dashboard/global report read | yes | no by default | own/team only if needed | own/team only if needed | read-only if explicitly enabled |
| Direct full CRUD over system data | yes | no | no | no | no |

Confirmed by user: admin can edit monthly-plan settings and can still upload/manage monthly plans; team_lead and user can upload/download only their own team plan while viewing every team row for awareness.

## Monthly Plan Replan

### Backend target

Prefer adding a year-level API to avoid 12 serial calls from the frontend:

- `GET /v1/monthly-plans/:year/overview`
  - returns 12 months with period, lock state, files, counts/status, and available actions for current actor.
- Existing month endpoints can remain for detail/download/upload.
- Ensure role policy:
  - `super_admin`: all.
  - `admin`: monthly plan upload/download/read and monthly-plan setting edits, not user/master-data admin.
  - `team_lead`/`user`: see all teams for planning awareness, upload/download own-team files only.
  - `viewer`: read-only monthly-plan awareness if enabled; no upload.

### Frontend target

- `/monthly-plan` shows year selector/default 2569 / 2026.
- Render all 12 months as card/table/accordion.
- Thai Buddhist year labels should be clear: `2569 (2026)` or Thai UI wording.
- No hardcoded Jan/Feb/Mar-only view.
- No fixed current-month-only UX.
- Role-aware buttons:
  - Admin/super_admin: upload/manage for allowed periods.
  - Team_lead/user: see all teams, upload/download only own-team files.
  - Viewer: view/read only; no upload.

## TDD Requirements For Workers

- Any RBAC change must start with backend policy tests and frontend role-policy tests that fail first.
- Monthly yearly view must start with frontend view-model/type tests or component tests where possible.
- Backend year overview must start with controller/service tests.
- Performance fixes must include measurable before/after notes; tests alone are not sufficient.

## Proposed Kanban Graph

1. HP0 — Performance baseline and bottleneck report.
2. HP1 — Performance fixes from HP0, with before/after proof.
3. HP2 — Backend RBAC split: super_admin full, admin monthly-plan only.
4. HP3 — Frontend RBAC/navigation split: separate super_admin console from admin monthly-plan tools.
5. HP4 — Backend monthly-plan yearly overview + lock policy.
6. HP5 — Frontend yearly monthly-plan UX for 2569 / 2026.
7. HP6 — QA, smoke, browser verification, PRD/release report update.

Dependency rule: HP2–HP6 must not start implementation until HP0 is done; ideally HP1 is completed before broad UI work.

## Verification Commands

Backend:

```bash
go test ./...
go vet ./...
go build -o /tmp/hotlines-api main.go
bash scripts/test_smoke.sh
BASE_URL=http://localhost:8080 USERNAME=<super_admin_6_digit> PASSWORD=<redacted> TEAM_ID=1 bash scripts/smoke.sh
```

Frontend:

```bash
npx tsc --noEmit
npm run build
```

Browser verification:

- Login as `super_admin`.
- Login as `admin`.
- Confirm admin no longer sees/uses super-admin-only user/role/global CRUD actions.
- Confirm admin can edit monthly-plan settings and upload/manage plans.
- Confirm `/monthly-plan` shows 2569 / 2026 yearly view starting with the June 2026 live cycle.
- Confirm team_lead/user can see every team’s plan rows, upload/download only own-team files, and cannot upload for another team.
- Confirm June 2026 upload is allowed until 23 May 2026 based on lock day 23.


## Backend Performance Follow-up — 2026-05-09

User decision: frontend performance fixes alone are not enough; backend/API bottlenecks from HP0 must be optimized before final QA/release readiness.

Backend performance scope to gate before HP6:

1. Dashboard summary latency
   - Baseline: `/v1/dashboard/summary?year=2026` averaged about 410 ms locally.
   - Investigate SQL/query shape, aggregation strategy, indexes, repeated queries, and safe response caching if freshness allows.

2. Monthly-plan status/yearly overview latency
   - Baseline: `/v1/monthly-plans/2026/6/status` averaged about 316 ms locally.
   - Optimize period/settings/team/file-count lookups and active monthly-plan-file queries.
   - Keep corrected permission rules: super_admin/admin broad manage, team_lead/user own-team upload/download, all authenticated roles can view awareness rows.

3. Auth/session restore latency
   - Baseline: `/v1/auth/me` averaged about 189 ms locally.
   - Profile and reduce avoidable relation loading or repeated DB work.

4. Public master-data latency
   - Baseline examples: `/v1/teams` about 346 ms, `/v1/peas` about 369 ms, `/v1/operation-centers` about 296 ms.
   - Optimize query shape and indexes for lookup endpoints used by forms/admin pages.

Required workflow:
- Follow systematic debugging: measure first, identify root cause, then fix.
- Follow TDD: add failing performance/shape regression tests or benchmark-style guards before changing production code where practical.
- Verify with `go test ./... && go vet ./... && go build -o /tmp/hotlines-api main.go` and smoke script checks.
- Update `plan/performance-baseline-2026-05-09.md` or create a follow-up backend performance report with before/after numbers.
