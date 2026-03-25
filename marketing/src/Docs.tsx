import { useState } from 'react';
import { Menu, X, ChevronRight, ExternalLink } from 'lucide-react';

const sections = [
  { id: 'quick-start', label: 'Quick Start' },
  { id: 'installation', label: 'Installation' },
  { id: 'configuration', label: 'Configuration' },
  { id: 'autocapture', label: 'Auto-Capture' },
  { id: 'batching', label: 'Event Batching' },
  { id: 'sdk-methods', label: 'SDK Methods' },
  { id: 'self-hosting', label: 'Self-Hosting' },
  { id: 'environment', label: 'Environment Variables' },
  { id: 'api-reference', label: 'API Reference' },
  { id: 'privacy', label: 'Privacy Model' },
];

const CodeBlock = ({ children, title }: { children: string; title?: string }) => (
  <div className="rounded-xl overflow-hidden border border-white/10 my-6 shadow-lg">
    {title && (
      <div className="bg-white/5 px-4 py-2.5 border-b border-white/10 flex items-center gap-2">
        <div className="w-2.5 h-2.5 rounded-full bg-red-500/40" />
        <div className="w-2.5 h-2.5 rounded-full bg-yellow-500/40" />
        <div className="w-2.5 h-2.5 rounded-full bg-green-500/40" />
        <span className="text-[11px] text-muted-foreground font-mono ml-2">{title}</span>
      </div>
    )}
    <pre className="bg-black/40 p-5 overflow-x-auto text-sm leading-relaxed">
      <code className="text-white/90 font-mono">{children}</code>
    </pre>
  </div>
);

const PropTable = ({ rows }: { rows: [string, string, string, string][] }) => (
  <div className="overflow-x-auto my-6 rounded-xl border border-white/10">
    <table className="w-full text-sm">
      <thead>
        <tr className="bg-white/5 border-b border-white/10">
          <th className="text-left px-5 py-3 font-semibold text-foreground">Property</th>
          <th className="text-left px-5 py-3 font-semibold text-foreground">Type</th>
          <th className="text-left px-5 py-3 font-semibold text-foreground">Default</th>
          <th className="text-left px-5 py-3 font-semibold text-foreground">Description</th>
        </tr>
      </thead>
      <tbody>
        {rows.map(([prop, type, def, desc], i) => (
          <tr key={i} className="border-b border-white/5 last:border-0 hover:bg-white/[0.02] transition-colors">
            <td className="px-5 py-3 font-mono text-primary text-xs">{prop}</td>
            <td className="px-5 py-3 font-mono text-accent text-xs">{type}</td>
            <td className="px-5 py-3 font-mono text-muted-foreground text-xs">{def}</td>
            <td className="px-5 py-3 text-muted-foreground">{desc}</td>
          </tr>
        ))}
      </tbody>
    </table>
  </div>
);

const Callout = ({ type, children }: { type: 'tip' | 'warning' | 'info'; children: React.ReactNode }) => {
  const styles = {
    tip: 'border-green-500/30 bg-green-500/5 text-green-400',
    warning: 'border-yellow-500/30 bg-yellow-500/5 text-yellow-400',
    info: 'border-primary/30 bg-primary/5 text-primary',
  };
  const labels = { tip: 'Tip', warning: 'Warning', info: 'Info' };
  return (
    <div className={`border rounded-xl px-5 py-4 my-6 ${styles[type]}`}>
      <span className="font-bold text-xs uppercase tracking-wider">{labels[type]}</span>
      <div className="mt-1.5 text-muted-foreground text-sm leading-relaxed">{children}</div>
    </div>
  );
};

