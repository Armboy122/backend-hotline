# Hotline Implementation PRD — Team Plan, Monthly Plan, Calendar, Contact Directory, and งานระดมทีม

Date: 2026-05-10
Status: implementation-ready PRD for the next workstream
Source docs: `plan/11-k0-decision-matrix.md`, `plan/12-performance-rbac-monthly-plan-replan.md`, `plan/13-work-planning-and-large-job-prd-discovery.md`, `plan/14-session-handoff-2026-05-09.md`
Backend repo / active vault: `/Users/sakdithat/Desktop/myproject/backend-hotline`
Frontend repo: `/Users/sakdithat/Desktop/myproject/hotlines3`

---

## 1. Goal

Move Hotline from a daily-report-first tool into an operations planning system where each team can see upcoming work, prepare resources, coordinate with other teams, and later reuse planning data when creating daily work reports.

This workstream defines the MVP for:

1. Monthly plan update for outside-area work.
2. Team plan for own-area work.
3. Planning calendar that combines monthly-plan, team-plan, and งานระดมทีม items.
4. Contact directory for team/user coordination.
5. งานระดมทีม as the large multi-team work planning feature.

Daily report sync/prefill is included as a data-contract requirement and light MVP entry point, but deep daily-report comparison/reporting is deferred unless called out by an implementation phase.

---

## 2. Latest rule that supersedes older contradictions

The latest monthly-plan rule is:

- Monthly plan is active and remains part of the product.
- `super_admin` and `admin` can manage monthly plans broadly.
- `team_lead` and `user` can view all team rows for awareness.
- `team_lead` and `user` can upload/download only their own team monthly-plan files.
- `team_lead` and `user` must not upload/download another team's files.
- `viewer`, if retained, is read-only and has no upload action.

This supersedes older text in `plan/12-performance-rbac-monthly-plan-replan.md` and `plan/14-session-handoff-2026-05-09.md` that said normal `user` cannot upload. For this PRD, normal `user` can upload/download own-team monthly-plan files only.

No in-system approval workflow is added for monthly plan. Approval remains an external business process handled by `admin`/`super_admin`; approved documents can be uploaded back into the system.

---

## 3. Domain definitions

### 3.1 monthly plan

Use monthly plan when a team plans work outside its own responsible area.

MVP behavior:

- Primarily document-based.
- Captures enough metadata for calendar and later daily report prefill:
  - work title / short description
  - target team
  - work date or start date
  - optional end date
  - optional time
  - location text
  - PEA / operation center / feeder / station when known
  - notes
  - submitted file(s)
  - approved/returned file(s), if uploaded later by admin/super_admin
- No in-system approval state machine.
- Should appear on the planning calendar.
- Existing yearly view for พ.ศ. 2569 / 2026 remains, but it must also serve as the source for calendar projections.

### 3.2 team plan

Use team plan when work is inside the team's own responsible area.

MVP behavior:

- No approval required.
- `user` and `team_lead` can create own-team plan items.
- Creator can edit their own item while it is not completed/cancelled.
- `team_lead` can edit/delete/cancel own-team items.
- `admin` can read for operational awareness but should not become a broad planner unless explicitly configured later.
- `super_admin` can manage all.
- Time is optional; date/location/work detail are required.
- Should appear on the planning calendar.
- Should be reusable as daily report prefill later.

### 3.3 planning calendar

Combined monthly calendar view for planned work.

MVP behavior:

- Shows work indicators on each date.
- Click/tap a date to see all planned work for that date:
  - team plans
  - monthly plans
  - งานระดมทีม events
- Multi-day plans appear on every date from start date through end date.
- Items without a time still appear on the date.
- Users should see enough information to prepare: work type/title, place, electric area, team(s), and contact path.

### 3.4 contact directory

Directory of active users and teams for coordination.

MVP behavior:

- All logged-in roles can view active directory entries.
- Search/filter by name, team, role/position, and phone where practical.
- Visible fields:
  - display name
  - position/title
  - phone number
  - team
  - role label where useful
- Every user can edit their own personal/contact information.
- `super_admin` can manage all contact/user fields.
- `admin` can manage only fields explicitly allowed for monthly-plan operations; default is read-only for other users' contact info.

### 3.5 งานระดมทีม

Large multi-team work planning feature.

MVP behavior:

