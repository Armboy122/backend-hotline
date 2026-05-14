# Hotline Current Status — 2026-05-13

## Why this file exists

This is the project-local status note for Hotline. Treat this repo-local plan folder as the current source of truth over the older central Obsidian vault at `/Users/sakdithat/Downloads/files 3`.

Related repos:

- Backend/API/schema owner: `/Users/sakdithat/Desktop/myproject/backend-hotline`
- Frontend/API client: `/Users/sakdithat/Desktop/myproject/hotlines3`

## Current architecture rule

- `backend-hotline` owns database schema and migrations through Goose.
- `hotlines3` is API-only: no Prisma, no frontend ORM, no schema/migration ownership, no direct DB access.
- Frontend data flow is React/service layer → Go REST API → PostgreSQL.

## Current product state

### Large-work / งานระดมทีม

Implemented through drained Kanban workstreams:

- assignment/team task queue fix
- planning board with desktop lanes and fallback form patterns
- lite location planning form
- operations UX and operations UX v2
- worker team todo flow with before/after photo visibility

Current behavior summary:

- Team lead can create/assign large-work task cards to teams.
- Planning supports minimal field-first location input: team, location/map/lat-long, detail, optional before photos.
- Workers see their team-specific todos.
- Operations view shows team progress, task details, map actions, and before/after photo state.
- Backend defaults large-work photo arrays and passes team filter to grouped task lists.

### Admin Panel Lite

Implemented in reduced scope as `จัดการระบบ`:

- No dashboard/analytics/monitoring/SLA work in this phase.
- Frontend includes admin hub, users, teams, and task/jobs management surfaces.
- Uses existing backend APIs where possible.
- Dangerous user/team write actions follow backend role policy (`super_admin` where required).
- Mobile/tablet responsive UX was part of the Kanban acceptance scope.

## Latest commits observed

### `hotlines3` main

- `6704847 feat(admin): add lite system management panel`
- `1853a91 feat: improve large-work operations UX`
- `e0d3f98 feat: make planning the mobile landing flow`
- `2947eb2 feat: improve large work operations UX`
- `3539c5f feat(large-work): simplify location planning form`

### `backend-hotline` main

- `a54e12c fix(tasks): pass team filter to grouped list`
- `34bd3e8 fix: default large work photo arrays`
- `03f3c48 feat(largework): accept lite location task input`

## Kanban state checked

Checked 2026-05-13 19:54 +07. These boards were drained: `todo=0`, `ready=0`, `running=0`, `blocked=0`.

- `hotline-largework-operations-ux`: done 13
- `hotline-largework-operations-ux-v2`: done 11
- `hotline-admin-panel-lite`: done 7
- `hotline-largework-location-lite-form`: done 5
- `hotline-largework-assignment-fix`: done 13

Related cron/watchdogs for these boards are paused after drain.

## Verification run during status refresh

```bash
# backend-hotline
go test ./internal/feature/largework/... ./internal/router/... ./pkg/db/migrations/... ./internal/feature/task/... -count=1

# hotlines3
npx --yes tsx src/features/large-work/worker-todo-flow.test.ts
npx --yes tsx src/features/large-work/operations-view-helpers.test.ts
npx --yes tsx src/components/pages/admin/teams-helpers.test.ts
npx --yes tsx src/lib/auth/role-policy.test.ts
npx tsc --noEmit
npm run lint
```

Result:

- Backend targeted tests passed.
- Frontend targeted tests and TypeScript passed.
- Frontend lint passed with 0 errors and 47 existing warnings.

## Working tree notes

- `backend-hotline` is clean.
- `hotlines3` has untracked QA artifacts only:
  - `.hermes/k6-responsive.spec.ts`
  - `.hermes/k6-storage.json`
  - `.hermes/screens/*.png`
  - `.hermes/tmp-k6-mock-api.mjs`
  - `test-results/.last-run.json`

## Recommended next steps

1. Choose one next lane:
   - Admin Panel Lite real browser/device polish,
   - full backend + frontend smoke test against real API/dev DB,
   - cleanup/archive untracked QA artifacts,
   - or initialize a project-local LLM Wiki for graph-linked knowledge.
2. If using LLM Wiki, initialize it at project scope, not in the old `files 3` vault. Suggested root: `/Users/sakdithat/Desktop/myproject/hotline-wiki` or a repo-local `docs/wiki/` if the user wants it versioned.
3. Keep future schema changes in `backend-hotline/pkg/db/migrations/` only.
