# 08 — Project Learning Path, Question Bank, and Ownership

> **Read/perform · From operator to architect**

## Project-specific curriculum

### Module 1 — Run and observe the complete loop

- **Concept:** processes, ports, HTTP, SQLite.
- **Why here:** Iris is understandable by watching one event travel through every boundary.
- **Files:** `Taskfile.yml`, `cmd/server/main.go`, `dashboard/vite.config.ts`, `pkg/db/sqlite.go`.
- **Prerequisites:** shell, JSON, basic HTTP.
- **Plain explanation:** a browser sends a fact; Go stores a row; dashboard asks Go to count rows.
- **Deep explanation:** distinguish process cwd/path resolution, synchronous request handling, embedded DB connection pooling, and static vs Vite serving.
- **You should answer:** which process owns port 8080/5173? Why are there two possible local DB locations? What makes data durable?
- **Exercise:** run both apps, `curl` a pageview, inspect SQLite, query `/api/stats`.
- **Modification:** change only a disposable event URL and predict every dashboard change.
- **Verify:** match HTTP 202, one DB row, stats/page output, dashboard.

### Module 2 — Browser SDK lifecycle and identity

- **Concept:** browser globals/storage/history/events/Beacon.
- **Why:** collection quality is defined before data reaches Go.
- **Files:** all `web/src`, browser baseline.
- **Prerequisites:** JS classes, DOM events, storage.
- **Beginner:** Iris listens to browser actions and creates anonymous IDs.
- **Deep:** module-global monkey patch, UTC rotation, session clone behavior, Beacon Boolean, volatile queue, start/stop lifecycle.
- **Questions:** what exactly triggers pageviews? What survives refresh/tab close/day change? Why can 202 never be known by Beacon code?
- **Exercise:** replay start, same URL push, replace, stop; compare expected/baseline.
- **Modification:** add a focused unit-test harness before changing behavior.
- **Verify:** browser report exact ordered names/URLs, not only counts.

### Module 3 — Go HTTP request path

- **Concept:** ServeMux, middleware, handlers, contexts, interfaces.
- **Why:** all security/error boundaries live here.
- **Files:** `cmd/server/main.go`, `pkg/api`, `pkg/core`.
- **Prerequisites:** Go functions/interfaces/errors.
- **Beginner:** routes select a handler; middleware wraps it; handler calls storage.
- **Deep:** methodless patterns, MaxBytesReader, decoder behavior, context cancellation, status mapping, `log.Fatal` defer semantics.
- **Questions:** where are IDs/times trusted? Which fields are validated? What methods can a handler receive?
- **Exercise:** call a read with POST and ingestion with malformed/large body.
- **Modification:** add handler tests for 405 before implementing method checks.
- **Verify:** status, `Allow`, no DB mutation, CORS behavior.

### Module 4 — SQL and metric semantics

- **Concept:** grouping, distinct, time filters, indexes, P75.
- **Why:** Iris’s product truth is primarily `pkg/db/query.go`.
- **Files:** schema/query/tests.
- **Prerequisites:** SQL SELECT/GROUP BY/index basics.
- **Beginner:** metrics are different ways of counting filtered event rows.
- **Deep:** site fallback expression, UTC/date conversion, nearest-rank percentile, in-Go aggregation, conversion denominator, query-plan effects.
- **Questions:** why do range visitors differ from sum of daily visitors? Can conversion exceed 100? Which indexes help?
- **Exercise:** construct five rows by hand and calculate every result before querying.
- **Modification:** add an edge-case DB test (empty IDs, boundary timestamp, invalid vital).
- **Verify:** manual expected output equals SQL/repository.

### Module 5 — React dashboard orchestration

- **Concept:** state/effects/async cancellation/rendering.
- **Why:** UI can make correct APIs appear wrong/stale.
- **Files:** `dashboard/src/App.tsx`, `api.ts`, EventsPage/components.
- **Prerequisites:** React hooks and Promise behavior.
- **Beginner:** state chooses site/date/view; effects fetch; props render.
- **Deep:** dependency arrays, abort controllers, `Promise.all` coupling, N+1, duplicated contracts/timezones.
- **Questions:** what happens if request 7/11 fails? Which state remains? How is navigation persisted?
- **Exercise:** block one endpoint in devtools and switch dates rapidly.
- **Modification:** add explicit per-view error UI with a test.
- **Verify:** no stale mislabeled data; aborts are ignored, real errors visible.

