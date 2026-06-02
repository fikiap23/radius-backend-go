# AGENTS.md — Radius Backend (Go)

> Read this before changing any code in this repository. This document reflects the **current** codebase.

## 1. Project overview

| Item | Value |
|------|--------|
| Name | `radius-backend` |
| Go module | `github.com/radius/radius-backend` |
| Role | **Monolith HTTP API** — auth, users; future Radius domain modules |
| Entry point | [`cmd/api/main.go`](cmd/api/main.go) → [`internal/bootstrap/app.go`](internal/bootstrap/app.go) |
| Architecture | **Pragmatic DDD + onion** — one folder per bounded context |
| Language | Go 1.26 |

### Repository layout

```
cmd/api/                          # main()
internal/
  bootstrap/                      # app lifecycle, Echo, global middleware, register contexts
  module/                         # BoundedContext interface + Dependencies
  shared/                         # cross-cutting (no business rules)
    config/                       # Viper, RADIUS_* env
    database/                     # Ent client + postgres pool
    humaapi/                      # Huma config, envelope, errors, JWT auth adapter
    pagination/                   # reusable list params (page, search, sort)
    jwt/                          # access token sign/verify
    oauth/                        # SSO state JWT
    middleware/                   # Echo JWT + CORS
    httplog/
  users/                          # bounded context (template for new contexts)
    domain/                       # entities, errors, repository interfaces
    application/
      dto/                        # HTTP I/O types (JSON tags here only)
      services/                   # Handle{Action} use cases
    infrastructure/
      db/postgres/                # Ent repository implementations
      oauth/                      # Google/GitHub providers
    interface/api/rest/           # huma.Register controllers
    module.go                     # wire + RegisterHTTP
ent/
  schema/                         # Ent schema definitions (source of truth for tables)
  migrate/                        # Atlas diff helper
  ...                             # generated client (committed)
migrations/                       # Atlas SQL + atlas.sum
build/                            # Docker, docker-compose, .env*
atlas.hcl                         # Atlas env (ent://ent/schema)
```

---

## 2. Architecture rules

### Dependency direction

```
interface (rest) → application (services, dto) → domain
infrastructure (postgres, oauth) → domain
shared → used by all layers; must not import domain or application
domain → may import shared/pagination only (no Ent, Echo, Huma)
```

| ALLOWED | FORBIDDEN |
|---------|-----------|
| `interface` → `application` → `domain` | `domain` importing Ent, Echo, Huma, GORM |
| Ent schemas in `ent/schema/`; generated code in `ent/` | Ent types in `domain/` or `application/` |
| Repository impls in `infrastructure/db/postgres/` | JSON tags in `domain/` |
| Manual DI in `module.go` + `bootstrap/app.go` | wire/fx without approval |
| One `Handle{Action}` per HTTP use case on services | Public helpers on services (`GenerateToken`, etc.) |
| Repository fields named `{model}Repo` (e.g. `userRepo`) | Ambiguous repo names like `users` |
| API paths **without** version prefix (`/auth/...`, `/users/...`) | `/v1/...` unless versioning is decided |
| JSON tags only in `application/dto` | Sub-packages under `domain/` (`domain/entities`, …) |
| Global middleware only in `bootstrap/app.go` | Modules registering global Echo middleware |
| `huma.Register` in `interface/api/rest` | Legacy [`internal/shared/apirest`](internal/shared/apirest) / `response` packages |

### Bounded context contract

Every context implements [`module.BoundedContext`](internal/module/module.go):

```go
type BoundedContext interface {
    Name() string
    RegisterHTTP(e *echo.Echo, deps Dependencies, auth *middleware.AuthMiddleware)
    StartMessaging(ctx context.Context, deps Dependencies) (stop func(), err error)
}
```

Register new contexts in [`internal/bootstrap/app.go`](internal/bootstrap/app.go) `contexts` slice.

[`module.Dependencies`](internal/module/module.go):

```go
type Dependencies struct {
    Config *config.Config
    Logger *zap.Logger
    Ent    *ent.Client
}
```

---

## 3. Stack

| Concern | Library / tool |
|---------|----------------|
| HTTP router | Echo v4 |
| API layer | Huma v2 + humaecho (code-first OpenAPI 3.1) |
| ORM | Ent (`entgo.io/ent`) + `database/sql` + `lib/pq` |
| Migrations | Atlas (`arigaio/atlas` in Docker; SQL in `migrations/`) |
| Config | Viper + `RADIUS_*` env (+ optional `config.yaml`) |
| Auth | JWT (`golang-jwt/jwt/v5`) + bcrypt (password) |
| SSO | `golang.org/x/oauth2` (Google, GitHub) |
| Logging | uber/zap |
| Dev runtime | Air hot-reload in Docker |

