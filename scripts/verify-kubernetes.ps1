[CmdletBinding()]
param(
    [int]$RecoveryTimeoutSeconds = 120,
    [string]$ApiUrl = 'http://localhost:30080'
)

$ErrorActionPreference = 'Stop'

function Invoke-Kubectl {
    param([string[]]$Arguments)
    & kubectl @Arguments
    if ($LASTEXITCODE -ne 0) {
        throw "kubectl $($Arguments -join ' ') failed."
    }
}

function Get-AvailableReplicas {
    $deployment = kubectl get deployment village-api -n village -o json | ConvertFrom-Json
    if ($null -eq $deployment.status.availableReplicas) { return 0 }
    return [int]$deployment.status.availableReplicas
}

function Wait-HttpOk {
    param([string]$Path, [int]$TimeoutSeconds)

    $deadline = (Get-Date).AddSeconds($TimeoutSeconds)
    do {
        try {
            $response = Invoke-WebRequest -Uri ($ApiUrl.TrimEnd('/') + $Path) -UseBasicParsing -TimeoutSec 5
            if ($response.StatusCode -eq 200) { return }
        } catch { }
        Start-Sleep -Seconds 2
    } while ((Get-Date) -lt $deadline)

    throw "$Path did not return HTTP 200 within $TimeoutSeconds seconds."
}

Write-Host '1/5 Cluster and database verification'
Invoke-Kubectl @('cluster-info')
Invoke-Kubectl @('rollout', 'status', 'statefulset/village-postgres', '-n', 'village', "--timeout=${RecoveryTimeoutSeconds}s")

$job = kubectl get job village-migrate -n village -o json | ConvertFrom-Json
if ($job.status.succeeded -lt 1) {
    throw 'village-migrate has not completed successfully.'
}

Write-Host '2/5 API rollout and probes'
Invoke-Kubectl @('rollout', 'status', 'deployment/village-api', '-n', 'village', "--timeout=${RecoveryTimeoutSeconds}s")
Wait-HttpOk '/health' $RecoveryTimeoutSeconds
Wait-HttpOk '/ready' $RecoveryTimeoutSeconds

Write-Host '3/5 Scaling from 2 to 3 replicas'
Invoke-Kubectl @('scale', 'deployment/village-api', '-n', 'village', '--replicas=3')
Invoke-Kubectl @('rollout', 'status', 'deployment/village-api', '-n', 'village', "--timeout=${RecoveryTimeoutSeconds}s")
if ((Get-AvailableReplicas) -lt 3) {
    throw "Expected 3 available API replicas after scaling, got $(Get-AvailableReplicas)."
}

Write-Host '4/5 Pod recovery'
$pod = kubectl get pods -n village -l app.kubernetes.io/name=village-api -o jsonpath='{.items[0].metadata.name}'
if ([string]::IsNullOrWhiteSpace($pod)) {
    throw 'No API pod found.'
}
Invoke-Kubectl @('delete', 'pod', $pod, '-n', 'village', '--wait=false')
Wait-HttpOk '/health' $RecoveryTimeoutSeconds
Wait-HttpOk '/ready' $RecoveryTimeoutSeconds

$deadline = (Get-Date).AddSeconds($RecoveryTimeoutSeconds)
do {
    if ((Get-AvailableReplicas) -ge 3) { break }
    Start-Sleep -Seconds 2
} while ((Get-Date) -lt $deadline)

if ((Get-AvailableReplicas) -lt 3) {
    throw "Expected 3 available API replicas after pod recovery, got $(Get-AvailableReplicas)."
}

Write-Host '5/5 HPA capability check'
kubectl top pods -n village | Out-Host
if ($LASTEXITCODE -eq 0) {
    Write-Host 'metrics-server is available. HPA can be enabled.'
    $hpaMetricsAvailable = $true
} else {
    Write-Host 'Metrics API is unavailable; manual scaling remains the verified scaling path.'
    Write-Host 'HPA behavior has not been verified. Enable metrics-server before enabling HPA.'
    $hpaMetricsAvailable = $false
}

Invoke-Kubectl @('get', 'deployment', '-n', 'village')
Invoke-Kubectl @('get', 'pods', '-n', 'village', '-o', 'wide')
Invoke-Kubectl @('get', 'svc', '-n', 'village')

Write-Host 'Sprint 28 Kubernetes deployment, scaling, probes, migration ordering, and pod recovery verification passed.'
if (-not $hpaMetricsAvailable) {
    Write-Host 'NOTE: HPA was not validated because the Kubernetes Metrics API is unavailable.'
}
