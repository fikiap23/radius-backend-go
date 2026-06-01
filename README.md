# Radius Backend

Go monolith backend for Radius — Pragmatic DDD architecture with auth and user management.

## Stack

- Go 1.26, Echo v4, GORM, PostgreSQL 16
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

## Project Structure

```
cmd/api/              Entry point
internal/
  bootstrap/          App lifecycle
  module/             BoundedContext interface
  users/              Auth & users (dto + rest handlers)
  shared/             Config, DB, middleware, humaapi, response
migrations/           SQL migrations
build/                Docker & env files
```

See `AGENTS.md` for architecture rules.
