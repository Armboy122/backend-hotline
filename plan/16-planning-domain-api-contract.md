# HNP-01 Planning Domain API Contract

Date: 2026-05-10
Status: implementation contract only; do not write production code from this card
Scope: team plan, planning calendar projection, contact directory, daily-report prefill, and งานระดมทีม
Source: `plan/15-team-plan-largework-implementation-plan.md` (HNP-00), plus inspection of current backend/frontend feature layout

---

## 1. Current implementation facts inspected

Backend repo: `/Users/sakdithat/Desktop/myproject/backend-hotline`
Frontend repo: `/Users/sakdithat/Desktop/myproject/hotlines3`

Current backend structure is feature-first and must remain this way:

```text
internal/feature/<feature>/
  controller/
  service/
  repository/
  dto/
  entity/
  mapper/
```

Relevant current modules:

- Daily report: `internal/feature/task`
  - Current routes: `/v1/tasks`, `/v1/tasks/by-team`, `/v1/tasks/by-filter`, `/v1/tasks/:id`
  - Current DB model: `models.TaskDaily`
  - Current scoping: service uses `entity.Actor` and `CanReadTeam` / `CanWriteTeam`.
- Monthly plan: `internal/feature/monthlyplan`
  - Current routes: `/v1/monthly-plans/:year/overview`, `/v1/monthly-plans/:year/:month/files`, `/v1/monthly-plans/files/:id/download`, settings, restore/delete.
  - Current DB models: `MonthlyPlan`, `PlanFile`, `MonthlyPlanSetting`, `FileSizeLog`.
  - Current metadata already exists on `PlanFile`: `workStartDate`, `workEndDate`, `destination`, `remarks`, `teamId`, `isMasterPlan`.
- Auth/RBAC: `internal/feature/auth/policy/roles.go`, middleware context values `user_id`, `role`, `team_id`.
- API response wrapper: `internal/dto.StandardResponse` = `{ success, data?, meta?, error? }`.
- Frontend service style: `src/lib/services/*.service.ts`, DTOs in `src/types/*.ts`, React Query keys/hooks in `src/hooks/useQueries.ts`.

Contract rules from HNP-00 that override older contradictory text:

- `team_lead` and `user` can view all monthly-plan team rows for awareness.
- `team_lead` and `user` can upload/download only their own team monthly-plan files.
- Team plan is own-area work; no approval.
- `user` and `team_lead` can add own-team team plans.
- Creator can edit own team-plan item.
- `team_lead` can delete/cancel own-team team-plan items.
- งานระดมทีม is now part of this large planning workstream and must use the exact Thai term in labels/docs.

---

## 2. Shared API conventions

All endpoints are under `/v1` and require `Authorization: Bearer <accessToken>` unless explicitly stated otherwise.

Success response:

```json
{
  "success": true,
  "data": {}
}
```

Paginated success response:

```json
{
  "success": true,
  "data": [],
  "meta": { "page": 1, "limit": 50, "total": 123 }
}
```

Error response:

```json
{
  "success": false,
  "error": {
    "code": "FORBIDDEN",
    "message": "forbidden"
  }
}
```

Date/time conventions:

- Dates are `YYYY-MM-DD` strings.
- Optional work time is `HH:mm` string or `null`.
- Multi-day items use inclusive date range: `startDate <= date <= endDate`; if `endDate` is `null`, it equals `startDate`.
- Timestamps are RFC3339 strings.
- IDs are numeric where backed by local DB tables; calendar `sourceId` and UI union IDs may be strings.

Error code conventions for new endpoints:

| HTTP | Code | Use |
|---:|---|---|
| 400 | `BAD_REQUEST` | malformed JSON/query/path, invalid date range, invalid status |
| 401 | `UNAUTHORIZED` | missing/invalid auth context |
| 403 | `FORBIDDEN` | authenticated but action not allowed |
| 404 | `NOT_FOUND` | item/source not found or hidden by policy |
| 409 | `CONFLICT` | invalid state transition, duplicate active relation |
| 500 | `ERROR` | unexpected server failure |

---

## 3. RBAC matrix

Legend: `all` = any team, `own` = actor's `teamId`, `creator` = row `createdByUserId`, `involved` = owner/participant team.

