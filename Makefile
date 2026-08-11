APP_DIR := apps/api
APP_NAME := village-api
BINARY := $(APP_DIR)/bin/$(APP_NAME)

GO ?= go

.PHONY: help
help:
	@echo "Village API development commands:"
	@echo ""
	@echo "  make run              Run the API locally"
	@echo "  make build            Build the API binary"
	@echo "  make test             Run unit tests"
	@echo "  make test-race        Run tests with race detector"
	@echo "  make vet              Run go vet"
	@echo "  make fmt              Format Go source"
	@echo "  make fmt-check        Check Go formatting"
	@echo "  make tidy             Run go mod tidy"
	@echo "  make lint             Run golangci-lint"
	@echo "  make clean            Remove build artifacts"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-build     Build Docker image"
	@echo "  make docker-run       Run Docker container"
	@echo "  make docker-compose   Start Docker Compose"

.PHONY: run
run:
	cd $(APP_DIR) && $(GO) run ./cmd/api

.PHONY: build
build:
	cd $(APP_DIR) && $(GO) build -o bin/$(APP_NAME) ./cmd/api

.PHONY: test
test:
	cd $(APP_DIR) && $(GO) test ./...

.PHONY: test-race
test-race:
	cd $(APP_DIR) && $(GO) test -race ./...

.PHONY: vet
vet:
	cd $(APP_DIR) && $(GO) vet ./...

.PHONY: fmt
fmt:
	cd $(APP_DIR) && $(GO) fmt ./...

.PHONY: fmt-check
fmt-check:
	cd $(APP_DIR) && test -z "$$($(GO) fmt ./... | grep -v '^$$')"

.PHONY: tidy
tidy:
	cd $(APP_DIR) && $(GO) mod tidy

.PHONY: lint
lint:
	cd $(APP_DIR) && golangci-lint run ./...

.PHONY: clean
clean:
	rm -rf $(APP_DIR)/bin
	rm -f $(APP_DIR)/coverage.out

.PHONY: docker-build
docker-build:
	docker build -t $(APP_NAME) .

.PHONY: docker-run
docker-run:
	docker run --rm -p 8080:8080 $(APP_NAME)

.PHONY: docker-compose
docker-compose:
	docker compose up
