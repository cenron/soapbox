# soapbox

A pre-2022 Twitter clone — chronological microblogging platform. No algorithms, no ads, just posts in order.

Built as a Go modular monolith with a React SPA frontend. Designed to scale into microservices when needed.

## Prerequisites

- [Go 1.23+](https://go.dev/dl/)
- [Docker](https://docs.docker.com/get-docker/) (for Postgres, MinIO, Mailpit)
- [Node.js 20+](https://nodejs.org/) (for the frontend, coming in Phase 0B)

## Setup

```bash
# Clone the repo
git clone git@github.com:radni/soapbox.git
cd soapbox

# Copy environment config
cp .env.example .env

# Install Go dev tools
go install github.com/air-verse/air@latest
go install github.com/swaggo/swag/cmd/swag@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Start dev infrastructure (Postgres, MinIO, Mailpit)
make docker-up

# Run the server (hot-reload via air)
make run
```

The server starts at `http://localhost:8080`.

## Make targets

| Command | Description |
|---------|-------------|
| `make run` | Start dev server with hot-reload (air) |
| `make build` | Build the binary to `bin/web` |
| `make test` | Run all tests |
| `make lint` | Run golangci-lint |
| `make swagger` | Regenerate Swagger docs |
| `make docker-up` | Start Postgres, MinIO, Mailpit |
| `make docker-down` | Stop dev infrastructure |

## Dev services

| Service | URL | Purpose |
|---------|-----|---------|
| API | http://localhost:8080 | Go backend |
| Swagger UI | http://localhost:8080/swagger/index.html | API documentation |
| Health check | http://localhost:8080/healthz | Liveness probe |
| MinIO Console | http://localhost:9001 | Object storage admin |
| Mailpit | http://localhost:8025 | Email testing UI |

## Project structure

```
soapbox/
├── cmd/web/main.go              # Composition root
├── internal/
│   ├── core/                    # Shared infrastructure
│   │   ├── bus/                 # Event + query bus
│   │   ├── cache/               # Cache interface + in-memory impl
│   │   ├── config/              # Env-based config
│   │   ├── db/                  # Postgres pool, migrations, transactions
│   │   ├── httpkit/             # Chi router, middleware, response helpers
│   │   ├── registry/            # Service discovery
│   │   ├── testutil/            # Test helpers and mocks
│   │   └── types/               # IDs, pagination, errors
│   ├── auth/                    # Auth module
│   ├── users/                   # User profiles and follows
│   ├── posts/                   # Posts, likes, reposts, hashtags
│   ├── feed/                    # Timeline assembly
│   ├── notifications/           # In-app notifications
│   ├── media/                   # S3 uploads
│   └── moderation/              # Reports, blocks, mutes
├── web/                         # React SPA (Vite + shadcn/ui + Tailwind)
├── api/swagger/                 # Generated Swagger docs
├── build/                       # Dockerfile + entrypoint
├── docker-compose.yml           # Dev infra
└── docs/                        # Specs, plan, architecture decisions
```

## Architecture

Modules are fully isolated — no cross-module imports. All inter-module communication goes through the bus (events for async, queries for sync). Each module owns its own Postgres schema.

See `docs/specs/2026-03-21-soapbox-design.md` for the full design specification.

## Tech stack

| Layer | Technology |
|-------|-----------|
| Backend | Go, chi, sqlx, pgx, goose |
| Frontend | React, Vite, shadcn/ui, Tailwind CSS |
| Database | PostgreSQL 16 (schema-per-module) |
| Object storage | S3-compatible (MinIO for dev) |
| Dev email | Mailpit |