| Capability | super_admin | admin | team_lead | user | viewer |
|---|---:|---:|---:|---:|---:|
| View planning calendar | all | all | all awareness | all awareness | all awareness/read-only |
| View monthly-plan rows | all | all | all awareness | all awareness | all awareness/read-only |
| Upload monthly-plan file | all + lock override | all per settings | own only per settings | own only per settings | no |
| Download monthly-plan file | all | all | own only | own only | no by default |
| Edit monthly-plan settings | yes | yes | no | no | no |
| Create team plan | all | no by default | own | own | no |
| List/read team plans | all | all awareness | all awareness | all awareness | all awareness/read-only |
| Update team plan | all | no by default | own team | creator only | no |
| Cancel/delete team plan | all | no by default | own team | no | no |
| Create งานระดมทีม | all | all | no in MVP | no | no |
| Read งานระดมทีม | all | all | involved + awareness rows | involved + awareness rows | involved/read-only if retained |
| Update/cancel งานระดมทีม | all | all operations | no in MVP | no | no |
| View contact directory | all | all | all active contacts | all active contacts | all active contacts/read-only |
| Update own contact profile | yes | yes | yes | yes | no unless viewer accounts are operational |
| Update another user's contact profile | yes | no by default | no | no | no |
| Start daily-report draft from plan | all allowed sources | all allowed sources | own/involved source | own/involved source | no |

Policy functions to add in the relevant domain/entity package before controller wiring:

- `CanCreateTeamPlan(actor, teamID)`
- `CanUpdateTeamPlan(actor, plan)`
- `CanCancelTeamPlan(actor, plan)`
- `CanCreateLargeWork(actor)`
- `CanUpdateLargeWork(actor, item)`
- `CanReadPlanningSource(actor, source)`
- `CanStartDailyReportDraft(actor, source)`

Each policy function must have table-driven tests before implementation changes.

---

## 4. Team Plan API contract

Backend feature folder: `internal/feature/teamplan`
Frontend DTO file: `src/types/team-plan.ts`
Frontend service: `src/lib/services/team-plan.service.ts`

### 4.1 Routes

| Method | Route | Name | Auth | Purpose |
|---|---|---|---|---|
| POST | `/v1/team-plans` | `teamPlan.create` | authenticated | Create own-area plan item |
| GET | `/v1/team-plans` | `teamPlan.list` | authenticated | Date-range list for page/calendar |
| GET | `/v1/team-plans/:id` | `teamPlan.get` | authenticated | Detail |
| PATCH | `/v1/team-plans/:id` | `teamPlan.update` | authenticated | Partial update |
| POST | `/v1/team-plans/:id/cancel` | `teamPlan.cancel` | authenticated | Status cancel; preferred over hard delete |
| DELETE | `/v1/team-plans/:id` | `teamPlan.delete` | authenticated | Optional alias to cancel/soft delete; do not hard delete |

### 4.2 Request DTOs

`POST /v1/team-plans`

```json
{
  "teamId": 1,
  "title": "Patrol feeder A",
  "workType": "patrol",
  "startDate": "2026-06-03",
  "endDate": "2026-06-05",
  "workTime": "09:00",
  "locationText": "PEA Bang Khen area",
  "peaId": 10,
  "operationCenterId": 2,
  "feederId": 33,
  "stationId": 7,
  "notes": "Prepare outage notice"
}
```

Required fields: `teamId`, `title`, `startDate`, `locationText`.

Validation:

- `teamId > 0`.
- `title` and `locationText` non-empty after trim.
- `startDate` valid `YYYY-MM-DD`.
- `endDate` optional; if present, `endDate >= startDate`.
- `workTime` optional; if present, `HH:mm`.
- `user` and `team_lead` cannot choose a `teamId` different from actor `teamId`.

`PATCH /v1/team-plans/:id`

```json
{
  "title": "Updated title",
  "workType": "maintenance",
  "startDate": "2026-06-04",
  "endDate": null,
  "workTime": null,
  "locationText": "Updated location",
  "peaId": null,
  "operationCenterId": 2,
  "feederId": 33,
  "stationId": null,
  "notes": "Updated notes",
  "status": "planned"
}
```

Allowed statuses: `planned`, `cancelled`, `completed`.

`GET /v1/team-plans` query params:

| Param | Type | Required | Notes |
|---|---|---:|---|
| `from` | `YYYY-MM-DD` | yes | inclusive |
| `to` | `YYYY-MM-DD` | yes | inclusive |
| `teamId` | number | no | filter; actions still scoped |
| `status` | string | no | comma-separated statuses |
| `page` | number | no | default 1 |
| `limit` | number | no | default 50, max 100 |

Date-range overlap rule:

```text
start_date <= :to AND COALESCE(end_date, start_date) >= :from
```

### 4.3 Response DTOs

`TeamPlanResponse`

```json
{
  "id": 101,
  "teamId": 1,
  "createdByUserId": 5,
  "title": "Patrol feeder A",
  "workType": "patrol",
  "startDate": "2026-06-03",
  "endDate": "2026-06-05",
  "workTime": "09:00",
  "locationText": "PEA Bang Khen area",
  "peaId": 10,
  "operationCenterId": 2,
  "feederId": 33,
  "stationId": 7,
  "notes": "Prepare outage notice",
  "status": "planned",
  "dailyTaskId": null,
  "team": { "id": 1, "name": "Team A" },
  "createdBy": { "id": 5, "username": "123456", "displayName": null },
  "actions": {
    "canEdit": true,
    "canCancel": true,
    "canDelete": true,
    "canStartDailyReport": true
  },
  "createdAt": "2026-05-10T01:00:00Z",
  "updatedAt": "2026-05-10T01:00:00Z",
  "deletedAt": null
}
```

