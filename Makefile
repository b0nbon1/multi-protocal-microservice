# Marketplace Microservices Makefile

.PHONY: help dev start stop restart clean test logs build status

# Default target
help: ## Show this help message
	@echo "Marketplace Microservices - Available Commands"
	@echo "=============================================="
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

# Development commands
dev: ## Start development environment
	@echo "🚀 Starting development environment..."
	@./scripts/start-dev.sh

start: ## Start all services with Docker Compose
	@echo "🚀 Starting all services..."
	@docker compose up -d --build
	@echo "✅ All services started!"

stop: ## Stop all services
	@echo "⏹️ Stopping all services..."
	@docker compose down
	@echo "✅ All services stopped!"

restart: ## Restart all services
	@echo "🔄 Restarting all services..."
	@docker compose restart
	@echo "✅ All services restarted!"

# Infrastructure commands
infra: ## Start only infrastructure services (databases, redis, rabbitmq)
	@echo "🏗️ Starting infrastructure services..."
	@docker compose up -d postgres-auth postgres-contract postgres-payment postgres-dispute postgres-audit redis rabbitmq
	@echo "✅ Infrastructure services started!"

# Cleanup commands
clean: ## Remove all containers, networks, and volumes
	@echo "🧹 Cleaning up..."
	@docker compose down -v --remove-orphans
	@docker system prune -f
	@echo "✅ Cleanup completed!"

clean-volumes: ## Remove all volumes (WARNING: This will delete all data)
	@echo "⚠️  This will delete ALL data. Are you sure? [y/N]" && read ans && [ $${ans:-N} = y ]
	@docker compose down -v
	@docker volume prune -f
	@echo "✅ All volumes removed!"

# Testing commands
test: ## Run API tests
	@echo "🧪 Running API tests..."
	@./scripts/test-api.sh

test-manual: ## Show manual testing examples
	@echo "📚 Manual Testing Examples"
	@echo "=========================="
	@echo ""
	@echo "1. Health Check:"
	@echo "   curl http://localhost:8080/api/v1/health"
	@echo ""
	@echo "2. Register User:"
	@echo "   curl -X POST http://localhost:8080/api/v1/auth/register \\"
	@echo "     -H 'Content-Type: application/json' \\"
	@echo "     -d '{\"email\":\"test@example.com\",\"password\":\"password123\"}'"
	@echo ""
	@echo "3. Login:"
	@echo "   curl -X POST http://localhost:8080/api/v1/auth/login \\"
	@echo "     -H 'Content-Type: application/json' \\"
	@echo "     -d '{\"email\":\"test@example.com\",\"password\":\"password123\"}'"
	@echo ""
	@echo "4. WebSocket Connection:"
	@echo "   ws://localhost:8081/ws?userId=USER_ID&clientId=test"

# Logging commands
logs: ## Show logs for all services
	@docker compose logs -f

logs-gateway: ## Show API Gateway logs
	@docker compose logs -f gateway

logs-auth: ## Show Auth Service logs
	@docker compose logs -f auth-service

logs-contract: ## Show Contract Service logs
	@docker compose logs -f contract-service

logs-payment: ## Show Payment Service logs
	@docker compose logs -f payment-service

logs-dispute: ## Show Dispute Service logs
	@docker compose logs -f dispute-service

logs-notification: ## Show Notification Service logs
	@docker compose logs -f notification-service

logs-audit: ## Show Audit Service logs
	@docker compose logs -f audit-service

# Build commands
build: ## Build all Docker images
	@echo "🔨 Building all Docker images..."
	@docker compose build
	@echo "✅ All images built!"

build-no-cache: ## Build all Docker images without cache
	@echo "🔨 Building all Docker images (no cache)..."
	@docker compose build --no-cache
	@echo "✅ All images built!"

# Status commands
status: ## Show status of all services
	@echo "📊 Service Status"
	@echo "================"
	@docker compose ps

health: ## Check health of all services
	@echo "🔍 Health Check"
	@echo "==============="
	@echo "API Gateway:      $$(curl -s -o /dev/null -w '%{http_code}' http://localhost:8080/api/v1/health || echo 'DOWN')"
	@echo "Auth Service:     $$(curl -s -o /dev/null -w '%{http_code}' http://localhost:3001/health || echo 'DOWN')"
	@echo "Contract Service: $$(curl -s -o /dev/null -w '%{http_code}' http://localhost:3002/health || echo 'DOWN')"
	@echo "Payment Service:  $$(curl -s -o /dev/null -w '%{http_code}' http://localhost:3003/health || echo 'DOWN')"
	@echo "Dispute Service:  $$(curl -s -o /dev/null -w '%{http_code}' http://localhost:3004/health || echo 'DOWN')"
	@echo "Notification:     $$(curl -s -o /dev/null -w '%{http_code}' http://localhost:8081/health || echo 'DOWN')"
	@echo "Audit Service:    $$(curl -s -o /dev/null -w '%{http_code}' http://localhost:8082/health || echo 'DOWN')"

# gRPC Development commands
proto-gen: ## Generate gRPC code from proto files
	@echo "⚙️ Generating gRPC code from proto files..."
	@./scripts/generate-proto.sh
	@echo "✅ gRPC code generated!"

proto-deps: ## Install gRPC dependencies for all services
	@echo "📦 Installing gRPC dependencies..."
	@cd services/auth && npm install
	@cd services/contract && npm install
	@cd services/payment && npm install
	@cd services/dispute && npm install
	@echo "✅ gRPC dependencies installed!"

proto-clean: ## Clean generated gRPC files
	@echo "🧹 Cleaning generated gRPC files..."
	@rm -rf gateway/internal/grpc/auth/*.pb.go || true
	@rm -rf gateway/internal/grpc/contract/*.pb.go || true
	@rm -rf gateway/internal/grpc/payment/*.pb.go || true
	@rm -rf gateway/internal/grpc/dispute/*.pb.go || true
	@rm -rf gateway/internal/grpc/notification/*.pb.go || true
	@rm -rf gateway/internal/grpc/audit/*.pb.go || true
	@rm -rf services/*/src/proto/* || true
	@echo "✅ Generated gRPC files cleaned!"

# Development helpers
install-deps: ## Install dependencies for all NestJS services
	@echo "📦 Installing dependencies..."
	@cd services/auth && npm install
	@cd services/contract && npm install
	@cd services/payment && npm install
	@cd services/dispute && npm install
	@echo "✅ Dependencies installed!"

format: ## Format code for all services
	@echo "🎨 Formatting code..."
	@cd services/auth && npm run format || true
	@cd services/contract && npm run format || true
	@cd services/payment && npm run format || true
	@cd services/dispute && npm run format || true
	@cd gateway && go fmt ./... || true
	@cd services/notification && go fmt ./... || true
	@cd services/audit && go fmt ./... || true
	@echo "✅ Code formatted!"

# Database commands
db-reset: ## Reset all databases (WARNING: This will delete all data)
	@echo "⚠️  This will delete ALL database data. Are you sure? [y/N]" && read ans && [ $${ans:-N} = y ]
	@docker compose down
	@docker volume rm marketplace_postgres_auth_data marketplace_postgres_contract_data marketplace_postgres_payment_data marketplace_postgres_dispute_data marketplace_postgres_audit_data || true
	@echo "✅ Databases reset!"

# Quick shortcuts
up: start ## Alias for start
down: stop ## Alias for stop

