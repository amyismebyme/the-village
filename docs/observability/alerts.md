# Initial high-signal alerts

Prometheus evaluates the rules and sends firing alerts to Alertmanager. Alertmanager controls grouping, repeat intervals and inhibition.

| Alert | Threshold | For |
|---|---:|---:|
| VillageAPIUnavailable | target unavailable | 2m |
| VillageAPIHigh5xxRate | >5% HTTP 5xx | 5m |
| VillageExternalHigh5xxRate | >10% external 5xx | 5m |
| VillageExternalRetryExhaustion | any exhaustion | immediate |
| VillageWorkerFailuresElevated | >0.005 failures/s | 10m |
| VillageExternalRateLimitPressure | >0.01 waits/s and >500ms avg wait | 15m |
| VillageCacheHitRatioLow | <70% | 15m |
| VillageDatabaseConnectionsUnavailable | 0 active DB connections | 2m |

These thresholds are conservative development/initial-production defaults. Tune them against real workload baselines before treating warning severity as paging-worthy.