List response is `StandardResponse` with `data: TeamPlanResponse[]` and `meta`.

### 4.4 Frontend TypeScript DTOs

```ts
export type TeamPlanStatus = 'planned' | 'cancelled' | 'completed'

export interface TeamPlanRequest {
  teamId: number
  title: string
  workType?: string | null
  startDate: string
  endDate?: string | null
  workTime?: string | null
  locationText: string
  peaId?: number | null
  operationCenterId?: number | null
  feederId?: number | null
  stationId?: number | null
  notes?: string | null
}

export type UpdateTeamPlanRequest = Partial<TeamPlanRequest> & {
  status?: TeamPlanStatus
}

export interface TeamPlanActions {
  canEdit: boolean
  canCancel: boolean
  canDelete: boolean
  canStartDailyReport: boolean
}

export interface TeamPlanResponse extends TeamPlanRequest {
  id: number
  createdByUserId: number
  status: TeamPlanStatus
  dailyTaskId: number | null
  team?: { id: number; name: string }
  createdBy?: { id: number; username: string; displayName: string | null }
  actions: TeamPlanActions
  createdAt: string
  updatedAt: string
  deletedAt: string | null
}
```

### 4.5 DB entity

Table: `team_plans`

| Column | Type | Null | Notes |
|---|---|---:|---|
| `id` | bigint identity | no | primary key |
| `team_id` | bigint | no | FK to `Team.id` |
| `created_by_user_id` | bigint | no | FK to `User.id` |
| `title` | text | no | short work title |
| `work_type` | text | yes | rough work type |
| `start_date` | date | no | calendar start |
| `end_date` | date | yes | null = one day |
| `work_time` | text | yes | `HH:mm`; text keeps migration simple |
| `location_text` | text | no | manual location |
| `pea_id` | bigint | yes | FK to `Pea.id` if present |
| `operation_center_id` | bigint | yes | FK to `OperationCenter.id` |
| `feeder_id` | bigint | yes | FK to `Feeder.id` |
| `station_id` | bigint | yes | FK to `Station.id` |
| `notes` | text | yes | optional |
| `status` | text | no | default `planned` |
| `daily_task_id` | bigint | yes | FK to `TaskDaily.id`, later sync marker |
| `created_at` | timestamptz | no | default current timestamp |
| `updated_at` | timestamptz | no | set by app |
| `deleted_at` | timestamptz | yes | soft delete |

Indexes:

- `idx_team_plans_team_start_date (team_id, start_date)`
- `idx_team_plans_date_range (start_date, end_date)`
- `idx_team_plans_created_by (created_by_user_id)`
- `idx_team_plans_status (status)`
- Optional partial index: `idx_team_plans_active_range (start_date, end_date) WHERE deleted_at IS NULL`

Migration safety:

- Additive table only; no existing table rewrite.
- Create indexes concurrently if using a manual SQL migration against production.
- Do not add `NOT NULL DEFAULT now()` to existing large tables in this phase.

Implementation note — HNT-01 (2026-05-10): persistence foundation exists as `models.TeamPlan`, is registered in `pkg/db.MigrationModels()` for existing AutoMigrate flow, and has a Goose-compatible SQL migration at `pkg/db/migrations/20260510014500_create_team_plans.sql`. Controllers/repository/service remain intentionally deferred to later HNT cards.

---

## 5. Planning Calendar Projection API contract

Backend feature folder: `internal/feature/planningcalendar`
Frontend DTO file: `src/types/planning-calendar.ts`
Frontend service: `src/lib/services/planning-calendar.service.ts`

The calendar module owns composition only. It must not duplicate source records into a calendar table for MVP.

### 5.1 Routes

| Method | Route | Name | Auth | Purpose |
|---|---|---|---|---|
| GET | `/v1/planning-calendar` | `planningCalendar.range` | authenticated | One date-range call for visible calendar |
| GET | `/v1/planning-calendar/day/:date` | `planningCalendar.day` | authenticated | Optional convenience wrapper for one day |

### 5.2 Query params

`GET /v1/planning-calendar`

| Param | Type | Required | Notes |
|---|---|---:|---|
| `from` | `YYYY-MM-DD` | yes | inclusive |
| `to` | `YYYY-MM-DD` | yes | inclusive; max span 62 days for MVP |
| `teamId` | number | no | filter display; not a permission bypass |
| `types` | comma string | no | `monthly_plan,team_plan,large_work`; default all |

