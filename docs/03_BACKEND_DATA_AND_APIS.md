# 03 — Backend, Domain, Database, and API Contracts

> **Essential · Business semantics reference**
>
> This chapter describes the v2 data architecture. For implementation rationale,
> projection invariants, and scaling triggers, also read
> [09 — v2 data architecture](./09_V2_DATA_ARCHITECTURE.md).

## Backend shape

```text
net/http ServeMux + CORS
        ↓
Handler (transport validation and error mapping)
        ↓
EventRepository
        ↓
SqliteRepository
        ├── serialized write connection
        └── four-connection read pool
```

Iris remains a modular monolith. Handlers call repository operations directly;
there is no separate service layer. The server also runs a maintenance goroutine
for checkpointed projections and retention.

API routes are wrapped in permissive CORS. Ingest endpoints and most read
endpoints do not authenticate callers. `POST /api/sites` is the exception: it
requires `Authorization: Bearer <IRIS_ADMIN_TOKEN>` and returns 503 when the
server has no admin token configured. Site listing and analytics reads remain
unauthenticated.

## Domain rules

### Site and domain

A site is a registered record with a stable ID, name, IANA timezone, retention
period, disable state, and one or more allowed hostnames. `POST /api/sites`
creates or updates it; `GET /api/sites` lists registered sites. Hostnames are
normalized to lowercase without a trailing dot and are unique across sites.

The browser's `site_id` is public identification, not a secret. Ingestion
requires the site to exist and the normalized event URL hostname to be in its
allowlist. The optional payload domain must agree with the URL hostname. This
protects data integrity but does not authenticate the browser or request Origin.

### Event

An event is an immutable raw fact within its retention window. Required client
fields are event ID, event name, URL, site ID, session ID, and visitor ID. The
server validates lengths, reserved event names, width, timestamp skew, schema
version, site, and hostname before persistence.

Client event IDs are unique idempotency keys. Replaying an accepted ID succeeds
without adding a second row. Client occurrence time (`ts`) is retained separately
from server receive time. Supported wire/schema version is currently `v: 1`, and
`sv` records the producing SDK version.

Tracked and referrer URLs must be absolute HTTP(S) URLs without user information.
Iris removes query strings and fragments before storage and extracts pathname and
referrer host into typed columns. Nested property strings are truncated; JSON
properties remain flexible and may still contain user-supplied sensitive data.

### Visitors and sessions

A visitor is a browser-local random ID that rotates at midnight in the SDK's
configured site timezone, which must match the registered site's timezone.
Range-level “unique visitors” is therefore a count of pseudonymous daily IDs,
not people.

The SDK stores its session ID in `localStorage`, shares it across same-origin
tabs, and renews it after 30 minutes without tracked activity or at the UTC
visitor boundary. Session aggregates are also materialized as rebuildable rows.

### Event taxonomy

- `$pageview` drives traffic, page, device, visitor, and session metrics.
- `$click` is reserved autocapture data and is excluded from custom events.
- `$web_vital` carries `$name` and numeric `$val` properties.
- A custom event has a nonempty name that does not begin with `$`.

The only accepted reserved names are `$pageview`, `$click`, and `$web_vital`.
Device class is derived from viewport width: below 768 is Mobile, below 1024 is
Tablet, and all larger widths are Desktop.

## Database architecture

Iris uses `github.com/mattn/go-sqlite3`. Every connection enables foreign keys,
WAL, a five-second busy timeout, and `NORMAL` synchronous mode. Writes are
serialized through one connection; reads use a separate pool capped at four.

Embedded SQL migrations live in `pkg/db/migrations` and applied versions are
recorded in `schema_migrations`. The first migration creates the v2 model and can
upgrade the legacy events-only database transactionally.

### Tables

| Category | Tables | Authority |
|---|---|---|
| Control plane | `sites`, `site_domains`, `ingest_keys` | Registered configuration; ingest keys are reserved for future use |
| Raw fact | `events` | Durable source of truth until retention deletes expired facts |
| Projection | `sessions`, `daily_site_metrics`, `daily_page_metrics`, `daily_referrer_visitors`, `daily_visitors`, `daily_sessions` | Rebuildable derived state |
| Operations | `schema_migrations`, `projection_checkpoints` | Migration history and ordered projection progress |

