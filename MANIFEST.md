# Sprint 28 — Kubernetes: Real Deployment + Scaling + Recovery

This is an incremental patch intended for the repository after Sprint 27. It adds only the Kubernetes layer and operator scripts; it does not change application Go code or Zscaler handling.

## Added

`infra/kubernetes/`
- `namespace.yaml`
- `configmap.yaml`
- `postgres.yaml`
- `migration-job.yaml`
- `api.yaml`
- `pdb.yaml`
- `hpa.yaml`
- `README.md`

`scripts/`
- `k8s-deploy.ps1`
- `k8s-deploy.sh`
- `k8s-scale.ps1`
- `k8s-scale.sh`
- `verify-kubernetes.ps1`
- `verify-kubernetes.sh`
- `k8s-down.ps1`
- `k8s-down.sh`

`docs/kubernetes/`
- `SPRINT28_KUBERNETES.md`
- `SPRINT28_CHECKLIST.md`

## Architecture

```text
Docker Desktop Kubernetes
        │
        ├── PostgreSQL StatefulSet
        │       └── 2Gi PVC
        │
        ├── migration Job
        │
        └── village-api Deployment
                ├── 2 replicas by default
                ├── NodePort 30080
                ├── startup/liveness -> /health
                ├── readiness -> /ready
                ├── PDB minAvailable=1
                └── optional HPA 2-5 replicas @ 70% CPU
```

## Deployment ordering

The scripts intentionally perform:

```text
Postgres -> ready -> migration Job -> complete -> API Deployment -> rollout
```

The migration image `migrate/migrate:v4.18.3` is Alpine-based and includes `/bin/sh`, so the migration Job can construct its connection string from Kubernetes Secret environment variables. citeturn495942view0

## Secret handling

No PostgreSQL Secret YAML is committed. The deployment scripts create `village-postgres-secret` only when it does not already exist. This avoids accidentally changing a running database's password when the deployment script is run again.

The default local development password is `village`; override the initial value with `VILLAGE_DB_PASSWORD` / `-DbPassword`.

## Run locally

1. Enable Kubernetes in Docker Desktop.
2. Build `village-api:local` with the existing Docker workflow.
3. Deploy:

```powershell
.\scripts\k8s-deploy.ps1
```

4. Verify:

```powershell
.\scripts\verify-kubernetes.ps1
```

5. API:

```text
http://localhost:30080
```

## Optional HPA

Check metrics-server:

```powershell
kubectl top pods -n village
```

Then:

```powershell
.\scripts\k8s-deploy.ps1 -EnableHPA
```

Manual scaling is still the deterministic Sprint 28 scaling path:

```powershell
.\scripts\k8s-scale.ps1 3
```

## Recovery

The verification script scales the API to three replicas, deletes one API pod, waits for `/health` and `/ready`, and verifies that three available replicas return.

## Teardown

Keep PostgreSQL data:

```powershell
.\scripts\k8s-down.ps1
```

Delete local PostgreSQL data too:

```powershell
.\scripts\k8s-down.ps1 -DeleteData
```

## Validation status

The generated YAML parsed successfully and all Bash scripts passed `bash -n`. Actual `kubectl` and Docker Desktop Kubernetes execution was not performed in the generation environment because `kubectl` is unavailable there.

## Deliberately deferred

Helm, GitOps/Argo CD, production ingress/TLS, external secret management, managed Kubernetes, multi-zone PostgreSQL, cluster-wide observability deployment, and cloud storage/backup policy remain later Sprint 29+ work.
