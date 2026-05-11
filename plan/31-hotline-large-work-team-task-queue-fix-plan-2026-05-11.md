# Hotline Large Work Team Task Queue Fix Plan

> **For Hermes:** Use systematic-debugging + test-driven-development before changing code. This plan is based on Neon production evidence from project `hotlines3` (`bitter-lake-05690037`) on 2026-05-11.

**Goal:** Make `งานระดมทีม` support real per-team work queues: team lead can split a large work plan into point tasks assigned to each team, and each worker/team can load and execute its own queue.

**Architecture:** Keep the existing large-work domain model, but repair the missing production persistence/runtime path and improve UX so each participating team has its own visible task lane. The immediate blocker is schema/runtime mismatch: production has large-work plans and participating teams but lacks the `large_work_tasks` table needed by task APIs.

**Tech Stack:** Go/Gin/GORM backend, PostgreSQL/Neon, Next.js frontend, TanStack Query.

---

## Evidence from Neon

Project: `bitter-lake-05690037` (`hotlines3`)
Database: `neondb`

Production tables found:
- `large_work_items`
- `large_work_item_teams`

Production table missing:
- `large_work_tasks`

No Goose tracking table found:
- `goose_db_version` does not exist

Latest large-work plan:
- id: `1`
- title: `เปลี่ยนลูกถ้วย`
- status: `planned`
- owner team: `2` / `สตูล`
- created by user: `31`
- date: `2026-05-12`

Participating teams already created for plan `1`:
- owner: `สตูล` (`team_id=2`)
- participants: `หาดใหญ่`, `สงขลา`, `สะเดา`, `พัทลุง`, `ผฮล.33`, `ปัตตานี`, `นราธิวาส`, `ยะลา`

Active users currently visible for those teams:
- team 1 `หาดใหญ่`: team lead + user
- team 2 `สตูล`: team lead + user

This means the plan/team assignment layer exists, but task queue persistence is missing.

---

## Root Cause Hypothesis

`คิวงานโหลดไม่ได้` and `จุดงานของแต่ละทีมกำหนดให้ไม่ได้` are caused by a schema/runtime mismatch:

1. Frontend calls task endpoints:
   - `GET /v1/large-work-tasks/my-todos`
   - `POST /v1/large-works/:id/tasks`
   - `GET /v1/large-works/:id/tasks`
2. Backend code expects `large_work_tasks` to exist.
3. Neon production DB does not have `large_work_tasks`.
4. Therefore task APIs cannot persist or load per-team queue items.
5. Existing plan creation still works because it only uses `large_work_items` + `large_work_item_teams`.

Secondary UX issue:
- Current `LargeWorkTasksDialog` is row-based. It can assign tasks to teams, but it does not show a clear “team lane / งานของทีมนี้” grouping yet.
- Requirement wants each team to have its own work set, so the UI should show team-grouped lanes, not just a free-form list.

---

## Proposed Fix Strategy

### Phase 0 — Safety / no direct production change without approval

Do not execute production SQL until user approves.

Before applying DB migration:
- Create/check restore point if available in Neon.
- Apply additive schema only.
- Verify with `information_schema` after migration.

### Phase 1 — Fix production persistence blocker

Add missing `large_work_tasks` table to Neon production using idempotent SQL copied from the committed migration.

SQL must be additive:
- `CREATE TABLE IF NOT EXISTS large_work_tasks (...)`
- `CREATE UNIQUE INDEX IF NOT EXISTS idx_large_work_tasks_plan_sequence ...`
- `CREATE INDEX IF NOT EXISTS idx_large_work_tasks_assigned_team_status ...`

Verification after migration:

```sql
SELECT table_name
FROM information_schema.tables
WHERE table_schema='public'
  AND table_name='large_work_tasks';
```

Expected: one row.

### Phase 2 — Runtime migration policy fix

Current repo has `config.yaml` with:

```yaml
database:
  auto_migrate: false
```

The production DB has no `goose_db_version`, so the SQL migration file alone is not automatically applied.

Choose one stable path:

Option A — Short term:
- Manually apply additive SQL to Neon production after approval.
- Keep code as-is.

Option B — Better long term:
- Add explicit migration command/runbook for deployment.
- Ensure production deploy runs `goose up` or a project migration command before app starts.
- Document in `plan/runbook.md`.

Recommended: do both — manual one-time fix now, then add a migration runner/runbook so this does not recur.

### Phase 3 — Improve team-specific task assignment UX

Modify frontend task assignment dialog so every team has its own lane:

Current component:
- `hotlines3/src/features/large-work/components/LargeWorkTasksDialog.tsx`

Target UX:
- Header: `กำหนดจุดงานให้แต่ละทีม`
- For each selected/participating team:
  - show team name
  - show count: todo / in_progress / done / blocked
  - button: `+ เพิ่มจุดให้ทีมนี้`
  - new row auto-fills `assignedTeamId` for that team
