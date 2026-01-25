# HotlineS3 API Specification

## Overview

เอกสารนี้กำหนด RESTful API สำหรับแยก Backend ออกจาก Frontend ของระบบ HotlineS3

**Total Endpoints:** 31 APIs
**Base URL:** `https://api.hotlines3.example.com/v1`
**Development:** `http://localhost:3001/api/v1`

---

## Response Format (มาตรฐาน)

### Success Response
```json
{
  "success": true,
  "data": { ... },
  "meta": {
    "page": 1,
    "limit": 50,
    "total": 150
  }
}
```

### Error Response
```json
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "ไม่พบข้อมูลที่ต้องการ",
    "details": {}
  }
}
```

### HTTP Status Codes
| Code | Description |
|------|-------------|
| 200 | Success (GET, PUT) |
| 201 | Created (POST) |
| 204 | No Content (DELETE) |
| 400 | Bad Request |
| 401 | Unauthorized |
| 404 | Not Found |
| 409 | Conflict (Duplicate) |
| 500 | Server Error |

---

## Development Priority (ลำดับการพัฒนา)

### Phase 1: Mobile Form MVP (6 endpoints) - สัปดาห์ที่ 1-2
> **เริ่มจากหน้าบันทึกข้อมูล** - API ที่จำเป็นสำหรับ Form หลัก

| # | Method | Endpoint | Description |
|---|--------|----------|-------------|
| 1 | GET | `/v1/teams` | ดึงรายชื่อทีม (dropdown) |
| 2 | GET | `/v1/job-types` | ดึงประเภทงาน (dropdown) |
| 3 | GET | `/v1/job-details` | ดึงรายละเอียดงาน (dropdown) |
| 4 | GET | `/v1/feeders` | ดึงฟีดเดอร์ (dropdown) |
| 5 | POST | `/v1/upload/image` | อัปโหลดรูปภาพ |
| 6 | POST | `/v1/tasks` | บันทึกงานใหม่ |

### Phase 2: Task Management (6 endpoints) - สัปดาห์ที่ 2-3
| # | Method | Endpoint | Description |
|---|--------|----------|-------------|
| 7 | GET | `/v1/tasks` | ดึงรายการงานทั้งหมด |
| 8 | GET | `/v1/tasks/:id` | ดึงงานตาม ID |
| 9 | PUT | `/v1/tasks/:id` | แก้ไขงาน |
| 10 | DELETE | `/v1/tasks/:id` | ลบงาน |
| 11 | GET | `/v1/tasks/by-filter` | กรองตามปี/เดือน |
| 12 | GET | `/v1/tasks/by-team` | จัดกลุ่มตามทีม |

### Phase 3: Admin Master Data (13 endpoints) - สัปดาห์ที่ 3-4
| # | Method | Endpoint | Description |
|---|--------|----------|-------------|
| 13-16 | CRUD | `/v1/teams/*` | จัดการทีม |
| 17-20 | CRUD | `/v1/job-types/*` | จัดการประเภทงาน |
| 21-25 | CRUD+Restore | `/v1/job-details/*` | จัดการรายละเอียดงาน |
| 26-30 | CRUD | `/v1/feeders/*` | จัดการฟีดเดอร์ |
| 31-34 | CRUD | `/v1/stations/*` | จัดการสถานี |
| 35-39 | CRUD+Bulk | `/v1/peas/*` | จัดการ PEA |
| 40-43 | CRUD | `/v1/operation-centers/*` | จัดการจุดรวมงาน |

### Phase 4: Dashboard Analytics (5 endpoints) - สัปดาห์ที่ 4-5
| # | Method | Endpoint | Description |
|---|--------|----------|-------------|
| 44 | GET | `/v1/dashboard/summary` | สรุปภาพรวม |
| 45 | GET | `/v1/dashboard/top-jobs` | Top 10 งาน |
| 46 | GET | `/v1/dashboard/top-feeders` | Top 10 ฟีดเดอร์ |
| 47 | GET | `/v1/dashboard/feeder-matrix` | Matrix ฟีดเดอร์-งาน |
| 48 | GET | `/v1/dashboard/stats` | สถิติขั้นสูง |

### Phase 5: Authentication (4 endpoints) - สัปดาห์ที่ 5-6
| # | Method | Endpoint | Description |
|---|--------|----------|-------------|
| 49 | POST | `/v1/auth/login` | เข้าสู่ระบบ |
| 50 | POST | `/v1/auth/logout` | ออกจากระบบ |
| 51 | POST | `/v1/auth/refresh` | รีเฟรช token |
| 52 | GET | `/v1/auth/me` | ข้อมูล user ปัจจุบัน |

---

# Phase 1: Mobile Form APIs (รายละเอียด)

## 1. GET /v1/teams

ดึงรายชื่อทีมทั้งหมดสำหรับ dropdown

