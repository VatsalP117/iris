# Iris Analytics Project Handbook

> **Read first · Ownership-transfer index**
>
> Evidence snapshot: `965cbc7` (`codex/marketing-redesign`), inspected 2026-07-30. The worktree was clean before this handbook was added.
>
> Current architecture correction: the v2 data architecture landed after this
> evidence snapshot. Statements below have been updated where behavior changed;
> dated reliability results remain historical evidence.

## Why this is a documentation set

Iris is small enough to understand end to end, but it has four independently built and deployed concerns: browser SDK, analytics server/database, dashboard, and marketing site. A single giant file would make API lookup, incident response, and deliberate study harder. This handbook therefore uses focused, interconnected Markdown chapters:

| Read | Purpose | Brief sections covered |
|---|---|---|
| [01 — System and repository](./01_SYSTEM_AND_REPOSITORY.md) | Product overview, repository map, C4-style architecture, startup, dependencies | 1–4, 13 |
| [02 — Runtime journeys and frontend](./02_RUNTIME_JOURNEYS_AND_FRONTEND.md) | Event flows, dashboard flows, SDK, dashboard, marketing | 5–6 |
| [03 — Backend, domain, data, and APIs](./03_BACKEND_DATA_AND_APIS.md) | Backend layers, rules, schema, queries, complete API catalogue | 7–10 |
| [04 — Security and configuration](./04_SECURITY_AND_CONFIGURATION.md) | Auth/authorization truth, configuration reference, security review | 11–12, 19 |
| [05 — Delivery, operations, and debugging](./05_DELIVERY_OPERATIONS_AND_DEBUGGING.md) | Deployment, CI/CD, resilience, observability, runbooks | 14–15, 17–18, 25–26 |
| [06 — Quality, scale, and debt](./06_QUALITY_SCALE_AND_DEBT.md) | Testing map, performance evidence, maintainability, debt, unknowns | 16, 20–23 |
| [07 — Decisions and safe change](./07_DECISIONS_AND_SAFE_CHANGE.md) | Recorded/reconstructed ADRs, change recipes, future ADRs | ADR requirement, 24, 30 |
| [08 — Learning and ownership](./08_LEARNING_AND_OWNERSHIP.md) | Curriculum, question bank with separate answers, interview narratives, final checklist | 27–29, 31 |
| [09 — Evidence coverage](./09_EVIDENCE_COVERAGE.md) | Inspection record, claim-quality rules, inaccessible evidence, final self-review | Coverage report |
| [v2 data architecture](./09_V2_DATA_ARCHITECTURE.md) | Current site/event/storage/projection/retention contracts and ClickHouse boundary | Current data reference |

The existing [roadmap](./ROADMAP.md) remains the product/reliability plan. The
root [`iris_architecture.md`](../iris_architecture.md) is the concise current
system map; use this handbook and the v2 data chapter for deeper evidence.

## Confidence labels

- **Confirmed** — directly established by current source, configuration, tests, or committed reports.
- **Strong inference** — multiple facts support the conclusion, but motivation or production state is not explicit.
- **Possible interpretation** — plausible but weakly supported.
- **Unknown** — the repository cannot answer it.
- **Contradictory evidence** — current sources disagree; both sides are cited.

Paths and line numbers are approximate and refer to this evidence snapshot. Symbols are included so references survive normal edits.

## Explain Iris in two minutes

Iris is a self-hosted web analytics system. A website embeds the
`iris-analytics` npm SDK. The SDK builds versioned, client-identified events from
the current URL, referrer, viewport width, a daily rotating visitor ID, and a
same-origin session ID that rolls after 30 minutes of inactivity. It can capture
pageviews, clicks, and Core Web Vitals or accept custom events.

The SDK sends JSON, immediately or in batches, to a Go `net/http` server. The
server validates registered site/domain ownership, sanitizes URLs, preserves
client ID/occurrence time, adds receive time, and inserts an idempotent raw event.
Query endpoints aggregate raw facts while a background projector builds
rebuildable session and daily tables.

A React/Vite dashboard calls public query endpoints and renders four in-memory
views: overview, registered sites, custom events, and Web Vitals. In production,
the dashboard and Go server ship in one image; SQLite lives on a mounted volume.
A separate React/Nginx image serves marketing and public documentation.

