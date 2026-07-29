# Iris Reliability Report: baseline-2026-07-29-target-500

**Verdict:** FAIL

Generated at: 2026-07-28 20:48:51 UTC

## Configuration

| Field | Value |
|---|---:|
| Target | `http://127.0.0.1:59390` |
| Database | `/Users/vatsalpatel/Desktop/Projects/iris/testing/load/baselines/2026-07-29/target-500/server/iris.db` |
| Site | `iris-lab-baseline-2026-07-29-target-500` |
| Offered rate | 500 events/s |
| Planned events | 150000 |
| Batch size | 1 |
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
| Events attempted | 150000 |
| Events accepted | 149963 |
| Events rejected | 37 |
| Request errors | 0 |
| Requests attempted | 150000 |
| Achieved event rate | 500.00 events/s |
| Achieved request rate | 500.00 requests/s |
| Maximum scheduling lag | 5513.07 ms |

### HTTP statuses

| Status | Requests |
|---|---:|
| 202 | 149963 |
| 500 | 37 |

## Storage reconciliation

| Measurement | Value |
|---|---:|
| Stored rows | 149963 |
| Unique sequences | 149963 |
| Missing accepted events | 0 |
| Duplicate rows | 0 |
| Unexpected rows | 0 |
| Field mismatches | 0 |
| Database size | 77381632 bytes |

## Request latency

| Percentile | Milliseconds |
|---|---:|
| Average | 19.85 |
| p50 | 1.12 |
| p90 | 1.81 |
| p95 | 3.00 |
| p99 | 580.52 |
| Maximum | 6320.67 |

## Server resources

| Measurement | Value |
|---|---:|
| Samples | 603 |
| Average CPU | 52.39% |
| Peak CPU | 74.40% |
| Average RSS | 39336842 bytes |
| Peak RSS | 72040448 bytes |
| Database growth | 77357056 bytes |
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

## Diagnostic samples

- Request errors: `HTTP 500 for 1 event(s): Failed to save event`, `HTTP 500 for 1 event(s): Failed to save event`, `HTTP 500 for 1 event(s): Failed to save event`, `HTTP 500 for 1 event(s): Failed to save event`, `HTTP 500 for 1 event(s): Failed to save event`, `HTTP 500 for 1 event(s): Failed to save event`, `HTTP 500 for 1 event(s): Failed to save event`, `HTTP 500 for 1 event(s): Failed to save event`, `HTTP 500 for 1 event(s): Failed to save event`, `HTTP 500 for 1 event(s): Failed to save event`, `HTTP 500 for 1 event(s): Failed to save event`, `HTTP 500 for 1 event(s): Failed to save event`, `HTTP 500 for 1 event(s): Failed to save event`, `HTTP 500 for 1 event(s): Failed to save event`, `HTTP 500 for 1 event(s): Failed to save event`, `HTTP 500 for 1 event(s): Failed to save event`, `HTTP 500 for 1 event(s): Failed to save event`, `HTTP 500 for 1 event(s): Failed to save event`, `HTTP 500 for 1 event(s): Failed to save event`, `HTTP 500 for 1 event(s): Failed to save event`, `HTTP 500 for 1 event(s): Failed to save event`, `HTTP 500 for 1 event(s): Failed to save event`, `HTTP 500 for 1 event(s): Failed to save event`, `HTTP 500 for 1 event(s): Failed to save event`, `HTTP 500 for 1 event(s): Failed to save event`, `HTTP 500 for 1 event(s): Failed to save event`, `HTTP 500 for 1 event(s): Failed to save event`, `HTTP 500 for 1 event(s): Failed to save event`, `HTTP 500 for 1 event(s): Failed to save event`, `HTTP 500 for 1 event(s): Failed to save event`, `HTTP 500 for 1 event(s): Failed to save event`, `HTTP 500 for 1 event(s): Failed to save event`, `HTTP 500 for 1 event(s): Failed to save event`, `HTTP 500 for 1 event(s): Failed to save event`, `HTTP 500 for 1 event(s): Failed to save event`, `HTTP 500 for 1 event(s): Failed to save event`, `HTTP 500 for 1 event(s): Failed to save event`