### Request
```
GET /v1/teams
```

### Response (200 OK)
```json
{
  "success": true,
  "data": [
    { "id": "1", "name": "ทีม A" },
    { "id": "2", "name": "ทีม B" },
    { "id": "3", "name": "ทีม C" }
  ]
}
```

---

## 2. GET /v1/job-types

ดึงประเภทงานทั้งหมดสำหรับ dropdown

### Request
```
GET /v1/job-types
```

### Response (200 OK)
```json
{
  "success": true,
  "data": [
    {
      "id": "1",
      "name": "ซ่อมบำรุง",
      "_count": { "tasks": 150 }
    },
    {
      "id": "2",
      "name": "ตรวจสอบ",
      "_count": { "tasks": 80 }
    }
  ]
}
```

---

## 3. GET /v1/job-details

ดึงรายละเอียดงานทั้งหมดสำหรับ dropdown (filter ตาม jobTypeId ที่ Frontend)

### Request
```
GET /v1/job-details
```

### Response (200 OK)
```json
{
  "success": true,
  "data": [
    {
      "id": "1",
      "name": "ตัดต้นไม้",
      "jobTypeId": "1",
      "createdAt": "2026-01-01T00:00:00.000Z",
      "updatedAt": "2026-01-01T00:00:00.000Z",
      "deletedAt": null,
      "_count": { "tasks": 75 }
    },
    {
      "id": "2",
      "name": "เปลี่ยนฟิวส์",
      "jobTypeId": "1",
      "createdAt": "2026-01-01T00:00:00.000Z",
      "updatedAt": "2026-01-01T00:00:00.000Z",
      "deletedAt": null,
      "_count": { "tasks": 45 }
    }
  ]
}
```

---

## 4. GET /v1/feeders

ดึงฟีดเดอร์ทั้งหมดพร้อมข้อมูลสถานี

### Request
```
GET /v1/feeders
```

### Response (200 OK)
```json
{
  "success": true,
  "data": [
    {
      "id": "1",
      "code": "SKT-01",
      "stationId": "1",
      "station": {
        "id": "1",
        "name": "สถานีไฟฟ้าย่อยสุขาภิบาล",
        "codeName": "SKT",
        "operationCenter": {
          "id": "1",
          "name": "จุดรวมงาน 1"
        }
      },
      "_count": { "tasks": 50 }
    }
  ]
}
```

---

## 5. POST /v1/upload/image

อัปโหลดรูปภาพไปยัง Cloud Storage (R2/S3)

