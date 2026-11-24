#!/usr/bin/env bash
set -euo pipefail

NAMESPACE="${NAMESPACE:-village}"
PROM_URL="${PROM_URL:-http://localhost:30090}"
SKIP_KUBELET_CHECK="${SKIP_KUBELET_CHECK:-0}"

kubectl get namespace "$NAMESPACE" >/dev/null
kubectl get deployment village-api -n "$NAMESPACE" >/dev/null
kubectl get deployment village-k8s-prometheus -n "$NAMESPACE" >/dev/null
kubectl get deployment village-kube-state-metrics -n "$NAMESPACE" >/dev/null
kubectl get daemonset village-node-exporter -n "$NAMESPACE" >/dev/null

curl -fsS "$PROM_URL/-/ready" >/dev/null
TARGETS="$(curl -fsS "$PROM_URL/api/v1/targets")"

die() { echo "ERROR: $*" >&2; exit 1; }

check_job() {
  local job="$1"
  echo "$TARGETS" | grep -q "\"job\":\"$job\"" || die "no active target for job $job"
  echo "$TARGETS" | grep -q "\"job\":\"$job\".*\"health\":\"up\"" || die "job $job exists but no healthy target was found"
  echo "$job healthy"
}

check_job village-api
check_job kube-state-metrics
check_job node-exporter

if [[ "$SKIP_KUBELET_CHECK" != "1" ]]; then
  check_job kubelet
  check_job kubelet-cadvisor
fi

query_has_data() {
  local query="$1"
  local description="$2"
  local encoded
  encoded="$(python3 -c 'import sys, urllib.parse; print(urllib.parse.quote(sys.argv[1]))' "$query")"
  local body
  body="$(curl -fsS "$PROM_URL/api/v1/query?query=$encoded")"
  python3 - "$body" <<'PY' || die "query failed or returned no data: $description"
import json, sys
obj = json.loads(sys.argv[1])
if obj.get('status') != 'success':
    raise SystemExit(1)
if not obj.get('data', {}).get('result'):
    raise SystemExit(1)
PY
  echo "$description metric data present"
}

query_has_data 'village_http_requests_total{namespace="village"}' 'Village API application metrics'
query_has_data 'kube_deployment_status_replicas_available{namespace="village",deployment="village-api"}' 'Kubernetes deployment metrics'
query_has_data 'node_uname_info' 'Node exporter metrics'

if [[ "$SKIP_KUBELET_CHECK" != "1" ]]; then
  query_has_data 'container_memory_working_set_bytes{namespace="village",pod=~"village-api.*"}' 'Workload container memory metrics'
fi

echo 'Kubernetes observability verification passed.'
echo "Prometheus UI: $PROM_URL"
echo 'Grafana dashboards: Village / Kubernetes'