- A large work item that involves more than one team.
- Created by `super_admin`, `admin`, or `team_lead` depending on phase policy; initial MVP should allow `super_admin`/`admin`, then add `team_lead` proposal if needed.
- Has an owner team and one or more participating teams.
- Appears on the planning calendar for all participating teams.
- Stores basic coordination data: title, date range, location, work type, owner team, participating teams, notes, optional attachment(s), and status.
- Does not require a complex approval workflow in MVP; status is for coordination only.

---

## 4. MVP user flows

### 4.1 Monthly plan update flow

1. User opens Monthly Plan or Calendar.
2. System shows yearly/monthly rows for all teams for awareness.
3. User selects a target month/team row.
4. Permission rules:
   - `super_admin`: upload, download, replace, delete/restore if implemented, override locks, manage settings.
   - `admin`: upload, download, replace, manage monthly-plan operations broadly according to lock policy; no broad system admin features.
   - `team_lead`: download and upload only if the row belongs to their own team and lock policy allows it.
   - `user`: download and upload only if the row belongs to their own team and lock policy allows it.
   - `viewer`: view rows only; no upload/download unless product later confirms read-only downloads.
5. For upload, actor supplies file and basic metadata: work date/range, location, optional PEA/feeder/station, notes.
6. System stores the active file and keeps enough metadata to project it into the calendar.
7. If the plan requires external approval, admin handles it outside the system and later uploads the approved/returned document.
8. Calendar shows monthly-plan item(s) on planned dates.
9. Later daily report creation can select this monthly-plan item to prefill date/place/team fields.

Acceptance criteria:

- `team_lead` and `user` can upload/download only own-team monthly-plan files.
- `team_lead` and `user` can see all team rows but action buttons are disabled/hidden for non-own teams.
- `admin` and `super_admin` can manage any team row.
- Older `user cannot upload` behavior is removed from tests and UI policy.
- Lock-day behavior remains based on `MonthlyPlanSetting.lockDay` unless overridden by `super_admin`.

### 4.2 Team plan flow

1. User opens Calendar or Team Plan page.
2. User chooses Add Team Plan.
3. System defaults team to actor's own team for `user`/`team_lead`.
4. User enters:
   - title/work type
   - date/start date
   - optional end date
   - optional time
   - location
   - optional PEA/operation center/feeder/station
   - notes
5. System creates a team-plan item with status `planned`.
6. Calendar shows the item on each date in range.
7. Creator can edit their own item.
8. `team_lead` can edit/delete/cancel own-team items.
9. `super_admin` can manage all team-plan items.
10. When creating a daily report, user can select a team-plan item to prefill relevant fields.

Acceptance criteria:

- `user` can create own-team team-plan items.
- `user` cannot create for another team.
- Creator can edit own team-plan item.
- `team_lead` can delete/cancel own-team team-plan item.
- Non-owner `user` cannot edit/delete another user's item unless product later allows team collaboration edits.
- Multi-day team plans render on every date in the range.

### 4.3 Calendar flow

1. User opens Calendar.
2. Default view is current month; year/month selector is available.
3. System loads calendar items in one request for visible date range, not one request per day.
4. Each day shows grouped counts/badges by plan type:
   - team plan
   - monthly plan
   - งานระดมทีม
5. User clicks a day.
6. Day detail drawer/list shows all items with:
   - type
   - title
   - team/teams
   - location/electric area
   - time if present
   - status
   - quick action if actor can edit/upload/download
7. User can filter by plan type and team.
8. User can open item detail from the day drawer.

Acceptance criteria:

- Calendar range API returns all items needed for the visible month.
- Items are permission-filtered for actions but broadly visible for planning awareness unless sensitive status is later added.
- Calendar does not make 30+ serial day calls.
- Empty dates render cleanly.
- Multi-day items appear on every date.

### 4.4 Contact directory flow

1. User opens Contacts.
2. System lists active teams/users.
3. User searches by name/team/position/phone.
4. User views contact card with name, position, phone, team.
5. User edits their own contact profile.
6. `super_admin` can edit any user/contact fields.

Acceptance criteria:

- Directory is available to all authenticated users.
- User can update own phone/position/contact fields.
- Normal user cannot edit another user's contact fields.
- Search works by name and team at minimum.
- Inactive users are hidden by default or clearly marked, depending on existing user model support.

