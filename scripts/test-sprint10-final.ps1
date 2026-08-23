$ErrorActionPreference = "Stop"
$ApiDir = Resolve-Path "$PSScriptRoot\..\apps\api"

Write-Host "SPRINT 10 FINAL PRODUCTION VERIFICATION" -ForegroundColor Cyan

Push-Location $ApiDir
try {
    go mod tidy
    if ($LASTEXITCODE -ne 0) { throw "go mod tidy failed" }

    $formatted = gofmt -l .
    if ($formatted) {
        Write-Host "Unformatted files:" -ForegroundColor Red
        $formatted
        throw "gofmt verification failed"
    }

    go build ./...
    if ($LASTEXITCODE -ne 0) { throw "go build failed" }

    go test ./... -count=1
    if ($LASTEXITCODE -ne 0) { throw "unit tests failed" }

    go test -race ./...
    if ($LASTEXITCODE -ne 0) { throw "race detector failed" }

    go vet ./...
    if ($LASTEXITCODE -ne 0) { throw "go vet failed" }

    golangci-lint run
    if ($LASTEXITCODE -ne 0) { throw "lint failed" }

    go test -tags=integration ./internal/integration/... -v -count=1
    if ($LASTEXITCODE -ne 0) { throw "integration tests failed" }
}
finally {
    Pop-Location
}

Write-Host "SPRINT 10 FINAL VERIFICATION PASSED" -ForegroundColor Green
