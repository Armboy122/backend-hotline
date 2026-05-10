# Performance Baseline and Bottleneck Report — 2026-05-09

## Scope

This HP0 report establishes the first objective performance baseline for HotlineS3 before HP1 implementation work.

Measured areas:

- Frontend app: `/login`, `/`, `/list`, `/monthly-plan`, `/admin/dashboard`
- Backend API: public master-data endpoints, authenticated dashboard endpoints, monthly-plan endpoints, task list filter endpoint
- Frontend build output: production route response timing and static chunk footprint
- Runtime architecture: auth/session restore, query layer, API proxy, and known monthly-plan permission behavior

Source context:

- `plan/10-hotline-prd-latest-and-kanban-scope.md`
- `plan/12-performance-rbac-monthly-plan-replan.md`
- Frontend repository: `/Users/sakdithat/Desktop/myproject/hotlines3`
- Backend repository: `/Users/sakdithat/Desktop/myproject/backend-hotline`

Sensitive values were intentionally excluded. Any token, password, API key, connection string, or storage credential observed during measurement is treated as `[REDACTED]` and is not copied into this document.

## Executive summary

The frontend feels slow mostly because user-visible screens depend on comparatively slow API responses and heavy client-side route chunks, not because the production HTML shell itself is slow.

Top findings:

1. Backend API latency is the primary HP1 bottleneck for dashboard and monthly-plan experiences.
   - Local `/v1/dashboard/summary?year=2026` averaged about 410 ms.
   - Local `/v1/monthly-plans/2026/6/status` averaged about 316 ms.
   - Public hierarchy endpoints such as `/v1/teams`, `/v1/peas`, and `/v1/operation-centers` ranged from about 296–346 ms locally.

2. Production route HTML is fast for mostly static shells, but data-heavy client pages are still slower.
   - Production `/login`, `/`, and `/monthly-plan` returned HTML in about 3–8 ms average.
   - Production `/list` averaged about 179 ms.
   - Production `/admin/dashboard` averaged about 265 ms.

3. The production client bundle is large.
   - Built JS chunks total about 3.0 MiB across 49 chunks.
   - Largest dependency-linked chunks are associated with `jspdf`, `recharts`, `leaflet`, `antd-mobile`, and repeated `axios`-containing chunks.

4. Development mode adds overhead, but production mode does not eliminate the main bottlenecks.
   - Dev route response times were slower than production for static pages, as expected.
   - Data-heavy pages remain slower in production because they still require client-side hydration and API work.

5. Monthly-plan status has both performance and authorization nuance.
   - `/v1/monthly-plans/2026/6/status` was 200 locally for the generated admin-like fixture, but returned 403 against the frontend-configured API base.
   - This likely reflects environment/auth-role differences rather than raw performance alone.
   - Current product rule correction: monthly-plan is active from June 2026. Teams can view all teams. `team_lead` can upload/download only own-team files. `user` can view/read but cannot upload. Admin/super_admin can upload/manage broadly. Admin can edit settings and upload/manage plans.

## Measurement environment

### Frontend

- Project: `/Users/sakdithat/Desktop/myproject/hotlines3`
- Framework: Next.js 16.1.6 with App Router and Turbopack
- Build command used: `npm run build`
- Production start used: `PORT=3002 npm run start`
- Production server verified listening on port 3002

### Backend

- Project: `/Users/sakdithat/Desktop/myproject/backend-hotline`
- Local API base: `http://localhost:8080`
- Backend appeared to be running locally during measurement.
- A temporary generated JWT fixture was used only for performance tests. Token value is not stored here.

### Frontend data architecture observed

- Frontend uses client-side auth with bearer tokens.
- Access token is in memory; refresh token and stored user are in localStorage.
- API calls go through `src/lib/api-client.ts` using Axios interceptors and automatic refresh-on-401 behavior.
- React Query defaults:
  - staleTime: 5 minutes
  - gcTime: 10 minutes
  - retry: 1
  - refetchOnWindowFocus: false
  - refetchOnReconnect: true
- `/api/[...path]` proxy can route client requests to the Go backend when configured through `NEXT_PUBLIC_API_URL=/api`.

## Baseline results

### Backend public endpoint latency — local backend

Repeated `curl` timing against `http://localhost:8080`:

| Endpoint | Avg latency | Status |
|---|---:|---:|
| `/health` | 2.8 ms | 200 |
| `/v1/teams` | 345.7 ms | 200 |
| `/v1/job-types` | 125.8 ms | 200 |
| `/v1/job-details` | 142.2 ms | 200 |
| `/v1/feeders` | 144.6 ms | 200 |
| `/v1/stations` | 116.8 ms | 200 |
| `/v1/peas` | 369.2 ms | 200 |
| `/v1/operation-centers` | 296.1 ms | 200 |

