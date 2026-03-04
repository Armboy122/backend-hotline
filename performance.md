# Performance Review: backend-hotline on GCP Cloud Run + Neon PostgreSQL

## สรุปภาพรวม

ระบบออกแบบมาได้ดีในหลายจุด เช่น CDN caching ผ่าน Cloudflare, connection pool ที่ tune สำหรับ Neon, batch query helpers ที่ป้องกัน N+1 แต่ยังมีจุดที่ปรับปรุงได้เพื่อลด latency และ DB round-trips อย่างมาก

---

## Priority 1: High Impact (ต้องทำ)

### 1.1 Dashboard/Stats ยิง 13 queries แบบ sequential

**ปัญหา:** `GET /v1/dashboard/stats` ยิง DB ทีละ query รวม ~13 ครั้ง แต่ละครั้งมี network round-trip ไป Neon (~5-10ms) รวมแล้ว 65-130ms เฉพาะ network ไม่รวม query time

**แก้ไข:** ใช้ `golang.org/x/sync/errgroup` ยิง query พร้อมกัน

```go
// Before: sequential (~130ms network)
totalTasks := countTasks(...)
activeTeams := countTeams(...)
topJobType := getTopJobType(...)
// ... อีก 10 queries

// After: concurrent (~10ms network)
g, ctx := errgroup.WithContext(c.Request.Context())
g.Go(func() error { totalTasks, err = countTasks(ctx, ...); return err })
g.Go(func() error { activeTeams, err = countTeams(ctx, ...); return err })
g.Go(func() error { topJobType, err = getTopJobType(ctx, ...); return err })
// ...
if err := g.Wait(); err != nil { ... }
```

**ผลลัพธ์:** ลด latency จาก sum-of-all-queries เหลือ max-of-all-queries (ประมาณ 5-10x เร็วขึ้น)

**ไฟล์:** `internal/handlers/v1/dashboard.go` — ฟังก์ชัน `Stats()`, `Summary()`, `FeederMatrix()`

---

### 1.2 ใช้ `Save()` ทั้ง row สำหรับ update แค่ 1-2 field

**ปัญหา:** หลายจุดใช้ `db.Save(&model)` ซึ่ง GORM จะ UPDATE ทุก column รวมถึง password hash

| Handler | Action | ส่ง UPDATE ทั้ง row | ควรแก้เป็น |
|---------|--------|---------------------|-----------|
| `auth.go` Login | update `lastLogin` | ทุก column รวม password | `.Model(&user).Update("lastLogin", now)` |
| `task.go` Delete | set `deletedat` | ทุก column | `.Model(&task).Update("deletedat", now)` |
| `job_detail.go` Delete | set `deletedAt` | ทุก column | `.Model(&jd).Update("deletedAt", now)` |

**ผลลัพธ์:** ลดขนาด SQL statement, ลด bandwidth ไป DB, ลดความเสี่ยงด้าน security (ไม่ส่ง password hash กลับไป DB โดยไม่จำเป็น)

---

### 1.3 Post-write reload ไม่มี Context

**ปัญหา:** หลัง Create/Update ใน `task.go`, `feeder.go`, `station.go`, `pea.go` จะ reload record พร้อม Preload แต่ไม่ใส่ `WithContext` ทำให้ถ้า client disconnect แล้ว query ยังรันต่อ

```go
// ปัญหา: ไม่มี WithContext
h.db.Preload("Team").Preload("JobType").First(&task, task.ID)

// แก้ไข: ใส่ WithContext
h.db.WithContext(c.Request.Context()).Preload("Team").Preload("JobType").First(&task, task.ID)
```

**ไฟล์:** `task.go` (Create, Update), `feeder.go` (Update), `station.go` (Create, Update), `pea.go` (Create, Update)

---

### 1.4 RecoveryMiddleware ลงทะเบียนหลัง routes

**ปัญหา:** `main.go` เพิ่ม `RecoveryMiddleware()` หลัง `SetupRouter()` return ทำให้ middleware ไม่ apply กับ route ที่ register แล้ว (Gin middleware ต้อง register ก่อน routes)

**แก้ไข:** ย้าย `RecoveryMiddleware()` ไปใน `SetupRouter()` ก่อน register routes หรือลบทิ้งเพราะ `gin.Default()` มี Recovery อยู่แล้ว (แต่ response format จะเป็น plain text ไม่ใช่ JSON)

**ไฟล์:** `main.go` line ~70, `internal/router/router.go`

---

## Priority 2: Medium Impact (ควรทำ)

### 2.1 เพิ่ม Request Timeout Middleware

**ปัญหา:** ไม่มี timeout per request ถ้า Neon ช้าหรือ cold start query อาจ hang ไม่มีกำหนด

```go
func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
        defer cancel()
        c.Request = c.Request.WithContext(ctx)
        c.Next()
    }
}
// ใช้: r.Use(TimeoutMiddleware(30 * time.Second))
```

---

### 2.2 เพิ่ม Gzip Compression

**ปัญหา:** Response ไม่มี compression task list ที่มี nested relations อาจมีขนาดใหญ่

```go
import "github.com/gin-contrib/gzip"
r.Use(gzip.Gzip(gzip.DefaultCompression))
```

**ผลลัพธ์:** ลดขนาด response 60-80% สำหรับ JSON, ลด bandwidth cost, เร็วขึ้นสำหรับ client

---

### 2.3 User List ไม่มี Pagination

**ปัญหา:** `GET /v1/users` return ทุก user ในระบบ ไม่มี pagination ถ้า user เยอะจะช้า

