param(
    [switch]$KeepRunning
)

$ErrorActionPreference = "Stop"


$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$RepoRoot  = Resolve-Path (Join-Path $ScriptDir "..")

$ApiDir = Join-Path $RepoRoot "apps\api"
$ComposeFile = Join-Path $RepoRoot "testdata\docker-compose.integration.yml"
$MigrationDir = Join-Path $RepoRoot "migrations"


Write-Host ""
Write-Host "========================================="
Write-Host " Village Integration Test Runner"
Write-Host "========================================="
Write-Host ""

Write-Host "Starting PostgreSQL..."

docker compose `
    -f $ComposeFile  `
    up -d

if ($LASTEXITCODE -ne 0) {
    throw "Failed to start docker compose."
}

Write-Host ""
Write-Host "Waiting for PostgreSQL..."

$healthy = $false

for ($i = 0; $i -lt 30; $i++) {

    $status = docker inspect `
        village-postgres-integration `
        --format="{{.State.Health.Status}}"

    if ($status -eq "healthy") {
        $healthy = $true
        break
    }

    Start-Sleep -Seconds 1
}

if (-not $healthy) {

    docker logs village-postgres-integration

    throw "Database never became healthy."
}

Write-Host ""
Write-Host "Applying migrations..."

$MigrationDir = ($MigrationDir -replace '\\','/')

migrate `
    -path $MigrationDir `
    -database "postgres://village:village@localhost:5433/village?sslmode=disable" `
    up

if ($LASTEXITCODE -ne 0) {

    Write-Host "Migration failed."

    exit 1
}

Write-Host ""
Write-Host "Running integration tests..."

go test `
    -tags=integration `
    ./internal/integration/... `
    -v

$testExitCode = $LASTEXITCODE

if (-not $KeepRunning) {

    Write-Host ""
    Write-Host "Stopping integration database..."

    docker compose `
        -f $ComposeFile  `
        down -v
}

exit $testExitCode