import { useMemo, useState } from "react";

import type { DateWindow } from "../App";
import type { SiteStat } from "../api";
import { Icon } from "./Icon";

interface Props {
    site: SiteStat;
    dateWindow: DateWindow;
}

export function EventsPage({ site }: Props) {
    const [query, setQuery] = useState("");
    const installCommand = useMemo(() => `npm install @iris-analytics/sdk`, []);
    const snippet = `import Iris from "@iris-analytics/sdk";

const iris = new Iris({
    siteId: "${site.site_id}",
    endpoint: "/api/events"
});

iris.track("user_signup_completed", {
    plan: "enterprise",
    referrer: "google_search"
});`;

    return (
        <div className="page-stack">
            <section className="events-toolbar">
                <div>
                    <span className="eyebrow">Behavior</span>
                    <h2>Event intelligence</h2>
                    <p>Understand product actions without collecting personal data.</p>
                </div>
                <label className="search-field">
                    <Icon name="search" size={17} />
                    <span className="sr-only">Search events</span>
                    <input value={query} onChange={(event) => setQuery(event.target.value)} placeholder="Quick search..." />
                </label>
            </section>

            <section className="event-trends-card card">
                <div className="event-trends-heading">
                    <div>
                        <h3>Event Trends</h3>
                        <p>{query || "Select an event below"} · selected date range</p>
                    </div>
                    <div>
                        <span>Total events</span>
                        <strong>—</strong>
                    </div>
                </div>
                <div className="events-empty-chart">
                    <div className="chart-grid-lines" />
                    <Icon name="activity" size={28} />
                    <strong>No custom event activity yet</strong>
                    <p>Tracked events will appear here as soon as the SDK sends an event name that does not start with <code>$</code>.</p>
                </div>
            </section>

            <section className="events-lower-grid">
                <article className="card tracked-events-card">
                    <div className="card-header">
                        <div>
                            <h3>Tracked Events</h3>
                            <p>Custom product actions recorded for this site</p>
                        </div>
                    </div>
                    <div className="event-table-head">
                        <span>Event name</span>
                        <span>Total count</span>
                        <span>Unique users</span>
                        <span>Trend</span>
                    </div>
                    <div className="table-empty">
                        <Icon name="calendar" size={24} />
                        <strong>Waiting for your first event</strong>
                        <p>Use the implementation snippet to start tracking conversions and product behavior.</p>
                    </div>
                </article>

                <article className="sdk-card">
                    <div className="sdk-card-header">
                        <div><span /><span /><span /></div>
                        <strong><Icon name="code" size={15} /> SDK Implementation</strong>
                    </div>
                    <pre><code><span className="code-comment"># Install the SDK</span>{"\n"}<span className="code-command">{installCommand}</span>{"\n\n"}{snippet}</code></pre>
                    <button onClick={() => navigator.clipboard?.writeText(snippet)}>
                        Copy implementation
                    </button>
                </article>
            </section>
        </div>
    );
}
