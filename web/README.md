# iris-analytics

Privacy-first, self-hosted web analytics SDK for browser apps.

## Install

```bash
npm install iris-analytics
# or
yarn add iris-analytics
# or
pnpm add iris-analytics
```

## Quick Start

Register `my-site` and its hostname with the backend before starting the SDK.
Site mutation requires the server's `IRIS_ADMIN_TOKEN`; see the root repository
README for the authenticated `POST /api/sites` example.

```ts
import { Iris } from "iris-analytics";

const analytics = new Iris({
  host: "https://analytics.yourdomain.com",
  siteId: "my-site",
  timezone: "UTC", // must match the registered site
  autocapture: {
    pageviews: true,
    clicks: true,
    webvitals: true,
  },
});

analytics.start();
```

## Where To Mount It

Initialize Iris once at the app root on the client.

### Next.js (App Router)

Create a client component and mount it in `app/layout.tsx`.

```tsx
// app/Analytics.tsx
"use client";

import { useEffect } from "react";
import { Iris } from "iris-analytics";

export function Analytics() {
  useEffect(() => {
    const iris = new Iris({
      host: "https://analytics.yourdomain.com",
      siteId: "my-site",
      autocapture: { pageviews: true, clicks: true, webvitals: true },
    });
    iris.start();
    return () => iris.stop();
  }, []);

  return null;
}
```

```tsx
// app/layout.tsx
import { Analytics } from "./Analytics";

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>
        <Analytics />
        {children}
      </body>
    </html>
  );
}
```

### Next.js (Pages Router)

Mount in `pages/_app.tsx`.

```tsx
import { useEffect } from "react";
import type { AppProps } from "next/app";
import { Iris } from "iris-analytics";

export default function App({ Component, pageProps }: AppProps) {
  useEffect(() => {
    const iris = new Iris({
      host: "https://analytics.yourdomain.com",
      siteId: "my-site",
      autocapture: { pageviews: true, clicks: true, webvitals: true },
    });
    iris.start();
    return () => iris.stop();
  }, []);

  return <Component {...pageProps} />;
}
```

### React (Vite / CRA)

Mount once in your root app component.

```tsx
import { useEffect } from "react";
import { Iris } from "iris-analytics";

export default function App() {
  useEffect(() => {
    const iris = new Iris({
      host: "https://analytics.yourdomain.com",
      siteId: "my-site",
      autocapture: { pageviews: true, clicks: true, webvitals: true },
    });
    iris.start();
    return () => iris.stop();
  }, []);

  return <main>Your app</main>;
}
```

### Nuxt 3

Create a client-only plugin at `plugins/iris.client.ts`.

```ts
import { Iris } from "iris-analytics";

export default defineNuxtPlugin(() => {
  const iris = new Iris({
    host: "https://analytics.yourdomain.com",
    siteId: "my-site",
    autocapture: { pageviews: true, clicks: true, webvitals: true },
  });
  iris.start();
});
```

## Manual Events

```ts
analytics.track("User Signed Up", { plan: "Pro" });
```

## Batching (Optional)

```ts
const analytics = new Iris({
  host: "https://analytics.yourdomain.com",
  siteId: "my-site",
  batching: {
    maxSize: 10,
    flushInterval: 5000,
    flushOnLeave: true,
  },
});
```

## Privacy Notes

- No third-party cookies.
- Visitor IDs are anonymous, isolated per site, and rotate at midnight in the
  configured site timezone.
- Session IDs use `localStorage`, are isolated per site, shared across
  same-origin tabs, and roll after 30 minutes without tracked activity.
- The backend removes URL/referrer query strings and fragments before storage
  and requires the event hostname to be registered for the configured site.

## Backend

This SDK sends versioned, client-identified events to the Iris backend
(`/api/event`, `/api/events`). Event IDs make replay idempotent; the transport
queue itself remains memory-only and best-effort.
Run the full stack from the main repo:

- Repository: https://github.com/VatsalP117/iris
