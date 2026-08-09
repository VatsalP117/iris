# Iris v2 Data Architecture

## Purpose

Iris uses a relational control plane, an immutable analytics fact log, and
rebuildable read models. This keeps the self-hosted deployment small while
preserving a clean migration path to a dedicated analytical database if the
workload eventually demands one.

The core invariant is:

> An accepted event is durable in the raw `events` table. Sessions and daily
> metrics are derived state and may be discarded and rebuilt.

## Data flow

```text
browser SDK
    -> validate and normalize
    -> append raw event
    -> project events in seq order
       -> sessions
       -> daily site/page/referrer data
```

The control-plane tables are `sites`, `site_domains`, and `ingest_keys`. Site
mutation is protected by a server-side admin bearer token. The `ingest_keys`
schema is reserved for a later browser credential flow; current browser ingestion
identifies a site with the public `site_id` and relies on hostname validation
rather than a secret embedded in client code.

## Register a site before tracking

`POST /api/sites` creates or updates a site. It requires `Authorization: Bearer
<IRIS_ADMIN_TOKEN>`; if `IRIS_ADMIN_TOKEN` is unset, mutation is disabled with a
`503` response. At least one hostname is required.
Hostnames are lowercased, trailing dots are removed, duplicates are collapsed,
and a hostname can belong to only one site.

```bash
curl -X POST http://localhost:8080/api/sites \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer replace-with-your-IRIS_ADMIN_TOKEN' \
  -d '{
    "site_id": "docs",
    "name": "Documentation",
    "timezone": "Asia/Kolkata",
    "retention_days": 365,
    "domains": ["docs.example.com"]
  }'
```

Defaults are the site ID for `name`, `UTC` for `timezone`, and 365 for
`retention_days`. Timezones must be valid IANA location names. `GET /api/sites`
returns the registered sites and their domains.

The SDK's `siteId` must match the registered `site_id`. During ingestion Iris
derives the hostname from `u`; if `d` is supplied it must match that derived
hostname. An unknown site returns `404`, and a hostname outside the site's
allowlist returns `403`.

The SDK's `timezone` option must also match the registered IANA timezone.
Visitor IDs rotate at midnight in that timezone; session IDs are isolated per
site and roll only after 30 minutes without tracked activity. Once a site has
events, Iris rejects timezone changes so historical day and identity semantics
cannot silently split.

Site IDs and browser-visible ingest identifiers are not secrets. The admin token
must remain server-side and should be a long random value. `GET /api/sites`,
analytics reads, and event ingestion remain unauthenticated; authorization for
those surfaces is still a separate concern.

## Event wire contract

The SDK sends an individual event to `POST /api/event`, or up to 50 events as a
JSON array to `POST /api/events`:

```json
{
  "id": "4a8c7ff0-8ace-4b2c-9c32-c9723b408623",
  "n": "$pageview",
  "u": "https://docs.example.com/guide?campaign=private#install",
  "d": "docs.example.com",
  "r": "https://search.example/results?q=iris",
  "w": 1440,
  "s": "docs",
  "sid": "d209545c-b823-42cb-832d-073abbdf8ef5",
  "vid": "8a12a39e-4297-464b-a929-90b8d09fb855",
  "p": {},
  "ts": "2026-08-09T10:30:00.000Z",
  "v": 1,
  "sv": "1.0.0"
}
```

The current contract includes four fields with distinct responsibilities:

| Field | Meaning |
|---|---|
| `id` | Client-generated idempotency key. Duplicate IDs are accepted without inserting a second raw event. |
| `ts` | Client occurrence time in ISO 8601 form. It defaults to receive time when omitted and cannot be more than five minutes in the future. |
| `v` | Wire/schema version. Omitted means version 1; other versions are currently rejected. |
| `sv` | SDK version used to produce the event. It is optional and limited to 64 characters. |

`id`, event name, site ID, session ID, and visitor ID are required and limited to
128 characters. Reserved event names are `$pageview`, `$click`, and
`$web_vital`; custom names must not begin with `$`.

### URL and property handling

Tracked and referrer URLs must be absolute `http` or `https` URLs without user
information. Before storage Iris lowercases the scheme and host, removes the
query string and fragment, and supplies `/` for an empty path. The normalized
pathname and referrer hostname are also stored as typed columns for common
queries. This means the example event stores
`https://docs.example.com/guide`, not its campaign query or fragment.

Custom properties remain JSON so new event types do not require migrations.
String property values are capped at 200 characters recursively. Frequently
queried, stable dimensions should become typed event columns in a future
migration rather than an entity-attribute-value table.

## Storage model

### Control plane

- `sites` owns the stable public site ID, display name, IANA timezone, retention
  setting, and disable state.
- `site_domains` is the hostname allowlist and identifies the primary hostname.
- `ingest_keys` reserves hashed, revocable ingest credentials for a future API.

### Raw fact log

`events` is append-only application data and the source of truth. Its integer
`seq` primary key provides deterministic projector order; the independent `id`
is unique for retry-safe ingestion. Both `occurred_at_us` and `received_at_us`
are stored as UTC integer microseconds. Common filters have compound indexes by
site and occurrence time, with variants for event name, session, and visitor.
Each row also stores `local_day`, calculated from occurrence time using the
site's timezone at ingestion. The timezone becomes immutable after the first
event, keeping historical and future calendar-day semantics consistent.

