# Hotline Backend API - Go (Gin + GORM)

## Project Overview

ระบบจัดการงานซ่อมบำรุงโครงข่ายไฟฟ้า (Hotline System) สำหรับการไฟฟ้าส่วนภูมิภาค
- **Frontend**: Next.js 15 (อยู่ที่ `/root/project/hotlines3`)
- **Backend**: Go + Gin + GORM (โปรเจคนี้)
- **Database**: PostgreSQL (Neon)
- **File Storage**: Cloudflare R2 (S3-compatible)

---

## Database Schema (9 Models)

### 1. OperationCenter (จุดรวมงาน)
```go
type OperationCenter struct {
    ID       int64     `gorm:"primaryKey;autoIncrement" json:"id"`
    Name     string    `gorm:"not null" json:"name"`
    Peas     []Pea     `gorm:"foreignKey:OperationID" json:"peas,omitempty"`
    Stations []Station `gorm:"foreignKey:OperationID" json:"stations,omitempty"`
}
```

### 2. Pea (การไฟฟ้า - Provincial Electricity Authority)
```go
type Pea struct {
    ID              int64            `gorm:"primaryKey;autoIncrement" json:"id"`
    Shortname       string           `gorm:"not null" json:"shortname"`
    Fullname        string           `gorm:"not null" json:"fullname"`
    OperationID     int64            `gorm:"not null" json:"operationId"`
    OperationCenter *OperationCenter `gorm:"foreignKey:OperationID" json:"operationCenter,omitempty"`
}
```

### 3. Station (สถานีไฟฟ้า)
```go
type Station struct {
    ID              int64            `gorm:"primaryKey;autoIncrement" json:"id"`
    Name            string           `gorm:"not null" json:"name"`
    CodeName        string           `gorm:"uniqueIndex;not null" json:"codeName"`
    OperationID     int64            `gorm:"not null" json:"operationId"`
    OperationCenter *OperationCenter `gorm:"foreignKey:OperationID" json:"operationCenter,omitempty"`
    Feeders         []Feeder         `gorm:"foreignKey:StationID" json:"feeders,omitempty"`
}
```

### 4. Feeder (ฟีดเดอร์/สายส่ง)
```go
type Feeder struct {
    ID        int64       `gorm:"primaryKey;autoIncrement" json:"id"`
    Code      string      `gorm:"uniqueIndex;not null" json:"code"`
    StationID int64       `gorm:"index;not null" json:"stationId"`
    Station   *Station    `gorm:"foreignKey:StationID" json:"station,omitempty"`
    Tasks     []TaskDaily `gorm:"foreignKey:FeederID" json:"tasks,omitempty"`
}
```

### 5. JobType (ประเภทงาน)
```go
type JobType struct {
    ID         int64       `gorm:"primaryKey;autoIncrement" json:"id"`
    Name       string      `gorm:"uniqueIndex;not null" json:"name"`
    Tasks      []TaskDaily `gorm:"foreignKey:JobTypeID" json:"tasks,omitempty"`
    JobDetails []JobDetail `gorm:"foreignKey:JobTypeID" json:"jobDetails,omitempty"`
}
```

### 6. JobDetail (รายละเอียดงาน)
```go
type JobDetail struct {
    ID        int64      `gorm:"primaryKey;autoIncrement" json:"id"`
    Name      string     `gorm:"uniqueIndex;not null" json:"name"`
    JobTypeID *int64     `gorm:"index" json:"jobTypeId"`
    JobType   *JobType   `gorm:"foreignKey:JobTypeID" json:"jobType,omitempty"`
    CreatedAt time.Time  `gorm:"autoCreateTime" json:"createdAt"`
    UpdatedAt time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`
    DeletedAt *time.Time `gorm:"index" json:"deletedAt,omitempty"`
}
```

### 7. Team (ทีมงาน)
```go
type Team struct {
    ID    int64       `gorm:"primaryKey;autoIncrement" json:"id"`
    Name  string      `gorm:"not null" json:"name"`
    Tasks []TaskDaily `gorm:"foreignKey:TeamID" json:"tasks,omitempty"`
}
```

### 8. TaskDaily (รายงานงานประจำวัน) - Core Model
```go
type TaskDaily struct {
    ID          int64           `gorm:"primaryKey;autoIncrement" json:"id"`
    WorkDate    time.Time       `gorm:"type:date;index;not null" json:"workDate"`
    TeamID      int64           `gorm:"not null" json:"teamId"`
    JobTypeID   int64           `gorm:"not null" json:"jobTypeId"`
    JobDetailID int64           `gorm:"not null" json:"jobDetailId"`
    FeederID    *int64          `gorm:"index" json:"feederId"`
    NumPole     *string         `json:"numPole"`
    DeviceCode  *string         `json:"deviceCode"`
    Detail      *string         `json:"detail"`
    UrlsBefore  pq.StringArray  `gorm:"type:text[]" json:"urlsBefore"`
    UrlsAfter   pq.StringArray  `gorm:"type:text[]" json:"urlsAfter"`
    Latitude    *float64        `gorm:"type:decimal(9,6)" json:"latitude"`
    Longitude   *float64        `gorm:"type:decimal(9,6)" json:"longitude"`
    CreatedAt   time.Time       `gorm:"autoCreateTime" json:"createdAt"`
    UpdatedAt   time.Time       `gorm:"autoUpdateTime" json:"updatedAt"`
    DeletedAt   *time.Time      `gorm:"index" json:"deletedAt,omitempty"`

    // Relations
    Team      *Team      `gorm:"foreignKey:TeamID" json:"team,omitempty"`
    JobType   *JobType   `gorm:"foreignKey:JobTypeID" json:"jobType,omitempty"`
    JobDetail *JobDetail `gorm:"foreignKey:JobDetailID" json:"jobDetail,omitempty"`
    Feeder    *Feeder    `gorm:"foreignKey:FeederID" json:"feeder,omitempty"`
}

// Indexes: workDate, (jobTypeId, jobDetailId), feederId, (latitude, longitude)
```

---

## API Endpoints (Total: 45+ endpoints)

### Master Data APIs

#### Operation Centers
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/operation-centers` | List all operation centers |
| GET | `/api/operation-centers/:id` | Get operation center by ID |
| POST | `/api/operation-centers` | Create operation center |
| PUT | `/api/operation-centers/:id` | Update operation center |
| DELETE | `/api/operation-centers/:id` | Delete operation center |

#### PEAs (การไฟฟ้า)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/peas` | List all PEAs |
| GET | `/api/peas/:id` | Get PEA by ID |
| POST | `/api/peas` | Create PEA |
| POST | `/api/peas/bulk` | Bulk create PEAs |
| PUT | `/api/peas/:id` | Update PEA |
| DELETE | `/api/peas/:id` | Delete PEA |