### 4.5 งานระดมทีม flow

1. Actor opens Large Work / งานระดมทีม page or creates from Calendar.
2. Actor creates a large-work item with:
   - title
   - owner team
   - participating teams
   - date/start date
   - optional end date
   - optional time
   - location
   - work type/details
   - notes
   - optional attachments
3. System validates at least two teams are involved.
4. System creates item with status `planned`.
5. Calendar shows item to all participating teams and to global viewers with planning access.
6. Owner actor can edit/cancel item according to role policy.
7. Participating team members can view details and contact owner team.
8. Later phases may add acknowledgement/approval, but MVP does not block on it.

Acceptance criteria:

- Large-work item requires at least owner team + one participating team.
- Item appears on calendar for every participating team.
- Day detail clearly labels item as `งานระดมทีม`.
- Normal team users can view large-work items affecting their team.
- Unauthorized users cannot modify another team's large-work item.

### 4.6 งานระดมทีม MVP contract

This section converts the earlier discovery-only concept into the implementation contract for the MVP. Treat `งานระดมทีม` as a large multi-team planning item, not as a monthly-plan document and not as a daily report.

#### Required fields

- Title / short work name.
- Work type or rough work description.
- Lead team / owner team (`owner_team_id`).
- Participating teams: at least one team in addition to the lead team.
- Planned start date and optional end date. If end date is omitted, use start date as the single work date.
- Optional time. Missing time must not block calendar visibility.
- Location text.
- Optional PEA / operation center / feeder / station references when the existing master data supports them.
- Notes / coordination details.
- Optional attachments only if existing file infrastructure can be reused without adding a new storage workflow.
- Participant assignment/status rows per participating team.
- Source reference for later daily-report prefill; do not create the daily report automatically in MVP.

#### State model

Use a small coordination state machine:

| State | Meaning | Allowed transitions |
|---|---|---|
| `draft` | Item is being prepared and may be incomplete; not shown as confirmed work to participants. | `planned`, `cancelled` |
| `planned` | Item is scheduled and visible on the calendar for lead and participating teams. | `in_progress`, `completed`, `cancelled` |
| `in_progress` | Work has started or is being executed. | `completed`, `cancelled` |
| `completed` | Planned work is done. | none in MVP except super_admin correction if later needed |
| `cancelled` | Work is cancelled and hidden from default active lists but can remain visible in history. | none in MVP except create a new replacement item |

MVP can create directly as `planned` if the form requires all required fields. Keep `draft` only if the UI needs save-before-complete behavior.

#### Participant team rules

- A valid item must include at least two distinct teams total: one lead team plus at least one participating team.
- The lead team must be represented in participant data with role `lead` or `owner` so calendar and permissions can query all affected teams from one join table.
- Participating team rows use role `participant`.
- A team can appear only once per item.
- Removing a participant team is allowed only while the item is `draft` or `planned`; once `in_progress`, use cancel/recreate or record a later follow-up decision.
- Calendar visibility includes every lead/participant team across every date in the planned date range.
- Participant rows should support lightweight status for coordination:
  - `assigned`: team is included but has not responded/started.
  - `acknowledged`: team lead or operations owner has confirmed awareness, if the UI exposes acknowledgement in the same phase.
  - `declined`: deferred for MVP unless user later requests team-response workflow.
  - `done`: deferred for MVP; completion is tracked at large-work item level first.
- MVP minimum participant status is `assigned`; acknowledgement can be added only if it does not create an approval workflow.

#### Role and permission rules

| Capability | super_admin | admin | team_lead | user | viewer |
|---|---:|---:|---:|---:|---:|
| Create งานระดมทีม | yes | yes | no in MVP; proposal later | no | no |
| Create for any lead team | yes | yes | no | no | no |
| Edit item details while `draft`/`planned` | yes | yes | no in MVP | no | no |
| Change date/range/location/participating teams | yes | yes | no in MVP | no | no |
| Cancel item | yes | yes | no in MVP | no | no |
| Mark in progress/completed | yes | yes | proposed later for lead team | no | no |
| View item affecting own team | yes | yes | yes | yes | yes/read-only if viewer remains |
| View all items | yes | yes | optional operations-wide; default yes for admin | awareness only if calendar policy allows | read-only if enabled |

Rules:

