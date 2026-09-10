$ErrorActionPreference = 'Stop'
$repoRoot = Split-Path -Parent $PSScriptRoot
Push-Location (Join-Path $repoRoot 'apps/api')
try {
    Write-Host 'Running HTTP telemetry overhead benchmark...'
    go test ./internal/telemetry -run '^$' -bench 'BenchmarkHTTPMiddlewareTelemetry$' -benchmem -count=3

    Write-Host 'Running Reddit ingestion telemetry throughput benchmark...'
    go test ./internal/external/reddit -run '^$' -bench 'BenchmarkIngestionThroughputTelemetry$' -benchmem -count=3
} finally {
    Pop-Location
}
