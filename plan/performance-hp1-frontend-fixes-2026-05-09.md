# HP1 Frontend Performance Fixes — 2026-05-09

## Scope

Task: `t_0fe48602` — implement highest-impact frontend performance fixes from HP0 before feature work proceeds.

Repositories:

- Frontend: `/Users/sakdithat/Desktop/myproject/hotlines3`
- Backend report location: `/Users/sakdithat/Desktop/myproject/backend-hotline/plan/performance-baseline-2026-05-09.md`

No secrets, passwords, tokens, or connection strings are included in this report.

## HP0 evidence used

HP0 identified that the largest proven bottlenecks are backend/API latency, with frontend route chunk cost as a secondary frontend-owned bottleneck:

- `/v1/dashboard/summary?year=2026`: about 410 ms local average.
- `/v1/monthly-plans/2026/6/status`: about 316 ms local average.
- `/admin/dashboard` route HTML: about 265 ms production average.
- Production JS chunks: about 3.0 MiB total across 49 chunks.
- Heavy libraries observed: `jspdf`, `recharts`, `leaflet`, `antd-mobile`, repeated `axios` chunks.

Because this HP1 card is scoped to frontend fixes, this pass targeted the dashboard `recharts` cost that was loaded by `dashboard-client.tsx` even though the only chart on that page is the feeder matrix chart shown after the user selects a feeder.

## TDD / regression coverage

Added a frontend performance regression check:

- File: `/Users/sakdithat/Desktop/myproject/hotlines3/scripts/performance-regression-check.mjs`
- Package script: `npm run test:performance`

Red verification before implementation:

```text
node scripts/performance-regression-check.mjs
# failed: dashboard-client.tsx must not statically import recharts
# failed: dashboard-client.tsx must dynamically import the feeder matrix chart component
```

Green verification after implementation:

```text
npm run test:performance
# performance regression checks passed
```

## Changes made

Frontend files changed:

- `/Users/sakdithat/Desktop/myproject/hotlines3/src/components/pages/admin/dashboard-client.tsx`
  - Removed static `recharts` import from the dashboard client shell.
  - Added a `next/dynamic` import for the feeder matrix chart.
  - The chart code now loads only when the feeder matrix section renders data.

- `/Users/sakdithat/Desktop/myproject/hotlines3/src/components/pages/admin/feeder-matrix-chart.tsx`
  - New client component containing the `recharts` imports and horizontal feeder matrix chart rendering.

- `/Users/sakdithat/Desktop/myproject/hotlines3/scripts/performance-regression-check.mjs`
  - Guards against accidentally reintroducing static `recharts` imports into the dashboard shell.

- `/Users/sakdithat/Desktop/myproject/hotlines3/package.json`
  - Added `test:performance` script.

## After-build evidence

Production build after the fix:

```text
npm run build
# passed
```

Chunk scan after the fix:

```text
total_js_kib=3016.1
chunk_count=51
567.6 KiB b2842c7ff4a901d2.js jspdf
330.5 KiB f9770e37883dbceb.js recharts
162.3 KiB e17f866064a25912.js leaflet,antd-mobile
149.1 KiB 98294bde2ea0a254.js leaflet
```

Interpretation:

- `recharts` is still present because the dashboard still needs the feeder matrix chart.
- The route shell no longer statically imports `recharts`; the route loadable manifest now lists the chart chunk as a dynamic loadable asset:

```text
.next/server/app/(main)/admin/dashboard/page/react-loadable-manifest.json
19099.files = ["static/chunks/0c74f0684a29692e.js", "static/chunks/f9770e37883dbceb.js"]
```

Browser smoke without authenticated session:

- `/login` loaded successfully in browser.
- Navigating to `/admin/dashboard` redirected to `/login`, as expected for unauthenticated access.
- Browser resource inspection after unauthenticated `/admin/dashboard` navigation did not load the `recharts` chunk.

Production route timing smoke on local server, after build:

```text
/login statuses=[200, 200, 200, 200, 200] avg_ms=8.9 max_ms=35.3
/monthly-plan statuses=[200, 200, 200, 200, 200] avg_ms=6.7 max_ms=19.6
/admin/dashboard warm_avg_ms=259.6 min_ms=218.4 max_ms=359.3
```

The `/admin/dashboard` HTML timing remains close to HP0 because this fix targets client-side route chunk readiness, not backend dashboard API latency or server-rendered route cost.

## Verification commands

Frontend:

```text
npm run test:performance  # passed
npx tsc --noEmit          # passed
npm run build             # passed
```

Backend quality gate, unchanged backend code:

```text
go test ./... && go vet ./... && go build -o /tmp/hotlines-api main.go  # passed
```

Shell smoke scripts from the task body were not present in the frontend repo:

```text
missing:scripts/test_smoke.sh
missing:scripts/smoke.sh
```

## Remaining performance work

The highest measured HP0 bottlenecks remain backend-owned and should be handled in follow-up backend/API work:

1. Optimize dashboard summary endpoint query shape or add measured caching/index changes.
2. Optimize monthly-plan status endpoint and confirm role-specific frontend/API contract.
3. Profile `/v1/auth/me` and refresh separately.
4. Add authenticated browser waterfall automation with a safe local fixture to measure duplicate calls and time-to-first-useful-data.

This HP1 frontend pass intentionally avoided cosmetic loading changes and did not hide API latency behind UI-only skeleton work.
