#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
docker run --rm -v "${repo_root}:/workspace" prom/prometheus:latest \
  promtool test rules /workspace/infra/docker/prometheus/tests/alerts.test.yml

echo "Prometheus alert rule tests passed."
