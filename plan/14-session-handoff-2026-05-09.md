# Session Handoff — Hotline stabilization and planning

> Latest HNQ release handoff: `plan/18-hnq-final-release-readiness-and-handoff-2026-05-10.md` documents the implemented planning workstream, role matrix, final QA gate references, release risks, and production validation steps.

Date: 2026-05-09 13:11 +07
Status: current stabilization round complete; ready for operator/manual validation.

## Goal of this session

Bring the Hotline project back into a reliable state after the large refactor, with emphasis on:

- Performance first, both frontend and backend.
- Clear RBAC split for `super_admin`, `admin`, `team_lead`, and `user`.
- Monthly plan yearly view for พ.ศ. 2569 / 2026.
- Correct monthly-plan upload/manage/download permissions.
- Updated PRD discovery for upcoming `team plan`, calendar, contact directory, and future `งานระดมทีม`.

## Repository and source-of-truth locations

- Backend / active Obsidian vault: `/Users/sakdithat/Desktop/myproject/backend-hotline`
- Frontend: `/Users/sakdithat/Desktop/myproject/hotlines3`
- Old frontend `hotline-2`: deleted from local workspace; do not use.
- Neon project: `hotlines3`
- Neon project ID: `bitter-lake-05690037`
- Neon branch: `production` / `br-snowy-thunder-a1teoes9`

## Kanban board

Board: `hotline-performance-rbac-2026`

Final status at handoff:

| Task | Status | Summary |
|---|---:|---|
| HP0 performance baseline and bottleneck report | done | Created measurable baseline and identified frontend/backend bottlenecks. |
| HP1 frontend performance fixes from baseline | done | Deferred heavy frontend chunks and added performance guard/reporting. |
| HP2 backend RBAC split | done | Split `super_admin` from `admin`; narrowed `admin` to monthly-plan operations. |
| HP3 frontend RBAC split | done | Updated navigation/actions/guards by role. |
| HP4 backend monthly-plan yearly overview and lock policy | done | Added yearly overview API and lock behavior. |
| HP5 frontend monthly-plan yearly 2569 UX | done | `/monthly-plan` now uses 2569/2026 yearly view and role-aware upload state. |
| HP4B backend API performance bottlenecks | done | Optimized backend/API hot paths and added measurement script/report. |
| HP6 QA smoke PRD and release readiness | done | Ran final gates, cleaned temp artifacts, updated release readiness report. |
| HN0 PRD discovery for team work planning and large-job planning | triage | Discovery-only; not implemented in current stabilization round. |

## Final automated verification

Backend gates passed:

```bash
go test ./...
go vet ./...
go build -o /tmp/hotlines-api main.go
bash scripts/test_smoke.sh
bash -n scripts/smoke.sh
bash -n scripts/measure_performance.sh
git diff --check
```

Frontend gates passed:

```bash
npx --yes tsx src/lib/auth/role-policy.test.ts
npx --yes tsx src/types/monthly-plan.test.ts
npx tsc --noEmit
npm run build
git diff --check
```

## Performance results

Baseline report:

- `plan/performance-baseline-2026-05-09.md`

Backend performance follow-up:

- `plan/performance-backend-hp4b-2026-05-09.md`
- `scripts/measure_performance.sh`

Key warm before/after numbers reported by HP4B:

| Endpoint | Before | After warm |
|---|---:|---:|
| `/v1/dashboard/summary?year=2026` | ~410.1 ms | ~0.9 ms |
| `/v1/auth/me` | ~188.8 ms | ~0.9 ms |
| `/v1/monthly-plans/2026/6/status` | ~315.7 ms | ~0.9 ms |
| `/v1/teams` | ~345.7 ms | ~1.3 ms |
| `/v1/peas` | ~369.2 ms | ~1.2 ms |
| `/v1/operation-centers` | ~296.1 ms | ~1.0 ms |

## Product/RBAC decisions confirmed

### Roles

- `super_admin`: full system admin; broad read/write/manage powers.
- `admin`: monthly-plan operational manager; can upload/manage monthly plans and edit monthly-plan settings, but is not full system admin.
- `team_lead`: can upload/manage monthly plan only for own team where permitted; no admin console.
- `user`: can upload monthly plan only for own team where permitted; view/read otherwise.

### Monthly plan

