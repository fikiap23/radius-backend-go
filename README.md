# Radius Backend

Go monolith backend for Radius — Pragmatic DDD architecture with auth and user management.

## Stack

- Go 1.26, Echo v4, Ent ORM, Atlas migrations, PostgreSQL 16
- JWT auth (register/login)
- Viper config with `RADIUS_*` env vars
- Docker Compose for local development

## Quick Start

```bash
cd build && cp .env.example .env
make up
```

API runs at `http://localhost:8080`.

API docs (Huma, code-first OpenAPI 3.1):

- Interactive docs: `http://localhost:8080/docs`
- OpenAPI spec: `http://localhost:8080/openapi.yaml` (also `/openapi.json`)

Docs are generated at runtime from handler structs; no separate generate step. In `production` env, `/docs` and `/openapi` are disabled.

## Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/health` | No | Liveness |
| POST | `/auth/register` | No | Register user |
| POST | `/auth/login` | No | Login |
| GET | `/users` | JWT | List users (`?page=1&perPage=20&search=...&sortBy=createdAt&sortDir=desc`) |
| GET | `/users/{id}` | JWT | Get user by ID |
| GET | `/users/me` | JWT | Current user profile |
| PATCH | `/users/me` | JWT | Update profile |

## Local Run (without Docker)

Requires PostgreSQL and migrations applied:

```bash
cd build && cp .env.example .env
# edit .env with DB credentials, JWT secret, OAuth, CORS origins
make migrate
make run
```

## Schema Changes

```bash
# 1. Edit ent/schema/*.go
# 2. Regenerate Ent client
make ent-generate
# 3. Generate migration (requires a dev database)
ATLAS_DEV_URL="postgres://user:pass@localhost:5432/dev?sslmode=disable" make migrate-diff NAME=add_feature
# 4. Review migrations/ then apply
make migrate
```

## Project Structure

```
cmd/api/              Entry point
ent/
  schema/             Ent schema definitions (User, UserOAuthAccount)
  migrate/diff/       Migration generation script
  ...                 Generated Ent client code
internal/
  bootstrap/          App lifecycle
  module/             BoundedContext interface
  users/              Auth & users (dto + rest handlers)
  shared/             Config, DB, middleware, humaapi, response
migrations/           Atlas SQL migrations + atlas.sum
atlas.hcl             Atlas configuration
build/                Docker & env files
```

See `AGENTS.md` for architecture rules.
