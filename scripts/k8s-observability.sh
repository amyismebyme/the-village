#!/usr/bin/env bash
set -euo pipefail

NAMESPACE="${NAMESPACE:-village}"
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
K8S_DIR="$REPO_ROOT/infra/kubernetes"

kubectl cluster-info
kubectl get namespace "$NAMESPACE" >/dev/null

kubectl apply -f "$K8S_DIR/prometheus-rbac.yaml"
kubectl apply -f "$K8S_DIR/kube-state-metrics.yaml"
kubectl rollout status deployment/village-kube-state-metrics -n "$NAMESPACE" --timeout=180s
kubectl apply -f "$K8S_DIR/node-exporter.yaml"
kubectl rollout status daemonset/village-node-exporter -n "$NAMESPACE" --timeout=180s
kubectl apply -f "$K8S_DIR/prometheus-k8s.yaml"
kubectl rollout status deployment/village-k8s-prometheus -n "$NAMESPACE" --timeout=180s

kubectl get pods -n "$NAMESPACE" -o wide
kubectl get svc -n "$NAMESPACE"

echo "Kubernetes observability deployment complete."
echo "Prometheus: http://localhost:30090"
echo "Grafana:    http://localhost:3000"
