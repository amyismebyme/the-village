#!/usr/bin/env bash
set -euo pipefail
replicas="${1:-3}"
timeout="${K8S_TIMEOUT:-120s}"

case "$replicas" in
  1|2|3|4|5) ;;
  *) echo 'Usage: $0 {1|2|3|4|5}' >&2; exit 2 ;;
esac

kubectl scale deployment/village-api -n village --replicas="$replicas"
kubectl rollout status deployment/village-api -n village --timeout="$timeout"
kubectl get deployment village-api -n village
kubectl get pods -n village -l app.kubernetes.io/name=village-api -o wide

echo 'Scaling verification complete. Test endpoint: http://localhost:30080/health'