### Request
```
POST /v1/upload/image
Content-Type: multipart/form-data
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| file | File | Yes | ไฟล์รูปภาพ |

### Constraints
- **Max Size:** 5MB
- **Allowed Types:** image/jpeg, image/png, image/webp, image/gif

### Response (200 OK)
```json
{
  "success": true,
  "data": {
    "url": "https://r2.example.com/images/1706123456789-abc123.jpg",
    "fileName": "images/1706123456789-abc123.jpg",
    "originalName": "photo.jpg",
    "size": 1234567,
    "type": "image/jpeg"
  }
}
```

### Error Response (400 Bad Request)
```json
{
  "success": false,
  "error": {
    "code": "INVALID_FILE_TYPE",
    "message": "ประเภทไฟล์ไม่ถูกต้อง รองรับเฉพาะ JPG, PNG, WebP, GIF"
  }
}
```

---

## 6. POST /v1/tasks

สร้างรายงานงานใหม่ (TaskDaily)

### Request
```
POST /v1/tasks
Content-Type: application/json
```

### Request Body
```json
{
  "workDate": "2026-01-24",
  "teamId": "1",
  "jobTypeId": "1",
  "jobDetailId": "1",
  "feederId": "1",
  "numPole": "A123",
  "deviceCode": "DEV-001",
  "detail": "รายละเอียดเพิ่มเติม",
  "urlsBefore": [
    "https://r2.example.com/images/before1.jpg",
    "https://r2.example.com/images/before2.jpg"
  ],
  "urlsAfter": [
    "https://r2.example.com/images/after1.jpg"
  ],
  "latitude": 13.756331,
  "longitude": 100.501762
}
```

### Field Specifications
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| workDate | string | Yes | วันที่ทำงาน (YYYY-MM-DD) |
| teamId | string | Yes | ID ทีม |
| jobTypeId | string | Yes | ID ประเภทงาน |
| jobDetailId | string | Yes | ID รายละเอียดงาน |
| feederId | string | No | ID ฟีดเดอร์ |
| numPole | string | No | หมายเลขเสา |
| deviceCode | string | No | รหัสอุปกรณ์ |
| detail | string | No | หมายเหตุ |
| urlsBefore | string[] | Yes | รูปก่อนทำงาน (ต้องมีอย่างน้อย 1 รูป) |
| urlsAfter | string[] | Yes | รูปหลังทำงาน (ว่างได้) |
| latitude | number | No | ละติจูด (Decimal 9,6) |
| longitude | number | No | ลองจิจูด (Decimal 9,6) |

### Response (201 Created)
```json
{
  "success": true,
  "data": {
    "id": "1",
    "workDate": "2026-01-24T00:00:00.000Z",
    "teamId": "1",
    "jobTypeId": "1",
    "jobDetailId": "1",
    "feederId": "1",
    "numPole": "A123",
    "deviceCode": "DEV-001",
    "detail": "รายละเอียดเพิ่มเติม",
    "urlsBefore": ["https://..."],
    "urlsAfter": ["https://..."],
    "latitude": 13.756331,
    "longitude": 100.501762,
    "team": {
      "id": "1",
      "name": "ทีม A"
    },
    "jobType": {
      "id": "1",
      "name": "ซ่อมบำรุง"
    },
    "jobDetail": {
      "id": "1",
      "name": "ตัดต้นไม้"
    },
    "feeder": {
      "id": "1",
      "code": "SKT-01",
      "station": {
        "name": "สถานีไฟฟ้าย่อยสุขาภิบาล",
        "operationCenter": {
          "name": "จุดรวมงาน 1"
        }
      }
    },
    "createdAt": "2026-01-24T10:30:00.000Z",
    "updatedAt": "2026-01-24T10:30:00.000Z",
    "deletedAt": null
  }
}
```

---

# Phase 2: Task Management APIs

## 7. GET /v1/tasks

ดึงรายการงานทั้งหมด (พร้อม filter และ pagination)

### Request
```
GET /v1/tasks?workDate=2026-01-24&teamId=1&page=1&limit=50
```

### Query Parameters
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| workDate | string | - | กรองตามวันที่ (YYYY-MM-DD) |
| teamId | string | - | กรองตามทีม |
| jobTypeId | string | - | กรองตามประเภทงาน |
| feederId | string | - | กรองตามฟีดเดอร์ |
| page | number | 1 | หน้าที่ต้องการ |
| limit | number | 50 | จำนวนต่อหน้า |

### Response (200 OK)
```json
{
  "success": true,
  "data": [
    { /* TaskDaily object */ }
  ],
  "meta": {
    "page": 1,
    "limit": 50,
    "total": 150
  }
}
```

---

## 8. GET /v1/tasks/:id

ดึงข้อมูลงานตาม ID

### Request
```
GET /v1/tasks/1
```

### Response (200 OK)
```json
{
  "success": true,
  "data": { /* TaskDaily object */ }
}
```

---

## 9. PUT /v1/tasks/:id

แก้ไขข้อมูลงาน

### Request
```
PUT /v1/tasks/1
Content-Type: application/json
```

### Request Body
```json
{
  "detail": "แก้ไขหมายเหตุ",
  "urlsAfter": ["https://r2.example.com/images/new-after.jpg"]
}
```
> ส่งเฉพาะ field ที่ต้องการแก้ไข

### Response (200 OK)
```json
{
  "success": true,
  "data": { /* Updated TaskDaily object */ }
}
```

---

## 10. DELETE /v1/tasks/:id

ลบงาน (Soft Delete)

### Request
```
DELETE /v1/tasks/1
```

### Response (204 No Content)
```
(empty body)
```

---

## 11. GET /v1/tasks/by-filter

ดึงงานตามปี/เดือน จัดกลุ่มตามทีม

### Request
```
GET /v1/tasks/by-filter?year=2026&month=01&teamId=1
```

### Query Parameters
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| year | string | Yes | ปี (e.g., "2026") |
| month | string | Yes | เดือน (e.g., "01") |
| teamId | string | No | กรองเฉพาะทีม |

### Response (200 OK)
```json
{
  "success": true,
  "data": {
    "ทีม A": {
      "team": { "id": "1", "name": "ทีม A" },
      "tasks": [ /* TaskDaily[] */ ]
    },
    "ทีม B": {
      "team": { "id": "2", "name": "ทีม B" },
      "tasks": [ /* TaskDaily[] */ ]
    }
  }
}
```

---

## 12. GET /v1/tasks/by-team

ดึงงานทั้งหมดจัดกลุ่มตามทีม

### Request
```
GET /v1/tasks/by-team
```

### Response (200 OK)
```json
{
  "success": true,
  "data": {
    "ทีม A": {
      "team": { "id": "1", "name": "ทีม A" },
      "tasks": [ /* TaskDaily[] */ ]
    }
  }
}
```

---

# Phase 3: Master Data CRUD APIs

## Teams CRUD

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/v1/teams` | ดึงทั้งหมด |
| GET | `/v1/teams/:id` | ดึงตาม ID |
| POST | `/v1/teams` | สร้างใหม่ |
| PUT | `/v1/teams/:id` | แก้ไข |
| DELETE | `/v1/teams/:id` | ลบ |

