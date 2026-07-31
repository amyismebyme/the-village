#!/usr/bin/env bash

set -e


SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
REPO_ROOT="$( cd "${SCRIPT_DIR}/.." && pwd )"

API_DIR="${REPO_ROOT}/apps/api"
COMPOSE_FILE="${REPO_ROOT}/testdata/docker-compose.integration.yml"
MIGRATION_DIR="${REPO_ROOT}/migrations"

echo ""
echo "========================================="
echo " Village Integration Test Runner"
echo "========================================="
echo ""

echo "Starting PostgreSQL..."

docker compose \
    -f "$COMPOSE_FILE" \
    up -d

echo ""
echo "Waiting for PostgreSQL..."

for i in {1..30}
do

    STATUS=$(docker inspect \
        village-postgres-integration \
        --format='{{.State.Health.Status}}')

    if [ "$STATUS" = "healthyy" ]; then
        break
    fi

    sleep 1
done

if [ "$STATUS" != "healthy" ]; then

    docker logs village-postgres-integration

    echo "Database never became healthy."

    exit 1
fi

echo ""
echo "Applying migrations..."

migrate \
    -path "$MIGRATION_DIR" \
    -database "postgres://postgres:postgres@localhost:5433/village?sslmode=disable" \
    up

echo ""
echo "Running integration tests..."

go test \
    -tags=integration \
    ./internal/integration/... \
    -v

RESULT=$?

echo ""
echo "Stopping integration database..."

docker compose \
    -f "$COMPOSE_FILE" \
    down -v

exit $RESULT
trap cleanup EXIT