Interpretation:

- The process and network path are healthy because `/health` is extremely fast.
- Master-data endpoints are likely spending time in database work, relation loading, serialization, or missing indexes/query shape rather than connection setup alone.
- These endpoints are important because many forms and admin pages depend on them.

### Backend authenticated endpoint latency — local backend

Repeated `curl` timing against local backend with generated admin-like fixture:

| Endpoint | Avg latency | Status |
|---|---:|---:|
| `/v1/auth/me` | 188.8 ms | 200 |
| `/v1/dashboard/summary?year=2026` | 410.1 ms | 200 |
| `/v1/dashboard/top-jobs?year=2026&limit=10` | 133.3 ms | 200 |
| `/v1/dashboard/top-feeders?year=2026&limit=10` | 130.5 ms | 200 |
| `/v1/monthly-plans/2026/6` | 70.1 ms | 200 |
| `/v1/monthly-plans/2026/6/files` | 116.0 ms | 200 |
| `/v1/monthly-plans/2026/6/status` | 315.7 ms | 200 |
| `/v1/monthly-plans/settings` | 49.3 ms | 200 |
| `/v1/tasks/by-filter?year=2026&month=5&teamId=all` | 68.9 ms | 200 |

Interpretation:

- Dashboard summary is the slowest measured authenticated endpoint.
- Monthly-plan status is the next most important backend bottleneck.
- `/v1/auth/me` is unexpectedly high for a session-check endpoint and can directly affect perceived startup/auth-guard delay.

### Backend authenticated endpoint latency — frontend-configured API base

The same endpoint set was measured against the API base configured for the frontend environment. The actual base URL is omitted here to avoid copying environment details.

| Endpoint | Avg latency | Status |
|---|---:|---:|
| `/v1/auth/me` | 283.0 ms | 200 |
| `/v1/dashboard/summary?year=2026` | 369.6 ms | 200 |
| `/v1/dashboard/top-jobs?year=2026&limit=10` | 227.4 ms | 200 |
| `/v1/dashboard/top-feeders?year=2026&limit=10` | 240.3 ms | 200 |
| `/v1/monthly-plans/2026/6` | 159.2 ms | 200 |
| `/v1/monthly-plans/2026/6/files` | 192.4 ms | 200 |
| `/v1/monthly-plans/2026/6/status` | 138.2 ms | 403 |
| `/v1/monthly-plans/settings` | 111.7 ms | 403 |
| `/v1/tasks/by-filter?year=2026&month=5&teamId=all` | 146.1 ms | 200 |

Interpretation:

- The configured API base is usually slower than local backend for comparable successful endpoints.
- Two monthly-plan endpoints returned 403 in this environment. Treat these as auth/role/environment findings, not direct latency wins.
- HP1 should verify role mapping and route guards for monthly-plan pages so team/admin users do not repeatedly call endpoints they cannot access.

### Frontend route response timing — dev vs production

Route HTML response timing comparison:

| Route | Dev avg | Production avg | Production max | Notes |
|---|---:|---:|---:|---|
| `/login` | 26.9 ms | 7.6 ms | 27.9 ms | Static auth shell is fast in prod. |
| `/` | 21.0 ms | 2.8 ms | 4.3 ms | Static shell is very fast in prod. |
| `/list` | 202.3 ms | 178.8 ms | 509.5 ms | Remains slower; likely route size and dynamic/client dependencies. |
| `/monthly-plan` | 26.5 ms | 2.8 ms | 8.5 ms | Static shell is fast; perceived speed depends on client data. |
| `/admin/dashboard` | 301.9 ms | 264.5 ms | 429.3 ms | Remains slower; likely route size and chart/dashboard dependencies. |

Interpretation:

- Dev mode is not the only cause of slow feel.
- `/list` and `/admin/dashboard` have meaningful production response cost even before considering authenticated data fetching.
- The pages most likely to feel slow are those combining route chunk cost, hydration, and API latency.

### Browser navigation snapshot — production `/login`

Browser navigation metrics on `http://localhost:3002/login`:

- navigation duration: about 54.6 ms
- domInteractive: about 24.7 ms
- HTML transfer size: about 3.5 KiB
- decodedBodySize: about 14.7 KiB
- Largest early JS resources on login were around 70 KiB each.

