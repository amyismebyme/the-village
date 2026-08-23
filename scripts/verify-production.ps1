param(
    [string]$ApiBaseUrl = "http://localhost:8080",
    [switch]$SkipDocker,
    [switch]$KeepDockerRunning
)

$ErrorActionPreference = "Stop"
$RepoRoot = Resolve-Path "$PSScriptRoot\.."
$ApiDir = Join-Path $RepoRoot "apps\api"
$ComposeFile = Join-Path $RepoRoot "docker-compose.yml"

function Section($title) {
    Write-Host ""
    Write-Host "============================================================" -ForegroundColor Cyan
    Write-Host " $title" -ForegroundColor Cyan
    Write-Host "============================================================" -ForegroundColor Cyan
}

function Run-Step($title, [scriptblock]$action) {
    Section $title
    & $action
    if ($LASTEXITCODE -ne 0) {
        throw "$title failed with exit code $LASTEXITCODE"
    }
}

Section "Production Verification"

Push-Location $ApiDir
try {
    Run-Step "Formatting check" {
        $files = gofmt -l .
        if ($files) {
            Write-Host "Unformatted files:" -ForegroundColor Red
            $files
            exit 1
        }
    }

    Run-Step "Build" { go build ./... }
    Run-Step "Unit tests" { go test ./... -count=1 }
    Run-Step "Race detector" { go test -race ./... }
    Run-Step "Vet" { go vet ./... }
    Run-Step "Lint" { golangci-lint run }
    Run-Step "Integration tests" { go test -tags=integration ./internal/integration/... -v -count=1 }

    Section "Context.Background audit"
    $matches = Get-ChildItem .\internal -Recurse -File -Filter *.go |
        Select-String -Pattern 'context\.Background\(\)' |
        Where-Object { $_.Path -notmatch '_test\.go$' }

    if ($matches) {
        Write-Host "Review these production-path uses of context.Background():" -ForegroundColor Yellow
        $matches | ForEach-Object { $_.ToString() }
        Write-Host "Startup/shutdown uses may be intentional; request paths must not replace incoming contexts." -ForegroundColor Yellow
    } else {
        Write-Host "[OK] No production-path context.Background() uses found"
    }

    Section "Sensitive logging audit"
    $loggingFiles = Get-ChildItem .\internal -Recurse -File -Filter *.go |
        Where-Object { $_.FullName -match 'logging|community_logging|recovery|app\.go|error\.go' -and $_.Name -notmatch '_test\.go$' }

    $sensitiveMatches = $loggingFiles |
        Select-String -Pattern 'password|authorization|bearer|token|secret|connection string|raw SQL|request body'

    if ($sensitiveMatches) {
        Write-Host "Potential sensitive-data references require review:" -ForegroundColor Yellow
        $sensitiveMatches | ForEach-Object { $_.ToString() }
    } else {
        Write-Host "[OK] No sensitive logging references found"
    }
}
finally {
    Pop-Location
}

if (-not $SkipDocker) {
    Section "Docker runtime verification"

    if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
        throw "docker is not installed or unavailable"
    }

    if (-not (Test-Path $ComposeFile)) {
        throw "docker-compose.yml not found: $ComposeFile"
    }

    try {
        docker compose -f $ComposeFile up -d --build
        if ($LASTEXITCODE -ne 0) {
            throw "docker compose up failed"
        }

        Section "Waiting for API health"
        $healthy = $false
        for ($i = 1; $i -le 30; $i++) {
            try {
                $health = Invoke-WebRequest "$ApiBaseUrl/health" -UseBasicParsing
                if ($health.StatusCode -eq 200) {
                    $healthy = $true
                    break
                }
            } catch { }
            Start-Sleep 1
        }

        if (-not $healthy) {
            docker compose -f $ComposeFile logs village-api
            throw "API did not become healthy."
        }
        Write-Host "[OK] /health = 200"

        Section "Readiness endpoint"
        $ready = Invoke-WebRequest "$ApiBaseUrl/ready" -UseBasicParsing
        if ($ready.StatusCode -ne 200) {
            throw "/ready expected 200 with database available, got $($ready.StatusCode)"
        }
        Write-Host "[OK] /ready = 200"

        Section "Dependency failure readiness"
        docker compose -f $ComposeFile stop postgres | Out-Null
        Start-Sleep 3

        try {
            Invoke-WebRequest "$ApiBaseUrl/ready" -UseBasicParsing | Out-Null
            throw "/ready unexpectedly returned 200 while PostgreSQL was stopped"
        } catch {
            if ($_.Exception.Response -and $_.Exception.Response.StatusCode.value__ -ne 503) {
                throw "/ready expected 503 with PostgreSQL stopped"
            }
        }

        docker compose -f $ComposeFile start postgres | Out-Null

        $readyAgain = $false
        for ($i = 1; $i -le 30; $i++) {
            try {
                $response = Invoke-WebRequest "$ApiBaseUrl/ready" -UseBasicParsing
                if ($response.StatusCode -eq 200) {
                    $readyAgain = $true
                    break
                }
            } catch { }
            Start-Sleep 1
        }

        if (-not $readyAgain) {
            docker compose -f $ComposeFile logs postgres village-api
            throw "/ready did not recover after PostgreSQL restart"
        }

        Write-Host "[OK] /ready = 503 during dependency outage and recovered to 200"

        Section "Metrics endpoint"
        $metrics = Invoke-WebRequest "$ApiBaseUrl/metrics" -UseBasicParsing
        if ($metrics.StatusCode -ne 200) {
            throw "/metrics expected 200, got $($metrics.StatusCode)"
        }
        Write-Host "[OK] /metrics = 200"

        $requiredMetrics = @(
            "village_http_requests_total",
            "village_http_request_duration_seconds",
            "village_http_requests_in_flight",
            "village_panics_total",
            "village_errors_total",
            "village_db_queries_total",
            "village_db_query_duration_seconds",
            "village_db_pool_",
            "village_build_info"
        )

        foreach ($metric in $requiredMetrics) {
            if ($metrics.Content -notmatch [regex]::Escape($metric)) {
                throw "/metrics missing $metric"
            }
        }

        Write-Host "[OK] Required metrics exposed"

        Section "Graceful shutdown"
        docker compose -f $ComposeFile stop village-api | Out-Null
        if ($LASTEXITCODE -ne 0) {
            throw "API graceful shutdown failed"
        }
        Write-Host "[OK] API container stopped cleanly"

        docker compose -f $ComposeFile start village-api | Out-Null
        Write-Host "[OK] API container restarted"
    }
    finally {
        if (-not $KeepDockerRunning) {
            docker compose -f $ComposeFile down --remove-orphans
        }
    }
}

Section "PRODUCTION VERIFICATION PASSED"