- `super_admin` retains full override/manage capability.
- `admin` is allowed here as an operations coordinator for large multi-team work; this does not restore broad user/role/master-data administration.
- `team_lead` and `user` can view items affecting their team through calendar/list/detail.
- `team_lead` creation or proposal is explicitly later-phase scope, not MVP, unless a later card updates this PRD.
- All write APIs must enforce server-side role checks even if UI hides buttons.

#### Create / edit / delete behavior

- Create validates required fields, date range, lead team, distinct participant teams, and actor permission.
- Edit is allowed only for `draft` and `planned` items in MVP; editing `in_progress`/`completed` requires a later change request or super_admin-only correction path.
- Date/range edits must update calendar projection immediately.
- Participant-team edits must preserve distinct-team validation after the edit.
- Delete should be implemented as cancel/soft delete, not hard delete, unless the item was never published (`draft`).
- Cancel requires a reason if the UI can support one cheaply; otherwise keep notes optional and preserve audit timestamps.
- Completed/cancelled items should not appear in default active planning lists unless the user filters for history.

#### Calendar and daily-report integration

- `GET /v1/planning-calendar` must include `large_work` items in the same response shape as team-plan and monthly-plan items.
- A multi-day large-work item appears on every date from start date through end date.
- Calendar item labels must use the exact Thai term `งานระดมทีม` in the UI.
- Day detail shows lead team, participating teams, location, work type/details, status, and available actions.
- Later daily report creation can select a large-work item to prefill date, location, team, work type/title, and source reference.
- Daily-report submission must not mutate the original large-work item in MVP.

#### MVP exclusions

- No approval workflow between participating teams.
- No resource allocation, manpower count, vehicle/equipment scheduling, or shift planning.
- No participant decline/negotiation workflow unless a later product decision adds it.
- No automatic daily report creation.
- No plan-vs-actual analytics for งานระดมทีม.
- No notifications/chat integration.
- No recurring large-work schedule generation.
- No hard-delete of published items.
- No separate calendar projection table unless performance testing proves the composed query is too slow.

---

## 5. Role capability matrix

| Capability | super_admin | admin | team_lead | user | viewer |
|---|---:|---:|---:|---:|---:|
| View calendar | yes | yes | yes | yes | yes/read-only |
| View all team rows for monthly plan awareness | yes | yes | yes | yes | yes/read-only |
| Upload/download monthly plan for any team | yes | yes | no | no | no |
| Upload/download monthly plan for own team | yes | yes | yes | yes | no |
| Override monthly-plan lock | yes | no by default | no | no | no |
| Edit monthly-plan settings | yes | yes | no | no | no |
| Create team plan for own team | yes | optional | yes | yes | no |
| Edit own-created team plan | yes | optional | yes | yes | no |
| Delete/cancel own-team team plan | yes | optional | yes | no | no |
| Manage all team plans | yes | no by default | no | no | no |
| View contact directory | yes | yes | yes | yes | yes/read-only |
| Edit own contact info | yes | yes | yes | yes | no or own only if viewer account is active |
| Edit other users contact info | yes | no by default | no | no | no |
| Create งานระดมทีม | yes | yes in MVP | proposed later | no | no |
| Edit/cancel งานระดมทีม | yes | yes for operations | own-team owned items if later enabled | no | no |
| Manage users/roles/passwords | yes | no | no | no | no |

Implementation note: if existing code has `admin` permissions broader than this matrix, narrow them only in the phase that owns that feature and cover with tests first.

---

## 6. API assumptions

Use existing backend style: `internal/feature/<feature>/{controller,service,repository,dto,entity,mapper}` and TDD before implementation.

### 6.1 Monthly plan APIs

Existing monthly-plan APIs should be extended rather than replaced where practical.

Assumed endpoints:

- `GET /v1/monthly-plans/:year/overview`
  - Returns all months and all team rows with action permissions for current actor.
  - Must include `canUpload`, `canDownload`, `canReplace`, `canOverrideLock`, `canManageSettings` per row/action.
- `POST /v1/monthly-plans/:year/:month/teams/:teamId/files`
  - Upload/presign metadata entry for a team/month.
  - Enforces own-team upload for `team_lead`/`user`.
- `GET /v1/monthly-plans/:year/:month/teams/:teamId/files/:fileId/download`
  - Enforces own-team download for `team_lead`/`user`; broad download for `admin`/`super_admin`.
