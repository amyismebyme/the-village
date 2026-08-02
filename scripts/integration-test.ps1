param(
    [switch]$KeepRunning
)

$ErrorActionPreference = "Stop"

$repoRoot = Resolve-Path "$PSScriptRoot\.."
$composeFile = Join-Path $repoRoot "testdata\docker-compose.integration.yml"
$migrations = Join-Path $repoRoot "migrations"
$apiDir = Join-Path $repoRoot "apps\api"

$databaseUrl = "postgres://village:village@localhost:5433/village?sslmode=disable"

Write-Host ""
Write-Host "======================================="
Write-Host " Village Integration Test Runner"
Write-Host "======================================="
Write-Host ""

try {

    ###############################################
    # Clean previous environment
    ###############################################

    Write-Host "Removing previous containers..."

    docker compose `
        -f $composeFile `
        down -v --remove-orphans

    ###############################################
    # Start PostgreSQL
    ###############################################

    Write-Host ""
    Write-Host "Starting PostgreSQL..."

    docker compose `
        -f $composeFile `
        up -d

    if ($LASTEXITCODE -ne 0) {
        throw "Failed to start docker compose."
    }

    ###############################################
    # Wait until healthy
    ###############################################

    Write-Host ""
    Write-Host "Waiting for PostgreSQL..."

    $healthy = $false

    for ($i = 1; $i -le 30; $i++) {

        $status = docker inspect `
            village-postgres-integration `
            --format "{{.State.Health.Status}}" `
            2>$null

        if ($status -eq "healthy") {
            $healthy = $true
            break
        }

        Start-Sleep 1
    }

    if (-not $healthy) {

        docker logs village-postgres-integration

        throw "PostgreSQL never became healthy."
    }

    Write-Host "PostgreSQL is healthy."

    ###############################################
    # Apply migrations
    ###############################################

    Write-Host ""
    Write-Host "Applying migrations..."

    Push-Location $repoRoot

    migrate `
        -verbose `
        -path "./migrations" `
        -database $databaseUrl `
        up

    if ($LASTEXITCODE -ne 0) {

        Pop-Location

        throw "Migration failed."
    }

    Pop-Location

    ###############################################
    # Run tests
    ###############################################

    Write-Host ""
    Write-Host "Running integration tests..."

    Push-Location $apiDir

    go test `
        -tags=integration `
        ./internal/integration/... `
        -v

    $testExitCode = $LASTEXITCODE

    Pop-Location

    if ($testExitCode -ne 0) {

        throw "Integration tests failed."
    }

    Write-Host ""
    Write-Host "======================================="
    Write-Host " ALL INTEGRATION TESTS PASSED"
    Write-Host "======================================="
}
finally {

    if (-not $KeepRunning) {

        Write-Host ""
        Write-Host "Stopping integration environment..."

        docker compose `
            -f $composeFile `
            down -v --remove-orphans
    }
    else {

        Write-Host ""
        Write-Host "Database left running (-KeepRunning)"
    }
}