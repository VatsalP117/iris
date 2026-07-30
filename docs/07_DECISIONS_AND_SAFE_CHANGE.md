# 07 — Architectural Decisions and Safe Change Guides

> **Deep dive · Decision record**

## ADR catalogue method

No formal ADR files were found. `iris_architecture.md` records architecture as a summary, and `docs/ROADMAP.md` records proposed direction rather than accepted decisions. All entries below are therefore **Reconstructed ADRs** unless explicitly labeled proposed. “Why” is historical fact only where a commit message/document states it.

## RADR-001 — Go `net/http` modular monolith

- **Status/date:** accepted, 2026-01-28; confidence medium.
- **Decision:** one Go server with standard `net/http`, handler/repository packages, and static dashboard.
- **Evidence:** commits `d1d5281`, `69c242b`; `cmd/server/main.go`; `pkg/api`; `pkg/db`.
- **Context/problem:** simple ingestion/query backend and dashboard delivery.
- **Why selected:** **Strong inference:** low operational footprint and direct control. No explicit alternative analysis.
- **Plausible alternatives:** Gin/Chi/Echo; Node backend; multiple services.
- **Positive consequences:** small dependency surface, simple tracing/deployment, Go concurrency.
- **Negative/risk:** manual method/middleware/error/server lifecycle; one failure boundary.
- **Operational/security/performance/cost:** one cheap process; no framework protections; static serving and APIs share capacity.
- **Developer experience:** approachable, but handler repetition grows.
- **Assessment/revisit:** still sensible. Revisit framework only if routing/middleware complexity demonstrably grows; split services only for measured scaling/ownership boundaries.
- **Questions for formal ADR:** expected endpoint count, team skills, hosting constraints, availability target.

## RADR-002 — SQLite as event store and reporting database

- **Status/date:** accepted, 2026-01-28; confidence high on decision, low on original motivation.
- **Decision:** embedded SQLite, one `events` table, startup DDL.
- **Evidence:** commit `66710fe`; `pkg/db/sqlite.go`; Docker volume.
- **Likely context:** self-hosted single binary with minimal dependencies.
- **Alternatives:** Postgres, ClickHouse, embedded analytics DB, append log + aggregates.
- **Benefits:** zero database service, transactional batch, portable file/backups, low cost.
- **Costs:** single-writer contention, one-node state, scan growth, migration/backup burden. Baseline confirms lock failures at 500 single writes/s.
- **Security:** filesystem/backup permissions become DB security.
- **Assessment:** appropriate for early self-hosted scope, underconfigured operationally.
- **Revisit trigger:** multi-replica requirement, database beyond supported envelope, sustained contention after WAL/pool/index/retention work, managed multi-tenant service.

## RADR-003 — One append-only generic event table

- **Status/date:** accepted, 2026-01-28 and evolved; confidence medium.
- **Decision:** system/custom/vital events share one row shape; properties JSON text.
- **Evidence:** `core.Event`, schema, all queries.
- **Plausible reason:** flexible SDK/event evolution without per-event migrations.
- **Trade-offs:** simple ingestion and new custom events; weak constraints/types, repeated strings, expensive aggregates, privacy risk, no site relations.
- **Assessment:** reasonable event-log start; needs schema/version/retention/site metadata around it before scale/security.
- **Revisit trigger:** frequent JSON querying, validation requirements, large scans, retention tiers, registered entities.

## RADR-004 — Browser-owned anonymous visitor/session IDs

- **Status/date:** accepted; daily rotation added 2026-03-25 in `5ed4028`.
- **Decision:** localStorage visitor ID rotates per UTC day; sessionStorage ID; memory fallback.
- **Evidence:** `web/src/storage.ts`; README/privacy docs.
- **Recorded motivation:** commit/package description and comments explicitly describe daily rotation/privacy.
- **Alternatives:** cookies, server/IP fingerprint, no ID, longer-lived pseudonym, timed sessions.
- **Benefits:** no cookies; simple distinct counts; no IP/user-agent logic.
- **Costs:** multi-day uniqueness inflation, storage clearing/privacy variability, tab semantics, duplicated-tab failure; localStorage is still client tracking state.
- **Assessment:** coherent privacy/product trade-off only if metrics are named/documented precisely.
- **Revisit trigger:** product requires range-level users/cohorts/true sessions or legal/privacy review changes.

