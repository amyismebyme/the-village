param(
    [switch]$KeepRunning
)
$ErrorActionPreference = "Stop"
$RepoRoot = Resolve-Path "$PSScriptRoot\.."
$ComposeFile = Join-Path $RepoRoot "testdata\docker-compose.integration.yml"
$MigrationDir = Join-Path $RepoRoot "migrations"
$ApiDir = Join-Path $RepoRoot "apps\api"
$DatabaseURL = "postgres://village:village@localhost:5433/village?sslmode=disable"
$ContainerName = "village-postgres-integration"
function Write-Section($title)
{
    Write-Host ""
    Write-Host "============================================================" -ForegroundColor Cyan
    Write-Host " $title" -ForegroundColor Cyan
    Write-Host "============================================================" -ForegroundColor Cyan
}
function Check-Command($name)
{
    if (-not (Get-Command $name -ErrorAction SilentlyContinue))
    {
        throw "$name is not installed or is not available in PATH."
    }
    Write-Host "[OK] $name found"
}
function Check-File($path)
{
    if (!(Test-Path $path))
    {
        throw "Required file not found: $path"
    }
    Write-Host "[OK] $path"
}
function Wait-ForDatabase {
    Write-Host ""
    Write-Host "Waiting for PostgreSQL..."
    for ($i = 1; $i -le 30; $i++) {
        $status = docker inspect `
            $ContainerName `
            --format "{{.State.Health.Status}}" `
            2>$null
        if ($status -eq "healthy") {
            Write-Host "[OK] PostgreSQL healthy"
            return
        }
        Start-Sleep 1
    }
    Write-Host ""
    Write-Host "Container logs:"
    docker logs $ContainerName
    throw "Database never became healthy."
}
Write-Section "Village Integration Test Runner"
try
{
    #############################################################
    # Environment Validation
    #############################################################
    Write-Section "Checking Environment"
    Check-Command docker
    Check-Command go
    Check-Command migrate
    Write-Host ""
    try
    {
        docker info *> $null
        Write-Host "[OK] Docker daemon running"
    }
    catch
    {
        throw "Docker Desktop is not running."
    }
    Check-File $ComposeFile
    Check-File $MigrationDir
    Check-File (Join-Path $ApiDir "go.mod")
    #############################################################
    # Cleanup old environment
    #############################################################
    Write-Section "Removing Previous Environment"
    docker compose `
        -f $ComposeFile `
        down `
        -v `
        --remove-orphans
    #############################################################
    # Start Database
    #############################################################
    Write-Section "Starting PostgreSQL"
    docker compose `
        -f $ComposeFile `
        up -d
    if ($LASTEXITCODE -ne 0)
    {
        throw "docker compose up failed."
    }
    Wait-ForDatabase
    #############################################################
    # Apply Migrations
    #############################################################
    Write-Section "Applying Database Migrations"
    Push-Location $RepoRoot
    migrate `
        -verbose `
        -path "./migrations" `
        -database $DatabaseURL `
        up
    if ($LASTEXITCODE -ne 0)
    {
        Pop-Location
        throw "Migration failed."
    }
    Pop-Location
    #############################################################
    # Run Tests
    #############################################################
    Write-Section "Running Integration Tests"
    Push-Location $ApiDir
    go test `
        -tags=integration `
        ./internal/integration/... `
        -count=1 `
        -v
    $testExit = $LASTEXITCODE
    Pop-Location
    if ($testExit -ne 0)
    {
        throw "Integration tests failed."
    }
    #############################################################
    # Success
    #############################################################
    Write-Section "SUCCESS"
    Write-Host "All integration tests passed." -ForegroundColor Green
}
finally
{
    if (-not $KeepRunning)
    {
        Write-Section "Cleaning Up"
        docker compose `
            -f $ComposeFile `
            down `
            -v `
            --remove-orphans
    }
    else
    {
        Write-Host ""
        Write-Host "Database left running (-KeepRunning)." -ForegroundColor Yellow
    }
}