export default function Docs() {
  const [sidebarOpen, setSidebarOpen] = useState(false);

  return (
    <div className="min-h-screen flex pt-16">
      {/* Mobile sidebar toggle */}
      <button
        onClick={() => setSidebarOpen(!sidebarOpen)}
        className="lg:hidden fixed bottom-6 right-6 z-50 w-12 h-12 rounded-full bg-primary flex items-center justify-center shadow-lg shadow-primary/30 hover:scale-110 transition-transform"
      >
        {sidebarOpen ? <X className="w-5 h-5 text-white" /> : <Menu className="w-5 h-5 text-white" />}
      </button>

      {/* Sidebar overlay on mobile */}
      {sidebarOpen && (
        <div className="lg:hidden fixed inset-0 z-30 bg-black/60 backdrop-blur-sm" onClick={() => setSidebarOpen(false)} />
      )}

      {/* Sidebar */}
      <aside className={`
        fixed top-16 bottom-0 w-72 border-r border-white/5 bg-background/95 backdrop-blur-xl z-40 overflow-y-auto
        transition-transform duration-300 ease-in-out
        lg:sticky lg:top-16 lg:h-[calc(100vh-4rem)] lg:self-start lg:translate-x-0 lg:z-0
        ${sidebarOpen ? 'translate-x-0' : '-translate-x-full'}
      `}>
        <div className="p-6">
          <h3 className="text-xs font-bold uppercase tracking-widest text-muted-foreground mb-5">Documentation</h3>
          <nav className="space-y-1">
            {sections.map((s) => (
              <a
                key={s.id}
                href={`#${s.id}`}
                onClick={() => setSidebarOpen(false)}
                className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm text-muted-foreground hover:text-foreground hover:bg-white/5 transition-all group"
              >
                <ChevronRight className="w-3.5 h-3.5 opacity-40 group-hover:opacity-100 group-hover:text-primary transition-all" />
                {s.label}
              </a>
            ))}
          </nav>

          <div className="mt-10 p-4 rounded-xl bg-white/[0.03] border border-white/5">
            <p className="text-xs text-muted-foreground leading-relaxed">
              Package:
              <a href="https://www.npmjs.com/package/iris-analytics" target="_blank" rel="noopener noreferrer" className="text-primary ml-1 inline-flex items-center gap-1 hover:underline">
                iris-analytics <ExternalLink className="w-3 h-3" />
              </a>
            </p>
            <p className="text-xs text-muted-foreground leading-relaxed mt-1.5">
              Source:
              <a href="https://github.com/VatsalP117/iris" target="_blank" rel="noopener noreferrer" className="text-primary ml-1 inline-flex items-center gap-1 hover:underline">
                GitHub <ExternalLink className="w-3 h-3" />
              </a>
            </p>
          </div>
        </div>
      </aside>

      {/* Main content */}
      <main className="flex-1 max-w-4xl mx-auto px-6 sm:px-8 lg:px-12 py-12 lg:py-16">
        {/* Header */}
        <div className="mb-16">
          <span className="inline-flex items-center px-3 py-1 rounded-full text-xs font-medium bg-primary/10 text-primary border border-primary/20 mb-4">
            v0.2.2
          </span>
          <h1 className="text-4xl md:text-5xl font-extrabold tracking-tight mb-4">Documentation</h1>
          <p className="text-lg text-muted-foreground leading-relaxed max-w-2xl">
            Everything you need to integrate Iris Analytics into your app.
            Privacy-first, cookieless, and self-hosted.
          </p>
        </div>

        {/* Quick Start */}
        <section id="quick-start" className="docs-section">
          <h2>Quick Start</h2>
          <p>Get up and running with Iris in under a minute. Install the SDK, initialize it with your server host, and start tracking.</p>

          <CodeBlock title="terminal">{`npm install iris-analytics`}</CodeBlock>

          <CodeBlock title="app.ts">{`import { Iris } from 'iris-analytics';

const iris = new Iris({
  host: 'https://your-iris-server.com',
  siteId: 'my-site',
});

iris.start();`}</CodeBlock>

          <Callout type="tip">
            That's it. With the default config, Iris sends events via <code>navigator.sendBeacon</code> for zero performance impact on your site.
          </Callout>
        </section>

        {/* Installation */}
        <section id="installation" className="docs-section">
          <h2>Installation</h2>
          <p>Iris is available as a lightweight npm package. It supports ESM and CommonJS out of the box and ships with full TypeScript types.</p>

          <h3>Package Manager</h3>
          <CodeBlock title="terminal">{`# npm
npm install iris-analytics

# pnpm
pnpm add iris-analytics

# yarn
yarn add iris-analytics`}</CodeBlock>

          <h3>Script Tag</h3>
          <p>If you're not using a bundler, you can load Iris directly from a CDN:</p>

          <CodeBlock title="index.html">{`<script type="module">
  import { Iris } from 'https://esm.sh/iris-analytics';

  const iris = new Iris({
    host: 'https://your-iris-server.com',
    siteId: 'my-site',
  });

  iris.start();
</script>`}</CodeBlock>
        </section>

        {/* Configuration */}
        <section id="configuration" className="docs-section">
          <h2>Configuration</h2>
          <p>Pass a configuration object when initializing the <code>Iris</code> class. Only <code>host</code> and <code>siteId</code> are required.</p>

          <h3>IrisConfig</h3>
          <PropTable rows={[
            ['host', 'string', '—', 'URL of your Iris server (required)'],
            ['siteId', 'string', '—', 'Unique site identifier (required)'],
            ['autocapture', 'AutocaptureConfig | false', 'false', 'Enable auto-capture features'],
            ['batching', 'BatchConfig', '—', 'Configure event batching'],
            ['debug', 'boolean', 'false', 'Log events and state to console'],
          ]} />

          <CodeBlock title="Full config example">{`const iris = new Iris({
  host: 'https://analytics.example.com',
  siteId: 'marketing-site',
  autocapture: {
    pageviews: true,
    clicks: true,
    webvitals: true,
  },
  batching: {
    maxSize: 10,
    flushInterval: 5000,
    flushOnLeave: true,
  },
  debug: true,
});`}</CodeBlock>
        </section>

        {/* Auto-Capture */}
        <section id="autocapture" className="docs-section">
          <h2>Auto-Capture</h2>
          <p>Iris can automatically track common interactions without manual instrumentation. Enable features via the <code>autocapture</code> config.</p>

          <h3>AutocaptureConfig</h3>
          <PropTable rows={[
            ['pageviews', 'boolean', 'undefined', 'Track page views on start and on client-side navigation (pushState, popstate)'],
            ['clicks', 'boolean', 'undefined', 'Track clicks on buttons, links, submit inputs, and [role="button"] elements'],
            ['webvitals', 'boolean', 'undefined', 'Report Core Web Vitals (CLS, INP, LCP) via the web-vitals library'],
          ]} />

          <h3>Pageview Tracking</h3>
          <p>When <code>pageviews: true</code>, Iris fires a <code>$pageview</code> event on <code>start()</code> and patches <code>history.pushState</code> to track SPA navigation. It also listens for <code>popstate</code> events.</p>

          <h3>Click Tracking</h3>
          <p>When <code>clicks: true</code>, Iris listens globally for click events (using capture phase) and tracks clicks on interactive elements. For each click, a <code>$click</code> event is generated with properties:</p>
          <PropTable rows={[
            ['$tag', 'string', '—', 'The element tag name (button, a, etc.)'],
            ['$id', 'string', '—', 'The element\'s ID attribute'],
            ['$class', 'string', '—', 'The element\'s class list'],
            ['$text', 'string', '—', 'Inner text of the element (first 50 chars)'],
            ['$href', 'string', '—', 'Href attribute (links only)'],
          ]} />

          <Callout type="tip">
            Add the <code>iris-ignore</code> CSS class to any element to exclude it from click auto-capture.
          </Callout>

          <h3>Web Vitals</h3>
          <p>
            When <code>webvitals: true</code>, Iris uses the <code>web-vitals</code> library to report Core Web Vitals. Each metric generates a <code>$web_vital</code> event with <code>$name</code> (CLS, INP, or LCP), <code>$val</code>, and <code>$rating</code> (good, needs-improvement, poor).
          </p>
        </section>

        {/* Event Batching */}
        <section id="batching" className="docs-section">
          <h2>Event Batching</h2>
          <p>Instead of sending each event immediately, Iris can batch events together and flush them on an interval or when the queue fills up.</p>

          <h3>BatchConfig</h3>
          <PropTable rows={[
            ['maxSize', 'number', '10', 'Max events to queue before flushing'],
            ['flushInterval', 'number', '5000', 'Flush interval in milliseconds'],
            ['flushOnLeave', 'boolean', 'true', 'Flush remaining events when the user leaves the page (pagehide / visibilitychange)'],
          ]} />

          <CodeBlock title="Batching example">{`const iris = new Iris({
  host: 'https://analytics.example.com',
  siteId: 'my-site',
  batching: {
    maxSize: 20,
    flushInterval: 10000,
  },
});`}</CodeBlock>

          <Callout type="info">
            When batching is enabled, events are sent to <code>/api/events</code> (batch endpoint). Without batching, each event goes to <code>/api/event</code> individually.
          </Callout>
        </section>

        {/* SDK Methods */}
        <section id="sdk-methods" className="docs-section">
          <h2>SDK Methods</h2>

          <h3><code>iris.start()</code></h3>
          <p>Starts the tracker. Enables all configured auto-capture features. Safe to call only once — subsequent calls are no-ops.</p>

          <h3><code>iris.track(name, props?)</code></h3>
          <p>Manually track a custom event. The name should describe the action. Props is an optional object of key-value pairs.</p>
          <CodeBlock title="Custom events">{`// Track a signup
iris.track('signup', { plan: 'pro' });

// Track a feature usage
iris.track('export_csv', { rows: 1500 });

// Simple event with no props
iris.track('cta_clicked');`}</CodeBlock>

          <h3><code>iris.stop()</code></h3>
          <p>Stops the tracker, flushes any queued events, restores <code>history.pushState</code>, and removes all event listeners. Use this for cleanup in SPAs.</p>

          <CodeBlock title="React cleanup">{`useEffect(() => {
  iris.start();
  return () => iris.stop();
}, []);`}</CodeBlock>
        </section>

        {/* Self-Hosting */}
        <section id="self-hosting" className="docs-section">
          <h2>Self-Hosting</h2>
          <p>Iris is a single Docker image containing the Go backend, SQLite database, and the analytics dashboard. No external databases or services needed.</p>

          <h3>Docker Run</h3>
          <CodeBlock title="terminal">{`docker run -d \\
  -p 8080:8080 \\
  -v ./data:/app/data \\
  ghcr.io/vatsalp117/iris:latest`}</CodeBlock>

          <h3>Docker Compose</h3>
          <CodeBlock title="docker-compose.yml">{`version: '3.8'

services:
  iris:
    image: ghcr.io/vatsalp117/iris:latest
    restart: unless-stopped
    ports:
      - "8080:8080"
    volumes:
      - ./iris-data:/app/data
    environment:
      - PORT=8080
      - DB_PATH=/app/data/iris.db
      - DASHBOARD_DIR=/app/dashboard/dist`}</CodeBlock>

          <Callout type="warning">
            Mount a volume for <code>/app/data</code> to persist your SQLite database across container restarts and updates.
          </Callout>
        </section>

        {/* Environment Variables */}
        <section id="environment" className="docs-section">
          <h2>Environment Variables</h2>
          <p>Configure the Iris server at runtime using environment variables.</p>

          <PropTable rows={[
            ['PORT', 'number', '8080', 'Port the HTTP server listens on'],
            ['DB_PATH', 'string', '/app/data/iris.db', 'Path to the SQLite database file'],
            ['DASHBOARD_DIR', 'string', '/app/dashboard/dist', 'Path to the dashboard static files directory'],
          ]} />
        </section>

        {/* API Reference */}
        <section id="api-reference" className="docs-section">
          <h2>API Reference</h2>
          <p>All API endpoints accept and return JSON. Stats endpoints require a <code>site_id</code> query parameter and accept optional <code>from</code> and <code>to</code> date filters (YYYY-MM-DD format).</p>

          <h3>Event Ingestion</h3>
          <div className="space-y-4 my-6">
            <div className="p-4 rounded-xl bg-white/[0.03] border border-white/10">
              <div className="flex items-center gap-3 mb-2">
                <span className="px-2 py-0.5 rounded text-[10px] font-bold uppercase tracking-wider bg-green-500/15 text-green-400 border border-green-500/20">POST</span>
                <code className="text-sm font-mono text-foreground">/api/event</code>
              </div>
              <p className="text-sm text-muted-foreground">Track a single event. Body is a JSON event payload.</p>
            </div>
            <div className="p-4 rounded-xl bg-white/[0.03] border border-white/10">
              <div className="flex items-center gap-3 mb-2">
                <span className="px-2 py-0.5 rounded text-[10px] font-bold uppercase tracking-wider bg-green-500/15 text-green-400 border border-green-500/20">POST</span>
                <code className="text-sm font-mono text-foreground">/api/events</code>
              </div>
              <p className="text-sm text-muted-foreground">Track a batch of events. Body is a JSON array of event payloads. Max batch size: 50.</p>
            </div>
          </div>

          <h3>Event Payload</h3>
          <PropTable rows={[
            ['n', 'string', '—', 'Event name (e.g. $pageview, $click, signup)'],
            ['u', 'string', '—', 'Full page URL'],
            ['d', 'string', '—', 'Domain / hostname'],
            ['r', 'string | null', '—', 'Referrer URL'],
            ['w', 'number', '—', 'Viewport width in pixels'],
            ['s', 'string', '—', 'Site ID'],
            ['sid', 'string', '—', 'Session ID'],
            ['vid', 'string', '—', 'Visitor ID (anonymous, daily-rotating)'],
            ['p', 'object', '—', 'Custom properties (optional)'],
          ]} />

          <h3>Stats &amp; Analytics</h3>
          <p>All stats endpoints require <code>?site_id=your-site</code> (or <code>?domain=</code>). Optionally filter with <code>&from=YYYY-MM-DD&to=YYYY-MM-DD</code>.</p>

          <div className="space-y-3 my-6">
            {[
              { method: 'GET', path: '/api/stats', desc: 'Aggregate stats: pageviews, unique visitors, sessions' },
              { method: 'GET', path: '/api/pages', desc: 'Top 10 pages by pageview count' },
              { method: 'GET', path: '/api/referrers', desc: 'Top 10 referrers by visitor count' },
              { method: 'GET', path: '/api/vitals', desc: 'Aggregated Core Web Vitals (CLS, INP, LCP)' },
              { method: 'GET', path: '/api/devices', desc: 'Device breakdown by screen width category' },
              { method: 'GET', path: '/api/timeseries', desc: 'Daily pageview counts (for chart visualization)' },
              { method: 'GET', path: '/api/sites', desc: 'List all tracked sites and their domains' },
            ].map(({ method, path, desc }) => (
              <div key={path} className="p-4 rounded-xl bg-white/[0.03] border border-white/10 flex flex-col sm:flex-row sm:items-center gap-2 sm:gap-4">
                <div className="flex items-center gap-3 shrink-0">
                  <span className="px-2 py-0.5 rounded text-[10px] font-bold uppercase tracking-wider bg-primary/15 text-primary border border-primary/20">{method}</span>
                  <code className="text-sm font-mono text-foreground">{path}</code>
                </div>
                <p className="text-sm text-muted-foreground">{desc}</p>
              </div>
            ))}
          </div>

          <h3>Example Request</h3>
          <CodeBlock title="terminal">{`curl "https://your-iris.com/api/stats?site_id=my-site&from=2026-03-01&to=2026-03-25"`}</CodeBlock>

          <CodeBlock title="Response">{`{
  "pageviews": 12847,
  "unique_visitors": 3291,
  "sessions": 5072
}`}</CodeBlock>
        </section>

        {/* Privacy Model */}
        <section id="privacy" className="docs-section">
          <h2>Privacy Model</h2>
          <p>Iris is designed from the ground up to be privacy-respecting. Here's how it works:</p>

          <div className="space-y-6 mt-8">
            <div className="p-6 rounded-2xl bg-white/[0.03] border border-white/10">
              <h4 className="font-bold mb-2">🍪 Zero Cookies</h4>
              <p className="text-sm text-muted-foreground leading-relaxed">Iris never sets any cookies. Visitor and session IDs are stored in <code>localStorage</code> and <code>sessionStorage</code>, which are first-party and inaccessible to third-party scripts.</p>
            </div>
            <div className="p-6 rounded-2xl bg-white/[0.03] border border-white/10">
              <h4 className="font-bold mb-2">🔄 Daily-Rotating Visitor IDs</h4>
              <p className="text-sm text-muted-foreground leading-relaxed">Visitor IDs are regenerated every UTC day. This prevents long-term tracking of individual users while still providing meaningful daily analytics.</p>
            </div>
            <div className="p-6 rounded-2xl bg-white/[0.03] border border-white/10">
              <h4 className="font-bold mb-2">📍 No PII Collection</h4>
              <p className="text-sm text-muted-foreground leading-relaxed">Iris doesn't collect IP addresses, user agents, or any personally identifiable information. All tracking data is anonymous by default.</p>
            </div>
            <div className="p-6 rounded-2xl bg-white/[0.03] border border-white/10">
              <h4 className="font-bold mb-2">🏠 Self-Hosted Only</h4>
              <p className="text-sm text-muted-foreground leading-relaxed">Your data lives on your server, in your SQLite database. Nothing is ever sent to third parties. No analytics SaaS, no tracking networks, no data brokers.</p>
            </div>
          </div>

          <Callout type="info">
            Because Iris doesn't use cookies or track personal data, most privacy regulations (GDPR, CCPA) do not require a consent banner.
          </Callout>
        </section>

        {/* Bottom spacer */}
        <div className="h-24" />
      </main>
    </div>
  );
}
