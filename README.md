# 👁 Iris Analytics

A dead-simple, privacy-friendly, self-hosted web analytics platform with a Next.js/Vite compatible npm client and a super-fast Go + SQLite backend.

## Architecture

* **Frontend Dashboard:** React 18 + Vite (Tailwind/Plain CSS)
* **Backend API:** Go (`net/http`) + SQLite
* **Client SDK:** Typescript (`@bigchill101/iris`)

---

## 1. Quickstart: Self-Hosting (Production)

The easiest way to run the Iris backend and dashboard is to use Docker.

1. Clone this repository (or copy the `docker-compose.yml`).
2. Run Docker Compose:

```bash
docker compose up -d
```

The server will automatically:
1. Spin up the backend API on `http://localhost:8080`.
2. Serve the built React Dashboard on the root URL `/`.
3. Create a SQLite database in `./data/iris.db` (persistent).

**View Dashboard:** Open `http://localhost:8080/`

---

## 2. Using the Client SDK (Your Website)

First, add the Iris client package to your project. Currently, this package sits inside this monorepo in the `/web` folder. 

Publish it to npm:
```bash
cd web
npm run build
npm publish --access public
```

### Installation

```bash
npm install @bigchill101/iris
# or
yarn add @bigchill101/iris
pnpm add @bigchill101/iris
```

### Initialization (React / Next.js / Vue / Vanilla)

Initialize Iris **once** at the root of your application (e.g., `_app.tsx`, `layout.tsx`, or `main.ts`).

```typescript
import { Iris } from '@bigchill101/iris';

const analytics = new Iris({
  // Point this to your hosted Iris server URL
  host: "https://analytics.yourdomain.com", 
  
  // The unique identifier for this specific website/project
  siteId: "my-awesome-site" 
});

// Starts listening to route changes and automatically sends pageviews
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

## 5. Security & Privacy

* **No Cookies:** User identity is tracked anonymously using standard `localStorage` (Visitor ID) and `sessionStorage` (Session ID). No third-party cookies are used.
* **CORS:** Currently configured broadly for development. Before moving your backend to a public production URL, ensure you lock down the CORS origins within `handler.go`.
