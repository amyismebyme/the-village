[CmdletBinding()]
param(
    [string]$ApiImage = $(if ($env:VILLAGE_API_IMAGE) { $env:VILLAGE_API_IMAGE } else { 'village-api:local' }),
    [string]$DbPassword = $(if ($env:VILLAGE_DB_PASSWORD) { $env:VILLAGE_DB_PASSWORD } else { 'village' }),
    [switch]$EnableHPA
)

$ErrorActionPreference = 'Stop'
$repoRoot = Split-Path -Parent $PSScriptRoot
$k8sDir = Join-Path $repoRoot 'infra/kubernetes'

function Invoke-Kubectl {
    param([string[]]$Arguments)
    & kubectl @Arguments
    if ($LASTEXITCODE -ne 0) {
        throw "kubectl $($Arguments -join ' ') failed with exit code $LASTEXITCODE."
    }
}

Write-Host "Using API image: $ApiImage"
Write-Host 'Checking Kubernetes context...'
Invoke-Kubectl @('cluster-info')

Write-Host 'Creating namespace...'
Invoke-Kubectl @('apply', '-f', (Join-Path $k8sDir 'namespace.yaml'))

Write-Host 'Ensuring development PostgreSQL secret exists...'
$existingSecret = kubectl get secret village-postgres-secret -n village --ignore-not-found
if ($LASTEXITCODE -ne 0) {
    throw 'Failed to check for existing PostgreSQL secret.'
}

if ([string]::IsNullOrWhiteSpace(($existingSecret | Out-String))) {
    $secretYaml = @"
apiVersion: v1
kind: Secret
metadata:
  name: village-postgres-secret
  namespace: village
type: Opaque
stringData:
  POSTGRES_DB: village
  POSTGRES_USER: village
  POSTGRES_PASSWORD: $DbPassword
"@
    $secretYaml | kubectl apply -f - | Out-Host
    if ($LASTEXITCODE -ne 0) { throw 'Failed to create PostgreSQL secret.' }
} else {
    Write-Host 'PostgreSQL secret already exists; retaining it for database compatibility.'
}

Write-Host 'Applying API configuration...'
Invoke-Kubectl @('apply', '-f', (Join-Path $k8sDir 'configmap.yaml'))

Write-Host 'Starting PostgreSQL...'
Invoke-Kubectl @('apply', '-f', (Join-Path $k8sDir 'postgres.yaml'))
Invoke-Kubectl @('rollout', 'status', 'statefulset/village-postgres', '-n', 'village', '--timeout=180s')

Write-Host 'Creating migration ConfigMap from repository migrations...'
Invoke-Kubectl @('create', 'configmap', 'village-migrations', '-n', 'village', '--from-file="C:\Users\ARaina\the-village\migrations"', '--dry-run=client', '-o', 'yaml') | kubectl apply -f - | Out-Host
if ($LASTEXITCODE -ne 0) { throw 'Failed to create migration ConfigMap.' }

Write-Host 'Running database migrations...'
Invoke-Kubectl @('delete', 'job', 'village-migrate', '-n', 'village', '--ignore-not-found')
Invoke-Kubectl @('apply', '-f', (Join-Path $k8sDir 'migration-job.yaml'))
Invoke-Kubectl @('wait', '--for=condition=complete', 'job/village-migrate', '-n', 'village', '--timeout=180s')

Write-Host "Applying API deployment with image $ApiImage..."
$apiManifest = Get-Content (Join-Path $k8sDir 'api.yaml') -Raw
$apiManifest = $apiManifest -replace 'image: village-api:local', "image: $ApiImage"
$apiManifest | kubectl apply -f - | Out-Host
if ($LASTEXITCODE -ne 0) { throw 'Failed to apply API manifests.' }

Invoke-Kubectl @('apply', '-f', (Join-Path $k8sDir 'pdb.yaml'))
if ($EnableHPA) {
    Write-Host 'Applying HPA...'
    Invoke-Kubectl @('apply', '-f', (Join-Path $k8sDir 'hpa.yaml'))
} else {
    Write-Host 'HPA not applied. Pass -EnableHPA after verifying metrics-server is available.'
}

Write-Host 'Waiting for API rollout...'
Invoke-Kubectl @('rollout', 'status', 'deployment/village-api', '-n', 'village', '--timeout=180s')
Invoke-Kubectl @('get', 'pods', '-n', 'village', '-o', 'wide')
Invoke-Kubectl @('get', 'svc', '-n', 'village')
Write-Host 'Kubernetes deployment complete. API NodePort: http://localhost:30080'
