# 06 — Testing, Performance, Maintainability, Debt, and Unknowns

> **Important · Engineering assessment**

## Testing map

### What exists

| Layer | Evidence | What it proves |
|---|---|---|
| API helper unit | `pkg/api/analytics_test.go` | previous-period duration and percentage edge cases |
| CORS unit | `pkg/api/cors_test.go` | arbitrary-origin reflection, credentials, preflight, wildcard |
| SQLite integration | `pkg/db/query_test.go` | site/domain grouping, devices, referrer normalization, P75, custom events, scores, date/time filters |
| Reliability unit/integration | `internal/reliability/reliability_test.go` | deterministic manifests, single/batch reconciliation, concurrent reads, stages, comparisons, missing-event detection, target safety |
| Full real-server correctness/load | `internal/reliability`; Task profiles | accepted rows/fields/public aggregates, latency/resources |
| Fault/recovery | `internal/reliability/faults.go` | restart, 503/network, lock/full, backup/integrity/restored boot |
| Browser system tests | `testing/load/browser/run.mjs` | SDK/navigation/storage/delivery/framework/lifecycle in Chromium |
| Browser comparator unit | `compare.test.mjs` | baseline regressions/improvements/coverage/harness detection |
| Independent load | `testing/load/k6` | arrival-rate writes and reads, basic thresholds |
| Build/static analysis | CI | Go race/vet/build, TS builds, marketing ESLint |

### What does not exist

- direct handler tests for method enforcement, body limits/status, field normalization/truncation, repository failures, or batch atomicity;
- SDK unit tests for storage, payloads, config validation, cleanup, or transport;
- dashboard component/state/API/error tests;
- marketing route/content tests;
- cross-browser Safari/Firefox/mobile tests;
- automated API contract/OpenAPI tests;
- migrations/schema upgrade tests (no migration system);
- auth/security tests (no auth);
- accessibility/visual regression tests;
- container image/startup/Compose tests;
- static security/dependency/secret/license scans;
- production infrastructure tests (no IaC);
- retention/export/deletion tests (features absent).

### Business rule/workflow coverage

| Rule/workflow | Coverage | Gap |
|---|---|---|
| Initial pageview | Browser PASS baseline | Cross-browser and backend row integration separated |
| SPA navigation | Browser detailed baseline | Many confirmed failures accepted as known |
| Manual/batch delivery | Lab + browser | No durable/retry/idempotency behavior to assert |
| Daily visitor rotation | Browser identity scenarios/readme | No focused deterministic UTC-boundary unit test found |
| Session identity | Browser; duplicate-tab failure | Semantics unsettled |
| Aggregate stats/pages/referrers/devices/series | DB tests + independent lab oracle | New reporting APIs not all in lab `readEndpoints`/aggregate checks |
| Custom events/performance | DB tests | Handler/frontend E2E limited; lab predates new endpoints |
| Site/date switching | No React tests | Abort/stale/all-or-nothing behavior unprotected |
| Deployment/restore | Lab restore only | No production image/volume drill in CI |

### Test quality and flakiness

DB tests exercise behavior with real temporary SQLite and are strong. The oracle independently derives expected values and checks raw storage/public APIs, a particularly strong safeguard for AI-generated changes.

Risks:

- browser scenarios depend on Chromium lifecycle/network behavior and preserve known failures rather than requiring green;
- quick load profiles can vary on hosted runners; commit history includes `ci: stabilize fault gate`;
- full baselines were captured on a modified M4 worktree at revision `4cf9a78`, so they are not directly comparable to every environment;
- resource sampling has OS-specific behavior;
- marketing lint/build proves compilation, not correctness.

### Prioritized missing tests

1. **P0 security contract:** unauthorized read/write/site enumeration must fail after auth/site registration is designed.
2. **P0 ingestion contract:** required fields/ranges/origin/site, method 405, body/batch limits, idempotent replay, occurrence vs receive time.
3. **P0 SDK delivery:** accepted/rejected Beacon Boolean, retryable response matrix, durable queue conditions, lost-response dedupe.
4. **P0 migration upgrade/rollback:** old schema snapshots through each migration.
5. **P1 dashboard orchestration:** site/date races, one endpoint failure, empty state, URL-deep-link contract.
6. **P1 new analytics oracle:** trends/custom events/distributions/pages/score under deterministic load.
7. **P1 production image smoke:** build, non-root, mounted persistence, health/readiness, static dashboard.
8. **P2 cross-browser/accessibility/visual:** Firefox/Safari-relevant lifecycle, keyboard/screen-reader, screenshots.
9. **P2 privacy:** query stripping, click allow/deny, property redaction, log redaction.

## Verification snapshot

Executed 2026-07-30:

- `go test -race ./...` — pass; macOS linker emitted non-fatal malformed `LC_DYSYMTAB` warnings for CGO test binaries.
- `pnpm test:browser-oracle` — 5/5 comparator tests pass.
- SDK build — pass; output about 4.78 kB ESM/5.24 kB CJS.
- Dashboard build — pass; 585.85 kB JS/168.24 kB gzip with Vite chunk-size warning.
- Marketing lint/build — pass; 398.47 kB JS/125.42 kB gzip.

