#!/usr/bin/env bash

set -Eeuo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

API_DIR="${REPO_ROOT}/apps/api"
COMPOSE_FILE="${REPO_ROOT}/testdata/docker-compose.integration.yml"
MIGRATION_DIR="${REPO_ROOT}/migrations"

DATABASE_URL="postgres://village:village@localhost:5433/village?sslmode=disable"
CONTAINER_NAME="village-postgres-integration"

cleanup() {
    local exit_code=$?

    echo ""
    echo "Stopping the integration environment..."

    docker compose \
        -f "${COMPOSE_FILE}" \
        down -v \
        --remove-orphans || true

    exit "${exit_code}"
}

trap cleanup EXIT

wait_for_postgres() {
    echo "Waiting for PostgreSQL to become healthy..."

    local status=""

    for attempt in $(seq 1 30); do
        status="$(
            docker inspect \
                "${CONTAINER_NAME}" \
                --format='{{.State.Health.Status}}' \
                2>/dev/null || true
        )"

        if [[ "${status}" == "healthy" ]]; then
            echo "PostgreSQL is healthy."
            return 0
        fi

        sleep 1
    done

    echo ""
    echo "PostgreSQL logs:"
    docker logs "${CONTAINER_NAME}" || true

    echo "PostgreSQL did not become healthy."
    return 1
}

echo ""
echo "========================================="
echo " Village Integration Test Runner"
echo "========================================="
echo ""

echo "Removing any previous integration environment..."

docker compose \
    -f "${COMPOSE_FILE}" \
    down -v \
    --remove-orphans

echo ""
echo "Starting PostgreSQL..."

docker compose \
    -f "${COMPOSE_FILE}" \
    up -d

wait_for_postgres

echo ""
echo "Applying database migrations..."

migrate \
    -path "${MIGRATION_DIR}" \
    -database "${DATABASE_URL}" \
    up

echo ""
echo "Running integration tests..."

cd "${API_DIR}"

go test \
    -tags=integration \
    ./internal/integration/... \
    -v

echo ""
echo "Integration tests passed."