# ESCROW Microservices Architecture

A production-ready microservices architecture built with NestJS and Go, featuring authentication, contract management, payments, disputes, notifications, and audit logging. The architecture uses **gRPC** for inter-service communication for improved performance and type safety.

## Architecture Overview

```
┌─────────────────────┐
│   API Gateway       │
│   (Go:8080)         │
│   HTTP → gRPC       │
└─────────┬───────────┘
          │ gRPC Communication
    ┌─────┴─────┐────────────┐─────────────┐
    │           │            |             |
    ▼           ▼            ▼             ▼
┌───────┐   ┌───────┐   ┌──────────┐   ┌─────────┐   ┌──────────────┐   ┌───────┐
│ Auth  │   │Contract│   │ Payment  │   │Dispute  │   │Notification  │   │ Audit │
│Service│   │Service │   │ Service  │   │Service  │   │   Service    │   │Service│
│HTTP:  │   │HTTP:   │   │ HTTP:    │   │HTTP:    │   │  HTTP: 8081  │   │HTTP:  │
│3001   │   │3002    │   │ 3003     │   │3004     │   │  gRPC: 50055 │   │8082   │
│gRPC:  │   │gRPC:   │   │ gRPC:    │   │gRPC:    │   │              │   │gRPC:  │
│50051  │   │50052   │   │ 50053    │   │50054    │   │              │   │50056  │
└───┬───┘   └───┬───┘   └────┬─────┘   └────┬────┘   └──────┬───────┘   └───┬───┘
    │           │            │              │               │               │
    ▼           ▼            ▼              ▼               ▼               ▼
┌───────┐   ┌───────┐   ┌───────┐     ┌───────┐       ┌───────┐       ┌───────┐
│Auth DB│   │Contract│   │Payment│     │Dispute│       │ Redis │       │MongoDB│
│(PgSQL)│   │  DB   │   │  DB   │     │  DB   │       │       │       │(Audit)│
└───────┘   └───────┘   └───────┘     └───────┘       └───────┘       └───────┘
```

## gRPC Communication

The microservices communicate with each other using gRPC, which provides:
- **Better Performance**: Binary protocol vs JSON/HTTP
- **Type Safety**: Protocol Buffers schema validation
- **Streaming Support**: Bidirectional streaming capabilities
- **Language Agnostic**: Generated clients for Go and TypeScript

### Protocol Buffers (Proto Files)

Located in the `/proto` directory:
- `auth.proto` - Authentication service definitions
- `contract.proto` - Contract service definitions  
- `payment.proto` - Payment service definitions
- `dispute.proto` - Dispute service definitions
- `notification.proto` - Notification service definitions
- `audit.proto` - Audit service definitions

### Shared Microservice Module

All NestJS services use a shared microservice module located at:
- `/shared/nestjs/modules/microservice/` - Common gRPC utilities, decorators, and interfaces

## Services

### API Gateway (Go - HTTP: 8080)
- **Purpose**: Central entry point for all client requests, converts HTTP to gRPC
- **Features**: JWT authentication, gRPC client management, request routing, audit logging
- **Communication**: Uses gRPC clients to communicate with all microservices
- **Endpoints**:
  - `GET /api/v1/health` - Health check
  - `POST /api/v1/auth/register` - User registration
  - `POST /api/v1/auth/login` - User login
  - `POST /api/v1/auth/validate` - Token validation
  - `GET /api/v1/users/:userId` - Get user profile
  - `POST /api/v1/contracts` - Create contract
  - `GET /api/v1/contracts` - List contracts
  - `GET /api/v1/contracts/:id` - Get contract details
  - `POST /api/v1/wallets` - Create wallet
  - `POST /api/v1/transfers` - Create transfer
  - `POST /api/v1/disputes` - Create dispute
  - `GET /api/v1/notifications` - Get notifications
  - `PUT /api/v1/notifications/:id/read` - Mark notification as read
  - `GET /api/v1/audit/logs` - Get audit logs

