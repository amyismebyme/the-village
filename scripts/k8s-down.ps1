[CmdletBinding()]
param(
    [switch]$DeleteData
)

$ErrorActionPreference = 'Stop'

$resources = @(
    'deployment/village-api',
    'service/village-api',
    'hpa/village-api',
    'pdb/village-api',
    'job/village-migrate',
    'statefulset/village-postgres',
    'service/village-postgres',
    'configmap/village-api-config',
    'configmap/village-migrations',
    'secret/village-postgres-secret'
)

foreach ($resource in $resources) {
    kubectl delete $resource -n village --ignore-not-found | Out-Host
    if ($LASTEXITCODE -ne 0) { throw "Failed to delete $resource." }
}

if ($DeleteData) {
    kubectl delete pvc -n village -l app.kubernetes.io/name=village-postgres --ignore-not-found | Out-Host
}

Write-Host 'Village Kubernetes workloads removed.'
if ($DeleteData) { Write-Host 'PostgreSQL PVC data removed.' }
