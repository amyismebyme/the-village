# Sprint 27 — Chaos: Controlled Failure + Recovery

This patch targets a The Village repository that already contains Sprint 26 k6 changes.

## Added

- `scripts/run-chaos.ps1` — Windows/PowerShell chaos runner.
- `scripts/run-chaos.sh` — shell chaos runner.
- `docs/chaos/CHAOS_RUNBOOK.md` — experiment procedure, expected behavior, safety and evidence guidance.
- `artifacts/chaos/.gitkeep` — keeps the evidence directory in Git.

## Modified

- `Makefile` — adds `chaos-api-kill`, `chaos-postgres-outage`, and `chaos-tempo-outage` targets.
- `.gitignore` — ignores generated chaos evidence while retaining `.gitkeep`.

The Makefile and `.gitignore` included here are complete replacement files based on the Sprint 26 state on `main`.

## Experiments

1. `api-kill`
   - Sends SIGKILL to `village-api`.
   - Verifies Docker restart recovery.
   - Verifies `/health`, `/ready`, then k6 smoke.

2. `postgres-outage`
   - Stops PostgreSQL.
   - Verifies `/health` remains 200.
   - Verifies `/ready` becomes 503.
   - Restarts PostgreSQL.
   - Verifies `/ready` returns 200 and k6 smoke passes.

3. `tempo-outage`
   - Stops Tempo.
   - Verifies `/health`, `/ready`, and community listing remain available.
   - Restarts Tempo.
   - Verifies k6 smoke.

## Intentionally not changed

- Docker Zscaler certificate handling.
- CI workflows.
- Application Go code.
- Kubernetes resources.
