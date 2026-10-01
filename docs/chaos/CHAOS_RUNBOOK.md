# Sprint 27 — Chaos Engineering Runbook

## Purpose

Sprint 27 introduces controlled failure experiments for The Village. The goal is to demonstrate that the local production-style stack can detect dependency failures, preserve serving behavior where appropriate, and recover cleanly after a failure.

These experiments are intentionally **not** run in CI. They mutate the running Docker Compose environment and must be started explicitly.

## Preconditions

Start the normal stack first:

```powershell
docker compose up -d --build
```

Confirm both liveness and readiness:

```powershell
curl.exe http://localhost:8080/health
curl.exe http://localhost:8080/ready
```

Run the Sprint 26 smoke test before chaos work:

```powershell
.\scripts\run-k6.ps1 smoke
```

The stack should be healthy before every experiment.

## Safety model

The runner requires explicit confirmation because it intentionally stops or kills containers.

PowerShell:

```powershell
.\scripts\run-chaos.ps1 api-kill -ConfirmChaos
```

Shell:

```bash
CHAOS_CONFIRM=1 ./scripts/run-chaos.sh api-kill
```

The runner writes a JSON evidence file under `artifacts/chaos/` and attempts to restore services in a `finally`/trap cleanup path.

## Experiment 1 — `api-kill`

### Failure

Send `SIGKILL` to `village-api`.

### Expected behavior

The container should be restarted by Docker Compose's `restart: unless-stopped` policy.

The recovery sequence is:

```text
API killed
   ↓
/health unavailable briefly
   ↓
container restarts
   ↓
/health = 200
   ↓
/ready = 200
   ↓
k6 smoke passes
```

### Run

```powershell
.\scripts\run-chaos.ps1 api-kill -ConfirmChaos
```

### Evidence

Record at least:

- outage duration until `/health` returns 200
- outage duration until `/ready` returns 200
- k6 smoke result after recovery
- container restart behavior in `docker compose ps`

## Experiment 2 — `postgres-outage`

### Failure

Stop the PostgreSQL service without stopping the API.

### Expected behavior

This experiment verifies the distinction between liveness and readiness:

```text
PostgreSQL stopped
       │
       ├── /health → 200
       │
       └── /ready  → 503
```

After PostgreSQL starts again:

```text
PostgreSQL restarted
       ↓
API reconnects / dependency becomes healthy
       ↓
/ready → 200
       ↓
k6 smoke passes
```

### Run

```powershell
.\scripts\run-chaos.ps1 postgres-outage -ConfirmChaos
```

This is a key SRE behavior: the process is alive, but the service should not advertise readiness while its required database dependency is unavailable.

## Experiment 3 — `tempo-outage`

### Failure

Stop Tempo, the OTLP trace backend.

### Expected behavior

Tempo is an observability dependency, not a serving dependency. The API should continue to serve requests:

```text
Tempo stopped
      │
      ├── /health → 200
      ├── /ready  → 200
      └── community list → 200
```

The experiment then restarts Tempo and runs the smoke test again.

### Run

```powershell
.\scripts\run-chaos.ps1 tempo-outage -ConfirmChaos
```

The important operational question is whether loss of telemetry storage compromises the application serving path.

## Evidence and observations

Every run produces a JSON file:

```text
artifacts/chaos/
├── api-kill-YYYYMMDD-HHMMSS.json
├── postgres-outage-YYYYMMDD-HHMMSS.json
└── tempo-outage-YYYYMMDD-HHMMSS.json
```

Use those files to capture a small incident-style record:

| Field | What to record |
|---|---|
| scenario | experiment name |
| started_at / ended_at | UTC timestamps |
| duration_ms | total experiment duration |
| observations | endpoint status timeline |
| success | whether recovery verification passed |
| error | failure reason, when present |

For a stronger SRE exercise, correlate the chaos run with Grafana/Prometheus/Loki/Tempo evidence after the experiment.

## What Sprint 27 proves

Sprint 26 established **how the service behaves under load**.

Sprint 27 establishes **how the service behaves when dependencies or processes fail**.

Together they give us the evidence needed for the next stage:

```text
Measure performance
        ↓
Break the system safely
        ↓
Observe recovery
        ↓
Use the evidence to configure Kubernetes
```

Kubernetes is deliberately deferred to Sprint 28 so resource requests, probes, restart behavior, and replica strategy can be based on observed behavior rather than arbitrary values.
