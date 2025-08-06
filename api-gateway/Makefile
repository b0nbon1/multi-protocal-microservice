# API Gateway Makefile

.PHONY: help run build docker-build docker-run stop clean dev logs test

# Default target
help:
	@echo "🚀 API Gateway Commands:"
	@echo ""
	@echo "  make run          - Run API Gateway locally with Go"
	@echo "  make build        - Build the API Gateway binary"
	@echo "  make docker-build - Build Docker image"
	@echo "  make docker-run   - Run with Docker"
	@echo "  make dev          - Run in development mode with hot reload"
	@echo "  make stop         - Stop running containers"
	@echo "  make logs         - View container logs"
	@echo "  make clean        - Clean up containers and images"
	@echo "  make test         - Run tests"
	@echo ""

# Run locally with Go
run:
	@echo "🚀 Starting API Gateway locally..."
	@if [ ! -f .env ]; then \
		echo "📝 Creating default .env file..."; \
		echo "JWT_SECRET=your-super-secret-jwt-key-change-this-in-production" > .env; \
		echo "USER_SERVICE_URL=http://localhost:3001" >> .env; \
		echo "ORDER_SERVICE_URL=http://localhost:3002" >> .env; \
	fi
	@go mod tidy
	@echo "📍 API Gateway will be available at: http://localhost:3000"
	@go run *.go

# Build binary
build:
	@echo "🔨 Building API Gateway binary..."
	@go mod tidy
	@go build -o bin/api-gateway *.go
	@echo "✅ Binary built at: bin/api-gateway"

# Build Docker image
docker-build:
	@echo "🐳 Building Docker image..."
	@docker build -t api-gateway .
	@echo "✅ Docker image built: api-gateway"

# Run with Docker
docker-run: docker-build
	@echo "🚀 Starting API Gateway with Docker..."
	@docker network ls | grep microservice-network > /dev/null || docker network create microservice-network
	@docker stop api-gateway-container 2>/dev/null || true
	@docker rm api-gateway-container 2>/dev/null || true
	@if [ ! -f .env ]; then \
		echo "JWT_SECRET=your-super-secret-jwt-key-change-this-in-production" > .env; \
		echo "USER_SERVICE_URL=http://user-service:3001" >> .env; \
		echo "ORDER_SERVICE_URL=http://order-service:3002" >> .env; \
	fi
	@docker run -d \
		--name api-gateway-container \
		--network microservice-network \
		-p 3000:3000 \
		--env-file .env \
		api-gateway
	@echo "✅ API Gateway container started!"
	@echo "📍 Available at: http://localhost:3000"

# Development mode with hot reload (requires air: go install github.com/cosmtrek/air@latest)
dev:
	@echo "🔥 Starting in development mode..."
	@if ! command -v air > /dev/null; then \
		echo "Installing air for hot reload..."; \
		go install github.com/cosmtrek/air@latest; \
	fi
	@if [ ! -f .env ]; then \
		echo "JWT_SECRET=your-super-secret-jwt-key-change-this-in-production" > .env; \
		echo "USER_SERVICE_URL=http://localhost:3001" >> .env; \
		echo "ORDER_SERVICE_URL=http://localhost:3002" >> .env; \
	fi
	@air

# Stop containers
stop:
	@echo "🛑 Stopping API Gateway container..."
	@docker stop api-gateway-container 2>/dev/null || true
	@docker rm api-gateway-container 2>/dev/null || true
	@echo "✅ Container stopped"

# View logs
logs:
	@docker logs -f api-gateway-container

# Clean up
clean: stop
	@echo "🧹 Cleaning up..."
	@docker rmi api-gateway 2>/dev/null || true
	@rm -f bin/api-gateway
	@echo "✅ Cleanup complete"

# Run tests
test:
	@echo "🧪 Running tests..."
	@go test ./...

# Health check
health:
	@echo "🔍 Checking API Gateway health..."
	@curl -s http://localhost:3000/health | jq . || echo "❌ API Gateway not responding or jq not installed"