### POST /v1/teams - Request Body
```json
{ "name": "ทีม D" }
```

---

## Job Types CRUD

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/v1/job-types` | ดึงทั้งหมด |
| GET | `/v1/job-types/:id` | ดึงตาม ID |
| POST | `/v1/job-types` | สร้างใหม่ |
| PUT | `/v1/job-types/:id` | แก้ไข |
| DELETE | `/v1/job-types/:id` | ลบ |

### POST /v1/job-types - Request Body
```json
{ "name": "ประเภทงานใหม่" }
```

---

## Job Details CRUD + Restore

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/v1/job-details` | ดึงทั้งหมด (ที่ไม่ถูกลบ) |
| GET | `/v1/job-details/:id` | ดึงตาม ID |
| POST | `/v1/job-details` | สร้างใหม่ |
| PUT | `/v1/job-details/:id` | แก้ไข |
| DELETE | `/v1/job-details/:id` | Soft Delete |
| POST | `/v1/job-details/:id/restore` | กู้คืน |

### POST /v1/job-details - Request Body
```json
{
  "name": "รายละเอียดงานใหม่",
  "jobTypeId": "1"
}
```

---

## Feeders CRUD

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/v1/feeders` | ดึงทั้งหมด |
| GET | `/v1/feeders/:id` | ดึงตาม ID |
| POST | `/v1/feeders` | สร้างใหม่ |
| PUT | `/v1/feeders/:id` | แก้ไข |
| DELETE | `/v1/feeders/:id` | ลบ |

### POST /v1/feeders - Request Body
```json
{
  "code": "SKT-02",
  "stationId": "1"
}
```

---

## Stations CRUD

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/v1/stations` | ดึงทั้งหมด |
| GET | `/v1/stations/:id` | ดึงตาม ID |
| POST | `/v1/stations` | สร้างใหม่ |
| PUT | `/v1/stations/:id` | แก้ไข |
| DELETE | `/v1/stations/:id` | ลบ |

### POST /v1/stations - Request Body
```json
{
  "name": "สถานีใหม่",
  "codeName": "NEW",
  "operationId": "1"
}
```

---

## PEAs CRUD + Bulk Import

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/v1/peas` | ดึงทั้งหมด |
| GET | `/v1/peas/:id` | ดึงตาม ID |
| POST | `/v1/peas` | สร้างใหม่ |
| POST | `/v1/peas/bulk` | นำเข้าหลายรายการ |
| PUT | `/v1/peas/:id` | แก้ไข |
| DELETE | `/v1/peas/:id` | ลบ |

### POST /v1/peas - Request Body
```json
{
  "shortname": "กฟย.ใหม่",
  "fullname": "การไฟฟ้าส่วนภูมิภาค ใหม่",
  "operationId": "1"
}
```

### POST /v1/peas/bulk - Request Body
```json
[
  { "shortname": "กฟย.1", "fullname": "การไฟฟ้า 1", "operationId": "1" },
  { "shortname": "กฟย.2", "fullname": "การไฟฟ้า 2", "operationId": "1" }
]
```

---

## Operation Centers CRUD

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/v1/operation-centers` | ดึงทั้งหมด |
| GET | `/v1/operation-centers/:id` | ดึงตาม ID |
| POST | `/v1/operation-centers` | สร้างใหม่ |
| PUT | `/v1/operation-centers/:id` | แก้ไข |
| DELETE | `/v1/operation-centers/:id` | ลบ |

### POST /v1/operation-centers - Request Body
```json
{ "name": "จุดรวมงาน 2" }
```

---

# Phase 4: Dashboard APIs

## GET /v1/dashboard/summary

สรุปภาพรวม Dashboard

### Request
```
GET /v1/dashboard/summary?year=2026&month=1&teamId=all&jobTypeId=all
```

### Query Parameters
| Parameter | Type | Description |
|-----------|------|-------------|
| year | number | ปี |
| month | number | เดือน (1-12) |
| teamId | string | ID ทีม หรือ "all" |
| jobTypeId | string | ID ประเภทงาน หรือ "all" |

### Response (200 OK)
```json
{
  "success": true,
  "data": {
    "totalTasks": 500,
    "totalJobTypes": 5,
    "totalFeeders": 25,
    "topTeam": {
      "id": "1",
      "name": "ทีม A",
      "count": 150
    }
  }
}
```

---

## GET /v1/dashboard/top-jobs

Top 10 รายละเอียดงานที่ทำบ่อย

### Request
```
GET /v1/dashboard/top-jobs?year=2026&limit=10
```

