#!/usr/bin/env bash
set -euo pipefail
repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root/apps/api"

go test ./internal/telemetry -run '^$' -bench 'BenchmarkHTTPMiddlewareTelemetry$' -benchmem -count=3
go test ./internal/external/reddit -run '^$' -bench 'BenchmarkIngestionThroughputTelemetry$' -benchmem -count=3
