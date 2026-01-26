# Backend Hotlines3 - Product Backlog & Todo List

> **Last Updated:** 2026-01-26
> **Project Status:** Phase 1-4 Completed | Phase 5+ Pending
> **Total Endpoints:** 48+ implemented

---

## Executive Summary

โปรเจค Backend Hotlines3 มีความสมบูรณ์ของ API endpoints สำหรับ Phase 1-4 แล้ว แต่ยังขาดส่วนสำคัญที่ต้องดำเนินการก่อน Production:

| หมวด | สถานะ | Priority |
|------|--------|----------|
| Core APIs (Phase 1-4) | ✅ Done | - |
| Authentication (Phase 5) | ❌ Not Started | P0 - Critical |
| Testing | ❌ Not Started | P0 - Critical |
| Security Hardening | ❌ Not Started | P1 - High |
| API Documentation | ❌ Not Started | P2 - Medium |
| Frontend Integration Fix | ⚠️ Issues Found | P1 - High |

---

## Phase 5: Authentication System [P0 - CRITICAL]

> **ความสำคัญ:** ระบบยังไม่มี Authentication เลย ทำให้ไม่ปลอดภัยสำหรับ Production

### 5.1 JWT Authentication Endpoints

- [ ] **POST /v1/auth/login** - เข้าสู่ระบบ
  - Input: username/email, password
  - Output: access_token, refresh_token, user info
  - Hash password ด้วย bcrypt

- [ ] **POST /v1/auth/logout** - ออกจากระบบ
  - Invalidate refresh token
  - Clear session if any

- [ ] **POST /v1/auth/refresh** - Refresh token
  - Input: refresh_token
  - Output: new access_token, new refresh_token
  - Implement token rotation

- [ ] **GET /v1/auth/me** - ดูข้อมูล user ปัจจุบัน
  - Requires valid JWT
  - Return user profile

### 5.2 User Management Model

- [ ] **สร้าง User Model**
  ```go
  type User struct {
      ID        uint
      Username  string (unique)
      Email     string (unique)
      Password  string (hashed)
      Role      string (admin/user/viewer)
      TeamID    *uint (nullable)
      IsActive  bool
      LastLogin time.Time
      CreatedAt, UpdatedAt, DeletedAt
  }
  ```

- [ ] **User CRUD Endpoints (Admin only)**
  - [ ] GET /v1/users - รายการ users
  - [ ] GET /v1/users/:id - ดู user
  - [ ] POST /v1/users - สร้าง user
  - [ ] PUT /v1/users/:id - แก้ไข user
  - [ ] DELETE /v1/users/:id - ลบ user

### 5.3 Middleware Implementation

- [ ] **JWT Middleware** - ตรวจสอบ token ทุก request
- [ ] **Role-based Access Control (RBAC)**
  - Admin: Full access
  - User: CRUD tasks, view master data
  - Viewer: Read-only access
- [ ] **Route Protection** - กำหนด routes ที่ต้อง auth

---

## Phase 6: Testing [P0 - CRITICAL]

> **ความสำคัญ:** ไม่มี tests เลย ทำให้ไม่มั่นใจในความถูกต้องของ code

### 6.1 Unit Tests

- [ ] **Handler Tests**
  - [ ] Task handlers (task.go)
  - [ ] Team handlers (team.go)
  - [ ] JobType handlers (job_type.go)
  - [ ] JobDetail handlers (job_detail.go)
  - [ ] Feeder handlers (feeder.go)
  - [ ] Station handlers (station.go)
  - [ ] PEA handlers (pea.go)
  - [ ] OperationCenter handlers (operation_center.go)
  - [ ] Dashboard handlers (dashboard.go)
  - [ ] Upload handlers (upload.go)

- [ ] **Model Tests**
  - [ ] StringArray custom type
  - [ ] Model relationships
  - [ ] Validations

### 6.2 Integration Tests

- [ ] **Database Integration**
  - [ ] Connection test
  - [ ] Migration test
  - [ ] CRUD operations test

- [ ] **API Integration Tests**
  - [ ] Full flow: Create task → Get task → Update → Delete
  - [ ] Filter & pagination tests
  - [ ] Dashboard aggregation tests