### Auth Service (NestJS - HTTP: 3001, gRPC: 50051)
- **Purpose**: User authentication and management
- **Database**: PostgreSQL (port 5432)
- **Features**: User registration, login, JWT tokens, password hashing
- **Communication**: Exposes both HTTP REST API and gRPC interface
- **gRPC Methods**: Register, Login, ValidateToken, GetUser, UpdateUser
- **HTTP Endpoints**:
  - `POST /auth/register` - Register new user
  - `POST /auth/login` - Login user
  - `GET /users/:id` - Get user by ID

### Contract Service (NestJS - HTTP: 3002, gRPC: 50052)
- **Purpose**: Contract creation and management
- **Database**: PostgreSQL (port 5432)
- **Features**: CRUD operations for contracts, user authorization
- **Communication**: Exposes both HTTP REST API and gRPC interface
- **gRPC Methods**: CreateContract, GetContract, GetContracts, UpdateContract, DeleteContract
- **HTTP Endpoints**:
  - `POST /contracts` - Create contract
  - `GET /contracts/:id` - Get contract
  - `GET /contracts/user/:userId` - List user's contracts

### Payment Service (NestJS - HTTP: 3003, gRPC: 50053)
- **Purpose**: Wallet and payment management
- **Database**: PostgreSQL (port 5432)
- **Features**: Digital wallets, deposits, transfers, transaction history
- **Communication**: Exposes both HTTP REST API and gRPC interface
- **gRPC Methods**: CreateWallet, GetWallet, UpdateWallet, CreateTransfer, GetTransactions, GetTransaction
- **HTTP Endpoints**:
  - `GET /wallets/:userId` - Get wallet balance
  - `POST /wallets/deposit` - Add funds
  - `POST /transfers` - Transfer between users

### Dispute Service (NestJS - HTTP: 3004, gRPC: 50054)
- **Purpose**: Dispute management for contracts
- **Database**: PostgreSQL (port 5432)
- **Features**: Create disputes, resolution workflow
- **Communication**: Exposes both HTTP REST API and gRPC interface
- **gRPC Methods**: CreateDispute, GetDispute, GetDisputes, UpdateDispute, ResolveDispute
- **HTTP Endpoints**:
  - `POST /disputes` - Create dispute
  - `GET /disputes/:id` - Get dispute
  - `PUT /disputes/:id/resolve` - Resolve dispute

### Notification Service (Go - HTTP: 8081, gRPC: 50055)
- **Purpose**: Real-time notifications via WebSocket and gRPC
- **Features**: WebSocket connections, real-time messaging, connection management
- **Communication**: Exposes both HTTP/WebSocket API and gRPC interface
- **gRPC Methods**: SendNotification, GetNotifications, MarkAsRead
- **HTTP Endpoints**:
  - `GET /ws` - WebSocket endpoint
  - `POST /notify` - Send notification

### Audit Service (Go - HTTP: 8082, gRPC: 50056)
- **Purpose**: Activity logging and audit trails
- **Database**: MongoDB (port 27017)
- **Features**: Action logging, user activity tracking, audit queries, document-based storage
- **Communication**: Exposes both HTTP REST API and gRPC interface
- **gRPC Methods**: CreateLog, GetLogs, GetLogsByUser
- **HTTP Endpoints**:
  - `POST /logs` - Create audit log
  - `GET /logs/user/:userId` - Get user activity

## Prerequisites

- Docker & Docker Compose
- Node.js 22+ (for local development)
- Go 1.24+ (for local development)
- PostgreSQL (for local development)
- MongoDB (for audit service)
- Protocol Buffers (protoc) - for generating gRPC code
- gRPC tooling for Go and TypeScript

## Quick Start

### Using Docker Compose (Recommended)

1. **Clone and navigate to the project**:
   ```bash
   git clone <repository-url>
   ```

