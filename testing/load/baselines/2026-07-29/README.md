# Iris full-duration baseline — 2026-07-29

This baseline ran each profile with a fresh Iris process and SQLite database on
an Apple M4 MacBook Air with 10 logical CPUs, 16 GiB RAM, Darwin arm64, and Go
1.25.3.

The binary recorded revision
`4cf9a78b9314864ebcb51074915ebb148f0034a4` with `git_modified: true` because
the testing-oracle implementation itself was the worktree being measured.

## Results

| Profile | Verdict | Planned | Accepted | Rejected | Events/s | Requests/s | p95 | p99 | Peak RSS |
|---|---|---:|---:|---:|---:|---:|---:|---:|---:|
| 500 events/s, single writes | FAIL | 150,000 | 149,963 | 37 | 500.00 | 500.00 | 3.00 ms | 580.52 ms | 68.7 MiB |
| 1,000 events/s, batch 10 | PASS | 300,000 | 300,000 | 0 | 1,000.02 | 100.00 | 4.10 ms | 46.42 ms | 61.9 MiB |

Both profiles stored every acknowledged event exactly once with no unexpected
rows or field mismatches. The failed single-write profile returned 37 HTTP 500
responses; every corresponding server error was `database is locked`. Its
maximum latency was 6.32 seconds and generator scheduling lag reached 5.51
seconds.

The batched profile had no delivery or reconciliation errors. Its maximum
latency was 2.28 seconds and maximum scheduling lag was 3.73 seconds, so the
tail still warrants attention even though the correctness gate passed.

## Interpretation

The tested binary does not have a safe 500-request/s single-write envelope:
SQLite contention rejects writes and causes tail collapse. The 1,000-event/s
result is materially better because batching reduces ingestion to about 100
transactions and HTTP requests per second.

The evidence points to transaction/connection contention as the immediate
bottleneck. Per-event synchronous logging is also a likely contributor and
should be isolated in a follow-up comparison. SQLite WAL was not enabled, so WAL
growth remained zero. The next product changes to measure are a bounded SQLite
connection pool, busy timeout and WAL configuration, and sampled ingestion
logging.

These results describe this machine, binary, database state, and workload. They
are not a production capacity guarantee.

Detailed human and machine reports are in
[`target-500/`](./target-500/) and [`target-1000/`](./target-1000/).

## Browser SDK baseline

The real-browser oracle passed 7 of 15 scenarios. It exposed same-URL
double-counting, missing `replaceState` tracking, double-counting from multiple
instances, no fallback when `sendBeacon` refuses a payload, no retry after a
transient 503 or offline period, inconsistent visitor identity across tabs, and
loss under a 1,000-event burst. These are product defects for the trust-release
roadmap, not oracle failures. See [`browser/browser-report.md`](./browser/browser-report.md).
