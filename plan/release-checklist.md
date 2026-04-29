# Backend Hotline Release Checklist

Use this before promoting a backend build.

## Pre-release

- [x] `go test ./...` passes
- [x] `go vet ./...` passes
- [x] `go build -o /tmp/hotlines-api main.go` passes
- [x] Docker image builds locally if Dockerfile is part of the release
- [x] Smoke script passes against the target environment
- [x] Database migration/AutoMigrate behavior reviewed
- [x] Secrets and config are set for the target environment
- [ ] `/v1` API compatibility checked against frontend expectations

## Functional checks

- [x] `GET /health` works
- [x] Login works
- [x] `/v1/auth/me` works with token
- [x] Task list/create/update/delete works
- [x] Task filters and pagination work
- [x] Master data list endpoints load
- [x] Monthly plan settings can be read by admin
- [x] Monthly plan presign/confirm upload flow works
- [x] Monthly plan file list/status/download/delete behavior works
- [x] Dashboard summary/top jobs/top feeders/feeder matrix/stats return data
- [x] User management admin routes work
- [x] Non-admin access restrictions still hold

## Security checks

- [x] Password hashes are not exposed
- [ ] JWT secret is not a default value in production
- [ ] CORS wildcard is not used in production unless explicitly accepted
- [x] R2 credentials are not logged
- [x] Admin-only routes require admin role
- [x] Rate limiting is enabled
- [x] Timeout middleware is enabled

## Deployment

- [ ] Database backup or point-in-time recovery is available
- [x] Application starts without migration errors
- [x] Health endpoint returns OK
- [x] Logs are visible in the target runtime
- [x] Rollback binary/image is available
- [x] Rollback plan is documented

## Post-release

- [ ] Confirm startup logs and API version/build
- [x] Run one end-to-end task flow
- [x] Run one monthly plan upload flow
- [ ] Monitor error logs for auth, DB, and R2 failures
- [ ] Confirm dashboard latency is acceptable
