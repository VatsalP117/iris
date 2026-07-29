# Iris Reliability Lab

The Iris Reliability Lab is a deterministic correctness oracle, load generator,
fault injector, profiler, and browser SDK test harness. It answers two separate
questions:

1. Did Iris store and report every event it accepted, exactly once and with the
   expected fields?
2. How much traffic can this server and SDK sustain, and where does it fail?

The Go lab starts a real Iris server on a random loopback port with a dedicated
SQLite database. Generated traffic carries a unique test run and sequence
number, so the lab can reconcile the HTTP results, raw rows, and public API
aggregates without relying on approximate counts.

## What is covered

The complete suite measures and checks:

- single-event and batch ingestion;
- fixed-rate 100, 500, and 1,000 events/s profiles;
- ramp, spike, mixed read/write, and 30-minute soak profiles;
- planned, attempted, accepted, rejected, failed, missing, duplicate,
  unexpected, and field-mismatched events;
- HTTP status counts, request throughput, scheduling lag, and average, p50,
  p90, p95, p99, and maximum latency;
- concurrent reads of stats, pages, referrers, vitals, devices, and all three
  time-series endpoints;
- exact stats, pages, referrers, vitals, devices, pageview, visitor, and
  session aggregate reconciliation;
- date-only and date-time boundary behavior;
- server CPU, RSS, process I/O, database growth, and WAL growth;
- CPU, heap, and goroutine profiles;
- restart, intermittent HTTP 503, network outage, SQLite lock, and SQLite-full
  behavior;
- online SQLite backup integrity, row reconciliation, restored-server boot,
  and restored API responses;
- browser SDK startup, duplicate starts, history navigation, hash changes,
  multiple instances, stopping, page-hide flushing, storage failure, multi-tab
  identity, offline/retry behavior, beacon fallback, and batching overhead;
- an independent k6 fixed-arrival, ramp, spike, mixed-read, and browser driver;
  and
- baseline/candidate regression comparison.

Reports are emitted as Markdown for people and JSON for CI or later analysis.

## Continuous integration

[GitHub Actions](../../.github/workflows/ci.yml) runs the oracle on every pull
request and every push to `main`. The workflow has four independent gates:

- Go tests with the race detector, vet, and backend builds;
- JavaScript builds, lint, and comparator tests;
- the complete quick correctness and fault suites; and
- the Chromium browser oracle compared with the committed baseline.

The browser job intentionally allows known SDK failures while generating its
candidate report. The comparator fails CI only when a previously passing
scenario regresses, a scenario contract changes, coverage is lost, or the
harness itself errors. Existing failures remain visible as known failures, and
fixes are reported as improvements.

Every oracle job uploads its Markdown, JSON, logs, databases, and profiles even
when a gate fails. GitHub retains these artifacts for 14 days. Configure the
single `CI required` check as the required branch-protection check so job names
and implementation details can change without updating repository settings.

## Safety and isolation

`iris-lab suite` and `iris-lab faults` create their own server, database, port,
and artifact directory. This is the preferred way to run the oracle.

The lower-level `iris-lab run` command accepts an existing server and database.
It refuses non-loopback targets unless `--allow-nonlocal` is supplied. A
loopback server can still point at valuable data, so always use a disposable
database.

The lab does not delete databases. Generated artifacts under
`artifacts/reliability/` are ignored by Git.

## Fast development gate

Build and run every profile with shortened durations:

```bash
IRIS_LAB_QUICK=true task lab:suite
IRIS_LAB_QUICK=true task lab:faults
task lab:browser
```

The quick suite takes roughly one minute and still performs complete
event-by-event and aggregate reconciliation. It is useful before a commit, but
it is not a substitute for the full-duration capacity and soak runs.

The browser command exits non-zero when it exposes an SDK defect. To collect a
report while allowing the overall command to continue:

```bash
pnpm exec playwright install chromium # only when no system Chrome/Chromium exists
pnpm lab:browser -- --allow-failures
```

Browser artifacts default to `artifacts/reliability/browser-<timestamp>/`.
Set `IRIS_BROWSER_OUTPUT` to choose a stable directory.

### Browser oracle layers

The Chromium oracle currently runs 35 scenarios in five groups:

- **SDK flows:** initialization, repeated starts, direct history navigation,
  multiple instances, storage failure, identity, batching, and high-volume
  enqueue behavior.
- **Generated state machine:** 16 deterministic lifecycle traces generated from
  `start`, `stop`, manual tracking, unique and same-URL `pushState`,
  `replaceState`, and hash actions. The independent model compares the exact
  ordered event names and URLs, not only totals. Failures include the seed and
  replayable action trace.
- **Delivery chaos:** offline recovery, 408/429/502/503/504 responses, permanent
  400 rejection, accepted requests whose response connection disappears,
  hanging and slow connections, beacon refusal, and allowed or rejected CORS
  preflights.
- **Framework fixtures:** real React Router declarative and data-router
  applications exercise links, programmatic navigation, replacement, and
  redirect behavior.
- **Lifecycle:** an actual Chromium BFCache round trip, pagehide/pageshow,
  Chrome freeze/resume, hidden-page batch flushing, duplicated tabs, navigation
  before SDK loading, and abrupt page close.

Playwright normally disables Chromium BFCache. The oracle deliberately removes
that launch flag and verifies `PageTransitionEvent.persisted` in both directions
before treating the BFCache assertion as exercised.

