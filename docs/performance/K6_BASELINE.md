# k6 Performance Baseline

## Purpose

Sprint 26 uses k6 to establish a repeatable performance baseline before Kubernetes resource sizing and autoscaling decisions.

## Workload model

The default workload is intentionally read-heavy:

- 10% `GET /health`
- 10% `GET /ready`
- 80% `GET /api/v1/communities?limit=20&offset=0`

The smoke profile additionally executes one create/update/delete cycle so the write path is exercised without creating sustained write pressure.

## Profiles

### Smoke

1 VU for 30 seconds.

Purpose: ensure the environment is usable before collecting a baseline.

### Baseline

10 VUs for 60 seconds.

Purpose: establish normal local behavior and repeatability.

### Sustained

10 → 25 → 50 VUs with a controlled ramp down.

Purpose: show how latency, throughput, database pressure, and failures change as concurrency grows.

### Spike

5 → 75 VUs rapidly, hold, then return to 5 VUs.

Purpose: observe transient overload behavior and recovery.

## Results template

Record results from the baseline environment here after running the tests.

| Measurement | Result | Notes |
|---|---:|---|
| p50 latency | TODO | |
| p95 latency | TODO | |
| p99 latency | TODO | |
| request failure rate | TODO | |
| iterations/sec | TODO | |
| API CPU | TODO | |
| API memory | TODO | |
| DB connections used | TODO | |
| DB query p95 | TODO | |
| cache hit ratio | TODO | |

These values are observations, not production SLOs.
