# Armor Cyber News

An automated cybersecurity news aggregation platform that collects, enriches, and delivers critical security information from trusted sources like CISA using a modern full-stack architecture.

## Table of Contents

- [Overview](#overview)
- [Key Features](#key-features)
- [Project Structure](#project-structure)
- [Tech Stack](#tech-stack)
- [Quick Start](#quick-start)
- [Component Documentation](#component-documentation)
- [Architecture](#architecture)
- [Development](#development)
- [Deployment](#deployment)
- [Contributing](#contributing)

## Overview

n8n Cyber News automates the collection and aggregation of cybersecurity threat information from multiple sources (CISA, security feeds, etc.) and delivers curated, enriched alerts to users. The system combines automated n8n workflows for data collection with a modern full-stack application for viewing and managing alerts.

The platform addresses the need for:
- Centralized security threat monitoring from multiple sources
- Automated enrichment with AI and contextual data
- Real-time alert delivery with WebSocket support
- Enterprise-grade security and authentication

## Key Features

- **Automated News Aggregation**: n8n workflows continuously pull from CISA and other security feeds
- **Multi-source Integration**: Support for RSS feeds, CISA alerts, and custom security sources
- **AI Enrichment**: Claude AI integration for summarizing and enriching threat information
- **Real-time Alerts**: WebSocket support for live alert streaming to authenticated users
- **JWT Authentication**: Secure API access with token-based authentication
- **Responsive UI**: Modern React frontend with real-time data visualization
- **Kubernetes Ready**: Production-grade deployment configurations
- **PostgreSQL & Redis**: Persistent storage with caching layer

## Project Structure

```
n8n-cyber-news/
├── aci-backend/                 # Go REST API server
│   ├── cmd/                      # Application entry points
│   ├── internal/
│   │   ├── api/                  # HTTP handlers and routes
│   │   ├── database/             # PostgreSQL operations
│   │   ├── auth/                 # JWT authentication
│   │   ├── websocket/            # Real-time alert streaming
│   │   └── models/               # Data structures
│   ├── migrations/               # Database schema migrations
│   ├── deployments/              # Kubernetes manifests
│   ├── go.mod                    # Go dependencies
│   └── README.md                 # Backend documentation
│
├── aci-frontend/                # React + TypeScript frontend
│   ├── src/
│   │   ├── components/           # Reusable React components
│   │   ├── pages/                # Page components
│   │   ├── hooks/                # Custom React hooks
│   │   ├── services/             # API clients
│   │   └── types/                # TypeScript interfaces
│   ├── package.json              # Dependencies and scripts
│   ├── vite.config.ts            # Vite configuration
│   └── README.md                 # Frontend documentation
│
├── workflows/                    # n8n workflow definitions
│   ├── cisa-aggregator.json      # CISA feed aggregation
│   ├── alert-enrichment.json     # AI enrichment workflow
│   └── README.md                 # Workflow documentation
│
├── specs/                        # Feature specifications and design docs
│   ├── authentication.md         # Auth system design
│   ├── n8n-integration.md        # n8n workflow architecture
│   ├── websocket-protocol.md     # Real-time communication spec
│   └── 001-aci-backend/          # Backend implementation specs
│
└── tests/                        # E2E and integration tests
```

## Tech Stack

### Backend
- **Language**: Go 1.21+
- **Framework**: Gin (REST API)
- **Database**: PostgreSQL 14+
- **Cache**: Redis 6+
- **Auth**: JWT tokens
- **Deployment**: Docker, Kubernetes
- **API**: RESTful with WebSocket support

### Frontend
- **Framework**: React 19.2
- **Language**: TypeScript 5.9
- **Build Tool**: Vite 7.2
- **UI Components**: shadcn/ui (Radix UI + Tailwind CSS)
- **State Management**: TanStack Query v5
- **Routing**: react-router-dom v7
- **Styling**: Tailwind CSS
- **Visualization**: Reviz

### Workflow Automation
- **Orchestration**: n8n
- **Trigger**: RSS feeds, HTTP polls
- **Enrichment**: Claude AI API
- **Storage**: PostgreSQL
- **Cache**: Redis

## Quick Start

### Prerequisites

Before getting started, ensure you have installed:

- **Docker** and **Docker Compose** (for local development)
- **Go 1.21+** (for backend development)
- **Node.js 18+** and **npm** (for frontend development)
- **PostgreSQL 14+** (if not using Docker)
- **Redis 6+** (if not using Docker)
- **n8n** (self-hosted or cloud instance for workflows)

### Option 1: Docker Compose (Recommended)

Run the entire stack locally with Docker:

```bash
# Clone the repository
git clone https://github.com/phillipboles/n8n-cyber-news.git
cd n8n-cyber-news

# Start all services (backend, frontend, PostgreSQL, Redis)
docker-compose up -d

# Check service status
docker-compose ps
```

Access the services:
- Frontend: http://localhost:5173
- Backend API: http://localhost:8080
- Health check: http://localhost:8080/health

### Option 2: Manual Setup

#### Backend Setup

```bash
cd aci-backend

# Create .env file with configuration
cp .env.example .env
# Edit .env with your database and authentication settings

# Install dependencies
go mod download

# Run database migrations
go run cmd/main.go migrate

# Start the server
go run cmd/main.go serve

# Server runs on http://localhost:8080
```

See [aci-backend/README.md](./aci-backend/README.md) for detailed backend setup.

#### Frontend Setup

```bash
cd aci-frontend

# Install dependencies
npm install

# Create environment file
cp .env.example .env.local
# Edit .env.local with API endpoint

# Start development server
npm run dev

# Access at http://localhost:5173
```

See [aci-frontend/README.md](./aci-frontend/README.md) for detailed frontend setup.

#### n8n Workflows

1. Deploy n8n (cloud or self-hosted)
2. Import workflow definitions from `workflows/` directory
3. Configure API keys and credentials in n8n
4. Enable auto-activation for continuous data collection

See [workflows/README.md](./workflows/README.md) for workflow setup details.

## Component Documentation

Each component has detailed documentation:

| Component | Purpose | Documentation |
|-----------|---------|-----------------|
| **Backend** | REST API, database, auth, WebSocket server | [aci-backend/README.md](./aci-backend/README.md) |
| **Frontend** | React UI, real-time dashboard | [aci-frontend/README.md](./aci-frontend/README.md) |
| **Workflows** | RSS aggregation, AI enrichment | [workflows/README.md](./workflows/README.md) |
| **Deployment** | Kubernetes manifests, deployment configs | [aci-backend/deployments/README.md](./aci-backend/deployments/README.md) |
| **Database** | Schema, migrations, queries | [aci-backend/migrations/README.md](./aci-backend/migrations/README.md) |

## Architecture

### System Overview

```
External Sources
├─ CISA Feeds (RSS)
├─ Security Feeds
└─ Custom Sources
        │
        v
   n8n Workflows
   ├─ Aggregation
   ├─ Enrichment
   └─ Storage
        │
        v
PostgreSQL Database
        │
        v
Go REST API
├─ HTTP Endpoints
├─ WebSocket Streaming
└─ JWT Auth
        │
        v
React Frontend
├─ Dashboard
├─ Alert Management
└─ Real-time Updates
```

### Data Flow

1. **Collection**: n8n workflows poll RSS feeds and CISA APIs on schedule
2. **Enrichment**: Raw alerts are enriched with AI summaries using Claude
3. **Storage**: Enriched data stored in PostgreSQL with Redis cache
4. **API**: Go backend exposes REST endpoints for frontend consumption
5. **Streaming**: WebSocket connections deliver real-time alert updates
6. **Display**: React frontend displays alerts with filtering, search, and tagging

### Authentication Flow

1. User logs in with credentials
2. Backend validates against user database
3. JWT token generated and returned to client
4. Client includes token in subsequent API requests
5. WebSocket connections authenticated via token parameter
6. Protected routes and endpoints validate token before processing

## Development

### Running Tests

```bash
# Backend tests
cd aci-backend
go test ./...

# Frontend tests
cd aci-frontend
npm test

# E2E tests
cd tests
npm test
```

### Code Style

- **Backend**: Follow Go conventions (gofmt, golint)
- **Frontend**: TypeScript with ESLint (check with `npm run lint`)

Run linters:

```bash
# Backend
cd aci-backend
golangci-lint run

# Frontend
cd aci-frontend
npm run lint
npm run lint:fix  # Auto-fix issues
```

### Development Workflow

1. Create a feature branch: `git checkout -b feature/your-feature`
2. Make your changes following the code style guidelines
3. Run tests: `npm test` or `go test ./...`
4. Run linter: `npm run lint` or `golangci-lint run`
5. Commit with clear messages: `git commit -m "feat: description"`
6. Push to your fork: `git push origin feature/your-feature`
7. Open a Pull Request for review

### Environment Variables

#### Backend (.env)

```bash
# Server
PORT=8080
ENV=development

# Database
DATABASE_URL=postgresql://user:password@localhost:5432/cyber_news
REDIS_URL=redis://localhost:6379

# Authentication
JWT_SECRET=your-secret-key-change-in-production
JWT_EXPIRY=24h

# n8n Integration
N8N_API_URL=http://localhost:5678
N8N_API_KEY=your-n8n-api-key

# AI Enrichment
CLAUDE_API_KEY=your-claude-api-key

# Logging
LOG_LEVEL=debug
```

#### Frontend (.env.local)

```bash
VITE_API_URL=http://localhost:8080
VITE_WS_URL=ws://localhost:8080
```

## Deployment

### Docker

The project includes Dockerfile for containerized deployment:

```bash
# Build backend image
cd aci-backend
docker build -t n8n-cyber-news-backend:latest .

# Build frontend image
cd aci-frontend
docker build -t n8n-cyber-news-frontend:latest .
```

### Kubernetes

Production deployment manifests included:

```bash
# Apply manifests
kubectl apply -f aci-backend/deployments/

# Verify deployment
kubectl get pods -l app=cyber-news
kubectl get services
kubectl logs -f deployment/cyber-news-backend
```

See [aci-backend/deployments/README.md](./aci-backend/deployments/README.md) for complete deployment guide.

### Environment Considerations

**Development**:
- All services run locally via Docker Compose
- Hot-reload enabled for frontend
- Debug logging enabled
- Mock data available

**Staging**:
- Services deployed to staging cluster
- Real n8n instance for workflow testing
- Reduced logging verbosity
- Performance testing enabled

**Production**:
- High-availability Kubernetes setup
- SSL/TLS termination at ingress
- Production database backups
- Monitoring and alerting enabled
- Rate limiting and DDoS protection

## Contributing

We welcome contributions! Please see our guidelines:

1. **Fork the repository**
2. **Create a feature branch** for your work
3. **Follow code style** conventions
4. **Add tests** for new functionality
5. **Update documentation** as needed
6. **Submit a Pull Request** with clear description

### Reporting Issues

Found a bug? Please open an issue with:
- Clear title and description
- Steps to reproduce
- Expected vs actual behavior
- Environment details (OS, versions, etc.)
- Screenshots or logs if applicable

### Getting Help

- Check [existing issues](https://github.com/phillipboles/n8n-cyber-news/issues)
- Review component READMEs
- Check specification documents in `specs/`
- Join discussions in project repository

## License

This software is proprietary and confidential. Copyright (c) 2025 Armor, Inc. All rights reserved. See [LICENSE](./LICENSE) for details.

## Support

For questions or issues:

- Open an issue on GitHub
- Check component documentation
- Review specification documents
- Contact the development team

## Related Documentation

- [Backend Integration Notes](./aci-backend/INTEGRATION_NOTES.md)
- [Authentication Spec](./specs/authentication.md)
- [n8n Integration Spec](./specs/n8n-integration.md)
- [WebSocket Protocol](./specs/websocket-protocol.md)

---

**Last Updated**: December 2025

For the latest updates, see recent commits and pull requests.
