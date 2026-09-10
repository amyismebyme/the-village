$ErrorActionPreference = 'Stop'

$root = Split-Path -Parent $PSScriptRoot

$forbidden = '(?i)(access[_-]?token|refresh[_-]?token|client[_-]?secret|authorization|password|cookie|credential|dsn|connection[_-]?string)'
$telemetryFiles = Get-ChildItem -Path $root -Recurse -File -Include *.go,*.yml,*.yaml |
    Where-Object { $_.FullName -notmatch '\\.git\\|vendor\\|api\.exe$' }

$violations = @()
foreach ($file in $telemetryFiles) {
    $isTelemetryFile = $file.FullName -match '\\internal\\(metrics|observability|telemetry|logger)\\|\\infra\\docker\\prometheus\\'
    if (-not $isTelemetryFile) { continue }

    $lineNo = 0
    foreach ($line in Get-Content -LiteralPath $file.FullName) {
        $lineNo++
        if ($line -match $forbidden -and $line -match 'NewCounterVec|NewGaugeVec|NewHistogramVec|attribute\.|WithAttributes|labels:|labelnames:') {
            $violations += "${($file.FullName)}:$lineNo`t$line"
        }
    }
}

if ($violations.Count -gt 0) {
    $violations | ForEach-Object { Write-Error $_ }
    exit 1
}

Write-Host 'Observability security scan passed.'