#### Stations (สถานี)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/stations` | List all stations with operation center |
| GET | `/api/stations/:id` | Get station by ID |
| POST | `/api/stations` | Create station |
| PUT | `/api/stations/:id` | Update station |
| DELETE | `/api/stations/:id` | Delete station |

#### Feeders (ฟีดเดอร์)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/feeders` | List all feeders with station & operation center |
| GET | `/api/feeders/:id` | Get feeder by ID |
| POST | `/api/feeders` | Create feeder |
| PUT | `/api/feeders/:id` | Update feeder |
| DELETE | `/api/feeders/:id` | Delete feeder |

#### Job Types (ประเภทงาน)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/job-types` | List all job types with task count |
| GET | `/api/job-types/:id` | Get job type by ID |
| POST | `/api/job-types` | Create job type |
| PUT | `/api/job-types/:id` | Update job type |
| DELETE | `/api/job-types/:id` | Delete job type |

#### Job Details (รายละเอียดงาน)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/job-details` | List all job details with task count |
| GET | `/api/job-details/:id` | Get job detail by ID |
| POST | `/api/job-details` | Create job detail |
| PUT | `/api/job-details/:id` | Update job detail |
| DELETE | `/api/job-details/:id` | Soft delete job detail |

#### Teams (ทีมงาน)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/teams` | List all teams |
| GET | `/api/teams/:id` | Get team by ID |
| POST | `/api/teams` | Create team |
| PUT | `/api/teams/:id` | Update team |
| DELETE | `/api/teams/:id` | Delete team |

---

### Task Daily APIs (Core Business)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/tasks` | List tasks with filters |
| GET | `/api/tasks/:id` | Get task by ID with all relations |
| POST | `/api/tasks` | Create task |
| PUT | `/api/tasks/:id` | Update task |
| DELETE | `/api/tasks/:id` | Soft delete task |
| GET | `/api/tasks/by-team` | Get all tasks grouped by team |

**Query Parameters for GET `/api/tasks`:**
```
?year=2024          - Filter by year
&month=1            - Filter by month (1-12)
&teamId=1           - Filter by team
&jobTypeId=1        - Filter by job type
&feederId=1         - Filter by feeder
&workDate=2024-01-15 - Filter by specific date
```

**Request Body for POST/PUT:**
```json
{
  "workDate": "2024-01-15",
  "teamId": 1,
  "jobTypeId": 1,
  "jobDetailId": 1,
  "feederId": 1,
  "numPole": "A001",
  "deviceCode": "SW-001",
  "detail": "รายละเอียดงาน",
  "urlsBefore": ["https://photo.example.com/before1.jpg"],
  "urlsAfter": ["https://photo.example.com/after1.jpg"],
  "latitude": 13.756331,
  "longitude": 100.501762
}
```

**Validation Rules:**
- `latitude`: must be between -90 and 90
- `longitude`: must be between -180 and 180
- Both coordinates must be provided together or neither

---

### Dashboard/Analytics APIs

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/dashboard/summary` | Dashboard summary statistics |
| GET | `/api/dashboard/stats` | Comprehensive dashboard stats |
| GET | `/api/dashboard/top-jobs` | Top 10 most performed jobs |
| GET | `/api/dashboard/top-feeders` | Top 10 most used feeders |
| GET | `/api/dashboard/feeder-matrix` | Job breakdown by feeder |

**GET `/api/dashboard/summary`**
```
Query: ?year=2024&month=1&teamId=1&jobTypeId=1

Response:
{
  "totalTasks": 150,
  "totalJobTypes": 5,
  "totalFeeders": 25,
  "topTeam": { "id": 1, "name": "ทีม A", "taskCount": 50 }
}
```

**GET `/api/dashboard/top-jobs`**
```
Query: ?year=2024&limit=10&month=1&teamId=1&jobTypeId=1

Response:
[
  {
    "jobDetailId": 1,
    "jobDetailName": "ตัดต้นไม้",
    "jobTypeName": "งานบำรุงรักษา",
    "taskCount": 45
  }
]
```

**GET `/api/dashboard/top-feeders`**
```
Query: ?year=2024&limit=10&month=1&teamId=1&jobTypeId=1

Response:
[
  {
    "feederId": 1,
    "feederCode": "LPB-01",
    "stationName": "สถานี A",
    "taskCount": 30
  }
]
```

**GET `/api/dashboard/feeder-matrix`**
```
Query: ?feederId=1&year=2024&month=1&teamId=1&jobTypeId=1

Response:
{
  "feederId": 1,
  "feederCode": "LPB-01",
  "jobDetails": [
    { "jobDetailId": 1, "name": "ตัดต้นไม้", "count": 15 },
    { "jobDetailId": 2, "name": "ซ่อมสาย", "count": 10 }
  ],
  "totalTasks": 25
}
```

**GET `/api/dashboard/stats`**
```
Query: ?startDate=2024-01-01&endDate=2024-12-31&teamId=1&feederId=1

Response:
{
  "summary": {
    "totalTasks": 500,
    "totalTeams": 5,
    "topJobDetail": { "name": "ตัดต้นไม้", "count": 100 },
    "topFeeder": { "code": "LPB-01", "count": 80 }
  },
  "charts": {
    "byFeeder": [...],
    "byJobType": [...],
    "byTeam": [...],
    "byDate": [...]
  }
}
```

---

### File Upload API

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/upload` | Get presigned URL for S3 upload |
| DELETE | `/api/upload/:key` | Delete file from S3 |

**POST `/api/upload`**
```json
Request:
{
  "fileName": "image.jpg",
  "fileType": "image/jpeg"
}

Response:
{
  "uploadUrl": "https://...",
  "fileUrl": "https://photo.example.com/images/xxx.jpg",
  "fileKey": "images/xxx.jpg"
}
```

**Allowed file types:** `image/jpeg`, `image/jpg`, `image/png`, `image/webp`, `image/gif`

---

## Project Structure (Module-Based Architecture)