- `PATCH /v1/monthly-plans/:year/:month/teams/:teamId/metadata`
  - Updates calendar metadata if needed; same write policy as upload.

### 6.2 Team plan APIs

New feature folder: `internal/feature/teamplan`.

Assumed endpoints:

- `POST /v1/team-plans`
- `GET /v1/team-plans?from=YYYY-MM-DD&to=YYYY-MM-DD&teamId=...&status=...`
- `GET /v1/team-plans/:id`
- `PATCH /v1/team-plans/:id`
- `DELETE /v1/team-plans/:id` or `POST /v1/team-plans/:id/cancel`
- `POST /v1/team-plans/:id/create-daily-report-draft` (may be deferred to sync phase)

Policy assumptions:

- Create: `user`/`team_lead` own team only; `super_admin` any team.
- Update: creator or own-team `team_lead`; `super_admin` any.
- Delete/cancel: own-team `team_lead`; `super_admin` any.
- List/read: all authenticated users can view planning items needed for coordination; actions remain scoped.

### 6.3 Calendar APIs

New feature folder can be `internal/feature/planningcalendar`, or a controller that composes monthlyplan/teamplan/largework services without owning their data.

Assumed endpoint:

- `GET /v1/planning-calendar?from=YYYY-MM-DD&to=YYYY-MM-DD&teamId=...&types=monthly_plan,team_plan,large_work`

Response shape assumption:

```json
{
  "items": [
    {
      "id": "string",
      "type": "team_plan | monthly_plan | large_work",
      "title": "string",
      "startDate": "YYYY-MM-DD",
      "endDate": "YYYY-MM-DD|null",
      "time": "HH:mm|null",
      "teamIds": [1],
      "teamNames": ["Team A"],
      "location": "string",
      "electricArea": "string|null",
      "status": "planned",
      "sourceId": "string",
      "actions": {
        "canEdit": true,
        "canDelete": false,
        "canUpload": false,
        "canDownload": true
      }
    }
  ]
}
```

### 6.4 Contact directory APIs

Can extend existing user/team feature folders instead of creating a new one if cleaner.

Assumed endpoints:

- `GET /v1/contact-directory?query=...&teamId=...&role=...`
- `GET /v1/contact-directory/:userId`
- `PATCH /v1/users/me/contact`
- `PATCH /v1/users/:id/contact` guarded by `super_admin` only unless later confirmed.

### 6.5 งานระดมทีม APIs

New feature folder: `internal/feature/largework`.

Assumed endpoints:

- `POST /v1/large-work-items`
- `GET /v1/large-work-items?from=YYYY-MM-DD&to=YYYY-MM-DD&teamId=...&status=...`
- `GET /v1/large-work-items/:id`
- `PATCH /v1/large-work-items/:id`
- `POST /v1/large-work-items/:id/cancel`
- `POST /v1/large-work-items/:id/attachments` optional if attachment infrastructure is ready.

Policy assumptions:

- MVP create/edit: `super_admin` and `admin`.
- Read: teams involved plus global planning viewers.
- Later option: allow `team_lead` to propose large-work items for own team, with no approval workflow unless separately confirmed.

---

## 7. Data-model assumptions

Prefer additive migrations. Avoid destructive changes. Use indexes for calendar date-range queries.

### 7.1 monthly plan metadata

Existing monthly-plan file tables should be extended only if needed.

Fields likely needed on active plan/file or related metadata table:

- `work_title`
- `work_start_date`
- `work_end_date`
- `work_time` nullable
- `location_text`
- `pea_id` nullable
- `operation_center_id` nullable
- `feeder_id` nullable, if model exists
- `station_id` nullable, if model exists
- `notes`
- `approved_file_id` or file purpose/type if approved documents are stored separately

### 7.2 `team_plans`

Assumed table:

- `id`
- `team_id`
- `created_by_user_id`
- `title`
- `work_type`
- `start_date`
- `end_date` nullable
- `work_time` nullable
- `location_text`
- `pea_id` nullable
- `operation_center_id` nullable
- `feeder_id` nullable
- `station_id` nullable
- `notes`
- `status` (`planned`, `cancelled`, `completed`)
- `daily_task_id` nullable for later sync
- `created_at`, `updated_at`, `deleted_at` if soft delete is used

