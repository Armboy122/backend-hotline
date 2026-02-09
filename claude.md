# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

ระบบจัดการงานซ่อมบำรุงโครงข่ายไฟฟ้า (Hotline Maintenance System) for Provincial Electricity Authority. Go backend with Gin, GORM, PostgreSQL (Neon), Cloudflare R2 file storage, and JWT authentication.

## Build & Run Commands

```bash
go run main.go              # Run dev server (reads config.yaml)
go build -o hotlines3-api   # Build binary
go build ./...              # Verify compilation
go vet ./...                # Static analysis
docker build -t hotline-backend .   # Docker build
docker-compose up                   # Docker run
```

No test suite exists. No Makefile.

## Architecture

**Entry point**: `main.go` → loads config (Viper/YAML) → connects DB → auto-migrates → sets up JWT → starts Gin router.

**All API routes are under `/v1/`** prefix. Auth routes require JWT Bearer token via middleware.

```
main.go                          # Startup: config → DB → JWT → router
config.yaml                      # Viper config (server, database, cloudflare, jwt, cors)

internal/
  config/config.go               # Viper config loader (Config struct)
  database/database.go           # GORM connection + AutoMigrate
  models/
    models.go                    # 9 GORM models (all table/column definitions)
    columns.go                   # Type-safe quoted column name constants
    scopes.go                    # Reusable GORM scopes (filters, soft-delete)
    helpers.go                   # CountTasksBy, CountTasksFor helpers
  dto/response.go                # All request/response DTOs, StandardResponse wrapper
  handlers/v1/                   # All HTTP handlers (one file per resource)
  middleware/
    auth.go                      # JWT auth: RequireAuth(), RequireRole(), OptionalAuth()
    error_handler.go             # RecoveryMiddleware, HandleValidationError
  router/router.go               # Route registration (all /v1/* groups)

pkg/
  jwt/jwt.go                     # JWTManager: GenerateTokenPair, ValidateToken
  password/password.go           # bcrypt: HashPassword, CheckPassword
  s3/r2.go                       # R2Client for Cloudflare R2 presigned URLs
```

## Database Column Naming (CRITICAL)

The database was originally created by Prisma. Table names are PascalCase, column names are camelCase. **You must use quoted identifiers in raw SQL.**

- **TaskDaily** timestamps are **lowercase**: `"createdat"`, `"updatedat"`, `"deletedat"`
- **JobDetail/User** timestamps are **camelCase**: `"createdAt"`, `"updatedAt"`, `"deletedAt"`
- Foreign keys are camelCase: `"teamId"`, `"jobTypeId"`, `"jobDetailId"`, `"feederId"`
- WorkDate is lowercase: `"workdate"`

**Always use the type-safe constants** from `models/columns.go` instead of raw strings:
```go
models.TaskCol.TeamID      // → `"teamId"`
models.TaskCol.DeletedAt   // → `"deletedat"`
models.JobDetailCol.DeletedAt // → `"deletedAt"`
```

Use scopes from `models/scopes.go` for common filters:
```go
db.Scopes(models.TaskNotDeleted)
db.Scopes(models.ApplyDashboardFilters(year, month, teamID, jobTypeID))
```

Use helpers from `models/helpers.go` for task counting:
```go
models.CountTasksBy(db, models.TaskCol.TeamID, ids)  // batch count
models.CountTasksFor(db, models.TaskCol.TeamID, id)   // single count
```

## Handler Pattern

Each handler in `internal/handlers/v1/` is a struct wrapping `*gorm.DB`. Methods: `List()`, `GetByID()`, `Create()`, `Update()`, `Delete()`.

Standard response format via `dto.StandardResponse`:
```go
c.JSON(http.StatusOK, dto.StandardResponse{
    Success: true,
    Data:    responseData,
    Meta:    &dto.Meta{Page: 1, Limit: 50, Total: 100}, // optional pagination
    Error:   &dto.ErrorInfo{Code: "NOT_FOUND", Message: "..."}, // on error
})
```

## Key Models

9 models in `internal/models/models.go`. Core model is **TaskDaily** which references Team, JobType, JobDetail, and Feeder. Models with soft delete: TaskDaily, JobDetail, User.

Custom types: `StringArray` for PostgreSQL `text[]` arrays (urlsBefore/urlsAfter), `decimal.Decimal` for lat/lng coordinates.

## Auth System

JWT-based. Roles: `admin`, `user`, `viewer`. Token pair: access (1h) + refresh (7d). Middleware in `internal/middleware/auth.go`. User management at `/v1/users/` requires admin role.

## API Routes

All under `/v1/`. See `internal/router/router.go` for complete route definitions.

- **Auth**: `/v1/auth/` (login, register, refresh, logout, me)
- **Master data**: teams, job-types, job-details, feeders, stations, peas, operation-centers (standard CRUD)
- **Tasks**: `/v1/tasks/` (CRUD + `/by-team` + `/by-filter?year=&month=&teamId=`)
- **Dashboard**: `/v1/dashboard/` (summary, top-jobs, top-feeders, feeder-matrix, stats)
- **Upload**: `/v1/upload/` (presigned URL for R2, delete)
- **Users**: `/v1/users/` (admin CRUD + password change)
- **Health**: `GET /health` (no version prefix)