```
backend-hotlines3/
├── cmd/
│   └── server/
│       └── main.go                     # Entry point
│
├── internal/
│   ├── config/
│   │   └── config.go                   # Environment config
│   ├── database/
│   │   └── database.go                 # GORM connection
│   ├── common/
│   │   ├── response.go                 # Standard response structs
│   │   ├── errors.go                   # Custom error types
│   │   └── repository.go               # Base repository interface
│   ├── middleware/
│   │   ├── cors.go                     # CORS configuration
│   │   ├── error.go                    # Error handler
│   │   ├── logger.go                   # Request logging
│   │   └── recovery.go                 # Panic recovery
│   ├── router/
│   │   └── router.go                   # Main router setup
│   │
│   └── modules/
│       ├── health/                     # Health check module
│       │   ├── controller.go           # /health, /ready endpoints
│       │   └── routes.go
│       │
│       ├── operation_center/
│       │   ├── model.go                # OperationCenter model
│       │   ├── dto.go                  # Request/Response DTOs
│       │   ├── repository.go           # Database operations
│       │   ├── service.go              # Business logic
│       │   ├── controller.go           # HTTP handlers
│       │   └── routes.go               # Module routes
│       │
│       ├── pea/
│       │   ├── model.go
│       │   ├── dto.go
│       │   ├── repository.go
│       │   ├── service.go
│       │   ├── controller.go
│       │   └── routes.go
│       │
│       ├── station/
│       │   ├── model.go
│       │   ├── dto.go
│       │   ├── repository.go
│       │   ├── service.go
│       │   ├── controller.go
│       │   └── routes.go
│       │
│       ├── feeder/
│       │   ├── model.go
│       │   ├── dto.go
│       │   ├── repository.go
│       │   ├── service.go
│       │   ├── controller.go
│       │   └── routes.go
│       │
│       ├── job_type/
│       │   ├── model.go
│       │   ├── dto.go
│       │   ├── repository.go
│       │   ├── service.go
│       │   ├── controller.go
│       │   └── routes.go
│       │
│       ├── job_detail/
│       │   ├── model.go
│       │   ├── dto.go
│       │   ├── repository.go
│       │   ├── service.go
│       │   ├── controller.go
│       │   └── routes.go
│       │
│       ├── team/
│       │   ├── model.go
│       │   ├── dto.go
│       │   ├── repository.go
│       │   ├── service.go
│       │   ├── controller.go
│       │   └── routes.go
│       │
│       ├── task/
│       │   ├── model.go
│       │   ├── dto.go
│       │   ├── repository.go
│       │   ├── service.go
│       │   ├── controller.go
│       │   └── routes.go
│       │
│       ├── dashboard/
│       │   ├── dto.go
│       │   ├── repository.go
│       │   ├── service.go
│       │   ├── controller.go
│       │   └── routes.go
│       │
│       └── upload/
│           ├── dto.go
│           ├── service.go
│           ├── controller.go
│           └── routes.go
│
├── pkg/
│   ├── s3/
│   │   └── r2.go                       # Cloudflare R2 client
│   ├── validator/
│   │   └── validator.go                # Custom validators
│   └── logger/
│       └── logger.go                   # Structured logging
│
├── scripts/
│   ├── migrate.sh                      # Database migration script
│   └── seed.sh                         # Seed data script
│
├── .dockerignore                       # Docker ignore file
├── .env.example                        # Environment template
├── .env.production.example             # Production env template
├── Dockerfile                          # Multi-stage Docker build
├── docker-compose.yml                  # Development compose
├── docker-compose.prod.yml             # Production compose
├── Makefile                            # Build & deploy commands
├── go.mod
├── go.sum
└── README.md
```

### .dockerignore

```
# Git
.git
.gitignore

# IDE
.idea
.vscode
*.swp
*.swo

# Binaries
bin/
*.exe
*.dll
*.so
*.dylib

# Test files
*_test.go
coverage.out

# Local env files
.env
.env.local
.env.*.local

# Documentation
*.md
!README.md
docs/

# Misc
.DS_Store
Thumbs.db
tmp/
```

### Module Structure Pattern

แต่ละ module จะมีโครงสร้างดังนี้:

```
module_name/
├── model.go        # GORM model struct
├── dto.go          # Request/Response DTOs
├── repository.go   # Database operations (interface + implementation)
├── service.go      # Business logic (interface + implementation)
├── controller.go   # HTTP handlers
└── routes.go       # Route registration function
```

### Module File Templates

**model.go** - GORM model
```go
package modulename

type ModuleName struct {
    ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
    Name      string    `gorm:"not null" json:"name"`
    CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
    UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}
```

**dto.go** - Data Transfer Objects
```go
package modulename

// Request DTOs
type CreateRequest struct {
    Name string `json:"name" binding:"required"`
}

type UpdateRequest struct {
    Name string `json:"name"`
}

// Response DTOs (if different from model)
type Response struct {
    ID   int64  `json:"id"`
    Name string `json:"name"`
}
```

**repository.go** - Database Layer
```go
package modulename

type Repository interface {
    FindAll() ([]ModuleName, error)
    FindByID(id int64) (*ModuleName, error)
    Create(entity *ModuleName) error
    Update(entity *ModuleName) error
    Delete(id int64) error
}

type repository struct {
    db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
    return &repository{db: db}
}
```

**service.go** - Business Logic Layer
```go
package modulename

type Service interface {
    GetAll() ([]ModuleName, error)
    GetByID(id int64) (*ModuleName, error)
    Create(req *CreateRequest) (*ModuleName, error)
    Update(id int64, req *UpdateRequest) (*ModuleName, error)
    Delete(id int64) error
}

type service struct {
    repo Repository
}

func NewService(repo Repository) Service {
    return &service{repo: repo}
}
```

**controller.go** - HTTP Handlers
```go
package modulename

type Controller struct {
    service Service
}

func NewController(service Service) *Controller {
    return &Controller{service: service}
}

func (c *Controller) GetAll(ctx *gin.Context) {
    // handle request
}
```

**routes.go** - Route Registration
```go
package modulename

func RegisterRoutes(router *gin.RouterGroup, db *gorm.DB) {
    repo := NewRepository(db)
    svc := NewService(repo)
    ctrl := NewController(svc)

    group := router.Group("/module-names")
    {
        group.GET("", ctrl.GetAll)
        group.GET("/:id", ctrl.GetByID)
        group.POST("", ctrl.Create)
        group.PUT("/:id", ctrl.Update)
        group.DELETE("/:id", ctrl.Delete)
    }
}
```

### Main Router Setup

```go
// internal/router/router.go
package router

import (
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"

    "backend-hotlines3/internal/modules/operation_center"
    "backend-hotlines3/internal/modules/pea"
    "backend-hotlines3/internal/modules/station"
    // ... other modules
)

func SetupRouter(db *gorm.DB) *gin.Engine {
    r := gin.Default()

    api := r.Group("/api")
    {
        operation_center.RegisterRoutes(api, db)
        pea.RegisterRoutes(api, db)
        station.RegisterRoutes(api, db)
        feeder.RegisterRoutes(api, db)
        job_type.RegisterRoutes(api, db)
        job_detail.RegisterRoutes(api, db)
        team.RegisterRoutes(api, db)
        task.RegisterRoutes(api, db)
        dashboard.RegisterRoutes(api, db)
        upload.RegisterRoutes(api, db)
    }

    return r
}
```

### Modules Summary (10 modules)

