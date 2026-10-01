[CmdletBinding()]
param(
    [Parameter(Mandatory = $true, Position = 0)]
    [ValidateSet('api-kill', 'postgres-outage', 'tempo-outage')]
    [string]$Scenario,

    [string]$BaseUrl = $(if ($env:CHAOS_BASE_URL) { $env:CHAOS_BASE_URL } else { 'http://localhost:8080' }),

    [int]$RecoveryTimeoutSec = 90,

    [int]$PollIntervalSec = 2,

    [switch]$ConfirmChaos
)

$ErrorActionPreference = 'Stop'

if (-not $ConfirmChaos) {
    throw "Chaos experiment '$Scenario' is destructive. Re-run with -ConfirmChaos."
}

$repoRoot = Split-Path -Parent $PSScriptRoot
$resultDir = Join-Path $repoRoot 'artifacts/chaos'
New-Item -ItemType Directory -Force -Path $resultDir | Out-Null

$runId = Get-Date -Format 'yyyyMMdd-HHmmss'
$resultPath = Join-Path $resultDir "$Scenario-$runId.json"
$base = $BaseUrl.TrimEnd('/')

$observations = New-Object System.Collections.Generic.List[object]
$startedAt = (Get-Date).ToUniversalTime()

function Get-HttpStatus {
    param([string]$Path)

    $uri = "$base$Path"
    $output = & curl.exe -sS -o NUL -w "%{http_code}" --max-time 5 $uri 2>$null

    if ($LASTEXITCODE -ne 0) {
        return 0
    }

    if ($output -match '^[0-9]{3}$') {
        return [int]$output
    }

    return 0
}

function Wait-ForStatus {
    param(
        [string]$Path,
        [int[]]$ExpectedStatus,
        [int]$TimeoutSec,
        [string]$Label
    )

    $deadline = (Get-Date).AddSeconds($TimeoutSec)

    do {
        $status = Get-HttpStatus -Path $Path

        $observations.Add([pscustomobject]@{
            timestamp = (Get-Date).ToUniversalTime().ToString('o')
            phase = 'poll'
            label = $Label
            path = $Path
            status = $status
        })

        if ($ExpectedStatus -contains $status) {
            return $true
        }

        Start-Sleep -Seconds $PollIntervalSec
    } while ((Get-Date) -lt $deadline)

    return $false
}

function Record-Status {
    param(
        [string]$Phase,
        [string]$Label,
        [string]$Path
    )

    $status = Get-HttpStatus -Path $Path
    $observations.Add([pscustomobject]@{
        timestamp = (Get-Date).ToUniversalTime().ToString('o')
        phase = $Phase
        label = $Label
        path = $Path
        status = $status
    })

    return $status
}

function Invoke-Compose {
    param([string[]]$Arguments)

    Push-Location $repoRoot
    try {
        & docker compose @Arguments
        if ($LASTEXITCODE -ne 0) {
            throw "docker compose $($Arguments -join ' ') failed with exit code $LASTEXITCODE."
        }
    }
    finally {
        Pop-Location
    }
}

function Run-K6Smoke {
    & (Join-Path $repoRoot 'scripts/run-k6.ps1') smoke -BaseUrl $base
    if ($LASTEXITCODE -ne 0) {
        throw 'k6 smoke verification failed after recovery.'
    }
}

function Preflight {
    Write-Host "Preflight: checking API liveness and readiness at $base"

    if ((Record-Status -Phase 'preflight' -Label 'liveness' -Path '/health') -ne 200) {
        throw 'Preflight failed: /health is not returning HTTP 200.'
    }

    if ((Record-Status -Phase 'preflight' -Label 'readiness' -Path '/ready') -ne 200) {
        throw 'Preflight failed: /ready is not returning HTTP 200.'
    }
}

function Run-ApiKill {
    Write-Host 'Experiment: SIGKILL village-api and verify automatic restart.'
    Invoke-Compose -Arguments @('kill', '--signal', 'SIGKILL', 'village-api')

    if (-not (Wait-ForStatus -Path '/health' -ExpectedStatus @(200) -TimeoutSec $RecoveryTimeoutSec -Label 'api liveness recovered')) {
        throw "API did not recover /health within $RecoveryTimeoutSec seconds."
    }

    if (-not (Wait-ForStatus -Path '/ready' -ExpectedStatus @(200) -TimeoutSec $RecoveryTimeoutSec -Label 'api readiness recovered')) {
        throw "API did not recover /ready within $RecoveryTimeoutSec seconds."
    }

    Run-K6Smoke
}