Indexes:

- `(team_id, start_date)`
- `(start_date, end_date)` or a range-friendly equivalent
- `(created_by_user_id)`
- `(status)`

### 7.3 `large_work_items`

Assumed tables:

`large_work_items`:

- `id`
- `owner_team_id`
- `created_by_user_id`
- `title`
- `work_type`
- `start_date`
- `end_date` nullable
- `work_time` nullable
- `location_text`
- `pea_id` nullable
- `operation_center_id` nullable
- `feeder_id` nullable
- `station_id` nullable
- `notes`
- `status` (`draft`, `planned`, `in_progress`, `completed`, `cancelled`)
- `created_at`, `updated_at`, `deleted_at` if soft delete is used

`large_work_item_teams`:

- `large_work_item_id`
- `team_id`
- `role` (`owner`/`lead`, `participant`)
- `participant_status` (`assigned`, `acknowledged`; reserve `declined`/`done` for later if needed)
- `acknowledged_by_user_id` nullable, only if acknowledgement is implemented in MVP UI
- `acknowledged_at` nullable, only if acknowledgement is implemented in MVP UI
- unique `(large_work_item_id, team_id)`

Validation:

- At least two distinct teams per item.
- Owner/lead team must exist in participant set or be represented by role `owner`/`lead`.
- Each team can appear only once per item.
- Participant status defaults to `assigned`; do not add a required approval/decline flow in MVP.

### 7.4 contact fields

Existing `User` and `Team` models should be reused.

Confirm/extend fields:

- user display name/name
- user phone number
- user position/title
- team id/name
- active status

If phone/position already exist, do not duplicate. If not, add nullable fields with safe migration.

### 7.5 calendar projection

Do not create a separate calendar table for MVP unless performance requires it. Prefer query composition from source tables:

- monthly-plan metadata/files
- team_plans
- large_work_items + large_work_item_teams

If performance becomes poor, add a projection/cache in a later phase with invalidation tests.

---

## 8. Frontend assumptions

Expected frontend areas in `/Users/sakdithat/Desktop/myproject/hotlines3`:

- Role policy helpers and tests should be updated first.
- Monthly plan UI should keep the yearly view but correct normal-user own-team upload/download.
- New routes/pages:
  - `/calendar` or `/planning-calendar`
  - `/team-plan` if separate from calendar create flow
  - `/contacts`
  - `/large-work` or Thai-labeled navigation for `งานระดมทีม`
- Calendar page should use a batched date-range API.
- Mobile layout matters: field users may check plans on phones.

Frontend acceptance gates:

- Role policy tests cover `user` own-team monthly-plan upload/download.
- Calendar date-range view-model tests cover multi-day items.
- Typecheck passes: `npx tsc --noEmit`.
- Build passes: `npm run build`.
- Browser smoke confirms buttons are visible/hidden correctly by role.

---

## 9. Implementation phases and checklist

### Phase HNP-01 — PRD contract and policy tests

Goal: lock the corrected role policy before implementation.

Backend checklist:

- [ ] Add/adjust monthly-plan policy tests proving `team_lead` can upload/download own team only.
- [ ] Add/adjust monthly-plan policy tests proving `user` can upload/download own team only.
- [ ] Add/adjust tests proving `team_lead`/`user` can see all monthly-plan rows for awareness.
- [ ] Add tests proving `admin`/`super_admin` broad monthly-plan management still works.
- [ ] Add tests proving `admin` does not regain broad user/role/master-data permissions.

Frontend checklist:

- [ ] Add role-policy tests for monthly-plan own-team upload/download by `team_lead` and `user`.
- [ ] Add tests for non-own-team action suppression.

Acceptance criteria:

- Tests fail before implementation if current behavior still says `user` cannot upload.
- Tests pass after policy correction.
- No feature UI is added before policy tests are in place.

### Phase HNP-02 — Monthly-plan metadata and corrected own-team actions

Goal: correct monthly-plan behavior and ensure calendar-ready metadata exists.

Backend checklist:

- [ ] Extend DTO/entity/mapping only as needed for date/location metadata.
- [ ] Add or update migration using safe nullable fields.
- [ ] Enforce `team_lead`/`user` own-team upload/download.
- [ ] Return per-row action permissions in yearly overview.
- [ ] Preserve admin/super_admin broad operations.