## RADR-005 — Best-effort Beacon-first delivery with optional batching

- **Status/date:** accepted initially; batching 2026-03-18 (`0b11462`); confidence high.
- **Decision:** Beacon whenever present, fetch only if absent; volatile batch queue; page-leave flush.
- **Recorded reason:** README says batching reduces request overhead.
- **Alternatives:** fetch with response/retry, durable IndexedDB queue, service worker, server collector proxy.
- **Benefits:** off critical path, tiny client, improved throughput with batching.
- **Costs:** does not observe acceptance; ignores Beacon false; queue removed before success; duplicate/loss ambiguity.
- **Evidence assessment:** committed browser failures prove consequences.
- **Current assessment:** insufficient for a “trustworthy” analytics contract; roadmap correctly proposes idempotent at-least-once where browser permits.
- **Revisit:** now, before v0.3.

## RADR-006 — Server assigns event ID and ingestion timestamp

- **Status/date:** accepted 2026-01-28; confidence medium.
- **Decision:** overwrite client ID/time.
- **Likely reason:** trust server authority and avoid malformed/missing fields.
- **Benefits:** canonical UUID and UTC ingestion time.
- **Costs:** cannot deduplicate retries or measure occurrence/delivery delay; historical replay changes time.
- **Assessment:** keep receive time but add stable client event ID and occurrence time/schema/SDK version as separate fields.

## RADR-007 — Site/domain compatibility without site registry

- **Status/date:** evolved 2026-02-22; confidence medium.
- **Decision:** aggregate by site ID, fall back/match legacy domain, infer site list from events.
- **Evidence:** commits `a202fa6`, `990bd32`, current `siteMatchClause/GetSites`, tests.
- **Plausible context:** preserve early domain-keyed data while introducing multi-domain logical sites.
- **Benefits:** backwards compatibility/no setup.
- **Costs:** weak tenancy, forged sites, ambiguous domain/site matches, impossible authorization.
- **Assessment:** useful transitional compatibility, not a permanent tenant model.
- **Revisit:** site registration/auth implementation; formal data migration required.

## RADR-008 — Same-origin React/Vite dashboard bundled with server

- **Status/date:** accepted 2026-02-21; confidence high.
- **Decision:** build dashboard into server image; API client uses relative paths.
- **Evidence:** root Dockerfile; `api.ts BASE=""`; FileServer.
- **Benefits:** one analytics deployment, no runtime CORS needed for dashboard, simple operator URL.
- **Costs:** UI/API deploy together, process failure shares boundary, no dashboard CDN/version independence.
- **Assessment:** good current trade-off. Do not split without a real need.

## RADR-009 — Separate marketing deployment

- **Status/date:** accepted 2026-03-25, commit `7a114bd`/PR #1; confidence high.
- **Recorded problem:** README explicitly says prevent marketing/SDK commits rebuilding/redeploying backend.
- **Decision:** separate Docker context/image/watch paths; root `.dockerignore` excludes marketing/web.
- **Benefits:** independent changes and smaller analytics build context.
- **Costs:** separate runtime/config/lockfile; marketing routing must be configured independently.
- **Assessment:** still sound.

## RADR-010 — Permissive CORS after removing allowlists

- **Status/date:** current/unknown rationale, 2026-03-22 `9aa63b2`.
- **Decision:** allow/reflect all origins; removed ingest/dashboard env allowlists.
- **Evidence:** commit diff, current middleware/tests, stale `.env` names.
- **Why:** **Unknown.** Do not infer that debugging convenience was an accepted security trade-off.
- **Benefits:** installation works from any browser origin.
- **Costs:** arbitrary cross-origin reads/writes; future credential hazard.
- **Assessment:** does not make sense for dashboard reads. Replace via security ADR.
- **Questions:** what broke with prior normalization? Which origins need public ingestion? Is dashboard behind external auth?

