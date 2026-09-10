[CmdletBinding()]
param(
    [switch]$SkipBackendOutageChecks
)

$ErrorActionPreference = 'Stop'
$repoRoot = Split-Path -Parent $PSScriptRoot
Push-Location $repoRoot
try {
    Write-Host 'Validating Docker Compose configuration...'
    docker compose config | Out-Null

    Write-Host 'Starting observability stack...'
    docker compose up -d

    function Wait-HttpOk([string]$Name, [string]$Uri, [int]$TimeoutSeconds = 90) {
        $deadline = (Get-Date).AddSeconds($TimeoutSeconds)
        do {
            try {
                $response = Invoke-WebRequest -Uri $Uri -Method Get -UseBasicParsing -TimeoutSec 5
                if ($response.StatusCode -ge 200 -and $response.StatusCode -lt 400) {
                    Write-Host "$Name healthy: $Uri"
                    return
                }
            } catch { }
            Start-Sleep -Seconds 2
        } while ((Get-Date) -lt $deadline)

        throw "$Name did not become healthy: $Uri"
    }

    Wait-HttpOk 'API' 'http://localhost:8080/health'
    Wait-HttpOk 'Prometheus' 'http://localhost:9090/-/ready'
    Wait-HttpOk 'Grafana' 'http://localhost:3000/api/health'
    Wait-HttpOk 'Loki' 'http://localhost:3100/ready'
    Wait-HttpOk 'Tempo' 'http://localhost:3200/ready'
    Wait-HttpOk 'Alertmanager' 'http://localhost:9093/-/ready'

    Write-Host 'Checking Compose service state...'
    docker compose ps

    if (-not $SkipBackendOutageChecks) {
        foreach ($service in @('tempo', 'loki', 'alertmanager')) {
            Write-Host "Testing API availability with $service unavailable..."
            docker compose stop $service | Out-Null
            try {
                Wait-HttpOk "API while $service is stopped" 'http://localhost:8080/health' 30
            } finally {
                docker compose start $service | Out-Null
                if ($service -eq 'tempo') {
                    Wait-HttpOk 'Tempo restored' 'http://localhost:3200/ready'
                } elseif ($service -eq 'loki') {
                    Wait-HttpOk 'Loki restored' 'http://localhost:3100/ready'
                } else {
                    Wait-HttpOk 'Alertmanager restored' 'http://localhost:9093/-/ready'
                }
            }
        }
    }

    Write-Host 'Docker observability stack verification passed.'
} finally {
    Pop-Location
}