Frontend checklist:

- [ ] Update yearly monthly-plan UI to show upload/download for own-team `user` and `team_lead`.
- [ ] Disable/hide actions for non-own teams with clear copy.
- [ ] Capture/show date/location metadata.

Acceptance criteria:

- `user` and `team_lead` own-team upload/download works.
- Non-own team upload/download is blocked in UI and API.
- Yearly overview still works for 2569/2026.
- Calendar projection has date/location fields available.

Verification:

- Backend: `go test ./... && go vet ./... && go build -o /tmp/hotlines-api main.go`
- Frontend: `npx tsc --noEmit && npm run build`
- Browser smoke by role.

### Phase HNP-03 — Team plan backend

Goal: add own-area planning data model and API.

Backend checklist:

- [ ] Create `internal/feature/teamplan` vertical slice.
- [ ] Add migration for `team_plans`.
- [ ] Write service policy tests first.
- [ ] Implement create/list/detail/update/cancel or delete.
- [ ] Add controller tests for role/team scoping.
- [ ] Add repository tests where practical.

Acceptance criteria:

- `user` creates own-team team plan.
- `user` cannot create for another team.
- Creator edits own item.
- `team_lead` cancels/deletes own-team item.
- Multi-day item can be queried by date range.

### Phase HNP-04 — Planning calendar backend

Goal: expose one range API that combines planning sources.

Backend checklist:

- [ ] Add calendar response DTO shared by all plan types.
- [ ] Implement range query for monthly-plan metadata.
- [ ] Implement range query for team plans.
- [ ] Compose results in `GET /v1/planning-calendar`.
- [ ] Add tests for multi-day expansion or frontend-ready date grouping.
- [ ] Add tests for action permissions by role.

Acceptance criteria:

- One API call can load a month view.
- Calendar returns monthly-plan and team-plan items.
- Actions are scoped without hiding awareness rows unexpectedly.
- No N+1 team/user lookups in obvious paths.

### Phase HNP-05 — Calendar frontend MVP

Goal: deliver Google Calendar-style monthly visibility.

Frontend checklist:

- [ ] Add calendar route/navigation.
- [ ] Add month selector.
- [ ] Add calendar grid with day badges/counts.
- [ ] Add day detail drawer/list.
- [ ] Add filters by type/team.
- [ ] Add create team-plan entry path.
- [ ] Add links/actions to monthly-plan files when permitted.

Acceptance criteria:

- Month loads through range API.
- Date click shows team plans and monthly plans.
- Multi-day items display across dates.
- Empty/loading/error states are readable.
- Mobile viewport remains usable.

### Phase HNP-06 — Contact directory

Goal: make coordination contact info available.

Backend checklist:

- [ ] Confirm existing user fields for phone/position.
- [ ] Add nullable fields if missing.
- [ ] Add directory list/detail endpoint.
- [ ] Add own-contact update endpoint.
- [ ] Add tests for own edit vs other-user edit.

Frontend checklist:

- [ ] Add Contacts page.
- [ ] Add search/filter.
- [ ] Add contact card/list.
- [ ] Add edit-own-contact form.

Acceptance criteria:

- All authenticated users can browse active contacts.
- User can edit own phone/position/contact fields.
- User cannot edit another user's contact info.
- Search by name/team works.

### Phase HNP-07 — งานระดมทีม backend

Goal: add large multi-team planning source.

Backend checklist:

- [ ] Create `internal/feature/largework` vertical slice.
- [x] Add DB/model foundation for `large_work_items` and `large_work_item_teams` via GORM AutoMigrate model registration.
- [ ] Add explicit SQL/Goose migration files if/when the project moves away from current AutoMigrate workflow.
- [ ] Write validation tests: at least two teams, valid owner/participant teams.
- [ ] Write policy tests for create/edit/read.
- [ ] Implement create/list/detail/update/cancel.
- [ ] Integrate large-work items into planning-calendar range API.

Acceptance criteria:

- Large-work item requires multiple teams.
- Appears on calendar for participating teams.
- Unauthorized roles cannot modify.
- Day detail labels type as `งานระดมทีม`.

### Phase HNP-08 — งานระดมทีม frontend

Goal: expose large multi-team planning to operators.

Frontend checklist:

