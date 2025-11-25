# Orders Microservice

A minimal Orders microservice built in Go that provides order management functionality with event-driven architecture, idempotency support, and observability features.

## Short Summary

This microservice exposes REST APIs for creating and retrieving orders. It includes:
- PostgreSQL persistence with Flyway migrations
- In-process event queue with background worker for OrderCreated events
- Idempotency support for POST /orders
- Structured logging with zap
- OpenTelemetry tracing
- Graceful shutdown and context propagation
- Docker containerization with multi-stage builds

## How to Run

### Using Docker Compose (Recommended)

1. Ensure Docker and Docker Compose are installed
2. Run the entire stack:
   ```bash
   docker-compose up
   ```

This will start:
- PostgreSQL database
- Flyway migrations (runs automatically)
- Orders microservice

The Compose file pulls service-specific configuration from the env files in `env/`. Update these before running if you need different values:

- `env/postgres.env`
- `env/flyway.env`
- `env/orders-service.env`

The service will be available at `http://localhost:8080`
The swagger url is: `http://localhost:8080/swagger/index.html`

### Local Development

1. Start PostgreSQL (or use Docker):
   ```bash
   docker run -d --name postgres \
     -e POSTGRES_USER=postgres \
     -e POSTGRES_PASSWORD=postgres \
     -e POSTGRES_DB=ordersdb \
     -p 5432:5432 \
     postgres:15-alpine
   ```

2. Run Flyway migrations:
   ```bash
   docker run --rm \
     -v $(pwd)/migrations:/flyway/sql \
     --network host \
     flyway/flyway:9 \
     -url=jdbc:postgresql://localhost:5432/ordersdb \
     -user=postgres \
     -password=postgres \
     migrate
   ```

3. Set environment variables:
   ```bash
   export SERVER_PORT=8080
   export DB_HOST=localhost
   export DB_PORT=5432
   export DB_USER=postgres
   export DB_PASSWORD=postgres
   export DB_NAME=ordersdb
   export DB_SSLMODE=disable
   export LOG_LEVEL=info
   export OTEL_ENABLED=true
   export EVENT_QUEUE_SIZE=100
   ```

4. Run the service:
   ```bash
   go run cmd/server/main.go
   ```

## How to Build

### Local Build

```bash
go build -o orders-service ./cmd/server
```

### Docker Build

```bash
docker build -t orders-service .
```

### Run Tests

```bash
go test ./...
```

### Generate Swagger Documentation

Swagger documentation is pre-generated in `docs/docs.go`. To regenerate:

```bash
go install github.com/swaggo/swag/cmd/swag@latest
swag init -g cmd/server/main.go
```

## Migration

Migrations are managed using Flyway. The migration files are located in the `migrations/` directory.

### Migration Files

- `V1__create_orders_table.sql` - Creates the orders and idempotency_keys tables

### Running Migrations

Migrations run automatically when using docker-compose. For manual execution:

```bash
docker run --rm \
  -v $(pwd)/migrations:/flyway/sql \
  --network host \
  flyway/flyway:9 \
  -url=jdbc:postgresql://localhost:5432/ordersdb \
  -user=postgres \
  -password=postgres \
  migrate
```

## Docker Compose

The `docker-compose.yml` file defines three services:

1. **postgres**: PostgreSQL 15 database
2. **flyway**: Runs database migrations before the application starts
3. **orders-service**: The main application service

The services are configured with proper health checks and dependencies to ensure correct startup order.

## API Endpoints

### POST /orders

Create a new order.

**Request Body:**
```json
{
  "customer_id": "customer-123",
  "product_id": "product-456",
  "quantity": 2,
  "total_price": 100.50,
  "idempotency_key": "unique-key-123"
}
```

**Response:** 201 Created
```json
{
  "id": "order-uuid",
  "customer_id": "customer-123",
  "product_id": "product-456",
  "quantity": 2,
  "total_price": 100.50,
  "status": "created",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

### GET /orders/{id}

Retrieve an order by ID.

**Response:** 200 OK
```json
{
  "id": "order-uuid",
  "customer_id": "customer-123",
  "product_id": "product-456",
  "quantity": 2,
  "total_price": 100.50,
  "status": "created",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

### GET /healthz

Health check endpoint.

**Response:** 200 OK
```json
{
  "status": "healthy"
}
```

### GET /swagger/index.html

Swagger UI documentation.

Access the interactive API documentation at `http://localhost:8080/swagger/index.html`

## Short Design Notes

### Architecture

- **Layered Architecture**: Handler → Service → Repository pattern
- **Event-Driven**: In-process Go channels for event emission with background worker
- **Idempotency**: Implemented via idempotency_keys table to prevent duplicate order creation
- **Context Propagation**: All operations use context.Context for cancellation and timeouts
- **Graceful Shutdown**: Handles SIGINT/SIGTERM with 30s timeout for in-flight requests

### Key Components

1. **Models**: Domain models for Order and CreateOrderRequest
2. **Repository**: Database access layer with PostgreSQL
3. **Service**: Business logic layer with event emission
4. **Handler**: HTTP request handlers using Gin framework
5. **Events**: OrderCreated event with background worker processing
6. **Config**: 12-factor configuration via environment variables
7. **Logging**: Structured logging with zap
8. **Tracing**: OpenTelemetry integration for distributed tracing

### Error Handling

- Basic error taxonomy with repository-level errors (e.g., ErrOrderNotFound)
- Proper HTTP status codes (400, 404, 500)
- Error logging with structured fields

### Observability

- **Logging**: Structured JSON logs with zap
- **Tracing**: OpenTelemetry with stdout exporter (can be extended to Jaeger/OTLP)
- **Health Checks**: Simple health endpoint for monitoring

### Idempotency

POST /orders requests require an `idempotency_key`. If the same key is used twice, the original order is returned instead of creating a duplicate.

### Event Processing

OrderCreated events are emitted to an in-process channel and processed by a background worker. 


## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| SERVER_PORT | 8080 | HTTP server port |
| DB_HOST | localhost | PostgreSQL host |
| DB_PORT | 5432 | PostgreSQL port |
| DB_USER | postgres | Database user |
| DB_PASSWORD | postgres | Database password |
| DB_NAME | ordersdb | Database name |
| DB_SSLMODE | disable | SSL mode |
| LOG_LEVEL | info | Log level (debug, info, warn, error) |
| OTEL_ENABLED | true | Enable OpenTelemetry tracing |
| EVENT_QUEUE_SIZE | 100 | Size of event channel buffer |
| GIN_MODE | debug | Detailed logs of gin module release/debug |

## What is missing

More unittests are needed. For proper unittests mocking is needed (like OrderRepository) and some refactor on model package (interface for proper mocking).

Some validation/protection against SQL injection is still missing.

The oder worker can be extended to:
- Send notifications
- Update inventory
- Trigger downstream services




