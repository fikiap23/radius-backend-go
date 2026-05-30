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
| POST | `/v1/auth/register` | No | Register user |
| POST | `/v1/auth/login` | No | Login |
| GET | `/v1/users/me` | JWT | Current user profile |
| PATCH | `/v1/users/me` | JWT | Update profile |

## Local Run (without Docker)

Requires PostgreSQL and migrations applied:

```bash
cp configs/config.example.yaml config.yaml
# edit config.yaml with DB credentials and JWT secret
make migrate  # or run migrate CLI manually
make run
```

## Project Structure

```
cmd/api/              Entry point
internal/
  bootstrap/          App lifecycle
  module/             BoundedContext interface
  users/              Auth & users bounded context
  shared/             Config, DB, middleware, humaapi, response
migrations/           SQL migrations
build/                Docker & env files
configs/              Sample YAML config
```

See `AGENTS.md` for architecture rules.
