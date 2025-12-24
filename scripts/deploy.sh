#!/bin/bash
# ===========================================
# Swiftlet Backend - Deploy Script
# ===========================================
# Run this on the server to deploy updates
# Usage: ./deploy.sh [version]

set -e

VERSION=${1:-latest}
APP_DIR="/opt/swiftlet"
COMPOSE_FILE="docker-compose.prod.yml"

echo "🚀 Deploying Swiftlet Backend v$VERSION"
echo "======================================="

cd $APP_DIR

# Pull latest images
echo "📦 Pulling Docker images..."
docker-compose -f $COMPOSE_FILE pull

# Stop and remove old containers
echo "🛑 Stopping old containers..."
docker-compose -f $COMPOSE_FILE down

# Start new containers
echo "▶️  Starting new containers..."
docker-compose -f $COMPOSE_FILE up -d

# Wait for health check
echo "⏳ Waiting for health check..."
sleep 10

# Check health
if curl -sf http://localhost:8080/health > /dev/null; then
    echo "✅ Health check passed!"
else
    echo "❌ Health check failed!"
    docker-compose -f $COMPOSE_FILE logs --tail=50 backend
    exit 1
fi

# Cleanup
echo "🧹 Cleaning up old images..."
docker image prune -f

echo ""
echo "✅ Deployment complete!"
echo "📊 Status:"
docker-compose -f $COMPOSE_FILE ps
