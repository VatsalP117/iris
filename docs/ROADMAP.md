# Iris Analytics Roadmap

## Product direction

Iris should become a boring, trustworthy, self-hosted analytics binary for
developers who want useful traffic and conversion metrics without operating a
large analytics platform.

The next releases should prioritize confidence in the numbers over feature
count. A small dashboard is valuable only when Iris can explain what every
metric means and demonstrate when events were accepted, rejected, duplicated,
or lost.

## Current position

The repository already has a strong foundation:

- a compact browser SDK;
- a Go ingestion and query server;
- transactional batch writes to SQLite;
- a bundled React dashboard;
- a single-container deployment model; and
- tests for the primary database aggregations.

Before Iris is treated as production-ready, it still needs an explicit delivery
contract, idempotent ingestion, precise metric definitions, stronger input and
access controls, operational tooling, and end-to-end reliability tests.

## Phase 0: Define the analytics contract

Document the behavior that all later implementation and tests must preserve:

- what constitutes a pageview;
- which route transitions count and which are deduplicated;
- what visitor and session metrics mean;
- how daily visitor rotation affects multi-day reports;
- whether time windows use UTC or a site-configured timezone;
- which URLs, query parameters, referrers, and properties are stored;
- the delivery guarantee and known browser loss conditions;
- how retries and duplicate submissions behave;
- supported payload and batch limits; and
- the initial supported deployment and traffic envelope.

### Exit criteria

- Dashboard labels match the written definitions.
- SDK and server documentation use the same terminology and defaults.
- Every contract rule has a planned automated test.

## Phase 1: v0.3 trust release

### Reliable delivery

- Generate a stable event ID in the SDK.
- Add event occurrence time, receive time, SDK version, and schema version.
- Enforce a unique event ID in storage.
- Make retries idempotent.
- Fall back to `fetch` when `sendBeacon` refuses to queue a payload.
- Add bounded retry with backoff while the page is alive.
- Preserve queued batches until success or a documented discard condition.
- Validate SDK batching options against backend limits.
- Expose useful delivery callbacks or debug status without claiming acceptance
  before the server responds.

The intended guarantee is idempotent at-least-once delivery where the browser
allows it. Iris should document unavoidable unload, browser, extension, and
network limitations instead of promising exactly-once browser delivery.

### Correct event semantics

- Prevent accidental multiple initialization.
- Deduplicate same-URL and framework-induced route events.
- Define support for `pushState`, `replaceState`, `popstate`, and hash changes.
- Correct SPA referrer behavior and exclude self-referrals.
- Separate daily unique visitors from range-level visitor metrics.
- Define session expiration rather than equating a browser tab with a session.
- Store event occurrence time separately from ingestion time.

### Secure and validated ingestion

- Register sites explicitly.
- Protect dashboard and read APIs with authentication.
- Treat CORS as browser policy, not authentication.
- Add configurable ingestion origin restrictions.
- Add rate limiting and abuse protection.
- Validate required fields, lengths, numeric ranges, event names, site IDs,
  URLs, and batch sizes.
- Reject unknown or unauthorized site IDs.
- Sanitize URLs and strip sensitive query parameters by default.
- Make click-text collection opt-in and provide property allow/deny hooks.

### Installation confidence

- Make the primary quickstart send a real pageview.
- Add an installation verifier and live-event inspector.
- Surface the last accepted event and actionable setup errors.
- Align the root README, SDK README, architecture notes, and marketing site.
- Replace absolute performance, privacy, and legal claims with precise ones.

### Exit criteria

- One initial pageview is emitted for one supported installation.
- One pageview is emitted for each supported route transition.
- Replaying the same event never increments analytics twice.
- A rejected beacon attempts the configured fallback.
- Temporary server failure does not silently discard an in-memory batch.
- Invalid and cross-site events are rejected.
- The browser-to-database path is covered by automated tests.

## Iris Reliability Lab

The Reliability Lab should be built before the trust release so the current
implementation has a preserved baseline. The same workloads will then measure
each roadmap change.

### Principles

- Run only against an isolated server and dedicated database.
- Generate deterministic events with a run ID and sequence number.
- Record expected events independently from Iris.
- Reconcile attempted, accepted, stored, missing, duplicate, unexpected, and
  incorrectly aggregated events.
