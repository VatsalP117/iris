# 👁 Iris Analytics

A dead-simple, privacy-friendly, self-hosted web analytics platform with a Next.js/Vite compatible npm client and a super-fast Go + SQLite backend.

## Project handbook

For the evidence-based architecture, API, data, security, operations, ADR, testing, and learning guide, start with [`docs/PROJECT_HANDBOOK.md`](docs/PROJECT_HANDBOOK.md).

For npm versioning and publishing, see [`docs/RELEASING.md`](docs/RELEASING.md).

For the current database model and scaling path, see
[`docs/09_V2_DATA_ARCHITECTURE.md`](docs/09_V2_DATA_ARCHITECTURE.md).

## Architecture

* **Frontend Dashboard:** React 18 + Vite (Tailwind/Plain CSS)
* **Backend API:** Go (`net/http`) + SQLite in WAL mode
* **Client SDK:** Typescript (`iris-analytics`)

SQLite uses a serialized writer and a separate read pool. Versioned migrations
run when the database opens. The append-only `events` table is the durable source
of truth; sessions and daily metrics are rebuildable projections.

---

## 1. Quickstart: Self-Hosting (Production)

The easiest way to run the Iris backend and dashboard is to use Docker.

1. Clone this repository (or copy the `docker-compose.yml`).
2. Run Docker Compose:

```bash
IRIS_ADMIN_TOKEN='replace-with-a-long-random-token' docker compose up -d
```

The server will automatically:
1. Spin up the backend API on `http://localhost:8081`.
2. Serve the built React Dashboard on the root URL `/`.
3. Create or migrate a persistent SQLite database in `./data/iris.db`.

**View Dashboard:** Open `http://localhost:8081/`

Before a browser can submit events, register its site ID and allowed hostname:

```bash
curl -X POST http://localhost:8081/api/sites \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer replace-with-a-long-random-token' \
  -d '{
    "site_id": "my-awesome-site",
    "name": "My Awesome Site",
    "timezone": "UTC",
    "retention_days": 365,
    "domains": ["www.example.com", "example.com"]
  }'
```

Use the same `site_id` when configuring the SDK. Event ingestion returns `404`
for an unknown site and `403` when an event URL's hostname is not registered for
that site. For local development, include the exact local hostname (usually
`localhost`) in `domains`; hostnames do not include a scheme or port.

### Dokploy Monorepo Split (Recommended)

To prevent marketing or SDK commits from rebuilding/redeploying the backend app, deploy this repo as two separate Dokploy apps:

1. **Backend + Dashboard app**
   - Build context: repository root
   - Dockerfile: `/Dockerfile`
   - Port: `8080`
   - Watch paths: `cmd/**,pkg/**,dashboard/**,go.mod,go.sum,Dockerfile,.dockerignore,pnpm-lock.yaml,package.json,docker/pnpm-workspace.dashboard.yaml`
2. **Marketing app**
   - Build context: `/marketing`
   - Dockerfile: `/marketing/Dockerfile` (or just `Dockerfile` when context is `/marketing`)
   - Port: `80`
   - Watch paths: `marketing/**`

The backend image is now isolated from monorepo-only changes via:
- root `.dockerignore` (excludes `marketing/` and `web/`)
- Dockerfile copy steps that include only backend/dashboard sources

---

## 2. Using the Client SDK (Your Website)

The client package is published from the `/web` workspace. Maintainers should
use the Changesets and GitHub Actions release flow documented in
[`docs/RELEASING.md`](docs/RELEASING.md); do not publish directly from a
developer machine.

### Installation

```bash
npm install iris-analytics
# or
yarn add iris-analytics
pnpm add iris-analytics
```

### Initialization (React / Next.js / Vue / Vanilla)

Initialize Iris **once** at the root of your application (e.g., `_app.tsx`, `layout.tsx`, or `main.ts`).

```typescript
import { Iris } from 'iris-analytics';

const analytics = new Iris({
  // Point this to your hosted Iris server URL
  host: "https://analytics.yourdomain.com", 
  
  // The unique identifier for this specific website/project
  siteId: "my-awesome-site",
  // Must match the timezone used when registering the site
  timezone: "UTC"
});

// Starts listening to route changes and automatically sends pageviews
analytics.start();
```

### Event Batching (Optional)