2. **Set up environment variables**:
   ```bash
   cp env.example .env
   # Edit .env with your configuration
   ```

3. **Start all services**:
   ```bash
   docker-compose up -d
   ```

4. **Wait for services to be ready** (check health):
   ```bash
   curl http://localhost:8080/api/v1/health
   ```

### Local Development

1. **Start databases**:
   ```bash
   docker-compose up -d postgres-auth postgres-contract postgres-payment postgres-dispute mongodb-audit redis rabbitmq
   ```

2. **Install dependencies for NestJS services**:
   ```bash
   cd services/auth && npm install
   cd ../contract && npm install
   cd ../payment && npm install
   cd ../dispute && npm install
   ```

3. **Start services individually**:
   ```bash
   # Terminal 1 - Auth Service
   cd services/auth && npm run start:dev

   # Terminal 2 - Contract Service
   cd services/contract && npm run start:dev

   # Terminal 3 - Payment Service
   cd services/payment && npm run start:dev

   # Terminal 4 - Dispute Service
   cd services/dispute && npm run start:dev

   # Terminal 5 - Notification Service
   cd services/notification && go run cmd/main.go

   # Terminal 6 - Audit Service
   cd services/audit && go run cmd/main.go

   # Terminal 7 - API Gateway
   cd gateway && go run cmd/main.go
   ```

## Testing the API

### 1. Register a User
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

### 2. Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

Save the `access_token` from the response for subsequent requests.

### 3. Create a Contract
```bash
curl -X POST http://localhost:8080/api/v1/contracts \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -d '{
    "sellerId": "USER_ID_1",
    "buyerId": "USER_ID_2",
    "title": "Web Development Contract",
    "amount": 1000.00
  }'
```

### 4. Check Wallet Balance
```bash
curl -X GET http://localhost:8080/api/v1/wallets/USER_ID \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

### 5. Make a Deposit
```bash
curl -X POST http://localhost:8080/api/v1/wallets/deposit \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -d '{
    "userId": "USER_ID",
    "amount": 500.00,
    "description": "Initial deposit"
  }'
```

### 6. Connect to WebSocket
```javascript
const ws = new WebSocket('ws://localhost:8081/ws?userId=USER_ID&clientId=web-client');
ws.onmessage = (event) => {
  console.log('Notification:', JSON.parse(event.data));
};
```

## Environment Variables

Key environment variables for configuration:

```env
# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-here

# Database URLs
AUTH_DATABASE_URL=postgresql://postgres:postgres@localhost:5432/auth_db
CONTRACT_DATABASE_URL=postgresql://postgres:postgres@localhost:5433/contract_db
PAYMENT_DATABASE_URL=postgresql://postgres:postgres@localhost:5434/payment_db
DISPUTE_DATABASE_URL=postgresql://postgres:postgres@localhost:5435/dispute_db

# MongoDB Configuration for Audit Service
MONGODB_URI=mongodb://audit_user:audit_password@localhost:27017/audit_db?authSource=audit_db

# Service URLs (for inter-service communication)
AUTH_SERVICE_URL=http://localhost:3001
CONTRACT_SERVICE_URL=http://localhost:3002
PAYMENT_SERVICE_URL=http://localhost:3003
DISPUTE_SERVICE_URL=http://localhost:3004
NOTIFICATION_SERVICE_URL=http://localhost:8081
AUDIT_SERVICE_URL=http://localhost:8082