- Report events per second separately from HTTP requests per second.
- Preserve the git revision, environment, workload, and random seed with every
  report.
- Prefer reproducible CLI runs over manually interpreted dashboards.

### Test layers

1. **Correctness:** complex but low-volume deterministic traffic covering every
   metric and boundary.
2. **Ingestion load:** individual and batched HTTP traffic at fixed arrival
   rates.
3. **Mixed load:** dashboard reads while ingestion continues.
4. **Browser reliability:** real SDK lifecycle, navigation, storage, beacon,
   offline, and multi-tab flows.
5. **Failure recovery:** server restarts, database locks, delayed responses,
   transient errors, disk failures, and backup restoration.

### Initial load profiles

| Profile | Offered load | Duration | Purpose |
|---|---:|---:|---|
| Smoke | 10 events/s | 30 seconds | Validate the lab |
| Baseline | 100 events/s | 2 minutes | Establish normal latency |
| Target | 500 events/s | 5 minutes | Measure expected higher traffic |
| Target high | 1,000 events/s | 5 minutes | Initial capacity target |
| Ramp | 100 to 2,000+ events/s | 8–10 minutes | Find saturation |
| Spike | 100 to 2,000 events/s | 30 seconds | Test sudden bursts |
| Soak | 250–500 events/s | 30–60 minutes | Find gradual degradation |

### Report contents

- git revision and build version;
- host CPU, memory, operating system, and deployment mode;
- planned, attempted, accepted, rejected, and failed requests;
- stored rows, missing sequences, logical duplicates, and unexpected rows;
- aggregate mismatches by metric;
- achieved event and request rates;
- p50, p90, p95, and p99 latency;
- CPU, memory, disk, database, and WAL growth;
- SQLite busy or locked errors;
- dashboard query latency during writes;
- failure recovery time;
- first failed load level; and
- evidence-backed bottleneck analysis.

### Initial implementation

The first implementation is a Go oracle and report generator that:

- produces deterministic event payloads;
- sends individual or batched requests at a configured event rate;
- records the exact accepted sequence set;
- reads the dedicated SQLite database;
- reconciles stored rows against accepted events;
- verifies headline API aggregates; and
- writes JSON and Markdown reports.

High-volume k6 profiles, browser automation, system metrics, profiling, and
failure injection will build on that deterministic core.

## Phase 2: Operational confidence

- Add versioned schema migrations.
- Enable and tune SQLite WAL and busy timeout behavior.
- Constrain the SQLite connection pool appropriately.
- Add compound indexes based on measured query plans.
- remove index-blocking timestamp expressions where possible;
- add health and readiness endpoints;
- add HTTP timeouts and graceful shutdown;
- sample or aggregate per-event logging;
- add ingestion counters for received, accepted, duplicate, rejected, and
  failed events;
- document backup, integrity-check, restore, retention, deletion, and export
  workflows; and
- publish a tested throughput and database-size envelope.

### Exit criteria

- Concurrent load does not produce unhandled SQLite lock errors inside the
  supported envelope.
- Backup restoration passes the correctness oracle.
- A server restart does not corrupt acknowledged data.
- Resource use and query latency remain bounded during the soak profile.

## Phase 3: Product usefulness

After the numbers are dependable, add the smallest high-value analytics layer:

- goals and conversions;
- custom-event reporting;
- UTM and campaign attribution;
- filters and segments;
- bounce rate and visit duration with explicit definitions;
- real-time traffic;
- internal-traffic and page exclusions;
- configurable site timezone;
- previous-period comparison; and
- CSV or API export.

## Phase 4: Product expansion

Only after real usage demonstrates demand, consider:

- funnels and user journeys;
- revenue attribution;
- retention and cohorts;
- multi-user teams and roles;
- scheduled reports and notifications;
- imports from other analytics products;
- additional databases or horizontal ingestion; and
- hosted or managed deployment options.

These features should not delay reliability, privacy safety, or operational
confidence.

## Release discipline

For every roadmap release:

1. Run the correctness oracle against the previous release.
2. Preserve its report as the baseline.
3. Run the same workload against the candidate release.
4. Explain every metric change.
5. Reject releases with unexplained loss, duplication, or aggregation drift.
6. Publish the supported scale and known limitations.

The goal is to make progress measurable: lower loss, no duplicate counting,
predictable query latency, safe recovery, and metric definitions users can
trust.
