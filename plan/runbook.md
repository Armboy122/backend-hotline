# Backend Hotline Runbook

## Configuration

Primary config is currently loaded through `config.yaml` and the project's config loader. Keep `.env.example` in sync when env-based config is used.

Common settings to verify:

- Server port and Gin mode
- PostgreSQL host/port/user/password/dbname or DSN equivalent
- JWT secret and token lifetime
- CORS allowed origins
- Cloudflare R2 endpoint/account/bucket/access key/secret key
- Upload limits and allowed file types
- Rate limit and timeout values

## First boot

1. Prepare PostgreSQL.
2. Configure `config.yaml` or environment variables for the target runtime.
3. Start the API.
4. Confirm database connection and AutoMigrate/startup logs.
5. Verify `GET /health`.
6. Log in with a known admin user.
7. Check one read-only endpoint such as `GET /v1/tasks?page=1&limit=1`.

## Migration behavior

- This repo currently uses GORM AutoMigrate plus helper commands under `cmd/`.
- Treat schema-fix commands as operational tools, not normal app startup.
- Back up production data before running schema-altering commands.
- Do not rename columns during architecture refactor work.
- If explicit migrations are introduced later, keep them idempotent and ordered.

## Restart procedure

1. Stop the running API.
2. Confirm no one is running manual schema-fix commands.
3. Start the new build with the same config.
4. Re-check `GET /health`.
5. Re-run smoke checks for auth/tasks/dashboard/monthly-plan.

## Rollback

1. Stop the new deployment.
2. Restore previous binary/image.
3. If schema changes caused the issue, restore DB from backup/PITR before retrying.
4. Re-run health and smoke checks.

## R2/upload troubleshooting

- Presign fails: verify R2 endpoint, credentials, bucket, and region/account config.
- Upload succeeds but confirm fails: verify object key format and monthly plan period/team policy.
- Download URL fails: verify file metadata exists and object still exists in R2.
- Hard delete fails: check both DB metadata and R2 object deletion result.

## Auth troubleshooting

- Login fails: confirm user exists, is active, and password hash is valid.
- Token rejected: confirm JWT secret and signing method match runtime.
- Admin route returns 403: confirm role claim and user role in DB.
- `/v1/auth/me` fails after login: confirm middleware stores claims in Gin context as expected.

## Dashboard troubleshooting

- Slow dashboard: check DB indexes on task date/team/job/feeder columns.
- Empty dashboard: verify TaskDaily soft-delete filters and date filters.
- Inconsistent counts: compare `/v1/tasks` filtered result with dashboard query filters.

## Smoke command

`./scripts/smoke.sh` checks health, login/refresh/me, user management access, task list/filter access, monthly-plan period/files/status/settings, and dashboard summary/top/stats endpoints. It prints a pass/fail/skip summary and exits non-zero if any checked endpoint fails.

```bash
BASE_URL=http://localhost:8080 USERNAME=admin PASSWORD=secret TEAM_ID=1 ./scripts/smoke.sh
```

Run static smoke-script regression checks before release:

```bash
bash scripts/test_smoke.sh
```

If `USERNAME` and `PASSWORD` are not provided, authenticated checks are skipped. A release candidate is not fully smoke-tested until authenticated checks pass against a seeded target environment.

## K5 frontend smoke checklist

Use `plan/k5-release-readiness-report.md` for the browser checklist covering login/session restore, daily task creation/listing, monthly plan submission/admin review, and admin user-management role restrictions.
