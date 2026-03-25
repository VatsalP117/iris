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

```ts
import { Iris } from "iris-analytics";

const analytics = new Iris({
  host: "https://analytics.yourdomain.com",
  siteId: "my-site",
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
- Visitor IDs are anonymous and rotate daily in UTC.
- Session IDs are tab-scoped using `sessionStorage`.

## Backend

This SDK sends events to the Iris backend (`/api/event`, `/api/events`).
Run the full stack from the main repo:

- Repository: https://github.com/VatsalP117/iris
