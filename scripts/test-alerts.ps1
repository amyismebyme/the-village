$ErrorActionPreference = 'Stop'
$repoRoot = Split-Path -Parent $PSScriptRoot
$prometheusImage = 'prom/prometheus:latest'
$mount = "${repoRoot}:/workspace"

Write-Host 'Running Prometheus alert rule tests...'
docker run --rm -v $mount $prometheusImage promtool test rules /workspace/infra/docker/prometheus/tests/alerts.test.yml
if ($LASTEXITCODE -ne 0) {
    throw 'Prometheus alert rule tests failed.'
}

Write-Host 'Alert rule tests passed.'
