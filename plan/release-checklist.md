# Backend Hotline Release Checklist

Use this before promoting a backend build.

## Pre-release

- [ ] `go test ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go build -o /tmp/hotlines-api main.go` passes
- [ ] Docker image builds locally if Dockerfile is part of the release
- [ ] Smoke script passes against the target environment
- [ ] Database migration/AutoMigrate behavior reviewed
- [ ] Secrets and config are set for the target environment
- [ ] `/v1` API compatibility checked against frontend expectations

## Functional checks

- [ ] `GET /health` works
- [ ] Login works
- [ ] `/v1/auth/me` works with token
- [ ] Task list/create/update/delete works
- [ ] Task filters and pagination work
- [ ] Master data list endpoints load
- [ ] Monthly plan settings can be read by admin
- [ ] Monthly plan presign/confirm upload flow works
- [ ] Monthly plan file list/status/download/delete behavior works
- [ ] Dashboard summary/top jobs/top feeders/feeder matrix/stats return data
- [ ] User management admin routes work
- [ ] Non-admin access restrictions still hold

## Security checks

- [ ] Password hashes are not exposed
- [ ] JWT secret is not a default value in production
- [ ] CORS wildcard is not used in production unless explicitly accepted
- [ ] R2 credentials are not logged
- [ ] Admin-only routes require admin role
- [ ] Rate limiting is enabled
- [ ] Timeout middleware is enabled

## Deployment

- [ ] Database backup or point-in-time recovery is available
- [ ] Application starts without migration errors
- [ ] Health endpoint returns OK
- [ ] Logs are visible in the target runtime
- [ ] Rollback binary/image is available
- [ ] Rollback plan is documented

## Post-release

- [ ] Confirm startup logs and API version/build
- [ ] Run one end-to-end task flow
- [ ] Run one monthly plan upload flow
- [ ] Monitor error logs for auth, DB, and R2 failures
- [ ] Confirm dashboard latency is acceptable
