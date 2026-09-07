param(
    [string]$Subreddit = "toronto"
)

$ErrorActionPreference = "Stop"

$required = @(
    "REDDIT_CLIENT_ID",
    "REDDIT_CLIENT_SECRET",
    "REDDIT_USER_AGENT"
)

foreach ($name in $required) {
    $value = (Get-Item "Env:$name" -ErrorAction SilentlyContinue).Value
    if ([string]::IsNullOrWhiteSpace($value)) {
        throw "$name must be set before running the live Reddit smoke test"
    }
}

$env:REDDIT_INGEST_SUBREDDIT = $Subreddit

Push-Location (Join-Path $PSScriptRoot "..\apps\api")
try {
    go test -tags=smoke ./internal/external/reddit `
        -run '^TestRedditLiveSmoke$' `
        -count=1 `
        -v

    if ($LASTEXITCODE -ne 0) {
        throw "Reddit live smoke test failed with exit code $LASTEXITCODE"
    }
}
finally {
    Pop-Location
}