## Full profile suite

Run all standard profiles:

```bash
task lab:suite
```

Profiles can be selected:

```bash
IRIS_LAB_PROFILES=target-500,target-1000 task lab:suite
IRIS_LAB_PROFILES=mixed,ramp,spike task lab:suite
IRIS_LAB_PROFILES=soak task lab:suite
```

Standard definitions:

| Profile | Traffic |
|---|---|
| `smoke` | 10 events/s for 30 seconds |
| `baseline` | 100 events/s for 2 minutes, single ingestion |
| `target-500` | 500 events/s for 5 minutes, single ingestion |
| `target-1000` | 1,000 events/s for 5 minutes, batches of 10 |
| `mixed` | 500 events/s plus 25 analytics reads/s for 5 minutes |
| `ramp` | 100, 500, 1,000, then 2,000 events/s; 2 minutes each |
| `spike` | 100 events/s, 2,000 events/s spike, then recovery |
| `soak` | 250 events/s plus 10 analytics reads/s for 30 minutes |

Events per second and HTTP requests per second are reported separately. At
1,000 events/s with batches of 10, Iris receives about 100 ingestion requests/s.

Profiles write `summary.json`, `report.md`, `resources.csv`, and, when enabled,
`cpu.pprof`, `heap.pprof`, and `goroutines.txt`. The suite root contains
`suite-summary.json` and `suite-report.md`. Each profile has its own fresh
server, database, and server log so earlier profiles cannot skew later results.

## Fault and recovery suite

Run the standard or shortened fault suite:

```bash
task lab:faults
IRIS_LAB_QUICK=true task lab:faults
```

Injected transport failures are expected to reject some events. A fault
scenario passes only when the server recovers and every event that received an
accepted response exists exactly once with the correct data. The SQLite-full
scenario must also demonstrate at least one rejected write.

If a connection disappears after storage but before the response reaches the
client, the report counts that row as unexpected/ambiguous. It remains visible
as a delivery-risk metric, but does not fail the storage invariant unless the
row is duplicated or malformed. Client retry behavior for this case is checked
separately by the browser oracle.

The backup scenario uses SQLite `VACUUM INTO`, runs `integrity_check`, reconciles
the copied rows, boots a fresh Iris process from the copy, and validates all
public aggregates against the accepted manifest.

## Independent k6 driver

k6 is run in Docker, so no host installation is required. Start an isolated
Iris server, then pass the address visible from the container:

```bash
IRIS_K6_TARGET=http://host.docker.internal:8080 \
IRIS_K6_RATE=500 \
IRIS_K6_DURATION=5m \
IRIS_K6_BATCH_SIZE=1 \
task lab:k6
```

For batched 1,000 events/s and concurrent reads:

```bash
IRIS_K6_TARGET=http://host.docker.internal:8080 \
IRIS_K6_RATE=1000 \
IRIS_K6_DURATION=5m \
IRIS_K6_BATCH_SIZE=10 \
IRIS_K6_READ_RATE=25 \
task lab:k6
```

Set `IRIS_K6_PROFILE=ramp` or `spike` for the staged k6 profiles. The raw script
is [iris.js](./k6/iris.js).

The optional k6 browser probe requires a k6 image with browser support and a
running page:

```bash
docker run --rm -i \
  -e K6_BROWSER_HEADLESS=true \
  -e TARGET_PAGE=http://host.docker.internal:5173 \
  grafana/k6:latest-with-browser \
  run - < testing/load/k6/browser.js
```

## One-off deterministic runs

For an existing disposable server:

```bash
go run ./cmd/iris-lab run \
  --target http://127.0.0.1:8080 \
  --db /tmp/iris-reliability/iris.db \
  --rate 500 \
  --duration 5m \
  --batch-size 1 \
  --read-rate 25
```

Use `--events 1000` for an exact count, or supply staged rates:

```bash
go run ./cmd/iris-lab run \
  --target http://127.0.0.1:8080 \
  --db /tmp/iris-reliability/iris.db \
  --batch-size 10 \
  --stages 100:30s,500:1m,1000:1m,2000:30s
```

## Compare revisions

Use reports produced with identical load configurations:

```bash
IRIS_LAB_BASELINE=/path/to/baseline/summary.json \
IRIS_LAB_CANDIDATE=/path/to/candidate/summary.json \
task lab:compare
```

The comparison requires matching load configurations, OS, architecture, and
visible CPU count. It fails for incompatible runs or regressions beyond the
built-in tolerances. Correctness errors, rejected events, missing rows, and
duplicates have zero tolerance. Throughput, latency, CPU, RSS, read latency,
and database bytes per event have explicit percentage thresholds.

## Reading a failure

A normal correctness run fails when:

- a planned event was not attempted and accepted;
- an accepted sequence is absent, duplicated, unexpected, or changed in
  SQLite;
- an analytics endpoint fails during mixed load; or
- a public aggregate differs from the accepted manifest.

The Markdown report includes diagnostic samples. `summary.json` contains the
complete counters, timings, resource metrics, endpoint results, and verdict.
The server log and Go profiles should be inspected next when reconciliation
passes but latency, CPU, memory, or scheduling lag deteriorates.

Committed full-duration reference runs live in
[`baselines/`](./baselines/). They describe the tested machine and revision;
they are evidence for comparison, not universal production capacity claims.
