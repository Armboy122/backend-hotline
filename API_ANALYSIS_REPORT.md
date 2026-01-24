# รายงานวิเคราะห์ความสอดคล้องระหว่าง Backend API และ Frontend Hotline App

## สรุปภาพรวม

### สถานะปัจจุบัน
| รายการ | Frontend (hotlines3) | Backend Go (claude.md) |
|--------|---------------------|------------------------|
| Database | PostgreSQL + Prisma ORM | PostgreSQL + GORM |
| Data Access | Server Actions (เรียก Prisma ตรง) | REST API (ยังไม่ได้ implement) |
| File Storage | Cloudflare R2 (direct upload) | Cloudflare R2 (presigned URL) |
| Auth | ไม่มี | ไม่มี (แนะนำเพิ่ม) |

### ข้อสังเกตสำคัญ
**Frontend ปัจจุบันใช้ Server Actions เรียก Prisma โดยตรง** ไม่ได้เรียก Backend Go API
หากต้องการให้ Frontend เรียก Backend Go API แทน ต้องปรับ Frontend ให้เรียก REST API

---

## 1. ตารางเปรียบเทียบ API Endpoints

### 1.1 Master Data APIs

| Entity | Frontend ใช้งาน | Backend claude.md | สถานะ |
|--------|----------------|-------------------|-------|
| **Operation Centers** | GET (list) | CRUD 5 APIs | ครบถ้วน |
| **PEAs** | GET (list) | CRUD 6 APIs (+ bulk) | ครบถ้วน |
| **Stations** | GET (list) | CRUD 5 APIs | ครบถ้วน |
| **Feeders** | GET (list) | CRUD 5 APIs | ครบถ้วน |
| **Job Types** | GET (list) | CRUD 5 APIs | ครบถ้วน |
| **Job Details** | GET (list) | CRUD 5 APIs | ครบถ้วน |
| **Teams** | GET (list), CRUD | CRUD 5 APIs | ครบถ้วน |

### 1.2 Task Daily APIs (Core)

| Function | Frontend ใช้งาน | Backend claude.md | สถานะ |
|----------|----------------|-------------------|-------|
| List tasks | GET /api/tasks | GET /api/tasks | ครบถ้วน |
| Get task by ID | GET /api/tasks/:id | GET /api/tasks/:id | ครบถ้วน |
| Create task | POST /api/tasks | POST /api/tasks | ครบถ้วน |
| Update task | PUT /api/tasks/:id | PUT /api/tasks/:id | ครบถ้วน |
| Delete task | DELETE /api/tasks/:id | DELETE /api/tasks/:id | ครบถ้วน |
| Group by team | getTaskDailiesByTeam() | GET /api/tasks/by-team | ครบถ้วน |
| **Filter by year/month/team** | getTaskDailiesByFilter() | Query params | ครบถ้วน |

### 1.3 Dashboard APIs

| Function | Frontend ใช้งาน | Backend claude.md | สถานะ |
|----------|----------------|-------------------|-------|
| Summary | getDashboardSummary() | GET /api/dashboard/summary | ครบถ้วน |
| Top Jobs | getTopJobDetails() | GET /api/dashboard/top-jobs | ครบถ้วน |
| Top Feeders | getTopFeeders() | GET /api/dashboard/top-feeders | ครบถ้วน |
| Feeder Matrix | getFeederJobMatrix() | GET /api/dashboard/feeder-matrix | ครบถ้วน |
| Stats (Charts) | getDashboardStats() | GET /api/dashboard/stats | ครบถ้วน |

### 1.4 Upload APIs

| Function | Frontend ใช้งาน | Backend claude.md | ความแตกต่าง |
|----------|----------------|-------------------|-------------|
| Upload | uploadImage() - Direct R2 | POST /api/upload - Presigned URL | **แตกต่าง** |
| Delete | deleteImage() - Direct R2 | DELETE /api/upload/:key | **แตกต่าง** |

---

## 2. สิ่งที่ต้องเพิ่มใน Backend API

### 2.1 API ที่ขาดหายไป

#### 2.1.1 Tasks Filter by Year/Month API
```
GET /api/tasks/by-filter?year=2024&month=1&teamId=1
```
Frontend ใช้ `getTaskDailiesByFilter()` ที่ return ข้อมูล grouped by team

