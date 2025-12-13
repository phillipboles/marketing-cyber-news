# Quickstart: ACI Backend

**Date**: 2025-12-11
**Feature**: 001-aci-backend

---

## Prerequisites

| Tool | Version | Installation |
|------|---------|--------------|
| Go | 1.25+ | https://go.dev/dl/ |
| Docker | 24+ | https://docs.docker.com/get-docker/ |
| Docker Compose | v2+ | Included with Docker Desktop |
| Make | Any | Pre-installed on macOS/Linux |
| PostgreSQL Client | 18+ | `brew install postgresql@18` |

---

## Quick Start (5 minutes)

### 1. Clone Repository

```bash
git clone https://github.com/armor/aci-backend.git
cd aci-backend
```

### 2. Generate JWT Keys

```bash
mkdir -p keys
openssl genrsa -out keys/private.pem 2048
openssl rsa -in keys/private.pem -pubout -out keys/public.pem
chmod 600 keys/private.pem
```

### 3. Configure Environment

```bash
cp .env.example .env.local
```

Edit `.env.local` with required values:

```bash
# Required
DATABASE_URL=postgres://aci:aci_password@localhost:5432/aci?sslmode=disable
JWT_PRIVATE_KEY_PATH=./keys/private.pem
JWT_PUBLIC_KEY_PATH=./keys/public.pem
N8N_WEBHOOK_SECRET=your-32-char-minimum-secret-key-here
ANTHROPIC_API_KEY=sk-ant-your-key-here

# Optional (defaults shown)
SERVER_PORT=8080
LOG_LEVEL=debug
```

### 4. Start Services

```bash
# Start PostgreSQL and Redis
make docker-up

# Wait for PostgreSQL to be ready
docker-compose logs -f postgres
# Look for: "database system is ready to accept connections"
```

### 5. Run Migrations

```bash
make migrate-up
```

### 6. Seed Data

```bash
make seed
```

### 7. Start Server

```bash
make run
```

### 8. Verify Installation

```bash
# Health check
curl http://localhost:8080/v1/health

# Expected response:
# {"status":"healthy","version":"1.0.0","timestamp":"..."}

# Readiness check
curl http://localhost:8080/v1/ready

# Expected response:
# {"status":"ready","checks":{"database":"ok","redis":"ok"}}
```

---

## Development Workflow

### Running with Hot Reload

