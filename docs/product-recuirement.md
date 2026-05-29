# TunnelKit — Product Requirements Document

> **Product:** TunnelKit | **Status:** Draft v0.1 | **Created:** May 2026
> **Stack:** Go · TanStack · TypeScript · shadcn/ui | **Classification:** Confidential Draft

---

## 1. Product Overview

TunnelKit adalah self-hosted tunneling platform yang memungkinkan developer mengekspos local service ke internet melalui encrypted tunnel, sebagai alternatif Ngrok yang dapat di-deploy sendiri (on-premise atau cloud).

> **Vision:** _"Memberikan developer kontrol penuh atas infrastruktur tunnel mereka — tanpa vendor lock-in, tanpa rate limit tersembunyi, dan dengan observability first-class."_

### 1.1 Problem Statement

Ngrok dan layanan serupa memiliki batasan struktural yang menjadi blocker bagi tim engineering skala menengah ke atas:

- Custom domain terkunci di tier berbayar dengan harga tidak transparan
- Data tunnel melewati server pihak ketiga — risiko compliance dan privacy
- Rate limiting agresif pada free/entry tier yang tidak cocok untuk development intensif
- Tidak ada kontrol atas log dan audit trail akses tunnel
- Sulit diintegrasikan ke pipeline CI/CD internal

### 1.2 Solution

TunnelKit menyediakan:

- **Server komponen (Go):** menerima koneksi dari client, meneruskan traffic ke public endpoint
- **Client CLI (Go):** dijalankan di mesin developer atau CI runner, membuka tunnel ke server
- **Dashboard Web (TanStack + shadcn/ui):** manajemen tunnel, user, domain, monitoring real-time, dan audit log
- **API REST (Go):** semua operasi dapat diakses programatik, cocok untuk integrasi CI/CD

---

## 2. Goals & Non-Goals

### 2.1 Goals (In Scope)

