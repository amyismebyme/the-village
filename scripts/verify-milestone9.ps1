$ErrorActionPreference = 'Stop'
$repoRoot = Split-Path -Parent $PSScriptRoot
Push-Location $repoRoot
try {
    Write-Host '=== Milestone 9: Compose configuration ==='
    docker compose config | Out-Null

    Write-Host '=== Milestone 9: security/cardinality ==='
    .\scripts\verify-observability-security.ps1

    Write-Host '=== Milestone 9: alert rule evaluation ==='
    .\scripts\test-alerts.ps1

    Write-Host '=== Milestone 9: alert chain ==='
    .\scripts\test-alert-chain.ps1

    Write-Host '=== Milestone 9: Docker stack ==='
    .\scripts\verify-observability-stack.ps1

    Write-Host '=== Milestone 9: Go tests ==='
    Push-Location (Join-Path $repoRoot 'apps/api')
    try {
        go test ./... -count=1
    } finally {
        Pop-Location
    }

    Write-Host '=== Milestone 9: performance smoke benchmarks ==='
    .\scripts\benchmark-observability.ps1

    Write-Host ''
    Write-Host 'Milestone 9 verification gate completed successfully.'
} finally {
    Pop-Location
}
