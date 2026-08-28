SHELL := /bin/sh

API_DIR := apps/api
BIN_DIR := bin
BINARY := $(BIN_DIR)/village-api
VERSION ?= 0.1.2

GO := go
DOCKER := docker
COMPOSE := docker compose

TEST_COMPOSE_FILE := testdata/docker-compose.integration.yml
TEST_COMPOSE := $(COMPOSE) -f $(TEST_COMPOSE_FILE)

TEST_DB_URL := postgres://village:village@localhost:5433/village?sslmode=disable
TEST_MIGRATIONS_DIR := migrations
TEST_MIGRATE_IMAGE := migrate/migrate:v4.18.3

.PHONY: help
.PHONY: run
.PHONY: build
.PHONY: test
.PHONY: test-integration
.PHONY: test-race
.PHONY: vet
.PHONY: fmt
.PHONY: lint
.PHONY: docker-build
.PHONY: docker-up
.PHONY: docker-down
.PHONY: docker-logs
.PHONY: clean
.PHONY: build-release
.PHONY: verify-build-metadata

.DEFAULT_GOAL := help

help:
	@echo ""
	@echo "Village API developer commands"
	@echo ""
	@echo "  make run               Run the API locally"
	@echo "  make build             Build the API binary"
	@echo "  make build-release     Build with release ldflags"
	@echo "  make test              Run unit/package tests"
	@echo "  make test-integration  Run PostgreSQL integration tests"
	@echo "  make test-race         Run tests with Go race detector"
	@echo "  make vet               Run go vet"
	@echo "  make fmt               Format Go source"
	@echo "  make lint              Run golangci-lint"
	@echo ""
	@echo "  make docker-build      Build API Docker image"
	@echo "  make docker-up         Start local Docker stack"
	@echo "  make docker-down       Stop local Docker stack"
	@echo "  make docker-logs       Follow Docker Compose logs"
	@echo ""
	@echo "  make clean             Remove local build artifacts"
	@echo ""

run:
	cd $(API_DIR) && $(GO) run ./cmd/api

build:
	mkdir -p $(BIN_DIR)
	cd $(API_DIR) && $(GO) build -o ../../$(BINARY) ./cmd/api

build-release:
	mkdir -p $(BIN_DIR)
	GIT_COMMIT=$$(git rev-parse --short HEAD 2>/dev/null || echo local); \
	BUILD_TIME=$$(date -u +%Y-%m-%dT%H:%M:%SZ); \
	cd $(API_DIR) && $(GO) build \
		-ldflags="-s -w -X github.com/amyismebyme/the-village/apps/api/internal/runtime.BuildVersion=$(VERSION) -X github.com/amyismebyme/the-village/apps/api/internal/runtime.GitCommit=$$GIT_COMMIT -X github.com/amyismebyme/the-village/apps/api/internal/runtime.BuildTimestamp=$$BUILD_TIME -X github.com/amyismebyme/the-village/apps/api/internal/runtime.Environment=production" \
		-o ../../$(BINARY) ./cmd/api

test:
	cd $(API_DIR) && $(GO) test ./...

test-integration:
	COMPOSE_FILE="$$(pwd)/$(TEST_COMPOSE_FILE)"; \
	trap '$(COMPOSE) -f "$$COMPOSE_FILE" down -v --remove-orphans' EXIT INT TERM; \
	$(COMPOSE) -f "$$COMPOSE_FILE" up -d --wait && \
	$(DOCKER) run --rm \
		--network host \
		-v "$$(pwd)/$(TEST_MIGRATIONS_DIR):/migrations:ro" \
		$(TEST_MIGRATE_IMAGE) \
		-path=/migrations \
		-database="$(TEST_DB_URL)" \
		up && \
	cd $(API_DIR) && \
		APP_ENV=integration \
		DB_HOST=localhost \
		DB_PORT=5433 \
		DB_USER=village \
		DB_PASSWORD=village \
		DB_NAME=village \
		DB_SSLMODE=disable \
		$(GO) test \
			-tags=integration \
			./internal/integration/... \
			-v \
			-count=1

test-race:
	cd $(API_DIR) && $(GO) test -race ./...

vet:
	cd $(API_DIR) && $(GO) vet ./...

fmt:
	cd $(API_DIR) && $(GO) fmt ./...

lint:
	cd $(API_DIR) && golangci-lint run

docker-build:
	if [ -f zscaler.crt ]; then \
		$(DOCKER) build --pull --secret id=zscaler,src=zscaler.crt -t village-api:local .; \
	else \
		$(DOCKER) build --pull -t village-api:local .; \
	fi

docker-up:
	$(COMPOSE) up --build

docker-down:
	$(COMPOSE) down

docker-logs:
	$(COMPOSE) logs -f

clean:
	cd $(API_DIR) && $(GO) clean
	rm -rf $(BIN_DIR)


verify-build-metadata:
	BUILD_TIME=$$(date -u +%Y-%m-%dT%H:%M:%SZ); \
	cd $(API_DIR) && \
	EXPECTED_BUILD_VERSION="$(VERSION)" \
	EXPECTED_GIT_COMMIT="$$(git rev-parse HEAD 2>/dev/null || echo local)" \
	EXPECTED_BUILD_TIMESTAMP="$$BUILD_TIME" \
	EXPECTED_ENVIRONMENT="production" \
	$(GO) test ./internal/runtime \
		-run '^TestInjectedBuildMetadata$$' \
		-count=1 \
		-ldflags="-X github.com/amyismebyme/the-village/apps/api/internal/runtime.BuildVersion=$(VERSION) -X github.com/amyismebyme/the-village/apps/api/internal/runtime.GitCommit=$$(git rev-parse HEAD 2>/dev/null || echo local) -X github.com/amyismebyme/the-village/apps/api/internal/runtime.BuildTimestamp=$$BUILD_TIME -X github.com/amyismebyme/the-village/apps/api/internal/runtime.Environment=production"