Monthly plan is primarily a document submission workflow, not an in-system approval workflow.

- Use monthly plan when work is outside the team's own responsibility area.
- User/team submits document and basic metadata such as work date range and place.
- There is no approval state inside the system.
- Admin takes submitted documents to handle approval outside the system.
- After external approval, admin uploads the approved document back into the system.
- Monthly-plan date/place metadata should be reusable for daily work report prefill later.
- Yearly monthly-plan UX focuses on year 2569/2026.
- June 2026 planning uses `MonthlyPlanSetting.lockDay = 23` as the lock rule anchor.

### Team plan / calendar / contacts — discovery only

Discovery note:

- `plan/13-work-planning-and-large-job-prd-discovery.md`

Current requirement understanding:

- Work inside team's own responsibility area = `team plan`.
- Work outside own responsibility area = `monthly plan`.
- `team plan` is for collaborative planning inside a team and does not require approval.
- `user` and `team_lead` can add team-plan items.
- `user` can edit team-plan items they created.
- `team_lead` can delete team-plan items for their own team.
- Calendar UX should resemble Google Calendar monthly view: see which days have work, click a day to see work/location/electric area; exact time is optional.
- Daily report should later sync/prefill from both monthly plan and team plan.
- Contact directory should let users see other users' name, position, phone, and team.
- Every user can edit their own personal/contact information.

### Future large multi-team feature

- Domain term: **งานระดมทีม**
- Meaning: future large multi-team work planning feature.
- Status: explicitly not in current implementation scope.

## Files and docs changed/created during this round

Important planning docs:

- `plan/10-hotline-prd-latest-and-kanban-scope.md`
- `plan/11-k0-decision-matrix.md`
- `plan/12-performance-rbac-monthly-plan-replan.md`
- `plan/13-work-planning-and-large-job-prd-discovery.md`
- `plan/14-session-handoff-2026-05-09.md`
- `plan/k5-release-readiness-report.md`
- `plan/performance-baseline-2026-05-09.md`
- `plan/performance-hp1-frontend-fixes-2026-05-09.md`
- `plan/performance-backend-hp4b-2026-05-09.md`

Important backend/frontend areas touched by workers include:

- Backend RBAC/auth/user/monthly-plan/dashboard/task policy and tests.
- Backend yearly monthly-plan overview and lock policy.
- Backend API performance/caching/query-shape improvements.
- Frontend role-policy helpers/tests.
- Frontend monthly-plan yearly page and role-aware upload state.
- Frontend dashboard performance improvements.
- Smoke and measurement scripts.

Use git diff/status in each repo before continuing:

```bash
cd /Users/sakdithat/Desktop/myproject/backend-hotline && git status --short
cd /Users/sakdithat/Desktop/myproject/hotlines3 && git status --short
```

## Cleanup completed

Final HP6 cleanup removed temporary artifacts/processes:

- Backend `.hermes-tmp/seed_smoke_user.go`
- Frontend `.env.local`
- `/tmp/hotlines_monthly_mock.py`
- Temporary listeners on ports `8080` and `3005`

Telegram completion notification was sent successfully to `telegram:pico Claw` with message id `2228`.
The broken cron watchdog was paused after completion because its bare `telegram` delivery target failed with `Chat not found`.

## Remaining risks / manual validation

Current stabilization scope is ready for operator/manual validation, but still validate manually before production promotion:

1. R2-backed monthly-plan upload/download requires valid Cloudflare R2 credentials and target environment.
2. Browser walkthrough should be run against the actual target backend, especially role behavior for monthly-plan upload/download.
3. Smoke script authenticated paths require real seeded credentials; without credentials they are skipped, not proven.
4. `team plan`, calendar, contact directory, and future `งานระดมทีม` remain discovery-only and need a new implementation phase.

## Recommended next session start

1. Read this note first.
2. Check current working tree status in both repos.
3. If the user wants validation: run manual browser smoke using real target backend/credentials.
4. If the user wants next features: create a new PRD-first board for `team plan + calendar + contact directory`; keep `งานระดมทีม` separate and later.

Suggested next commands:

```bash
cd /Users/sakdithat/Desktop/myproject/backend-hotline
hermes kanban --board hotline-performance-rbac-2026 list
git status --short

cd /Users/sakdithat/Desktop/myproject/hotlines3
git status --short
```
