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
if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose is not installed. Please install Docker Compose first."
    exit 1
fi

# Create .env file if it doesn't exist
if [ ! -f .env ]; then
    echo "📝 Creating .env file from template..."
    cp env.example .env
    echo "✅ Created .env file. Please review and update as needed."
fi

echo "🐘 Starting infrastructure services (Databases, Redis, RabbitMQ)..."
docker-compose up -d postgres-auth postgres-contract postgres-payment postgres-dispute postgres-audit redis rabbitmq

echo "⏳ Waiting for databases to be ready..."
sleep 10

# Check database connections
echo "🔍 Checking database connections..."
for i in {1..30}; do
    if docker-compose exec -T postgres-auth pg_isready -U postgres > /dev/null 2>&1 &&
       docker-compose exec -T postgres-contract pg_isready -U postgres > /dev/null 2>&1 &&
       docker-compose exec -T postgres-payment pg_isready -U postgres > /dev/null 2>&1 &&
       docker-compose exec -T postgres-dispute pg_isready -U postgres > /dev/null 2>&1 &&
       docker-compose exec -T postgres-audit pg_isready -U postgres > /dev/null 2>&1; then
        echo "✅ All databases are ready!"
        break
    fi
    echo "⏳ Waiting for databases... ($i/30)"
    sleep 2
done

echo "🚀 Starting all microservices..."
docker-compose up -d

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
echo "├── API Gateway:      http://localhost:8080"
echo "├── Auth Service:     http://localhost:3001"
echo "├── Contract Service: http://localhost:3002"
echo "├── Payment Service:  http://localhost:3003"
echo "├── Dispute Service:  http://localhost:3004"
echo "├── Notification:     http://localhost:8081"
echo "├── Audit Service:    http://localhost:8082"
echo "├── Redis:           redis://localhost:6379"
echo "└── RabbitMQ:        http://localhost:15672 (admin/admin)"
echo ""
echo "🔗 API Gateway Health: http://localhost:8080/api/v1/health"
echo "📚 WebSocket Test:     ws://localhost:8081/ws?userId=test&clientId=dev"
echo ""
echo "🛠️  To view logs: docker-compose logs -f [service-name]"
echo "⏹️  To stop:      docker-compose down"
echo "🔄 To restart:    docker-compose restart [service-name]"
echo ""
echo "Happy coding! 🚀"

