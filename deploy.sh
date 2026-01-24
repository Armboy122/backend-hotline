#!/bin/bash

echo "🚀 Starting deployment..."

# Stop existing container
echo "📦 Stopping existing container..."
docker stop hotlines3-api 2>/dev/null || true
docker rm hotlines3-api 2>/dev/null || true

# Build new image
echo "🔨 Building Docker image..."
docker build -t hotlines3-api:latest .

if [ $? -ne 0 ]; then
    echo "❌ Build failed!"
    exit 1
fi

# Start container
echo "🏃 Starting container..."
docker-compose up -d

if [ $? -ne 0 ]; then
    echo "❌ Failed to start container!"
    exit 1
fi

echo "✅ Deployment completed successfully!"
echo "📊 Container status:"
docker ps | grep hotlines3-api

echo ""
echo "📝 Logs:"
docker logs hotlines3-api --tail 20
