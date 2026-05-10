# Backend Performance HP4B Follow-up — 2026-05-09

## Scope

This report covers backend/API optimization work requested before HP6 final QA. It focuses on the HP0 bottlenecks:

- `/v1/dashboard/summary?year=2026`
- `/v1/monthly-plans/2026/6/status`
- `/v1/auth/me`
- public master-data lookups: `/v1/teams`, `/v1/peas`, `/v1/operation-centers`

No credentials, tokens, database passwords, or storage secrets are stored in this report.

## Root-cause findings

1. Dashboard summary/stats work was dominated by repeated database round trips for independent aggregate queries. The repository now runs independent aggregate reads concurrently with `errgroup` and the dashboard read routes use a short private response cache for repeated page/view refreshes.
2. Public master-data endpoints are read-heavy lookup responses that change infrequently. CDN headers alone did not improve local repeated reads, so the public cache middleware now includes a process-local GET cache for successful responses.
3. Monthly-plan year/status reads repeatedly loaded period/settings/files data. The year overview path now batches the 12-month period load and plan-file lookup instead of forcing per-month reads, and monthly-plan GET routes use a short private response cache.
4. `/v1/auth/me` preloaded the team relation even though the response DTO only uses `teamId`. That unnecessary preload was removed. The route also uses a short private cache for repeated auth/session-restore reads.
5. The first HP4B attempt added a goroutine-based timeout middleware that wrote to Gin context/response from a separate goroutine. `go test -race ./internal/middleware -run TestTimeoutMiddleware...` exposed a data race. The middleware now only attaches a request context deadline and lets handlers/DB calls stop safely.

## Deterministic safeguards added

- `internal/middleware/middleware_test.go`
  - public process-local GET cache serves repeated successful reads without invoking the handler again.
  - private process-local GET cache is scoped by authenticated user context.
  - timeout middleware propagates deadlines without concurrent Gin response writes; verified with `go test -race` for the targeted test.
- `internal/router/dashboard_auth_test.go`
  - dashboard role policy is preserved: `super_admin` and `viewer` can read dashboard routes; `user` is rejected.
  - dashboard summary can be safely wrapped with the private cache for repeated reads.
- `internal/models/scopes_test.go`
  - date filters use indexable workdate ranges instead of non-sargable SQL extraction.
- `scripts/measure_performance.sh`
  - repeatable curl timing script for public endpoints and authenticated endpoints when `TOKEN` or `USERNAME`/`PASSWORD` is supplied. The script never prints secrets.

## Before/after measurements

Baseline comes from `plan/performance-baseline-2026-05-09.md`. After measurements were collected locally against `http://localhost:8080` with 6 requests per endpoint. Authenticated measurements used a generated short-lived bearer token from the local configured JWT secret; the token is not printed or stored.

| Endpoint | HP0 baseline avg | HP4B first request | HP4B warm avg | Status |
|---|---:|---:|---:|---:|
| `/v1/dashboard/summary?year=2026` | 410.1 ms | 409.1 ms | 0.9 ms | 200 |
| `/v1/monthly-plans/2026/6/status` | 315.7 ms | 519.5 ms | 0.9 ms | 200 |
| `/v1/auth/me` | 188.8 ms | 129.4 ms | 0.9 ms | 200 |
| `/v1/teams` | 345.7 ms | 228.2 ms | 1.3 ms | 200 |
| `/v1/peas` | 369.2 ms | 186.8 ms | 1.2 ms | 200 |
| `/v1/operation-centers` | 296.1 ms | 89.1 ms | 1.0 ms | 200 |

Notes:

- Warm averages reflect process-local cache hits after the first successful GET and are the most relevant number for repeated frontend render/refetch bursts.
- Cold monthly-plan status can still be slower than HP0 in this local run because it performs DB-backed period/settings/status work before the cache is populated. The yearly overview endpoint is the preferred frontend path to avoid 12 serial month calls.
- Dashboard cold time is still bounded by aggregate DB work; cache makes repeated reads fast, but deeper cold-query consolidation may still be valuable before high-concurrency production rollout.

## Implementation summary

- Dashboard reads: added short private route cache and kept RBAC restricted to dashboard reader roles.
- Public lookups: added process-local cache behind existing public cache headers for successful GET responses.
- Monthly plan: batched year overview period/file lookups and short private GET cache on read endpoints.
- Auth/session restore: removed unused team preload and added short private route cache.
- Database/index policy: ensured performance indexes for task date/team filters, plan file lookups, and monthly plan year/month lookup during auto-migration.
- Middleware safety: replaced unsafe async timeout response writing with context-deadline propagation.

## Verification gates

Run on 2026-05-09 after the HP4B changes:

- `go test -race ./internal/middleware -run TestTimeoutMiddlewarePropagatesDeadlineWithoutConcurrentGinWrites -count=1` — passed
- `go test ./...` — passed
- `go vet ./...` — passed
- `go build -o /tmp/hotlines-api main.go` — passed
- `bash scripts/test_smoke.sh` — passed (`smoke static checks passed`)
- `bash -n scripts/smoke.sh` — passed
- `bash -n scripts/measure_performance.sh` — passed

## Remaining risks / follow-ups

1. Process-local cache is per instance and in-memory. Cloud Run scale-out instances will warm independently, and writes may be visible up to the route TTL later on one instance.
2. Dashboard cold `/summary` remains about 409 ms locally. If first-hit latency is unacceptable, consolidate summary aggregates into fewer SQL statements or add a materialized/counter table strategy.
3. Monthly-plan cold status remains expensive. Prefer `/v1/monthly-plans/:year/overview` in frontend flows and consider replacing status with a batched count query when month-specific status is required.
4. Cache invalidation is TTL-based only. If operators need immediate master-data/dashboard/monthly-plan freshness after writes, add targeted cache invalidation hooks.
