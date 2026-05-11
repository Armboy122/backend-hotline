# HN-LW Execution QA Handoff: งานระดมทีม

Date: 2026-05-11
Scope: final integration QA after backend and frontend งานระดมทีม execution slices
Source of truth: `plan/28-hotline-large-work-execution-replan-2026-05-11.md`

## Product flow verified against diffs

Corrected flow:

1. `team_lead` creates/edits own-team งานระดมทีม plans.
2. `team_lead` assigns point-by-point work to other teams without admin approval.
3. All authenticated roles can view overview/progress.
4. Assigned team members see only own-team execution todos.
5. Workers start a point, save before photo URL, save after photo URL, add completion note, complete the point, then continue to the next point.

## What works

### Backend: `/Users/sakdithat/Desktop/myproject/backend-hotline`

Implemented and verified in the largework slice:

- Policy allows admin/super_admin plus own-team `team_lead` to manage plans and assign execution tasks.
- Users cannot edit plans but can execute tasks assigned to their own team.
- Execution task entity/model/DTO/repository/service/controller coverage exists for:
  - overview and per-team progress
  - replacing/assigning tasks and auto-adding assigned teams as participants
  - own-team todo queue
  - start task
  - attach before/after photo URLs
  - complete task with required before photo, after photo, and completion note
  - block task with reason
- Routes are wired for:
  - `GET /v1/large-work-items/:id/overview`
  - `POST /v1/large-work-items/:id/tasks`
  - `GET /v1/large-work-items/:id/tasks`
  - `GET /v1/large-work-tasks/my-todos`
  - `PATCH /v1/large-work-tasks/:taskId/start`
  - `POST /v1/large-work-tasks/:taskId/photos`
  - `PATCH /v1/large-work-tasks/:taskId/complete`
  - `PATCH /v1/large-work-tasks/:taskId/block`
- New migration/model coverage exists for large-work task storage and photo URL arrays.

### Frontend: `/Users/sakdithat/Desktop/myproject/hotlines3`

Implemented and verified in the planning/large-work slice:

- Role policy allows `team_lead` create/edit/assign for owner team and own-team users to execute assigned tasks.
- Service/hooks expose overview, task assignment, my-todos, start, photo, complete, and block routes.
- Planning UI includes overview progress and assignment card/row entry for task points.
- Worker todo UI shows own-team queue and supports start -> before URL -> after URL -> completion note -> complete -> next point.
- QA fix applied during this handoff: `canCompleteWorkerTask` now requires a non-empty completion note before enabling completion, matching backend validation.

## Verification performed

### Backend gates

From `/Users/sakdithat/Desktop/myproject/backend-hotline`:

- PASS: `go test ./internal/feature/largework/... ./internal/router/... ./pkg/db/migrations/... -count=1`
- BLOCKED by out-of-scope filesystem/resource issue: `go test ./... -count=1`
  - Failing setup packages are under `internal/feature/masterdata/*` with `resource deadlock avoided` while reading files.
  - `internal/feature/largework/*`, `internal/router`, `internal/models`, and `pkg/db/migrations` passed within the full run.
- BLOCKED by same out-of-scope filesystem/resource issue: `go vet ./...`
  - Failure reads `internal/feature/masterdata/*` files with `resource deadlock avoided`.
- PASS: `go build -o /tmp/hotlines-api main.go`
- PASS: `git diff --check`
- PASS: static added-line security scan found no hardcoded secrets, shell injection, eval/exec, unsafe deserialization, or obvious SQL formatting injection patterns in scoped backend/frontend diffs.

### Frontend gates

From `/Users/sakdithat/Desktop/myproject/hotlines3`:

- PASS: `npx --yes tsx src/lib/auth/role-policy.test.ts`
- PASS: `npx --yes tsx src/types/large-work.test.ts`
- PASS: `npx --yes tsx src/features/large-work/worker-todo-flow.test.ts`
- PASS: `npx tsc --noEmit`
- PASS: `npm run build`
  - Note: Next logs `.env` read `Unknown system error -11`, but build exits 0 and completes successfully.
- PASS: scoped `git -c core.preloadIndex=false diff --check -- ...large-work/planning paths...`
- BLOCKED for full-repo git operations by pre-existing `CLAUDE.md` resource-deadlock behavior; use scoped pathspecs until fixed.

### Independent review

- PASS: independent reviewer approved the QA fix requiring completion note before task completion.
- Non-blocking suggestions: add more helper assertions for whitespace-only notes and all existing/draft photo combinations if this area receives another test-hardening pass.

## Risks and known blockers

1. Full backend `go test ./...` and `go vet ./...` are not clean because newly added/out-of-scope `internal/feature/masterdata/*` files hit macOS `resource deadlock avoided` read failures. The largework-scoped backend gate is clean.
2. Frontend full git status/diff can fail on `CLAUDE.md` with `Resource deadlock avoided`; scoped pathspec checks are clean.
3. Frontend production build logs `.env` read error `Unknown system error -11`, but exits 0. Treat as environment/filesystem noise unless runtime env values are missing in manual QA.
4. MVP uses photo URL arrays instead of real file upload. This matches plan/28’s allowed storage seam, but a later task should wire the existing upload/R2 flow if field users need native photo capture/upload.
5. Assignment UI helper currently drops rows without `assignedTeamId` when building payload. Backend rejects invalid task rows, but UX should eventually surface row-level validation before submit.

## Manual QA script

Use a fresh local/dev database with migrations applied.

1. Start backend.
   - `cd /Users/sakdithat/Desktop/myproject/backend-hotline`
   - `go build -o /tmp/hotlines-api main.go`
   - run app with the normal project environment.
2. Start frontend.
   - `cd /Users/sakdithat/Desktop/myproject/hotlines3`
   - `npm run dev`
3. Login as a `team_lead` for team A.
4. Open `/planning` and select `งานระดมทีม`.
5. Create a new large-work plan owned by team A.
   - Expected: create action is visible for `team_lead`.
   - Expected: plan saves and appears in overview.
6. Add at least three task points:
   - one assigned to team A
   - two assigned to team B
   - include point label, work type/detail, quantity fields, and location text.
   - Expected: save succeeds and overview counts include both teams.
7. Login as a user/team member from team B.
8. Open `/planning`.
   - Expected: overall plan overview is visible.
   - Expected: plan edit/assignment controls are not visible.
9. Open the worker todo queue.
   - Expected: only team B assigned tasks appear.
   - Expected: team A task does not appear in my-todos.
10. Select the first team B task and press start.
    - Expected: status becomes `in_progress`.
11. Add a before photo URL and save it.
    - Expected: before count increments.
12. Add an after photo URL and save it.
    - Expected: after count increments.
13. Try to complete without a completion note.
    - Expected: complete button remains disabled.
14. Add a completion note and complete.
    - Expected: task becomes `done`, completed timestamp/user are set, UI advances to the next incomplete point.
15. Return to overview.
    - Expected: done/progress counts update.
16. Block another task with a reason via API or UI when available.
    - Expected: task status becomes `blocked`, reason is retained in notes, and blocked count updates.

## Release recommendation

Largework-scoped implementation is ready for human review. Do not treat the whole repo as fully green until the out-of-scope resource-deadlock blockers for backend `masterdata/*` and frontend `CLAUDE.md` are resolved or explicitly waived.
