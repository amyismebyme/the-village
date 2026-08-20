#############################################################
# static/code-quality regression gate — formatting, unit tests, race, vet, lint, integration tests.
$ErrorActionPreference = "Stop"

Write-Host "== Formatting ==" -ForegroundColor Cyan
gofmt -l .
if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}

$unformatted = gofmt -l .
if ($unformatted) {
    Write-Host "Unformatted files:" -ForegroundColor Red
    $unformatted
    exit 1
}

Write-Host "== Validation package ==" -ForegroundColor Cyan
go test ./internal/validation/... -v
if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}

Write-Host "== Model validation ==" -ForegroundColor Cyan
go test ./internal/model/... -v
if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}

Write-Host "== Service validation ==" -ForegroundColor Cyan
go test ./internal/service/... -v
if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}

Write-Host "== Handler validation ==" -ForegroundColor Cyan
go test ./internal/handlers/... -v
if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}

Write-Host "== Full unit suite ==" -ForegroundColor Cyan
go test ./... -count=1
if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}

Write-Host "== Race detector ==" -ForegroundColor Cyan
go test -race ./...
if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}

Write-Host "== Vet ==" -ForegroundColor Cyan
go vet ./...
if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}

Write-Host "== Lint ==" -ForegroundColor Cyan
golangci-lint run
if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}

Write-Host "== Integration suite ==" -ForegroundColor Cyan
go test -tags=integration ./internal/integration/... -count=1
if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}

Write-Host ""
Write-Host "Validation regression gate PASSED." -ForegroundColor Green