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

2. แก้ไขค่าใน `config.yaml`:
```yaml
database:
  host: localhost
  port: 5432
  user: postgres
  password: your-password
  dbname: hotlines3
```

3. รัน PostgreSQL:
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
- `GET /api/operation-centers` - รายการทั้งหมด
- `GET /api/operation-centers/:id` - ดูรายละเอียด
- `POST /api/operation-centers` - สร้างใหม่
- `PUT /api/operation-centers/:id` - แก้ไข
- `DELETE /api/operation-centers/:id` - ลบ

#### PEAs (การไฟฟ้าส่วนภูมิภาค)
- `GET /api/peas` - รายการทั้งหมด
- `GET /api/peas/:id` - ดูรายละเอียด
- `POST /api/peas` - สร้างใหม่
- `POST /api/peas/bulk` - สร้างหลายรายการ
- `PUT /api/peas/:id` - แก้ไข
- `DELETE /api/peas/:id` - ลบ

#### Stations (สถานี)
- `GET /api/stations` - รายการทั้งหมด
- `GET /api/stations/:id` - ดูรายละเอียด
- `POST /api/stations` - สร้างใหม่
- `PUT /api/stations/:id` - แก้ไข
- `DELETE /api/stations/:id` - ลบ

#### Feeders (สายป้อน)
- `GET /api/feeders` - รายการทั้งหมด
- `GET /api/feeders/:id` - ดูรายละเอียด
- `POST /api/feeders` - สร้างใหม่
- `PUT /api/feeders/:id` - แก้ไข
- `DELETE /api/feeders/:id` - ลบ

#### Job Types (ประเภทงาน)
- `GET /api/job-types` - รายการทั้งหมด
- `GET /api/job-types/:id` - ดูรายละเอียด
- `POST /api/job-types` - สร้างใหม่
- `PUT /api/job-types/:id` - แก้ไข
- `DELETE /api/job-types/:id` - ลบ

#### Job Details (รายละเอียดงาน)
- `GET /api/job-details` - รายการทั้งหมด
- `GET /api/job-details/:id` - ดูรายละเอียด
- `POST /api/job-details` - สร้างใหม่
- `PUT /api/job-details/:id` - แก้ไข
- `DELETE /api/job-details/:id` - ลบ

#### Teams (ทีมงาน)
- `GET /api/teams` - รายการทั้งหมด
- `GET /api/teams/:id` - ดูรายละเอียด
- `POST /api/teams` - สร้างใหม่
- `PUT /api/teams/:id` - แก้ไข
- `DELETE /api/teams/:id` - ลบ

### Task Daily APIs (งานประจำวัน)
- `GET /api/tasks` - รายการทั้งหมด (รองรับ query: year, month, teamId)
- `GET /api/tasks/by-team` - รายการ grouped by team (query: year, month)
- `GET /api/tasks/:id` - ดูรายละเอียด
- `POST /api/tasks` - สร้างใหม่
- `PUT /api/tasks/:id` - แก้ไข
- `DELETE /api/tasks/:id` - ลบ

### Dashboard APIs
- `GET /api/dashboard/summary` - สรุปภาพรวม
- `GET /api/dashboard/top-jobs` - งานที่ทำบ่อยที่สุด
- `GET /api/dashboard/top-feeders` - สายป้อนที่ทำงานบ่อยที่สุด
- `GET /api/dashboard/stats` - สถิติต่างๆ สำหรับกราฟ

## ตัวอย่างการใช้งาน

### สร้าง Operation Center
```bash
curl -X POST http://localhost:8080/api/operation-centers \
  -H "Content-Type: application/json" \
  -d '{
    "name": "ศูนย์ปฏิบัติการภาคเหนือ",
    "code": "NORTH"
  }'
```

### ดูรายการ Teams
```bash
curl http://localhost:8080/api/teams
```

### ดูงานตามเดือน
```bash
curl "http://localhost:8080/api/tasks?year=2024&month=1"
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