| Module | Endpoints | Description |
|--------|-----------|-------------|
| `operation_center` | CRUD (5) | จุดรวมงาน |
| `pea` | CRUD + Bulk (6) | การไฟฟ้า |
| `station` | CRUD (5) | สถานีไฟฟ้า |
| `feeder` | CRUD (5) | ฟีดเดอร์ |
| `job_type` | CRUD (5) | ประเภทงาน |
| `job_detail` | CRUD (5) | รายละเอียดงาน |
| `team` | CRUD (5) | ทีมงาน |
| `task` | CRUD + Special (6) | งานประจำวัน (Core) |
| `dashboard` | Read-only (5) | สถิติและรายงาน |
| `upload` | Create/Delete (2) | File upload (R2) |

---

## Module API Specifications (รายละเอียดแต่ละ Module)

---

### 1. Module: `operation_center` (5 APIs)

**Path:** `internal/modules/operation_center/`

#### APIs
| # | Method | Endpoint | Description |
|---|--------|----------|-------------|
| 1 | GET | `/api/operation-centers` | List all |
| 2 | GET | `/api/operation-centers/:id` | Get by ID |
| 3 | POST | `/api/operation-centers` | Create |
| 4 | PUT | `/api/operation-centers/:id` | Update |
| 5 | DELETE | `/api/operation-centers/:id` | Delete |

#### DTO (dto.go)
```go
package operation_center

// === Request DTOs ===
type CreateRequest struct {
    Name string `json:"name" binding:"required"`
}

type UpdateRequest struct {
    Name string `json:"name" binding:"required"`
}

// === Response DTOs ===
// ใช้ Model โดยตรง (ไม่ต้องสร้าง Response แยก)
```

#### Response Examples
```json
// GET /api/operation-centers
{
  "success": true,
  "data": [
    { "id": 1, "name": "จุดรวมงานภาค 1" },
    { "id": 2, "name": "จุดรวมงานภาค 2" }
  ]
}

// GET /api/operation-centers/:id
{
  "success": true,
  "data": {
    "id": 1,
    "name": "จุดรวมงานภาค 1",
    "peas": [...],      // optional: include relations
    "stations": [...]   // optional: include relations
  }
}

// POST/PUT Response
{
  "success": true,
  "data": { "id": 1, "name": "จุดรวมงานภาค 1" },
  "message": "Created successfully"
}

// DELETE Response
{
  "success": true,
  "message": "Deleted successfully"
}
```

---

### 2. Module: `pea` (6 APIs)

**Path:** `internal/modules/pea/`

#### APIs
| # | Method | Endpoint | Description |
|---|--------|----------|-------------|
| 1 | GET | `/api/peas` | List all |
| 2 | GET | `/api/peas/:id` | Get by ID |
| 3 | POST | `/api/peas` | Create single |
| 4 | POST | `/api/peas/bulk` | Bulk create |
| 5 | PUT | `/api/peas/:id` | Update |
| 6 | DELETE | `/api/peas/:id` | Delete |

#### DTO (dto.go)
```go
package pea

// === Request DTOs ===
type CreateRequest struct {
    Shortname   string `json:"shortname" binding:"required"`
    Fullname    string `json:"fullname" binding:"required"`
    OperationID int64  `json:"operationId" binding:"required"`
}

type BulkCreateRequest struct {
    Peas []CreateRequest `json:"peas" binding:"required,dive"`
}

type UpdateRequest struct {
    Shortname   string `json:"shortname"`
    Fullname    string `json:"fullname"`
    OperationID int64  `json:"operationId"`
}

// === Response DTOs ===
type Response struct {
    ID              int64  `json:"id"`
    Shortname       string `json:"shortname"`
    Fullname        string `json:"fullname"`
    OperationID     int64  `json:"operationId"`
    OperationCenter *OperationCenterResponse `json:"operationCenter,omitempty"`
}

type OperationCenterResponse struct {
    ID   int64  `json:"id"`
    Name string `json:"name"`
}
```

#### Response Examples
```json
// GET /api/peas
{
  "success": true,
  "data": [
    {
      "id": 1,
      "shortname": "กฟส.ลพบุรี",
      "fullname": "การไฟฟ้าส่วนภูมิภาคจังหวัดลพบุรี",
      "operationId": 1,
      "operationCenter": { "id": 1, "name": "จุดรวมงานภาค 1" }
    }
  ]
}

// POST /api/peas/bulk
{
  "success": true,
  "data": [...],
  "message": "Created 5 PEAs successfully"
}
```

---

### 3. Module: `station` (5 APIs)

**Path:** `internal/modules/station/`

#### APIs
| # | Method | Endpoint | Description |
|---|--------|----------|-------------|
| 1 | GET | `/api/stations` | List all with operation center |
| 2 | GET | `/api/stations/:id` | Get by ID |
| 3 | POST | `/api/stations` | Create |
| 4 | PUT | `/api/stations/:id` | Update |
| 5 | DELETE | `/api/stations/:id` | Delete |

#### DTO (dto.go)
```go
package station

// === Request DTOs ===
type CreateRequest struct {
    Name        string `json:"name" binding:"required"`
    CodeName    string `json:"codeName" binding:"required"`
    OperationID int64  `json:"operationId" binding:"required"`
}

type UpdateRequest struct {
    Name        string `json:"name"`
    CodeName    string `json:"codeName"`
    OperationID int64  `json:"operationId"`
}

// === Response DTOs ===
type Response struct {
    ID              int64                    `json:"id"`
    Name            string                   `json:"name"`
    CodeName        string                   `json:"codeName"`
    OperationID     int64                    `json:"operationId"`
    OperationCenter *OperationCenterResponse `json:"operationCenter,omitempty"`
    Feeders         []FeederResponse         `json:"feeders,omitempty"`
}

type OperationCenterResponse struct {
    ID   int64  `json:"id"`
    Name string `json:"name"`
}

type FeederResponse struct {
    ID   int64  `json:"id"`
    Code string `json:"code"`
}
```

---

### 4. Module: `feeder` (5 APIs)

**Path:** `internal/modules/feeder/`

#### APIs
| # | Method | Endpoint | Description |
|---|--------|----------|-------------|
| 1 | GET | `/api/feeders` | List all with station & operation |
| 2 | GET | `/api/feeders/:id` | Get by ID |
| 3 | POST | `/api/feeders` | Create |
| 4 | PUT | `/api/feeders/:id` | Update |
| 5 | DELETE | `/api/feeders/:id` | Delete |

#### DTO (dto.go)
```go
package feeder

// === Request DTOs ===
type CreateRequest struct {
    Code      string `json:"code" binding:"required"`
    StationID int64  `json:"stationId" binding:"required"`
}

type UpdateRequest struct {
    Code      string `json:"code"`
    StationID int64  `json:"stationId"`
}

// === Response DTOs ===
type Response struct {
    ID        int64            `json:"id"`
    Code      string           `json:"code"`
    StationID int64            `json:"stationId"`
    Station   *StationResponse `json:"station,omitempty"`
}

type StationResponse struct {
    ID              int64                    `json:"id"`
    Name            string                   `json:"name"`
    CodeName        string                   `json:"codeName"`
    OperationCenter *OperationCenterResponse `json:"operationCenter,omitempty"`
}

type OperationCenterResponse struct {
    ID   int64  `json:"id"`
    Name string `json:"name"`
}
```