### Module 6 — Build, containers, and persistence

- **Concept:** multi-stage images, CGO, volumes, reverse proxies.
- **Why:** an analytics system is valuable only if its DB survives.
- **Files:** Dockerfiles, Compose, `.dockerignore`, README.
- **Prerequisites:** Docker layers/mounts/networking.
- **Beginner:** build tools disappear from final image; volume survives the container.
- **Deep:** CGO ABI, build contexts, mutable tags, bind mounts, one-writer SQLite, external TLS.
- **Questions:** what is in each image? Why exclude web/marketing? What happens without the mount?
- **Exercise:** create data in a disposable mounted container, replace container, verify row.
- **Modification:** add a healthcheck only after implementing meaningful health/readiness.
- **Verify:** image smoke and persistence test.

### Module 7 — Reliability oracle

- **Concept:** deterministic testing, accepted-manifest reconciliation, fault injection, baselines.
- **Why:** this is the strongest mechanism for safely reviewing AI changes.
- **Files:** `internal/reliability`, `testing/load`.
- **Prerequisites:** concurrency, percentiles, test doubles vs real systems.
- **Beginner:** generator knows each event and checks HTTP, DB, and reports agree.
- **Deep:** scheduling lag, rate vs request rate, ambiguous commits, compatible environment comparisons, known browser failures.
- **Questions:** why check accepted rather than attempted events? Why is 1000 batched better than 500 single? Why can CI pass with browser failures?
- **Exercise:** quick smoke and inspect summary/report/server log.
- **Modification:** add a new public aggregate to independent verification.
- **Verify:** intentionally break query and watch oracle fail.

### Module 8 — Security/privacy and tenant boundaries

- **Concept:** identity vs authorization, browser CORS, abuse, pseudonymous data.
- **Why:** current architecture has critical gaps.
- **Files:** routes/CORS/SDK capture/docs/roadmap.
- **Prerequisites:** HTTP origin, sessions/tokens, threat modeling.
- **Beginner:** CORS controls browsers, not who owns data; current server trusts everyone.
- **Deep:** credentialed reflection, public write design, site keys in browsers, data minimization, logs/backups, proxy trust.
- **Questions:** why does an origin allowlist not authenticate curl? Which route leaks all site names? What data can be personal?
- **Exercise:** make a threat model and safely demonstrate unauthorized calls locally.
- **Modification:** write the authentication/ingestion ADR and tests before implementation.
- **Verify:** tenant-isolation/security test matrix.

### Module 9 — Make and defend architectural decisions

- **Concept:** ADRs, trade-offs, reversibility, evidence.
- **Why:** several current choices are good, but motives are partially reconstructed.
- **Files:** chapter 07, Git history, roadmap/baselines.
- **Prerequisites:** all prior modules.
- **Exercise:** defend SQLite, then argue the strongest case against it using measured evidence.
- **Modification:** author one formal ADR with decision triggers and rollback.
- **Verify:** another engineer can distinguish facts, inference, unknowns, and alternatives.

## Comprehension question bank

Stop before the answer section if self-testing.

### Basic

1. What four deliverable concerns exist?
2. Which process writes SQLite?
3. What is the default server port and production host port in Compose?
4. Which event names power pageviews, clicks, and Web Vitals?
5. When does the visitor ID rotate?
6. What makes a custom event?
7. How is a “site” created?
8. Does the dashboard use SSR or client rendering?
9. What does batching change?
10. Where is the current schema defined?
11. Which production background jobs exist?
12. What is the dashboard API base URL?

### Intermediate

13. Trace `iris.track("signup")` to a dashboard custom-event row.
14. Why can range-level unique visitors be inflated across days?
15. What happens if one event in a batch fails insertion?
16. What happens if one of 11 overview requests fails?
17. Why can direct marketing `/docs` fail while clicking Docs works?
18. Why is `IRIS_ALLOWED_INGEST_ORIGINS` ineffective?
19. Explain `siteMatchClause`.
20. Why might timestamp indexes be underused?
21. How is P75 calculated?
22. How does previous-period comparison treat a zero baseline?
23. What does dashboard AbortController solve and not solve?
24. Why is 1,000 batched events/s not the same as 1,000 requests/s?
25. What is the transaction boundary?
26. Which actual endpoint does the lab use as a health check?