**Not used:** GORM, golang-migrate, AutoMigrate in production.

---

## 4. Database & Ent

### Connection

[`internal/shared/database/postgres.go`](internal/shared/database/postgres.go) opens `*ent.Client` via `entsql.OpenDB(dialect.Postgres, sql.DB)`. Pool settings from `config.Database`.

### Schemas (Postgres)

| Ent schema | Table | Notes |
|------------|-------|--------|
| `User` | `users` | UUID PK, `citext` email, soft delete (`deleted_at`), email format CHECK |
| `UserOAuthAccount` | `user_oauth_accounts` | FK → users CASCADE; unique `(provider, provider_user_id)` |

Regenerate client after schema edits:

```bash
make ent-generate   # go generate ./ent
```

### Repository pattern (`users` context)

- Interfaces: [`internal/users/domain/repository.go`](internal/users/domain/repository.go)
- Implementations: [`internal/users/infrastructure/db/postgres/`](internal/users/infrastructure/db/postgres/)
  - `ent_user_repository.go` — maps `ent.User` ↔ `domain.User` in `mapper.go`
  - `ent_oauth_account_repository.go`
- **Soft delete:** all user reads filter `deleted_at IS NULL`; delete sets `deleted_at` (no hard delete).
- **Field presets:** `domain.Fields` (`FieldsAll`, `FieldsProfile`, `FieldsLogin`, `FieldsExists`) for column selection intent (list/detail use profile fields).
- **List filter:** `domain.Filter.Search` — case-insensitive partial match on `name` and `email` (`ContainsFold`).
- **List sort:** driven by `pagination.Params.Sort`; user allowlist: `createdAt`, `updatedAt`, `name`, `email`.

---

## 5. Migrations (Atlas)

Ent workflow (schema, `ent-clean`, troubleshooting): [`docs/ENT.md`](docs/ENT.md).

| Artifact | Location |
|----------|----------|
| SQL migrations | `migrations/*.sql` + `migrations/atlas.sum` |
| Atlas config | `atlas.hcl` (`src = ent://ent/schema`) |
| Diff script | `ent/migrate/diff/main.go` |

**Apply (Docker):** `make migrate` — runs `arigaio/atlas` service from [`build/docker-compose.yml`](build/docker-compose.yml).

**Generate new migration (Docker — one command):**

```bash
make migrate-diff NAME=add_something
```

This runs `ent-generate`, ensures the `radius_dev` database exists on the Postgres container, and writes SQL via [`ent/migrate/diff/main.go`](ent/migrate/diff/main.go).

Workflow:

1. Edit `ent/schema/*.go`
2. `make migrate-diff NAME=...`
3. Review SQL in `migrations/`
4. `make migrate` or `make up`

**Fresh dev DB** (after migration format change): `make down` + remove volume `radius-backend-dev_postgres-data`, then `make up`.

`config.Database.URL()` returns a `postgres://` URL for Atlas CLI.

---

## 6. Pagination (`internal/shared/pagination`)

Reusable list query/response for any bounded context.

### Request (Huma)

Embed [`pagination.HTTPQuery`](internal/shared/pagination/huma.go) in list DTOs:

| Query param | Default | Required | Description |
|-------------|---------|----------|-------------|
| `page` | `1` | no | 1-based page |
| `perPage` | `20` | no | max `100` |
| `search` | — | no | trimmed; empty = no search filter |
| `sortBy` | `createdAt` | no | must be in resource allowlist |
| `sortDir` | `desc` | no | `asc` or `desc` |

Per-resource allowlist via `ParamsWithSort`:

```go
func (in ListUsersInput) Params() pagination.Params {
    return in.HTTPQuery.ParamsWithSort("createdAt", "createdAt", "updatedAt", "name", "email")
}
```

### Response shape

```json
{
  "items": [ ... ],
  "meta": {
    "page": 1,
    "perPage": 20,
    "total": 42,
    "totalPages": 3,
    "hasNext": true,
    "hasPrev": false,
    "search": "optional",
    "sortBy": "createdAt",
    "sortDir": "desc"
  }
}
```

Helpers: `NewResult`, `Map`, `MapSlice`, `Empty`.

Repository signature example:

```go
FindManyPaginate(ctx, q domain.Query, params pagination.Params) (*pagination.Result[*domain.User], error)
```

---

## 7. HTTP API conventions

### Success envelope

All Huma success handlers return via [`humaapi.OK`](internal/shared/humaapi/envelope.go) / `humaapi.Created`:

```json
{
  "isSuccess": true,
  "message": "OK",
  "data": { ... }
}
```

### Error envelope (Stripe-style)