---

### 5. Module: `job_type` (5 APIs)

**Path:** `internal/modules/job_type/`

#### APIs
| # | Method | Endpoint | Description |
|---|--------|----------|-------------|
| 1 | GET | `/api/job-types` | List all with task count |
| 2 | GET | `/api/job-types/:id` | Get by ID |
| 3 | POST | `/api/job-types` | Create |
| 4 | PUT | `/api/job-types/:id` | Update |
| 5 | DELETE | `/api/job-types/:id` | Delete |

#### DTO (dto.go)
```go
package job_type

// === Request DTOs ===
type CreateRequest struct {
    Name string `json:"name" binding:"required"`
}

type UpdateRequest struct {
    Name string `json:"name" binding:"required"`
}

// === Response DTOs ===
type Response struct {
    ID        int64  `json:"id"`
    Name      string `json:"name"`
    TaskCount int64  `json:"taskCount,omitempty"` // จำนวน task ที่ใช้ job type นี้
}

type ResponseWithDetails struct {
    ID         int64               `json:"id"`
    Name       string              `json:"name"`
    JobDetails []JobDetailResponse `json:"jobDetails,omitempty"`
}

type JobDetailResponse struct {
    ID   int64  `json:"id"`
    Name string `json:"name"`
}
```

---

### 6. Module: `job_detail` (5 APIs)

**Path:** `internal/modules/job_detail/`

#### APIs
| # | Method | Endpoint | Description |
|---|--------|----------|-------------|
| 1 | GET | `/api/job-details` | List all with task count |
| 2 | GET | `/api/job-details/:id` | Get by ID |
| 3 | POST | `/api/job-details` | Create |
| 4 | PUT | `/api/job-details/:id` | Update |
| 5 | DELETE | `/api/job-details/:id` | Soft delete |

#### DTO (dto.go)
```go
package job_detail

import "time"

// === Request DTOs ===
type CreateRequest struct {
    Name      string `json:"name" binding:"required"`
    JobTypeID *int64 `json:"jobTypeId"` // optional
}

type UpdateRequest struct {
    Name      string `json:"name"`
    JobTypeID *int64 `json:"jobTypeId"`
}

// === Response DTOs ===
type Response struct {
    ID        int64            `json:"id"`
    Name      string           `json:"name"`
    JobTypeID *int64           `json:"jobTypeId"`
    JobType   *JobTypeResponse `json:"jobType,omitempty"`
    TaskCount int64            `json:"taskCount,omitempty"`
    CreatedAt time.Time        `json:"createdAt"`
    UpdatedAt time.Time        `json:"updatedAt"`
}

type JobTypeResponse struct {
    ID   int64  `json:"id"`
    Name string `json:"name"`
}
```

**Note:** ใช้ Soft Delete (มี deletedAt field)

---

### 7. Module: `team` (5 APIs)

**Path:** `internal/modules/team/`

#### APIs
| # | Method | Endpoint | Description |
|---|--------|----------|-------------|
| 1 | GET | `/api/teams` | List all |
| 2 | GET | `/api/teams/:id` | Get by ID |
| 3 | POST | `/api/teams` | Create |
| 4 | PUT | `/api/teams/:id` | Update |
| 5 | DELETE | `/api/teams/:id` | Delete |

#### DTO (dto.go)
```go
package team

// === Request DTOs ===
type CreateRequest struct {
    Name string `json:"name" binding:"required"`
}

type UpdateRequest struct {
    Name string `json:"name" binding:"required"`
}

// === Response DTOs ===
type Response struct {
    ID        int64 `json:"id"`
    Name      string `json:"name"`
    TaskCount int64  `json:"taskCount,omitempty"` // optional: นับจำนวน task
}
```

---

### 8. Module: `task` (6 APIs) ⭐ Core Module

**Path:** `internal/modules/task/`

#### APIs
| # | Method | Endpoint | Description |
|---|--------|----------|-------------|
| 1 | GET | `/api/tasks` | List with filters |
| 2 | GET | `/api/tasks/:id` | Get by ID with relations |
| 3 | POST | `/api/tasks` | Create |
| 4 | PUT | `/api/tasks/:id` | Update |
| 5 | DELETE | `/api/tasks/:id` | Soft delete |
| 6 | GET | `/api/tasks/by-team` | Group by team |

#### Query Parameters (GET /api/tasks)
```
?year=2024           - Filter by year
&month=1             - Filter by month (1-12)
&teamId=1            - Filter by team
&jobTypeId=1         - Filter by job type
&feederId=1          - Filter by feeder
&workDate=2024-01-15 - Filter by specific date
```

#### DTO (dto.go)
```go
package task

import (
    "time"
    "github.com/lib/pq"
)

// === Request DTOs ===
type CreateRequest struct {
    WorkDate    string   `json:"workDate" binding:"required"`    // "2024-01-15"
    TeamID      int64    `json:"teamId" binding:"required"`
    JobTypeID   int64    `json:"jobTypeId" binding:"required"`
    JobDetailID int64    `json:"jobDetailId" binding:"required"`
    FeederID    *int64   `json:"feederId"`                       // optional
    NumPole     *string  `json:"numPole"`                        // optional
    DeviceCode  *string  `json:"deviceCode"`                     // optional
    Detail      *string  `json:"detail"`                         // optional
    UrlsBefore  []string `json:"urlsBefore"`                     // array of URLs
    UrlsAfter   []string `json:"urlsAfter"`                      // array of URLs
    Latitude    *float64 `json:"latitude" binding:"omitempty,min=-90,max=90"`
    Longitude   *float64 `json:"longitude" binding:"omitempty,min=-180,max=180"`
}

type UpdateRequest struct {
    WorkDate    *string  `json:"workDate"`
    TeamID      *int64   `json:"teamId"`
    JobTypeID   *int64   `json:"jobTypeId"`
    JobDetailID *int64   `json:"jobDetailId"`
    FeederID    *int64   `json:"feederId"`
    NumPole     *string  `json:"numPole"`
    DeviceCode  *string  `json:"deviceCode"`
    Detail      *string  `json:"detail"`
    UrlsBefore  []string `json:"urlsBefore"`
    UrlsAfter   []string `json:"urlsAfter"`
    Latitude    *float64 `json:"latitude"`
    Longitude   *float64 `json:"longitude"`
}

type FilterParams struct {
    Year      *int    `form:"year"`
    Month     *int    `form:"month"`
    TeamID    *int64  `form:"teamId"`
    JobTypeID *int64  `form:"jobTypeId"`
    FeederID  *int64  `form:"feederId"`
    WorkDate  *string `form:"workDate"`
}

// === Response DTOs ===
type Response struct {
    ID          int64              `json:"id"`
    WorkDate    string             `json:"workDate"`
    TeamID      int64              `json:"teamId"`
    JobTypeID   int64              `json:"jobTypeId"`
    JobDetailID int64              `json:"jobDetailId"`
    FeederID    *int64             `json:"feederId"`
    NumPole     *string            `json:"numPole"`
    DeviceCode  *string            `json:"deviceCode"`
    Detail      *string            `json:"detail"`
    UrlsBefore  pq.StringArray     `json:"urlsBefore"`
    UrlsAfter   pq.StringArray     `json:"urlsAfter"`
    Latitude    *float64           `json:"latitude"`
    Longitude   *float64           `json:"longitude"`
    CreatedAt   time.Time          `json:"createdAt"`
    UpdatedAt   time.Time          `json:"updatedAt"`

    // Relations
    Team      *TeamResponse      `json:"team,omitempty"`
    JobType   *JobTypeResponse   `json:"jobType,omitempty"`
    JobDetail *JobDetailResponse `json:"jobDetail,omitempty"`
    Feeder    *FeederResponse    `json:"feeder,omitempty"`
}

type TeamResponse struct {
    ID   int64  `json:"id"`
    Name string `json:"name"`
}

type JobTypeResponse struct {
    ID   int64  `json:"id"`
    Name string `json:"name"`
}

type JobDetailResponse struct {
    ID   int64  `json:"id"`
    Name string `json:"name"`
}

type FeederResponse struct {
    ID      int64            `json:"id"`
    Code    string           `json:"code"`
    Station *StationResponse `json:"station,omitempty"`
}

type StationResponse struct {
    ID       int64  `json:"id"`
    Name     string `json:"name"`
    CodeName string `json:"codeName"`
}

// Group by team response
type TasksByTeamResponse struct {
    TeamID    int64      `json:"teamId"`
    TeamName  string     `json:"teamName"`
    Tasks     []Response `json:"tasks"`
    TaskCount int        `json:"taskCount"`
}
```

