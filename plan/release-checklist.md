# Backend Hotline Release Checklist

Use this before promoting a backend build.

## Pre-release

- [x] `go test ./...` passes
- [x] `go vet ./...` passes
- [x] `go build -o /tmp/hotlines-api main.go` passes
- [x] Docker image builds locally if Dockerfile is part of the release
- [x] Smoke script passes against the target environment
- [x] Database migration/AutoMigrate behavior reviewed
- [x] Secrets and config are set for the target environment
- [x] `/v1` API compatibility checked against frontend expectations for auth/user/task/monthly-plan/dashboard K5 paths

## Functional checks

- [x] `GET /health` works
- [x] Login works
- [x] `/v1/auth/me` works with token
- [x] Task list/create/update/delete works
- [x] Task filters and pagination work
- [x] Master data list endpoints load
- [x] Monthly plan settings can be read by admin
- [x] Monthly plan presign/confirm upload flow works
- [x] Monthly plan file list/status/download/delete behavior works
- [x] Dashboard summary/top jobs/top feeders/feeder matrix/stats return data
- [x] User management admin routes work
- [x] Non-admin access restrictions still hold
- [x] Frontend smoke checklist documented for login, daily task, monthly plan, and admin user management
- [x] HNQ planning workstream release handoff documents team plan, planning calendar, contact directory, daily-report prefill, and งานระดมทีม validation scope
- [ ] Execute frontend smoke checklist in browser against target backend
- [x] Confirm frontend large-work action RBAC fix card `t_1bd8c697` is complete before production promotion

## HNQ planning workstream release gate

See `plan/18-hnq-final-release-readiness-and-handoff-2026-05-10.md` for the latest release readiness and handoff report.

- [x] Backend QA parent references final gates for Go tests, vet, build, smoke static checks, and live unauthenticated smoke.
- [x] Frontend QA parent references role policy/type tests, typecheck, build, performance test, browser unauth/login smoke, and `git diff --check`.
- [x] Implemented flows are documented: monthly-plan correction, team plan, planning calendar, contact directory, daily-report prefill, and งานระดมทีม.
- [x] Role matrix is documented for `super_admin`, `admin`, `team_lead`, `user`, and `viewer`.
- [x] Remaining production risks and manual validation steps are documented.
- [ ] Safe credentials available for authenticated backend smoke and browser role walkthrough.
- [ ] R2-backed monthly-plan upload/download validated in target environment.
- [x] `t_1bd8c697` frontend large-work RBAC follow-up completed and verified locally with role-policy tests and TypeScript.

## K5 remaining risks

- Backend dashboard APIs are still public and filter-scoped only; frontend admin guards protect normal UI usage, but direct API role scoping requires follow-up if dashboard data must be confidential at API level. Follow-up card created from K5 for backend/server-side dashboard auth scoping.
- Authenticated smoke coverage requires seeded `USERNAME`/`PASSWORD`; otherwise authenticated checks are skipped by design.
- R2 monthly-plan upload cannot be fully proven without valid target Cloudflare R2 credentials.

## Security checks

- [x] Password hashes are not exposed
- [ ] JWT secret is not a default value in production
- [ ] CORS wildcard is not used in production unless explicitly accepted
- [x] R2 credentials are not logged
- [x] Admin-only routes require admin role
- [x] Rate limiting is enabled
- [x] Timeout middleware is enabled

## Deployment

- [ ] Database backup or point-in-time recovery is available
- [x] Application starts without migration errors
- [x] Health endpoint returns OK
- [x] Logs are visible in the target runtime
- [x] Rollback binary/image is available
- [x] Rollback plan is documented

## Post-release

- [ ] Confirm startup logs and API version/build
- [x] Run one end-to-end task flow
- [x] Run one monthly plan upload flow
- [ ] Monitor error logs for auth, DB, and R2 failures
- [ ] Confirm dashboard latency is acceptable
