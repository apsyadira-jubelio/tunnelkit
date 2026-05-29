# TunnelKit Infrastructure

## Local Development

```bash
# Start all services
docker compose up -d

# View logs
docker compose logs -f

# Stop
docker compose down
```

## Services

| Service | Port | Description |
|---------|------|-------------|
| PostgreSQL | 5432 | Main database |
| Redis | 6379 | Cache & pub/sub |
| Server | 8080 | API + tunnel server |
| Web | 3000 | Dashboard UI |

## Database

```bash
# Run migrations
docker compose exec server ./scripts/migrate.sh

# Connect to psql
docker compose exec postgres psql -U tunnelkit -d tunnelkit
```

## Production

```bash
# Build server binary
docker compose build server

# Or build locally
cd ../services/server
make build
```

## NGINX (Optional)

For wildcard subdomain routing:

```bash
# Add to /etc/hosts
127.0.0.1 tunnel.localhost
*.tunnel.localhost

# Run nginx
docker compose --profile ingress up nginx
```