#### Response Examples
```json
// GET /api/tasks/:id
{
  "success": true,
  "data": {
    "id": 1,
    "workDate": "2024-01-15",
    "teamId": 1,
    "jobTypeId": 1,
    "jobDetailId": 1,
    "feederId": 1,
    "numPole": "A001",
    "deviceCode": "SW-001",
    "detail": "รายละเอียดงาน",
    "urlsBefore": ["https://..."],
    "urlsAfter": ["https://..."],
    "latitude": 13.756331,
    "longitude": 100.501762,
    "createdAt": "2024-01-15T10:00:00Z",
    "updatedAt": "2024-01-15T10:00:00Z",
    "team": { "id": 1, "name": "ทีม A" },
    "jobType": { "id": 1, "name": "งานบำรุงรักษา" },
    "jobDetail": { "id": 1, "name": "ตัดต้นไม้" },
    "feeder": {
      "id": 1,
      "code": "LPB-01",
      "station": { "id": 1, "name": "สถานี A", "codeName": "ST-A" }
    }
  }
}

// GET /api/tasks/by-team
{
  "success": true,
  "data": [
    {
      "teamId": 1,
      "teamName": "ทีม A",
      "tasks": [...],
      "taskCount": 25
    }
  ]
}
```

**Validation Rules:**
- `latitude`: -90 ถึง 90
- `longitude`: -180 ถึง 180
- ถ้าส่ง lat ต้องส่ง lng ด้วย (หรือไม่ส่งทั้งคู่)

---

### 9. Module: `dashboard` (5 APIs) 📊

**Path:** `internal/modules/dashboard/`

#### APIs
| # | Method | Endpoint | Description |
|---|--------|----------|-------------|
| 1 | GET | `/api/dashboard/summary` | Summary statistics |
| 2 | GET | `/api/dashboard/stats` | Comprehensive stats |
| 3 | GET | `/api/dashboard/top-jobs` | Top 10 jobs |
| 4 | GET | `/api/dashboard/top-feeders` | Top 10 feeders |
| 5 | GET | `/api/dashboard/feeder-matrix` | Job breakdown by feeder |

#### Query Parameters (ใช้ร่วมกันทุก API)
```
?year=2024          - Filter by year
&month=1            - Filter by month
&teamId=1           - Filter by team
&jobTypeId=1        - Filter by job type
&feederId=1         - Filter by feeder (feeder-matrix)
&limit=10           - Limit results (top-jobs, top-feeders)
&startDate=2024-01-01  - Start date (stats)
&endDate=2024-12-31    - End date (stats)
```

#### DTO (dto.go)
```go
package dashboard

// === Request DTOs ===
type FilterParams struct {
    Year      *int   `form:"year"`
    Month     *int   `form:"month"`
    TeamID    *int64 `form:"teamId"`
    JobTypeID *int64 `form:"jobTypeId"`
    FeederID  *int64 `form:"feederId"`
    Limit     *int   `form:"limit"`
    StartDate *string `form:"startDate"`
    EndDate   *string `form:"endDate"`
}

// === Response DTOs ===

// 1. Summary Response
type SummaryResponse struct {
    TotalTasks    int64        `json:"totalTasks"`
    TotalJobTypes int64        `json:"totalJobTypes"`
    TotalFeeders  int64        `json:"totalFeeders"`
    TopTeam       *TopTeam     `json:"topTeam"`
}

type TopTeam struct {
    ID        int64  `json:"id"`
    Name      string `json:"name"`
    TaskCount int64  `json:"taskCount"`
}

// 2. Top Jobs Response
type TopJobResponse struct {
    JobDetailID   int64  `json:"jobDetailId"`
    JobDetailName string `json:"jobDetailName"`
    JobTypeName   string `json:"jobTypeName"`
    TaskCount     int64  `json:"taskCount"`
}

// 3. Top Feeders Response
type TopFeederResponse struct {
    FeederID    int64  `json:"feederId"`
    FeederCode  string `json:"feederCode"`
    StationName string `json:"stationName"`
    TaskCount   int64  `json:"taskCount"`
}

// 4. Feeder Matrix Response
type FeederMatrixResponse struct {
    FeederID   int64                  `json:"feederId"`
    FeederCode string                 `json:"feederCode"`
    JobDetails []JobDetailCount       `json:"jobDetails"`
    TotalTasks int64                  `json:"totalTasks"`
}

type JobDetailCount struct {
    JobDetailID int64  `json:"jobDetailId"`
    Name        string `json:"name"`
    Count       int64  `json:"count"`
}

// 5. Stats Response (Comprehensive)
type StatsResponse struct {
    Summary StatsSummary `json:"summary"`
    Charts  StatsCharts  `json:"charts"`
}

type StatsSummary struct {
    TotalTasks   int64          `json:"totalTasks"`
    TotalTeams   int64          `json:"totalTeams"`
    TopJobDetail *TopJobDetail  `json:"topJobDetail"`
    TopFeeder    *TopFeederInfo `json:"topFeeder"`
}

type TopJobDetail struct {
    Name  string `json:"name"`
    Count int64  `json:"count"`
}

type TopFeederInfo struct {
    Code  string `json:"code"`
    Count int64  `json:"count"`
}

type StatsCharts struct {
    ByFeeder  []ChartItem `json:"byFeeder"`
    ByJobType []ChartItem `json:"byJobType"`
    ByTeam    []ChartItem `json:"byTeam"`
    ByDate    []ChartItem `json:"byDate"`
}

type ChartItem struct {
    Label string `json:"label"`
    Value int64  `json:"value"`
}
```

