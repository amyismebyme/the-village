#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

./scripts/test-alerts.sh

fingerprint="$(cat /proc/sys/kernel/random/uuid | tr -d '-')"
now="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
end="$(date -u -d '+5 minutes' +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || date -u -v+5M +%Y-%m-%dT%H:%M:%SZ)"

curl -fsS \
  -X POST \
  -H 'Content-Type: application/json' \
  http://localhost:9093/api/v2/alerts \
  -d "[{\"labels\":{\"alertname\":\"VillageControlledTestAlert\",\"service\":\"village-api\",\"severity\":\"critical\",\"source\":\"test\",\"fingerprint\":\"${fingerprint}\"},\"annotations\":{\"summary\":\"Milestone 9 controlled alert test\",\"description\":\"Synthetic alert used to prove Alertmanager receives and retains an alert.\"},\"startsAt\":\"${now}\",\"endsAt\":\"${end}\"}]" >/dev/null

sleep 2
curl -fsS http://localhost:9093/api/v2/alerts | grep -q 'VillageControlledTestAlert'
echo 'Controlled alert -> Alertmanager routing-state test passed.'