## RADR-011 — Reliability oracle as a release gate

- **Status/date:** accepted 2026-07-29 (`4cf9a78` through `b3aa0ac`); confidence high.
- **Recorded context:** roadmap requires confidence in numbers and preserved baseline.
- **Decision:** deterministic accepted-manifest reconciliation, faults, browser baseline comparison, quick CI gates.
- **Benefits:** catches silent loss/duplication/aggregate drift and makes known defects explicit.
- **Costs:** complex test subsystem, CI time, baseline maintenance, environment sensitivity.
- **Assessment:** excellent differentiator. Extend to new APIs and keep implementation independent.
- **Revisit:** tune profile duration/tolerances based on CI evidence; never weaken correctness to reduce flakes.

## RADR-012 — Handwritten API types/no formal schema

- **Status/date:** current; confidence medium.
- **Decision:** Go JSON structs and manually duplicated TS interfaces/examples.
- **Benefits:** no codegen/tooling.
- **Costs:** drift already visible in SDK snippets/docs; no runtime validation/versioning.
- **Assessment:** acceptable for prototype, now a change-safety liability.
- **Revisit:** before external API stability/auth/versioning.

## Safe change guides

### Add a dashboard page/view

1. Define whether it is URL-addressable; current enum state is not. If deep links matter, make routing an explicit decision.
2. Add response contract/query first if data is new.
3. Add `DashboardView`, title/nav in `DashboardShell`, component, and render branch in `App`.
4. Keep fetching near the orchestration owner or establish a consistent query layer; do not silently create another N+1.
5. Add loading, empty, partial-error, keyboard, and responsive behavior.
6. Add component/orchestration tests (currently missing), then build and inspect bundle.
7. Rollback is static asset rollback unless API/schema also changes.

### Add an API endpoint

1. Write method/path/auth/site ownership/request/response/error/time semantics.
2. Add core DTO only if it is truly shared; avoid an ever-broader repository interface without reason.
3. Implement repository query with parameterized values and context.
4. Implement handler validation/method enforcement/error mapping.
5. Register route/middleware explicitly.
6. Add DB behavior tests, handler status tests, reliability aggregate/read check, and TS client type/caller.
7. Update canonical API docs and public docs. Verify CORS/auth policy.
8. Query-plan/load test if it scans events. Rollback is binary-only if no schema change.

### Add a database field to `events`

1. Define nullability/default/backfill, old binary compatibility, privacy/retention, index need.
2. Introduce versioned migration infrastructure first; editing `CREATE TABLE IF NOT EXISTS` is not an upgrade.
3. Update schema migration, `core.Event`, insertion SQL/args, lab stored-event verification/manifests, query code, SDK wire contract if applicable.
4. Test fresh DB, old snapshot upgrade, partial failure, backup/restore, and downgrade policy.
5. Deploy migration with backup and single-writer control. Prefer expand/contract if old/new binaries overlap.

### Add a new table

Define owner and relationships; use foreign keys only after confirming enabling `PRAGMA foreign_keys`; add migration/version; repository methods/transactions; indexes/query plans; backup/retention/export; integration/migration tests. Do not infer the table from event traffic if it carries authorization state.

### Add/change authentication

1. Create a formal ADR: self-hosted mode, provider, session/token, user/site schema, proxy trust, recovery.
2. Fix CORS before credentialed cookies.
3. Separate dashboard read auth, admin operations, and public ingestion credentials.
4. Add users/memberships/sites/keys through migrations.
5. Enforce backend ownership on every read/write; UI checks are presentation only.
6. Add CSRF if cookies, secure HttpOnly SameSite cookies, logout/revocation/expiry/audit/rate limits.
7. Security and tenant-isolation tests precede deployment. Provide bootstrap/recovery and rollback without leaving routes open.

### Add a background job

