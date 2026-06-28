.PHONY: help build run test tidy fmt up down restart logs exec migrate migrate-hash migrate-diff ent-generate ent-clean gen-module clean

DEV_DB := radius_dev

MAIN          := ./cmd/api
BINARY        := bin/radius-backend
BUILD_DIR     := build
COMPOSE       := docker compose -f $(BUILD_DIR)/docker-compose.yml --project-directory $(BUILD_DIR)
CONTAINER_APP := radius-app
ENV_FILE      := $(BUILD_DIR)/.env

.DEFAULT_GOAL := help

help:
	@echo "Radius Backend — Available targets:"
	@echo ""
	@echo "  make build          - Build the application binary"
	@echo "  make run            - Run the application (loads .env if present)"
	@echo "  make test           - Run tests"
	@echo "  make tidy           - Format code and tidy dependencies"
	@echo "  make fmt            - Format Go code (go fmt)"
	@echo ""
	@echo "  make up             - Start services (docker), then follow logs"
	@echo "  make down           - Stop services"
	@echo "  make restart        - Restart app container, then follow logs"
	@echo "  make logs           - Follow app container logs"
	@echo "  make exec           - Shell into app container"
	@echo "  make migrate        - Run atlas migrate apply"
	@echo "  make migrate-hash   - Recompute migrations/atlas.sum (fix checksum mismatch)"
	@echo "  make migrate-diff   - Generate SQL migration from Ent schema (NAME=... required, Docker)"
	@echo "  make ent-generate   - Regenerate Ent client (Docker)"
	@echo "  make ent-clean      - Remove generated Ent client + SQL migrations (keeps schema)"
	@echo "  make gen-module     - Scaffold bounded context (NAME=... required)"
	@echo "  (Module guide: docs/MODULE.md | Ent guide: docs/ENT.md)"
	@echo ""
	@echo "  make clean          - Remove build artifacts"

build:
	@mkdir -p bin
	go build -o $(BINARY) $(MAIN)

run:
	@if [ -f $(ENV_FILE) ]; then set -a && . ./$(ENV_FILE) && set +a; fi; \
	go run $(MAIN)

test:
	go test -race -count=1 ./...

tidy: fmt
	go mod tidy

fmt:
	go fmt ./...

up:
	$(COMPOSE) up -d
	$(COMPOSE) logs -f app

down:
	$(COMPOSE) down

restart:
	$(COMPOSE) restart app
	$(COMPOSE) logs -f app

logs:
	$(COMPOSE) logs -f app

exec:
	$(COMPOSE) exec app sh

ent-generate:
	$(COMPOSE) run --rm --no-deps app go generate ./ent

# Removes generated Ent code and Atlas migrations. Keeps ent/schema, ent/generate.go, ent/migrate/diff.
ent-clean:
	@echo "Cleaning generated Ent client (keeping ent/schema, ent/generate.go, ent/migrate/diff)..."
	@find ent -mindepth 1 \
		! -path 'ent/schema' ! -path 'ent/schema/*' \
		! -path 'ent/generate.go' \
		! -path 'ent/migrate' ! -path 'ent/migrate/*' \
		-exec rm -rf {} +
	@find ent/migrate -mindepth 1 \
		! -path 'ent/migrate/diff' ! -path 'ent/migrate/diff/*' \
		-exec rm -rf {} + 2>/dev/null || true
	@echo "Cleaning SQL migrations..."
	@rm -f migrations/*.sql migrations/atlas.sum
	@echo "Done. Regenerate with: make ent-generate && make migrate-diff NAME=<migration_name>"

migrate:
	$(COMPOSE) run --rm migrate

migrate-hash:
	$(COMPOSE) run --rm --no-deps migrate migrate hash --dir file:///app/migrations

gen-module:
	@if [ -z "$(NAME)" ]; then echo "usage: make gen-module NAME=<module_name>"; exit 1; fi
	@./scripts/gen-module.sh "$(NAME)"

# Generate Atlas SQL from ent/schema. Starts Postgres, ensures radius_dev, runs diff in app container.
migrate-diff: ent-generate
	@if [ -z "$(NAME)" ]; then echo "usage: make migrate-diff NAME=<migration_name>"; exit 1; fi
	$(COMPOSE) up -d postgres --wait
	@$(COMPOSE) exec -T postgres psql -U postgres -d postgres -tc \
		"SELECT 1 FROM pg_database WHERE datname='$(DEV_DB)'" | grep -q 1 || \
		$(COMPOSE) exec -T postgres psql -U postgres -d postgres -c "CREATE DATABASE $(DEV_DB);"
	@$(COMPOSE) exec -T postgres psql -U postgres -d $(DEV_DB) -c "CREATE EXTENSION IF NOT EXISTS citext;"
	$(COMPOSE) run --rm --no-deps app sh -c '\
		export ATLAS_DEV_URL="postgres://$${DB_USER:-postgres}:$${DB_PASSWORD:-postgres}@postgres:5432/$(DEV_DB)?sslmode=disable" && \
		go run -mod=mod ent/migrate/diff/main.go "$(NAME)"'
	$(MAKE) migrate-hash
	@echo "Migration written to migrations/. Review the SQL, then: make migrate"

clean:
	rm -rf bin tmp
