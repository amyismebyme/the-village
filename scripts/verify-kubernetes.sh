#!/usr/bin/env bash
set -euo pipefail
recovery_timeout="${K8S_RECOVERY_TIMEOUT:-120s}"
api_url="${K8S_API_URL:-http://localhost:30080}"

wait_http_ok() {
  local path="$1"
  local timeout_s="${2:-120}"
  local deadline=$((SECONDS + timeout_s))
  while (( SECONDS < deadline )); do
    if curl -fsS --max-time 5 "$api_url$path" >/dev/null; then return 0; fi
    sleep 2
  done
  echo "$path did not return HTTP 200 within ${timeout_s}s" >&2
  return 1
}

kubectl cluster-info >/dev/null
kubectl rollout status statefulset/village-postgres -n village --timeout="$recovery_timeout"
succeeded="$(kubectl get job village-migrate -n village -o jsonpath='{.status.succeeded}')"
[[ "$succeeded" == '1' ]] || { echo 'village-migrate has not completed successfully.' >&2; exit 1; }

kubectl rollout status deployment/village-api -n village --timeout="$recovery_timeout"
wait_http_ok /health 120
wait_http_ok /ready 120

kubectl scale deployment/village-api -n village --replicas=3
kubectl rollout status deployment/village-api -n village --timeout="$recovery_timeout"
available="$(kubectl get deployment village-api -n village -o jsonpath='{.status.availableReplicas}')"
[[ "$available" =~ ^[0-9]+$ ]] && (( available >= 3 )) || { echo "Expected 3 available replicas, got $available" >&2; exit 1; }

pod="$(kubectl get pods -n village -l app.kubernetes.io/name=village-api -o jsonpath='{.items[0].metadata.name}')"
[[ -n "$pod" ]] || { echo 'No API pod found.' >&2; exit 1; }
kubectl delete pod "$pod" -n village --wait=false
wait_http_ok /health 120
wait_http_ok /ready 120

deadline=$((SECONDS + 120))
while (( SECONDS < deadline )); do
  available="$(kubectl get deployment village-api -n village -o jsonpath='{.status.availableReplicas}')"
  if [[ "$available" =~ ^[0-9]+$ ]] && (( available >= 3 )); then break; fi
  sleep 2
done
available="$(kubectl get deployment village-api -n village -o jsonpath='{.status.availableReplicas}')"
[[ "$available" =~ ^[0-9]+$ ]] && (( available >= 3 )) || { echo "Expected 3 available replicas after recovery, got $available" >&2; exit 1; }

kubectl get deployment village-api -n village
kubectl get pods -n village -o wide
kubectl get svc -n village
kubectl top pods -n village || true
kubectl get hpa -n village || true

echo 'Sprint 28 Kubernetes deployment, scaling, probes, migration ordering, and pod recovery verification passed.'
