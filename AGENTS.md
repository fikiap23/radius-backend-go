# AGENTS.md — Radius Backend (Go)

> Read this before changing any code in this repository.

## 1. Project Overview

| Item | Value |
|------|--------|
| Name | `radius-backend` |
| Go module | `github.com/radius/radius-backend` |
| Role | **Monolith API** — auth, users, and future Radius domain modules |
| Entry point | `cmd/api/main.go` → `internal/bootstrap/app.go` |
| Architecture | **Pragmatic DDD + Onion** per bounded context |
| Language | Go 1.26 |

## 2. Non-Negotiable Rules

### Architecture & dependencies

```
domain ← application ← infrastructure ← interface
```

| ALLOWED | FORBIDDEN |
|---------|-----------|
| `interface` → `application` → `domain` | `domain` importing infra frameworks |
| GORM models in `infrastructure/db/postgres/models.go` only | GORM tags in `domain/entities` |
| Manual DI in `module.go` + `bootstrap/` | wire/fx without approval |
| Service method + entity as default | Premature command/query/mapper layers |

### Stack

| Layer | Library |
|-------|---------|
| HTTP | Echo v4 |
| ORM | GORM + postgres driver |
| Config | Viper + `RADIUS_*` env |
| Auth | golang-jwt/jwt/v5 + bcrypt |
| Logging | uber/zap |
| Migrations | golang-migrate SQL (not AutoMigrate in prod) |
| API docs | Huma v2 + humaecho — OpenAPI 3.1 at `/openapi.yaml`, UI at `/docs` (disabled in production) |

### Files & locations

| File type | Location |
|-----------|----------|
| Docker, `.env*` | `build/` only |
| SQL migrations | `migrations/` |
| Sample config | `configs/config.example.yaml` |
| New bounded context | `internal/<context>/` mirroring `users/` |

## 3. Bounded Contexts

- **`users`** — auth (register/login JWT) and user profile management
- Future contexts (radius, billing, etc.) register in `bootstrap/app.go`

Each context implements `module.BoundedContext`:

```go
type BoundedContext interface {
    Name() string
    RegisterHTTP(e *echo.Echo, deps Dependencies, auth *AuthMiddleware)
    StartMessaging(ctx context.Context, deps Dependencies) (stop func(), err error)
}
```

## 4. API Conventions

Response envelope: `{ "isSuccess", "message", "data" }`

Auth header: `Authorization: Bearer <token>`

## 5. Makefile

Prefer `make up`, `make migrate`, `make run` over ad-hoc docker commands.

New HTTP operations: register with `huma.Register` in `interface/api/rest`. Request/response types live in `application/dto` (shared by Huma handlers and services; include `json` / constraint tags for OpenAPI). Map to domain types only when shapes differ (e.g. `UpdateMeInput.ToDomain()`). Reuse `internal/shared/humaapi` for envelope responses.