**Response format ที่ต้องการ:**
```json
{
  "success": true,
  "data": {
    "ทีม A": {
      "team": { "id": "1", "name": "ทีม A" },
      "tasks": [...]
    },
    "ทีม B": {
      "team": { "id": "2", "name": "ทีม B" },
      "tasks": [...]
    }
  }
}
```

#### 2.1.2 Dashboard Stats Response Format
Backend claude.md กำหนด response format ต่างจาก Frontend เล็กน้อย

**Frontend ต้องการ:**
```json
{
  "summary": {
    "totalTasks": 500,
    "activeTeams": 5,        // ใช้ชื่อ activeTeams
    "topJobType": "งานซ่อม", // เป็น string ไม่ใช่ object
    "topFeeder": "LPB-01"    // เป็น string ไม่ใช่ object
  },
  "charts": {
    "tasksByFeeder": [{"name": "LPB-01", "value": 30}],  // ใช้ name/value
    "tasksByJobType": [...],
    "tasksByTeam": [...],
    "tasksByDate": [{"date": "2024-01-01", "count": 5}]  // ใช้ date/count
  }
}
```

**Backend claude.md กำหนด:**
```json
{
  "summary": {
    "totalTasks": 500,
    "totalTeams": 5,         // ชื่อต่างกัน
    "topJobDetail": {...},   // เป็น object
    "topFeeder": {...}       // เป็น object
  },
  "charts": {
    "byFeeder": [{"label": "...", "value": 30}],  // ใช้ label แทน name
    "byJobType": [...],
    "byTeam": [...],
    "byDate": [{"label": "...", "value": 5}]
  }
}
```

**แนะนำ:** ปรับ Backend ให้ตรงกับ Frontend หรือปรับ Frontend ให้รับ format ใหม่

### 2.2 Upload API - ความแตกต่างที่สำคัญ

#### Frontend ปัจจุบัน (Direct Upload)
```typescript
// 1. รับ File จาก form
// 2. แปลงเป็น Buffer
// 3. Upload ตรงไป R2 ด้วย AWS SDK
// 4. Return URL
```

#### Backend claude.md (Presigned URL)
```
POST /api/upload
Request: { fileName, fileType }
Response: { uploadUrl, fileUrl, fileKey }

// Client ต้อง:
// 1. เรียก API ขอ presigned URL
// 2. Upload ไปที่ presigned URL ด้วย PUT
// 3. ใช้ fileUrl เก็บใน database
```

**แนะนำ:**
- วิธี Presigned URL ปลอดภัยกว่า (ไม่ต้องเก็บ credentials ใน frontend)
- ถ้าเปลี่ยนไปใช้ Backend API ต้องปรับ Frontend upload flow

---

## 3. สิ่งที่อาจลดหรือไม่จำเป็น

### 3.1 APIs ที่ Frontend ไม่ได้ใช้ (แต่ควรเก็บไว้)

| API | เหตุผลที่ควรเก็บ |
|-----|-----------------|
| POST /api/peas/bulk | สำหรับ Admin import ข้อมูล |
| PUT/DELETE Master Data | สำหรับ Admin Panel จัดการข้อมูล |

### 3.2 Fields ที่ Frontend ไม่ได้ใช้ใน Response

Frontend serialize TaskDaily ไม่ได้ใช้:
- `deletedAt` ในการแสดงผล (แต่ควรเก็บไว้สำหรับ soft delete)

---

## 4. ระบบ Login/Authentication สำหรับอนาคต

### 4.1 โครงสร้างที่แนะนำ

