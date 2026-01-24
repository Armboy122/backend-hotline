# 🚀 Backend Hotlines3 - Deployment Guide

## ข้อมูลเบื้องต้น
- **Port**: 8080
- **Docker Image**: hotlines3-api:latest
- **Config**: config.yaml

## 📋 ข้อกำหนด
- Docker Engine
- Config file: `config.yaml` (อยู่ใน root directory)

## 🔨 Build Docker Image

```bash
docker build -t hotlines3-api:latest .
```

## 🏃 รัน Container

### วิธีที่ 1: ใช้ script (แนะนำ)

```bash
./run-docker.sh
```

### วิธีที่ 2: ใช้คำสั่ง Docker โดยตรง

```bash
docker run -d \
  --name hotlines3-api \
  --restart unless-stopped \
  -p 8080:8080 \
  -e TZ=Asia/Bangkok \
  -v $(pwd)/config.yaml:/app/config.yaml:ro \
  hotlines3-api:latest
```

## 📊 ตรวจสอบสถานะ

```bash
# ดู container ที่รันอยู่
docker ps | grep hotlines3

# ดู logs
docker logs hotlines3-api

# ดู logs แบบ real-time
docker logs -f hotlines3-api

# ตรวจสอบ health
curl http://localhost:8080/health
```

## 🔄 อัพเดท/Deploy ใหม่

```bash
# หยุด container เก่า
docker stop hotlines3-api
docker rm hotlines3-api

# Build image ใหม่
docker build -t hotlines3-api:latest .

# รัน container ใหม่
./run-docker.sh
```

## 🛑 หยุด Container

```bash
docker stop hotlines3-api
docker rm hotlines3-api
```

## 🧪 ทดสอบ API

```bash
# Health check
curl http://localhost:8080/health

# Get operation centers
curl http://localhost:8080/api/operation-centers

# Get teams
curl http://localhost:8080/api/teams

# Dashboard summary
curl http://localhost:8080/api/dashboard/summary
```

## 📝 API Endpoints

- `GET /health` - Health check
- `GET /api/operation-centers` - ดูข้อมูลจุดรวมงาน
- `GET /api/peas` - ดูข้อมูลการไฟฟ้า
- `GET /api/stations` - ดูข้อมูลสถานี
- `GET /api/feeders` - ดูข้อมูลฟีดเดอร์
- `GET /api/job-types` - ดูประเภทงาน
- `GET /api/job-details` - ดูรายละเอียดงาน
- `GET /api/teams` - ดูทีมงาน
- `GET /api/tasks` - ดูงานประจำวัน
- `GET /api/dashboard/summary` - สรุปภาพรวม
- `GET /api/dashboard/top-jobs` - งานที่มีมากที่สุด
- `GET /api/dashboard/top-feeders` - Feeders ที่มีงานมากที่สุด
- `GET /api/dashboard/stats` - สถิติต่างๆ

## 🔧 Troubleshooting

### Port 8080 ถูกใช้งานอยู่
```bash
# หา process ที่ใช้ port
lsof -i :8080

# หรือหยุด container เก่า
docker stop $(docker ps -q --filter "publish=8080")
```

### Container ไม่ start
```bash
# ดู logs ละเอียด
docker logs hotlines3-api

# ตรวจสอบ config
cat config.yaml
```

## 📦 ไฟล์ที่สำคัญ

- `Dockerfile` - คำสั่งสร้าง Docker image
- `config.yaml` - Configuration file
- `run-docker.sh` - สคริปต์รัน container
- `.dockerignore` - ไฟล์ที่ไม่ต้องการใน image

## 🔐 ข้อมูลการเชื่อมต่อฐานข้อมูล

อยู่ในไฟล์ `config.yaml`:
- Host: ep-sweet-hill-a1a76thg.ap-southeast-1.aws.neon.tech
- Database: neondb
- Port: 5432
