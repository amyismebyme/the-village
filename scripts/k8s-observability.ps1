[CmdletBinding()]
param(
    [string]$Namespace = 'village'
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

Write-Host "Deploying Kubernetes observability into namespace '$Namespace'..."
Invoke-Kubectl @('cluster-info')
Invoke-Kubectl @('get', 'namespace', $Namespace)

Write-Host 'Applying Prometheus and discovery RBAC...'
Invoke-Kubectl @('apply', '-f', (Join-Path $k8sDir 'prometheus-rbac.yaml'))

Write-Host 'Applying kube-state-metrics...'
Invoke-Kubectl @('apply', '-f', (Join-Path $k8sDir 'kube-state-metrics.yaml'))
Invoke-Kubectl @('rollout', 'status', 'deployment/village-kube-state-metrics', '-n', $Namespace, '--timeout=180s')

Write-Host 'Applying node-exporter...'
Invoke-Kubectl @('apply', '-f', (Join-Path $k8sDir 'node-exporter.yaml'))
Invoke-Kubectl @('rollout', 'status', 'daemonset/village-node-exporter', '-n', $Namespace, '--timeout=180s')

Write-Host 'Applying Kubernetes Prometheus...'
Invoke-Kubectl @('apply', '-f', (Join-Path $k8sDir 'prometheus-k8s.yaml'))
Invoke-Kubectl @('rollout', 'status', 'deployment/village-k8s-prometheus', '-n', $Namespace, '--timeout=180s')

Write-Host 'Checking observability pods...'
Invoke-Kubectl @('get', 'pods', '-n', $Namespace, '-o', 'wide')
Invoke-Kubectl @('get', 'svc', '-n', $Namespace)

Write-Host 'Kubernetes observability deployment complete.'
Write-Host 'Prometheus: http://localhost:30090'
Write-Host 'Grafana:    http://localhost:3000'
