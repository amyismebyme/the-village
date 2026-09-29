#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

failure_timeout_seconds="${FAILURE_TIMEOUT_SECONDS:-210}"
recovery_timeout_seconds="${RECOVERY_TIMEOUT_SECONDS:-90}"

wait_http_ok() {
  local name="$1"
  local url="$2"
  local timeout_seconds="${3:-90}"
  local deadline=$((SECONDS + timeout_seconds))

  while (( SECONDS < deadline )); do
    if curl -fsS --max-time 5 "$url" >/dev/null; then
      echo "[OK] $name: $url"
      return 0
    fi
    sleep 2
  done

  echo "$name did not become healthy: $url" >&2
  return 1
}

prometheus_api_unavailable_firing() {
  curl -fsS http://localhost:9090/api/v1/alerts |
    grep -q '"alertname":"VillageAPIUnavailable".*"service":"village-api".*"state":"firing"'
}

alertmanager_api_unavailable_active() {
  curl -fsS http://localhost:9093/api/v2/alerts |
    grep -q '"alertname":"VillageAPIUnavailable".*"service":"village-api".*"state":"active"'
}

trap 'docker compose start village-api >/dev/null 2>&1 || true' EXIT

echo '1/3 Start stack and validate alerting components...'
docker compose up -d
wait_http_ok API-liveness http://localhost:8080/health
wait_http_ok Prometheus http://localhost:9090/-/ready
wait_http_ok Alertmanager http://localhost:9093/-/ready

echo '2/3 Trigger controlled API outage and wait for Prometheus evaluation...'
docker compose stop village-api >/dev/null

deadline=$((SECONDS + failure_timeout_seconds))
prometheus_alert_firing=false
while (( SECONDS < deadline )); do
  if prometheus_api_unavailable_firing; then
    prometheus_alert_firing=true
    break
  fi
  sleep 5
done

if [[ "$prometheus_alert_firing" != "true" ]]; then
  echo "Prometheus did not fire VillageAPIUnavailable within $failure_timeout_seconds seconds." >&2
  exit 1
fi

echo '[OK] Prometheus is firing VillageAPIUnavailable.'

deadline=$((SECONDS + 60))
alertmanager_received=false
while (( SECONDS < deadline )); do
  if alertmanager_api_unavailable_active; then
    alertmanager_received=true
    break
  fi
  sleep 2
done

if [[ "$alertmanager_received" != "true" ]]; then
  echo 'Alertmanager did not receive the Prometheus-fired VillageAPIUnavailable alert.' >&2
  exit 1
fi

echo '[OK] Alertmanager received VillageAPIUnavailable.'

echo '3/3 Restore API and verify recovery...'
docker compose start village-api >/dev/null
wait_http_ok API-liveness-after-recovery http://localhost:8080/health "$recovery_timeout_seconds"
wait_http_ok API-readiness-after-recovery http://localhost:8080/ready "$recovery_timeout_seconds"

echo 'Live alert chain passed: API failure -> Prometheus firing alert -> Alertmanager receipt -> API recovery.'