The most important caveat: **analytics reads, site listing, and browser ingestion
remain unauthenticated and there is no rate limiting.** Site mutation is protected
by `IRIS_ADMIN_TOKEN`; registered sites, versioned migrations, status/health,
graceful shutdown, and retry-safe event identity now exist. Anyone who can reach
the service can still read analytics or submit events for a known site/domain.

## Architecture-review explanation

Iris is a deliberately compact modular monolith with an external browser collector:

1. The browser SDK owns client lifecycle, anonymous identifiers, event construction, and best-effort delivery.
2. The Go process owns validation, receive time, persistence, projection and
   retention scheduling, aggregate semantics, and static dashboard serving.
3. SQLite owns registered sites, constrained raw facts, migration history,
   checkpoints, and rebuildable projections.
4. The dashboard owns selection/date state and presentation; it duplicates response types and some Core Web Vitals thresholds.
5. The marketing site is a separate static deployment boundary and has no runtime dependency on analytics.
6. The Reliability Lab is test infrastructure, not a production worker. It launches isolated real servers, reconciles accepted events against SQLite/public aggregates, injects faults, and runs browser state/lifecycle scenarios (`internal/reliability`; `testing/load`).

Communication is synchronous HTTP except for browser Beacon/fetch and the
in-process maintenance loop. There is no external queue, worker process, cron,
webhook, or cache. Raw events are the write model/source of truth; sessions and
daily metrics are checkpointed projections in the same SQLite file.

## Concepts to learn first

1. **Event taxonomy:** `$pageview`, `$click`, `$web_vital`, and names not beginning with `$`.
2. **Identity semantics:** visitor = browser-local random ID per site-local day;
   session = per-site same-origin ID renewed after 30 minutes of tracked inactivity.
3. **Site registration:** event site/hostname must match a registered allowlist;
   this is integrity validation, not browser authentication.
4. **Two event times:** client occurrence and server receive time are distinct;
   client ID is the idempotency key.
5. **Facts and projections:** raw events are authoritative; session and daily
   tables can be rebuilt from a versioned checkpoint.
6. **Best-effort delivery:** server insertion is idempotent by client event ID,
   but the SDK removes queued items before acceptance and has no durable queue or retry.
7. **Same-origin dashboard:** production uses `BASE = ""`; Vite proxies `/api` in development.
8. **One main container plus persistent volume:** server and dashboard share an image/process boundary; SQLite durability depends on the mount.
9. **Evidence-backed known failures:** the committed browser baseline is a defect catalogue, not a fully passing E2E suite.

## Suggested reading order

### One-hour orientation

Read this file, the system diagram and process map in chapter 01, the pageview sequence in chapter 02, the event/domain/schema sections in chapter 03, and the critical findings in chapter 04.

### One-day understanding

Read chapters 01–05 in order. Then run the local setup, send one event with `curl`, inspect the row with `sqlite3`, and trace the dashboard’s 11 parallel overview requests.

### One-week deep study

Complete chapters 06–08, run Go tests and the quick Reliability Lab, replay selected browser failures, inspect Git commits cited by the ADR catalogue, and perform the exercises.

### Full project mastery

Satisfy every item in the final ownership checklist, execute a backup/restore
drill on disposable data, formalize the remaining read/ingest security and
migration-lifecycle ADRs, and implement one small end-to-end change without
breaking the reliability oracle.

## Current posture at a glance

| Dimension | Assessment | Evidence |
|---|---|---|
| Functional shape | Coherent small analytics stack | `web/src`, `cmd/server`, `pkg`, `dashboard/src` |
| Production security | **Not suitable for public dashboard exposure without an external access layer** | Admin token protects site mutation, but reads/ingestion remain public and CORS is permissive |
| Data correctness | Good aggregate tests and exceptional deterministic oracle; incomplete ingestion contract | `pkg/db/query_test.go`; `internal/reliability`; `docs/ROADMAP.md` |
| Delivery reliability | Known browser loss/duplication/navigation defects | `testing/load/baselines/2026-07-29/browser/browser-report.md` |
| Capacity | Machine-specific evidence: single writes fail at 500 requests/s; batches passed at 1,000 events/s/100 requests/s | baseline README and reports |
| Operations | Status, migrations, shutdown, projections, and retention exist; backups and telemetry remain operator work | `cmd/server/main.go`; `pkg/db` |
| Maintainability | Small and readable; weak contracts, UI duplication, and historical artifacts add risk | chapters 03 and 06 |