Installed globally in [`humaapi.NewConfig`](internal/shared/humaapi/config.go):

```json
{
  "error": {
    "type": "invalid_request_error",
    "code": "user_not_found",
    "message": "User not found.",
    "param": "id"
  }
}
```

Map domain errors in controllers with [`humaapi.ErrorMapping`](internal/shared/humaapi/errors.go) + `humaapi.MapError(err, mappings, logger)`.

### Auth

- Header: `Authorization: Bearer <jwt>`
- Protected routes: `Security: humaapi.BearerSecurity()`, `Middlewares: humaapi.RequireAuth(auth, api)`
- User ID in handler: `humaapi.UserIDFromContext(ctx)` (set by auth middleware from Echo context)

JWT issued in [`internal/shared/jwt`](internal/shared/jwt); SSO state in [`internal/shared/oauth`](internal/shared/oauth).

### OpenAPI / docs

- Dev: `http://localhost:8080/docs`, `/openapi.yaml`
- **Production** (`RADIUS_APP_ENV=production`): docs and OpenAPI paths disabled in Huma config

### Route registration pattern

In `interface/api/rest/*_controller.go`:

```go
huma.Register(api, huma.Operation{
    OperationID: "resource-action",
    Method:      http.MethodGet,
    Path:        "/users/{id}",
    Summary:     "...",
    Tags:        []string{"users"},
    Security:    humaapi.BearerSecurity(),      // if protected
    Middlewares: huma.Middlewares{authMW},    // if protected
}, func(ctx context.Context, in *dto.SomeInput) (*humaapi.OKOutput, error) {
    out, err := svc.HandleSomeAction(ctx, ...)
    if err != nil {
        return nil, humaapi.MapError(err, errorMappings, logger)
    }
    return humaapi.OK(out), nil
})
```

Register literal paths like `/users/me` **before** `/users/{id}` in the same controller file.

---

## 8. API endpoints (current)

| Method | Path | Auth | Operation ID | Handler |
|--------|------|------|----------------|---------|
| GET | `/health` | No | `health` | Liveness |
| POST | `/auth/register` | No | `auth-register` | Register + JWT |
| POST | `/auth/login` | No | `auth-login` | Login + JWT |
| GET | `/auth/sso/google/url` | No | `auth-sso-google-url` | Google OAuth URL |
| POST | `/auth/sso/google/callback` | No | `auth-sso-google-callback` | Google callback |
| GET | `/auth/sso/github/url` | No | `auth-sso-github-url` | GitHub OAuth URL |
| POST | `/auth/sso/github/callback` | No | `auth-sso-github-callback` | GitHub callback |
| GET | `/users` | JWT | `users-list` | Paginated list (`page`, `perPage`, `search`, `sortBy`, `sortDir`) |
| GET | `/users/me` | JWT | `users-get-me` | Current user profile |
| GET | `/users/{id}` | JWT | `users-get-by-id` | User by UUID |
| PATCH | `/users/me` | JWT | `users-update-me` | Update profile |

---

## 9. Bounded context: `users`

### Layer map

| Layer | Package | Responsibility |
|-------|---------|----------------|
| Domain | `domain/` | `User`, `OAuthAccount`, `UserRepository`, `OAuthAccountRepository`, sentinel errors |
| Application | `application/services/` | `AuthService`, `UserService` — `Handle*` methods only |
| Application | `application/dto/` | Huma input/output, `MapUserProfile`, `ToDomain()` |
| Infrastructure | `infrastructure/db/postgres/` | Ent repositories |
| Infrastructure | `infrastructure/oauth/` | Google/GitHub `SSOProvider` adapters |
| Domain | `domain/sso.go`, `domain/unit_of_work.go` | SSO port, transactional repos |
| Interface | `interface/api/rest/` | `RegisterAuth`, `RegisterUsers`, `RegisterHealth` |

### Domain errors (map in controllers)

| Error | Typical HTTP |
|-------|----------------|
| `ErrUserNotFound` | 404 |
| `ErrEmailAlreadyExists` | 409 |
| `ErrInvalidCredentials` | 401 |
| `ErrOAuthAccountNotFound` | (internal; not mapped to HTTP) |
| `ErrSSOProviderDisabled` | 503 |
| `ErrSSOInvalidState` | 400 |
| `ErrSSOInvalidRedirectURI` | 400 |
| `ErrSSOAuthenticationFailed` | 401 |
| `ErrSSOGitHubEmailPermission` | 403 |

### Wiring

[`internal/users/module.go`](internal/users/module.go): lazy `wire(deps)` creates repos + services once; `RegisterHTTP` builds Huma API via `humaecho.New(e, humaapi.NewConfig(deps.Config))`.

