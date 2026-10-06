[CmdletBinding()]
param(
    [ValidateSet(1,2,3,4,5)]
    [int]$Replicas = 3,
    [int]$TimeoutSeconds = 120
)

$ErrorActionPreference = 'Stop'

& kubectl scale deployment/village-api -n village --replicas=$Replicas | Out-Host
if ($LASTEXITCODE -ne 0) { throw 'kubectl scale failed.' }

& kubectl rollout status deployment/village-api -n village --timeout="${TimeoutSeconds}s" | Out-Host
if ($LASTEXITCODE -ne 0) { throw 'Deployment did not converge after scaling.' }

& kubectl get deployment village-api -n village | Out-Host
& kubectl get pods -n village -l app.kubernetes.io/name=village-api -o wide | Out-Host

Write-Host "Scaling verification complete. Test endpoint: http://localhost:30080/health"
