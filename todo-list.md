# Performance Optimization Todo List

## Priority 1: High Impact (ต้องทำ)

- [x] **1.1** Dashboard/Stats concurrent queries using errgroup (Summary done)
  - ไฟล์: `internal/handlers/v1/dashboard.go` (Stats, Summary✅, FeederMatrix)
  - สถานะ: Summary() เสร็จแล้ว, ยังเหลือ Stats() และ FeederMatrix()
  
- [ ] **1.2** แก้ `Save()` → `.Update()` สำหรับ update แค่ 1-2 field
  - ไฟล์: `internal/handlers/v1/auth.go` (Login - lastLogin)
  - ไฟล์: `internal/handlers/v1/task.go` (Delete - deletedat)
  - ไฟล์: `internal/handlers/v1/job_detail.go` (Delete - deletedAt)
  
- [ ] **1.3** เพิ่ม `WithContext` ใน post-write reload
  - ไฟล์: `internal/handlers/v1/task.go` (Create, Update)
  - ไฟล์: `internal/handlers/v1/feeder.go` (Update)
  - ไฟล์: `internal/handlers/v1/station.go` (Create, Update)
  - ไฟล์: `internal/handlers/v1/pea.go` (Create, Update)
  
- [ ] **1.4** แก้ RecoveryMiddleware order (ย้ายก่อน routes หรือลบ)
  - ไฟล์: `main.go`, `internal/router/router.go`

## Priority 2: Medium Impact (ควรทำ)

- [ ] **2.1** เพิ่ม Request Timeout Middleware
  - สร้าง middleware ใหม่
  
- [ ] **2.2** เพิ่ม Gzip Compression
  - ไฟล์: `internal/router/router.go`
  
- [ ] **2.3** User List pagination
  - ไฟล์: `internal/handlers/v1/user.go`
  
- [ ] **2.4** ลบ unnecessary reload ใน UserHandler.Update
  - ไฟล์: `internal/handlers/v1/user.go`
  
- [ ] **2.5** แก้ Stats activeTeams bug (filter ตาม date range)
  - ไฟล์: `internal/handlers/v1/dashboard.go`
  
- [ ] **2.6** แก้ ListByTeam COUNT bug (apply filters)
  - ไฟล์: `internal/handlers/v1/task.go`
  
- [ ] **2.7** Bcrypt cost consistency
  - ไฟล์: `internal/handlers/v1/user.go`

## Priority 3: Architecture / Nice-to-Have

- [ ] **3.1** Cloud Run Cold Start Optimization (min-instances)
- [ ] **3.2** Rate Limiting middleware
- [ ] **3.3** ลบ Dead Code
  - ไฟล์: `internal/handlers/v1/upload.go`
  - ไฟล์: `internal/models/...`
- [ ] **3.4** ListByFilter / ListByTeam return `map` → slice
  - ไฟล์: `internal/handlers/v1/task.go`

## Workflow

ทุกครั้งที่ทำงาน:
1. อ่านไฟล์ที่ต้องแก้ไข
2. แก้ไขตาม requirement
3. รัน `go build ./...` ตรวจสอบ error
4. ถ้าไม่มี error → commit ด้วยข้อความที่บ่งบอกถึง task ที่ทำ
5. อัพเดท todo-list.md (checkbox)
6. ไป task ถัดไป