### Advanced/trade-off

27. Defend SQLite for Iris’s current product.
28. Identify the first evidence-based triggers for leaving SQLite.
29. Design separate read and ingestion authorization boundaries.
30. Why is server-generated event identity incompatible with idempotent retries?
31. Define a migration from legacy domain/site matching to registered sites.
32. What metric history changes if sessions become 30-minute inactivity sessions?
33. How would you make Web Vital calculation scalable without silently changing P75?
34. Explain ambiguous delivery when commit succeeds but response disappears.
35. Which code belongs in a service layer if one is introduced?
36. What should a readiness check verify without making every request expensive?
37. Which observability labels could create privacy/cardinality problems?
38. When should marketing remain separate, and when might it not?

### Debugging scenarios

39. Dashboard suddenly shows zero sites after deploy. List the first checks.
40. Single ingestion returns intermittent 500 under load; batches work.
41. Events view shows no custom events despite clicks.
42. “Unique visitors” doubled over a two-day window.
43. Dashboard displays stale numbers after one endpoint 500.
44. SDK records duplicate pageviews on navigation.
45. A custom conversion rate is above 100%.
46. Container restart erased data.
47. CI is green but the SDK still loses offline events.
48. Web Vitals score disagrees with a UI threshold.
49. A newly added DB column works on a fresh local DB but production says “no such column.”
50. Marketing home works, `/docs` refresh 404s.

### What-would-happen/code/database/security/deployment

51. What if `batching.maxSize` is 100?
52. What if `stop()` is followed by `start()` on the same instance?
53. What if a caller sends `id` and `ts`?
54. What if screen width is missing?
55. What if `to=2026-07-01` and an event occurs at `23:59:59.500`?
56. Can a remote script read another site’s stats?
57. Can SQL injection occur through `site_id` in current queries?
58. What if `DB_PATH=/new/missing/path/iris.db`?
59. What if `IRIS_LAB_PPROF=1` in production?
60. What if two server replicas use separate local volumes?
61. What if they share a normal network-mounted SQLite file?
62. Which files must change to add `sdk_version`?
63. How would you verify a backup?
64. Which evidence proves the current single-write capacity is unsafe at 500 requests/s?

## Answers

### Basic answers

1. SDK, Go server/SQLite+dashboard image, dashboard, separate marketing site; Reliability Lab is engineering infrastructure.
2. The Go `iris-server`; SQLite is embedded.
3. Server 8080; Compose publishes 8081.
4. `$pageview`, `$click`, `$web_vital`.
5. Once per UTC calendar day in official SDK localStorage.
6. Nonempty event name not starting with `$`.
7. Any ingested row produces/influences inferred `/api/sites`; no registry.
8. React client rendering.
9. Queues in memory and sends `/api/events`; reduces requests and creates a batch transaction.
10. Startup SQL in `pkg/db/sqlite.go`.
11. None.
12. Empty/same-origin; Vite proxies `/api`.

### Intermediate answers

13. SDK payload → transport `/api/event(s)` → handler UUID/time/truncation → event row → `GetCustomEvents` filters non-`$` → EventsPage fetch/render.
14. ID rotates daily; the same browser uses different IDs.
15. transaction rolls back all; client has already removed its queue.
16. `Promise.all` rejects; no dataset updates in that block; previous state may stay visible.
17. BrowserRouter handles client navigation, stock Nginx lacks a fallback rewrite.
18. current code does not read it; allowlist implementation was removed.
19. matches effective site (`site_id` unless empty, else domain) or raw domain for legacy compatibility.
20. functions/OR on indexed columns and absent compound indexes impede straightforward lookups.
21. sort values, nearest-rank `ceil(.75*n)-1`.
22. 0→0 gives 0%; nonzero current over zero gives 100%.
23. cancels prior overview fetches and ignores AbortError; it does not provide caching/timeout/partial errors and has no explicit unmount cleanup.
24. batch size 10 means roughly 100 HTTP requests/s.
25. single INSERT implicit transaction or entire `/api/events` batch explicit transaction.
26. `/api/sites`.

### Advanced answers