### Query Parameters
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| year | number | ปีปัจจุบัน | ปี |
| limit | number | 10 | จำนวนผลลัพธ์ |
| month | number | - | เดือน |
| teamId | string | - | กรองตามทีม |
| jobTypeId | string | - | กรองตามประเภทงาน |

### Response (200 OK)
```json
{
  "success": true,
  "data": [
    {
      "id": "1",
      "name": "ตัดต้นไม้",
      "count": 75,
      "jobTypeName": "ซ่อมบำรุง"
    },
    {
      "id": "2",
      "name": "เปลี่ยนฟิวส์",
      "count": 45,
      "jobTypeName": "ซ่อมบำรุง"
    }
  ]
}
```

---

## GET /v1/dashboard/top-feeders

Top 10 ฟีดเดอร์ที่มีงานมากที่สุด

### Request
```
GET /v1/dashboard/top-feeders?year=2026&limit=10
```

### Response (200 OK)
```json
{
  "success": true,
  "data": [
    {
      "id": "1",
      "code": "SKT-01",
      "stationName": "สถานีไฟฟ้าย่อยสุขาภิบาล",
      "count": 50
    }
  ]
}
```

---

## GET /v1/dashboard/feeder-matrix

Matrix แสดงงานแต่ละประเภทในฟีดเดอร์

### Request
```
GET /v1/dashboard/feeder-matrix?feederId=1&year=2026
```

### Query Parameters
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| feederId | string | Yes | ID ฟีดเดอร์ |
| year | number | No | ปี |
| month | number | No | เดือน |

### Response (200 OK)
```json
{
  "success": true,
  "data": {
    "feederId": "1",
    "feederCode": "SKT-01",
    "stationName": "สถานีไฟฟ้าย่อยสุขาภิบาล",
    "totalCount": 50,
    "jobDetails": [
      {
        "id": "1",
        "name": "ตัดต้นไม้",
        "count": 25,
        "jobTypeName": "ซ่อมบำรุง"
      },
      {
        "id": "2",
        "name": "เปลี่ยนฟิวส์",
        "count": 15,
        "jobTypeName": "ซ่อมบำรุง"
      }
    ]
  }
}
```

---

## GET /v1/dashboard/stats

สถิติขั้นสูงพร้อมข้อมูลสำหรับ Charts

### Request
```
GET /v1/dashboard/stats?startDate=2026-01-01&endDate=2026-12-31
```

### Query Parameters
| Parameter | Type | Description |
|-----------|------|-------------|
| startDate | string | วันที่เริ่มต้น (ISO) |
| endDate | string | วันที่สิ้นสุด (ISO) |
| teamId | string | กรองตามทีม |
| feederId | string | กรองตามฟีดเดอร์ |

### Response (200 OK)
```json
{
  "success": true,
  "data": {
    "summary": {
      "totalTasks": 500,
      "activeTeams": 5,
      "topJobType": "ซ่อมบำรุง",
      "topFeeder": "SKT-01"
    },
    "charts": {
      "tasksByFeeder": [
        { "name": "SKT-01", "value": 50 },
        { "name": "SKT-02", "value": 35 }
      ],
      "tasksByJobType": [
        { "name": "ซ่อมบำรุง", "value": 200 },
        { "name": "ตรวจสอบ", "value": 150 }
      ],
      "tasksByTeam": [
        { "name": "ทีม A", "value": 150 },
        { "name": "ทีม B", "value": 120 }
      ],
      "tasksByDate": [
        { "date": "2026-01-01", "count": 10 },
        { "date": "2026-01-02", "count": 15 }
      ]
    }
  }
}
```

---

# Phase 5: Authentication APIs

## POST /v1/auth/login

เข้าสู่ระบบ

### Request
```
POST /v1/auth/login
Content-Type: application/json
```

### Request Body
```json
{
  "username": "admin",
  "password": "password123"
}
```

### Response (200 OK)
```json
{
  "success": true,
  "data": {
    "accessToken": "eyJhbGciOiJIUzI1NiIs...",
    "refreshToken": "eyJhbGciOiJIUzI1NiIs...",
    "expiresIn": 3600,
    "user": {
      "id": "1",
      "username": "admin",
      "role": "admin"
    }
  }
}
```

---

## POST /v1/auth/logout

ออกจากระบบ

### Request
```
POST /v1/auth/logout
Authorization: Bearer <accessToken>
```

### Response (200 OK)
```json
{
  "success": true
}
```

---

## POST /v1/auth/refresh

รีเฟรช Access Token

### Request
```
POST /v1/auth/refresh
Content-Type: application/json
```

### Request Body
```json
{
  "refreshToken": "eyJhbGciOiJIUzI1NiIs..."
}
```

### Response (200 OK)
```json
{
  "success": true,
  "data": {
    "accessToken": "eyJhbGciOiJIUzI1NiIs...",
    "expiresIn": 3600
  }
}
```

