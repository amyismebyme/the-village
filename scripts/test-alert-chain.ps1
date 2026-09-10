$ErrorActionPreference = 'Stop'
$repoRoot = Split-Path -Parent $PSScriptRoot
Push-Location $repoRoot
try {
    Write-Host '1/2 Prometheus rule evaluation tests...'
    .\scripts\test-alerts.ps1

    Write-Host '2/2 Alertmanager routing-state test...'
    $fingerprint = [Guid]::NewGuid().ToString('N')
    $now = (Get-Date).ToUniversalTime()
    $end = $now.AddMinutes(5)
    $payload = @(
        @{
            labels = @{
                alertname = 'VillageControlledTestAlert'
                service = 'village-api'
                severity = 'critical'
                source = 'test'
                fingerprint = $fingerprint
            }
            annotations = @{
                summary = 'Milestone 9 controlled alert test'
                description = 'Synthetic alert used to prove Alertmanager receives and retains an alert.'
            }
            startsAt = $now.ToString('o')
            endsAt = $end.ToString('o')
        }
    ) | ConvertTo-Json -Depth 8

    Invoke-RestMethod `
        -Uri 'http://localhost:9093/api/v2/alerts' `
        -Method Post `
        -ContentType 'application/json' `
        -Body $payload | Out-Null

    Start-Sleep -Seconds 2
    $alerts = Invoke-RestMethod -Uri 'http://localhost:9093/api/v2/alerts' -Method Get
    $found = $alerts | Where-Object {
        $_.labels.alertname -eq 'VillageControlledTestAlert' -and
        $_.labels.service -eq 'village-api'
    }

    if (-not $found) {
        throw 'Alertmanager did not retain the controlled test alert.'
    }

    Write-Host 'Controlled alert -> Alertmanager routing-state test passed.'
} finally {
    Pop-Location
}