Validation:

- `from` and `to` must parse.
- `to >= from`.
- Maximum range 62 days initially to protect query cost.

### 5.3 Response DTO

```json
{
  "success": true,
  "data": {
    "from": "2026-06-01",
    "to": "2026-06-30",
    "items": [
      {
        "id": "team_plan:101",
        "type": "team_plan",
        "sourceId": 101,
        "title": "Patrol feeder A",
        "startDate": "2026-06-03",
        "endDate": "2026-06-05",
        "workTime": "09:00",
        "dateKeys": ["2026-06-03", "2026-06-04", "2026-06-05"],
        "teamIds": [1],
        "teams": [{ "id": 1, "name": "Team A", "role": "owner" }],
        "locationText": "PEA Bang Khen area",
        "electricArea": {
          "peaId": 10,
          "peaName": "PEA A",
          "operationCenterId": 2,
          "operationCenterName": "OC A",
          "feederId": 33,
          "feederCode": "F01",
          "stationId": 7,
          "stationName": "ST A"
        },
        "status": "planned",
        "source": {
          "route": "/v1/team-plans/101",
          "dailyReportPrefillRoute": "/v1/daily-report-drafts/from-plan"
        },
        "actions": {
          "canView": true,
          "canEdit": true,
          "canCancel": true,
          "canUpload": false,
          "canDownload": false,
          "canStartDailyReport": true
        }
      }
    ],
    "summary": {
      "total": 1,
      "byType": { "team_plan": 1, "monthly_plan": 0, "large_work": 0 }
    }
  }
}
```

Allowed `type` values:

- `team_plan`
- `monthly_plan`
- `large_work`

Calendar source mapping:

| Type | Source table/API | sourceId | Notes |
|---|---|---:|---|
| `team_plan` | `team_plans` | `team_plans.id` | own-area plan |
| `monthly_plan` | `PlanFile` | `PlanFile.id` | include only files with work date metadata; master plan can appear only if it has date/location metadata |
| `large_work` | `large_work_items` | `large_work_items.id` | งานระดมทีม |

Monthly-plan calendar projection uses current `PlanFile` fields:

- `workStartDate` -> `startDate`
- `workEndDate` -> `endDate`
- `destination` -> `locationText`
- `description` -> `title` fallback to `fileName`
- `remarks` -> detail/notes if exposed later

### 5.4 Frontend TypeScript DTOs

```ts
export type PlanningItemType = 'team_plan' | 'monthly_plan' | 'large_work'

export interface PlanningCalendarTeamRef {
  id: number
  name: string
  role: 'owner' | 'participant' | 'target'
}

export interface PlanningElectricAreaRef {
  peaId: number | null
  peaName: string | null
  operationCenterId: number | null
  operationCenterName: string | null
  feederId: number | null
  feederCode: string | null
  stationId: number | null
  stationName: string | null
}

export interface PlanningCalendarActions {
  canView: boolean
  canEdit: boolean
  canCancel: boolean
  canUpload: boolean
  canDownload: boolean
  canStartDailyReport: boolean
}

export interface PlanningCalendarItem {
  id: string
  type: PlanningItemType
  sourceId: number
  title: string
  startDate: string
  endDate: string | null
  workTime: string | null
  dateKeys: string[]
  teamIds: number[]
  teams: PlanningCalendarTeamRef[]
  locationText: string | null
  electricArea: PlanningElectricAreaRef
  status: string
  source: {
    route: string
    dailyReportPrefillRoute?: string | null
  }
  actions: PlanningCalendarActions
}

export interface PlanningCalendarResponse {
  from: string
  to: string
  items: PlanningCalendarItem[]
  summary: {
    total: number
    byType: Record<PlanningItemType, number>
  }
}
```

---

## 6. Contact Directory API contract

Backend feature folder: `internal/feature/contactdirectory` or extension of `internal/feature/user` with a separate controller; prefer `contactdirectory` if policy grows.
Frontend DTO file: `src/types/contact-directory.ts`
Frontend service: `src/lib/services/contact-directory.service.ts`

### 6.1 Routes

| Method | Route | Name | Auth | Purpose |
|---|---|---|---|---|
| GET | `/v1/contact-directory` | `contactDirectory.list` | authenticated | Search active contacts |
| GET | `/v1/contact-directory/:userId` | `contactDirectory.get` | authenticated | Contact detail |
| PATCH | `/v1/users/me/contact` | `contactDirectory.updateOwn` | authenticated | Actor updates own phone/position/display fields |
| PATCH | `/v1/users/:id/contact` | `contactDirectory.updateOther` | super_admin | Super admin updates another user's contact fields |

### 6.2 Query params

`GET /v1/contact-directory`

