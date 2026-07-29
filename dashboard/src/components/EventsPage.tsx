import { useEffect, useMemo, useState } from "react";
import { format, parseISO } from "date-fns";
import {
    Area,
    AreaChart,
    CartesianGrid,
    ResponsiveContainer,
    Tooltip,
    XAxis,
    YAxis,
} from "recharts";

import type { DateWindow } from "../App";
import {
    api,
    CustomEventsResult,
    CustomEventTimeSeriesBucket,
    SiteStat,
} from "../api";
import { Icon } from "./Icon";

interface Props {
    site: SiteStat;
    dateWindow: DateWindow;
}

function formatChange(value: number): string {
    return `${value > 0 ? "+" : ""}${value.toFixed(1)}%`;
}

export function EventsPage({ site, dateWindow }: Props) {
    const [query, setQuery] = useState("");
    const [result, setResult] = useState<CustomEventsResult | null>(null);
    const [selectedEvent, setSelectedEvent] = useState("");
    const [series, setSeries] = useState<CustomEventTimeSeriesBucket[]>([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const controller = new AbortController();
        setLoading(true);
        api.customEvents(site.site_id, dateWindow.queryFrom, dateWindow.queryTo, controller.signal)
            .then((data) => {
                setResult(data);
                setSelectedEvent((current) => {
                    if (data.events.some((event) => event.event_name === current)) return current;
                    return data.events[0]?.event_name ?? "";
                });
            })
            .catch((error) => {
                if (error instanceof DOMException && error.name === "AbortError") return;
                console.error("Iris: failed to fetch custom events", error);
            })
            .finally(() => {
                if (!controller.signal.aborted) setLoading(false);
            });
        return () => controller.abort();
    }, [dateWindow.queryFrom, dateWindow.queryTo, site.site_id]);

    useEffect(() => {
        if (!selectedEvent) {
            setSeries([]);
            return;
        }
        const controller = new AbortController();
        api.customEventTimeseries(
            site.site_id,
            selectedEvent,
            dateWindow.queryFrom,
            dateWindow.queryTo,
            controller.signal,
        )
            .then(setSeries)
            .catch((error) => {
                if (error instanceof DOMException && error.name === "AbortError") return;
                console.error("Iris: failed to fetch custom event time series", error);
            });
        return () => controller.abort();
    }, [dateWindow.queryFrom, dateWindow.queryTo, selectedEvent, site.site_id]);

    const filteredEvents = useMemo(() => {
        const normalized = query.trim().toLowerCase();
        if (!normalized) return result?.events ?? [];
        return (result?.events ?? []).filter((event) => event.event_name.toLowerCase().includes(normalized));
    }, [query, result]);

    const installCommand = "npm install @iris-analytics/sdk";
    const snippet = `import Iris from "@iris-analytics/sdk";

const iris = new Iris({
    siteId: "${site.site_id}",
    endpoint: "/api/events"
});

iris.track("user_signup_completed", {
    plan: "enterprise",
    referrer: "google_search"
});`;

    const summary = result?.summary;
    const selectedStat = result?.events.find((event) => event.event_name === selectedEvent);
    const hasEvents = (result?.events.length ?? 0) > 0;

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

            <section className="event-summary-grid" aria-label="Custom event summary">
                <article className="metric-card">
                    <div className="metric-card-heading"><span>Total Events</span><Icon name="activity" size={16} /></div>
                    <strong>{loading ? "—" : (summary?.total_events ?? 0).toLocaleString()}</strong>
                    <small className={(summary?.change_percent ?? 0) >= 0 ? "positive" : "negative"}>
                        {summary ? `${formatChange(summary.change_percent)} vs previous period` : "No comparison available"}
                    </small>
                </article>
                <article className="metric-card">
                    <div className="metric-card-heading"><span>Unique Users</span><Icon name="users" size={16} /></div>
                    <strong>{loading ? "—" : (summary?.unique_users ?? 0).toLocaleString()}</strong>
                    <small>Privacy-safe visitor identifiers</small>
                </article>
                <article className="metric-card">
                    <div className="metric-card-heading"><span>Conversion Rate</span><Icon name="trend" size={16} /></div>
                    <strong>{loading ? "—" : `${(summary?.conversion_rate ?? 0).toFixed(1)}%`}</strong>
                    <small>Sessions with at least one custom event</small>
                </article>
            </section>

            <section className="event-trends-card card">
                <div className="event-trends-heading">
                    <div>
                        <h3>Event Trends</h3>
                        <p>{selectedEvent || "Select an event below"} · selected date range</p>
                    </div>
                    <div>
                        <span>Total events</span>
                        <strong>{selectedStat?.total_count.toLocaleString() ?? "—"}</strong>
                    </div>
                </div>
                {hasEvents ? (
                    <div className="event-chart">
                        <ResponsiveContainer width="100%" height="100%">
                            <AreaChart data={series} margin={{ top: 12, right: 12, left: -24, bottom: 0 }}>
                                <defs>
                                    <linearGradient id="eventArea" x1="0" y1="0" x2="0" y2="1">
                                        <stop offset="0%" stopColor="#4f46e5" stopOpacity={0.3} />
                                        <stop offset="100%" stopColor="#4f46e5" stopOpacity={0} />
                                    </linearGradient>
                                </defs>
                                <CartesianGrid vertical={false} stroke="var(--outline)" strokeDasharray="3 3" />
                                <XAxis
                                    axisLine={false}
                                    dataKey="date"
                                    tick={{ fill: "var(--text-subtle)", fontFamily: "var(--font-mono)", fontSize: 10 }}
                                    tickFormatter={(value) => format(parseISO(value), "MMM d")}
                                    tickLine={false}
                                />
                                <YAxis
                                    allowDecimals={false}
                                    axisLine={false}
                                    tick={{ fill: "var(--text-subtle)", fontFamily: "var(--font-mono)", fontSize: 10 }}
                                    tickLine={false}
                                />
                                <Tooltip
                                    labelFormatter={(value) => format(parseISO(String(value)), "MMM d, yyyy")}
                                    formatter={(value) => [Number(value).toLocaleString(), "Events"]}
                                />
                                <Area
                                    activeDot={{ fill: "#3525cd", r: 4, strokeWidth: 0 }}
                                    dataKey="count"
                                    dot={false}
                                    fill="url(#eventArea)"
                                    stroke="#4f46e5"
                                    strokeWidth={2}
                                    type="monotone"
                                />
                            </AreaChart>
                        </ResponsiveContainer>
                    </div>
                ) : (
                    <div className="events-empty-chart">
                        <div className="chart-grid-lines" />
                        <Icon name="activity" size={28} />
                        <strong>No custom event activity yet</strong>
                        <p>Tracked events will appear here as soon as the SDK sends an event name that does not start with <code>$</code>.</p>
                    </div>
                )}
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
                    {filteredEvents.length > 0 ? (
                        <div className="event-table-body">
                            {filteredEvents.map((event) => (
                                <button
                                    className={event.event_name === selectedEvent ? "active" : ""}
                                    key={event.event_name}
                                    onClick={() => setSelectedEvent(event.event_name)}
                                >
                                    <span>{event.event_name}</span>
                                    <span>{event.total_count.toLocaleString()}</span>
                                    <span>{event.unique_users.toLocaleString()}</span>
                                    <span className={event.change_percent >= 0 ? "up" : "down"}>
                                        {formatChange(event.change_percent)}
                                    </span>
                                </button>
                            ))}
                        </div>
                    ) : (
                        <div className="table-empty">
                            <Icon name="calendar" size={24} />
                            <strong>{query ? "No matching events" : "Waiting for your first event"}</strong>
                            <p>{query ? "Try another event name." : "Use the implementation snippet to start tracking conversions and product behavior."}</p>
                        </div>
                    )}
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