**Note:** Dashboard module ไม่มี model เป็นของตัวเอง ใช้ query จาก task module

---

### 10. Module: `upload` (2 APIs) 📁

**Path:** `internal/modules/upload/`

#### APIs
| # | Method | Endpoint | Description |
|---|--------|----------|-------------|
| 1 | POST | `/api/upload` | Get presigned URL |
| 2 | DELETE | `/api/upload/:key` | Delete file |

#### DTO (dto.go)
```go
package upload

// === Request DTOs ===
type PresignRequest struct {
    FileName string `json:"fileName" binding:"required"`
    FileType string `json:"fileType" binding:"required,oneof=image/jpeg image/jpg image/png image/webp image/gif"`
}

// === Response DTOs ===
type PresignResponse struct {
    UploadURL string `json:"uploadUrl"` // Presigned URL for upload
    FileURL   string `json:"fileUrl"`   // Public URL after upload
    FileKey   string `json:"fileKey"`   // Key for delete
}

type DeleteResponse struct {
    Message string `json:"message"`
}
```

#### Response Examples
```json
// POST /api/upload
{
  "success": true,
  "data": {
    "uploadUrl": "https://r2.cloudflarestorage.com/bucket/images/xxx.jpg?X-Amz-...",
    "fileUrl": "https://photo.example.com/images/xxx.jpg",
    "fileKey": "images/xxx.jpg"
  }
}

// DELETE /api/upload/:key
{
  "success": true,
  "message": "File deleted successfully"
}
```

**Allowed File Types:**
- `image/jpeg`
- `image/jpg`
- `image/png`
- `image/webp`
- `image/gif`

**Note:** Upload module ไม่มี model และ repository เพราะไม่ใช้ database (ใช้ S3/R2)

---

## Environment Variables

### .env.example (Development)

```env
# Server
PORT=8080
GIN_MODE=debug
APP_ENV=development

# Database (PostgreSQL)
DATABASE_URL=postgresql://user:password@host:5432/database?sslmode=require

# Cloudflare R2 (S3)
R2_ACCOUNT_ID=your_account_id
R2_ACCESS_KEY_ID=your_access_key
R2_SECRET_ACCESS_KEY=your_secret_key
R2_BUCKET_NAME=your_bucket
R2_PUBLIC_URL=https://photo.example.com

# Logging
LOG_LEVEL=debug
LOG_FORMAT=text
```

### .env.production.example (Production/Docker)

```env
# Server
PORT=8080
GIN_MODE=release
APP_ENV=production

# Database (PostgreSQL - Neon or other)
DATABASE_URL=postgresql://user:password@host:5432/database?sslmode=require

# Cloudflare R2 (S3)
R2_ACCOUNT_ID=your_account_id
R2_ACCESS_KEY_ID=your_access_key
R2_SECRET_ACCESS_KEY=your_secret_key
R2_BUCKET_NAME=your_bucket
R2_PUBLIC_URL=https://photo.example.com

# Logging
LOG_LEVEL=info
LOG_FORMAT=json

# CORS (comma-separated origins)
CORS_ALLOWED_ORIGINS=https://your-frontend.com,https://www.your-frontend.com
```

---

## Dependencies (go.mod)

```go
require (
    github.com/gin-gonic/gin v1.9.1
    gorm.io/gorm v1.25.5
    gorm.io/driver/postgres v1.5.4
    github.com/aws/aws-sdk-go-v2 v1.24.0
    github.com/aws/aws-sdk-go-v2/service/s3 v1.47.0
    github.com/aws/aws-sdk-go-v2/credentials v1.16.12
    github.com/joho/godotenv v1.5.1
    github.com/lib/pq v1.10.9
    github.com/go-playground/validator/v10 v10.16.0
)
```

---

## Standard Response Format

```go
// Success Response
type Response struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Message string      `json:"message,omitempty"`
}

// Error Response
type ErrorResponse struct {
    Success bool   `json:"success"`
    Error   string `json:"error"`
}

// Pagination Response
type PaginatedResponse struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data"`
    Meta    Meta        `json:"meta"`
}

type Meta struct {
    Total       int64 `json:"total"`
    Page        int   `json:"page"`
    PerPage     int   `json:"perPage"`
    TotalPages  int   `json:"totalPages"`
}
```

---

## Notes

1. **No Authentication**: ระบบเดิมไม่มี auth - พิจารณาเพิ่ม JWT หรือ API Key ในอนาคต
2. **Soft Delete**: JobDetail และ TaskDaily ใช้ soft delete (deletedAt field)
3. **BigInt IDs**: PostgreSQL ใช้ BigInt สำหรับ primary keys
4. **Array Fields**: urlsBefore และ urlsAfter เป็น PostgreSQL array (`text[]`)
5. **Coordinates**: ใช้ Decimal(9,6) สำหรับ latitude/longitude
6. **CORS**: ต้อง config ให้รองรับ frontend origin

---

## Quick Start

```bash
# 1. Init go module
go mod init backend-hotlines3

# 2. Install dependencies
go mod tidy

# 3. Copy env
cp .env.example .env

# 4. Run migrations (GORM AutoMigrate)
go run cmd/server/main.go migrate

# 5. Start server
go run cmd/server/main.go
```

---

## Docker Deployment

### Docker Hub Repository
- **Registry**: Docker Hub
- **Repository**: `yourusername/hotline-backend`
- **Tags**: `latest`, `v1.0.0`, `dev`

### Dockerfile (Multi-stage Build)

```dockerfile
# ===== Build Stage =====
FROM golang:1.22-alpine AS builder

# Install dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.Version=${VERSION:-dev}" \
    -o /app/server ./cmd/server

# ===== Production Stage =====
FROM alpine:3.19

# Install ca-certificates for HTTPS and tzdata for timezone
RUN apk --no-cache add ca-certificates tzdata

# Set timezone
ENV TZ=Asia/Bangkok

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/server .

# Change ownership
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
ENTRYPOINT ["./server"]
```

### docker-compose.yml (Development)

```yaml
version: '3.8'

