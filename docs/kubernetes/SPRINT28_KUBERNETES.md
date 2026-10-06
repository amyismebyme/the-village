# Sprint 28 — Kubernetes: Real Deployment + Scaling + Recovery

## Purpose

Sprint 28 introduces a real Kubernetes deployment for The Village. The first target is a local Docker Desktop Kubernetes cluster. The existing Docker Compose stack remains available for local application + observability development.

## Deployed components

```text
namespace: village

PostgreSQL
  ├─ StatefulSet: village-postgres
  ├─ headless Service: village-postgres
  └─ 2Gi persistent volume claim

Database migration
  ├─ ConfigMap generated from ./migrations by the deployment script
  └─ Job: village-migrate

API
  ├─ Deployment: village-api
  ├─ default replicas: 2
  ├─ NodePort: 30080
  ├─ ServiceAccount with token automount disabled
  ├─ startup/liveness: /health
  ├─ readiness: /ready
  ├─ PodDisruptionBudget: minAvailable=1
  └─ optional HPA: 2–5 replicas, CPU target 70%
```

## Why deployment is scripted

Kubernetes does not provide a native `depends_on` equivalent for arbitrary Jobs. The local deployment script therefore creates resources in an explicit sequence:

```text
namespace
  ↓
DB Secret
  ↓
API ConfigMap
  ↓
PostgreSQL StatefulSet
  ↓
PostgreSQL readiness
  ↓
Migration ConfigMap
  ↓
Migration Job
  ↓
Migration completion
  ↓
API Deployment + Service
  ↓
API rollout
```

This prevents the deployment workflow from advertising the API before the schema migration completes.

## Local API image

The default image is:

```text
village-api:local
```

The deployment script does not build the image. Build it with the existing Docker workflow first. This intentionally preserves the current local Zscaler certificate handling.

Override the image with:

```powershell
.\scripts\k8s-deploy.ps1 -ApiImage village-api:some-tag
```

or:

```bash
./scripts/k8s-deploy.sh --api-image village-api:some-tag
```

## Deploy on Docker Desktop Kubernetes

Enable Kubernetes in Docker Desktop and check the context:

```powershell
kubectl config current-context
kubectl cluster-info
```

Build the local image using the existing project workflow, then:

```powershell
.\scripts\k8s-deploy.ps1
```

The default service is:

```text
http://localhost:30080
```

## Verify

```powershell
.\scripts\verify-kubernetes.ps1
```

The verification proves:

1. PostgreSQL becomes Ready.
2. Migration Job completes successfully.
3. API Deployment reaches Ready.
4. `/health` returns 200.
5. `/ready` returns 200.
6. API scales from 2 to 3 replicas.
7. One API pod can be deleted.
8. Kubernetes creates a replacement pod.
9. Three API replicas become available again.
10. `/health` and `/ready` recover.

## Manual scaling

PowerShell:

```powershell
.\scripts\k8s-scale.ps1 3
```

Bash:

```bash
./scripts/k8s-scale.sh 3
```

The scale scripts restrict local verification to 1–5 replicas. That keeps the development exercise bounded.

## HPA

The HPA manifest is included but not applied by default.

Local HPA requires a Metrics API implementation, normally metrics-server. Check:

```powershell
kubectl top pods -n village
```

If that command reports that Metrics API is unavailable, manual scaling is the verified path for this sprint.

Enable HPA only when metrics-server is available:

```powershell
.\scripts\k8s-deploy.ps1 -EnableHPA
```

or:

```bash
./scripts/k8s-deploy.sh --enable-hpa
```

Do not interpret the HPA as a production tuning decision yet. Sprint 26 performance results are the basis for later resource tuning.

## Resource configuration

Initial development values are:

```text
API request: 100m CPU / 128Mi memory
API limit:   500m CPU / 512Mi memory

PostgreSQL request: 100m CPU / 128Mi memory
PostgreSQL limit:   500m CPU / 512Mi memory
PostgreSQL storage: 2Gi
```

These values are intentionally conservative starting points. They should be revisited after sustained/spike load measurements and before any cloud production deployment.

## Recovery model

When a pod is deleted:

```text
API pod deleted
      ↓
Deployment controller
      ↓
new pod scheduled
      ↓
startup probe /health
      ↓
readiness probe /ready
      ↓
Service routes to Ready replicas
```

Because the API Deployment defaults to 2 replicas and the PDB requires one available pod during voluntary disruption, this is a basic high-availability/recovery exercise rather than a full multi-node production topology.

## PostgreSQL persistence

PostgreSQL is deployed as a StatefulSet with a persistent volume claim. Normal teardown removes workloads but leaves the PVC.

```powershell
.\scripts\k8s-down.ps1
```

To remove local PostgreSQL data too:

```powershell
.\scripts\k8s-down.ps1 -DeleteData
```

## Deferred to later sprints

- Helm
- GitOps / Argo CD
- production Ingress/TLS
- production Secret Manager / External Secrets
- managed Kubernetes / GKE
- multi-zone PostgreSQL
- production storage classes and backup strategy
- cluster-wide observability deployment
- service mesh

Those are intentionally not mixed into Sprint 28.
