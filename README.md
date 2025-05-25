# Wallets

![Coverage](https://img.shields.io/badge/coverage-31%25-red)
[![Go Report Card](https://goreportcard.com/badge/github.com/ezhdanovskiy/wallets)](https://goreportcard.com/report/github.com/ezhdanovskiy/wallets)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Project Overview

Wallets is a microservice for managing electronic wallets written in Go. The service provides a REST API for creating wallets, depositing funds, transferring between wallets, and viewing transaction history.

### Key Features:
- Create named wallets
- Deposit funds to wallets
- Transfer funds between wallets
- View transaction history with date filtering
- Export transactions to CSV format

## Quick Start

### Requirements
- Go 1.20+
- PostgreSQL 12+
- Docker and Docker Compose (optional)

### Running with Docker Compose
```bash
# Start all services (PostgreSQL + application)
docker-compose up

# Start in background
docker-compose up -d
```

### Local Development
```bash
# 1. Start PostgreSQL
make postgres/up

# 2. Apply database migrations
make migrate/up

# 3. Run the application
make run
```

### Testing
```bash
# Run unit tests
make test

# Run integration tests
make test/int

# Run tests with coverage
make test/coverage

# Run integration tests with coverage
make test/coverage/int
```

### Building
```bash
# Build binary
make build

# Build Docker image
make build/docker
```

## Project Structure

```
.
├── api/
│   └── v1/
│       └── swagger.yaml         # OpenAPI specification
├── cmd/
│   └── main.go                  # Application entry point
├── docs/
│   └── diagrams/                # Architecture diagrams
├── internal/
│   ├── application/             # Application initialization and startup
│   │   ├── application.go
│   │   └── logger.go
│   ├── config/                  # Configuration from env variables
│   │   └── config.go
│   ├── consts/                  # Application constants
│   │   └── consts.go
│   ├── csv/                     # CSV report generation
│   │   └── operations.go
│   ├── dto/                     # Data Transfer Objects
│   │   ├── amount.go
│   │   ├── deposit.go
│   │   ├── operation.go
│   │   ├── transfer.go
│   │   └── wallet.go
│   ├── http/                    # HTTP layer
│   │   ├── dependencies.go
│   │   ├── errors.go
│   │   ├── handlers.go
│   │   └── server.go
│   ├── httperr/                 # HTTP errors
│   │   └── errors.go
│   ├── repository/              # Database layer
│   │   ├── entities.go
│   │   └── repository.go
│   ├── service/                 # Business logic
│   │   ├── dependencies.go
│   │   ├── errors.go
│   │   ├── service.go
│   │   └── service_test.go
│   └── tests/                   # Integration tests
│       └── integration_test.go
├── migrations/                  # SQL migrations
├── docker-compose.yml
├── Dockerfile
├── go.mod
├── go.sum
├── Makefile
├── README.md                    # This file (English)
└── README_ru.md                 # Russian documentation
```

## Architecture

The application follows a three-layer architecture with clear boundaries between layers:

![Package Dependencies Diagram](docs/diagrams/package-dependencies.png)

### Application Layers:

1. **HTTP Layer** (`internal/http/`) - HTTP request handling, input validation, routing
2. **Service Layer** (`internal/service/`) - business logic, transaction management, business rule validation
3. **Repository Layer** (`internal/repository/`) - database operations, SQL queries

### Key Architectural Decisions:

- **Transaction Isolation Level**: Uses `sql.LevelSerializable` for all wallet operations, ensuring data consistency in concurrent operations
- **Money Storage**: Amounts are stored as `numeric(18,2)` in the database and `float64` in Go code
- **Operation Logging**: All monetary operations are recorded in the `operations` table for audit purposes
- **Validation**: Minimum operation amount is 0.01, validated at the HTTP layer

### Database

PostgreSQL with two main tables:
- **wallets** - wallet information (id, name, balance, created_at, updated_at)
- **operations** - transaction history (id, wallet_id, type, amount, created_at)

## Configuration

The application is configured via environment variables:

| Variable | Description | Default Value |
|----------|-------------|---------------|
| `DB_HOST` | PostgreSQL host | `localhost` |
| `DB_PORT` | PostgreSQL port | `5432` |
| `DB_USER` | Database user | `postgres` |
| `DB_PASSWORD` | Database password | `postgres` |
| `DB_NAME` | Database name | `wallets` |
| `APP_PORT` | HTTP server port | `8080` |
| `LOG_LEVEL` | Logging level | `info` |

## API Endpoints

### POST /v1/wallets
Create a new wallet
```json
{
  "name": "My Wallet"
}
```

### POST /v1/wallets/deposit
Deposit funds to a wallet
```json
{
  "wallet_id": "123e4567-e89b-12d3-a456-426614174000",
  "amount": 100.50
}
```

### POST /v1/wallets/transfer
Transfer funds between wallets
```json
{
  "from_wallet_id": "123e4567-e89b-12d3-a456-426614174000",
  "to_wallet_id": "987fcdeb-51a2-43d1-9012-345678901234",
  "amount": 50.00
}
```

### GET /v1/wallets/operations
Get transaction history with optional filters:
- `wallet_id` - Wallet ID
- `from_date` - Start date (RFC3339)
- `to_date` - End date (RFC3339)
- `offset` - Pagination offset
- `limit` - Number of records

Detailed API specification is available in [api/v1/swagger.yaml](api/v1/swagger.yaml).

## Available Commands

### Makefile Commands

| Command | Description |
|---------|-------------|
| `make build` | Build application binary |
| `make test` | Run unit tests |
| `make test/int` | Run integration tests |
| `make test/coverage` | Run tests with coverage report |
| `make test/coverage/int` | Run integration tests with coverage |
| `make run` | Run the application |
| `make postgres/up` | Start PostgreSQL in Docker container |
| `make postgres/down` | Stop PostgreSQL container |
| `make migrate/up` | Apply all migrations |
| `make migrate/down` | Rollback last migration |
| `make build/docker` | Build Docker image |
| `make diagrams` | Generate diagrams from DOT files |

### Development Commands

```bash
# View package documentation
go doc internal/service

# Generate mocks for testing
go generate ./...

# Check test coverage
go test -cover ./...

# Run specific test
go test -run TestServiceTransfer ./internal/service
```

## Development Features

### Testing
- **Unit Tests**: Use repository mocks to isolate business logic. See example in `internal/service/service_test.go`
- **Integration Tests**: Test the full stack with a real database. See `internal/tests/integration_test.go`
- **Code Coverage**: The project maintains comprehensive test coverage. In PRs, coverage is compared with the master branch to track improvements

### Error Handling
- Custom errors are defined in `internal/httperr/` for proper HTTP semantics
- Business errors (insufficient funds, wallet not found) return appropriate HTTP status codes

### Logging
- Structured logging via Zap
- All monetary operations are logged for audit purposes

### Metrics
- Prometheus integration for metrics collection
- Available at `/metrics` endpoint

## Documentation

- [README.md](README.md) - English documentation (this file)
- [README_ru.md](README_ru.md) - Russian documentation
- [api/v1/swagger.yaml](api/v1/swagger.yaml) - OpenAPI specification

## License

This project is distributed under the MIT License. See [LICENSE](LICENSE) file for details.