| Param | Type | Required | Notes |
|---|---|---:|---|
| `query` | string | no | username/display/phone search |
| `teamId` | number | no | filter |
| `role` | string | no | `super_admin`, `admin`, `team_lead`, `user`, `viewer` |
| `includeInactive` | bool | no | super_admin only; default false |
| `page` | number | no | default 1 |
| `limit` | number | no | default 50, max 100 |

### 6.3 Request DTO

`PATCH /v1/users/me/contact`

```json
{
  "displayName": "Somchai",
  "position": "Line technician",
  "phoneNumber": "0812345678"
}
```

All fields optional, but at least one field must be present.

Validation:

- `displayName`: max 120 chars.
- `position`: max 120 chars.
- `phoneNumber`: max 40 chars, store as text to preserve leading zero and separators.

### 6.4 Response DTO

```json
{
  "id": 5,
  "username": "123456",
  "displayName": "Somchai",
  "position": "Line technician",
  "phoneNumber": "0812345678",
  "role": "user",
  "teamId": 1,
  "team": { "id": 1, "name": "Team A" },
  "isActive": true,
  "actions": {
    "canEdit": true,
    "canEditRoleOrTeam": false
  },
  "updatedAt": "2026-05-10T01:00:00Z"
}
```

### 6.5 Frontend TypeScript DTOs

```ts
import type { UserRole } from './auth'

export interface ContactDirectoryEntry {
  id: number
  username: string
  displayName: string | null
  position: string | null
  phoneNumber: string | null
  role: UserRole
  teamId: number | null
  team?: { id: number; name: string } | null
  isActive: boolean
  actions: {
    canEdit: boolean
    canEditRoleOrTeam: boolean
  }
  updatedAt: string
}

export interface UpdateContactRequest {
  displayName?: string | null
  position?: string | null
  phoneNumber?: string | null
}
```

### 6.6 DB entity changes

Existing table: `User`

Add nullable columns only if they do not already exist in target DB:

| Column | Type | Null | Notes |
|---|---|---:|---|
| `displayName` | text | yes | user-facing name |
| `position` | text | yes | role/title shown in directory |
| `phoneNumber` | text | yes | preserve leading zero |

Indexes:

- Optional `idx_user_team_active (teamId, isActive)` if directory list is slow.
- Optional trigram/search index later; not MVP unless measured need exists.

Migration safety:

- Nullable columns: low risk.
- Do not backfill required values in migration.
- Keep existing `username` auth semantics unchanged.

---

## 7. งานระดมทีม API contract

Backend feature folder: `internal/feature/largework`
Frontend DTO file: `src/types/large-work.ts`
Frontend service: `src/lib/services/large-work.service.ts`
UI label: use exact term `งานระดมทีม`.

### 7.1 Routes

| Method | Route | Name | Auth | Purpose |
|---|---|---|---|---|
| POST | `/v1/large-work-items` | `largeWork.create` | super_admin/admin | Create multi-team work item |
| GET | `/v1/large-work-items` | `largeWork.list` | authenticated | Date-range/list page |
| GET | `/v1/large-work-items/:id` | `largeWork.get` | authenticated | Detail |
| PATCH | `/v1/large-work-items/:id` | `largeWork.update` | super_admin/admin | Partial update |
| POST | `/v1/large-work-items/:id/cancel` | `largeWork.cancel` | super_admin/admin | Cancel item |
| POST | `/v1/large-work-items/:id/attachments` | `largeWork.attach` | super_admin/admin | Optional after storage policy is ready |

### 7.2 Request DTOs

`POST /v1/large-work-items`

```json
{
  "ownerTeamId": 1,
  "participantTeamIds": [1, 2, 3],
  "title": "งานระดมทีม เปลี่ยนอุปกรณ์หลัก",
  "workType": "major_maintenance",
  "startDate": "2026-06-10",
  "endDate": "2026-06-12",
  "workTime": "08:30",
  "locationText": "Station A feeder F01",
  "peaId": 10,
  "operationCenterId": 2,
  "feederId": 33,
  "stationId": 7,
  "notes": "Need bucket truck and safety observer"
}
```

Required fields: `ownerTeamId`, `participantTeamIds`, `title`, `startDate`, `locationText`.

Validation:

- `ownerTeamId > 0`.
- `participantTeamIds` must contain at least two distinct team IDs.
- `ownerTeamId` must be included in `participantTeamIds`; service may add it automatically but tests should document chosen behavior.
- `endDate >= startDate` if present.
- `status` values: `planned`, `cancelled`, `completed`.

`PATCH /v1/large-work-items/:id` uses the same fields as create, all optional except at least one field must be present.

`GET /v1/large-work-items` query params:

