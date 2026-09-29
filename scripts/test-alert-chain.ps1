[CmdletBinding()]
param(
    [int]$FailureTimeoutSeconds = 210,
    [int]$RecoveryTimeoutSeconds = 90
)

$ErrorActionPreference = 'Stop'
$repoRoot = Split-Path -Parent $PSScriptRoot
Push-Location $repoRoot

function Wait-HttpOk([string]$Name, [string]$Uri, [int]$TimeoutSeconds = 90) {
    $deadline = (Get-Date).AddSeconds($TimeoutSeconds)
    do {
        try {
            $response = Invoke-WebRequest -Uri $Uri -Method Get -UseBasicParsing -TimeoutSec 5
            if ($response.StatusCode -ge 200 -and $response.StatusCode -lt 400) {
                Write-Host "[OK] $Name: $Uri"
                return
            }
        } catch { }
        Start-Sleep -Seconds 2
    } while ((Get-Date) -lt $deadline)

    throw "$Name did not become healthy: $Uri"
}

function Get-PrometheusApiUnavailableAlert {
    $response = Invoke-RestMethod -Uri 'http://localhost:9090/api/v1/alerts' -Method Get
    if ($response.status -ne 'success') {
        throw 'Prometheus alerts API did not return success.'
    }

    return @($response.data.alerts | Where-Object {
        $_.labels.alertname -eq 'VillageAPIUnavailable' -and
        $_.labels.service -eq 'village-api' -and
        $_.state -eq 'firing'
    })
}

try {
    Write-Host '1/3 Start stack and validate alerting components...'
    docker compose up -d
    if ($LASTEXITCODE -ne 0) {
        throw 'docker compose up failed.'
    }

    Wait-HttpOk 'API liveness' 'http://localhost:8080/health'
    Wait-HttpOk 'Prometheus' 'http://localhost:9090/-/ready'
    Wait-HttpOk 'Alertmanager' 'http://localhost:9093/-/ready'

    Write-Host '2/3 Trigger controlled API outage and wait for Prometheus evaluation...'
    docker compose stop village-api | Out-Null
    if ($LASTEXITCODE -ne 0) {
        throw 'Failed to stop village-api.'
    }

    $deadline = (Get-Date).AddSeconds($FailureTimeoutSeconds)
    $prometheusAlert = $null

    do {
        try {
            $prometheusAlert = Get-PrometheusApiUnavailableAlert
            if ($prometheusAlert.Count -gt 0) {
                break
            }
        } catch { }
        Start-Sleep -Seconds 5
    } while ((Get-Date) -lt $deadline)

    if (-not $prometheusAlert) {
        throw "Prometheus did not fire VillageAPIUnavailable within $FailureTimeoutSeconds seconds."
    }

    Write-Host '[OK] Prometheus is firing VillageAPIUnavailable.'

    $alertmanagerDeadline = (Get-Date).AddSeconds(60)
    $alertmanagerAlert = $null
    do {
        try {
            $alerts = Invoke-RestMethod -Uri 'http://localhost:9093/api/v2/alerts' -Method Get
            $alertmanagerAlert = @($alerts | Where-Object {
                $_.labels.alertname -eq 'VillageAPIUnavailable' -and
                $_.labels.service -eq 'village-api' -and
                $_.status.state -eq 'active'
            })
            if ($alertmanagerAlert.Count -gt 0) {
                break
            }
        } catch { }
        Start-Sleep -Seconds 2
    } while ((Get-Date) -lt $alertmanagerDeadline)

    if (-not $alertmanagerAlert) {
        throw 'Alertmanager did not receive the Prometheus-fired VillageAPIUnavailable alert.'
    }

    Write-Host '[OK] Alertmanager received VillageAPIUnavailable.'

    Write-Host '3/3 Restore API and verify recovery...'
    docker compose start village-api | Out-Null
    if ($LASTEXITCODE -ne 0) {
        throw 'Failed to start village-api.'
    }

    Wait-HttpOk 'API liveness after recovery' 'http://localhost:8080/health' $RecoveryTimeoutSeconds
    Wait-HttpOk 'API readiness after recovery' 'http://localhost:8080/ready' $RecoveryTimeoutSeconds

    Write-Host 'Live alert chain passed: API failure -> Prometheus firing alert -> Alertmanager receipt -> API recovery.'
}
finally {
    try {
        docker compose start village-api | Out-Null
    } catch { }
    Pop-Location
}
