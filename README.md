# Backend Hotlines3 - Golang API

Backend API สำหรับระบบจัดการงานประจำวัน Hotlines

## เทคโนโลยีที่ใช้

- **Golang** 1.23+
- **Gin** - Web framework
- **GORM** - ORM สำหรับ PostgreSQL
- **Viper** - Configuration management (ใช้ `config.yaml`)
- **PostgreSQL** - Database

## โครงสร้างโปรเจกต์

```
backend-hotlines3/
├── config.yaml              # Configuration file (Viper)
├── main.go                  # Entry point
├── internal/
│   ├── config/             # Config loader
│   ├── database/           # Database connection & migrations
│   ├── models/             # GORM models
│   ├── handlers/           # HTTP handlers
│   └── router/             # Route setup
└── go.mod
```

## การติดตั้ง

1. ติดตั้ง dependencies:
```bash
go mod tidy
```

2. สร้างไฟล์ `.env` จาก `.env.example` แล้วเติมค่า secrets ที่จำเป็น

3. แก้ไขค่าใน `config.yaml` ถ้าต้องการปรับค่าเซิร์ฟเวอร์หรือฐานข้อมูล:
```yaml
database:
  host: localhost
  port: 5432
  user: postgres
  password: your-password
  dbname: hotlines3
```

4. รัน PostgreSQL:
```bash
# ตัวอย่างใช้ Docker
docker run --name postgres-hotlines \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=hotlines3 \
  -p 5432:5432 \
  -d postgres:16
```

## การรัน

```bash
go run main.go
```

Server จะรันที่ `http://localhost:8080`

## API Endpoints

### Health Check
- `GET /health` - ตรวจสอบสถานะเซิร์ฟเวอร์

### Master Data APIs

#### Operation Centers (ศูนย์ปฏิบัติการ)
- `GET /v1/operation-centers` - รายการทั้งหมด
- `GET /v1/operation-centers/:id` - ดูรายละเอียด
- `POST /v1/operation-centers` - สร้างใหม่
- `PUT /v1/operation-centers/:id` - แก้ไข
- `DELETE /v1/operation-centers/:id` - ลบ

#### PEAs (การไฟฟ้าส่วนภูมิภาค)
- `GET /v1/peas` - รายการทั้งหมด
- `GET /v1/peas/:id` - ดูรายละเอียด
- `POST /v1/peas` - สร้างใหม่
- `POST /v1/peas/bulk` - สร้างหลายรายการ
- `PUT /v1/peas/:id` - แก้ไข
- `DELETE /v1/peas/:id` - ลบ

#### Stations (สถานี)
- `GET /v1/stations` - รายการทั้งหมด
- `GET /v1/stations/:id` - ดูรายละเอียด
- `POST /v1/stations` - สร้างใหม่
- `PUT /v1/stations/:id` - แก้ไข
- `DELETE /v1/stations/:id` - ลบ

#### Feeders (สายป้อน)
- `GET /v1/feeders` - รายการทั้งหมด
- `GET /v1/feeders/:id` - ดูรายละเอียด
- `POST /v1/feeders` - สร้างใหม่
- `PUT /v1/feeders/:id` - แก้ไข
- `DELETE /v1/feeders/:id` - ลบ

#### Job Types (ประเภทงาน)
- `GET /v1/job-types` - รายการทั้งหมด
- `GET /v1/job-types/:id` - ดูรายละเอียด
- `POST /v1/job-types` - สร้างใหม่
- `PUT /v1/job-types/:id` - แก้ไข
- `DELETE /v1/job-types/:id` - ลบ

#### Job Details (รายละเอียดงาน)
- `GET /v1/job-details` - รายการทั้งหมด
- `GET /v1/job-details/:id` - ดูรายละเอียด
- `POST /v1/job-details` - สร้างใหม่
- `PUT /v1/job-details/:id` - แก้ไข
- `DELETE /v1/job-details/:id` - ลบ

#### Teams (ทีมงาน)
- `GET /v1/teams` - รายการทั้งหมด
- `GET /v1/teams/:id` - ดูรายละเอียด
- `POST /v1/teams` - สร้างใหม่
- `PUT /v1/teams/:id` - แก้ไข
- `DELETE /v1/teams/:id` - ลบ

### Task Daily APIs (งานประจำวัน)
- `GET /v1/tasks` - รายการทั้งหมด (รองรับ query: year, month, teamId)
- `GET /v1/tasks/by-team` - รายการ grouped by team (query: year, month)
- `GET /v1/tasks/:id` - ดูรายละเอียด
- `POST /v1/tasks` - สร้างใหม่
- `PUT /v1/tasks/:id` - แก้ไข
- `DELETE /v1/tasks/:id` - ลบ

### Dashboard APIs
- `GET /v1/dashboard/summary` - สรุปภาพรวม
- `GET /v1/dashboard/top-jobs` - งานที่ทำบ่อยที่สุด
- `GET /v1/dashboard/top-feeders` - สายป้อนที่ทำงานบ่อยที่สุด
- `GET /v1/dashboard/stats` - สถิติต่างๆ สำหรับกราฟ

## ตัวอย่างการใช้งาน

### สร้าง Operation Center
```bash
curl -X POST http://localhost:8080/v1/operation-centers \
  -H "Content-Type: application/json" \
  -d '{
    "name": "ศูนย์ปฏิบัติการภาคเหนือ",
    "code": "NORTH"
  }'
```

### ดูรายการ Teams
```bash
curl http://localhost:8080/v1/teams
```

### ดูงานตามเดือน
```bash
curl "http://localhost:8080/v1/tasks?year=2024&month=1"
```

## Configuration (Viper)

โปรเจกต์นี้ใช้ **Viper** จัดการ configuration ผ่าน `config.yaml`:

```yaml
server:
  port: 8080
  mode: debug

database:
  host: localhost
  port: 5432
  user: postgres
  password: postgres
  dbname: hotlines3
```

ข้อดีของ Viper:
- อ่าน config จากหลายแหล่ง (YAML, JSON, ENV, etc.)
- Hot reload configuration
- Type-safe configuration
- Easy to test

## Database Schema

GORM จะสร้างตารางอัตโนมัติเมื่อรัน (Auto Migration):
- `operation_centers`
- `peas`
- `stations`
- `feeders`
- `job_types`
- `job_details`
- `teams`
- `task_dailies`

## Development

รันแบบ development mode:
```bash
# แก้ไข config.yaml
server:
  mode: debug

go run main.go
```

## Production

รันแบบ production mode:
```bash
# แก้ไข config.yaml
server:
  mode: release

# Build
go build -o hotlines-api main.go

# Run
./hotlines-api
```

## ที่ต้องทำต่อ

- [ ] Authentication (JWT)
- [ ] Upload API (Cloudflare R2)
- [ ] Rate limiting
- [ ] Request validation
- [ ] Unit tests
- [ ] API documentation (Swagger)
- [ ] Docker compose สำหรับ development
# backend-hotline