| #   | Goal                                           | Ukuran Keberhasilan                                    |
| --- | ---------------------------------------------- | ------------------------------------------------------ |
| G1  | HTTP & HTTPS tunneling dengan latency rendah   | p99 latency tambahan < 20ms vs direct connection       |
| G2  | TCP tunneling untuk protokol non-HTTP          | Berhasil tunnel SSH, database, gRPC                    |
| G3  | Custom domain dengan auto TLS (Let's Encrypt)  | 100% tunnel dapat diakses via custom subdomain         |
| G4  | Dashboard web real-time dengan metrics & logs  | Refresh metrics < 2 detik via WebSocket/SSE            |
| G5  | Multi-user dengan RBAC (Admin, Member, Viewer) | Granular permission per resource tunnel                |
| G6  | API key management & audit log                 | Setiap akses tercatat dengan IP, timestamp, user agent |
| G7  | Self-hosted — single binary server deployment  | Server + agent berjalan dalam < 128MB RAM idle         |

### 2.2 Non-Goals (Out of Scope v1)

- GUI desktop app untuk client (CLI-first di v1)
- Billing / monetization layer
- Multi-region relay server (single server di v1)
- Browser DevTools inspection (a la Ngrok inspector) — roadmap v2
- Mobile app

---

## 3. Target Users & Persona

### Persona 1 — Backend/Fullstack Developer

| Atribut             | Detail                                                                                                                       |
| ------------------- | ---------------------------------------------------------------------------------------------------------------------------- |
| Konteks             | Perlu ekspos local dev server untuk testing webhook, demo ke klien, atau integrasi third-party (payment, OAuth, marketplace) |
| Pain point saat ini | Ngrok free tier terlalu dibatasi; ngrok berbayar mahal untuk tim kecil; tidak mau data tunnel lewat server eksternal         |
| Ekspektasi          | Bisa setup < 5 menit, URL stabil per session, bisa lihat log real-time di dashboard                                          |

### Persona 2 — DevOps / Platform Engineer

| Atribut             | Detail                                                                                 |
| ------------------- | -------------------------------------------------------------------------------------- |
| Konteks             | Mengelola infrastruktur tunnel untuk seluruh tim, perlu audit trail dan access control |
| Pain point saat ini | Tidak ada visibility siapa menggunakan tunnel apa, kapan, dari IP mana                 |
| Ekspektasi          | RBAC, audit log, revoke akses instan, deployment mudah via Docker/K8s                  |

---

## 4. Architecture Overview

### 4.1 High-Level Architecture

```
Traffic Flow:
Browser / Client Request  →  Public Server (Go)  →  Tunnel Channel (WebSocket/mux)  →  TunnelKit Client  →  Local Service

Control Plane:  Dashboard Web  ↔  REST API (Go)  ↔  PostgreSQL + Redis
Data Plane:     HTTP/TCP Request  ↔  Reverse Proxy (Go net/http)  ↔  Persistent WS Connection  ↔  Local Port
```

### 4.2 Core Components

| Komponen          | Teknologi                        | Tanggung Jawab                                              |
| ----------------- | -------------------------------- | ----------------------------------------------------------- |
| Tunnel Server     | Go (net/http, gorilla/websocket) | Terima koneksi client, routing HTTP/TCP, TLS termination    |
| Tunnel Client CLI | Go (cobra, yamux)                | Koneksi persistent ke server, forward traffic ke local port |
| API Server        | Go (Echo v4)                     | REST API untuk auth, tunnel CRUD, user management, metrics  |
| Web Dashboard     | TanStack Start + React           | UI management tunnel, real-time monitoring, settings        |
| UI Components     | shadcn/ui + Tailwind CSS v4      | Design system konsisten, accessible components              |
| Database          | PostgreSQL 16                    | User, tunnel, audit log, API key persistence                |
| Cache / PubSub    | Redis 7                          | Session, rate limiting, real-time event fan-out             |
| Proxy / Ingress   | Nginx (opsional)                 | Wildcard subdomain routing ke tunnel server                 |

### 4.3 Monorepo Structure

```
tunnelkit/                          # root monorepo
├── apps/
│   ├── web/                        # TanStack Start (dashboard)
│   │   ├── src/
│   │   │   ├── routes/             # TanStack Router file-based routes
│   │   │   ├── components/         # App-specific React components
│   │   │   ├── hooks/              # Custom hooks (useRealtime, useTunnels, ...)
│   │   │   └── lib/                # API client, query keys, utils
│   │   ├── package.json
│   │   └── vite.config.ts
│   └── cli/                        # Go tunnel client
│       ├── cmd/                    # cobra root & sub-commands
│       ├── internal/               # tunnel logic, multiplexer
│       └── main.go
├── services/
│   └── server/                     # Go tunnel + API server
│       ├── cmd/server/             # entrypoint
│       ├── internal/
│       │   ├── api/                # Echo handlers, middleware
│       │   ├── tunnel/             # WS accept, mux, proxy
│       │   ├── domain/             # business logic (hexagonal)
│       │   ├── repository/         # PostgreSQL (pgx), Redis
│       │   └── config/             # env, flags
│       └── go.mod
├── packages/
│   ├── ui/                         # shadcn/ui shared component lib
│   │   ├── src/components/ui/      # Button, Badge, Card, Table, ...
│   │   └── package.json
│   ├── api-types/                  # Shared TypeScript API types (openapi-ts)
│   └── config/                     # Shared ESLint, TS, Tailwind configs
├── infra/
│   ├── docker-compose.yml          # local dev stack
│   ├── Dockerfile.server           # Go multi-stage build
│   └── nginx/                      # wildcard subdomain config
├── scripts/                        # codegen, migration helpers
├── pnpm-workspace.yaml
└── turbo.json                      # Turborepo task graph
```

---

## 5. Functional Requirements

### 5.1 Tunnel Management

| ID     | Feature              | Deskripsi                                                                | Priority |
| ------ | -------------------- | ------------------------------------------------------------------------ | -------- |
| TUN-01 | Create HTTP Tunnel   | Client membuka tunnel, server assign subdomain unik (random atau custom) | P0       |
| TUN-02 | Create TCP Tunnel    | Forward port TCP arbitrer (SSH, DB, gRPC)                                | P0       |
| TUN-03 | Custom Subdomain     | User bisa request subdomain spesifik jika tersedia & authorized          | P0       |
| TUN-04 | Persistent Subdomain | Subdomain tetap sama antar restart (dikonfigurasi per user)              | P1       |
| TUN-05 | Tunnel Inspect       | Dashboard menampilkan request/response log real-time per tunnel          | P1       |
| TUN-06 | Replay Request       | Replay HTTP request dari log (a la Ngrok Inspector)                      | P2       |
| TUN-07 | Tunnel Password      | Basic auth pada public URL tunnel                                        | P1       |
| TUN-08 | IP Allowlist         | Restrict akses tunnel ke CIDR tertentu                                   | P1       |

### 5.2 Authentication & Authorization

| ID      | Feature                | Deskripsi                                                            | Priority |
| ------- | ---------------------- | -------------------------------------------------------------------- | -------- |
| AUTH-01 | Email + Password Login | JWT-based auth, refresh token via httpOnly cookie                    | P0       |
| AUTH-02 | API Key                | Generate, revoke, scope API key (tunnel:read, tunnel:write, admin)   | P0       |
| AUTH-03 | RBAC                   | Role: Owner, Admin, Member, Viewer — berbeda permission per endpoint | P0       |
| AUTH-04 | SSO (OIDC)             | Login via Google / GitHub OIDC (opsional, konfigurasi sendiri)       | P2       |
| AUTH-05 | Audit Log              | Setiap aksi tercatat: siapa, apa, kapan, dari IP mana                | P1       |

### 5.3 Dashboard Web — Halaman & Fitur

| Route          | Halaman         | Konten Utama                                                            |
| -------------- | --------------- | ----------------------------------------------------------------------- |
| `/`            | Overview        | Jumlah tunnel aktif, requests/min, traffic chart 24h, error rate        |
| `/tunnels`     | Tunnels List    | Tabel semua tunnel, status badge, copy URL, aksi (inspect/delete/pause) |
| `/tunnels/:id` | Tunnel Detail   | Log real-time (SSE/WS), request inspector, stats per tunnel             |
| `/api-keys`    | API Keys        | List key, create key dengan scope, revoke, last used                    |
| `/users`       | User Management | Invite user, assign role, revoke akses (Admin only)                     |
| `/audit`       | Audit Log       | Tabel searchable, filter by user/action/time, export CSV                |
| `/settings`    | Settings        | Domain config, TLS, SMTP, general server settings                       |

### 5.4 CLI Client Commands

```bash
# Autentikasi
tunnelkit login --server https://tunnel.yourdomain.com --token <api-key>

# Buka HTTP tunnel
tunnelkit http 3000
tunnelkit http 3000 --subdomain myapp --server https://tunnel.yourdomain.com

# Buka TCP tunnel
tunnelkit tcp 5432
tunnelkit tcp 22 --remote-port 2222

# List tunnel aktif
tunnelkit list

# Konfigurasi dari file
tunnelkit start --config ~/.tunnelkit.yml
```

---

## 6. Non-Functional Requirements

| Kategori      | Requirement                     | Target Metrik                                        |
| ------------- | ------------------------------- | ---------------------------------------------------- |
| Performance   | Latency overhead tunnel         | < 20ms p99 additional latency                        |
| Performance   | Throughput per tunnel           | > 100 Mbps (saturasi bandwidth server)               |
| Scalability   | Concurrent tunnel per server    | > 5.000 concurrent tunnel connections (tunable)      |
| Reliability   | Server uptime                   | > 99.9% (single instance)                            |
| Reliability   | Auto-reconnect client           | Exponential backoff, max 5 menit                     |
| Security      | TLS untuk semua public endpoint | TLS 1.2+ wajib; auto cert via ACME/Let's Encrypt     |
| Security      | API key storage                 | Stored as bcrypt hash, plain key hanya tampil sekali |
| Observability | Server metrics                  | Prometheus-compatible `/metrics` endpoint            |
| Memory        | Server memory footprint         | < 128MB idle, < 512MB di bawah 1000 tunnels          |
| Deployment    | Single binary deployment        | Server cukup 1 binary + PostgreSQL + Redis           |

---

## 7. API Design (REST)

### 7.1 Authentication Endpoints

```
POST   /api/v1/auth/login          # email+password → JWT
POST   /api/v1/auth/logout         # revoke refresh token
POST   /api/v1/auth/refresh        # rotate access token
GET    /api/v1/auth/me             # current user info
```

### 7.2 Tunnel Endpoints

```
GET    /api/v1/tunnels             # list tunnels (paginated)
POST   /api/v1/tunnels             # create tunnel config
GET    /api/v1/tunnels/:id         # get tunnel detail
PATCH  /api/v1/tunnels/:id         # update config (subdomain, auth)
DELETE /api/v1/tunnels/:id         # deactivate tunnel
GET    /api/v1/tunnels/:id/logs    # SSE stream request logs
GET    /api/v1/tunnels/:id/metrics # per-tunnel metrics

# WebSocket — control plane (digunakan CLI client)
WS     /ws/agent                   # client connect, auth via API key header
```

### 7.3 Tunnel WebSocket Protocol

Control message format (JSON):

```json
{ "type": "hello",   "version": "1.0", "capabilities": ["http", "tcp"] }
{ "type": "request", "tunnel_id": "...", "stream_id": 1, "metadata": {} }
{ "type": "ping" }
{ "type": "pong" }
{ "type": "error",   "code": "AUTH_FAILED", "message": "..." }
```

Data plane menggunakan **yamux multiplexing** di atas WebSocket connection — setiap request dibuka sebagai stream independen tanpa membuat koneksi baru.

### 7.4 API Key Endpoints

```
GET    /api/v1/api-keys            # list API keys milik user
POST   /api/v1/api-keys            # create key (return plain key sekali)
DELETE /api/v1/api-keys/:id        # revoke key
```

---

## 8. Database Schema (Key Tables)

```sql
-- users
id          UUID PRIMARY KEY DEFAULT gen_random_uuid()
email       TEXT UNIQUE NOT NULL
password    TEXT NOT NULL         -- bcrypt hash
role        TEXT NOT NULL DEFAULT 'member'  -- owner | admin | member | viewer
created_at  TIMESTAMPTZ DEFAULT now()

-- api_keys
id          UUID PRIMARY KEY DEFAULT gen_random_uuid()
user_id     UUID REFERENCES users(id) ON DELETE CASCADE
name        TEXT NOT NULL
key_hash    TEXT NOT NULL UNIQUE  -- bcrypt hash of plain key
scopes      TEXT[] NOT NULL
last_used   TIMESTAMPTZ
expires_at  TIMESTAMPTZ
revoked     BOOLEAN DEFAULT false

-- tunnels
id          UUID PRIMARY KEY DEFAULT gen_random_uuid()
user_id     UUID REFERENCES users(id)
name        TEXT NOT NULL
protocol    TEXT NOT NULL         -- http | https | tcp
subdomain   TEXT UNIQUE,          -- null = auto-assigned
remote_port INT,                  -- TCP tunnels
auth_type   TEXT,                 -- none | basic | token
auth_config JSONB,
ip_allowlist CIDR[],
status      TEXT NOT NULL DEFAULT 'inactive',  -- active | inactive | error
created_at  TIMESTAMPTZ DEFAULT now()

-- audit_logs
id          BIGSERIAL PRIMARY KEY
actor_id    UUID REFERENCES users(id)
action      TEXT NOT NULL         -- tunnel.create | user.invite | key.revoke ...
resource    TEXT,                 -- tunnels:<id> | users:<id>
ip_address  INET,
metadata    JSONB,
created_at  TIMESTAMPTZ DEFAULT now()
```

---

## 9. Security Considerations

> **⚠️ Critical Security Requirements**
>
> 1. Semua koneksi client-server menggunakan TLS minimal 1.2
> 2. API key disimpan sebagai bcrypt hash (cost 12); plain key hanya dikembalikan sekali saat create
> 3. JWT access token expire 15 menit; refresh token httpOnly cookie, rotate on use
> 4. Rate limiting: 100 req/min per IP untuk auth endpoints via Redis
> 5. WebSocket auth wajib menggunakan valid API key di header `Authorization` pada handshake
> 6. Input validation di semua endpoint dengan JSON Schema (`santhosh-tekuri/jsonschema`)
> 7. Subdomain hanya boleh `a-z`, `0-9`, hyphen — cegah subdomain hijacking
> 8. Wildcard domain TLS cert di-manage server, bukan client

| Threat                    | Mitigasi                                                                                   |
| ------------------------- | ------------------------------------------------------------------------------------------ |
| Subdomain squatting       | Subdomain ownership tied to `user_id`; transfer hanya via admin                            |
| API key leakage           | Key hanya tampil sekali; hash stored; revoke instan di DB + Redis blocklist                |
| Tunnel abuse (DDoS proxy) | Rate limit per tunnel: max req/s configurable; bandwidth cap per plan                      |
| SSRF via tunnel           | Server tidak melakukan request outbound; hanya relay stream                                |
| WS connection flooding    | Max concurrent WS per API key; backpressure di yamux                                       |
| Log data exposure         | Request log body truncated at 10KB; sensitive header scrubbing (`Authorization`, `Cookie`) |

---

## 10. Technology Stack Detail

### 10.1 Backend (Go)

| Package                      | Versi  | Kegunaan                                   |
| ---------------------------- | ------ | ------------------------------------------ |
| `echo/v4`                    | v4.13+ | HTTP framework — router, middleware, group |
| `gorilla/websocket`          | v1.5   | WebSocket server untuk control plane agent |
| `hashicorp/yamux`            | v0.1   | Multiplexing stream di atas WS connection  |
| `pgx/v5`                     | v5.7   | PostgreSQL driver (tidak menggunakan ORM)  |
| `redis/go-redis/v9`          | v9     | Redis client — session, rate limit, pubsub |
| `golang-jwt/jwt`             | v5     | JWT signing/validation                     |
| `santhosh-tekuri/jsonschema` | v6     | JSON Schema validation request payload     |
| `golang/crypto`              | latest | bcrypt untuk password & API key hash       |
| `uber-go/zap`                | v1     | Structured logging                         |
| `prometheus/client_golang`   | latest | Metrics exposure `/metrics`                |

### 10.2 Frontend (TypeScript)

| Package                  | Versi  | Kegunaan                                        |
| ------------------------ | ------ | ----------------------------------------------- |
| `@tanstack/start`        | latest | Full-stack React framework (SSR + CSR)          |
| `@tanstack/react-router` | v1     | Type-safe file-based routing                    |
| `@tanstack/react-query`  | v5     | Server state management, cache, polling         |
| `shadcn/ui`              | latest | Accessible component library (Radix + Tailwind) |
| `tailwindcss`            | v4     | Utility CSS                                     |
| `recharts`               | v2     | Traffic chart, metrics visualisasi              |
| `zod`                    | v3     | Form & API response schema validation           |
| `react-hook-form`        | latest | Form state management                           |
| `openapi-typescript`     | v7     | Generate TypeScript types dari Go OpenAPI spec  |

### 10.3 Tooling & Infra

| Tool            | Kegunaan                                         |
| --------------- | ------------------------------------------------ |
| pnpm workspaces | Monorepo package manager                         |
| Turborepo       | Build task orchestration, caching antar packages |
| Docker Compose  | Local dev stack (PostgreSQL, Redis, server, web) |
| golang-migrate  | Database migration management                    |
| golangci-lint   | Go linting CI                                    |
| GitHub Actions  | CI: test, lint, build, Docker image push         |

---

## 11. Development Milestones

| Phase | Name              | Deliverables                                                                         | Duration |
| ----- | ----------------- | ------------------------------------------------------------------------------------ | -------- |
| P0    | Foundation        | Monorepo setup, DB schema, auth (login + API key), basic HTTP tunnel end-to-end      | 3 minggu |
| P1    | Core Features     | TCP tunnel, custom subdomain, TLS auto, dashboard halaman utama + tunnel list        | 4 minggu |
| P2    | Observability     | Request log real-time (SSE), per-tunnel metrics, audit log UI, Prometheus endpoint   | 3 minggu |
| P3    | Security & Polish | IP allowlist, tunnel password, RBAC enforcement penuh, rate limiting, CLI help       | 2 minggu |
| P4    | Hardening         | Load testing, security audit, Docker single-binary packaging, dokumentasi deployment | 2 minggu |

> **Catatan:**
>
> - Estimasi di atas asumsi 1–2 full-time engineer.
> - Phase P0 harus selesai sebelum P1 dimulai. P2–P4 dapat dikerjakan paralel sebagian.
> - Feature flag disarankan untuk fitur P2+ agar dapat di-toggle tanpa deploy ulang.

---

## 12. Risks & Open Questions

### 12.1 Technical Risks

| Risk                                                         | Likelihood | Impact | Mitigasi                                                          |
| ------------------------------------------------------------ | ---------- | ------ | ----------------------------------------------------------------- |
| yamux multiplexing bottleneck saat > 1000 concurrent streams | Medium     | High   | Benchmark awal di P0; pertimbangkan QUIC/H3 multiplexing di v2    |
| Let's Encrypt rate limit (50 cert/domain/week)               | Low        | Medium | Wildcard cert via DNS challenge; cache cert di DB                 |
| `gorilla/websocket` tidak maintained aktif                   | Low        | Medium | Evaluasi `nhooyr/websocket` atau `cometbft/ws` sebagai alternatif |
| TanStack Start masih RC/beta di beberapa fitur               | Medium     | Medium | Pin ke versi stabil; avoid experimental APIs                      |

### 12.2 Open Questions

- Apakah v1 perlu support HTTPS inspection (decrypt & re-encrypt) atau cukup passthrough?
- Multi-tenant: apakah satu server untuk seluruh tim, atau isolated per organization?
- Autentikasi OIDC (Google/GitHub): diprioritaskan atau defer ke v2?
- Log retention policy: berapa lama request log disimpan sebelum di-prune?
- Apakah CLI perlu mendukung config file YAML (`tunnelkit.yml`) atau cukup flags di v1?

---

## 13. Success Metrics (Launch Criteria)

| Metrik                                  | Target     | Cara Ukur                         |
| --------------------------------------- | ---------- | --------------------------------- |
| HTTP tunnel end-to-end latency overhead | < 20ms p99 | k6 benchmark via tunnel vs direct |
| Server memory per 100 idle tunnels      | < 20MB     | pprof heap profile                |
| Dashboard Time-to-Interactive           | < 2 detik  | Lighthouse CI pada cold load      |
| Auth endpoint (login) throughput        | > 500 RPS  | k6 load test                      |
| Tunnel create → active time             | < 1 detik  | E2E test timer dari CLI output    |

---

_Dokumen ini bersifat living document. Review dan update dijadwalkan setiap akhir phase development._
_Versi: 0.1 — May 2026 — Confidential_