---

## GET /v1/auth/me

ดึงข้อมูล User ปัจจุบัน

### Request
```
GET /v1/auth/me
Authorization: Bearer <accessToken>
```

### Response (200 OK)
```json
{
  "success": true,
  "data": {
    "id": "1",
    "username": "admin",
    "role": "admin",
    "teamId": "1",
    "team": { "id": "1", "name": "ทีม A" }
  }
}
```

---

# TypeScript Types

## Common Types

```typescript
// Standard API Response
interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: {
    code: string;
    message: string;
    details?: Record<string, unknown>;
  };
  meta?: {
    page?: number;
    limit?: number;
    total?: number;
  };
}
```

## Task Daily Types

```typescript
interface CreateTaskDailyRequest {
  workDate: string;           // YYYY-MM-DD
  teamId: string;
  jobTypeId: string;
  jobDetailId: string;
  feederId?: string;
  numPole?: string;
  deviceCode?: string;
  detail?: string;
  urlsBefore: string[];
  urlsAfter: string[];
  latitude?: number;
  longitude?: number;
}

interface UpdateTaskDailyRequest extends Partial<CreateTaskDailyRequest> {}

interface TaskDailyResponse {
  id: string;
  workDate: string;
  teamId: string;
  jobTypeId: string;
  jobDetailId: string;
  feederId: string | null;
  numPole: string | null;
  deviceCode: string | null;
  detail: string | null;
  urlsBefore: string[];
  urlsAfter: string[];
  latitude: number | null;
  longitude: number | null;
  team: { id: string; name: string };
  jobType: { id: string; name: string };
  jobDetail: { id: string; name: string };
  feeder?: {
    id: string;
    code: string;
    station: {
      name: string;
      operationCenter: { name: string };
    };
  };
  createdAt: string;
  updatedAt: string;
  deletedAt: string | null;
}
```

## Upload Types

```typescript
interface UploadResponse {
  url: string;
  fileName: string;
  originalName: string;
  size: number;
  type: string;
}
```

## Dashboard Types

```typescript
interface DashboardSummary {
  totalTasks: number;
  totalJobTypes: number;
  totalFeeders: number;
  topTeam: {
    id: string;
    name: string;
    count: number;
  } | null;
}

interface TopJobDetail {
  id: string;
  name: string;
  count: number;
  jobTypeName: string;
}

interface TopFeeder {
  id: string;
  code: string;
  stationName: string;
  count: number;
}
```

---

# Backend Technology Recommendations

## Option 1: Express.js + Prisma (แนะนำสำหรับความคุ้นเคย)
- ย้ายจาก Server Actions ได้ง่าย
- ใช้ Prisma ORM เดิมได้

## Option 2: Fastify + Prisma (แนะนำสำหรับ Performance)
- เร็วกว่า Express
- TypeScript support ดี

## Option 3: NestJS + Prisma (แนะนำสำหรับ Enterprise)
- Modular architecture
- Built-in validation

---

# Migration Strategy

## Step 1: สร้าง Backend API Server
- Copy Prisma schema
- Copy service layer จาก `src/server/services/`

## Step 2: สร้าง API Routes
- สร้าง routes ที่เรียก services
- ใช้ response format เดียวกัน

## Step 3: แก้ไข Frontend
- เปลี่ยน Server Actions เป็น API calls
- Update React Query hooks

## Step 4: ลบ Server Actions
- เมื่อ migrate ครบแล้ว ลบ Server Actions

---

# Phase 6: PDF Report Generation APIs (Backend Go)

## Overview

PDF Generation จะย้ายไปทำที่ Backend (Go) เพื่อคุณภาพที่ดีกว่า:
- รองรับ PDF/A มาตรฐาน
- ภาษาไทยจัดการง่ายกว่า (ไม่ต้อง embed font ใน JS bundle)
- Performance ดีกว่า ไม่ block browser
- Caching ได้ (เก็บ PDF ที่สร้างแล้วใน S3)

### Go Libraries ที่แนะนำ

| Library | ข้อดี | ใช้งาน |
|---------|-------|--------|
| `github.com/go-pdf/fpdf` | ง่าย, รองรับ Unicode | Simple reports |
| `github.com/johnfercher/maroto/v2` | Template-based, ใช้ง่ายมาก | Complex layouts |
| `github.com/unidoc/unipdf` | Commercial, feature ครบ | Enterprise |

---

## GET /v1/reports/tasks/pdf

สร้าง PDF รายงานงานประจำเดือน

### Request
```
GET /v1/reports/tasks/pdf?year=2026&month=1&teamId=all&format=stream
```

### Query Parameters
| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| year | number | Yes | - | ปี (ค.ศ.) |
| month | number | Yes | - | เดือน (1-12) |
| teamId | string | No | "all" | ID ทีม หรือ "all" สำหรับทุกทีม |
| format | string | No | "stream" | `stream` = ดาวน์โหลดทันที, `url` = return S3 URL |

