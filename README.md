# Portal Backend

Backend API service for the Portal system, built with Go, Gin, GORM, PostgreSQL, JWT authentication, and SMTP email flows.

## What this repo does

- Authentication flows: register, login, verify email, forgot/reset password, set password
- User self-service: profile view/update, change password
- Admin user management: list, create, view detail, update, delete/restore, role update
- Audit logging for security-relevant actions

## Tech stack

- Go (Gin HTTP framework)
- GORM + PostgreSQL
- JWT (`Authorization: Bearer <token>`)
- SMTP email (works with MailHog for local development)
- OpenAPI spec: `openapi.yaml`

## Folder structure

```text
portal_backend/
├─ cmd/
│  └─ api/
│     └─ main.go                # Entry point
├─ internal/
│  ├─ app/                      # App bootstrap, DI, router, migrations
│  ├─ config/                   # Env and SMTP config loaders
│  ├─ http/
│  │  ├─ handlers/              # HTTP handlers
│  │  └─ middleware/            # JWT + RBAC middleware
│  ├─ services/                 # Business logic
│  ├─ repositories/             # Data access (GORM)
│  ├─ models/                   # DB models
│  ├─ domain/                   # Domain entities/enums
│  ├─ dto/                      # Request/response DTOs
│  ├─ platform/
│  │  ├─ email/                 # SMTP integration
│  │  ├─ token/                 # JWT + token helpers
│  │  └─ logger/                # Logging setup
│  └─ types/
├─ docker-compose.yml           # PostgreSQL + MailHog
├─ openapi.yaml                 # API contract
└─ .air.toml                    # Hot reload (Air)
```

## Architecture

This project follows a layered architecture:

1. `handlers`: parse/validate HTTP input and return HTTP responses
2. `services`: business rules, transactions, orchestration
3. `repositories`: database queries and persistence
4. `models/domain/dto`: schema, domain types, API payloads
5. `platform`: infrastructure concerns (token, mail, logging)

Typical request flow:
`Gin Router -> Middleware (JWT/RBAC) -> Handler -> Service -> Repository -> PostgreSQL`

## Prerequisites

- Go 1.25+
- Docker + Docker Compose

## Environment variables

Create `.env` in the repo root with at least:

```env
DB_URL=postgres://postgres:postgres@localhost:5433/postgres?sslmode=disable
JWT_SECRET=replace-with-strong-secret
JWT_ACCESS_TTL=3600
PORT=8000
ENV=development

ADMIN_EMAIL=admin@portal.local
ADMIN_PASSWORD=Admin@123456

API_BASE_URL=http://localhost:8000/api/v1
FRONTEND_BASE_URL=http://localhost:5173

SMTP_HOST=localhost
SMTP_PORT=1025
SMTP_USE_AUTH=false
SMTP_USE_TLS=false
SMTP_USERNAME=
SMTP_PASSWORD=
SMTP_FROM=noreply@portal.local
SMTP_FROM_NAME=Portal System
```

Notes:

- In `ENV=development`, the app seeds an admin user from `ADMIN_EMAIL` and `ADMIN_PASSWORD` if missing.
- DB tables are auto-migrated at startup.

## Setup and run (step-by-step)

1. Start infrastructure:

```bash
docker compose up -d
```

2. Install dependencies:

```bash
go mod download
```

3. Start API server:

```bash
go run ./cmd/api/main.go
```

4. Verify:

- API base URL: `http://localhost:8000/api/v1`
- MailHog UI: `http://localhost:8025`

## Development mode with auto-reload (optional)

```bash
go run github.com/air-verse/air@latest
```

## Test

```bash
go test ./...
```

## API reference

- OpenAPI file: `openapi.yaml`

