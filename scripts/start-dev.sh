#!/bin/bash

# Development startup script for marketplace microservices

set -e

echo "🚀 Starting Marketplace Microservices Development Environment"
echo "=============================================================="

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker first."
    exit 1
fi

# Check if Docker Compose is available
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed. Please install Docker first."
    exit 1
fi

if ! docker compose version &> /dev/null; then
    echo "❌ Docker Compose is not available. Please ensure Docker Desktop is running."
    exit 1
fi

# Create .env file if it doesn't exist
if [ ! -f .env ]; then
    echo "📝 Creating .env file from template..."
    cp env.example .env
    echo "✅ Created .env file. Please review and update as needed."
fi

echo "🔨 Building Docker images..."
docker compose build

echo "🐘 Starting infrastructure services (Databases, MongoDB, Redis, RabbitMQ)..."
docker compose up -d postgres-auth postgres-contract postgres-payment postgres-dispute mongodb-audit redis rabbitmq

echo "⏳ Waiting for databases to be ready..."
sleep 10

# Check database connections
echo "🔍 Checking database connections..."
for i in {1..30}; do
    if docker compose exec -T postgres-auth pg_isready -U postgres > /dev/null 2>&1 &&
       docker compose exec -T postgres-contract pg_isready -U postgres > /dev/null 2>&1 &&
       docker compose exec -T postgres-payment pg_isready -U postgres > /dev/null 2>&1 &&
       docker compose exec -T postgres-dispute pg_isready -U postgres > /dev/null 2>&1 &&
       docker compose exec -T mongodb-audit mongosh --eval "db.adminCommand('ping')" > /dev/null 2>&1; then
        echo "✅ All databases are ready!"
        break
    fi
    echo "⏳ Waiting for databases... ($i/30)"
    sleep 2
done

echo "🚀 Starting all microservices..."
docker compose up -d

echo "⏳ Waiting for services to start..."
sleep 15

# Health check function
check_health() {
    local service_name=$1
    local url=$2
    local max_attempts=30
    local attempt=1

    while [ $attempt -le $max_attempts ]; do
        if curl -s "$url" > /dev/null 2>&1; then
            echo "✅ $service_name is healthy"
            return 0
        fi
        echo "⏳ Waiting for $service_name... ($attempt/$max_attempts)"
        sleep 2
        ((attempt++))
    done
    echo "❌ $service_name failed to start"
    return 1
}

# Check service health
echo "🔍 Checking service health..."
check_health "API Gateway" "http://localhost:8080/api/v1/health"
check_health "Auth Service" "http://localhost:3001/health"
check_health "Contract Service" "http://localhost:3002/health"
check_health "Payment Service" "http://localhost:3003/health"
check_health "Dispute Service" "http://localhost:3004/health"
check_health "Notification Service" "http://localhost:8081/health"
check_health "Audit Service" "http://localhost:8082/health"

echo ""
echo "🎉 All services are running!"
echo ""
echo "📋 Service Status:"
echo "├── API Gateway:      http://localhost:8080 (HTTP → gRPC)"
echo "├── Auth Service:     http://localhost:3001 | gRPC: localhost:50051"
echo "├── Contract Service: http://localhost:3002 | gRPC: localhost:50052"
echo "├── Payment Service:  http://localhost:3003 | gRPC: localhost:50053"
echo "├── Dispute Service:  http://localhost:3004 | gRPC: localhost:50054"
echo "├── Notification:     http://localhost:8081 | gRPC: localhost:50055"
echo "├── Audit Service:    http://localhost:8082 | gRPC: localhost:50056"
echo "├── Redis:           redis://localhost:6379"
echo "├── MongoDB:         mongodb://localhost:27017"
echo "└── RabbitMQ:        http://localhost:15672 (admin/admin)"
echo ""
echo "🔗 API Gateway Health: http://localhost:8080/api/v1/health"
echo "📚 WebSocket Test:     ws://localhost:8081/ws?userId=test&clientId=dev"
echo "⚙️  gRPC Endpoints:    Each service exposes gRPC on ports 50051-50056"
echo ""
echo "🛠️  Development Commands:"
echo "├── Generate proto:   make proto-gen"
echo "├── View logs:        docker compose logs -f [service-name]"
echo "├── Stop services:    docker compose down"
echo "└── Restart service:  docker compose restart [service-name]"
echo ""
echo "Happy coding! 🚀"

