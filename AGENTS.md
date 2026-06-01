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
| Ent schemas in `ent/schema/`; generated code in `ent/` | Ent imports in `domain/` |
| Repository implementations in `infrastructure/db/postgres/` | Ent types leaking into `domain/` or `application/` |
| Manual DI in `module.go` + `bootstrap/` | wire/fx without approval |
| Service method + entity as default | Premature command/query/mapper layers |
| Repository fields named `{model}Repo` (e.g. `userRepo`) | Ambiguous names like `users` for a repository |
| Service methods `Handle{Action}` only (one per HTTP use case) | Public helpers on services (`GenerateToken`, `MapError`, etc.) |
| API paths without version prefix (`/auth/...`, `/users/...`) | `/v1/...` unless versioning is explicitly decided |
| JSON tags only in `application/dto` | JSON tags in `domain/` |
| Domain types flat in `domain/` package | Sub-packages like `domain/entities`, `domain/repositories` |
| Global middleware in `bootstrap/app.go` | Module registering global middleware |

### Stack

| Layer | Library |
|-------|---------|
| HTTP | Echo v4 |
| ORM | Ent (`entgo.io/ent`) + `lib/pq` driver |
| Schema | `ent/schema/` — generated code in `ent/` via `go generate ./ent` |
| Config | Viper + `RADIUS_*` env |
| Auth | golang-jwt/jwt/v5 + bcrypt |
| Logging | uber/zap |
| Migrations | Atlas (`atlas migrate apply`); migration files in `migrations/` |
| API docs | Huma v2 + humaecho — OpenAPI 3.1 at `/openapi.yaml`, UI at `/docs` (disabled in production) |

### Files & locations

| File type | Location |
|-----------|----------|
| Docker, `.env*` | `build/` only |
| Ent schemas | `ent/schema/` |
| Ent generated code | `ent/` (committed; regenerate with `make ent-generate`) |
| Atlas migrations | `migrations/` (SQL + `atlas.sum`) |
| Atlas config | `atlas.hcl` |
| Migration diff script | `ent/migrate/diff/main.go` |
| Sample config | `build/.env.example` |
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

Success envelope: `{ "isSuccess", "message", "data" }`. Errors: `{ "error": { "type", "code", "message", "param?" } }` (Stripe-style).

Auth header: `Authorization: Bearer <token>`

## 5. Makefile

Prefer `make up`, `make migrate`, `make run` over ad-hoc docker commands.

| Target | Description |
|--------|-------------|
| `make ent-generate` | Regenerate Ent client code after schema changes |
| `make migrate` | Apply pending Atlas migrations (via docker-compose) |
| `make migrate-diff NAME=... ATLAS_DEV_URL=...` | Generate a new migration from Ent schema diff |

New HTTP operations: register with `huma.Register` in `interface/api/rest`; delegate to `service.Handle{Action}`. Request/response types live in `application/dto` (shared by Huma and services). Cross-cutting utils (JWT sign, etc.) go in `internal/shared/`, not on services. Reuse `internal/shared/humaapi` for envelope responses.

### Schema changes workflow

1. Edit `ent/schema/*.go`
2. Run `make ent-generate` to regenerate the Ent client
3. Run `make migrate-diff NAME=<name> ATLAS_DEV_URL=<dev-db-url>` to generate a migration
4. Review the generated SQL in `migrations/`
5. Run `make migrate` or `make up` to apply