| Param | Type | Required | Notes |
|---|---|---:|---|
| `from` | `YYYY-MM-DD` | no | date range start |
| `to` | `YYYY-MM-DD` | no | date range end |
| `teamId` | number | no | owner/participant filter |
| `status` | string | no | comma-separated statuses |
| `page` | number | no | default 1 |
| `limit` | number | no | default 50, max 100 |

### 7.3 Response DTO

```json
{
  "id": 501,
  "ownerTeamId": 1,
  "createdByUserId": 9,
  "title": "งานระดมทีม เปลี่ยนอุปกรณ์หลัก",
  "workType": "major_maintenance",
  "startDate": "2026-06-10",
  "endDate": "2026-06-12",
  "workTime": "08:30",
  "locationText": "Station A feeder F01",
  "peaId": 10,
  "operationCenterId": 2,
  "feederId": 33,
  "stationId": 7,
  "notes": "Need bucket truck and safety observer",
  "status": "planned",
  "teams": [
    { "id": 1, "name": "Team A", "role": "owner" },
    { "id": 2, "name": "Team B", "role": "participant" }
  ],
  "actions": {
    "canEdit": true,
    "canCancel": true,
    "canStartDailyReport": false
  },
  "createdAt": "2026-05-10T01:00:00Z",
  "updatedAt": "2026-05-10T01:00:00Z",
  "deletedAt": null
}
```

### 7.4 Frontend TypeScript DTOs

```ts
export type LargeWorkStatus = 'planned' | 'cancelled' | 'completed'
export type LargeWorkTeamRole = 'owner' | 'participant'

export interface LargeWorkRequest {
  ownerTeamId: number
  participantTeamIds: number[]
  title: string
  workType?: string | null
  startDate: string
  endDate?: string | null
  workTime?: string | null
  locationText: string
  peaId?: number | null
  operationCenterId?: number | null
  feederId?: number | null
  stationId?: number | null
  notes?: string | null
}

export type UpdateLargeWorkRequest = Partial<LargeWorkRequest> & {
  status?: LargeWorkStatus
}

export interface LargeWorkTeamRef {
  id: number
  name: string
  role: LargeWorkTeamRole
}

export interface LargeWorkResponse extends Omit<LargeWorkRequest, 'participantTeamIds'> {
  id: number
  createdByUserId: number
  status: LargeWorkStatus
  teams: LargeWorkTeamRef[]
  actions: {
    canEdit: boolean
    canCancel: boolean
    canStartDailyReport: boolean
  }
  createdAt: string
  updatedAt: string
  deletedAt: string | null
}
```

### 7.5 DB entities

Table: `large_work_items`

| Column | Type | Null | Notes |
|---|---|---:|---|
| `id` | bigint identity | no | primary key |
| `owner_team_id` | bigint | no | FK to `Team.id` |
| `created_by_user_id` | bigint | no | FK to `User.id` |
| `title` | text | no | includes Thai display name if desired |
| `work_type` | text | yes | rough type |
| `start_date` | date | no | calendar start |
| `end_date` | date | yes | null = one day |
| `work_time` | text | yes | `HH:mm` |
| `location_text` | text | no | manual location |
| `pea_id` | bigint | yes | FK to `Pea.id` |
| `operation_center_id` | bigint | yes | FK to `OperationCenter.id` |
| `feeder_id` | bigint | yes | FK to `Feeder.id` |
| `station_id` | bigint | yes | FK to `Station.id` |
| `notes` | text | yes | optional |
| `status` | text | no | default `planned` |
| `created_at` | timestamptz | no | default current timestamp |
| `updated_at` | timestamptz | no | set by app |
| `deleted_at` | timestamptz | yes | soft delete |

Table: `large_work_item_teams`

| Column | Type | Null | Notes |
|---|---|---:|---|
| `large_work_item_id` | bigint | no | FK to `large_work_items.id` |
| `team_id` | bigint | no | FK to `Team.id` |
| `role` | text | no | `owner` or `participant` |
| `created_at` | timestamptz | no | default current timestamp |

Constraints/indexes:

- Primary key or unique: `(large_work_item_id, team_id)`.
- Index `idx_large_work_items_owner_start_date (owner_team_id, start_date)`.
- Index `idx_large_work_items_date_range (start_date, end_date)`.
- Index `idx_large_work_item_teams_team (team_id)`.
- Service validation must enforce at least two distinct team IDs.

Migration safety:

- Additive tables only.
- Current backend runtime uses GORM AutoMigrate, so HNW-01 registers `models.LargeWorkItem` and `models.LargeWorkItemTeam` in `pkg/db.MigrationModels()` rather than introducing a parallel SQL migration path.
- Add FK constraints as `NOT VALID` then validate in production if/when explicit SQL migrations replace AutoMigrate.
- Create indexes concurrently for production migration if not wrapped by GORM AutoMigrate.

---

## 8. Daily-report prefill API contract