- Existing tasks grouped by `assignedTeamId`
- Save sends a single `tasks` array to existing endpoint.

This avoids users needing to choose team from a dropdown for every row and makes it obvious that each team has its own work.

### Phase 4 — My queue loading / role handling

Ensure `WorkerTodoQueue` handles these states explicitly:
- DB table missing/API 500: show actionable message: `ยังไม่ได้เปิดใช้งานตารางคิวงาน ติดต่อผู้ดูแลระบบ`
- no assigned task: show `ยังไม่มีจุดงานที่มอบหมายให้ทีมของคุณ`
- user has no active team: show `บัญชีนี้ยังไม่ผูกทีม จึงโหลดคิวงานไม่ได้`

Backend should return stable errors if table exists but user has no team/permission.

### Phase 5 — Backfill optional seed task for existing plan

After schema exists, team lead should manually create tasks through UI.

If urgent and user wants a temporary seed for plan `1`, insert sample tasks for each assigned team only after approval. But default plan is no seed SQL — let the team lead define real points.

---

## TDD Task Breakdown

### Task 1: Add backend migration regression test

**Objective:** Ensure the SQL migration includes `large_work_tasks` and required indexes.

**Files:**
- Modify/test: `backend-hotline/pkg/db/migrations/team_plan_migration_test.go` or create `large_work_migration_test.go`

**Expected test:**
- migration contains `CREATE TABLE IF NOT EXISTS large_work_tasks`
- migration contains `idx_large_work_tasks_plan_sequence`
- migration contains `idx_large_work_tasks_assigned_team_status`

### Task 2: Add frontend grouped task-lane helper test

**Objective:** Group tasks/rows by team and auto-create rows assigned to a selected team.

**Files:**
- Create/modify: `hotlines3/src/lib/large-work-helpers.test.ts`
- Modify: `hotlines3/src/lib/large-work-helpers.ts`

**Expected behavior:**
- `buildTeamTaskLanes(teams, tasks)` returns a lane per team.
- `newTaskRowForTeam(teamId)` returns a row with `assignedTeamId` set.

### Task 3: Refactor LargeWorkTasksDialog into team lanes

**Objective:** Let team lead add points directly under each team.

**Files:**
- Modify: `hotlines3/src/features/large-work/components/LargeWorkTasksDialog.tsx`

**Acceptance:**
- each team lane visible
- add point under lane auto assigns team
- existing tasks grouped under correct lane
- save remains compatible with `LargeWorkAddTasksRequest`

### Task 4: Improve WorkerTodoQueue error/no-team states

**Objective:** Make queue load failures explain why instead of generic failure.

**Files:**
- Modify: `hotlines3/src/features/large-work/components/WorkerTodoQueue.tsx`
- Test: `hotlines3/src/features/large-work/worker-todo-flow.test.ts`

**Acceptance:**
- DB/schema error message is actionable
- empty queue message is clear
- no-team message is clear if API exposes it

### Task 5: Add backend API smoke check for production readiness

**Objective:** Prevent “frontend shipped but DB missing table” recurrence.

**Files:**
- Add/update API smoke script or docs.

Checks:
- `GET /v1/large-works/1/tasks` returns 200/empty array, not 500.
- `GET /v1/large-work-tasks/my-todos` returns 200/empty array for a team user, not 500.

### Task 6: Update runbook

**Objective:** Document Neon migration prerequisite.

**Files:**
- Modify: `backend-hotline/plan/runbook.md`
- Modify: `backend-hotline/plan/README.md`

Include:
- project id `bitter-lake-05690037`
- migration check query
- no credentials
- exact verification steps

---

## Verification Commands

Backend:

```bash
cd /Users/sakdithat/Desktop/myproject/backend-hotline
go test ./internal/feature/largework/... ./internal/router/... ./pkg/db/migrations/... -count=1
go build -o /tmp/hotlines-api main.go
git diff --check
```

Frontend:

```bash
cd /Users/sakdithat/Desktop/myproject/hotlines3
npx --yes tsx src/types/large-work.test.ts
npx --yes tsx src/features/large-work/worker-todo-flow.test.ts
npx --yes tsx src/lib/auth/role-policy.test.ts
npx tsc --noEmit
npm run build
git diff --check
```

Neon verification:

```sql
SELECT table_name
FROM information_schema.tables
WHERE table_schema='public'
  AND table_name IN ('large_work_items','large_work_item_teams','large_work_tasks')
ORDER BY table_name;
```

Expected:
- `large_work_item_teams`
- `large_work_items`
- `large_work_tasks`

---

## Decision Needed

Ask user approval before touching production DB:

1. Apply missing `large_work_tasks` schema to Neon production now.
2. Then implement UI lane improvement so each team has its own task section.
3. Then verify by creating tasks under existing plan `1` and loading worker queue.
