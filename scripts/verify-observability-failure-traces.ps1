param(
    [Parameter(Mandatory=$true)][string]$TraceId,
    [string]$TempoUrl = 'http://localhost:3200'
)

$ErrorActionPreference = 'Stop'

if ($TraceId -notmatch '^[0-9a-fA-F]{32}$') {
    throw 'TraceId must be exactly 32 hexadecimal characters.'
}

$trace = Invoke-WebRequest -Uri "$TempoUrl/api/traces/$TraceId" -Method Get -UseBasicParsing
if ($trace.StatusCode -lt 200 -or $trace.StatusCode -ge 400) {
    throw "Tempo trace lookup failed with HTTP $($trace.StatusCode)."
}

$content = $trace.Content

$scenarios = @(
    @{ Name = 'Reddit 503'; Tokens = @('reddit', '5xx') },
    @{ Name = 'Rate limiting'; Tokens = @('rate', '429') },
    @{ Name = 'Timeout'; Tokens = @('timeout') },
    @{ Name = 'Retry exhaustion'; Tokens = @('retry', 'exhaust') },
    @{ Name = 'Database failure'; Tokens = @('db', 'failure', 'postgres') },
    @{ Name = 'Worker failure'; Tokens = @('worker', 'failure') },
    @{ Name = 'Context cancellation'; Tokens = @('canceled', 'cancelled', 'cancellation') }
)

Write-Host "Trace $TraceId loaded."
foreach ($scenario in $scenarios) {
    $hits = @($scenario.Tokens | Where-Object { $content -match $_ })
    if ($hits.Count -eq 0) {
        Write-Warning "$($scenario.Name): no obvious marker found. Inspect the trace manually and confirm the current bounded span/error attribute names."
    } else {
        Write-Host "$($scenario.Name): marker(s) found -> $($hits -join ', ')"
    }
}

Write-Host 'Failure-trace verification complete. This script validates the supplied trace payload; failure injection remains environment-specific.'
