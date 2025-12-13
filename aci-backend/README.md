# ACI Backend

Armor Cyber Intelligence (ACI) Backend - A cybersecurity news aggregation platform with AI-powered summarization and real-time alerts.

## Architecture

This project follows **Clean Architecture** principles with clear separation of concerns:

```
aci-backend/
├── cmd/server/              # Application entry point
├── internal/
│   ├── config/              # Configuration management
│   ├── domain/              # Business entities and logic
│   │   ├── entities/        # Core business entities
│   │   ├── valueobjects/    # Immutable value objects
│   │   └── errors/          # Domain-specific errors
│   ├── repository/          # Data persistence layer
│   │   ├── postgres/        # PostgreSQL implementations
│   │   └── redis/           # Redis implementations
│   ├── service/             # Business logic layer
│   │   ├── auth/            # Authentication service
│   │   ├── article/         # Article management
│   │   ├── alert/           # Alert management
│   │   └── ai/              # AI integration
│   ├── api/                 # HTTP layer
│   │   ├── handlers/        # HTTP handlers
│   │   ├── middleware/      # HTTP middleware
│   │   └── dto/             # Data transfer objects
│   ├── websocket/           # WebSocket layer
│   │   ├── hub/             # WebSocket hub
│   │   └── client/          # Client management
│   ├── ai/                  # AI integration
│   │   ├── client/          # Anthropic client
│   │   └── prompts/         # Prompt templates
│   └── pkg/                 # Shared utilities
│       ├── logger/          # Logging utilities
│       ├── validator/       # Validation utilities
│       └── crypto/          # Cryptographic utilities
├── migrations/              # Database migrations
├── tests/                   # Test suites
│   ├── unit/                # Unit tests
│   ├── integration/         # Integration tests
│   └── e2e/                 # End-to-end tests
├── deployments/             # Deployment configurations
│   ├── docker/              # Docker configurations
│   └── k8s/                 # Kubernetes manifests
└── docs/                    # Documentation
```

## Tech Stack

- **Language**: Go 1.22+
- **Web Framework**: Chi (v5)
- **WebSocket**: Gorilla WebSocket
- **Database**: PostgreSQL with pgx (v5)
- **Cache**: Redis (optional)
- **Authentication**: JWT with RS256
- **AI**: Anthropic Claude SDK
- **Logging**: Zerolog
- **Validation**: go-playground/validator
- **Migrations**: golang-migrate
- **Testing**: testify

## Prerequisites

- Go 1.22 or higher
- PostgreSQL 14+
- Redis (optional, for sessions/cache)
- golangci-lint (for linting)
- golang-migrate (for migrations)

## Setup

1. **Clone the repository**
   ```bash
   cd aci-backend
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Install development tools**
   ```bash
   make dev-deps
   ```

4. **Configure environment**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

5. **Generate JWT keys**
   ```bash
   mkdir -p keys
   openssl genrsa -out keys/jwt-private.pem 2048
   openssl rsa -in keys/jwt-private.pem -pubout -out keys/jwt-public.pem
   ```

6. **Run database migrations**
   ```bash
   make migrate-up
   ```

7. **Build and run**
   ```bash
   make build
   ./bin/aci-backend
   ```

## Development

### Available Make Targets

```bash
make build           # Build the application
make test            # Run all tests with coverage
make test-unit       # Run unit tests only
make test-integration # Run integration tests only
make lint            # Run golangci-lint
make fmt             # Format code
make tidy            # Tidy Go modules
make migrate-up      # Run migrations up
make migrate-down    # Rollback migrations
make migrate-create  # Create new migration
make docker-build    # Build Docker image
make docker-up       # Start Docker Compose
make docker-down     # Stop Docker Compose
make run             # Run application locally
make clean           # Clean build artifacts
make help            # Show all targets
```

### Running Tests

```bash
# All tests with coverage
make test

# Unit tests only
make test-unit

# Integration tests only
make test-integration
```

### Code Quality

```bash
# Run linter
make lint

# Format code
make fmt
```

### Database Migrations

```bash
# Create a new migration
make migrate-create

# Run migrations up
make migrate-up

# Rollback migrations
make migrate-down
```

## Docker Development

```bash
# Build Docker image
make docker-build

# Start services (PostgreSQL, Redis, app)
make docker-up

# View logs
make docker-logs

# Stop services
make docker-down
```

## Configuration

Configuration is loaded from environment variables. See `.env.example` for all available options.

Required environment variables:
- `DATABASE_URL` - PostgreSQL connection string
- `JWT_PRIVATE_KEY_PATH` - Path to JWT private key
- `JWT_PUBLIC_KEY_PATH` - Path to JWT public key
- `N8N_WEBHOOK_SECRET` - Secret for n8n webhook authentication
- `ANTHROPIC_API_KEY` - Anthropic API key for AI features

## Project Status

🚧 **Under Development** - Project structure created, implementation in progress.

## Next Steps

1. Implement domain entities (User, Article, Alert)
2. Implement repository layer with PostgreSQL
3. Implement service layer with business logic
4. Implement HTTP handlers and middleware
5. Implement WebSocket real-time updates
6. Implement AI integration with Anthropic
7. Write comprehensive tests
8. Create API documentation

## License

Proprietary - All rights reserved
