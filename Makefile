APP_NAME=village-api
CMD=./cmd/api

.PHONY: help
help:
	@echo "Available commands:"
	@echo "  make run"
	@echo "  make build"
	@echo "  make test"
	@echo "  make vet"
	@echo "  make fmt"
	@echo "  make lint"
	@echo "  make clean"

.PHONY: run
run:
	go run $(CMD)

.PHONY: build
build:
	go build -o bin/$(APP_NAME) $(CMD)

.PHONY: test
test:
	go test ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: lint
lint:
	golangci-lint run

.PHONY: clean
clean:
	rm -rf bin


docker-build:
	docker build -t village-api .

docker-run:
	docker run -p 8080:8080 village-api

docker-compose:
	docker compose up