By default, each event fires an immediate HTTP request. You can enable **batching** to queue events and flush them in a single request — reducing network overhead for high-traffic pages.

```typescript
const analytics = new Iris({
  host: "https://analytics.yourdomain.com",
  siteId: "my-awesome-site",
  batching: {
    maxSize: 10,        // flush after 10 queued events (default: 10)
    flushInterval: 5000, // flush every 5 seconds (default: 5000ms)
    flushOnLeave: true,  // flush on tab switch / close (default: true)
  },
});

analytics.start();
```

### Manual Tracking (Custom Events)

You can track custom events manually anywhere in your app:

```typescript
analytics.track("User Signed Up", { plan: "Pro" });
analytics.track("Added to Cart", { itemId: 42, price: 99.99 });
```

---

## 3. Developing Locally

If you want to modify the dashboard or the Go backend, the workspace uses `pnpm` and `Taskfile`.

### Requirements
- Go 1.22+
- Node.js 20+ & pnpm
- Task (https://taskfile.dev)

### Start Development Server

```bash
# Terminal 1: Starts the Go Backend on :8080
task dev:backend

# Terminal 2: Starts the React Dashboard on :5173 (proxies /api to :8080)
task dev:dashboard
```

**Note:** The Go server will create an `iris.db` file in the `./data` directory relative to where you run it.

---

## 4. Environment Variables (Backend)

When running the Go backend, you can configure it using the following environment variables:

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | The port the HTTP server binds to. |
| `DB_PATH` | `./data/iris.db` | The path to the SQLite database file. |
| `DASHBOARD_DIR` | `./dashboard/dist` | Path to the directory containing the built frontend. |
| `IRIS_ADMIN_TOKEN` | unset | Bearer token required by `POST /api/sites`. Site mutation returns `503` while unset. |

`IRIS_LAB_PPROF` and `IRIS_LAB_DB_EXTRA_PAGES` are reliability-lab controls,
not production configuration. Site timezone and retention are configured through
`POST /api/sites`. The maintenance loop projects new events every 250 ms and
applies each site's retention policy at startup and every 24 hours.

Operational commands use the same `DB_PATH` configuration:

```bash
iris-server rebuild-projections
iris-server apply-retention
```

## 5. Dashboard Analytics APIs

Dashboard reporting uses `site_id`, `from`, and `to` query parameters:

| Endpoint | Purpose |
|---|---|
| `/api/site-trends` | Current and previous-period pageviews, visitors, sessions, and percentage changes |
| `/api/custom-events` | Custom-event totals, unique users, conversion rate, event rows, and trends |
| `/api/custom-events/timeseries` | Daily volume for a selected `event_name` |
| `/api/vitals/distribution` | Good, needs-improvement, and poor sample counts for LCP, INP, and CLS |
| `/api/vitals/pages` | Per-page P75 LCP, INP, CLS, and pageview traffic |
| `/api/vitals/score` | Overall 0–100 performance score and per-metric scores |
| `/api/status` | Database health, raw-event sequence, projection checkpoint, and projection lag |

The custom-event conversion rate is the percentage of pageview sessions that
recorded at least one custom event in the selected period. The performance score
maps each metric's P75 value onto a 0–100 scale: the Core Web Vitals "good"
threshold maps to 90, the "poor" threshold maps to 50, and values at twice the
poor threshold or worse map to 0. The overall score is the mean of the available
LCP, INP, and CLS scores.

## 6. Security & Privacy

* **No Cookies:** Anonymous visitor IDs rotate at midnight in the configured site timezone. Session IDs use `localStorage`, are isolated per site, shared across same-origin tabs, and roll after 30 minutes of inactivity. No third-party cookies are used.
* **URL minimization:** The backend accepts only absolute HTTP(S) URLs, strips query strings and fragments before storage, and verifies the resulting hostname against the site's domain allowlist.
* **Site administration:** `POST /api/sites` requires `Authorization: Bearer <IRIS_ADMIN_TOKEN>`. Use a long random value and keep it server-side. Analytics reads, site listing, and browser ingestion are currently unauthenticated.
* **CORS:** The backend allows cross-origin browser requests by default so the SDK and hosted dashboard can talk to the API without additional setup. The domain allowlist is an ingestion-integrity check, not authentication.
