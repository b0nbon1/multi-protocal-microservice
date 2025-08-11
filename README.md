# ESCROW Microservices Architecture

A production-ready microservices architecture built with NestJS and Go, featuring authentication, contract management, payments, disputes, notifications, and audit logging.

## Architecture Overview

```
┌─────────────────┐
│   API Gateway   │
│    (Go:8080)    │
└─────────┬───────┘
          │
    ┌─────┴─────┐
    │           │
    ▼           ▼
┌───────┐   ┌───────┐   ┌──────────┐   ┌─────────┐   ┌──────────────┐   ┌───────┐
│ Auth  │   │Contract│   │ Payment  │   │Dispute  │   │Notification  │   │ Audit │
│Service│   │Service │   │ Service  │   │Service  │   │   Service    │   │Service│
│(3001) │   │ (3002) │   │  (3003)  │   │ (3004)  │   │    (8081)    │   │(8082) │
└───┬───┘   └───┬───┘   └────┬─────┘   └────┬────┘   └──────┬───────┘   └───┬───┘
    │           │            │              │               │               │
    ▼           ▼            ▼              ▼               ▼               ▼
┌───────┐   ┌───────┐   ┌───────┐     ┌───────┐       ┌───────┐       ┌───────┐
│Auth DB│   │Contract│   │Payment│     │Dispute│       │ Redis │       │Audit  │
│       │   │  DB   │   │  DB   │     │  DB   │       │       │       │  DB   │
└───────┘   └───────┘   └───────┘     └───────┘       └───────┘       └───────┘
```

## Services

### API Gateway (Go - Port 8080)
- **Purpose**: Central entry point for all client requests
- **Features**: JWT authentication, request routing, audit logging
- **Endpoints**:
  - `GET /api/v1/health` - Health check
  - `POST /api/v1/auth/*` - Authentication routes (no auth required)
  - `GET|POST|PUT|DELETE /api/v1/*` - Protected routes (auth required)

### Auth Service (NestJS - Port 3001)
- **Purpose**: User authentication and management
- **Database**: PostgreSQL (port 5432)
- **Features**: User registration, login, JWT tokens, password hashing
- **Endpoints**:
  - `POST /auth/register` - Register new user
  - `POST /auth/login` - Login user
  - `GET /users/:id` - Get user by ID

### Contract Service (NestJS - Port 3002)
- **Purpose**: Contract creation and management
- **Database**: PostgreSQL (port 5433)
- **Features**: CRUD operations for contracts, user authorization
- **Endpoints**:
  - `POST /contracts` - Create contract
  - `GET /contracts/:id` - Get contract
  - `GET /contracts/user/:userId` - List user's contracts

### Payment Service (NestJS - Port 3003)
- **Purpose**: Wallet and payment management
- **Database**: PostgreSQL (port 5434)
- **Features**: Digital wallets, deposits, transfers, transaction history
- **Endpoints**:
  - `GET /wallets/:userId` - Get wallet balance
  - `POST /wallets/deposit` - Add funds
  - `POST /transfers` - Transfer between users

### Dispute Service (NestJS - Port 3004)
- **Purpose**: Dispute management for contracts
- **Database**: PostgreSQL (port 5435)
- **Features**: Create disputes, resolution workflow
- **Endpoints**:
  - `POST /disputes` - Create dispute
  - `GET /disputes/:id` - Get dispute
  - `PUT /disputes/:id/resolve` - Resolve dispute

### Notification Service (Go - Port 8081)
- **Purpose**: Real-time notifications via WebSocket
- **Features**: WebSocket connections, real-time messaging, connection management
- **Endpoints**:
  - `GET /ws` - WebSocket endpoint
  - `POST /notify` - Send notification

### Audit Service (Go - Port 8082)
- **Purpose**: Activity logging and audit trails
- **Database**: PostgreSQL (port 5436)
- **Features**: Action logging, user activity tracking, audit queries
- **Endpoints**:
  - `POST /logs` - Create audit log
  - `GET /logs/user/:userId` - Get user activity

## Prerequisites

- Docker & Docker Compose
- Node.js 22+ (for local development)
- Go 1.24+ (for local development)
- PostgreSQL (for local development)

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
   docker-compose up -d postgres-auth postgres-contract postgres-payment postgres-dispute postgres-audit redis rabbitmq
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
AUDIT_DATABASE_URL=postgres://postgres:postgres@localhost:5436/audit_db

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

### Adding New Features

1. **New endpoints**: Add to respective service controllers
2. **New entities**: Create entities and run migrations
3. **Inter-service communication**: Use HTTP clients or message queues
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

1. **Port conflicts**: Ensure ports 3001-3004, 8080-8082, 5432-5436, 6379, 5672, 15672 are available
2. **Database connection**: Check PostgreSQL containers are running and accessible
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

