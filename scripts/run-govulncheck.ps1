$ErrorActionPreference = "Stop"

Push-Location (Join-Path $PSScriptRoot "..\apps\api")
try {
    go run golang.org/x/vuln/cmd/govulncheck@v1.7.0 ./...
    if ($LASTEXITCODE -ne 0) {
        throw "govulncheck failed with exit code $LASTEXITCODE"
    }
}
finally {
    Pop-Location
}