- [ ] Add navigation label using exact term `งานระดมทีม`.
- [ ] Add list page.
- [ ] Add create/edit form for owner team, participating teams, date range, location, details.
- [ ] Add calendar detail rendering for large-work items.
- [ ] Add role-aware actions.

Acceptance criteria:

- Admin/super_admin can create large-work item.
- Calendar shows item on each affected date.
- Participating teams can view details.
- Mobile layout is usable.

### Phase HNP-09 — Daily report prefill MVP

Goal: allow planned work to reduce duplicate daily-report entry.

Backend checklist:

- [ ] Define source reference fields for daily report draft without forcing plan mutation.
- [ ] Add endpoint or query param to start daily report from plan source.
- [ ] Keep actual report editable even when plan data differs.

Frontend checklist:

- [ ] Add Start Daily Report action from plan detail where relevant.
- [ ] Prefill date/location/team/work title fields.
- [ ] Allow user to edit actual report independently.

Acceptance criteria:

- Daily report can be started from team plan or monthly plan.
- Actual report edits do not overwrite original plan unless explicitly supported later.

### Phase HNP-10 — QA, smoke, docs, and release readiness

Goal: verify the full planning workstream.

Checklist:

- [ ] Backend tests pass.
- [ ] Backend vet/build pass.
- [ ] Frontend typecheck/build pass.
- [ ] Browser smoke by role:
  - `super_admin`
  - `admin`
  - `team_lead`
  - `user`
- [ ] Verify `user` own-team monthly-plan upload/download.
- [ ] Verify `user` cannot upload/download another team's monthly plan.
- [ ] Verify calendar combines all implemented plan types.
- [ ] Verify contact directory search and own-profile edit.
- [ ] Update `plan/README.md`, release checklist, and any user-facing docs changed by implementation.

Acceptance criteria:

- All automated gates pass.
- Role behavior matches this PRD.
- No contradictory monthly-plan rule remains in active implementation docs without a superseded note.

---

## 10. Open questions that should not block MVP

These can be decided during implementation without blocking the whole workstream:

1. Exact source of responsible-area truth for team plan validation:
   - MVP can use actor's `teamId` and manual location text.
   - Later can add PEA/feeder/station mapping.
2. Whether `admin` can create team plans:
   - MVP default: no broad team-plan creation; admin focuses on monthly-plan operations and งานระดมทีม.
3. Whether `team_lead` can create งานระดมทีม directly:
   - MVP default: admin/super_admin create; team_lead proposal can be later.
4. Whether cancelled/deleted plans are soft-deleted or status-only:
   - Prefer status/soft delete for auditability.
5. Whether contact phone is visible to every authenticated user:
   - MVP says yes because directory purpose is coordination; revisit if privacy requirement appears.

---

## 11. Non-goals for MVP

- No complex in-system approval workflow for monthly plan.
- No plan-vs-actual analytics dashboard beyond source reference fields.
- No separate calendar projection table unless performance requires it.
- No broad admin resurrection; `admin` remains monthly-plan/operations scoped.
- No destructive migration or data cleanup without explicit operator approval.

---

## 12. Required verification commands per implementation phase

Backend:

```bash
go test ./...
go vet ./...
go build -o /tmp/hotlines-api main.go
git diff --check
```

Frontend:

```bash
npx tsc --noEmit
npm run build
git diff --check
```

Browser/manual smoke:

- Login as `super_admin`: broad monthly-plan and large-work management works.
- Login as `admin`: monthly-plan broad operations work; no broad user/role/master-data admin access.
- Login as `team_lead`: own-team monthly-plan upload/download works; non-own team blocked; team-plan own-team delete/cancel works.
- Login as `user`: own-team monthly-plan upload/download works; non-own team blocked; own team-plan create/edit works.
- Verify calendar month loads with one range request and displays multi-day items correctly.

---

## 13. Implementation order recommendation

1. Correct monthly-plan policy tests and implementation first because it reconciles the latest product override.
2. Add team-plan backend and calendar backend together enough to prove the data contract.
3. Build calendar frontend MVP after the range API exists.
4. Add contact directory as a small independent vertical slice.
5. Add งานระดมทีม after calendar primitives exist so it becomes one more planning source, not a separate UI island.
6. Add daily-report prefill last after plan source IDs are stable.
