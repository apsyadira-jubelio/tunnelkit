# TunnelKit API

## Authentication

### Login
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123"
}
```

### Register
```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "email": "newuser@example.com",
  "password": "password123"
}
```

### Get Current User
```http
GET /api/v1/auth/me
Authorization: Bearer <token>
```

## Tunnels

### List Tunnels
```http
GET /api/v1/tunnels
Authorization: Bearer <token>
```

### Create Tunnel
```http
POST /api/v1/tunnels
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "my-app",
  "protocol": "http",
  "subdomain": "myapp",
  "auth_type": "none"
}
```

### Delete Tunnel
```http
DELETE /api/v1/tunnels/:id
Authorization: Bearer <token>
```

## API Keys

### List API Keys
```http
GET /api/v1/api-keys
Authorization: Bearer <token>
```

### Create API Key
```http
POST /api/v1/api-keys
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "my-key",
  "scopes": ["tunnel:read", "tunnel:write"]
}
```

### Revoke API Key
```http
DELETE /api/v1/api-keys/:id
Authorization: Bearer <token>
```

## WebSocket

### Agent Connection
```javascript
const ws = new WebSocket('ws://localhost:8080/ws/agent')

ws.onopen = () => {
  ws.send(JSON.stringify({
    type: 'hello',
    version: '1.0',
    tunnel_id: 'tunnel-uuid'
  }))
}

ws.onmessage = (event) => {
  const msg = JSON.parse(event.data)
  // Handle: hello, request, ping, pong, error
}
```

## Error Responses

All errors return:
```json
{
  "message": "error description"
}
```

Status codes:
- `400` - Bad Request
- `401` - Unauthorized
- `403` - Forbidden
- `404` - Not Found
- `409` - Conflict
- `500` - Internal Server Error