function Run-PostgresOutage {
    Write-Host 'Experiment: stop PostgreSQL and verify readiness fails without killing liveness.'
    Invoke-Compose -Arguments @('stop', 'postgres')

    if (-not (Wait-ForStatus -Path '/health' -ExpectedStatus @(200) -TimeoutSec 20 -Label 'liveness remains healthy during database outage')) {
        throw 'Liveness did not remain healthy during PostgreSQL outage.'
    }

    if (-not (Wait-ForStatus -Path '/ready' -ExpectedStatus @(503) -TimeoutSec 20 -Label 'readiness reports dependency outage')) {
        throw 'Readiness did not report HTTP 503 while PostgreSQL was stopped.'
    }

    Write-Host 'Recovery: starting PostgreSQL.'
    Invoke-Compose -Arguments @('start', 'postgres')

    if (-not (Wait-ForStatus -Path '/ready' -ExpectedStatus @(200) -TimeoutSec $RecoveryTimeoutSec -Label 'readiness recovered after database restart')) {
        throw "API did not recover /ready within $RecoveryTimeoutSec seconds after PostgreSQL restart."
    }

    Run-K6Smoke
}

function Run-TempoOutage {
    Write-Host 'Experiment: stop Tempo and verify serving traffic remains healthy.'
    Invoke-Compose -Arguments @('stop', 'tempo')

    if (-not (Wait-ForStatus -Path '/health' -ExpectedStatus @(200) -TimeoutSec 20 -Label 'liveness remains healthy during Tempo outage')) {
        throw 'Liveness did not remain healthy during Tempo outage.'
    }

    if (-not (Wait-ForStatus -Path '/ready' -ExpectedStatus @(200) -TimeoutSec 20 -Label 'readiness remains healthy during Tempo outage')) {
        throw 'Readiness did not remain healthy during Tempo outage.'
    }

    $communityStatus = Record-Status -Phase 'degraded' -Label 'community list remains available without Tempo' -Path '/api/v1/communities?limit=20&offset=0'
    if ($communityStatus -ne 200) {
        throw 'Community listing failed while Tempo was unavailable.'
    }

    Write-Host 'Recovery: starting Tempo.'
    Invoke-Compose -Arguments @('start', 'tempo')

    Run-K6Smoke
}

$success = $false
$errorMessage = $null

try {
    Preflight

    switch ($Scenario) {
        'api-kill' { Run-ApiKill }
        'postgres-outage' { Run-PostgresOutage }
        'tempo-outage' { Run-TempoOutage }
    }

    $success = $true
    Write-Host "Chaos experiment '$Scenario' passed."
}
catch {
    $errorMessage = $_.Exception.Message
    Write-Host "Chaos experiment '$Scenario' failed: $errorMessage" -ForegroundColor Red
    throw
}
finally {
    if ($Scenario -eq 'postgres-outage') {
        try {
            Invoke-Compose -Arguments @('start', 'postgres')
        }
        catch {
            if (-not $errorMessage) {
                $errorMessage = $_.Exception.Message
            }
        }
    }

    if ($Scenario -eq 'tempo-outage') {
        try {
            Invoke-Compose -Arguments @('start', 'tempo')
        }
        catch {
            if (-not $errorMessage) {
                $errorMessage = $_.Exception.Message
            }
        }
    }

    $endedAt = (Get-Date).ToUniversalTime()
    $durationMs = [math]::Round(($endedAt - $startedAt).TotalMilliseconds)

    $result = [pscustomobject]@{
        scenario = $Scenario
        base_url = $base
        started_at = $startedAt.ToString('o')
        ended_at = $endedAt.ToString('o')
        duration_ms = $durationMs
        success = $success
        error = $errorMessage
        observations = $observations
    }

    $result | ConvertTo-Json -Depth 8 | Set-Content -Path $resultPath -Encoding utf8
    Write-Host "Evidence: $resultPath"
}