The full real-browser oracle was not rerun because its committed baseline intentionally has 16 known failures and requires browser system execution; its report was inspected. Full-duration load tests were not rerun because they take many minutes and capacity evidence is already committed.

## Performance and scalability

### Confirmed baseline

On an Apple M4/10 logical CPU/16 GiB, Go 1.25.3, revision `4cf9a78` modified:

- 500 events/s as 500 single HTTP writes/s: 149,963/150,000 accepted; 37 SQLite lock 500s; p95 3 ms, p99 580.52 ms, max 6.32 s; FAIL.
- 1,000 events/s in batches of 10 (100 requests/s): 300,000 accepted; p95 4.10 ms, p99 46.42 ms, max 2.28 s; PASS.

Both stored every acknowledged event exactly once and aggregates matched. The report explicitly says this is not a production capacity guarantee (`testing/load/baselines/2026-07-29`).

### Current bottlenecks

1. SQLite write/connection contention: default pool, rollback journal, no busy timeout/WAL; confirmed lock failures.
2. Per-request synchronous logging at high ingestion rates.
3. Best-effort browser enqueue: 1,000 immediate events produced only 270 actual events in baseline, so client delivery fails before server capacity.
4. Query expressions and missing compound indexes.
5. In-memory referrer/vital/page aggregation and sorting.
6. Repeated dashboard queries and Sites N+1.
7. Ever-growing single table/no retention.
8. Dashboard bundle and eager charting.

### Scaling path (estimates)

Assumptions: one server, local SSD, properly mounted SQLite, current code, traffic measured as events rather than people.

- **Small number of users/sites:** likely adequate; operational/security mistakes and event semantic defects fail before compute.
- **Hundreds of active visitors:** normally fine if event rates are modest and batching works; browser loss and query/privacy correctness remain limiting.
- **Thousands of active visitors:** burst rate, many single Beacon requests, lock errors, log I/O, and unindexed scans become material. Batch configuration and storage growth matter.
- **Much larger scale:** one process/file is a write and availability boundary; read scans compete with writes, backup/retention becomes expensive, and horizontal replicas cannot safely share ordinary local SQLite.

Do not jump immediately to distributed services. First: define contract; validate/batch/idempotency; bound SQLite connections; WAL/busy timeout; compound indexes based on plans; sample logs; retention; observability; publish measured envelope. Consider a server DB/queue only when measured workload, multi-instance availability, or operations require it.

### Single points/state

- one Go process;
- one SQLite file/volume;
- no replicated queue;
- browser-only volatile queue;
- no caching/CDN config for dashboard;
- external reverse proxy/DNS unknown.

## Code-quality assessment

### Strong choices

- Small, traceable codebase using standard Go HTTP and repository interface.
- Parameterized SQL and context-aware database methods.
- Atomic batch insert.
- Server-owned UUID/ingestion time avoids trusting those client fields, although it blocks idempotency/occurrence semantics.
- Deep-copy property truncation avoids caller mutation (commit `3291446`).
- AbortController mitigates stale overview requests.
- Exceptionally strong deterministic reliability/fault/browser evidence and CI aggregation.
- Separate marketing deployment prevents unrelated analytics redeploys.

### Accidental complexity/weak boundaries

- `pkg/db/query.go` mixes SQL, domain definitions, normalization, scoring, and in-memory algorithms in ~770 lines.
- Handler has repeated query/error/write blocks without a consistent method/validation/error abstraction.
- `core` looks like a domain layer but is mostly JSON DTOs and a broad repository interface.
- UI duplicates all response types and thresholds.
- `App.tsx` orchestrates site/date/navigation plus 11 datasets.
- EventsPage embeds a factually incorrect SDK snippet.
- Marketing has a parallel public-doc source, root README, SDK README, old architecture doc, and roadmap—high documentation drift.
- Two JS package managers/locks are intentional for Docker isolation but increase upgrade complexity.
- Tracked generated TypeScript artifacts and unused root tools create noise.
- Synthetic UI content (site health, SparkBars, marketing metrics) looks operationally real.

### Framework/design misuse

- BrowserRouter with stock Nginx has no documented fallback.
- SDK history monkey-patching is incomplete and global.
- Transport creation/start/stop lifecycle is inconsistent.
- Default Go server lacks production controls.
- Static FileServer does not provide SPA fallback, although dashboard currently uses no URL routes.
- `Promise.all` creates coupled failure despite independent widgets.

### Hidden side effects

- constructing Transport can start timers/add listeners;
- `start()` mutates global `history.pushState`;
- server startup can mutate schema and, in lab mode, DB max pages;
- `/api/sites` changes solely because any event is inserted;
- date preset semantics depend on operator timezone;
- marketing content makes unsupported product/legal claims.

No circular imports were identified. React components import `DATE_PRESETS` from App, which is reverse-ish coupling but not a runtime cycle in current graph.

## Technical-debt register