**แก้ไข:** เพิ่ม pagination เหมือน task list (page/limit query params)

---

### 2.4 ลบ Reload ที่ไม่จำเป็นใน UserHandler.Update

**ปัญหา:** หลัง update user จะ reload ด้วย `Preload("Team")` แต่ response DTO ไม่ได้ใช้ Team object

**แก้ไข:** ลบ reload ออก หรือถ้าต้องการ return Team ให้เพิ่มใน DTO

**ไฟล์:** `internal/handlers/v1/user.go`

---

### 2.5 Stats endpoint bug: `activeTeams` ไม่ filter ตาม date range

**ปัญหา:** COUNT DISTINCT teamId ไม่ apply `baseScope` ทำให้ count teams ทั้งหมดตลอดไม่ว่า filter อะไร

**ไฟล์:** `internal/handlers/v1/dashboard.go` — ฟังก์ชัน `Stats()`

---

### 2.6 ListByTeam COUNT ไม่ apply filters

**ปัญหา:** Total count ใน pagination meta เป็นจำนวน task ทั้งหมด ไม่ว่าจะ filter อะไร

**ไฟล์:** `internal/handlers/v1/task.go` — ฟังก์ชัน `ListByTeam()`

---

### 2.7 Bcrypt cost ไม่ consistent

**ปัญหา:**
- `auth.go` (Register/Login) ใช้ `pkg/password` → bcrypt cost 12 (~400-600ms)
- `user.go` (ChangePassword) ใช้ `bcrypt.DefaultCost` → cost 10 (~100ms)

**แก้ไข:** ใช้ `pkg/password` ทุกที่ หรือพิจารณาลด cost เป็น 10 สำหรับ Cloud Run ที่ CPU อาจถูก throttle

---

## Priority 3: Architecture / Nice-to-Have

### 3.1 Cloud Run Cold Start Optimization

| ปัญหา | แนวทาง |
|--------|---------|
| Scale to zero + Neon suspend = 2-4s first request | ตั้ง `min-instances: 1` ใน Cloud Run |
| ไม่มี retry ตอน connect DB | เพิ่ม retry with backoff ใน `database.Connect()` |
| `gin.Default()` มี Logger middleware log ทุก request | ใน production ใช้ `gin.New()` + เฉพาะ Recovery |

### 3.2 Rate Limiting

**ปัญหา:** Public endpoints ไม่มี rate limit ถ้า CDN ถูก bypass จะกระทบ Cloud Run และ Neon

**แนวทาง:** เพิ่ม rate limiting middleware หรือใช้ Cloud Armor / Cloudflare WAF rules

### 3.3 ลบ Dead Code

- `UploadDirect()` ใน `upload.go` — ไม่ได้ register ใน router
- `StringArray` custom type — ใช้ `lib/pq.StringArray` แทนได้
- `replaceAll()`, `indexOfString()` helpers — ใช้ `strings.ReplaceAll`, `strings.Index` จาก stdlib

### 3.4 ListByFilter / ListByTeam return `map`

**ปัญหา:** JSON object key order ไม่ guaranteed ถ้า 2 team ชื่อซ้ำจะ collision

**แก้ไข:** Return เป็น ordered slice แทน map

---

## สรุป Action Items

| # | รายการ | Impact | Effort | ไฟล์หลัก |
|---|--------|--------|--------|----------|
| 1 | Stats/Dashboard concurrent queries | สูงมาก | กลาง | `dashboard.go` |
| 2 | แก้ `Save()` → `.Update()` | สูง | ต่ำ | `auth.go`, `task.go`, `job_detail.go` |
| 3 | เพิ่ม `WithContext` ใน reload | สูง | ต่ำ | `task.go`, `feeder.go`, `station.go`, `pea.go` |
| 4 | แก้ RecoveryMiddleware order | สูง | ต่ำ | `main.go`, `router.go` |
| 5 | เพิ่ม Request Timeout | กลาง | ต่ำ | middleware ใหม่ |
| 6 | เพิ่ม Gzip Compression | กลาง | ต่ำ | `router.go` |
| 7 | User List pagination | กลาง | ต่ำ | `user.go` |
| 8 | ลบ unnecessary reload | กลาง | ต่ำ | `user.go` |
| 9 | แก้ Stats activeTeams bug | กลาง | ต่ำ | `dashboard.go` |
| 10 | แก้ ListByTeam COUNT bug | กลาง | ต่ำ | `task.go` |
| 11 | Bcrypt cost consistency | กลาง | ต่ำ | `user.go` |
| 12 | Cloud Run min-instances | กลาง | ต่ำ | Cloud Run config |
| 13 | Rate limiting | ต่ำ | กลาง | middleware ใหม่ |
| 14 | ลบ dead code | ต่ำ | ต่ำ | `upload.go`, `models.go` |
| 15 | map → slice response | ต่ำ | กลาง | `task.go` |

---

## Performance Estimates (Before vs After)

| Scenario | ปัจจุบัน | หลังแก้ไข |
|----------|---------|-----------|
| Stats endpoint (warm) | ~150-200ms | ~30-50ms (concurrent queries) |
| Stats endpoint (cold Neon) | ~3-4s | ~2-3s (concurrent + min-instances) |
| Login request | ~500-700ms (bcrypt + full Save) | ~400-600ms (bcrypt + targeted Update) |
| Task Create | ~50-80ms (7 queries) | ~30-50ms (+ WithContext safety) |
| Task List (50 items) | ~20-40ms | ~15-25ms (+ gzip = smaller payload) |
| Response size (50 tasks) | ~50-100KB | ~10-20KB (with gzip) |
