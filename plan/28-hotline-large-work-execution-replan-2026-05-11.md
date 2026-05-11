# HN-LW Execution Replan: งานระดมทีม Worker Flow

> **For Hermes:** Use Kanban workers with strict TDD. Backend and frontend tasks must not rely on the old MVP assumption that only admin/super_admin can manage งานระดมทีม.

Date: 2026-05-11
Status: approved by product correction in Telegram
Repos:
- Backend: `/Users/sakdithat/Desktop/myproject/backend-hotline`
- Frontend: `/Users/sakdithat/Desktop/myproject/hotlines3`

---

## Goal

Turn `งานระดมทีม` from a simple multi-team planning record into a real field-work execution workflow:

1. `team_lead` can create and edit งานระดมทีม for the owning team.
2. `team_lead` can divide/assign work to other teams immediately, without admin approval.
3. All teams can see the overall plan/progress.
4. Assigned workers see their own todos.
5. Workers complete work point-by-point: before photo -> after photo -> save result -> next point until complete.

---

## Corrected product rules

### Roles

- `super_admin`: full access.
- `admin`: full operational access.
- `team_lead`:
  - create งานระดมทีม for own/own-led area.
  - edit งานระดมทีม they created or own-team owns.
  - assign work points to other teams.
  - see overall plan and team progress.
- `user` / worker:
  - see overall plan.
  - see only own-team assigned todos in the execution queue.
  - upload before/after photos and save completion notes for assigned points.
- `viewer`: read-only overview if retained.

### Planning shape

A large-work plan contains common fields:

- owner team
- title / work type
- work date/range/time
- high-level location text
- total point count / total tree count / total quantity count where applicable
- participating teams
- status and progress summary

The plan has many execution points/tasks. Each point can differ slightly but should share MVP fields:

- sequence/order number
- assigned team
- point label / location name
- latitude
- longitude
- work type / work detail
- quantity fields: point count, tree count, item count
- notes
- status: `todo`, `in_progress`, `done`, `blocked`, `cancelled`
- before photo attachments
- after photo attachments
- completion note
- timestamps and actor IDs for start/complete

Future rounds may need different fields. Keep an extensibility slot such as `metadata` JSON on the point/task entity, but do not block MVP on dynamic-form UI.

---

## Backend implementation direction

### Existing old files to update

- `internal/feature/largework/entity/policy.go`
- `internal/feature/largework/entity/policy_test.go`
- `internal/feature/largework/entity/entity.go`
- `internal/feature/largework/service/v1.go`
- `internal/feature/largework/service/v1_test.go`
- `internal/feature/largework/controller/v1.go`
- `internal/feature/largework/controller/v1_test.go`
- `internal/feature/largework/repository/v1.go`
- `internal/feature/largework/repository/v1_test.go`
- `internal/feature/largework/dto/dto.go`
- `internal/router/router.go`
- `pkg/db/migrations/*large_work*`

### New backend capability

Add execution-point/task support under the existing `largework` feature instead of a separate disconnected feature.

Suggested API:

- `GET /v1/large-work-items/:id/overview`
  - all authenticated roles can view if policy allows overview visibility.
  - returns plan + progress + per-team assignment counts.
- `POST /v1/large-work-items/:id/tasks`
  - admin/super_admin/team_lead owner can add or replace assigned work points.
- `GET /v1/large-work-items/:id/tasks`
  - all involved roles can list; non-privileged users should see overview fields, but action permissions only for own-team assigned tasks.
- `GET /v1/large-work-tasks/my-todos`
  - current user's team queue.
- `PATCH /v1/large-work-tasks/:taskId/start`
  - own assigned team starts work; status -> `in_progress`.
- `POST /v1/large-work-tasks/:taskId/photos`
  - attach before/after photos or record photo URLs if storage layer already exists.
- `PATCH /v1/large-work-tasks/:taskId/complete`
  - own assigned team completes point with after photo and result note.
- `PATCH /v1/large-work-tasks/:taskId/block`
  - own assigned team can block with reason.

