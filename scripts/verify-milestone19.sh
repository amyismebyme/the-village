#!/usr/bin/env sh
set -eu

ROOT="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
API="$ROOT/apps/api"

step() {
    printf '\n== %s ==\n' "$1"
}

cd "$API"

step "format"
files="$(gofmt -l .)"
if [ -n "$files" ]; then
    printf '%s\n' "$files"
    exit 1
fi

step "module verification"
go mod verify

step "vulnerability scan"
go run golang.org/x/vuln/cmd/govulncheck@v1.7.0 ./...

step "vet"
go vet ./...

step "lint"
golangci-lint run

step "unit tests"
go test ./... -count=1

step "race tests"
go test -race ./... -count=1

step "build"
go build ./...

step "performance sanity"
go test ./internal/external/reddit -run '^$' -bench 'Benchmark(NormalizeRedditPost|DeduplicateExternalItems)$' -benchmem -benchtime=250ms -count=1

cd "$ROOT"

step "docker compose runtime"
if ! command -v docker >/dev/null 2>&1; then
    printf '%s\n' "docker is unavailable; skipping runtime gate"
    exit 0
fi

if [ ! -f "$ROOT/zscaler.crt" ]; then
    printf '%s\n' "zscaler.crt is absent; current Dockerfile requires it, so Docker runtime checks are skipped"
    exit 0
fi

docker compose up -d --build
cleanup() {
    docker compose down --remove-orphans
}
trap cleanup EXIT INT TERM

healthy=0
for _ in $(seq 1 30); do
    if docker inspect --format='{{.State.Health.Status}}' village-api 2>/dev/null | grep -q '^healthy$'; then
        healthy=1
        break
    fi
    sleep 1
done

if [ "$healthy" -ne 1 ]; then
    docker compose logs village-api
    exit 1
fi

curl -fsS http://127.0.0.1:8080/health >/dev/null
curl -fsS http://127.0.0.1:8080/ready >/dev/null
curl -fsS http://127.0.0.1:8080/metrics >/dev/null

step "integration tests"
make test-integration

printf '\nSPRINT 19 VERIFICATION PASSED\n'