Do not update raw events to correct a dashboard metric. Change the projection
logic, bump its version, and rebuild derived state. Raw-event deletion is
appropriate only for an explicit retention or privacy operation.

### Derived projections

The analytics projection contains:

- `sessions`, keyed by site and SDK session ID, with first/last timestamps,
  entry and exit pathnames, initial referrer, pageview/event counts, and bounce
  state;
- `daily_site_metrics`, containing pageview and custom-event counts in the
  site's configured timezone;
- `daily_page_metrics`, containing pageviews by site-local day and pathname;
- `daily_referrer_visitors`, retaining distinct visitor keys needed to answer
  daily referrer visitor counts;
- `daily_visitors` and `daily_sessions`, retaining exact site-local distinct
  sets for daily visitor and session charts;
- `projection_checkpoints`, recording the last raw `seq` and projection version.

The background projector reads a bounded batch strictly after its checkpoint.
Derived writes and the checkpoint advance commit in one transaction, so a failed
batch can be retried without double counting. Sessions touched by a batch are
recomputed from their raw events through that batch's final sequence, which also
handles events received out of occurrence-time order. Projection lag never
changes whether an event was accepted: durability is established by the raw
event transaction first. The server drains pending work at startup, polls every
250 milliseconds, and processes batches of up to 1,000 until caught up.

`RebuildProjections` clears only derived tables, resets the checkpoint, and
replays the complete raw log in batches. A checkpoint version mismatch triggers
a rebuild before the HTTP server starts, and projection-backed reads also verify
the version. Old and new projection semantics are therefore not silently mixed.
Dashboard queries may continue to use raw events where an exact or
not-yet-projected answer is required.

Run a manual rebuild with `iris-server rebuild-projections`. `GET /api/status`
reports the last raw sequence, projection checkpoint, and current lag.

## SQLite runtime topology

Every connection enables foreign keys, WAL journal mode, a five-second busy
timeout, and `NORMAL` synchronous mode. The repository opens:

- one write connection (`MaxOpenConns(1)`) for ingestion, site mutations,
  projections, and maintenance;
- a separate read pool with up to four connections for dashboard queries.

Single-event writes and batch writes use the serialized writer. A batch is one
transaction, and conflicts on the event ID do nothing. The HTTP API returns
`202 Accepted` only after that raw transaction succeeds; there is no in-memory
queue whose loss can precede persistence.

WAL permits readers to continue while the writer commits, but SQLite remains a
single-writer database. Schema migrations, projection writes, retention, and
ingestion must stay serialized. Monitor WAL growth and reader duration, and back
up the database together with its WAL state using a SQLite-aware procedure.

## Versioned migrations

SQL migrations are embedded from `pkg/db/migrations` and recorded in
`schema_migrations`. They run transactionally when `NewSqliteDB` opens the
database. Migration 1 creates the v2 schema; migration 2 adds stable site-local
day storage and exact daily visitor/session projection sets.

For a pre-v2 database, startup renames the legacy `events` table, creates v2
tables, derives sites and domains from legacy rows, copies events with occurrence
and receive time set to the legacy timestamp, and sanitizes legacy URL/referrer
queries. Null or invalid legacy timestamps use migration time. The temporary
legacy table is then dropped in the same transaction. Back up the SQLite
database before deploying a migration despite the transactional boundary.

Future migrations must be append-only numbered files. Never edit an already
released migration: add a new version, migrate data explicitly, and cover both a
fresh database and an upgrade database in tests.

## Retention and maintenance

`retention_days` is enforced per site at server startup and every 24
hours. The cutoff is calculated from the current UTC instant. One serialized
transaction removes raw events whose occurrence time is before the cutoff,
sessions whose end time is before it, and daily projection rows before the
cutoff date. Sessions spanning the boundary are rebuilt from retained raw events
so expired visitor, referrer, entry-page, and count data cannot linger. The
operation is idempotent and logs the number of raw events removed.

Retention does not currently batch large deletions, checkpoint WAL, reclaim file
space, or expose progress beyond server logs. Those should be added before
retention workloads become large enough to hold the writer for a material
period. Privacy-driven deletion will also need a separate explicit workflow.
It can be invoked manually with `iris-server apply-retention`.

Backups, restore drills, projection lag, database size, WAL size, ingestion
latency, and busy/locked errors should be treated as production signals.

## When to introduce ClickHouse

SQLite is the default while Iris is a single-node, self-hosted service. Do not
add distributed infrastructure based only on table size. Consider a split where
SQLite or PostgreSQL keeps control-plane data and ClickHouse keeps event facts
and analytical projections when measurements show one or more of these:

- sustained ingestion no longer fits one serialized writer even after batching;
- multiple application replicas must ingest concurrently;
- arbitrary segmentation, funnels, or long-range raw scans miss latency goals
  after appropriate indexes and projections;
- event volume reaches hundreds of millions and retention, compression, or
  partition dropping becomes a dominant operational concern;
- projection lag cannot remain bounded during normal peak traffic.

The raw-fact-plus-rebuildable-projection boundary is intentionally independent
of SQLite. That lets a future ClickHouse projector consume the same logical
event stream without changing the browser contract or making ClickHouse a
requirement for small deployments.
