# Monthly schedule: Hotline to Clinic Tool

Hotline is the source of truth for monthly team schedules. Clinic Tool must read
only a published revision through the integration endpoint; drafts never leave
Hotline.

## Deployment checklist

1. Deploy the backend migration that creates schedule periods, revisions, and
   assignments, and adds monthly-plan metadata to teams.
2. Set `CLINIC_TOOL_INTEGRATION_KEY` to a long random secret in the Hotline
   backend environment. Store the same value only in Clinic Tool's server-side
   environment.
3. In Hotline team management, configure every visible team with a unique stable
   code, base area, crew type, and display order.
4. Optionally validate and import the reviewed July 2026 fixture:

   ```sh
   go run ./cmd/backfill-monthly-schedule \
     --fixture fixtures/monthly-schedule/2026-07-clinic-tool.json
   ```

   The command is dry-run by default. After all team codes match, create the
   draft explicitly:

   ```sh
   go run ./cmd/backfill-monthly-schedule \
     --fixture fixtures/monthly-schedule/2026-07-clinic-tool.json \
     --apply \
     --actor-user-id <super-admin-user-id>
   ```

5. Review the timeline in Hotline, save the draft, and publish it. The import
   command never publishes.

## API contract

Hotline workspace routes require the normal authenticated user session:

- `GET /v1/monthly-plans/{year}/{month}/schedule`
- `PUT /v1/monthly-plans/{year}/{month}/schedule/draft`
- `POST /v1/monthly-plans/{year}/{month}/schedule/publish`

The Clinic Tool server reads:

```http
GET /v1/integrations/clinic-tool/monthly-plans/{year}/{month}
X-Integration-Key: <secret>
X-Request-ID: <optional trace id>
```

The response contains the immutable published revision, checksum, ordered teams,
and a complete set of dated segments. Hotline derives `home` segments from days
not covered by an away assignment. Clinic Tool must not infer or edit schedule
data independently.

The response includes an `ETag` equal to the published checksum. Clinic Tool may
send `If-None-Match`; Hotline returns `304 Not Modified` when the revision has
not changed.

Expected failure modes:

- `401` for a missing or invalid integration key
- `404` when the requested month has no published revision
- `503` when the Hotline integration secret is not configured

## Rollback and recovery

Publishing a new revision supersedes the previous revision without mutating its
stored projection. If a schedule is wrong, correct the Hotline draft and publish
again. Clinic Tool should keep its last successfully fetched revision when a
refresh fails and surface the request ID for investigation.
