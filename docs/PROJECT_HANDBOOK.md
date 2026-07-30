# Iris Analytics Project Handbook

> **Read first · Ownership-transfer index**
>
> Evidence snapshot: `965cbc7` (`codex/marketing-redesign`), inspected 2026-07-30. The worktree was clean before this handbook was added.

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

The existing [roadmap](./ROADMAP.md) remains the product/reliability plan. The root [`iris_architecture.md`](../iris_architecture.md) is an older architectural summary; use this handbook as the current evidence-based source.

## Confidence labels

- **Confirmed** — directly established by current source, configuration, tests, or committed reports.
- **Strong inference** — multiple facts support the conclusion, but motivation or production state is not explicit.
- **Possible interpretation** — plausible but weakly supported.
- **Unknown** — the repository cannot answer it.
- **Contradictory evidence** — current sources disagree; both sides are cited.

Paths and line numbers are approximate and refer to this evidence snapshot. Symbols are included so references survive normal edits.

## Explain Iris in two minutes

Iris is a self-hosted web analytics system. A website embeds the `iris-analytics` npm SDK. In the browser, the SDK builds anonymous event payloads from the current URL, referrer, viewport width, a daily rotating visitor ID, and a tab-scoped session ID. It can capture pageviews, clicks, and Core Web Vitals or accept custom events (`web/src/index.ts → Iris`).

The SDK sends JSON, immediately or in batches, to a Go `net/http` server. The server overwrites the event ID and timestamp, truncates nested property strings, and inserts the row into one SQLite `events` table (`pkg/api/handler.go → TrackEvent()/TrackBatchEvents()`; `pkg/db/sqlite.go`). Query endpoints aggregate those rows into stats, pages, referrers, device categories, daily series, custom events, and performance reports (`pkg/db/query.go`).

A React/Vite dashboard calls those public query endpoints and renders four in-memory views: overview, inferred sites, custom events, and Web Vitals (`dashboard/src/App.tsx`). In production, the dashboard and Go server ship in one image; SQLite lives on a mounted volume. A separate React/Nginx image serves the marketing and public documentation site (`Dockerfile`; `marketing/Dockerfile`).

The most important caveat: **Iris currently has no authentication, authorization, registered-site model, rate limiting, migration system, health endpoint, graceful shutdown, or retry-safe event identity.** Anyone who can reach it can write events for any site ID and read all analytics. The repository’s own roadmap calls these pre-production gaps (`docs/ROADMAP.md:25-27,50-108`).

## Architecture-review explanation

Iris is a deliberately compact modular monolith with an external browser collector:

1. The browser SDK owns client lifecycle, anonymous identifiers, event construction, and best-effort delivery.
2. The Go process owns HTTP routing, server-assigned identity/time, persistence, aggregate semantics, and static dashboard serving.
3. SQLite owns durable event rows and primary-key uniqueness, but almost no domain constraints.
4. The dashboard owns selection/date state and presentation; it duplicates response types and some Core Web Vitals thresholds.
5. The marketing site is a separate static deployment boundary and has no runtime dependency on analytics.
6. The Reliability Lab is test infrastructure, not a production worker. It launches isolated real servers, reconciles accepted events against SQLite/public aggregates, injects faults, and runs browser state/lifecycle scenarios (`internal/reliability`; `testing/load`).

Communication is synchronous HTTP except for the browser’s fire-and-forget Beacon/fetch behavior. There are no queues, workers, cron jobs, webhooks, caches, or external application APIs. A single event table is both the write model and reporting source, so deployment is simple but ingestion contention, analytical scan cost, retention, tenancy, and schema evolution all converge on one file.

## Concepts to learn first

1. **Event taxonomy:** `$pageview`, `$click`, `$web_vital`, and names not beginning with `$`.
2. **Identity semantics:** visitor = browser-local random ID per UTC day; session = `sessionStorage` ID, usually per tab—not a timed analytics session.
3. **Site isolation is only a query value:** `site_id` is caller supplied, unregistered, and unauthenticated.
4. **Server time wins:** the backend discards any client event ID/time and applies ingestion time.
5. **One event table:** every metric is computed from `events`; no materialized aggregates or caches exist.
6. **Best-effort delivery:** the SDK removes queued items before acceptance and implements no durable queue, retry, or idempotency.
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

Satisfy every item in the final ownership checklist, execute a backup/restore drill on disposable data, propose/record the security and schema-migration ADRs, and implement one small end-to-end change without breaking the reliability oracle.

## Current posture at a glance

| Dimension | Assessment | Evidence |
|---|---|---|
| Functional shape | Coherent small analytics stack | `web/src`, `cmd/server`, `pkg`, `dashboard/src` |
| Production security | **Not suitable for public exposure without an external access layer** | No auth in routes; permissive CORS; arbitrary site IDs |
| Data correctness | Good aggregate tests and exceptional deterministic oracle; incomplete ingestion contract | `pkg/db/query_test.go`; `internal/reliability`; `docs/ROADMAP.md` |
| Delivery reliability | Known browser loss/duplication/navigation defects | `testing/load/baselines/2026-07-29/browser/browser-report.md` |
| Capacity | Machine-specific evidence: single writes fail at 500 requests/s; batches passed at 1,000 events/s/100 requests/s | baseline README and reports |
| Operations | Container is simple; health, migrations, backups, shutdown, telemetry are mostly manual/absent | `Dockerfile`; `cmd/server/main.go`; roadmap Phase 2 |
| Maintainability | Small and readable; weak contracts, UI duplication, and historical artifacts add risk | chapters 03 and 06 |