The raw event row has an integer `seq` for projector order and a separate unique
client `id`. It stores event/site identity, occurrence/receive microseconds,
normalized URL dimensions, stable site-local day, session/visitor IDs, JSON
properties, and wire/SDK versions. Compound indexes cover site/time and
site/name/day query shapes plus session and visitor variants.

The projector commits derived changes and checkpoint advancement atomically. It
drains pending batches at startup and polls every 250 ms. Rebuilding clears only
derived tables, resets the versioned checkpoint, and replays raw events.

Retention runs at startup and every 24 hours. It deletes expired raw events,
sessions, and daily projections according to each site's
`retention_days`. Backups and file-space reclamation remain operator concerns.

## API conventions

- Base path: `/api` on the Go server.
- Ingestion consumes JSON; successful acceptance is empty `202`.
- Read endpoints return JSON; errors are plain text.
- Event bodies are limited to 1 MiB; batches contain at most 50 events.
- Analytics queries require `site_id`; the `domain` query name remains a legacy
  alias for that value.
- Date-only windows are interpreted in the registered site's timezone. Explicit
  timestamps are interpreted as UTC/RFC3339 values.
- CORS reflects a supplied Origin and allows credentials. CORS is not access
  control.
- Browser ingestion rejects an `Origin` whose hostname differs from the tracked
  event hostname. Non-browser clients can still spoof request fields, so
  deployment-level rate limiting remains necessary.

## API catalogue

| Method/path | Purpose | Important behavior |
|---|---|---|
| POST `/api/sites` | Register/update site | Requires admin bearer token; body has `site_id`, `name`, `timezone`, `retention_days`, `domains`; returns 201 |
| GET `/api/sites` | List site records | Currently unauthenticated |
| POST `/api/event` | Ingest one event | Validates and normalizes; idempotent by client `id`; returns 202 |
| POST `/api/events` | Ingest batch | Maximum 50; one atomic transaction; returns 202 |
| GET `/api/stats` | Pageviews, unique visitors, sessions | Raw pageview aggregates |
| GET `/api/site-trends` | Current/previous stats and changes | Equal-duration previous period when dates are supplied |
| GET `/api/pages` | Top paths | Up to 10 |
| GET `/api/referrers` | Top referrer hosts | Distinct visitor IDs |
| GET `/api/vitals` | P75 LCP/INP/CLS | Nearest-rank P75 |
| GET `/api/vitals/distribution` | Vital quality buckets | Good/needs-improvement/poor |
| GET `/api/vitals/pages` | Per-path vitals and traffic | Up to 20 |
| GET `/api/vitals/score` | Overall and per-metric score | Current Iris scoring formula |
| GET `/api/custom-events` | Custom-event summary and rows | Non-reserved names |
| GET `/api/custom-events/timeseries` | Daily selected-event volume | Requires `event_name` |
| GET `/api/devices` | Viewport device classes | Pageviews only |
| GET `/api/timeseries` | Daily pageviews | Site-local date semantics for date-only windows |
| GET `/api/timeseries/visitors` | Daily distinct visitor IDs | Daily pseudonymous identity |
| GET `/api/timeseries/sessions` | Daily distinct session IDs | SDK session identity |

Analytics reads still query raw events where exact or not-yet-projected answers
are required. Projection availability therefore does not make dashboard results
eventually consistent.

## Current limitations

- Dashboard reads, site listing, and event ingestion do not yet have user/site
  authorization; only site mutation has an admin bearer token.
- The SDK queue is memory-only and does not retry failed or rejected delivery.
- CORS remains permissive and there is no rate limit.
- Go and TypeScript response types are manually duplicated; there is no OpenAPI
  or runtime response validation.
- Arbitrary event properties and click text can contain sensitive data despite
  URL query minimization.
- Backups, restore drills, metrics, and readiness are not scheduled production
  services.

## Historical evolution

The original 2026 schema used one loosely constrained `events` table, inferred
sites from event strings, generated event identity/time on the server, opened an
untuned default SQLite pool, and edited startup `CREATE TABLE IF NOT EXISTS` DDL.
The 2026-07-29 reliability baseline measured that architecture and remains useful
historical evidence; its lock failures and browser delivery defects are not a
description of every current behavior.

The v2 migration deliberately preserves legacy rows: it creates registered site
and domain records from the old values, gives occurrence and receive time the
legacy timestamp, and then removes the temporary legacy table. This is the
supported upgrade path; future schema changes must add numbered migrations.
