# 04 — Authentication, Configuration, and Security Review

> **Security critical**

## Authentication and authorization: the complete flow

There is no authentication flow.

- No identity provider, users table, login/logout route, password, OAuth/OIDC, token, cookie, refresh, provisioning, or synchronization code exists.
- No route is protected.
- No role/permission/resource ownership model exists.
- The dashboard stores/sends no credential.
- All site IDs and event identity values are caller-controlled strings.
- There is no service-to-service authentication or webhook verification because there are no service calls/webhooks.

Consequently, authentication and authorization must be traced separately as:

```text
Authentication: request arrives → no identity established → handler runs
Authorization: no principal/resource check → caller-provided site_id selects data
```

Every operation has missing backend authorization:

- write one or 50 events for any `site_id`;
- enumerate all inferred sites;
- read all stats/custom events/Web Vitals for a guessed or enumerated site;
- access pprof if a production process is accidentally started with the lab flag.

An external reverse proxy/VPN could constrain access, but none is configured in this repository. Do not infer its production presence.

## Configuration reference

Actual secret values were not read into this documentation.

### Production server

| Name | Read at | Required/default | Secret/browser | Failure and risk |
|---|---|---|---|---|
| `PORT` | `cmd/server/main.go → getEnv` | Optional, `8080` | No; server only | Invalid/non-bindable value causes ListenAndServe fatal. No numeric startup validation. |
| `DB_PATH` | same; Docker/Compose | Optional, `./data/iris.db`; image `/app/data/iris.db` | Path, not secret | Parent is not created based on value. Wrong path can create ephemeral DB or fail. Changing volume/path can silently split data. |
| `DASHBOARD_DIR` | same; Docker/Compose | Optional, `./dashboard/dist`; image `/app/dashboard/dist` | No | Missing path yields static 404s while APIs still run. No startup validation. |

### Lab hooks consumed by server

| Name | Purpose | Default/process | Risk |
|---|---|---|---|
| `IRIS_LAB_PPROF` | Enables `/debug/pprof/` and growth-limit handling | unset; iris-server | Exposes sensitive runtime profiles without auth if set publicly. Name also gates DB page-limit reset. |
| `IRIS_LAB_DB_EXTRA_PAGES` | Sets `PRAGMA max_page_count` relative to current pages | unset; only interpreted when pprof flag is `1` | Deliberately induces disk-full behavior; catastrophic to normal writes if misconfigured. |

### Task/Lab variables

| Names | Used by | Purpose/default |
|---|---|---|
| `IRIS_LAB_DB`, `IRIS_LAB_TARGET`, `IRIS_LAB_RATE`, `IRIS_LAB_DURATION`, `IRIS_LAB_BATCH_SIZE` | `task lab:run` | Disposable target/path and load settings; DB required |
| `IRIS_LAB_PROFILES`, `IRIS_LAB_QUICK`, `IRIS_LAB_PPROF` | suite task | Profile list; full durations; pprof true |
| `IRIS_LAB_BASELINE`, `IRIS_LAB_CANDIDATE` | compare task | Both report paths required |
| `IRIS_K6_TARGET`, `IRIS_K6_RATE`, `IRIS_K6_DURATION`, `IRIS_K6_BATCH_SIZE`, `IRIS_K6_READ_RATE`, `IRIS_K6_PROFILE` | k6 task | Existing target and arrival profile |
| `TARGET`, `EVENT_RATE`, `DURATION`, `BATCH_SIZE`, `READ_RATE`, `PROFILE`, `RUN_ID`, `SITE_ID`, `K6_SUMMARY` | k6 script | Container-level equivalents/defaults |
| `IRIS_BROWSER_OUTPUT`, `IRIS_BROWSER_EXECUTABLE` | browser harness | Artifact directory and browser override |

These are non-secret by definition but can target/create large workloads. The Go lab refuses non-loopback targets unless explicit `--allow-nonlocal`; k6 has no equivalent guard.

### Unused/contradictory configuration

- Local `.env` defines `IRIS_ALLOWED_INGEST_ORIGINS` and `IRIS_ALLOWED_DASHBOARD_ORIGINS`, but current code never reads them. These names were implemented then removed in commit `9aa63b2`.
- README and marketing docs omit lab variables, which is appropriate for basic operators but incomplete operational reference.
- No frontend environment variables exist.
- No secrets are required because the system has no auth/integrations; that absence is a security gap, not a secrets-management success.
- Configuration has no typed struct/central validation. Recommended future validation: parse port, resolve/create DB parent, verify dashboard directory, reject lab flags outside an explicit lab mode, and log non-secret effective config.

## Security findings

Severity is impact under a plausible Internet-accessible deployment. Confidence describes repository evidence, not exploit likelihood.

### Critical

#### S-01: all analytics reads and site enumeration are unauthenticated

- **Status:** Confirmed; high confidence.
- **Evidence:** every route in `cmd/server/main.go:47-64`; handlers contain no identity check; dashboard API supplies no credential.
- **Impact:** anyone reaching the server can call `/api/sites`, then read URLs, referrers, traffic, custom product events, and performance by site.
- **Exploitability caveat:** network-layer restrictions may exist outside the repository; unknown.
- **Action:** protect read/dashboard routes with authenticated identities and per-site authorization; keep ingestion authorization a separate design.

#### S-02: arbitrary unauthenticated event injection and tenant spoofing

- **Status:** Confirmed; high confidence.
- **Evidence:** `TrackEvent/TrackBatchEvents` trust JSON `s`, `d`, `u`, IDs, names and properties; no site registry.
- **Impact:** analytics poisoning, unbounded storage/log growth, forged sites, misleading decisions, SQLite contention/availability loss.
- **Action:** register sites, issue scoped ingest keys or verify allowed site/origin, validate fields, rate-limit, and add idempotent client event IDs.

