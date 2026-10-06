#!/usr/bin/env bash
set -euo pipefail

kubectl delete deployment/village-api -n village --ignore-not-found
kubectl delete service/village-api -n village --ignore-not-found
kubectl delete hpa/village-api -n village --ignore-not-found
kubectl delete pdb/village-api -n village --ignore-not-found
kubectl delete job/village-migrate -n village --ignore-not-found
kubectl delete statefulset/village-postgres -n village --ignore-not-found
kubectl delete service/village-postgres -n village --ignore-not-found
kubectl delete configmap/village-api-config -n village --ignore-not-found
kubectl delete configmap/village-migrations -n village --ignore-not-found
kubectl delete secret/village-postgres-secret -n village --ignore-not-found

if [[ "${1:-}" == '--delete-data' ]]; then
  kubectl delete pvc -n village -l app.kubernetes.io/name=village-postgres --ignore-not-found
  echo 'PostgreSQL PVC data removed.'
fi

echo 'Village Kubernetes workloads removed.'
