# Phase E - Auth, User, and Deploy Hardening

> **Structure rule:** All code in this phase MUST follow the module-first vertical-slice layout defined in [`00-structure-reset.md`](./00-structure-reset.md). Each module lives under `internal/modules/<module>/` with `controller.go`, `service.go`, `repository.go`, `repository_impl.go`, `dto.go`, `errors.go`, and `entity.go` as needed. The pilot at `internal/modules/task/` is the reference. Do **not** create new files under `internal/domain/`, `internal/app/`, `internal/port/`, or `internal/adapter/out/` — put auth under `internal/modules/auth/` and user under `internal/modules/user/`.

## Goal

Refactor security-sensitive auth/user behavior after core operational modules are stable, then finish release/deploy readiness.

## E1 - Auth module

**Objective:** Move login/refresh/logout/me behavior out of direct handler/GORM code.

**Files:**

- Create: `internal/modules/auth/controller.go`
- Create: `internal/modules/auth/service.go`
- Create: `internal/modules/auth/repository.go`
- Create: `internal/modules/auth/repository_impl.go`
- Create: `internal/modules/auth/dto.go`
- Create: `internal/modules/auth/errors.go`
- Create: `internal/modules/auth/entity.go`
- Retire or thin: `internal/handlers/v1/auth.go` once module is routed

**Task cards:**

### E1.1 Login

- [ ] Reject empty username/password
- [ ] Reject inactive user
- [ ] Verify bcrypt through password port or existing package seam
- [ ] Generate JWT using `pkg/jwt`
- [ ] Update `lastLogin` with field-specific update, not full `Save()`
- [ ] Response excludes password hash

### E1.2 Refresh/logout

- [ ] Preserve current refresh behavior
- [ ] Preserve logout behavior
- [ ] Invalid token maps to 401

### E1.3 Me

- [ ] Reads actor from auth context
- [ ] Returns current user shape
- [ ] Inactive/missing user behavior is tested

## E2 - User module

**Objective:** Move user management and change password behavior into module services.

**Files:**

- Create: `internal/modules/user/controller.go`
- Create: `internal/modules/user/service.go`
- Create: `internal/modules/user/repository.go`
- Create: `internal/modules/user/repository_impl.go`
- Create: `internal/modules/user/dto.go`
- Create: `internal/modules/user/errors.go`
- Create: `internal/modules/user/entity.go`
- Retire or thin: `internal/handlers/v1/user.go` once module is routed

**Task cards:**

- [ ] List users with pagination if product decides to add it; otherwise document current full-list behavior
- [ ] Get user by id
- [ ] Create user validates username/password/role/team
- [ ] Update user does not overwrite password accidentally
- [ ] Delete/deactivate behavior preserved
- [ ] Change password requires current password unless current behavior says otherwise
- [ ] Admin-only behavior remains enforced at route/middleware level

## E3 - Production config validation

**Objective:** Prevent unsafe runtime defaults in production.

**Files:**

- Modify: `internal/config/config.go`
- Modify: `.env.example`
- Create/modify: `internal/config/config_test.go`

**Task cards:**

- [ ] Require strong JWT secret in production mode
- [ ] Require database config in production mode
- [ ] Require R2 config if upload/monthly-plan routes are enabled
- [ ] Make CORS wildcard invalid in production unless explicitly allowed
- [ ] Document all variables in `.env.example`

## E4 - Build/deploy assets

**Objective:** Make another agent or developer able to build/run consistently.

**Files:**

- Create or update: `Dockerfile`
- Create if desired: `docker-compose.yml`
- Create: `scripts/smoke.sh` if not completed in Phase A
- Modify: `README.md`
- Update: `plan/runbook.md`

**Task cards:**

- [ ] Confirm Dockerfile builds `main.go`
- [ ] Confirm container can receive config/env
- [ ] Confirm health endpoint works locally
- [ ] Smoke script supports `BASE_URL`, `USERNAME`, `PASSWORD`

## E5 - Final architecture cleanup

**Objective:** Remove stale partial patterns and confirm dependency direction.

**Files:**

- Modify only as needed across `internal/`

**Task cards:**

- [ ] No new direct business GORM in migrated controllers
- [ ] Service packages do not import Gin
- [ ] Entity/error packages do not import GORM/Gin/Viper/AWS
- [ ] Repository interfaces return domain entities/results
- [ ] README describes the module-first structure

## Phase E acceptance

```bash
go test ./...
go vet ./...
go build -o /tmp/hotlines-api main.go
scripts/smoke.sh
```
