# K5 QA, Smoke Tests, Reporting, and Release Readiness

Date: 2026-05-09

## Backend smoke test coverage

Script: `scripts/smoke.sh`
Static regression check: `scripts/test_smoke.sh`

The smoke script checks:

1. Public readiness
   - `GET /health` must return 200.
2. Auth
   - `POST /v1/auth/login` with `USERNAME`/`PASSWORD`.
   - Extracts `accessToken` and `refreshToken` from either wrapped `data.*` or top-level response fields.
   - `GET /v1/auth/me` with bearer token.
   - `POST /v1/auth/refresh` with refresh token.
3. User management
   - `GET /v1/users` with bearer token accepts 200 for admin/super_admin and 403 for non-admin credentials.
4. Daily task
   - `GET /v1/tasks?page=1&limit=1` without token must return 401.
   - `GET /v1/tasks?page=1&limit=1` with token must return 200.
   - `GET /v1/tasks/by-filter?page=1&limit=1&year=$CURRENT_YEAR&month=$CURRENT_MONTH` with token must return 200.
5. Monthly plan
   - `GET /v1/monthly-plans/:year/:month` with token.
   - `GET /v1/monthly-plans/:year/:month/files` with token.
   - `GET /v1/monthly-plans/:year/:month/status` with token.
   - `GET /v1/monthly-plans/settings` accepts 200 for admin/super_admin and 403 for other roles.
6. Dashboard/report
   - `GET /v1/dashboard/summary`.
   - `GET /v1/dashboard/summary?teamId=$TEAM_ID`.
   - `GET /v1/dashboard/top-jobs`.
   - `GET /v1/dashboard/top-feeders`.
   - `GET /v1/dashboard/stats`.

Run command:

```bash
BASE_URL=http://localhost:8080 USERNAME=<login> PASSWORD=<password> TEAM_ID=1 scripts/smoke.sh
```

If `USERNAME`/`PASSWORD` are omitted, authenticated checks are reported as skipped, not passed.

## Frontend smoke checklist

Use this against the Next.js app at `http://localhost:3000` with `NEXT_PUBLIC_API_URL` pointed at the target backend.

### Login and session

- [ ] Visit `/login` while logged out.
- [ ] Login with `super_admin` or `admin` credential.
- [ ] Confirm redirect to `/` and username badge is visible.
- [ ] Confirm admin badge displays `Super Admin` for `super_admin` and `Admin` for `admin`.
- [ ] Refresh browser; session restores without returning to login.
- [ ] Logout; protected routes redirect back to `/login`.

### Daily task flow

- [ ] Visit `/`.
- [ ] Confirm team, job type, job detail, work date, evidence image fields, and location fields render.
- [ ] Submit a normal daily task with required fields.
- [ ] Confirm success toast and no unexpected validation error.
- [ ] Visit `/list` and confirm the new/known task appears.
- [ ] Confirm non-privileged credentials only see own-team task data.

### Monthly plan flow

- [ ] Login as `team_lead` with a team.
- [ ] Visit `/monthly-plan`.
- [ ] Confirm default period is next-month planning context.
- [ ] Open upload dialog and confirm fields exist for work start date, work end date, destination, remarks, and file.
- [ ] Upload or dry-run with a valid file and confirm own-team submission is allowed when period is unlocked.
- [ ] Confirm locked-period UI blocks team_lead upload.
- [ ] Login as `admin` or `super_admin`; confirm admin can upload on behalf of team and see settings/status view.

### Admin user management

- [ ] Login as `super_admin`.
- [ ] Visit `/admin`; confirm dashboard, master data, monthly plan, and user management entries are reachable where present in the UI.
- [ ] Confirm creating/promoting `admin` is available only to `super_admin`.
- [ ] Confirm resetting another user's password is available only to `super_admin`.
- [ ] Login as `admin`; confirm admin cannot create another admin or reset password for others.
- [ ] Login as `team_lead` or `user`; confirm `/admin` redirects or shows access denied.

## Dashboard/report scoping verification

Current verified implementation state after follow-up `t_ae577223`:

- Frontend admin area is guarded by `canAccessAdminConsole`, which allows only `super_admin` and `admin`.
- Backend `/v1/dashboard/*` endpoints now require bearer authentication and are covered by router tests for unauthorized/forbidden/authorized access.
- Frontend dashboard data loading forwards bearer tokens from the client auth flow.
- Smoke script verifies dashboard endpoints with authenticated requests.

## Quality gates run

Backend final gate on 2026-05-09:

- `go test ./...` — passed
- `go vet ./...` — passed
- `go build -o /tmp/hotlines-api main.go` — passed
- `bash scripts/test_smoke.sh` — passed
- `bash -n scripts/smoke.sh` — passed
- `bash -n scripts/measure_performance.sh` — passed
- `git diff --check` — passed

Frontend final gate on 2026-05-09:

- `npx --yes tsx src/lib/auth/role-policy.test.ts` — passed
- `npx --yes tsx src/types/monthly-plan.test.ts` — passed
- `npx tsc --noEmit` — passed
- `npm run build` — passed
- `git diff --check` — passed

Known frontend lint status:

- `npm run lint` still maps to `next lint`, which is not valid for Next 16 in this project. Previous K4 handoff already recorded this as an existing tooling issue.

## HP6 final QA notes — 2026-05-09

- HP5 delivered `/monthly-plan` as a 2569/2026 yearly page backed by the year-overview API and role-aware upload action state.
- HP4B delivered backend/API performance optimization and report: `plan/performance-backend-hp4b-2026-05-09.md`.
- HP4B added repeatable measurement script: `scripts/measure_performance.sh`.
- Final cleanup removed temporary QA files and local mock/server processes:
  - `.hermes-tmp/seed_smoke_user.go`
  - frontend `.env.local`
  - `/tmp/hotlines_monthly_mock.py`
  - temporary listeners on ports `8080` and `3005`
- Browser QA was attempted by workers with local mock/session setup, but final readiness still treats production-like browser walkthrough as manual validation because local proxy/session setup was not stable enough to be authoritative.

## Release readiness verdict

Ready for operator/manual validation of the current performance/RBAC/monthly-plan stabilization scope after successful automated backend and frontend quality gates.

Remaining risks:

1. Smoke script requires real seeded credentials and records authenticated checks as skipped without `USERNAME`/`PASSWORD`.
2. R2-backed monthly-plan upload can only be fully proven in an environment with valid Cloudflare R2 credentials.
3. Final browser walkthrough should still be executed by an operator against the target backend before production promotion, especially monthly-plan upload/download per role.
4. `team plan`, calendar planning, contact directory, and future `งานระดมทีม` remain discovery-only scope in `plan/13-work-planning-and-large-job-prd-discovery.md`.