If real file upload integration is too large for this slice, implement the DB/API contract as photo URL arrays and leave a clearly named adapter seam for existing storage/R2.

### Backend tests required first

- Policy tests:
  - `team_lead` can create large-work for own team.
  - `team_lead` can update own-team/created large-work.
  - `team_lead` can assign tasks to other teams.
  - `user` cannot edit plan but can update own-team task execution.
- Service tests:
  - create/update validates at least owner + one participant or at least one assigned external team.
  - add tasks validates assigned teams are involved or auto-adds participants by chosen rule.
  - my todos only returns current team tasks.
  - start/complete transitions enforce status order and photo requirements.
- Controller tests:
  - happy path for creating plan as team_lead.
  - adding task points as team_lead.
  - worker my-todos/start/complete endpoints.
- Repository/migration tests:
  - new table/columns exist.
  - list queries support plan overview and team todo queue.

---

## Frontend implementation direction

### Existing old files to update

- `src/types/large-work.ts`
- `src/lib/services/large-work.service.ts`
- `src/hooks/useQueries.ts`
- `src/hooks/mutations/useLargeWorkMutations.ts`
- `src/lib/auth/role-policy.ts`
- `src/lib/auth/role-policy.test.ts`
- `src/app/(main)/planning/page.tsx`

### New UX

Inside `/planning` / `งานระดมทีม`:

1. **Overview tab/card**
   - all teams see plan list and progress.
   - show total points, total trees/items, done/todo/in-progress/blocked.
   - show per-team assignment counts.
2. **Plan create/edit**
   - `team_lead` has create/edit action.
   - form keeps common plan fields.
3. **Task breakdown / assign work**
   - owner teamlead/admin can add rows for points:
     - assigned team
     - point label/location
     - lat/lng
     - work detail/type
     - point/tree/item counts
     - note
   - mobile-first row/card editing.
4. **My team todos**
   - assigned workers see only own team execution queue.
   - card flow: todo -> start -> before photo -> after photo -> completion note -> done -> next.

### Frontend tests required first

- role-policy tests:
  - team_lead can create/edit/assign large-work.
  - user can update own assigned task execution but cannot edit plan.
- service type tests or small tsx tests:
  - request DTOs include task point fields.
  - my-todos service route exists.
- UI smoke/unit test if current stack supports it; otherwise add TypeScript-level assertion tests following existing project pattern.

---

## Kanban task graph

1. **K0 PRD/API contract patch**
   - update docs and accepted product rules.
   - no production code.
2. **Backend policy + DTO foundation** depends on K0.
3. **Backend execution task model/repository/migration** depends on K0.
4. **Backend service/controller endpoints** depends on 2 and 3.
5. **Frontend policy/types/services** depends on K0.
6. **Frontend planning + assignment UI** depends on 5 and backend contract.
7. **Frontend worker todo execution UI** depends on 5 and backend contract.
8. **Integration QA + release handoff** depends on backend + frontend implementation.

---

## Verification gates

Backend:

```bash
cd /Users/sakdithat/Desktop/myproject/backend-hotline
go test ./internal/feature/largework/... ./internal/router/... ./pkg/db/migrations/...
go test ./...
go vet ./...
go build -o /tmp/hotlines-api main.go
git diff --check
```

Frontend:

```bash
cd /Users/sakdithat/Desktop/myproject/hotlines3
npx --yes tsx src/lib/auth/role-policy.test.ts
npx tsc --noEmit
npm run build
git diff --check
```

Final manual/browser QA:

- team_lead creates a งานระดมทีม plan.
- team_lead assigns multiple points to another team.
- all teams can view overview.
- assigned user sees only own-team todos.
- worker submits before/after photo data and completes point.
- progress updates after each completion.

---

## Non-goals for first implementation wave

- Full dynamic form builder.
- Complex approval workflow.
- Offline sync.
- Sophisticated map drawing.
- New storage provider setup if existing photo upload infra is not ready.

Keep seams for these but ship the execution workflow first.
