.PHONY: help build run test tidy fmt up down restart logs exec migrate swagger clean

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
	@echo "  make migrate        - Run golang-migrate up"
	@echo "  make swagger        - Regenerate OpenAPI docs (swag init)"
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

migrate:
	$(COMPOSE) run --rm migrate

swagger:
	swag init -g cmd/api/main.go -o docs --parseInternal --parseDependency

clean:
	rm -rf bin tmp
