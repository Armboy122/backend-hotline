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

## Large-work production schema gate: `large_work_tasks`

Use the backend's migration path, not ad-hoc SQL execution through an external console/tool. The repo currently owns schema through GORM `AutoMigrate` guarded by `database.auto_migrate`; there is no active Goose runner wired into deploy/startup yet.

Root cause found on 2026-05-11:

- `internal/models/models.go` already defines `LargeWorkTask` with `TableName() == "large_work_tasks"`.
- `pkg/db/db.go` registered `LargeWorkItem` and `LargeWorkItemTeam` in `MigrationModels()` but missed `LargeWorkTask`.
- Therefore even when `auto_migrate: true` is used, startup cannot create `large_work_tasks`.

Correct operational path:

1. Keep `models.LargeWorkTask` registered in `MigrationModels()`.
2. Deploy a backend build containing that registration.
3. Run the backend migration path once with `database.auto_migrate: true` against the target DB, or run the app in a controlled migration job using the same config.
4. Switch `database.auto_migrate` back to `false` for normal production runtime if that is the deployment convention.
5. Post-check production schema:

```sql
SELECT table_name
FROM information_schema.tables
WHERE table_schema = 'public'
  AND table_name IN ('large_work_items', 'large_work_item_teams', 'large_work_tasks')
ORDER BY table_name;

SELECT indexname
FROM pg_indexes
WHERE schemaname = 'public'
  AND tablename = 'large_work_tasks'
ORDER BY indexname;
```

Expected post-check: `large_work_tasks` exists with indexes generated from the GORM model tags, including plan/sequence and assigned-team/status indexes. Then smoke `GET /v1/large-works/1/tasks` and `GET /v1/large-work-tasks/my-todos` with an authenticated team user; both should avoid 500 errors caused by the missing table.

If the project is later standardized on Goose like the reference backend, add a real Goose migration runner/script first, then move this schema ownership there. Do not manually run one-off SQL through MCP/console because that bypasses version/migration ownership and can leave deploy state unclear.

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