27. embedded, cheap, portable, atomic, simple single-container operation; current bottlenecks have untried tuning.
28. measured contention after tuning, multi-writer/replica availability, unbounded scans/size, managed multi-tenant operations.
29. authenticated/per-site dashboard reads; public browser ingestion needs scoped site credentials/origin+rate controls and cannot rely on a secret embedded in JS.
30. every retry gets a new UUID, so DB cannot recognize the logical event.
31. create sites/domains/memberships/keys through migration; backfill from effective IDs; dual-read/write; resolve ambiguity; enforce FK/auth; retire fallback after verification.
32. historic/session counts and conversion denominator change; existing IDs lack event occurrence/inactivity semantics, so backfill may be impossible.
33. preserve nearest-rank definition; use SQL/window functions or maintained distributions/materialization validated against oracle.
34. DB has event, client sees error; retry can duplicate unless stable ID/unique constraint.
35. cross-repository use cases, validation/policy, comparison/metric rules; not thin pass-through indirection.
36. process accepting traffic, DB reachable/writable as appropriate, schema compatible, static files optional depending routing; use bounded checks.
37. raw site IDs, URLs, event names, visitor/session IDs produce sensitivity/high cardinality.
38. separate while release/caching/ownership differ; combine only if operational simplicity outweighs independent lifecycle.

### Scenario answers

39. effective DB_PATH/env, mount, accidental fresh file, permissions, row count, logs, `/api/sites`.
40. SQLite lock contention/default pool/logging; confirmed baseline pattern.
41. `$click` is reserved and excluded; only names without `$` appear.
42. expected daily visitor rotation or storage changes; inspect semantics before declaring bug.
43. all-or-nothing `Promise.all`; add error UI and inspect failed endpoint.
44. same-URL or multiple-instance/history patch defects; compare browser baseline/trace.
45. event-session numerator is not intersected with pageview sessions.
46. missing/wrong volume or DB_PATH; inspect exact file before restore.
47. comparator tolerates known failures; green means no regression vs baseline.
48. thresholds duplicated; compare Go `vitalThresholds` with VitalsPage/META.
49. no migration system; startup `IF NOT EXISTS` did not alter old table.
50. missing Nginx SPA fallback.

### What-would-happen answers

51. client may send 100; server returns 413; batch is already discarded.
52. destroyed transport timer/listeners are not recreated; behavior is incomplete/broken for batching.
53. handler overwrites both.
54. Go zero → stored 0 → Mobile for pageviews.
55. current date-only upper bound uses second precision and may exclude fractional timestamp after 23:59:59.
56. yes if network reachable; arbitrary CORS also permits browser access.
57. no evident injection; parameters are bound. Logic/authorization abuse still exists.
58. literal `data` is created, not custom parent; schema initialization likely fails if parent absent.
59. unauthenticated pprof and lab DB-growth logic activate.
60. analytics split/inconsistent by replica.
61. unsupported contention/filesystem locking risk; do not assume correctness.
62. migration/schema, `core.Event`, insert SQL, SDK payload/config, lab manifest/verification, contracts/docs/tests.
63. `VACUUM INTO`, integrity check, counts/reconciliation, boot restored server, verify APIs.
64. committed target-500 report: 37 500s with “database is locked,” p99/max collapse.

## Interview-style explanations

### Frontend interview

Emphasize two React/Vite clients and the browser SDK: local state, typed fetch wrapper, aborting stale requests, charts, browser storage/lifecycle, and known routing/delivery edge cases. Be honest that dashboard tests/state management are light and SDK history patching needs redesign.

Expected follow-ups: why no query library/router, handling partial errors, bundle size, accessibility, Beacon semantics, multi-tab IDs, SSR safety, React 18 vs 19.

### Backend interview

Emphasize standard Go, repository interface, transactional batch, SQLite aggregate definitions, parameterized SQL, context propagation, and deterministic reconciliation. Discuss absent service layer as a deliberate small-codebase posture, while naming what would justify one.

Follow-ups: locking/pool/WAL, method validation, shutdown/timeouts, migrations, P75, idempotency, schema constraints.

### Full-stack interview

Trace pageview end to end and explain contract duplication, same-origin deployment, date/time/identity semantics, and UI failure coupling. Mention improvements with evidence.

Follow-ups: site registration/auth, API versioning, E2E tests, privacy, feature addition.