### High

#### S-03: permissive credentialed CORS reflects every Origin

- **Status:** Confirmed; high confidence.
- **Evidence:** `pkg/api/cors.go:13-19`; tests explicitly allow `https://evil.example`.
- **Impact:** arbitrary web origins can read/write APIs from browsers. There are currently no cookies, so credential reflection does not add privilege today, but it makes a future cookie-auth rollout immediately vulnerable unless fixed first.
- **Action:** separate ingestion/read CORS. Allow configured ingestion origins; restrict dashboard origin; never reflect arbitrary origins with credentials.

#### S-04: no rate limit, quotas, or abuse control

- **Status:** Confirmed; high confidence.
- **Impact:** disk exhaustion, DB lock amplification, log flood, CPU/memory pressure. Baseline already shows 500 single writes/s yielding lock failures without malicious load.
- **Action:** proxy and application limits, body/field/batch quotas, site quotas, retention, monitoring.

#### S-05: full URLs/referrers/click text/custom properties can store secrets or personal data

- **Status:** Confirmed capability; data sensitivity depends on consumers.
- **Evidence:** `web/src/index.ts:47-54`; `autocapture.ts:20-30`; untyped properties; server logs URLs.
- **Impact:** query tokens, email/order IDs, visible text, hrefs, or user-supplied properties can enter DB/logs/backups.
- **Action:** strip query/fragment by default or configurable allowlist, make click text opt-in, property allow/deny hooks, document data classification, redact logs.

### Medium

#### S-06: no explicit HTTP server timeouts or graceful shutdown

- **Status:** Confirmed.
- **Evidence:** `http.ListenAndServe` directly.
- **Impact:** slow-client resource exhaustion and acknowledged/in-flight uncertainty on deploy.
- **Action:** configured `http.Server` with header/read/write/idle timeouts, signal-driven shutdown, readiness.

#### S-07: lab pprof and destructive DB-limit controls rely on one environment flag

- **Status:** Confirmed.
- **Impact:** unauthenticated runtime/heap information and induced write failure if enabled in production.
- **Action:** compile/separate lab instrumentation or bind pprof to loopback/admin listener; explicit environment-mode guard.

#### S-08: missing security headers/TLS configuration

- **Status:** Confirmed in application; proxy state unknown.
- **Evidence:** no CSP, HSTS, frame options, nosniff, referrer policy in Go/Nginx config.
- **Impact:** defense-in-depth gaps for dashboard/marketing. Go serves plaintext itself.
- **Action:** set at trusted reverse proxy or application and document TLS termination.

#### S-09: unbounded semantic input and weak JSON constraints

- **Status:** Confirmed.
- **Evidence:** only 1 MiB body, 50 rows, and property-string truncation. No URL/name/key/depth/number range validation.
- **Impact:** malformed analytics, deep/large property structures, stored XSS risk if future UI renders properties unsafely, resource pressure.
- **Current XSS note:** React escapes displayed strings and current UI does not render arbitrary properties as HTML, so no confirmed stored XSS.

#### S-10: container runs as root

- **Status:** Strong inference from Dockerfiles: no `USER` directive in Alpine or Nginx image.
- **Impact:** larger impact if the process/container is compromised; mounted DB is writable.
- **Action:** create non-root user with volume permissions; verify Nginx runtime expectations.

### Low/informational

- No CSRF token exists. With no authenticated browser session, CSRF is not the primary current issue; unauthenticated access is. Revisit with cookie auth.
- SQL injection is not evident: values use parameters. The only formatted SQL is an integer produced by `strconv.Atoi` for a lab PRAGMA.
- SSRF is not present: the server does not fetch caller-provided URLs.
- File upload/path traversal endpoints do not exist. `DASHBOARD_DIR` is operator config, not request data.
- Dependencies are version-ranged in manifests but lockfiles/checksums provide builds. No dependency vulnerability scan/SBOM exists.
- No cloud/CI deployment credentials are used by the visible workflow; permissions are read-only, a strong choice.
- SQLite file permissions inherit process/host defaults; encryption at rest, backup encryption, and least-privilege volume ownership are unknown.

## Privacy assessment

“No cookies” is confirmed. “No personal data” and “most regulations do not require consent” are not repository-provable and can be false depending on URLs, referrers, text, custom properties, localStorage rules, jurisdiction, and operator use. LocalStorage identifiers remain online identifiers/pseudonymous data in many contexts.

Data disclosed to providers:

- Google Fonts receives dashboard browser requests because CSS imports remote fonts (`dashboard/src/index.css`).
- Hosting/DNS/TLS providers necessarily see traffic, but are unknown.
- GitHub/npm receive source/package metadata through development/publishing.
- No analytics events are sent to an application SaaS by current SDK code.

## Security verification checklist

Before public deployment:

1. Confirm network exposure and reverse-proxy rules.
2. Attempt `/api/sites` and a site stats request without credentials; current expected result is 200, which must not remain acceptable.
3. Send an event with another site ID; current expected result is 202.
4. Preflight from an untrusted Origin; current tests expect it to be allowed.
5. Verify pprof is absent.
6. Inspect stored/logged URLs and properties for query secrets/PII.
7. Load-test rate controls and disk quota on disposable infrastructure.
8. Test direct `/docs` on marketing and security headers with `curl -I`.
9. Verify volume ownership, encryption, backup access, and restore authorization outside the repo.
