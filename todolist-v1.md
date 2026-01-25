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

## Phase 5: Authentication (4 endpoints) - PENDING

- [ ] POST `/v1/auth/login`
- [ ] POST `/v1/auth/logout`
- [ ] POST `/v1/auth/refresh`
- [ ] GET `/v1/auth/me`

---

## Phase 6: PDF Reports (4 endpoints) - PENDING

- [ ] GET `/v1/reports/tasks/pdf`
- [ ] GET `/v1/reports/tasks/pdf/preview`
- [ ] POST `/v1/reports/tasks/pdf/batch`
- [ ] GET `/v1/reports/jobs/:jobId`

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
| 2026-01-24 | Setup | - | Initial checklist created |
| 2026-01-24 | 1-4 | feat(api-v1): implement all Phase 1-4 endpoints | All Phase 1-4 endpoints implemented |

---

## API Summary

**Total v1 Endpoints Implemented: 53**

| Category | Count | Status |
|----------|-------|--------|
| Phase 1: Mobile Form | 6 | ✅ |
| Phase 2: Task Management | 6 | ✅ |
| Phase 3: Master Data CRUD | 36 | ✅ |
| Phase 4: Dashboard | 5 | ✅ |
| Phase 5: Authentication | 4 | ⏳ Pending |
| Phase 6: PDF Reports | 4 | ⏳ Pending |

---

## Key Features

1. **Standard Response Format**: All v1 endpoints use `{ success, data, meta?, error? }` format
2. **_count Field**: Teams, JobTypes, JobDetails, Feeders include task count
3. **Nested Relations**: Feeders include Station → OperationCenter
4. **Soft Delete**: JobDetails and Tasks use soft delete with restore capability
5. **Pagination**: Tasks list supports page/limit with meta
6. **Presigned URL**: Upload uses R2 presigned URLs for direct client upload
7. **Backward Compatible**: Legacy `/api/*` routes still work