Install [air](https://github.com/air-verse/air) for hot reload:

```bash
go install github.com/air-verse/air@latest
make dev
```

### Running Tests

```bash
# All tests
make test

# With coverage
make test-coverage
# Open coverage.html in browser

# Integration tests (requires running database)
make test-integration
```

### Linting

```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

make lint
```

### Database Operations

```bash
# Create new migration
make migrate-create name=add_new_table

# Run all pending migrations
make migrate-up

# Rollback last migration
make migrate-down

# Check migration status
migrate -path migrations -database "$DATABASE_URL" version
```

---

## API Examples

### Authentication

#### Register User

```bash
curl -X POST http://localhost:8080/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!",
    "name": "John Doe"
  }'
```

Response:
```json
{
  "user": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "email": "user@example.com",
    "name": "John Doe",
    "role": "user",
    "created_at": "2025-12-11T10:00:00Z"
  },
  "tokens": {
    "access_token": "eyJ...",
    "refresh_token": "abc123...",
    "expires_at": "2025-12-11T10:15:00Z"
  }
}
```

#### Login

```bash
curl -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!"
  }'
```

#### Refresh Token

```bash
curl -X POST http://localhost:8080/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "abc123..."
  }'
```

### Articles

#### List Articles

```bash
curl http://localhost:8080/v1/articles \
  -H "Authorization: Bearer <access_token>"
```

Query parameters:
- `page` (default: 1)
- `limit` (default: 20, max: 100)
- `category` (category slug)
- `severity` (critical/high/medium/low/informational)
- `sort` (published_at/relevance, default: published_at)
- `order` (asc/desc, default: desc)

#### Get Single Article

```bash
curl http://localhost:8080/v1/articles/<article_id> \
  -H "Authorization: Bearer <access_token>"
```

#### Search Articles

```bash
curl "http://localhost:8080/v1/articles/search?q=ransomware&severity=critical" \
  -H "Authorization: Bearer <access_token>"
```

### Categories

```bash
curl http://localhost:8080/v1/categories \
  -H "Authorization: Bearer <access_token>"
```

### Alerts

#### Create Alert

```bash
curl -X POST http://localhost:8080/v1/alerts \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "VMware Vulnerabilities",
    "type": "vendor",
    "value": "vmware"
  }'
```

#### List Alerts

```bash
curl http://localhost:8080/v1/alerts \
  -H "Authorization: Bearer <access_token>"
```

### WebSocket

#### Connect

```javascript
const ws = new WebSocket('ws://localhost:8080/ws?token=<access_token>');

ws.onopen = () => {
  console.log('Connected');

  // Subscribe to channels
  ws.send(JSON.stringify({
    type: 'subscribe',
    id: crypto.randomUUID(),
    timestamp: new Date().toISOString(),
    payload: {
      channels: ['articles:critical', 'alerts:user']
    }
  }));
};

ws.onmessage = (event) => {
  const msg = JSON.parse(event.data);
  console.log('Received:', msg.type, msg.payload);
};
```

---

## Docker Commands

```bash
# Start all services
make docker-run

# Stop all services
make docker-stop

# View logs
make docker-logs

# Rebuild after code changes
make docker-build
```

---

## Environment Variables Reference

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| DATABASE_URL | Yes | - | PostgreSQL connection string |
| JWT_PRIVATE_KEY_PATH | Yes | - | Path to RS256 private key |
| JWT_PUBLIC_KEY_PATH | Yes | - | Path to RS256 public key |
| N8N_WEBHOOK_SECRET | Yes | - | Shared secret for webhook verification |
| ANTHROPIC_API_KEY | Yes | - | Claude API key |
| SERVER_HOST | No | 0.0.0.0 | Server bind address |
| SERVER_PORT | No | 8080 | Server port |
| SERVER_READ_TIMEOUT | No | 30s | HTTP read timeout |
| SERVER_WRITE_TIMEOUT | No | 30s | HTTP write timeout |
| REDIS_URL | No | - | Redis connection string (optional) |
| JWT_ACCESS_TOKEN_EXPIRY | No | 15m | Access token lifetime |
| JWT_REFRESH_TOKEN_EXPIRY | No | 168h | Refresh token lifetime (7 days) |
| CORS_ALLOWED_ORIGINS | No | http://localhost:3000 | Allowed CORS origins |
| RATE_LIMIT_REQUESTS | No | 100 | Requests per window |
| RATE_LIMIT_WINDOW | No | 1m | Rate limit window |
| LOG_LEVEL | No | info | Log level (debug/info/warn/error) |
| LOG_FORMAT | No | json | Log format (json/console) |
| FEATURE_SEMANTIC_SEARCH | No | true | Enable vector search |
| FEATURE_AI_ENRICHMENT | No | true | Enable AI enrichment |

---

## Troubleshooting

### Database Connection Failed

```bash
# Check PostgreSQL is running
docker-compose ps

# Check PostgreSQL logs
docker-compose logs postgres

# Verify connection string
psql "$DATABASE_URL" -c "SELECT 1"
```

### Migration Errors

```bash
# Check current version
migrate -path migrations -database "$DATABASE_URL" version

# Force specific version (use with caution)
migrate -path migrations -database "$DATABASE_URL" force <version>
```

### JWT Key Issues

```bash
# Verify key format
openssl rsa -in keys/private.pem -check
openssl rsa -pubin -in keys/public.pem -text

# Regenerate if corrupted
rm -rf keys
mkdir keys
openssl genrsa -out keys/private.pem 2048
openssl rsa -in keys/private.pem -pubout -out keys/public.pem
```

### Port Already in Use

```bash
# Find process using port 8080
lsof -i :8080

# Kill process
kill -9 <PID>
```

---

## Common Makefile Commands

| Command | Description |
|---------|-------------|
| `make build` | Build binary |
| `make run` | Run server |
| `make dev` | Run with hot reload |
| `make test` | Run all tests |
| `make test-coverage` | Run tests with coverage |
| `make test-integration` | Run integration tests |
| `make lint` | Run linter |
| `make fmt` | Format code |
| `make migrate-up` | Run migrations |
| `make migrate-down` | Rollback migration |
| `make seed` | Seed database |
| `make docker-build` | Build Docker image |
| `make docker-run` | Start containers |
| `make docker-stop` | Stop containers |
| `make docker-logs` | View container logs |
| `make clean` | Clean build artifacts |
| `make help` | Show all commands |

---

## Next Steps

1. **Explore API**: Use the examples above to test endpoints
2. **Set up n8n**: Configure n8n workflows for content scraping
3. **Configure Claude**: Test AI enrichment with sample articles
4. **Deploy**: See `/deployments/k8s/` for Kubernetes manifests