# External Services
REDIS_URL=redis://localhost:6379
RABBITMQ_URL=amqp://admin:admin@localhost:5672
```

## Data Models

### User
```typescript
{
  id: string (uuid)
  email: string (unique)
  password: string (hashed)
  createdAt: Date
  updatedAt: Date
}
```

### Contract
```typescript
{
  id: string (uuid)
  sellerId: string (uuid)
  buyerId: string (uuid)
  title: string
  amount: number (decimal)
  status: 'DRAFT' | 'ACTIVE' | 'COMPLETED'
  createdAt: Date
  updatedAt: Date
}
```

### Wallet & Transaction
```typescript
{
  // Wallet
  id: string (uuid)
  userId: string (uuid)
  balance: number (decimal)
  createdAt: Date
  updatedAt: Date

  // Transaction
  id: string (uuid)
  fromWalletId?: string (uuid)
  toWalletId?: string (uuid)
  amount: number (decimal)
  type: 'DEPOSIT' | 'WITHDRAWAL' | 'TRANSFER' | 'PAYMENT' | 'REFUND'
  description?: string
  createdAt: Date
}
```

### Dispute
```typescript
{
  id: string (uuid)
  contractId: string (uuid)
  raisedBy: string (uuid)
  description: string
  status: 'OPEN' | 'RESOLVED'
  resolvedAt?: Date
  resolution?: string
  resolvedBy?: string (uuid)
  createdAt: Date
  updatedAt: Date
}
```

## Monitoring & Health Checks

All services expose health check endpoints:
- Gateway: `GET /api/v1/health`
- Auth: `GET /health`
- Contract: `GET /health`
- Payment: `GET /health`
- Dispute: `GET /health`
- Notification: `GET /health`
- Audit: `GET /health`

## Security Features

- JWT-based authentication
- Password hashing with bcrypt
- Request validation and sanitization
- CORS protection
- Rate limiting ready (implement as needed)
- SQL injection prevention (ORM/query builder)
- XSS protection through validation

## Development

### Protocol Buffers (gRPC) Development

1. **Install gRPC dependencies**:
   ```bash
   # Install dependencies for all NestJS services
   make proto-deps
   ```

2. **Generate gRPC code from proto files**:
   ```bash
   # Generate Go and TypeScript gRPC code
   make proto-gen
   ```

3. **Modify proto files**: Edit files in `/proto` directory when:
   - Adding new service methods
   - Changing message structures
   - Adding new services

4. **After proto changes**: Always regenerate code and restart services
   ```bash
   make proto-gen
   docker-compose restart
   ```

5. **Clean generated files** (if needed):
   ```bash
   make proto-clean
   ```

### Adding New Features

1. **New endpoints**: Add to respective service controllers and gRPC handlers
2. **New entities**: Create entities and run migrations
3. **Inter-service communication**: Use gRPC clients through the API gateway
4. **Real-time features**: Use WebSocket service for notifications

### Database Migrations

For NestJS services (using TypeORM):
```bash
npm run typeorm migration:generate -- -n MigrationName
npm run typeorm migration:run
```

For Go services (using GORM):
Auto-migration is enabled in development. For production, implement proper migration scripts.

## Production Considerations

- [ ] Replace `synchronize: true` with proper migrations
- [ ] Implement proper secret management
- [ ] Add rate limiting and API throttling
- [ ] Implement proper logging aggregation
- [ ] Add monitoring and alerting
- [ ] Implement circuit breakers for service communication
- [ ] Add distributed tracing
- [ ] Implement proper backup strategies
- [ ] Add SSL/TLS termination
- [ ] Implement proper error handling and retry logic

## Troubleshooting

### Common Issues

1. **Port conflicts**: Ensure ports 3001-3004, 8080-8082, 5432-5435, 27017, 6379, 5672, 15672 are available
2. **Database connection**: Check PostgreSQL and MongoDB containers are running and accessible
3. **JWT errors**: Ensure JWT_SECRET is consistent across all services
4. **CORS issues**: Check CORS configuration in each service
5. **WebSocket connection**: Ensure no proxy/firewall blocking WebSocket connections

### Debugging

Enable debug logging:
```bash
export LOG_LEVEL=debug
export NODE_ENV=development
```

Check service logs:
```bash
docker-compose logs -f service-name
```

## License

MIT License - see LICENSE file for details.

