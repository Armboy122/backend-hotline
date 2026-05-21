# B2 Role and Capability Enforcement Audit — 2026-05-22

Scope: backend-only audit and stabilization for final Hotline roles and monthly-plan capability enforcement.

## Sources reviewed

- `/Users/sakdithat/Desktop/myproject/hotline/plan/37-world-class-production-grade-benchmark-kanban-2026-05-21.md`
- `/Users/sakdithat/Desktop/myproject/hotline/plan/39-main-route-production-acceptance-checklist-2026-05-22.md`
- `/Users/sakdithat/Desktop/myproject/wiki/entities/projects/hotline.md`
- Backend router and service policy code under `internal/router`, `internal/middleware`, `internal/feature/*`.

## Role policy baseline

Final roles are:

- `super_admin`
- `team_lead`
- `user`
- `viewer`

Legacy `admin` remains only as a migration/rejection value and must not be accepted as a privileged final role.

## Audit result

### Pass

- Final role constants and validation are centralized in `internal/feature/auth/policy/roles.go`; `RoleAdmin` is explicitly legacy and excluded from `IsValidRole`.
- Super-admin-only router helpers return only `policy.RoleSuperAdmin` for admin/master-data/capability/user-management write routes.
- Capability management routes are protected by `RequireAuth` plus `RequireRole(superAdminOnlyRoles()...)`.
- User-management capability grant/revoke routes are inside the same super-admin-only users group.
- Master-data mutations (`teams`, `job-types`, `job-details`, `feeders`, `stations`, `peas`, `operation-centers`) are guarded by `superAdminOnlyRoles()`.
- Generic upload endpoints reject non-`super_admin` roles in the controller.
- Task, work-report, team-plan, large-work, contacts, and monthly-plan services already include role/team-scope tests for common write/read paths.
- Monthly-plan approved/master upload and conversion use backend capability data via `Actor.Capabilities` and `CanUploadMasterPlan`; frontend capability assumptions are not trusted by the service.

### Fixed in this B2 pass

- Gap: `monthlyplan.Service.GetFile` blocked `viewer` for team-scoped files but allowed `viewer` to fetch approved/master files because master files bypassed `CanAccessFile`.
- Fix: `GetFile` now applies `Actor.CanDownloadFile()` before file lookup/access checks, so `viewer`, legacy `admin`, unknown, and unauthorised roles cannot get monthly-plan download URLs even for approved/master files.
- Regression: `TestGetFileRejectsViewerDownloadEvenForApprovedMasterPlan` was added and verified RED before the fix.

### Deferred / needs separate design card

- `/v1/dashboard/*` still exists as backend legacy/read API and `dashboardReadRoles()` allows `super_admin` and `viewer`. The production UX says Dashboard must not return as a main/default route, but backend removal/deprecation is larger than B2 and should be coordinated with frontend/QA route policy.
- `/v1/monthly-plans/:year/:month` uses `EnsurePeriod`, which can create a period from a GET path. This predates B2 and may conflict with a strict “viewer causes no writes” interpretation. Recommended follow-up: split read-only period lookup from explicit super-admin/system creation after confirming frontend flow.
- Export/download endpoints outside monthly-plan/work-report are not present or not active in backend scope; viewer no-export remains mostly a frontend/QA route acceptance item unless new export APIs are added.

## Verification

RED:

```bash
go test ./internal/feature/monthlyplan/service -run TestGetFileRejectsViewerDownloadEvenForApprovedMasterPlan -count=1
# failed as expected: viewer got nil error before fix
```

GREEN / scoped:

```bash
go test ./internal/feature/monthlyplan/service -run 'TestGetFileRejectsViewerDownloadEvenForApprovedMasterPlan|TestSoftDeleteFileRequiresMasterPlanCapabilityForApprovedFiles|TestConvertApprovedToPlanningRequiresApprovedMasterFileAndCapability' -count=1
# passed

go test ./internal/feature/monthlyplan/... ./internal/router ./internal/feature/upload/controller ./internal/feature/workreport/... ./internal/feature/task/... -count=1
# passed
```

Full gates:

```bash
go test ./... -count=1
# passed

go vet ./... && go build ./...
# passed
```

## Files changed by this B2 task

- `internal/feature/monthlyplan/service/service_test.go`
- `internal/feature/monthlyplan/service/v1.go`
- `plan/40-b2-role-capability-enforcement-audit-2026-05-22.md`

## Workspace note

Pre-existing unrelated contact-directory/migration changes were present and left untouched:

- `pkg/db/db.go`
- `pkg/db/migration_models_test.go`
- `pkg/db/migrations/20260521112000_create_external_contacts.sql`
- `pkg/db/migrations/external_contacts_migration_test.go`
