#!/usr/bin/env bash
set -euo pipefail

API_IMAGE="${VILLAGE_API_IMAGE:-village-api:local}"
DB_PASSWORD="${VILLAGE_DB_PASSWORD:-village}"
ENABLE_HPA=false

while [[ $# -gt 0 ]]; do
  case "$1" in
    --api-image) API_IMAGE="$2"; shift 2 ;;
    --enable-hpa) ENABLE_HPA=true; shift ;;
    *) echo "Usage: $0 [--api-image IMAGE] [--enable-hpa]" >&2; exit 2 ;;
  esac
done

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
k8s_dir="$repo_root/infra/kubernetes"

echo "Using API image: $API_IMAGE"
kubectl cluster-info >/dev/null
kubectl apply -f "$k8s_dir/namespace.yaml"

if ! kubectl get secret village-postgres-secret -n village >/dev/null 2>&1; then
  cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: village-postgres-secret
  namespace: village
type: Opaque
stringData:
  POSTGRES_DB: village
  POSTGRES_USER: village
  POSTGRES_PASSWORD: ${DB_PASSWORD}
EOF
else
  echo 'PostgreSQL secret already exists; retaining it for database compatibility.'
fi

kubectl apply -f "$k8s_dir/configmap.yaml"
kubectl apply -f "$k8s_dir/postgres.yaml"
kubectl rollout status statefulset/village-postgres -n village --timeout=180s

kubectl create configmap village-migrations -n village \
  --from-file="$repo_root/migrations" \
  --dry-run=client -o yaml | kubectl apply -f -

kubectl delete job village-migrate -n village --ignore-not-found
kubectl apply -f "$k8s_dir/migration-job.yaml"
kubectl wait --for=condition=complete job/village-migrate -n village --timeout=180s

sed "s#image: village-api:local#image: ${API_IMAGE}#" "$k8s_dir/api.yaml" | kubectl apply -f -
kubectl apply -f "$k8s_dir/pdb.yaml"

if [[ "$ENABLE_HPA" == true ]]; then
  kubectl apply -f "$k8s_dir/hpa.yaml"
else
  kubectl delete hpa village-api -n village --ignore-not-found >/dev/null
fi

kubectl rollout status deployment/village-api -n village --timeout=180s
kubectl get pods -n village -o wide
kubectl get svc -n village

echo "Kubernetes deployment complete. API NodePort: http://localhost:30080"
