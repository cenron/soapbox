# soapbox

A pre-2022 Twitter clone — chronological microblogging platform. No algorithms, no ads, just posts in order.

Built as a Go modular monolith with a React SPA frontend. Designed to scale into microservices when needed.

## Prerequisites

- [Go 1.23+](https://go.dev/dl/)
- [Docker](https://docs.docker.com/get-docker/) (for Postgres, MinIO, Mailpit)
- [Node.js 20+](https://nodejs.org/) (for the frontend)

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

# Install frontend dependencies
cd web && npm install && cd ..

# Start dev infrastructure (Postgres, MinIO, Mailpit)
make docker-up

# Run the backend (hot-reload via air)
make run

# In a separate terminal, run the frontend dev server
make web-dev
```

The backend starts at `http://localhost:8080`. The frontend dev server starts at `http://localhost:5173` and proxies API requests to the backend.

## Make targets

| Command | Description |
|---------|-------------|
| `make run` | Start backend dev server with hot-reload (air) |
| `make build` | Build frontend + Go binary to `bin/web` |
| `make test` | Run all tests (frontend + backend) |
| `make lint` | Run all linters (ESLint + golangci-lint) |
| `make swagger` | Regenerate Swagger docs |
| `make docker-up` | Start Postgres, MinIO, Mailpit |
| `make docker-down` | Stop dev infrastructure |
| `make web-install` | Install frontend dependencies |
| `make web-dev` | Start Vite dev server (port 5173) |
| `make web-build` | Build frontend to `web/dist/` |
| `make web-test` | Run frontend tests (Vitest) |
| `make web-lint` | Run frontend linter (ESLint) |

## Dev services

| Service | URL | Purpose |
|---------|-----|---------|
| Frontend (dev) | http://localhost:5173 | Vite dev server with HMR |
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
│   └── modules/                 # Feature modules (isolated, communicate via bus)
│       ├── auth/                # Auth module
│       ├── users/               # User profiles and follows
│       ├── posts/               # Posts, likes, reposts, hashtags
│       ├── feed/                # Timeline assembly
│       ├── notifications/       # In-app notifications
│       ├── media/               # S3 uploads
│       └── moderation/          # Reports, blocks, mutes
├── web/                         # React SPA (Vite + shadcn/ui + Tailwind)
│   ├── src/
│   │   ├── shared/              # API client, auth context, WebSocket, UI components
│   │   ├── layouts/             # App shell (nav bar, sidebar)
│   │   ├── pages/               # Route pages
│   │   └── modules/             # Feature-specific frontend code
│   └── embed.go                 # go:embed for production SPA serving
├── api/swagger/                 # Generated Swagger docs
├── build/                       # Dockerfile + entrypoint
├── docker-compose.yml           # Dev infra
└── docs/                        # Specs, plan, architecture decisions
```

## Architecture

Modules are fully isolated — no cross-module imports. All inter-module communication goes through the bus (events for async, queries for sync). Each module owns its own Postgres schema.

See `docs/specs/2026-03-21-soapbox-design.md` for the full design specification.

## Frontend

The frontend is a React SPA in `web/`, built with Vite and embedded into the Go binary for production.

**Stack:** React 19, TypeScript, Tailwind CSS v4, shadcn/ui, React Router v7, TanStack Query v5

**Development:** `make web-dev` runs the Vite dev server with HMR. API calls (`/api`, `/ws`, `/swagger`, `/healthz`) are proxied to the Go backend at `:8080`.

**Production:** `make build` compiles the frontend into `web/dist/`, then embeds it into the Go binary via `go:embed`. The Go server serves the SPA and falls back to `index.html` for client-side routing.

**Testing:** Vitest + React Testing Library. Run with `make web-test`.

**Adding UI components:** `cd web && npx shadcn@latest add <component>` — components land in `src/shared/ui/`.

## Tech stack

| Layer | Technology |
|-------|-----------|
| Backend | Go, chi, sqlx, pgx, goose |
| Frontend | React 19, Vite, TypeScript, shadcn/ui, Tailwind CSS v4 |
| Routing | React Router v7 (declarative mode) |
| Server state | TanStack Query v5 |
| Testing | Vitest, React Testing Library (frontend); Go test (backend) |
| Database | PostgreSQL 16 (schema-per-module) |
| Object storage | S3-compatible (MinIO for dev) |
| Dev email | Mailpit |