| ID/type | Evidence/impact | Risk/urgency/difficulty | Recommendation |
|---|---|---|---|
| TD-01 Security bug | no auth/authorization | Critical / now / high | Formal ADR then identity + per-site read controls |
| TD-02 Security/abuse | arbitrary ingestion/site spoofing/rate | Critical / now / high | site registry, scoped ingestion, validation, quotas |
| TD-03 Reliability bug | browser baseline 16 failures | High / now / high | roadmap v0.3 delivery contract/idempotency |
| TD-04 Data architecture | no migrations; startup DDL | High / before schema change / medium | versioned migration tool and old-schema fixtures |
| TD-05 Operations | no shutdown/timeouts/health | High / before production / medium | explicit server, signals, probes |
| TD-06 Performance | DB defaults/lock failures | High / measured now / medium | pool=bounded, busy/WAL ADR and baseline comparison |
| TD-07 Privacy | full URL/text/properties/logs | High / now / medium | safe defaults/redaction/data policy |
| TD-08 Contract | no field validation/idempotency/version | High / now / high | JSON schema/validation and event ID/time fields |
| TD-09 Backup/retention | only lab proof, no schedule | High / before valuable data / medium | RPO/RTO, automated encrypted backups/restore drills |
| TD-10 Testing | new APIs absent from full oracle | Medium / soon / medium | extend deterministic manifest/checks |
| TD-11 Frontend reliability | all-or-nothing overview, console errors | Medium / soon / medium | per-widget settled results/error UI/retry |
| TD-12 UI truth bug | broken SDK snippet; synthetic states | High product / now / low | generate/reuse real SDK examples; label/remove mock data |
| TD-13 Contract duplication | Go/TS/docs types | Medium / later / high | choose OpenAPI/schema generation after API versioning ADR |
| TD-14 Query scale | expression filters/in-memory scans | Medium / measured / medium | explain plans, compound indexes, SQL aggregation |
| TD-15 Bundle | 586 kB dashboard | Low-medium / later / low | lazy views/chart library analysis |
| TD-16 Deployment | no image CI/deploy/IaC/digests | Medium / before managed prod / medium | image build/scan/sign and documented promotion |
| TD-17 Maintainability | unused/tracked generated files/tools | Low / cleanup / low | confirm then remove/ignore in focused change |
| TD-18 Documentation | five drifting public sources | High product / now / medium | canonical contract, generated excerpts, doc checks |
| TD-19 Product limitation | no site management/users/export/retention | Deliberate early-stage | do not “fix” until product/security ADRs |
| TD-20 Date semantics | daily rotation/UTC vs local date presets | Medium correctness / soon / medium | site timezone/time-contract ADR and tests |

“Leave alone for now”: modular monolith, SQLite, plain `net/http`, local React state, and no distributed queue are reasonable at current demonstrated scope. Revisit based on explicit triggers, not fashion.

## Unknowns and contradictions

### Production topology

- **Question:** where/how is Iris actually hosted?
- **Inspected:** Dockerfiles, Compose, README Dokploy notes, CI, Git branches/history.
- **Evidence:** suggested Dokploy split only.
- **Unknown:** region, domains, TLS, registry, replicas, resources, volume/backups, access controls, deploy trigger.
- **Danger:** high; security/DR/availability cannot be assessed.
- **Resolve:** export sanitized platform config/IaC/runbooks.

### Historic database upgrades

- Current schema has site/visitor columns but no migrations. Startup `IF NOT EXISTS` cannot alter old tables.
- Existing deployments’ creation version/manual changes are unknown.
- Resolve by inventorying production `PRAGMA table_info/index_list/user_version` read-only and creating migration baselines.

### CORS intent

- Git history implemented allowlists then removed them (`ceaa1de` through `9aa63b2`), while local `.env` still names them.
- Motivation for removal is not recorded; commit title only says remove checks.
- Treat current permissiveness as code fact, not a reasoned accepted security decision.

### Documentation contradictions

- README says “Next.js/Vite compatible,” which is usage compatibility, not a Next.js component.
- Root quickstart says Compose creates `./data`; current mount is `../files/iris-data`.
- Dashboard snippet uses nonexistent package/import/config.
- Marketing Docker example references `ghcr.io/vatsalp117/iris:latest`; no image-publish workflow proves it exists/current.
- Marketing preview displays bounce rate, which backend does not calculate.
- Marketing says first-party storage is inaccessible to third-party scripts; scripts running on the same origin can generally access it.
- Public consent/legal guidance cannot be established by code.
- Old `iris_architecture.md` defaults/autocapture and file links are stale.

### Tests vs product behavior

- CI comparator accepts known browser failures; green CI does not mean browser oracle fully passes.
- CORS tests intentionally codify access from `evil.example`.
- Reliability read endpoints do not include newly added reporting endpoints.

### Unknown operational policies

RPO/RTO, retention/deletion requests, legal basis, support access, log rotation, dependency cadence, incident ownership, branch protection, production SLOs, secret rotation, cost, DNS/TLS, and data residency cannot be determined.