Interpretation:

- The login HTML shell itself is not a bottleneck.
- Initial JS chunk loading is acceptable on local machine, but should still be monitored on lower-end mobile devices.

### Production bundle footprint

Measured `.next/static/chunks` output:

- Total JS chunk size: about 3002.3 KiB across 49 chunks
- Largest CSS chunk: about 149.4 KiB

Largest observed JS chunks:

| Chunk | Approx size | Detected contents |
|---|---:|---|
| `b2842c7ff4a901d2.js` | 567.6 KiB | `jspdf` |
| `980d0adb1740c498.js` | 348.0 KiB | `recharts` |
| `e65fae6ba0a4c496.js` | 219.2 KiB | not fully classified |
| `adabfc2d4bff09a9.js` | 193.4 KiB | not fully classified |
| `6ce80328e988ab20.js` | 165.1 KiB | `leaflet`, `antd-mobile` |
| `dfff2fc9aec5c357.js` | 153.5 KiB | not fully classified |
| `98294bde2ea0a254.js` | 149.1 KiB | `leaflet` |
| multiple smaller chunks | varies | `axios` appears in multiple chunks |

Interpretation:

- PDF export, charts, map/location picker, and mobile picker libraries are heavy enough to affect user-perceived route readiness.
- HP1 should defer or dynamically import heavy feature-specific libraries so they do not tax initial route startup.

## Root-cause classification

### B1 — API latency: dashboard summary

Severity: High

Evidence:

- `/v1/dashboard/summary?year=2026` averaged about 410 ms locally.
- `/admin/dashboard` also has one of the slowest production route responses before data fetching is considered.

Likely root cause area:

- Dashboard summary query shape, aggregation strategy, relation loading, missing database indexes, or repeated separate queries.

HP1 objective:

- Add benchmark or integration timing guard around dashboard summary.
- Inspect SQL generated for the summary endpoint.
- Add/verify indexes for task date/year, job type/detail, feeder/station hierarchy, and team/operation-center filters.
- Consider pre-aggregation or endpoint-level response caching if data does not need second-level freshness.

### B2 — API latency: monthly-plan submission status

Severity: High

Evidence:

- `/v1/monthly-plans/2026/6/status` averaged about 316 ms locally.
- Repository path shows status construction uses period lookup, settings lookup, full team list, and grouped file counts.
- Environment-configured API returned 403 for status/settings, so route calls must be aligned with role permissions.

Likely root cause area:

- Monthly-plan status query shape and repeated full-team status materialization.
- Missing compound indexes on plan files by monthly plan, team, and delete flag.
- Frontend may call admin-only status/settings paths when current role should not.

HP1 objective:

- Add indexes for active monthly-plan file lookups.
- Verify status endpoint does only the minimum queries needed.
- Ensure frontend monthly-plan pages call role-appropriate endpoints.
- Add tests covering corrected monthly-plan permissions:
  - monthly-plan is active from June 2026
  - team/team_lead can view all teams; team_lead can upload/download only own-team files
  - user can view/read but cannot upload
  - upload/manage/settings remain admin/super_admin broadly, with admin allowed to edit monthly-plan settings

### B3 — API latency: auth/session restore

Severity: Medium

Evidence:

- `/v1/auth/me` averaged about 189 ms locally and about 283 ms through the frontend-configured API base.
- Client-side auth guard depends on restoring the session before protected pages settle.

Likely root cause area:

- User lookup and relation loading during session restore.
- Network path to configured API base.
- Refresh flow may compound perceived wait if access token is absent and only refresh token is available.

HP1 objective:

- Profile `/v1/auth/me` and refresh endpoint separately.
- Return only the fields needed for auth guard/session display.
- Avoid extra relation preload unless required.
- Add frontend instrumentation for time-to-auth-ready.

### B4 — API latency: public master-data endpoints

Severity: Medium

Evidence:

- `/v1/teams`: about 346 ms locally.
- `/v1/peas`: about 369 ms locally.
- `/v1/operation-centers`: about 296 ms locally.

Likely root cause area:

- Broad list endpoints, relation counts/preloads, serialization, or missing indexes.
- These endpoints can accumulate if multiple selectors load in parallel on page entry.

HP1 objective:

- Add route-level timing logs or query plan inspection for master-data endpoints.
- Cache rarely changing master data in React Query and possibly backend memory/cache with invalidation on admin mutations.
- Confirm frontend avoids duplicate queries under auth/layout/page boundaries.

### B5 — Bundle size and heavy client dependencies

Severity: Medium

Evidence:

