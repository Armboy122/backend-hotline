# Hotline Backend API v1 - Implementation Checklist

## Overview
Implementation tracking for API v1 endpoints as specified in `api.md`

---

## Phase 1: Mobile Form MVP (6 endpoints) ✅ COMPLETED

### Setup
- [x] Create standard response DTOs (`internal/dto/response.go`)
- [x] Setup v1 route group in router

### Endpoints
- [x] GET `/v1/teams` - ดึงรายชื่อทีม (with `_count`)
- [x] GET `/v1/job-types` - ดึงประเภทงาน (with `_count`)
- [x] GET `/v1/job-details` - ดึงรายละเอียดงาน (with timestamps, `_count`)
- [x] GET `/v1/feeders` - ดึงฟีดเดอร์ (with nested station.operationCenter, `_count`)
- [x] POST `/v1/upload/image` - อัปโหลดรูปภาพ (presigned URL)
- [x] POST `/v1/tasks` - บันทึกงานใหม่

---

## Phase 2: Task Management (6 endpoints) ✅ COMPLETED

- [x] GET `/v1/tasks` - รายการงาน (with pagination meta)
- [x] GET `/v1/tasks/:id` - ดึงงานตาม ID
- [x] PUT `/v1/tasks/:id` - แก้ไขงาน
- [x] DELETE `/v1/tasks/:id` - ลบงาน (return 204)
- [x] GET `/v1/tasks/by-filter` - กรองตามปี/เดือน/ทีม
- [x] GET `/v1/tasks/by-team` - จัดกลุ่มตามทีม

---

## Phase 3: Master Data CRUD (36 endpoints) ✅ COMPLETED

### Teams (5)
- [x] GET `/v1/teams`
- [x] GET `/v1/teams/:id`
- [x] POST `/v1/teams`
- [x] PUT `/v1/teams/:id`
- [x] DELETE `/v1/teams/:id`

### Job Types (5)
- [x] GET `/v1/job-types`
- [x] GET `/v1/job-types/:id`
- [x] POST `/v1/job-types`
- [x] PUT `/v1/job-types/:id`
- [x] DELETE `/v1/job-types/:id`

### Job Details (6)
- [x] GET `/v1/job-details`
- [x] GET `/v1/job-details/:id`
- [x] POST `/v1/job-details`
- [x] PUT `/v1/job-details/:id`
- [x] DELETE `/v1/job-details/:id`
- [x] POST `/v1/job-details/:id/restore` - กู้คืน (NEW)

### Feeders (5)
- [x] GET `/v1/feeders`
- [x] GET `/v1/feeders/:id`
- [x] POST `/v1/feeders`
- [x] PUT `/v1/feeders/:id`
- [x] DELETE `/v1/feeders/:id`

### Stations (5)
- [x] GET `/v1/stations`
- [x] GET `/v1/stations/:id`
- [x] POST `/v1/stations`
- [x] PUT `/v1/stations/:id`
- [x] DELETE `/v1/stations/:id`

### PEAs (6)
- [x] GET `/v1/peas`
- [x] GET `/v1/peas/:id`
- [x] POST `/v1/peas`
- [x] POST `/v1/peas/bulk` - Bulk import
- [x] PUT `/v1/peas/:id`
- [x] DELETE `/v1/peas/:id`

### Operation Centers (5)
- [x] GET `/v1/operation-centers`
- [x] GET `/v1/operation-centers/:id`
- [x] POST `/v1/operation-centers`
- [x] PUT `/v1/operation-centers/:id`
- [x] DELETE `/v1/operation-centers/:id`

---

## Phase 4: Dashboard (5 endpoints) ✅ COMPLETED

- [x] GET `/v1/dashboard/summary` - สรุปภาพรวม
- [x] GET `/v1/dashboard/top-jobs` - Top 10 งาน
- [x] GET `/v1/dashboard/top-feeders` - Top 10 ฟีดเดอร์
- [x] GET `/v1/dashboard/feeder-matrix` - Matrix ฟีดเดอร์-งาน (NEW)
- [x] GET `/v1/dashboard/stats` - สถิติขั้นสูง

---

## 🧪 Testing Checklist (Phase 1-4)

### Pre-requisites
- [ ] Start server: `go run ./cmd/server`
- [ ] Verify server running: `curl http://localhost:8080/health`
- [ ] Database connected and has test data

---

### Phase 1: Mobile Form MVP Testing

#### GET /v1/teams
```bash
curl http://localhost:8080/v1/teams
```
- [ ] Returns `{ success: true, data: [...] }`
- [ ] Each team has `id`, `name`, `_count.tasks`