### 6.3 Test Infrastructure

- [ ] Setup test database (separate from production)
- [ ] Setup test fixtures/factories
- [ ] Setup CI/CD pipeline for automated testing
- [ ] Code coverage reporting (target: 80%+)

---

## Phase 7: Security Hardening [P1 - HIGH]

### 7.1 Credentials Management

- [ ] **ย้าย secrets ไปใช้ Environment Variables**
  - [ ] Database credentials
  - [ ] Cloudflare R2 keys
  - [ ] JWT secret
  - [ ] สร้าง .env.example template

- [ ] **ลบ credentials ออกจาก config.yaml**
  - [ ] ใช้ Viper AutomaticEnv()
  - [ ] อัพเดท deployment scripts

### 7.2 Rate Limiting

- [ ] **Implement Rate Limiter Middleware**
  - [ ] Per-IP rate limiting
  - [ ] Per-user rate limiting (when authenticated)
  - [ ] Different limits for different endpoints
  - Suggested limits:
    - Login: 5 requests/minute
    - API general: 100 requests/minute
    - Upload: 10 requests/minute

### 7.3 Input Validation

- [ ] **เพิ่ม Request Validation ให้ครบ**
  - [ ] Task creation validation (required fields, date format)
  - [ ] Master data validation (unique constraints)
  - [ ] Upload validation (file type, size)
  - [ ] Query parameter validation (pagination limits)

### 7.4 Security Headers

- [ ] **เพิ่ม Security Headers Middleware**
  - [ ] X-Content-Type-Options: nosniff
  - [ ] X-Frame-Options: DENY
  - [ ] X-XSS-Protection: 1; mode=block
  - [ ] Content-Security-Policy
  - [ ] Strict-Transport-Security (HSTS)

---

## Phase 8: API Documentation [P2 - MEDIUM]

### 8.1 Swagger/OpenAPI

