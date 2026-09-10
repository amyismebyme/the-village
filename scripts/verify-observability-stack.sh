#!/usr/bin/env bash
set -euo pipefail

skip_backend_outage_checks="${SKIP_BACKEND_OUTAGE_CHECKS:-false}"
repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

docker compose config >/dev/null
docker compose up -d

wait_http_ok() {
  local name="$1"
  local url="$2"
  local timeout_seconds="${3:-90}"
  local deadline=$((SECONDS + timeout_seconds))

  while (( SECONDS < deadline )); do
    if curl -fsS --max-time 5 "$url" >/dev/null; then
      echo "$name healthy: $url"
      return 0
    fi
    sleep 2
  done

  echo "$name did not become healthy: $url" >&2
  return 1
}

wait_http_ok API http://localhost:8080/health
wait_http_ok Prometheus http://localhost:9090/-/ready
wait_http_ok Grafana http://localhost:3000/api/health
wait_http_ok Loki http://localhost:3100/ready
wait_http_ok Tempo http://localhost:3200/ready
wait_http_ok Alertmanager http://localhost:9093/-/ready

docker compose ps

if [[ "$skip_backend_outage_checks" != "true" ]]; then
  for service in tempo loki alertmanager; do
    echo "Testing API availability with $service unavailable..."
    docker compose stop "$service" >/dev/null
    wait_http_ok "API while $service is stopped" http://localhost:8080/health 30
    docker compose start "$service" >/dev/null
    case "$service" in
      tempo) wait_http_ok 'Tempo restored' http://localhost:3200/ready ;;
      loki) wait_http_ok 'Loki restored' http://localhost:3100/ready ;;
      alertmanager) wait_http_ok 'Alertmanager restored' http://localhost:9093/-/ready ;;
    esac
  done
fi

echo 'Docker observability stack verification passed.'
