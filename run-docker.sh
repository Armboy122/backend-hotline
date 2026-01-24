#!/bin/bash

# สคริปต์สำหรับรัน backend-hotlines3 ด้วย Docker

echo "🚀 Starting backend-hotlines3..."

# หยุดและลบ container เก่า
docker stop hotlines3-api 2>/dev/null
docker rm hotlines3-api 2>/dev/null

# รัน container ใหม่
docker run -d \
  --name hotlines3-api \
  --restart unless-stopped \
  -p 8080:8080 \
  -e TZ=Asia/Bangkok \
  -v $(pwd)/config.yaml:/app/config.yaml:ro \
  hotlines3-api:latest

if [ $? -eq 0 ]; then
    echo "✅ Container started successfully!"
    echo ""
    echo "📊 Container status:"
    docker ps | grep hotlines3-api
    echo ""
    echo "📝 Logs (last 10 lines):"
    sleep 2
    docker logs hotlines3-api --tail 10
    echo ""
    echo "🔗 API is running at: http://localhost:8080"
    echo "🏥 Health check: http://localhost:8080/health"
else
    echo "❌ Failed to start container!"
    exit 1
fi