- [ ] **Setup Swagger**
  - [ ] Install swaggo/swag
  - [ ] Add swagger annotations to handlers
  - [ ] Generate swagger.json/yaml
  - [ ] Setup Swagger UI endpoint (/swagger/*)

- [ ] **Document All Endpoints**
  - [ ] Request/Response schemas
  - [ ] Authentication requirements
  - [ ] Error codes
  - [ ] Examples

### 8.2 API Versioning Documentation

- [ ] Document versioning strategy (v1, v2, etc.)
- [ ] Document deprecation policy
- [ ] Breaking changes changelog

---

## Phase 9: Frontend Integration Issues [P1 - HIGH]

> **ความสำคัญ:** มีความไม่สอดคล้องระหว่าง Backend Go กับ Frontend expectations

### 9.1 Dashboard Stats Response Format Fix

**ปัญหา:** Field names ต่างกัน

| Frontend Expects | Backend Returns |
|-----------------|-----------------|
| `byFeeder` | `chartByFeeder` |
| `byJobType` | `chartByJobType` |
| `byTeam` | `chartByTeam` |
| `byDate` | `chartByDate` |

- [ ] **แก้ไข response format ใน dashboard.go**
  - [ ] เปลี่ยน `chartByFeeder` → `byFeeder`
  - [ ] เปลี่ยน `chartByJobType` → `byJobType`
  - [ ] เปลี่ยน `chartByTeam` → `byTeam`
  - [ ] เปลี่ยน `chartByDate` → `byDate`

### 9.2 Upload Strategy Alignment

**ปัญหา:** Frontend ใช้ direct upload แต่ Backend เตรียม presigned URL

- [ ] **Clarify upload strategy กับ Frontend team**
  - Option A: Frontend ใช้ presigned URL จาก Backend
  - Option B: Backend รับ file upload โดยตรง
- [ ] **Implement ตาม strategy ที่เลือก**

### 9.3 Response Format Consistency

- [ ] **ตรวจสอบ response format ทุก endpoint**
  - [ ] Consistent field naming (camelCase)
  - [ ] Consistent error format
  - [ ] Consistent pagination meta

---

## Phase 10: Infrastructure & DevOps [P2 - MEDIUM]

### 10.1 Logging

- [ ] **Implement Structured Logging**
  - [ ] ใช้ zerolog หรือ zap
  - [ ] Log levels (debug, info, warn, error)
  - [ ] Request/response logging
  - [ ] Error tracking

### 10.2 Monitoring

- [ ] **Health Checks Enhancement**
  - [ ] Database health
  - [ ] R2 connection health
  - [ ] Memory/CPU metrics

- [ ] **Metrics Endpoint**
  - [ ] Prometheus metrics (/metrics)
  - [ ] Request latency
  - [ ] Error rates
  - [ ] Active connections

### 10.3 CI/CD

- [ ] **GitHub Actions Setup**
  - [ ] Lint on PR
  - [ ] Test on PR
  - [ ] Build Docker image
  - [ ] Auto deploy to staging

### 10.4 Docker Improvements

- [ ] Multi-stage build optimization
- [ ] Health check in Dockerfile
- [ ] Docker compose for full stack (db + api + frontend)

---

## Phase 11: Performance Optimization [P3 - LOW]

### 11.1 Database

- [ ] Add database indexes for common queries
- [ ] Query optimization for dashboard
- [ ] Connection pooling configuration
- [ ] Implement database query caching (Redis)

### 11.2 API

- [ ] Response caching for master data
- [ ] Pagination optimization (cursor-based for large datasets)
- [ ] Gzip compression middleware

---

## Bug Fixes & Technical Debt

### Known Issues

- [ ] **Soft Delete Inconsistency**
  - TaskDaily มี DeletedAt แต่ไม่ได้ใช้ soft delete consistently
  - ต้องตรวจสอบว่า query ทุกที่ใช้ `Unscoped()` เมื่อจำเป็น

- [ ] **Error Handling Inconsistency**
  - บาง handler return error message ตรงๆ
  - ต้อง standardize error codes และ messages

- [ ] **Database Transaction**
  - Bulk operations ไม่ได้ใช้ transaction
  - ต้อง wrap bulk operations ใน transaction

---

## Sprint Planning Recommendation

### Sprint 1 (Week 1-2): Security Foundation
- [ ] JWT Authentication (Phase 5.1, 5.2, 5.3)
- [ ] Credentials to Environment Variables (Phase 7.1)

### Sprint 2 (Week 3-4): Testing & Quality
- [ ] Unit Tests for core handlers (Phase 6.1)
- [ ] Integration Tests (Phase 6.2)
- [ ] Input Validation (Phase 7.3)

### Sprint 3 (Week 5-6): Integration & Documentation
- [ ] Frontend Integration Fixes (Phase 9)
- [ ] Swagger Documentation (Phase 8.1)
- [ ] Rate Limiting (Phase 7.2)

### Sprint 4 (Week 7-8): Production Readiness
- [ ] Structured Logging (Phase 10.1)
- [ ] Monitoring (Phase 10.2)
- [ ] CI/CD Pipeline (Phase 10.3)
- [ ] Security Headers (Phase 7.4)

---

## Acceptance Criteria for Production Release

- [ ] ✅ All Phase 1-4 APIs working (Done)
- [ ] JWT Authentication implemented and tested
- [ ] Unit test coverage > 80%
- [ ] All security headers in place
- [ ] Rate limiting active
- [ ] Swagger documentation complete
- [ ] CI/CD pipeline running
- [ ] Monitoring and alerting setup
- [ ] Frontend integration verified
- [ ] Load testing passed (100 concurrent users)

---

## Notes

### API Status Summary

| Phase | Name | Endpoints | Status |
|-------|------|-----------|--------|
| 1 | Mobile Form MVP | 6 | ✅ Done |
| 2 | Task Management | 6 | ✅ Done |
| 3 | Master Data CRUD | 36 | ✅ Done |
| 4 | Dashboard Analytics | 5 | ✅ Done |
| 5 | Authentication | 8+ | ❌ Pending |
| 6+ | Testing & Security | N/A | ❌ Pending |

### Tech Stack
- **Language:** Go 1.21+
- **Framework:** Gin
- **ORM:** GORM
- **Database:** PostgreSQL (Neon)
- **File Storage:** Cloudflare R2
- **Config:** Viper

---

*Document maintained by: Product Owner*
*Last review: 2026-01-26*
