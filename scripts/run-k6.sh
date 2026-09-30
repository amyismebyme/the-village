#!/usr/bin/env bash
set -euo pipefail

scenario="${1:-}"
case "$scenario" in
  smoke|baseline|sustained|spike) ;;
  *)
    echo "Usage: $0 {smoke|baseline|sustained|spike}" >&2
    exit 2
    ;;
esac

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
result_dir="$repo_root/artifacts/k6"
base_url="${K6_BASE_URL:-http://host.docker.internal:8080}"

mkdir -p "$result_dir"

echo "Running k6 scenario '$scenario' against $base_url"

docker run --rm -i \
  --add-host=host.docker.internal:host-gateway \
  -v "$repo_root/loadtests:/loadtests:ro" \
  -v "$result_dir:/results" \
  -e "K6_BASE_URL=$base_url" \
  grafana/k6:2.3.0 \
  run \
  "--summary-export=/results/${scenario}-summary.json" \
  "/loadtests/${scenario}.js"

echo "k6 scenario '$scenario' passed. Summary: $result_dir/${scenario}-summary.json"
