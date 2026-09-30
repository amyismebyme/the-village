# k6 Load Testing

Sprint 26 establishes the performance baseline and load model for The Village.

The tests use a read-heavy workload because the current application is primarily a community API and the goal is to measure serving-path behavior without turning the performance test into a data-generation test.

## Test profiles

| Profile | Default load | Purpose |
|---|---:|---|
| `smoke` | 1 VU / 1 iteration | Endpoint sanity plus one CRUD cycle |
| `baseline` | 10 VUs / 60s | Establish repeatable normal-load baseline |
| `sustained` | 10 → 25 → 50 VUs | Observe behavior as load increases and then falls |
| `spike` | 5 → 75 VUs | Observe sudden traffic pressure and recovery |

The default thresholds are guardrails for local testing, not production SLOs. Override them with environment variables when the baseline is known:

- `K6_BASE_URL`
- `K6_VUS`
- `K6_DURATION`
- `K6_P95_MS`
- `K6_ERROR_RATE`

## Run with Docker

The project scripts use the official `grafana/k6:2.3.0` image so the host does not need a k6 installation.

PowerShell:

```powershell
.\scripts\run-k6.ps1 smoke
.\scripts\run-k6.ps1 baseline
.\scripts\run-k6.ps1 sustained
.\scripts\run-k6.ps1 spike
```

Bash:

```bash
./scripts/run-k6.sh smoke
./scripts/run-k6.sh baseline
./scripts/run-k6.sh sustained
./scripts/run-k6.sh spike
```

The default Docker target is `http://host.docker.internal:8080` so the test container can exercise the locally published API port.

## Results

Summary JSON files are written to:

```text
artifacts/k6/
```

Those are local test artifacts and should not be committed.

## What to record

For the baseline, record:

- requests per second / iterations per second
- p50, p95 and p99 request latency
- HTTP failure rate
- checks passed/failed
- allocations are measured separately by Go benchmarks
- API CPU and memory
- PostgreSQL connection pool utilization
- database query latency
- cache hit ratio
- external/worker activity where relevant

## Interpretation

Use the progression:

```text
smoke
  ↓
baseline
  ↓
sustained
  ↓
spike
```

The first objective is to find the operating shape and saturation behavior. Do not convert these local numbers directly into a production SLO without production usage data.