- About 3.0 MiB total JS chunks in production build.
- `jspdf`, `recharts`, `leaflet`, and `antd-mobile` appear in large chunks.
- `/list` and `/admin/dashboard` remain slow in production compared with static pages.

Likely root cause area:

- Heavy libraries imported at route/component top level.
- PDF, chart, and map features included before user action.

HP1 objective:

- Dynamically import PDF export code only when the user requests export.
- Dynamically import map/location picker only on pages/steps that need it.
- Ensure chart libraries are isolated to dashboard route chunks.
- Run bundle analyzer after each split and record before/after route chunks.

### B6 — Duplicate requests / React render/refetch behavior

Severity: Medium, not yet proven

Evidence:

- Query defaults are reasonable and refetch-on-focus is disabled.
- Browser-side authenticated request inspection was limited because browser token injection/fetch from the page context failed in the automation environment.
- No confirmed duplicate request trace was captured during this HP0 pass.

Likely root cause area:

- Possible duplicate data loading across layout, auth guard, page components, and feature components.
- Possible role-gated pages invoking admin-only endpoints before role check.

HP1 objective:

- Add temporary browser/network instrumentation or use Playwright to capture route-level request waterfalls after login.
- Count requests for `/`, `/list`, `/monthly-plan`, and `/admin/dashboard` after first authenticated navigation and after soft navigation.
- Fix only confirmed duplicate calls.

## Prioritized HP1 remediation tasks

### HP1-1 — Profile and optimize dashboard summary endpoint

Acceptance criteria:

- Add reproducible benchmark/integration timing script for `/v1/dashboard/summary?year=YYYY`.
- Capture SQL/query plan evidence before changing code.
- Add or adjust indexes/memoization only with tests or measured proof.
- Target local average under 200 ms for seeded realistic data, or document why not feasible.

### HP1-2 — Optimize monthly-plan status and permission-aware frontend calls

Acceptance criteria:

- Add tests for monthly-plan role behavior matching the latest correction.
- Confirm status endpoint uses minimal queries and proper indexes.
- Confirm non-admin roles do not repeatedly call admin-only settings/status endpoints if not needed.
- Target local monthly-plan status average under 150 ms.

### HP1-3 — Reduce route chunk cost for PDF/chart/map-heavy pages

Acceptance criteria:

- Use dynamic imports for `jspdf` export path.
- Verify chart library remains isolated to dashboard route.
- Lazy-load Leaflet/map code only when the location picker/map is opened or rendered.
- Record bundle before/after with largest chunk sizes and total route JS.

### HP1-4 — Add authenticated browser waterfall baseline

Acceptance criteria:

- Add a safe local script or Playwright smoke that logs in with non-secret test credentials or injected test fixture.
- Capture request count, duplicate endpoint calls, failed 401/403 calls, and time-to-first-useful-data for key routes.
- Store only sanitized timing summaries, never tokens or passwords.

### HP1-5 — Profile auth/session restore path

Acceptance criteria:

- Measure `/v1/auth/me` and refresh endpoint separately.
- Reduce response fields and relation loading if unnecessary.
- Add frontend timing marker for auth-ready state.
- Target `/v1/auth/me` local average under 100 ms.

## Recommended order

1. HP1-4 first if the team needs proof of duplicate requests and client waterfalls before code changes.
2. HP1-1 and HP1-2 in parallel because they target the two slowest measured backend paths.
3. HP1-3 after route-level waterfall confirms which chunks block each page.
4. HP1-5 if protected-route startup still feels slow after API and bundle fixes.

## Verification performed

- Production frontend build completed successfully with `npm run build`.
- Production frontend was started on port 3002 and measured via repeated `curl`.
- Browser loaded production `/login` and `/` successfully.
- Backend endpoints were measured with repeated `curl` against local backend and frontend-configured API base.
- Static build chunk sizes were inspected from `.next/static/chunks`.

## Known limitations

- Browser-side authenticated waterfall was not completed because the automation environment could not fetch the local token fixture from page context.
- Some remote/API-base benchmark attempts were blocked by safety tooling; no unsafe workaround was used.
- Measurements are local-machine baselines, not production user monitoring data.
- No source behavior changes were made as part of this report.

## Conclusion

HP0 baseline is complete. The next work should not start with UI rewrites. The highest-confidence fixes are backend profiling/optimization for dashboard summary and monthly-plan status, followed by targeted bundle splitting for PDF/chart/map-heavy routes and authenticated browser waterfall instrumentation to prove or disprove duplicate request hypotheses.