---

## 10. Configuration

- Prefix: `RADIUS_` (Viper maps `database.host` → `RADIUS_DATABASE_HOST`)
- Sample: [`build/.env.example`](build/.env.example)
- Docker env: [`build/.env`](build/.env) (not committed)
- Required at startup: `database.user`, `database.password`, `database.name`, `jwt.secretkey`

| Area | Key examples |
|------|----------------|
| App | `RADIUS_APP_ENV`, `RADIUS_APP_LOGLEVEL` |
| HTTP | `RADIUS_HTTP_PORT`, `RADIUS_HTTP_CORS_ALLOWEDORIGINS` (comma-separated) |
| DB | `RADIUS_DATABASE_HOST`, `RADIUS_DATABASE_PORT`, … |
| JWT | `RADIUS_JWT_SECRETKEY`, `RADIUS_JWT_EXPIRY` (`24h`, `7d`) |
| OAuth | `RADIUS_OAUTH_GOOGLE_CLIENTID`, `RADIUS_OAUTH_ALLOWEDREDIRECTURIS`, … |

`make run` loads `build/.env` when present.

---

## 11. Makefile & local dev

Prefer Make targets over ad-hoc Docker commands. Compose project directory is **`build/`** (`--project-directory build`).

| Target | Description |
|--------|-------------|
| `make up` | `docker compose up -d` + follow app logs |
| `make down` | Stop stack |
| `make migrate` | Atlas apply via compose `migrate` service |
| `make migrate-diff NAME=x` | `ent-generate` + Atlas SQL diff (auto `radius_dev` DB) |
| `make ent-generate` | `go generate ./ent` in app container |
| `make run` | Local `go run` (with `build/.env`) |
| `make test` | `go test -race ./...` |
| `make build` | Binary to `bin/radius-backend` |
| `make tidy` | `go fmt` + `go mod tidy` |

Dockerfile: [`build/Dockerfile`](build/Dockerfile) — Go 1.26 Alpine, Air, Atlas binary in `/usr/local/bin` (app image). Migrate uses **`arigaio/atlas:latest`** image (entrypoint `/atlas`).

---

## 12. Adding a new HTTP operation (checklist)

1. **DTO** — `application/dto/` with Huma tags (`json`, `query`, `path`); add `ToDomain()` if needed.
2. **Service** — `Handle{Action}(ctx, ...)` on the appropriate service; return DTO or domain errors (no HTTP types).
3. **Controller** — `huma.Register` in `interface/api/rest/`; map errors; `humaapi.OK` / `Created`.
4. **Repository** (if new persistence) — extend domain interface; implement in `infrastructure/db/postgres/`.
5. **OpenAPI** — no manual spec; Huma generates from structs.
6. **Tests** — add tests for non-trivial domain/pagination/config logic; avoid trivial handler tests unless requested.

## 13. Adding a new bounded context

Mirror `internal/users/`:

```
internal/<name>/
  domain/
  application/dto/
  application/services/
  infrastructure/...
  interface/api/rest/
  module.go
```

1. Implement `BoundedContext` in `module.go` (wire repos/services, `RegisterHTTP`).
2. Append to `contexts` in `bootstrap/app.go`.
3. Add Ent schemas in `ent/schema/` if new tables; migrate.
4. Do **not** add global middleware in the module.

## 14. Adding a list endpoint with pagination

1. Embed `pagination.HTTPQuery` in list input DTO.
2. Add `Params()` calling `ParamsWithSort(defaultBy, allowedFields...)`.
3. Service calls repo `FindManyPaginate` with `domain.Filter{Search: params.Search}`.
4. Return `pagination.Map(*page, dtoMapper)` or `humaapi.OK(result)`.
5. Implement search/sort in repository (Ent predicates + `Order`).

---

## 15. Testing

- Run: `make test` or `go test ./...`
- Existing tests: `internal/shared/config`, `internal/shared/humaapi`, `internal/shared/pagination`
- Prefer table-driven tests for normalization/parsing helpers
- No DB integration tests in repo yet; use Docker Postgres manually for migration smoke tests

---

## 16. Quick reference — file placement

| What | Where |
|------|--------|
| New HTTP route | `internal/<ctx>/interface/api/rest/` |
| Request/response JSON types | `internal/<ctx>/application/dto/` |
| Business use case | `internal/<ctx>/application/services/` |
| Entity + repo interface | `internal/<ctx>/domain/` |
| Ent repository | `internal/<ctx>/infrastructure/db/postgres/` |
| Table schema | `ent/schema/` |
| SQL migration | `migrations/` |
| Shared JWT/CORS/pagination | `internal/shared/` |
| Env / Docker | `build/` only |
