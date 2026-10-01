#!/usr/bin/env bash
set -euo pipefail

SCENARIO="${1:-}"
BASE_URL="${CHAOS_BASE_URL:-http://localhost:8080}"
RECOVERY_TIMEOUT_SEC="${CHAOS_RECOVERY_TIMEOUT_SEC:-90}"
POLL_INTERVAL_SEC="${CHAOS_POLL_INTERVAL_SEC:-2}"
CONFIRM="${CHAOS_CONFIRM:-}"

case "$SCENARIO" in
  api-kill|postgres-outage|tempo-outage) ;;
  *)
    echo "Usage: CHAOS_CONFIRM=1 $0 {api-kill|postgres-outage|tempo-outage}" >&2
    exit 2
    ;;
esac

if [[ "$CONFIRM" != "1" ]]; then
  echo "Chaos experiment '$SCENARIO' is destructive. Re-run with CHAOS_CONFIRM=1." >&2
  exit 2
fi

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
RESULT_DIR="$REPO_ROOT/artifacts/chaos"
mkdir -p "$RESULT_DIR"
RUN_ID="$(date -u +%Y%m%d-%H%M%S)"
RESULT_PATH="$RESULT_DIR/$SCENARIO-$RUN_ID.json"
BASE_URL="${BASE_URL%/}"
START_EPOCH="$(date +%s)"
STARTED_AT="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
OBSERVATIONS=""

SUCCESS=false
ERROR_MESSAGE=''

record_observation() {
  local phase="$1"
  local label="$2"
  local path="$3"
  local status="$4"
  local timestamp
  timestamp="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  local item
  item="{\"timestamp\":\"$timestamp\",\"phase\":\"$phase\",\"label\":\"$label\",\"path\":\"$path\",\"status\":$status}"
  if [[ -z "$OBSERVATIONS" ]]; then
    OBSERVATIONS="$item"
  else
    OBSERVATIONS="$OBSERVATIONS,$item"
  fi
}

http_status() {
  local path="$1"
  curl -sS -o /dev/null -w '%{http_code}' --max-time 5 "$BASE_URL$path" 2>/dev/null || true
}

record_status() {
  local phase="$1"
  local label="$2"
  local path="$3"
  local status
  status="$(http_status "$path")"
  [[ "$status" =~ ^[0-9]{3}$ ]] || status=0
  record_observation "$phase" "$label" "$path" "$status"
  printf '%s\n' "$status"
}

wait_for_status() {
  local path="$1"
  local expected="$2"
  local timeout="$3"
  local label="$4"
  local deadline=$((SECONDS + timeout))
  local status

  while (( SECONDS < deadline )); do
    status="$(http_status "$path")"
    [[ "$status" =~ ^[0-9]{3}$ ]] || status=0
    record_observation 'poll' "$label" "$path" "$status"

    if [[ "$status" == "$expected" ]]; then
      return 0
    fi

    sleep "$POLL_INTERVAL_SEC"
  done

  return 1
}

compose() {
  docker compose "$@"
}

preflight() {
  echo "Preflight: checking API liveness and readiness at $BASE_URL"

  local status
  status="$(http_status /health)"
  [[ "$status" =~ ^[0-9]{3}$ ]] || status=0
  record_observation preflight liveness /health "$status"
  [[ "$status" == "200" ]] || {
    echo 'Preflight failed: /health is not returning HTTP 200.' >&2
    return 1
  }

  status="$(http_status /ready)"
  [[ "$status" =~ ^[0-9]{3}$ ]] || status=0
  record_observation preflight readiness /ready "$status"
  [[ "$status" == "200" ]] || {
    echo 'Preflight failed: /ready is not returning HTTP 200.' >&2
    return 1
  }
}

run_k6_smoke() {
  "$REPO_ROOT/scripts/run-k6.sh" smoke
}

run_api_kill() {
  echo 'Experiment: SIGKILL village-api and verify automatic restart.'
  compose kill --signal SIGKILL village-api

  wait_for_status /health 200 "$RECOVERY_TIMEOUT_SEC" 'api liveness recovered' || {
    echo "API did not recover /health within ${RECOVERY_TIMEOUT_SEC}s." >&2
    return 1
  }

  wait_for_status /ready 200 "$RECOVERY_TIMEOUT_SEC" 'api readiness recovered' || {
    echo "API did not recover /ready within ${RECOVERY_TIMEOUT_SEC}s." >&2
    return 1
  }

  run_k6_smoke
}

run_postgres_outage() {
  echo 'Experiment: stop PostgreSQL and verify readiness fails without killing liveness.'
  compose stop postgres

  wait_for_status /health 200 20 'liveness remains healthy during database outage' || {
    echo 'Liveness did not remain healthy during PostgreSQL outage.' >&2
    return 1
  }

  wait_for_status /ready 503 20 'readiness reports dependency outage' || {
    echo 'Readiness did not report HTTP 503 while PostgreSQL was stopped.' >&2
    return 1
  }

  echo 'Recovery: starting PostgreSQL.'
  compose start postgres

  wait_for_status /ready 200 "$RECOVERY_TIMEOUT_SEC" 'readiness recovered after database restart' || {
    echo "API did not recover /ready within ${RECOVERY_TIMEOUT_SEC}s after PostgreSQL restart." >&2
    return 1
  }

  run_k6_smoke
}

run_tempo_outage() {
  echo 'Experiment: stop Tempo and verify serving traffic remains healthy.'
  compose stop tempo

  wait_for_status /health 200 20 'liveness remains healthy during Tempo outage' || {
    echo 'Liveness did not remain healthy during Tempo outage.' >&2
    return 1
  }

  wait_for_status /ready 200 20 'readiness remains healthy during Tempo outage' || {
    echo 'Readiness did not remain healthy during Tempo outage.' >&2
    return 1
  }

  local status
  status="$(record_status degraded 'community list remains available without Tempo' '/api/v1/communities?limit=20&offset=0')"
  [[ "$status" == "200" ]] || {
    echo 'Community listing failed while Tempo was unavailable.' >&2
    return 1
  }

  echo 'Recovery: starting Tempo.'
  compose start tempo

  run_k6_smoke
}

write_evidence() {
  local ended_at duration_s
  ended_at="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  duration_s="$(( $(date +%s) - START_EPOCH ))"
  cat > "$RESULT_PATH" <<EOF
{
  "scenario": "${SCENARIO}",
  "base_url": "${BASE_URL}",
  "started_at": "${STARTED_AT}",
  "ended_at": "${ended_at}",
  "duration_seconds": ${duration_s},
  "success": ${SUCCESS},
  "error": $(printf '%s' "${ERROR_MESSAGE:-}" | sed 's/\\/\\\\/g; s/"/\\"/g' | awk '{printf "\"%s\"", $0}'),
  "observations": [${OBSERVATIONS}]
}
EOF
  echo "Evidence: $RESULT_PATH"
}

cleanup() {
  set +e
  if [[ "$SCENARIO" == 'postgres-outage' ]]; then
    compose start postgres >/dev/null 2>&1
  fi
  if [[ "$SCENARIO" == 'tempo-outage' ]]; then
    compose start tempo >/dev/null 2>&1
  fi
  write_evidence
}
trap cleanup EXIT

if ! preflight; then
  ERROR_MESSAGE='Preflight failed.'
  exit 1
fi

case "$SCENARIO" in
  api-kill) run_api_kill ;;
  postgres-outage) run_postgres_outage ;;
  tempo-outage) run_tempo_outage ;;
esac

SUCCESS=true
echo "Chaos experiment '$SCENARIO' passed."