There is no worker framework. First ADR should decide in-process goroutine vs separate command/process, durable state, leader election, retries/idempotency, shutdown, observability. Add an entry point only after defining at-least-once semantics. A timer goroutine in `main` is acceptable only for noncritical single-instance work with documented duplication/loss behavior.

### Add an external integration

Define data sent, credentials, privacy/cost/rate limits, timeout/retry/idempotency, outage behavior, webhook verification, and deletion. Put client behind a small interface, use explicit HTTP timeouts, redact logs, inject secret via runtime platform, add fake-contract tests and fault scenarios. Never block ingestion indefinitely.

### Add an environment variable

Name with `IRIS_` unless it is a conventional runtime variable, centralize parsing/validation, document process/default/secret/failure, add Docker/Compose/platform only where required, test missing/invalid values, and never expose a server secret through Vite. Remove obsolete variables deliberately.

### Change deployment

Inventory actual platform first. Preserve `/app/data` semantics, DB ownership, port, dashboard files, and CGO architecture. Build immutable image, smoke on a copy, back up DB, define health/readiness and rollback, avoid two writers sharing unsupported volume. Update README/Dokploy and CI image verification.

### Upgrade a major dependency

Read primary changelog, update the authoritative manifest/lock (both pnpm and marketing npm if relevant), compile, run tests/quick lab/browser comparison, inspect bundle/API behavior, build containers, and compare baselines. SQLite-driver/Go/React Router/Vite upgrades deserve focused migration/runtime checks.

### Add a business rule

Write it in plain language with examples/edge cases/timezone/tenant scope. Decide authoritative layer: ingestion validation, query definition, or DB constraint. Add independent tests and update duplicated UI/docs. For metrics, preserve/rebaseline historical semantics explicitly—silent number changes are release-blocking per roadmap.

## Recommended future ADRs

| Decision | Trigger/why | Options/evaluation/evidence | Reversibility/risk of early decision |
|---|---|---|---|
| Authentication and site authorization | before public dashboard | built-in local auth, OIDC, trusted reverse proxy; self-host UX, recovery, tenant isolation | schema/API high cost; delaying leaves critical exposure |
| Ingestion authentication/origin model | before trusted analytics | public site key, signed token, allowlisted origin, server proxy; spoof resistance vs browser-secret limits | moderate; pretending browser key is secret is risky |
| Event delivery/idempotency contract | v0.3 | client event ID, durable queue, retry matrix; measured loss/duplication | schema/protocol change; must decide before stable API |
| Versioned schema migrations | next schema change | embedded Go migrator vs external tool; locking/rollback/old snapshots | choice replaceable, migration history not |
| SQLite runtime tuning | after compatible baselines | WAL, busy timeout, connection bounds, synchronous level | reversible PRAGMAs, but durability trade-offs require evidence |
| Time/visitor/session semantics | before metric claims | UTC/site timezone; daily vs range visitors; timed vs tab session | metric history hard to change; collect user requirements |
| URL/property privacy policy | before broad use | path-only default, query allowlist, redaction hooks, click opt-in | safer defaults may break desired attribution; early is better |
| Retention/export/deletion | before valuable/regulated data | global/site TTL, partitions/archive, operator APIs | late cleanup is expensive |
| Supported deployment envelope | after full tests | single container/volume, external Postgres mode, managed offering | keep narrow until demand |
| API versioning/schema generation | before external integrations | OpenAPI/JSON Schema/handwritten; runtime validation/client generation | tooling replaceable; public compatibility is not |
| Observability/SLOs | before production support | Prometheus/OpenTelemetry/structured logs/provider-native | low code reversibility; avoid collecting sensitive labels |
| Marketing rendering/hosting | if SEO/deep links matter | Nginx fallback, static generation, SSR platform | easily reversible; do not overbuild without data |

For every ADR, collect actual traffic/DB size, failure reports, operator constraints, privacy/legal review, hosting capabilities/cost, and rollback experiments. Avoid choosing distributed infrastructure before current SQLite improvements are measured.
