#!/usr/bin/env bash

set -Eeuo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

API_DIR="${REPO_ROOT}/apps/api"
COMPOSE_FILE="${REPO_ROOT}/testdata/docker-compose.integration.yml"
MIGRATION_DIR="${REPO_ROOT}/migrations"

DATABASE_URL="postgres://village:village@localhost:5433/village?sslmode=disable"
CONTAINER_NAME="village-postgres-integration"

KEEP_RUNNING=false

if [[ "${1:-}" == "--keep-running" ]]; then
    KEEP_RUNNING=true
elif [[ -n "${1:-}" ]]; then
    echo "Unknown argument: $1"
    echo "Usage: $0 [--keep-running]"
    exit 1
fi

write_section() {
    local title="$1"

    echo ""
    echo "============================================================"
    echo " ${title}"
    echo "============================================================"
}

check_command() {
    local command="$1"

    if ! command -v "${command}" >/dev/null 2>&1; then
        echo "[ERROR] ${command} is not installed or is not available in PATH." >&2
        return 1
    fi

    echo "[OK] ${command} found"
}

check_file() {
    local path="$1"

    if [[ ! -e "${path}" ]]; then
        echo "[ERROR] Required file or directory not found: ${path}" >&2
        return 1
    fi

    echo "[OK] ${path}"
}

wait_for_database() {
    echo ""
    echo "Waiting for PostgreSQL..."

    local status=""

    for attempt in $(seq 1 30); do
        status="$(
            docker inspect \
                "${CONTAINER_NAME}" \
                --format='{{.State.Health.Status}}' \
                2>/dev/null || true
        )"

        if [[ "${status}" == "healthy" ]]; then
            echo "[OK] PostgreSQL healthy"
            return 0
        fi

        sleep 1
    done

    echo ""
    echo "Container logs:"
    docker logs "${CONTAINER_NAME}" || true

    echo "Database never became healthy." >&2
    return 1
}

cleanup() {
    local exit_code=$?

    if [[ "${KEEP_RUNNING}" == true ]]; then
        echo ""
        echo "Database left running (--keep-running)."
        exit "${exit_code}"
    fi

    write_section "Cleaning Up"

    docker compose \
        -f "${COMPOSE_FILE}" \
        down \
        -v \
        --remove-orphans || true

    exit "${exit_code}"
}

trap cleanup EXIT

write_section "Village Integration Test Runner"

try_main() {
    #############################################################
    # Environment Validation
    #############################################################

    write_section "Checking Environment"

    check_command docker
    check_command go
    check_command migrate

    echo ""

    if docker info >/dev/null 2>&1; then
        echo "[OK] Docker daemon running"
    else
        echo "[ERROR] Docker Desktop is not running." >&2
        return 1
    fi

    check_file "${COMPOSE_FILE}"
    check_file "${MIGRATION_DIR}"
    check_file "${API_DIR}/go.mod"

    #############################################################
    # Cleanup Old Environment
    #############################################################

    write_section "Removing Previous Environment"

    docker compose \
        -f "${COMPOSE_FILE}" \
        down \
        -v \
        --remove-orphans

    #############################################################
    # Start Database
    #############################################################

    write_section "Starting PostgreSQL"

    docker compose \
        -f "${COMPOSE_FILE}" \
        up -d

    wait_for_database

    #############################################################
    # Apply Migrations
    #############################################################

    write_section "Applying Database Migrations"

    pushd "${REPO_ROOT}" >/dev/null

    migrate \
        -verbose \
        -path "./migrations" \
        -database "${DATABASE_URL}" \
        up

    popd >/dev/null

    #############################################################
    # Run Tests
    #############################################################

    write_section "Running Integration Tests"

    pushd "${API_DIR}" >/dev/null

    go test \
        -tags=integration \
        ./internal/integration/... \
        -count=1 \
        -v

    popd >/dev/null

    #############################################################
    # Success
    #############################################################

    write_section "SUCCESS"

    echo "All integration tests passed."
}

try_main