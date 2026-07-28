# Iris Reliability Lab

The Reliability Lab sends deterministic synthetic analytics traffic to an
isolated Iris server and reconciles the accepted event sequences against the
server's SQLite database.

This first implementation measures:

- planned, attempted, accepted, rejected, and failed events;
- individual and batched request throughput;
- HTTP status distribution;
- average, p50, p90, p95, p99, and maximum request latency;
- maximum load-generator scheduling lag;
- stored, missing, duplicate, unexpected, and mismatched event rows;
- database file size; and
- headline stats, page, and device API aggregates.

It writes both `report.md` and `summary.json`.

## Safety

Use a dedicated database and isolated Iris process. The CLI refuses non-loopback
targets unless `--allow-nonlocal` is explicitly supplied, but a loopback server
can still point at valuable data.

The lab never deletes the database. Every run uses a unique site and test run ID
so its rows can be identified.

Generated reports under `artifacts/reliability/` are ignored by Git.

## Run locally

Start Iris with a dedicated database:

```bash
mkdir -p /tmp/iris-reliability
DB_PATH=/tmp/iris-reliability/iris.db PORT=8080 go run ./cmd/server
```

In another terminal, run a short smoke profile:

```bash
go run ./cmd/iris-lab run \
    --target http://127.0.0.1:8080 \
    --db /tmp/iris-reliability/iris.db \
    --rate 10 \
    --duration 30s \
    --batch-size 1
```

The equivalent Task command is:

```bash
IRIS_LAB_DB=/tmp/iris-reliability/iris.db task lab:run
```

An exact event count can replace `rate × duration`, which is useful for quick
correctness checks:

```bash
go run ./cmd/iris-lab run \
    --db /tmp/iris-reliability/iris.db \
    --rate 100 \
    --events 1000 \
    --batch-size 10
```

## Initial profiles

Run these only after the smoke profile passes:

```bash
# 100 events/s for two minutes
go run ./cmd/iris-lab run \
    --db /tmp/iris-reliability/iris.db \
    --rate 100 \
    --duration 2m

# 500 events/s for five minutes
go run ./cmd/iris-lab run \
    --db /tmp/iris-reliability/iris.db \
    --rate 500 \
    --duration 5m

# 1,000 events/s for five minutes, in batches of 10
go run ./cmd/iris-lab run \
    --db /tmp/iris-reliability/iris.db \
    --rate 1000 \
    --duration 5m \
    --batch-size 10
```

Events per second and HTTP requests per second are intentionally reported
separately. At 1,000 events/s with batches of 10, Iris receives approximately
100 HTTP requests/s.

## Interpreting failures

The command exits unsuccessfully when:

- not every planned event was attempted and accepted;
- an accepted event sequence is absent from SQLite;
- a sequence is stored more than once;
- a row exists for a sequence that was not accepted;
- stored core fields differ from the generated event; or
- stats, page, or device aggregates differ from the accepted manifest.

The Markdown report includes up to 50 example sequences for each mismatch type.
The complete machine-readable summary is available in `summary.json`.

## Current boundary

This is the deterministic correctness and HTTP-load core. Planned additions
include:

- k6 fixed-arrival, ramp, spike, and soak profiles;
- concurrent dashboard reads;
- browser SDK lifecycle and navigation flows;
- CPU, memory, disk, WAL, and Go profile capture;
- server restart and database fault injection; and
- comparison reports across git revisions.

See [the product roadmap](../../docs/ROADMAP.md) for the complete plan.
