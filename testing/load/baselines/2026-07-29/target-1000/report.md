# Iris Reliability Report: baseline-2026-07-29-target-1000

**Verdict:** PASS

Generated at: 2026-07-28 22:47:08 UTC

## Configuration

| Field | Value |
|---|---:|
| Target | `http://127.0.0.1:62003` |
| Database | `/Users/vatsalpatel/Desktop/Projects/iris/testing/load/baselines/2026-07-29/target-1000/server/iris.db` |
| Site | `iris-lab-baseline-2026-07-29-target-1000` |
| Offered rate | 1000 events/s |
| Planned events | 300000 |
| Batch size | 10 |
| Workers | 128 |
| Concurrent read rate | 0 requests/s |
| Read workers | 8 |

## Environment

| Field | Value |
|---|---|
| Git revision | `4cf9a78b9314864ebcb51074915ebb148f0034a4` |
| Git worktree modified | true |
| Go | `go1.25.3` |
| Platform | `darwin/arm64` |
| CPUs visible to load generator | 10 |

## Delivery

| Measurement | Value |
|---|---:|
| Events attempted | 300000 |
| Events accepted | 300000 |
| Events rejected | 0 |
| Request errors | 0 |
| Requests attempted | 30000 |
| Achieved event rate | 1000.02 events/s |
| Achieved request rate | 100.00 requests/s |
| Maximum scheduling lag | 3728.43 ms |

### HTTP statuses

| Status | Requests |
|---|---:|
| 202 | 30000 |

## Storage reconciliation

| Measurement | Value |
|---|---:|
| Stored rows | 300000 |
| Unique sequences | 300000 |
| Missing accepted events | 0 |
| Duplicate rows | 0 |
| Unexpected rows | 0 |
| Field mismatches | 0 |
| Database size | 155357184 bytes |

## Request latency

| Percentile | Milliseconds |
|---|---:|
| Average | 6.78 |
| p50 | 2.66 |
| p90 | 3.77 |
| p95 | 4.10 |
| p99 | 46.42 |
| Maximum | 2276.20 |

## Server resources

| Measurement | Value |
|---|---:|
| Samples | 599 |
| Average CPU | 24.95% |
| Peak CPU | 107.80% |
| Average RSS | 37923872 bytes |
| Peak RSS | 64897024 bytes |
| Database growth | 155340800 bytes |
| Peak WAL | 0 bytes |

## Aggregate checks

| Check | Result | Error |
|---|---|---|
| stats | PASS |  |
| pages | PASS |  |
| referrers | PASS |  |
| vitals | PASS |  |
| devices | PASS |  |
| pageviews-timeseries | PASS |  |
| visitors-timeseries | PASS |  |
| sessions-timeseries | PASS |  |
| date-window-day | PASS |  |
| date-window-time | PASS |  |