### Response - format=stream (200 OK)
```
Content-Type: application/pdf
Content-Disposition: attachment; filename="WorkReport_2026-01_TeamA.pdf"

[PDF Binary Data]
```

### Response - format=url (200 OK)
```json
{
  "success": true,
  "data": {
    "url": "https://r2.example.com/reports/WorkReport_2026-01_TeamA.pdf",
    "expiresAt": "2026-01-25T10:30:00.000Z",
    "fileName": "WorkReport_2026-01_TeamA.pdf",
    "fileSize": 125430,
    "cachedAt": "2026-01-24T10:30:00.000Z"
  }
}
```

### Error Response (400 Bad Request)
```json
{
  "success": false,
  "error": {
    "code": "INVALID_PARAMETERS",
    "message": "กรุณาระบุปีและเดือนที่ถูกต้อง"
  }
}
```

### Error Response (404 Not Found)
```json
{
  "success": false,
  "error": {
    "code": "NO_DATA",
    "message": "ไม่พบข้อมูลงานในช่วงเวลาที่เลือก"
  }
}
```

---

## GET /v1/reports/tasks/pdf/preview

Preview ข้อมูลก่อนสร้าง PDF (สำหรับแสดง summary)

### Request
```
GET /v1/reports/tasks/pdf/preview?year=2026&month=1&teamId=all
```

### Response (200 OK)
```json
{
  "success": true,
  "data": {
    "year": 2026,
    "month": 1,
    "monthName": "มกราคม",
    "totalTasks": 150,
    "teams": [
      {
        "id": "1",
        "name": "ทีม A",
        "taskCount": 75
      },
      {
        "id": "2",
        "name": "ทีม B",
        "taskCount": 50
      },
      {
        "id": "3",
        "name": "ทีม C",
        "taskCount": 25
      }
    ],
    "estimatedPages": 5,
    "estimatedFileSize": "~500KB"
  }
}
```

---

## POST /v1/reports/tasks/pdf/batch

สร้าง PDF หลายทีมพร้อมกัน (Async job)

### Request
```
POST /v1/reports/tasks/pdf/batch
Content-Type: application/json
```

### Request Body
```json
{
  "year": 2026,
  "month": 1,
  "teamIds": ["1", "2", "3"],
  "mode": "separate",
  "notifyUrl": "https://webhook.example.com/pdf-ready"
}
```

### Field Specifications
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| year | number | Yes | ปี (ค.ศ.) |
| month | number | Yes | เดือน (1-12) |
| teamIds | string[] | Yes | รายการ ID ทีม (ว่าง = ทุกทีม) |
| mode | string | No | `separate` = แยกไฟล์, `combined` = รวมไฟล์เดียว |
| notifyUrl | string | No | Webhook URL สำหรับแจ้งเตือนเมื่อเสร็จ |

### Response (202 Accepted)
```json
{
  "success": true,
  "data": {
    "jobId": "pdf-job-abc123",
    "status": "processing",
    "totalFiles": 3,
    "estimatedTime": "30 seconds",
    "statusUrl": "/v1/reports/jobs/pdf-job-abc123"
  }
}
```

---

## GET /v1/reports/jobs/:jobId

ตรวจสอบสถานะ batch job

### Request
```
GET /v1/reports/jobs/pdf-job-abc123
```

### Response - Processing (200 OK)
```json
{
  "success": true,
  "data": {
    "jobId": "pdf-job-abc123",
    "status": "processing",
    "progress": {
      "completed": 1,
      "total": 3,
      "percentage": 33
    },
    "files": [
      {
        "teamId": "1",
        "teamName": "ทีม A",
        "status": "completed",
        "url": "https://r2.example.com/reports/WorkReport_2026-01_TeamA.pdf"
      },
      {
        "teamId": "2",
        "teamName": "ทีม B",
        "status": "processing",
        "url": null
      },
      {
        "teamId": "3",
        "teamName": "ทีม C",
        "status": "pending",
        "url": null
      }
    ]
  }
}
```

### Response - Completed (200 OK)
```json
{
  "success": true,
  "data": {
    "jobId": "pdf-job-abc123",
    "status": "completed",
    "completedAt": "2026-01-24T10:31:00.000Z",
    "files": [
      {
        "teamId": "1",
        "teamName": "ทีม A",
        "status": "completed",
        "url": "https://r2.example.com/reports/WorkReport_2026-01_TeamA.pdf",
        "fileSize": 125430
      },
      {
        "teamId": "2",
        "teamName": "ทีม B",
        "status": "completed",
        "url": "https://r2.example.com/reports/WorkReport_2026-01_TeamB.pdf",
        "fileSize": 98200
      },
      {
        "teamId": "3",
        "teamName": "ทีม C",
        "status": "completed",
        "url": "https://r2.example.com/reports/WorkReport_2026-01_TeamC.pdf",
        "fileSize": 45600
      }
    ],
    "zipUrl": "https://r2.example.com/reports/WorkReport_2026-01_AllTeams.zip"
  }
}
```