#### Model: User
```go
type User struct {
    ID           int64      `gorm:"primaryKey;autoIncrement" json:"id"`
    Username     string     `gorm:"uniqueIndex;not null" json:"username"`
    Email        string     `gorm:"uniqueIndex;not null" json:"email"`
    PasswordHash string     `gorm:"not null" json:"-"`
    FullName     string     `json:"fullName"`
    Role         string     `gorm:"default:'user'" json:"role"` // admin, supervisor, user
    TeamID       *int64     `json:"teamId"`
    IsActive     bool       `gorm:"default:true" json:"isActive"`
    LastLoginAt  *time.Time `json:"lastLoginAt"`
    CreatedAt    time.Time  `gorm:"autoCreateTime" json:"createdAt"`
    UpdatedAt    time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`
    DeletedAt    *time.Time `gorm:"index" json:"deletedAt,omitempty"`

    Team *Team `gorm:"foreignKey:TeamID" json:"team,omitempty"`
}
```

#### Model: RefreshToken (สำหรับ JWT)
```go
type RefreshToken struct {
    ID        int64     `gorm:"primaryKey;autoIncrement"`
    UserID    int64     `gorm:"not null"`
    Token     string    `gorm:"uniqueIndex;not null"`
    ExpiresAt time.Time `gorm:"not null"`
    CreatedAt time.Time `gorm:"autoCreateTime"`

    User User `gorm:"foreignKey:UserID"`
}
```

### 4.2 Auth APIs ที่ต้องเพิ่ม

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/auth/register` | สมัครสมาชิก (Admin only) |
| POST | `/api/auth/login` | เข้าสู่ระบบ |
| POST | `/api/auth/logout` | ออกจากระบบ |
| POST | `/api/auth/refresh` | Refresh token |
| GET | `/api/auth/me` | ข้อมูล user ปัจจุบัน |
| PUT | `/api/auth/change-password` | เปลี่ยนรหัสผ่าน |

### 4.3 Auth DTOs

```go
// Login
type LoginRequest struct {
    Username string `json:"username" binding:"required"`
    Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
    AccessToken  string       `json:"accessToken"`
    RefreshToken string       `json:"refreshToken"`
    ExpiresIn    int64        `json:"expiresIn"`
    User         UserResponse `json:"user"`
}

// Register
type RegisterRequest struct {
    Username string `json:"username" binding:"required,min=3,max=50"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=8"`
    FullName string `json:"fullName" binding:"required"`
    TeamID   *int64 `json:"teamId"`
    Role     string `json:"role" binding:"omitempty,oneof=admin supervisor user"`
}

// User Response
type UserResponse struct {
    ID        int64         `json:"id"`
    Username  string        `json:"username"`
    Email     string        `json:"email"`
    FullName  string        `json:"fullName"`
    Role      string        `json:"role"`
    TeamID    *int64        `json:"teamId"`
    Team      *TeamResponse `json:"team,omitempty"`
    IsActive  bool          `json:"isActive"`
    CreatedAt time.Time     `json:"createdAt"`
}
```

### 4.4 Role-Based Access Control (RBAC)

```go
// Roles
const (
    RoleAdmin      = "admin"      // จัดการทุกอย่าง
    RoleSupervisor = "supervisor" // ดู dashboard, จัดการทีมตัวเอง
    RoleUser       = "user"       // บันทึกงานประจำวัน
)

// Permissions Matrix
/*
| Resource          | Admin | Supervisor | User |
|-------------------|-------|------------|------|
| Master Data CRUD  | CRUD  | R          | R    |
| Task Daily        | CRUD  | CRUD (team)| CR (self) |
| Dashboard         | Full  | Team-only  | -    |
| Users             | CRUD  | R (team)   | -    |
| Upload            | Yes   | Yes        | Yes  |
*/
```

### 4.5 JWT Middleware

```go
// internal/middleware/auth.go
package middleware

func JWTAuth() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. ดึง token จาก Authorization header
        // 2. Validate JWT
        // 3. ดึง user จาก token claims
        // 4. Set user ใน context
        // 5. Call next handler
    }
}

func RequireRole(roles ...string) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. ดึง user จาก context
        // 2. ตรวจสอบ role
        // 3. Allow/Deny
    }
}
```

### 4.6 Protected Routes

```go
// internal/router/router.go

// Public routes
public := r.Group("/api")
{
    public.POST("/auth/login", authCtrl.Login)
    public.POST("/auth/refresh", authCtrl.Refresh)
}

// Protected routes (require login)
protected := r.Group("/api")
protected.Use(middleware.JWTAuth())
{
    protected.GET("/auth/me", authCtrl.Me)
    protected.POST("/auth/logout", authCtrl.Logout)
    protected.PUT("/auth/change-password", authCtrl.ChangePassword)

    // ... other protected routes
}

