# TunnelKit CLI

## Installation

```bash
go install github.com/tunnelkit/apps/cli@latest
```

## Commands

### Login
```bash
tunnelkit login --email user@example.com --password password123
```

### HTTP Tunnel
```bash
# Basic
tunnelkit http 3000

# With custom subdomain
tunnelkit http 3000 --subdomain myapp

# With custom server
tunnelkit http 3000 --server https://tunnel.example.com
```

### TCP Tunnel
```bash
# Basic
tunnelkit tcp 5432

# With custom remote port
tunnelkit tcp 22 --remote-port 2222
```

### List Tunnels
```bash
tunnelkit list
```

### Status
```bash
tunnelkit status
```

## Configuration

Config file: `~/.tunnelkit.yml`

```yaml
server: https://tunnel.example.com
token: your-api-token-here
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `TUNNELKIT_SERVER` | Server URL | `https://tunnel.localhost` |
| `TUNNELKIT_TOKEN` | API token | - |
| `TUNNELKIT_CONFIG` | Config file path | `~/.tunnelkit.yml` |