#### GET /v1/job-types
```bash
curl http://localhost:8080/v1/job-types
```
- [ ] Returns list with `_count.tasks` field
- [ ] Each job type has `id`, `name`

#### GET /v1/job-details
```bash
curl http://localhost:8080/v1/job-details
```
- [ ] Returns list with `createdAt`, `updatedAt`, `deletedAt`
- [ ] Has `_count.tasks` field
- [ ] Has `jobTypeId` field

#### GET /v1/feeders
```bash
curl http://localhost:8080/v1/feeders
```
- [ ] Returns nested `station.operationCenter`
- [ ] Has `_count.tasks` field
- [ ] Each feeder has `id`, `code`, `stationId`

#### POST /v1/upload/image
```bash
curl -X POST http://localhost:8080/v1/upload/image \
  -H "Content-Type: application/json" \
  -d '{"fileName": "test.jpg", "fileType": "image/jpeg"}'
```
- [ ] Returns `uploadUrl` (presigned URL)
- [ ] Returns `fileUrl` (public URL)
- [ ] Returns `fileKey`
- [ ] Rejects invalid file types (e.g., `application/pdf`)

#### POST /v1/tasks
```bash
curl -X POST http://localhost:8080/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "workDate": "2026-01-25",
    "teamId": 1,
    "jobTypeId": 1,
    "jobDetailId": 1,
    "feederId": 1,
    "urlsBefore": ["https://example.com/before.jpg"],
    "urlsAfter": []
  }'
```
- [ ] Returns 201 Created
- [ ] Returns task with all relations (team, jobType, jobDetail, feeder)
- [ ] `workDate` format correct
- [ ] Validates required fields

---

### Phase 2: Task Management Testing

#### GET /v1/tasks
```bash
curl "http://localhost:8080/v1/tasks?page=1&limit=10"
```
- [ ] Returns with pagination `meta: { page, limit, total }`
- [ ] Supports `workDate`, `teamId`, `jobTypeId`, `feederId` filters

#### GET /v1/tasks/:id
```bash
curl http://localhost:8080/v1/tasks/1
```
- [ ] Returns task with all relations
- [ ] Returns 404 for non-existent ID

#### PUT /v1/tasks/:id
```bash
curl -X PUT http://localhost:8080/v1/tasks/1 \
  -H "Content-Type: application/json" \
  -d '{"detail": "Updated detail"}'
```
- [ ] Returns updated task
- [ ] Only updates provided fields

#### DELETE /v1/tasks/:id
```bash
curl -X DELETE http://localhost:8080/v1/tasks/1
```
- [ ] Returns 204 No Content
- [ ] Task is soft-deleted (not hard delete)

#### GET /v1/tasks/by-filter
```bash
curl "http://localhost:8080/v1/tasks/by-filter?year=2026&month=1"
```
- [ ] Returns tasks grouped by team
- [ ] Supports `teamId` filter

#### GET /v1/tasks/by-team
```bash
curl http://localhost:8080/v1/tasks/by-team
```
- [ ] Returns tasks grouped by team name
- [ ] Each group has `team` info and `tasks` array

---

### Phase 3: Master Data CRUD Testing

#### Teams CRUD
```bash
# Create
curl -X POST http://localhost:8080/v1/teams \
  -H "Content-Type: application/json" \
  -d '{"name": "ทีม Test"}'

# Read
curl http://localhost:8080/v1/teams/1

# Update
curl -X PUT http://localhost:8080/v1/teams/1 \
  -H "Content-Type: application/json" \
  -d '{"name": "ทีม Updated"}'

# Delete
curl -X DELETE http://localhost:8080/v1/teams/1
```
- [ ] Create returns 201
- [ ] Read returns single item
- [ ] Update returns updated item
- [ ] Delete returns 204

#### Job Types CRUD
- [ ] Same CRUD pattern as Teams

#### Job Details CRUD + Restore
```bash
# Delete (soft)
curl -X DELETE http://localhost:8080/v1/job-details/1

# Restore
curl -X POST http://localhost:8080/v1/job-details/1/restore
```
- [ ] Soft delete sets `deletedAt`
- [ ] Restore clears `deletedAt`

#### Feeders CRUD
- [ ] Create requires `stationId`
- [ ] Returns nested station info

#### Stations CRUD
- [ ] Create requires `operationId`, `codeName`
- [ ] `codeName` must be unique

#### PEAs CRUD + Bulk
```bash
# Bulk create
curl -X POST http://localhost:8080/v1/peas/bulk \
  -H "Content-Type: application/json" \
  -d '[
    {"shortname": "กฟย.1", "fullname": "การไฟฟ้า 1", "operationId": 1},
    {"shortname": "กฟย.2", "fullname": "การไฟฟ้า 2", "operationId": 1}
  ]'
```
- [ ] Bulk create works
- [ ] Returns array of created items

