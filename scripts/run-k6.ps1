[CmdletBinding()]
param(
    [Parameter(Mandatory = $true, Position = 0)]
    [ValidateSet('smoke', 'baseline', 'sustained', 'spike')]
    [string]$Scenario,

    [string]$BaseUrl = $(if ($env:K6_BASE_URL) { $env:K6_BASE_URL } else { 'http://host.docker.internal:8080' })
)

$ErrorActionPreference = 'Stop'
$repoRoot = Split-Path -Parent $PSScriptRoot
$resultDir = Join-Path $repoRoot 'artifacts/k6'
$scriptPath = "/loadtests/$Scenario.js"
$summaryPath = "/results/$Scenario-summary.json"

New-Item -ItemType Directory -Force -Path $resultDir | Out-Null

Write-Host "Running k6 scenario '$Scenario' against $BaseUrl"

docker run --rm -i `
    --add-host=host.docker.internal:host-gateway `
    -v "$(Join-Path $repoRoot 'loadtests'):/loadtests:ro" `
    -v "${resultDir}:/results" `
    -e "K6_BASE_URL=$BaseUrl" `
    grafana/k6:2.3.0 `
    run `
    --summary-export=$summaryPath `
    $scriptPath

if ($LASTEXITCODE -ne 0) {
    throw "k6 scenario '$Scenario' failed."
}

Write-Host "k6 scenario '$Scenario' passed. Summary: $resultDir/$Scenario-summary.json"
