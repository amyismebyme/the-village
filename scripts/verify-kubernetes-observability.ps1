[CmdletBinding()]
param(
    [string]$Namespace = 'village',
    [string]$PrometheusUrl = 'http://localhost:30090',
    [switch]$SkipKubeletCheck
)

$ErrorActionPreference = 'Stop'

function Assert-True([bool]$Condition, [string]$Message) {
    if (-not $Condition) { throw $Message }
}

function Get-PrometheusTargets {
    $response = Invoke-RestMethod -Uri "$PrometheusUrl/api/v1/targets" -Method Get -TimeoutSec 10
    Assert-True ($response.status -eq 'success') 'Prometheus targets API did not return status=success.'
    return @($response.data.activeTargets)
}

function Assert-TargetUp([array]$Targets, [string]$Job, [string]$Reason = '') {
    $matches = @($Targets | Where-Object { $_.labels.job -eq $Job })
    Assert-True ($matches.Count -gt 0) "No active Prometheus target found for job '$Job'. $Reason"
    $up = @($matches | Where-Object { $_.health -eq 'up' })
    Assert-True ($up.Count -gt 0) "Prometheus target job '$Job' exists but has no healthy targets. First errors: $((@($matches | ForEach-Object { $_.lastError } | Where-Object { $_ }) -join ' | '))"
    Write-Host "$Job healthy: $($up.Count) target(s)."
}

kubectl get namespace $Namespace | Out-Null
kubectl get deployment village-api -n $Namespace | Out-Null
kubectl get deployment village-k8s-prometheus -n $Namespace | Out-Null
kubectl get deployment village-kube-state-metrics -n $Namespace | Out-Null
kubectl get daemonset village-node-exporter -n $Namespace | Out-Null

$promReady = Invoke-WebRequest -Uri "$PrometheusUrl/-/ready" -Method Get -UseBasicParsing -TimeoutSec 10
Assert-True ($promReady.StatusCode -eq 200) 'Kubernetes Prometheus /-/ready is not healthy.'

$targets = Get-PrometheusTargets
Assert-TargetUp $targets 'village-api' 'Check pod annotations, API /metrics, and Prometheus RBAC.'
Assert-TargetUp $targets 'kube-state-metrics'
Assert-TargetUp $targets 'node-exporter'

if (-not $SkipKubeletCheck) {
    Assert-TargetUp $targets 'kubelet' 'Check nodes/proxy RBAC and API-server proxy access.'
    Assert-TargetUp $targets 'kubelet-cadvisor' 'Check nodes/proxy RBAC and kubelet /metrics/cadvisor availability.'
}

function Assert-QueryHasData([string]$Query, [string]$Description) {
    $encoded = [Uri]::EscapeDataString($Query)
    $response = Invoke-RestMethod -Uri "$PrometheusUrl/api/v1/query?query=$encoded" -Method Get -TimeoutSec 10
    Assert-True ($response.status -eq 'success') "Prometheus query failed for $Description."
    Assert-True (@($response.data.result).Count -gt 0) "Prometheus query returned no data for ${Description}: $Query"
    Write-Host "$Description metric data present."
}

Assert-QueryHasData 'village_http_requests_total{namespace="village"}' 'Village API application metrics'
Assert-QueryHasData 'kube_deployment_status_replicas_available{namespace="village",deployment="village-api"}' 'Kubernetes deployment metrics'
Assert-QueryHasData 'node_uname_info' 'Node exporter metrics'

if (-not $SkipKubeletCheck) {
    Assert-QueryHasData 'container_memory_working_set_bytes{namespace="village",pod=~"village-api.*"}' 'Workload container memory metrics'
}

Write-Host 'Kubernetes observability verification passed.'
Write-Host "Prometheus UI: $PrometheusUrl"
Write-Host 'Grafana dashboards: Village / Kubernetes'