#### Operation Centers CRUD
- [ ] Basic CRUD works

---

### Phase 4: Dashboard Testing

#### GET /v1/dashboard/summary
```bash
curl "http://localhost:8080/v1/dashboard/summary?year=2026&month=1"
```
- [ ] Returns `totalTasks`, `totalJobTypes`, `totalFeeders`
- [ ] Returns `topTeam` with `id`, `name`, `count`

#### GET /v1/dashboard/top-jobs
```bash
curl "http://localhost:8080/v1/dashboard/top-jobs?year=2026&limit=10"
```
- [ ] Returns sorted by count desc
- [ ] Each item has `jobTypeName`

#### GET /v1/dashboard/top-feeders
```bash
curl "http://localhost:8080/v1/dashboard/top-feeders?year=2026&limit=10"
```
- [ ] Returns sorted by count desc
- [ ] Each item has `stationName`

#### GET /v1/dashboard/feeder-matrix
```bash
curl "http://localhost:8080/v1/dashboard/feeder-matrix?feederId=1&year=2026"
```
- [ ] Requires `feederId` parameter
- [ ] Returns `jobDetails` array with counts
- [ ] Returns `totalCount`

#### GET /v1/dashboard/stats
```bash
curl "http://localhost:8080/v1/dashboard/stats?startDate=2026-01-01&endDate=2026-12-31"
```
- [ ] Returns `summary` object
- [ ] Returns `charts` with `tasksByFeeder`, `tasksByJobType`, `tasksByTeam`, `tasksByDate`

---

## Files Created/Modified

### New Files
- `internal/dto/response.go` - Standard response structs
- `internal/handlers/v1/team.go` - Team v1 handler
- `internal/handlers/v1/job_type.go` - JobType v1 handler
- `internal/handlers/v1/job_detail.go` - JobDetail v1 handler (with restore)
- `internal/handlers/v1/feeder.go` - Feeder v1 handler
- `internal/handlers/v1/station.go` - Station v1 handler
- `internal/handlers/v1/pea.go` - PEA v1 handler (with bulk)
- `internal/handlers/v1/operation_center.go` - OperationCenter v1 handler
- `internal/handlers/v1/task.go` - Task v1 handler (with by-filter)
- `internal/handlers/v1/upload.go` - Upload handler (presigned URL)
- `internal/handlers/v1/dashboard.go` - Dashboard v1 handler (with feeder-matrix)
- `pkg/s3/r2.go` - Cloudflare R2 client

### Modified Files
- `internal/router/router.go` - Added v1 route group

---

## Progress Log

| Date | Phase | Commits | Notes |
|------|-------|---------|-------|
| 2026-01-25 | Setup | - | Initial checklist created |
| 2026-01-25 | 1-4 | feat(api-v1): implement all Phase 1-4 endpoints | All Phase 1-4 endpoints implemented |
| 2026-01-25 | Testing | - | Added testing checklist for Phase 1-4 |

---

## API Summary

**Total v1 Endpoints Implemented: 53**

| Category | Count | Status |
|----------|-------|--------|
| Phase 1: Mobile Form | 6 | ✅ Implemented |
| Phase 2: Task Management | 6 | ✅ Implemented |
| Phase 3: Master Data CRUD | 36 | ✅ Implemented |
| Phase 4: Dashboard | 5 | ✅ Implemented |
| Phase 5: Authentication | 4 | ⏸️ On Hold |
| Phase 6: PDF Reports | 4 | ⏸️ On Hold |

---

## Pending Phases (On Hold)

### Phase 5: Authentication (4 endpoints)
- [ ] POST `/v1/auth/login`
- [ ] POST `/v1/auth/logout`
- [ ] POST `/v1/auth/refresh`
- [ ] GET `/v1/auth/me`

### Phase 6: PDF Reports (4 endpoints)
- [ ] GET `/v1/reports/tasks/pdf`
- [ ] GET `/v1/reports/tasks/pdf/preview`
- [ ] POST `/v1/reports/tasks/pdf/batch`
- [ ] GET `/v1/reports/jobs/:jobId`

---

## Key Features

1. **Standard Response Format**: All v1 endpoints use `{ success, data, meta?, error? }` format
2. **_count Field**: Teams, JobTypes, JobDetails, Feeders include task count
3. **Nested Relations**: Feeders include Station → OperationCenter
4. **Soft Delete**: JobDetails and Tasks use soft delete with restore capability
5. **Pagination**: Tasks list supports page/limit with meta
6. **Presigned URL**: Upload uses R2 presigned URLs for direct client upload
7. **Backward Compatible**: Legacy `/api/*` routes still work