### System-design interview

Frame Iris as a single-node event analytics monolith optimized for self-hosting simplicity. State measured envelope carefully and propose an incremental scale path—contract/tuning/retention before distributed infrastructure.

Follow-ups: write amplification, partitions/materialized aggregates, multi-tenancy, availability, exact-once myth, backups.

### DevOps interview

Describe two multi-stage images, persistent volume, CI gates, immutable rollback needs, CGO, and missing deploy/IaC/probes. Do not claim a registry or production process the repo does not show.

Follow-ups: zero downtime, migrations, restore, non-root, image signing, SLO/alerts.

### Security review

Lead with the truth: no application identity or authorization; permissive CORS/public ingestion; pseudonymous/sensitive URL/property risks. Separate current facts from possible external proxy controls.

Follow-ups: browser public keys, CSRF after cookie auth, origin vs authentication, tenant schema, rate limits, data deletion.

### Product-engineering discussion

Emphasize the product goal “boring, trustworthy self-hosted analytics,” current useful dashboards, and why delivery/metric definitions/security precede more features. The Reliability Lab makes prioritization measurable.

Follow-ups: user value of goals/funnels, correctness vs speed, installation verification, privacy claims, roadmap sequencing.

### Role/ownership wording

Say: “I own and operate the repository, and I reconstructed/validated the design with code, history, and reliability evidence. Some early implementation decisions were AI-assisted, so I distinguish the implemented trade-off from motivations that were never recorded.” Do not claim unrecorded personal reasoning.

## Final project ownership checklist

### Architecture and code

- [ ] I can draw browser SDK → Go → SQLite → dashboard and separate marketing/lab from memory.
- [ ] I can name every runtime process, port, data store, and failure boundary.
- [ ] I can trace pageview, custom event, batch, SPA navigation, dashboard refresh, and Web Vital.
- [ ] I can explain actual layers and why there is no service layer.
- [ ] I know which code defines each metric and its edge cases.
- [ ] I can identify unused/generated/stale files without treating them as authority.

### Data and contracts

- [ ] I can recreate the one-table schema and indexes.
- [ ] I can explain site/domain compatibility, visitor rotation, tab sessions, P75, conversion, score.
- [ ] I can list every API, status, DB effect, and caller.
- [ ] I can explain transactions, missing constraints, time semantics, and schema-evolution risk.
- [ ] I can use query plans and the reliability oracle to review metric changes.

### Security/privacy

- [ ] I can separately explain absence of authentication and authorization.
- [ ] I can demonstrate current tenant spoof/read exposure safely.
- [ ] I know why CORS is not authentication and why current credential reflection is dangerous.
- [ ] I can inventory URL/referrer/text/property/log/backup data risks.
- [ ] I will not repeat “no PII/consent unnecessary” as a technical fact.

### Operations

- [ ] I can set up from a clean machine and explain cwd/path differences.
- [ ] I can build/test both images and SDK.
- [ ] I know the actual production platform/config—or have explicitly documented that it remains unknown.
- [ ] I can locate the live DB/volume without guessing.
- [ ] I have performed and timed a backup/restore drill on disposable data.
- [ ] I can deploy/verify/roll back with an immutable image and preserve data.
- [ ] I can diagnose API, DB, latency, deploy, frontend/proxy, and config incidents.

### Reliability and scale

- [ ] I understand attempted vs accepted vs stored vs duplicate/unexpected.
- [ ] I can run quick Lab/browser comparisons and interpret reports.
- [ ] I can explain why current CI can be green with known browser defects.
- [ ] I can state baseline results with machine/revision caveats.
- [ ] I know what to tune/measure before proposing a new database/service.

### Decision making and AI review

- [ ] I can distinguish recorded decision, reconstructed rationale, and unknown.
- [ ] I can defend or challenge each reconstructed ADR with evidence.
- [ ] I can write an ADR with alternatives, impacts, revisit triggers, and rollback.
- [ ] I can add a page/endpoint/field/table/rule safely with tests and deployment plan.
- [ ] I can review AI output for method/auth/tenant/time/idempotency/migration/error/observability contracts.
- [ ] I reject changes that silently alter metrics or weaken oracle correctness.
- [ ] I can identify intentional simplicity versus accidental underengineering.