Daily report source module remains `internal/feature/task`; the prefill use case can live in `internal/feature/task/service` or a small `internal/feature/dailyreportprefill` composer if it needs multiple repositories.

Frontend integrates with existing `src/types/task-daily.ts` and task-daily form.

### 8.1 Routes

| Method | Route | Name | Auth | Purpose |
|---|---|---|---|---|
| GET | `/v1/daily-report-drafts/from-plan` | `dailyReportDraft.fromPlan` | authenticated | Return a CreateTask prefill payload from a plan source |
| POST | `/v1/tasks/from-plan` | `task.createFromPlan` | authenticated | Optional convenience: create TaskDaily with source reference; can be deferred |

### 8.2 Query params

`GET /v1/daily-report-drafts/from-plan`

| Param | Type | Required | Notes |
|---|---|---:|---|
| `sourceType` | string | yes | `team_plan`, `monthly_plan`, `large_work` |
| `sourceId` | number | yes | source record ID |
| `workDate` | `YYYY-MM-DD` | no | required only when source spans multiple days and UI wants a specific date |

### 8.3 Response DTO

```json
{
  "success": true,
  "data": {
    "source": {
      "sourceType": "team_plan",
      "sourceId": 101,
      "title": "Patrol feeder A"
    },
    "prefill": {
      "workDate": "2026-06-03",
      "teamId": 1,
      "jobTypeId": null,
      "jobDetailId": null,
      "feederId": 33,
      "numPole": null,
      "deviceCode": null,
      "detail": "Patrol feeder A\nPEA Bang Khen area\nPrepare outage notice",
      "urlsBefore": [],
      "urlsAfter": [],
      "latitude": null,
      "longitude": null
    },
    "warnings": [
      "jobTypeId and jobDetailId must be selected before saving"
    ]
  }
}
```

Prefill mapping:

| Daily field | team_plan | monthly_plan | large_work |
|---|---|---|---|
| `workDate` | selected date or `startDate` | selected date or `workStartDate` | selected date or `startDate` |
| `teamId` | plan `teamId` | file `teamId` if present else actor team | actor team if participating else owner team for admin |
| `feederId` | plan `feederId` | null unless monthly metadata later adds it | item `feederId` |
| `detail` | title/location/notes | description/destination/remarks | title/location/notes |
| `jobTypeId` | null | null | null |
| `jobDetailId` | null | null | null |

Important rule: creating/editing the actual daily report must not mutate the source plan. Store source-reference fields only if explicitly added in a later migration.

### 8.4 Frontend TypeScript DTOs

```ts
import type { CreateTaskRequest } from './task-daily'
import type { PlanningItemType } from './planning-calendar'

export interface DailyReportDraftSource {
  sourceType: PlanningItemType
  sourceId: number
  title: string
}

export interface DailyReportDraftFromPlanResponse {
  source: DailyReportDraftSource
  prefill: CreateTaskRequest
  warnings: string[]
}
```

### 8.5 Optional future DB fields on `TaskDaily`

Do not add these in HNP-01. If a later implementation needs traceability, add nullable fields safely:

| Column | Type | Null | Notes |
|---|---|---:|---|
| `plan_source_type` | text | yes | `team_plan`, `monthly_plan`, `large_work` |
| `plan_source_id` | bigint | yes | source record ID |

Add index only if queried: `(plan_source_type, plan_source_id)`.

---

## 9. Monthly-plan contract corrections needed by this workstream

Existing monthly-plan routes stay, but the contract for HNP implementation must be stricter than the current DTO.

### 9.1 Existing routes to preserve

| Method | Route | Contract change |
|---|---|---|
| GET | `/v1/monthly-plans/:year/overview` | Must eventually expose all team rows for awareness, not only actor-owned files. Row actions must be per team. |
| POST | `/v1/monthly-plans/:year/:month/files/presign` | `team_lead`/`user` can presign only for own team. |
| POST | `/v1/monthly-plans/:year/:month/files` | `team_lead`/`user` can confirm only for own team. |
| GET | `/v1/monthly-plans/files/:id/download` | `team_lead`/`user` can download only own-team files. |
| GET | `/v1/monthly-plans/:year/:month/files?teamId=` | Non-admin request for another team must not leak file download data. |

### 9.2 Future year overview row shape

Current `MonthlyPlanActionResponse` has only `canUpload`. HNP phases should extend it to row/team-aware actions:

```json
{
  "year": 2026,
  "months": [
    {
      "month": 6,
      "deadline": "2026-05-23",
      "isLocked": false,
      "teams": [
        {
          "team": { "id": 1, "name": "Team A" },
          "status": "has_files",
          "fileCount": 2,
          "actions": {
            "canUpload": true,
            "canDownload": true,
            "canReplace": true,
            "canSoftDelete": true
          },
          "files": []
        }
      ]
    }
  ]
}
```

