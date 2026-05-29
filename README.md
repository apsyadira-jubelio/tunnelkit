# TunnelKit

Self-hosted tunneling platform — Ngrok alternative.

## Quick Start

```bash
# Start services (PostgreSQL + Redis + Server + Web Dashboard)
cd infra
docker compose up -d

# Or run locally
pnpm install
pnpm dev
```

## Components

| Component | Path | Tech |
|-----------|------|------|
| Server | `services/server/` | Go, Echo v4, PostgreSQL, Redis |
| CLI Client | `apps/cli/` | Go, Cobra, WebSocket, yamux |
| Web Dashboard | `apps/web/` | React, TanStack Start, shadcn/ui |

## Development

### Server

```bash
cd services/server
make build
make run
```

### CLI

```bash
cd apps/cli
go build -o tunnelkit ./cmd

# Login
./tunnelkit login --email dev@test.com --password test123

# Open HTTP tunnel
./tunnelkit http 3000 --subdomain myapp
```

### Web Dashboard

```bash
cd apps/web
pnpm install
pnpm dev
# Open http://localhost:3000
```

## API Endpoints

```
POST   /api/v1/auth/login
POST   /api/v1/auth/register
GET    /api/v1/auth/me

GET    /api/v1/tunnels
POST   /api/v1/tunnels
GET    /api/v1/tunnels/:id
DELETE /api/v1/tunnels/:id

GET    /api/v1/api-keys
POST   /api/v1/api-keys
DELETE /api/v1/api-keys/:id

WS     /ws/agent
```

## Database

```bash
# Run migrations
cd services/server
make migrate
```

## Architecture

```
Client (Browser/CLI) → Public Server (Go) → WebSocket → TunnelKit CLI → Local Service
```

## License

MIT