---

## PDF Report Specifications

### รูปแบบเอกสาร
- **Orientation**: Landscape (A4)
- **Font**: THSarabunNew (รองรับภาษาไทย)
- **Header**: ชื่อรายงาน + เดือน/ปี (พุทธศักราช)
- **Footer**: หมายเลขหน้า + "สร้างโดยระบบ HotlineS3"

### ตารางข้อมูล
| Column | Width | Description |
|--------|-------|-------------|
| # | 15mm | ลำดับ |
| วันที่ | 25mm | DD/MM/YYYY (พ.ศ.) |
| รายละเอียดงาน | 75mm | ชื่องานจาก JobDetail |
| ฟีดเดอร์ | 25mm | รหัสฟีดเดอร์ |
| เบอร์เสา | 25mm | หมายเลขเสา |
| รหัสอุปกรณ์ | 30mm | Device code |
| เพิ่มเติม | 75mm | หมายเหตุ |

### Styling
- **Header Row**: สีน้ำเงิน (#2563EB), ตัวอักษรขาว
- **Body Rows**: Zebra striping (สลับสีเทาอ่อน)
- **Font Size**: Header 12pt, Body 8pt
- **Summary**: แสดงจำนวนงานทั้งหมดท้ายรายงาน

### Caching Strategy
- PDF ที่สร้างแล้วเก็บใน S3/R2
- Cache key: `reports/{year}/{month}/{teamId}.pdf`
- TTL: 24 ชั่วโมง (หรือจนกว่าจะมีข้อมูลใหม่)
- Invalidate cache เมื่อมี Task ใหม่ในเดือนนั้น

---

## TypeScript Types (Frontend)

```typescript
// PDF Report Types
interface PDFReportRequest {
  year: number;
  month: number;
  teamId?: string;
  format?: 'stream' | 'url';
}

interface PDFReportUrlResponse {
  url: string;
  expiresAt: string;
  fileName: string;
  fileSize: number;
  cachedAt: string;
}

interface PDFPreviewResponse {
  year: number;
  month: number;
  monthName: string;
  totalTasks: number;
  teams: Array<{
    id: string;
    name: string;
    taskCount: number;
  }>;
  estimatedPages: number;
  estimatedFileSize: string;
}

interface PDFBatchRequest {
  year: number;
  month: number;
  teamIds: string[];
  mode?: 'separate' | 'combined';
  notifyUrl?: string;
}

interface PDFJobStatus {
  jobId: string;
  status: 'pending' | 'processing' | 'completed' | 'failed';
  progress?: {
    completed: number;
    total: number;
    percentage: number;
  };
  files: Array<{
    teamId: string;
    teamName: string;
    status: 'pending' | 'processing' | 'completed' | 'failed';
    url: string | null;
    fileSize?: number;
    error?: string;
  }>;
  completedAt?: string;
  zipUrl?: string;
}
```

---

## Go Backend Implementation Notes

### Folder Structure
```
backend/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── handler/
│   │   └── report_handler.go
│   ├── service/
│   │   └── pdf_service.go
│   ├── repository/
│   │   └── task_repository.go
│   └── pdf/
│       ├── generator.go
│       ├── templates/
│       │   └── monthly_report.go
│       └── fonts/
│           └── THSarabunNew.ttf
├── pkg/
│   └── storage/
│       └── s3.go
└── go.mod
```

### Key Dependencies
```go
// go.mod
require (
    github.com/johnfercher/maroto/v2 v2.0.0
    github.com/aws/aws-sdk-go-v2 v1.x.x
    github.com/jackc/pgx/v5 v5.x.x  // PostgreSQL driver
)
```

### Example Service Interface
```go
type PDFService interface {
    GenerateMonthlyReport(ctx context.Context, year, month int, teamID string) ([]byte, error)
    GetCachedReport(ctx context.Context, year, month int, teamID string) (string, error)
    CreateBatchJob(ctx context.Context, req BatchRequest) (string, error)
    GetJobStatus(ctx context.Context, jobID string) (*JobStatus, error)
}
```

---

# Files to Reference

| File | Description |
|------|-------------|
| `prisma/schema.prisma` | Database schema |
| `src/lib/actions/*.ts` | Server Actions (ใช้เป็น reference) |
| `src/server/services/*.ts` | Business logic (copy ไปใช้ได้) |
| `src/types/*.ts` | TypeScript types |
| `src/hooks/useQueries.ts` | React Query hooks |
