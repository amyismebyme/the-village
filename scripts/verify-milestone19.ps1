param(
    [switch]$SkipDocker,
    [switch]$SkipIntegration
)

$ErrorActionPreference = "Stop"
$root = Resolve-Path (Join-Path $PSScriptRoot "..")
$api = Join-Path $root "apps\api"

function Invoke-Step {
    param(
        [string]$Name,
        [scriptblock]$Action
    )

    Write-Host ""
    Write-Host "== $Name ==" -ForegroundColor Cyan
    & $Action
    if ($LASTEXITCODE -ne 0) {
        throw "$Name failed with exit code $LASTEXITCODE"
    }
}

Push-Location $api
try {
    Invoke-Step "Formatting" {
        $files = gofmt -l .
        if ($files) {
            $files
            exit 1
        }
    }

    Invoke-Step "Module verification" {
        go mod verify
    }

    Invoke-Step "Vulnerability scan" {
        go run golang.org/x/vuln/cmd/govulncheck@v1.7.0 ./...
    }

    Invoke-Step "Vet" {
        go vet ./...
    }

    Invoke-Step "Lint" {
        golangci-lint run
    }

    Invoke-Step "Unit tests" {
        go test ./... -count=1
    }

    Invoke-Step "Race tests" {
        go test -race ./... -count=1
    }

    Invoke-Step "Build" {
        go build ./...
    }

    Invoke-Step "Ingestion performance sanity" {
        go test ./internal/external/reddit -run '^$' `
            -bench 'Benchmark(NormalizeRedditPost|DeduplicateExternalItems)$' `
            -benchmem -benchtime=250ms -count=1
    }

    if (-not $SkipIntegration) {
        Invoke-Step "PostgreSQL integration tests" {
            Push-Location $root
            try {
                make test-integration
            }
            finally {
                Pop-Location
            }
        }
    }
}
finally {
    Pop-Location
}

if (-not $SkipDocker) {
    if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
        Write-Host "Docker unavailable; skipping Docker runtime checks." -ForegroundColor Yellow
    }
    elseif (-not (Test-Path (Join-Path $root "zscaler.crt"))) {
        Write-Host "zscaler.crt is absent; current Dockerfile requires it. Docker checks skipped." -ForegroundColor Yellow
    }
    else {
        Push-Location $root
        try {
            Invoke-Step "Docker build and runtime" {
                docker compose up -d --build
            }

            try {
                $healthy = $false
                for ($i = 0; $i -lt 45; $i++) {
                    $status = docker inspect --format='{{.State.Health.Status}}' village-api 2>$null
                    if ($status -eq "healthy") {
                        $healthy = $true
                        break
                    }
                    Start-Sleep -Seconds 1
                }

                if (-not $healthy) {
                    docker compose logs village-api
                    throw "village-api did not become healthy"
                }

                (Invoke-WebRequest "http://127.0.0.1:8080/health" -UseBasicParsing).StatusCode | Out-Null
                (Invoke-WebRequest "http://127.0.0.1:8080/ready" -UseBasicParsing).StatusCode | Out-Null
                (Invoke-WebRequest "http://127.0.0.1:8080/metrics" -UseBasicParsing).StatusCode | Out-Null
            }
            finally {
                docker compose down --remove-orphans | Out-Null
            }
        }
        finally {
            Pop-Location
        }
    }
}

Write-Host ""
Write-Host "MILESTONE 8 VERIFICATION PASSED" -ForegroundColor Green