services:
  api:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: hotline-api
    ports:
      - "8080:8080"
    environment:
      - PORT=8080
      - GIN_MODE=release
      - DATABASE_URL=${DATABASE_URL}
      - R2_ACCOUNT_ID=${R2_ACCOUNT_ID}
      - R2_ACCESS_KEY_ID=${R2_ACCESS_KEY_ID}
      - R2_SECRET_ACCESS_KEY=${R2_SECRET_ACCESS_KEY}
      - R2_BUCKET_NAME=${R2_BUCKET_NAME}
      - R2_PUBLIC_URL=${R2_PUBLIC_URL}
    restart: unless-stopped
    networks:
      - hotline-network
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 10s

networks:
  hotline-network:
    driver: bridge
```

### docker-compose.prod.yml (Production with Docker Hub)

```yaml
version: '3.8'

services:
  api:
    image: yourusername/hotline-backend:latest
    container_name: hotline-api-prod
    ports:
      - "8080:8080"
    env_file:
      - .env.production
    restart: always
    networks:
      - hotline-network
    deploy:
      resources:
        limits:
          cpus: '1'
          memory: 512M
        reservations:
          cpus: '0.25'
          memory: 128M
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"

networks:
  hotline-network:
    driver: bridge
```

### Build & Push Commands

```bash
# Login to Docker Hub
docker login

# Build image with tag
docker build -t yourusername/hotline-backend:latest .
docker build -t yourusername/hotline-backend:v1.0.0 .

# Push to Docker Hub
docker push yourusername/hotline-backend:latest
docker push yourusername/hotline-backend:v1.0.0

# Pull and run on server
docker pull yourusername/hotline-backend:latest
docker-compose -f docker-compose.prod.yml up -d
```

### Makefile Commands

```makefile
.PHONY: build run test docker-build docker-push deploy

# Variables
APP_NAME=hotline-backend
DOCKER_USER=yourusername
VERSION?=latest

# Local development
build:
	go build -o bin/server ./cmd/server

run:
	go run ./cmd/server

test:
	go test -v ./...

# Docker commands
docker-build:
	docker build -t $(DOCKER_USER)/$(APP_NAME):$(VERSION) .

docker-push:
	docker push $(DOCKER_USER)/$(APP_NAME):$(VERSION)

docker-run:
	docker-compose up -d

docker-stop:
	docker-compose down

# Production deployment
deploy: docker-build docker-push
	@echo "Deployed $(DOCKER_USER)/$(APP_NAME):$(VERSION)"

# Build and push with version
release:
	@read -p "Enter version (e.g., v1.0.0): " version; \
	docker build -t $(DOCKER_USER)/$(APP_NAME):$$version .; \
	docker build -t $(DOCKER_USER)/$(APP_NAME):latest .; \
	docker push $(DOCKER_USER)/$(APP_NAME):$$version; \
	docker push $(DOCKER_USER)/$(APP_NAME):latest
```

---

## Health Check Endpoint

เพิ่ม health check สำหรับ Docker และ Load Balancer:

```go
// internal/modules/health/controller.go
package health

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)

type Controller struct {
    db *gorm.DB
}

func NewController(db *gorm.DB) *Controller {
    return &Controller{db: db}
}

func (c *Controller) Health(ctx *gin.Context) {
    ctx.JSON(http.StatusOK, gin.H{
        "status": "healthy",
        "service": "hotline-backend",
    })
}

func (c *Controller) Ready(ctx *gin.Context) {
    // Check database connection
    sqlDB, err := c.db.DB()
    if err != nil {
        ctx.JSON(http.StatusServiceUnavailable, gin.H{
            "status": "not ready",
            "error": "database connection failed",
        })
        return
    }

    if err := sqlDB.Ping(); err != nil {
        ctx.JSON(http.StatusServiceUnavailable, gin.H{
            "status": "not ready",
            "error": "database ping failed",
        })
        return
    }

    ctx.JSON(http.StatusOK, gin.H{
        "status": "ready",
        "database": "connected",
    })
}

// routes.go
func RegisterRoutes(router *gin.Engine, db *gorm.DB) {
    ctrl := NewController(db)
    router.GET("/health", ctrl.Health)
    router.GET("/ready", ctrl.Ready)
}
```

---

## CI/CD with GitHub Actions

### .github/workflows/docker-build.yml

```yaml
name: Build and Push Docker Image

on:
  push:
    branches: [main]
    tags: ['v*']
  pull_request:
    branches: [main]

env:
  DOCKER_USER: ${{ secrets.DOCKER_USERNAME }}
  IMAGE_NAME: hotline-backend

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Run tests
        run: go test -v ./...

  build-and-push:
    needs: test
    runs-on: ubuntu-latest
    if: github.event_name == 'push'

    steps:
      - uses: actions/checkout@v4

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Login to Docker Hub
        uses: docker/login-action@v3
        with:
          username: ${{ secrets.DOCKER_USERNAME }}
          password: ${{ secrets.DOCKER_TOKEN }}

      - name: Extract metadata
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: ${{ env.DOCKER_USER }}/${{ env.IMAGE_NAME }}
          tags: |
            type=ref,event=branch
            type=semver,pattern={{version}}
            type=semver,pattern={{major}}.{{minor}}
            type=sha,prefix=

      - name: Build and push
        uses: docker/build-push-action@v5
        with:
          context: .
          push: true
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
          cache-from: type=gha
          cache-to: type=gha,mode=max
```

### GitHub Secrets Required

ตั้งค่าใน GitHub Repository Settings → Secrets:

| Secret Name | Description |
|-------------|-------------|
| `DOCKER_USERNAME` | Docker Hub username |
| `DOCKER_TOKEN` | Docker Hub access token (ไม่ใช้ password) |

### Deployment Workflow

```
1. Developer push code → GitHub
2. GitHub Actions runs tests
3. Build Docker image
4. Push to Docker Hub (yourusername/hotline-backend)
5. Server pulls latest image
6. Restart container with docker-compose
```

---

## Server Deployment Script

### deploy.sh (ใช้บน Production Server)

```bash
#!/bin/bash
set -e

# Configuration
DOCKER_USER="yourusername"
IMAGE_NAME="hotline-backend"
COMPOSE_FILE="docker-compose.prod.yml"

echo "🚀 Starting deployment..."

# Pull latest image
echo "📥 Pulling latest image..."
docker pull ${DOCKER_USER}/${IMAGE_NAME}:latest

# Stop current container
echo "⏹️  Stopping current container..."
docker-compose -f ${COMPOSE_FILE} down

# Start new container
echo "▶️  Starting new container..."
docker-compose -f ${COMPOSE_FILE} up -d

# Health check
echo "🏥 Waiting for health check..."
sleep 5
if curl -s http://localhost:8080/health | grep -q "healthy"; then
    echo "✅ Deployment successful!"
else
    echo "❌ Health check failed!"
    exit 1
fi

# Cleanup old images
echo "🧹 Cleaning up old images..."
docker image prune -f

echo "🎉 Done!"
```
