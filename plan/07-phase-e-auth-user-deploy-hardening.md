# Phase E - Auth, User, Server, And Deploy Hardening

## Goal

Move security-sensitive auth/user behavior into Hinghoi-style feature folders, then finish runtime/deploy readiness.

## Auth Target

```text
internal/feature/auth/controller/
internal/feature/auth/service/
internal/feature/auth/repository/
internal/feature/auth/dto/
internal/feature/auth/entity/
internal/feature/auth/mapper/
```

Tasks:

- [x] Create `internal/feature/auth` entrypoint packages
- [x] Wire auth routes through `internal/feature/auth`
- [x] Login rejects empty username/password
- [x] Login rejects inactive users
- [x] Password verification uses the existing password/crypto package boundary
- [x] JWT generation stays behind a token/JWT dependency
- [x] `lastLogin` update uses a field-specific update
- [x] Refresh/logout/me behavior remains compatible
- [x] Responses never expose password hashes

## User Target

```text
internal/feature/user/controller/
internal/feature/user/service/
internal/feature/user/repository/
internal/feature/user/dto/
internal/feature/user/entity/
internal/feature/user/mapper/
```

Tasks:

- [x] Create `internal/feature/user` entrypoint packages
- [x] Wire user routes through `internal/feature/user`
- [x] Move user list/get/create/update/delete into feature packages
- [x] Move change password behavior into service
- [x] Preserve admin-only middleware behavior
- [x] Prevent password overwrite during ordinary update

## Server And Runtime Target

Later, after feature migration is stable:

```text
internal/server/hotline_server/
pkg/db/
```

Tasks:

- [x] Move router/server composition toward `internal/server/hotline_server`
- [x] Review whether DB connection/models should move toward `pkg/db`
- [x] Add production config validation
- [x] Review `.env.example`
- [x] Verify Dockerfile/build/runbook/release checklist
- [x] Expand smoke script for auth/task/monthly-plan/dashboard flows

## Acceptance

```bash
go test ./...
go vet ./...
go build -o /tmp/hotlines-api main.go
scripts/smoke.sh
```
