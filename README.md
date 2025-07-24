# Subscription Service

A REST API service for managing user subscriptions built with Go, Gin framework, and PostgreSQL.

## Stack

- **Language**: Go 1.24.2
- **Framework**: Gin Web Framework
- **Database**: PostgreSQL 15
- **Containerization**: Docker & Docker Compose
- **Documentation**: Swagger/OpenAPI
- **Logging**: Logrus
- **Database Driver**: pgx/v5

## API Endpoints

### Subscriptions

- `POST /api/v1/subscriptions` - Create a new subscription
- `GET /api/v1/subscriptions` - List subscriptions with filters and pagination
- `GET /api/v1/subscriptions/{id}` - Get subscription by ID
- `PUT /api/v1/subscriptions/{id}` - Update subscription
- `DELETE /api/v1/subscriptions/{id}` - Delete subscription

### Cost Analysis

- `GET /api/v1/cost/summary` - Calculate total cost summary

### Documentation

- `GET /swagger/index.html` - Swagger UI documentation

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Make (optional, for using Makefile commands)

### Using Make Commands

1. **Start the application**:
   ```bash
   make start
   ```

2. **Check service health**:
   ```bash
   make health
   ```

3. **Stop services**:
   ```bash
   make stop
   ```

4. **See all available commands**:
   ```bash
   make help
   ```

### Manual Docker Compose

1. **Clone the repository and navigate to the project directory**

2. **Start services**:
   ```bash
   docker-compose up -d
   ```

3. **Check if services are running**:
   ```bash
   docker-compose ps
   ```

4. **Access the API**:
   - API Base URL: `http://localhost:8080/api/v1`
   - Swagger Documentation: `http://localhost:8080/swagger/index.html`

## Configuration

The application uses `config.yaml` for configuration. Key settings include:

- **Server**: Port, timeouts
- **Database**: Connection parameters, pool settings
- **Logging**: Level, file location, rotation settings

Environment variables can override config file settings:
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE`
- `SERVER_PORT`, `LOG_LEVEL`

## Database Schema

The application uses PostgreSQL with the following main table:

```sql
CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    service_name VARCHAR(255) NOT NULL,
    price INTEGER NOT NULL CHECK (price > 0),
    user_id UUID NOT NULL,
    start_date VARCHAR(7) NOT NULL,  -- Format: MM-YYYY
    end_date VARCHAR(7),             -- Format: MM-YYYY
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

## API Usage Examples

### Create Subscription

```bash
curl -X POST http://localhost:8080/api/v1/subscriptions \
  -H "Content-Type: application/json" \
  -d '{
    "service_name": "Netflix",
    "price": 1299,
    "user_id": "123e4567-e89b-12d3-a456-426614174000",
    "start_date": "01-2024",
    "end_date": "12-2024"
  }'
```

### List Subscriptions

```bash
curl "http://localhost:8080/api/v1/subscriptions?page=1&page_size=10&user_id=123e4567-e89b-12d3-a456-426614174000"
```

### Get Cost Summary

```bash
curl "http://localhost:8080/api/v1/cost/summary?start_month=01-2024&end_month=12-2024&user_id=123e4567-e89b-12d3-a456-426614174000"
```

### Code Quality

- **Run linter**: `make lint`
- **Run tests**: `make test`
- **Run tests with coverage**: `make test-coverage`

## Logging

Logs are written to both stdout and `./logs/app.log` 

## Health Checks

The application includes health checks for:

- Database connectivity (PostgreSQL)
- API endpoint responsiveness
- Container health status

Check health with: `make health`


