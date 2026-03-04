# Performance Optimization Todo List

## ✅ Priority 1: High Impact (เสร็จสิ้น)

- [x] **1.1** Dashboard/Stats concurrent queries using errgroup
  - Summary() ✅, Stats() ✅, FeederMatrix() ✅
  - ไฟล์: `internal/handlers/v1/dashboard.go`
  
- [x] **1.2** แก้ `Save()` → `.Update()` สำหรับ update แค่ 1-2 field
  - auth.go (Login - lastLogin) ✅
  - task.go (Delete - deletedat) ✅
  - job_detail.go (Delete - deletedAt) ✅
  
- [x] **1.3** เพิ่ม `WithContext` ใน post-write reload
  - task.go (Create, Update) ✅
  - feeder.go, station.go, pea.go มี WithContext อยู่แล้ว ✅
  
- [x] **1.4** แก้ RecoveryMiddleware order
  - ย้ายไปใน `SetupRouter()` ก่อน routes ✅

## ✅ Priority 2: Medium Impact (เสร็จสิ้น)

- [x] **2.1** เพิ่ม Request Timeout Middleware ✅
- [x] **2.2** เพิ่ม Gzip Compression ✅
- [x] **2.3** User List pagination ✅
- [x] **2.4** ลบ unnecessary reload ใน UserHandler.Update ✅
- [x] **2.5** แก้ Stats activeTeams bug ✅ (fixed ใน Stats() concurrent refactor)
- [x] **2.6** แก้ ListByTeam COUNT bug ✅
- [x] **2.7** Bcrypt cost consistency ✅

## ✅ Priority 3: Architecture / Nice-to-Have (เสร็จสิ้น)

- [x] **3.2** Rate Limiting middleware ✅
- [x] **3.3** ลบ Dead Code ✅
  - UploadDirect() ลบแล้ว
  - replaceAll/indexOfString → strings.ReplaceAll ✅
- [x] **3.4** ListByFilter / ListByTeam return map → slice ✅

## ⏸️ Infrastructure (ไม่ใช่ code change)

- [ ] **3.1** Cloud Run min-instances
  - ต้องตั้งค่าใน Cloud Run Console: `min-instances: 1`

## สรุปการ Commit (เรียงตามเวลา)

1. `perf(dashboard): concurrent queries in Summary() using errgroup`
2. `perf(dashboard): concurrent queries in Stats() using errgroup`
3. `perf(handlers): use .Update() instead of .Save() for single-field updates`
4. `fix(task): add WithContext to post-write reload queries`
5. `fix(router): move RecoveryMiddleware before routes in SetupRouter()`
6. `feat(middleware): add Request Timeout Middleware (30s)`
7. `feat(router): add Gzip compression middleware`
8. `feat(user): add pagination to User List endpoint`
9. `perf(user): remove unnecessary reload in UserHandler.Update`
10. `fix(task): apply filters to ListByTeam COUNT query`
11. `fix(user): use pkg/password for bcrypt consistency`
12. `refactor(task): return ordered slice instead of map for ListByFilter/ListByTeam`
13. `refactor: remove dead code`
14. `perf(dashboard): concurrent queries in FeederMatrix() using errgroup`
15. `feat(middleware): add Rate Limiting middleware`

**ทั้งหมด 15 commits** - ทุก commit มีการ build ตรวจสอบ error ก่อน commit เสมอ
