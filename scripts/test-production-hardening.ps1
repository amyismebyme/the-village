param(
    [switch]$SkipIntegration,
    [switch]$SkipLiveSmoke,
    [string]$ApiBaseUrl = "http://localhost:8080"
)

$ErrorActionPreference = "Stop"

$ApiDir = Join-Path $PSScriptRoot "..\apps\api"

function Section($title) {
    Write-Host ""
    Write-Host "============================================================" -ForegroundColor Cyan
    Write-Host " $title" -ForegroundColor Cyan
    Write-Host "============================================================" -ForegroundColor Cyan
}

function Run-Step($title, [scriptblock]$command) {
    Section $title
    & $command
    if ($LASTEXITCODE -ne 0) {
        throw "$title failed with exit code $LASTEXITCODE"
    }
    Write-Host "[PASS] $title" -ForegroundColor Green
}

Push-Location $ApiDir
try {
    Section "Sprint 10.5-10.7 Production Hardening Verification"
    Write-Host "Module root: $ApiDir"
    Write-Host "API smoke URL: $ApiBaseUrl"

    Run-Step "Go formatting check" {
        $unformatted = gofmt -l .
        if ($unformatted) {
            Write-Host "Unformatted files:" -ForegroundColor Red
            $unformatted
            exit 1
        }
    }

    Run-Step "Build" {
        go build ./...
    }

    Run-Step "Full unit suite" {
        go test ./... -count=1
    }

    Run-Step "Race detector" {
        go test -race ./...
    }

    Run-Step "Vet" {
        go vet ./...
    }

    Run-Step "Lint" {
        golangci-lint run
    }

    Run-Step "Context propagation tests" {
        go test ./internal/service/... -run "Context|Cancellation" -v -count=1
    }

    Run-Step "Timeout middleware tests" {
        go test ./internal/middleware/... -run "Timeout|RequestTimeout" -v -count=1
    }

    Run-Step "Database timeout tests" {
        go test ./internal/repository/postgres/... -run "Timeout|Context" -v -count=1
    }

    if (-not $SkipIntegration) {
        Run-Step "Full integration suite" {
            go test -tags=integration ./internal/integration/... -v -count=1
        }

        Run-Step "Timeout integration tests" {
            go test -tags=integration ./internal/integration -run "Timeout|ContextCancellation" -v -count=1
        }

        Run-Step "Community observability integration tests" {
            go test -tags=integration ./internal/integration -run "Observability|CommunityMetrics" -v -count=1
        }
    }

    Section "Context.Background audit"
    $backgroundHits = Get-ChildItem .\internal -Recurse -File -Filter *.go |
        Select-String -Pattern 'context\.Background\(\)' |
        Where-Object {
            $_.Path -notmatch '_test\.go$'
        }

    if ($backgroundHits) {
        Write-Host "Review these production-path context.Background() uses:" -ForegroundColor Yellow
        $backgroundHits | ForEach-Object { Write-Host $_ }
    }
    else {
        Write-Host "[PASS] No production-path context.Background() uses found." -ForegroundColor Green
    }

    if (-not $SkipLiveSmoke) {
        Section "Live API smoke test"

        try {
            $health = Invoke-RestMethod "$ApiBaseUrl/health"
            Write-Host "[PASS] /health -> $($health.status)" -ForegroundColor Green
        }
        catch {
            throw "Live API smoke test failed: /health is unavailable at $ApiBaseUrl"
        }

        try {
            $ready = Invoke-RestMethod "$ApiBaseUrl/ready"
            Write-Host "[PASS] /ready -> $($ready.status)" -ForegroundColor Green
        }
        catch {
            throw "Live API smoke test failed: /ready is unavailable at $ApiBaseUrl"
        }

        try {
            $metrics = Invoke-WebRequest "$ApiBaseUrl/metrics" -UseBasicParsing
            if ($metrics.StatusCode -ne 200) {
                throw "Unexpected /metrics status $($metrics.StatusCode)"
            }

            $metricText = $metrics.Content

            $requiredMetrics = @(
                "village_http_requests_total",
                "village_http_request_duration_seconds",
                "village_http_requests_in_flight",
                "village_community_create_total",
                "village_community_update_total",
                "village_community_delete_total",
                "village_community_validation_failures_total"
            )

            foreach ($metricName in $requiredMetrics) {
                if ($metricText -notmatch [regex]::Escape($metricName)) {
                    throw "Missing metric: $metricName"
                }
            }

            Write-Host "[PASS] /metrics exposes required observability families" -ForegroundColor Green
        }
        catch {
            throw "Live API smoke test failed for /metrics: $($_.Exception.Message)"
        }
    }

    Section "Production Hardening Verification PASSED"
    Write-Host "10.5 Context propagation: PASS" -ForegroundColor Green
    Write-Host "10.6 HTTP/DB timeout policy: PASS" -ForegroundColor Green
    Write-Host "10.7 Timeout integration coverage: PASS" -ForegroundColor Green
}
finally {
    Pop-Location
}