// Admin only routes
admin := r.Group("/api/admin")
admin.Use(middleware.JWTAuth(), middleware.RequireRole("admin"))
{
    admin.POST("/auth/register", authCtrl.Register)
    // ... admin routes
}
```

---

## 5. Module Structure สำหรับ Auth

```
internal/modules/auth/
├── model.go           # User, RefreshToken models
├── dto.go             # Request/Response DTOs
├── repository.go      # User & Token database operations
├── service.go         # Auth business logic (login, register, JWT)
├── controller.go      # HTTP handlers
└── routes.go          # Route registration

pkg/
├── jwt/
│   └── jwt.go         # JWT utilities (generate, validate, refresh)
└── hash/
    └── bcrypt.go      # Password hashing utilities
```

---

## 6. Environment Variables ที่ต้องเพิ่ม (Auth)

```env
# JWT
JWT_SECRET=your-super-secret-key-min-32-chars
JWT_ACCESS_TOKEN_EXPIRY=15m
JWT_REFRESH_TOKEN_EXPIRY=7d

# (Optional) OAuth
GOOGLE_CLIENT_ID=
GOOGLE_CLIENT_SECRET=
```

---

## 7. สรุปสิ่งที่ต้องทำ

### ระดับความสำคัญ: สูง (ทำทันที)

1. **ปรับ Response Format ของ Dashboard Stats**
   - `totalTeams` -> `activeTeams`
   - Charts ใช้ `name/value` แทน `label/value`
   - `tasksByDate` ใช้ `date/count`

2. **เพิ่ม API: GET /api/tasks/by-filter**
   - รับ query params: year, month, teamId
   - Return grouped by team

3. **ตัดสินใจเรื่อง Upload Strategy**
   - Option A: ใช้ Presigned URL (ปลอดภัยกว่า) - ต้องปรับ Frontend
   - Option B: ใช้ Direct Upload เหมือนเดิม - ต้องปรับ Backend

### ระดับความสำคัญ: กลาง (ทำก่อน production)

4. **เพิ่มระบบ Auth**
   - User model + migration
   - Auth module (login, register, JWT)
   - Middleware (JWTAuth, RequireRole)
   - Protected routes

5. **ปรับ TaskDaily Response**
   - Include station.operationCenter ใน feeder relation

### ระดับความสำคัญ: ต่ำ (Nice to have)

6. **Pagination**
   - เพิ่ม pagination สำหรับ list APIs ที่มีข้อมูลมาก

7. **Rate Limiting**
   - ป้องกัน API abuse

8. **Logging & Monitoring**
   - Request/Response logging
   - Error tracking

---

## 8. API Checklist

### Master Data APIs
- [x] Operation Centers - CRUD 5 APIs
- [x] PEAs - CRUD 6 APIs (+ bulk)
- [x] Stations - CRUD 5 APIs
- [x] Feeders - CRUD 5 APIs
- [x] Job Types - CRUD 5 APIs
- [x] Job Details - CRUD 5 APIs
- [x] Teams - CRUD 5 APIs

### Task Daily APIs
- [x] GET /api/tasks (with filters)
- [x] GET /api/tasks/:id
- [x] POST /api/tasks
- [x] PUT /api/tasks/:id
- [x] DELETE /api/tasks/:id
- [x] GET /api/tasks/by-team
- [ ] **GET /api/tasks/by-filter** (ต้องเพิ่ม - grouped response)

### Dashboard APIs
- [x] GET /api/dashboard/summary
- [x] GET /api/dashboard/top-jobs
- [x] GET /api/dashboard/top-feeders
- [x] GET /api/dashboard/feeder-matrix
- [ ] **GET /api/dashboard/stats** (ต้องปรับ response format)

### Upload APIs
- [ ] **POST /api/upload** (ต้องตัดสินใจ strategy)
- [ ] **DELETE /api/upload/:key**

### Auth APIs (ใหม่)
- [ ] POST /api/auth/register
- [ ] POST /api/auth/login
- [ ] POST /api/auth/logout
- [ ] POST /api/auth/refresh
- [ ] GET /api/auth/me
- [ ] PUT /api/auth/change-password

---

**สร้างโดย:** Claude Code Analysis
**วันที่:** 21 มกราคม 2026