This allows `team_lead`/`user` to see every team row while only receiving own-team file actions.

---

## 10. Test plan

No production code should be written under HNP-01, but later implementation cards must start with these tests.

### 10.1 Backend policy tests

Files to create first:

- `internal/feature/teamplan/entity/policy_test.go`
- `internal/feature/largework/entity/policy_test.go`
- `internal/feature/planningcalendar/service/service_test.go`
- `internal/feature/contactdirectory/service/service_test.go`
- Extend `internal/feature/monthlyplan/entity/policy_test.go`
- Extend `internal/feature/monthlyplan/service/service_test.go`

Required cases:

1. Monthly plan:
   - `user` can upload/download own-team file.
   - `user` cannot upload/download non-own-team file.
   - `team_lead` can upload/download own-team file.
   - `team_lead` cannot upload/download non-own-team file.
   - `team_lead`/`user` can see all team rows in overview with false actions for non-own rows.
   - `admin` and `super_admin` can manage all team rows.
2. Team plan:
   - `user` creates own-team item.
   - `user` cannot create other-team item.
   - Creator edits own item.
   - Non-creator `user` cannot edit another user's item.
   - `team_lead` edits/cancels own-team item.
   - `team_lead` cannot cancel other-team item.
   - Multi-day query returns item when date ranges overlap.
3. Calendar:
   - One range request returns `monthly_plan`, `team_plan`, and `large_work` items.
   - Multi-day item includes all inclusive `dateKeys`.
   - Unsupported `types` param returns 400.
   - Range longer than max returns 400.
   - Action flags are scoped without hiding awareness rows.
4. Contact directory:
   - All authenticated roles can list active contacts.
   - Own contact update succeeds.
   - Normal user cannot update another user's contact.
   - `super_admin` can update another user's contact.
   - Inactive users hidden by default.
5. Daily-report prefill:
   - Team-plan prefill maps date/team/feeder/detail.
   - Monthly-plan prefill maps workStartDate/team/destination/detail.
   - Large-work prefill rejects actor not involved unless admin/super_admin.
   - Multi-day source uses requested `workDate` when valid.

### 10.2 Backend controller tests

Add controller tests after service tests are red:

- `internal/feature/teamplan/controller/v1_test.go`
- `internal/feature/largework/controller/v1_test.go`
- `internal/feature/planningcalendar/controller/v1_test.go`
- `internal/feature/contactdirectory/controller/v1_test.go`

Assert:

- Route params and query params bind exactly as documented.
- Invalid dates return 400.
- Forbidden role/team actions return 403.
- Response JSON uses `StandardResponse` wrapper.
- `meta` exists on paginated list endpoints.

### 10.3 Repository tests

Add repository tests where SQLite/GORM test setup can cover query shape. Minimum useful coverage:

- Date-range overlap query for `team_plans`.
- Date-range overlap query for `large_work_items` joined through `large_work_item_teams`.
- Contact directory search and active filter.
- Calendar monthly-plan projection excludes files with no usable date unless UI explicitly asks for undated items.

### 10.4 Frontend type/view-model tests

Files to create first:

- `src/types/planning-calendar.test.ts`
- `src/types/team-plan.test.ts`
- `src/types/large-work.test.ts`
- Extend `src/types/monthly-plan.test.ts`
- Extend `src/lib/auth/role-policy.test.ts`

Required cases:

- `user` own-team monthly-plan upload/download actions true; non-own false.
- Calendar groups multi-day item into every date key.
- Calendar accepts `งานระดมทีม` label for large-work display.
- Team-plan form allows optional `workTime` and optional electric-area IDs.
- Daily-report draft response is assignable to existing `CreateTaskRequest`.

### 10.5 Verification commands

Backend:

```bash
go test ./...
go vet ./...
go build -o /tmp/hotlines-api main.go
git diff --check
```

Frontend after frontend contracts are added:

```bash
npx tsc --noEmit
npm run build
git diff --check
```

HNP-01 docs-only verification:

```bash
git diff --check
```

---

## 11. Implementation order after this contract

1. Correct monthly-plan policy tests and year overview team-row actions.
2. Add `teamplan` backend vertical slice.
3. Add `planningcalendar` backend composition for monthly-plan + team-plan.
4. Add contact directory backend and own-contact update.
5. Add `largework` backend vertical slice and include it in calendar.
6. Add daily-report prefill endpoint after source IDs and DTOs are stable.
7. Only then build the frontend calendar/team-plan/contact/large-work pages.

---

## 12. Non-goals for this contract

- No production code in HNP-01.
- No in-system approval workflow for monthly plan.
- No destructive migration.
- No separate materialized calendar table for MVP.
- No broad `admin` resurrection; admin remains monthly-plan/operations scoped.
- No plan-vs-actual analytics yet.
