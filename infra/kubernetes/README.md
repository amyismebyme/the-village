# Kubernetes

Sprint 28 local Kubernetes manifests for The Village.

Use `scripts/k8s-deploy.ps1` or `scripts/k8s-deploy.sh` rather than applying all files blindly because PostgreSQL readiness and the migration Job must complete before the API deployment is introduced.

Default local API endpoint:

`http://localhost:30080`

The deployment is designed for Docker Desktop Kubernetes and keeps production-grade concerns such as Helm, GitOps, managed secrets, and cloud networking for later sprints.
