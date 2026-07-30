# 09 — Evidence and Coverage Report

> **Reference · Investigation audit**

## Scope and method

The investigation followed actual entry points, imports, handler calls, repository methods, SQL, frontend callers, build/deployment definitions, tests, Reliability Lab verification, committed baseline results, and Git history. Existing prose was treated as evidence of stated intent, not as proof of behavior.

Evidence snapshot:

- branch: `codex/marketing-redesign`;
- HEAD before documentation: `965cbc7` (2026-07-29);
- worktree: clean before handbook files;
- inspection date: 2026-07-30;
- local DB contents were inspected only for schema and aggregate row/site/time counts; no secret values or event payload content were copied into docs;
- `.env` variable **names** were inspected and values redacted.

## Inspected application/runtime sources

- `cmd/server/main.go`
- `pkg/core/event.go`
- `pkg/api/handler.go`, `cors.go`, all API tests
- `pkg/db/sqlite.go`, `query.go`, query tests
- every `web/src/*.ts`, package manifest, SDK README and generated build metadata
- dashboard bootstrap, App/API, every component, global CSS/build config/index/package manifest
- marketing bootstrap, router, Home/layout, Docs content, CSS/build/lint/package files, Dockerfile/template README

## Inspected build/config/infrastructure

- root README, architecture summary, roadmap, Taskfile
- Go and JS manifests/checksums/lockfiles/workspace definitions
- root and marketing Dockerfiles, Compose, Docker ignore files
- Git ignore and local environment variable names/usages
- complete GitHub Actions workflow
- tracked/generated/ignored artifact inventory
- actual local SQLite `.schema`, index definitions, and non-sensitive counts
- direct Go and pnpm dependency listings

## Inspected reliability/testing

- every Go test name and relevant bodies
- `internal/reliability` types, manifests, load execution, read load, storage/public aggregate verification, resource sampling, reporting, suites, process lifecycle, faults, backup/restore, comparison
- browser harness scenario inventory, comparator tests, React Router fixtures
- k6 ingestion/read script and browser probe presence
- Reliability Lab handbook
- committed 2026-07-29 suite, target-500, target-1000, and browser reports/summaries

## Inspected history

Complete commit list and per-file history, plus focused diffs for:

- monorepo/backend/SQLite origins;
- site/domain evolution;
- batching;
- CORS allowlists and their removal;
- Compose volume change;
- daily visitor rotation;
- split Docker deployables;
- reliability roadmap/lab/CI;
- current reporting APIs/dashboard.

Only two PR merge commits were locally evident. Remote issue/PR bodies, review discussions, branch protection, releases/tags, and deployment-provider history were not available from local repository evidence and were not invented.

## Verification performed

| Check | Result |
|---|---|
| `go test -race ./...` | Pass; non-fatal macOS CGO linker warnings |
| browser comparator unit tests | 5 pass |
| SDK production build | Pass |
| dashboard TypeScript/Vite build | Pass; >500 kB chunk warning |
| marketing ESLint | Pass |
| marketing TypeScript/Vite build | Pass |

Not rerun:

- full browser oracle: committed report inspected; currently 19/35 pass and command is intentionally allowed to produce known failures in CI before comparison;
- full/quick load/fault suites: committed full-duration evidence and source/tests inspected; rerunning was unnecessary for documentation and would create substantial generated load artifacts;
- Docker image builds/runtime: Docker evidence inspected but daemon execution was not required to establish code architecture;
- production deployment/backup: no production authority/config was available.

## Required-section coverage

| Request section | Location |
|---|---|
| 1 Executive overview/read order | handbook index |
| 2 Repository map | chapter 01 |
| 3 System context/architecture/trust boundaries | chapter 01 |
| 4 Entry points/startup/shutdown | chapter 01 |
| 5 End-to-end journeys/edge cases | chapter 02 |
| 6 Frontend architecture | chapter 02 |
| 7 Backend architecture | chapter 03 |
| 8 Domain model/rules/glossary | chapter 03 |
| 9 Database/schema/evolution/ERD | chapter 03 |
| 10 APIs/contracts | chapter 03 |
| 11 Authentication/authorization | chapter 04 |
| 12 Configuration/env reference | chapter 04 |
| 13 Dependencies/integrations | chapter 01 |
| 14 Deployment/infrastructure | chapter 05 |
| 15 CI/CD | chapter 05 |
| 16 Testing strategy | chapter 06 |
| 17 Error/resilience | chapter 05 |
| 18 Observability/debug playbooks | chapter 05 |
| 19 Security review | chapter 04 |
| 20 Performance/scalability | chapter 06 |
| 21 Maintainability | chapter 06 |
| 22 Technical-debt register | chapter 06 |
| 23 Unknowns/contradictions | chapter 06 |
| 24 Safe-change guides | chapter 07 |
| 25 Local development | chapter 05 |
| 26 Operations runbook | chapter 05 |
| 27 Learning path | chapter 08 |
| 28 Questions and separate answers | chapter 08 |
| 29 Interview explanations | chapter 08 |
| 30 Future ADRs | chapter 07 |
| 31 Ownership checklist | chapter 08 |

## Final self-review

- [x] Re-scanned tracked and hidden repository files; ignored build/local state categorized.
- [x] Documented every production runtime component and explicit absence of workers/cron/queues.
- [x] Connected the sole table to ingestion and every query workflow.
- [x] Connected all registered APIs to caller/data effects; pprof/static routes included.
- [x] Traced authentication and authorization separately.
- [x] Limited deployment claims to Docker/Compose/README/CI evidence.
- [x] Marked all ADRs reconstructed; unknown motivations remain unknown.
- [x] Used file paths, symbols, lines, commits, tests, and reports for significant claims.
- [x] Dedicated unknown/contradiction register.
- [x] Project-specific curriculum/exercises/question answers.
- [x] Safe-change and AI-review checkpoints included.
- [x] Every diagram has adjacent text explanation.

## Evidence limitations requiring owner follow-up

1. Export sanitized actual Dokploy/reverse-proxy/DNS/TLS/volume/backup configuration.
2. Record production SQLite schema/version/size and whether it originated before current columns.
3. Document who can access the dashboard/server network today.
4. Provide production logs/metrics/incident history without personal data.
5. Confirm npm and container publication/release mechanism and currently supported versions.
6. Decide product metric/time/privacy contracts listed in roadmap.
7. Capture original intent for CORS removal and site/domain compatibility if the author remembers it.
8. Define RPO/RTO/retention/deletion/legal and support policies.

Until these exist, production security, availability, cost, data residency, recovery readiness, and historical decision motivation remain partly or wholly unknown.
