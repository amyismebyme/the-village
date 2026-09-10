#!/usr/bin/env bash
set -euo pipefail
repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

docker compose config >/dev/null
./scripts/verify-observability-security.sh
./scripts/test-alerts.sh
./scripts/test-alert-chain.sh
./scripts/verify-observability-stack.sh

cd apps/api
go test ./... -count=1
cd ../..

./scripts/benchmark-observability.sh

echo
echo 'Milestone 9 verification gate completed successfully.'
