# Execution Backlog

## Dependency graph

```text
A. Test + architecture hardening                     [TODO]
  -> B. TaskDaily vertical slice                      [TODO]
       -> C. Monthly plan workflow                    [TODO]
            -> D. Dashboard + master data pattern     [TODO]
                 -> E. Auth/user + deploy hardening   [TODO]
```

## Milestones

### M1 - Safe foundation [TODO]

- [ ] A1 architecture boundary tests
- [ ] A2 config and middleware tests
- [ ] A3 standard response/error mapping tests
- [ ] A4 TaskDaily list behavior tests around current usecase
- [ ] A5 smoke script skeleton for core `/v1` endpoints

Acceptance:

```bash
go test ./...
go vet ./...
go build -o /tmp/hotlines-api main.go
```

### M2 - TaskDaily module complete [TODO]

- [ ] B1 task repository interface covers get/create/update/delete/list-by-team/list-by-filter
- [ ] B2 task usecases added for get/create/update/delete
- [ ] B3 `TaskHandler` delegates all task endpoints to usecases
- [ ] B4 task validation and date/ID parsing locked by tests
- [ ] B5 response compatibility verified for nested team/job/feeder/station data

Acceptance:

- Existing `/v1/tasks*` routes still work.
- Pagination defaults remain `page=1`, `limit=50`, max `100`.
- Soft delete behavior is preserved.
- `go test ./...` passes.

### M3 - Monthly plan workflow modularized [TODO]

- [ ] C1 define monthly plan domain entities and errors
- [ ] C2 define monthly plan repository/storage ports
- [ ] C3 extract settings and period usecases
- [ ] C4 extract presign and confirm upload usecases
- [ ] C5 extract list/status/download/delete/restore/hard-delete usecases
- [ ] C6 keep role/team policy tested

Acceptance:

- Existing `/v1/monthly-plans*` routes remain compatible.
- R2 details are behind a storage port.
- Admin-only behavior remains enforced.
- Non-admin team restrictions are tested.

### M4 - Dashboard and master data cleaned [TODO]

- [ ] D1 dashboard query repository and query service
- [ ] D2 dashboard filter parsing tests
- [ ] D3 extract master data pattern for one representative module
- [ ] D4 apply pattern to remaining master data handlers in small batches
- [ ] D5 update README route documentation from `/api` to `/v1`

Acceptance:

- Dashboard aggregation responses remain compatible.
- Master data CRUD behavior remains compatible.
- Handler direct GORM usage is reduced without broad generic abstractions.

### M5 - Auth/user/release hardening [TODO]

- [ ] E1 auth service/usecase tests for login/refresh/me/logout
- [ ] E2 user service/usecase tests for CRUD/change password
- [ ] E3 replace direct auth/user GORM logic with usecases
- [ ] E4 production config validation and `.env.example` review
- [ ] E5 Docker/build/runbook/release checklist verified
- [ ] E6 smoke script runs a representative auth/task/monthly-plan/dashboard path

Acceptance:

- Password hashes are never exposed.
- Login updates `lastLogin` without `Save()` full-row side effects.
- Admin-only user management remains enforced.
- Release checklist is usable by another agent.

## Agent dispatch rule

Each unchecked checkbox above maps to one task assignment. Do not assign a whole milestone to one agent. Use [`99-agent-task-template.md`](./99-agent-task-template.md).
