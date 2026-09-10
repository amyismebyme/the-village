$ErrorActionPreference = 'Stop'

param(
    [string]$BaseUrl = 'http://localhost:8080',
    [string]$TempoUrl = 'http://localhost:3200',
    [int]$LookbackMinutes = 2
)

function Get-LatestTraceId {
    $logs = docker compose logs village-api --since "${LookbackMinutes}m" --no-color 2>$null
    if (-not $logs) { return $null }

    $traceIds = foreach ($line in $logs) {
        if ($line -match '"trace_id"\s*:\s*"([0-9a-f]{32})"') {
            $matches[1]
        } elseif ($line -match 'trace_id=([0-9a-f]{32})') {
            $matches[1]
        }
    }

    $traceIds | Select-Object -Last 1
}

Write-Host "Calling $BaseUrl/health ..."
$response = Invoke-WebRequest -Uri "$BaseUrl/health" -Method Get -UseBasicParsing
if ($response.StatusCode -lt 200 -or $response.StatusCode -ge 400) {
    throw "Health request failed with HTTP $($response.StatusCode)."
}

$traceId = Get-LatestTraceId
if (-not $traceId) {
    throw 'Could not find a trace_id in recent village-api logs. Ensure trace/log correlation is enabled and use InfoContext/WarnContext/ErrorContext.'
}

Write-Host "Found trace_id: $traceId"
$trace = Invoke-WebRequest -Uri "$TempoUrl/api/traces/$traceId" -Method Get -UseBasicParsing
if ($trace.StatusCode -lt 200 -or $trace.StatusCode -ge 400) {
    throw "Tempo trace lookup failed with HTTP $($trace.StatusCode)."
}

$body = $trace.Content
foreach ($spanName in @('http', 'community', 'reddit', 'db')) {
    if ($body -notmatch [regex]::Escape($spanName)) {
        Write-Warning "Trace payload did not visibly contain '$spanName'. Verify the exact span name used by the current instrumentation."
    }
}

Write-Host 'Tempo returned the trace successfully.'
Write-Host 'Use the trace_id in Grafana/Loki to verify log -> trace correlation and inspect child span names.'
