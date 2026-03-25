import { useState, useEffect, useCallback } from "react";
import { format, subDays, subHours } from "date-fns";
import { api, StatsResult, PageStat, ReferrerStat, VitalStat, DeviceStat, SiteStat } from "./api";
import { StatsCards } from "./components/StatsCards";
import { TopPages } from "./components/TopPages";
import { TopReferrers } from "./components/TopReferrers";
import { WebVitals } from "./components/WebVitals";
import { DeviceBreakdown } from "./components/DeviceBreakdown";
import { PageviewsChart, buildEmptyBuckets, DayBucket } from "./components/PageviewsChart";

type Tab = "overview" | "pages" | "referrers" | "vitals" | "devices";
type PresetKey = "24h" | "7d" | "30d" | "90d";

type DatePreset = {
    key: PresetKey;
    label: string;
    unit: "hours" | "days";
    amount: number;
};

type DateWindow = {
    from: Date;
    to: Date;
    queryFrom: string;
    queryTo: string;
};

const DATE_PRESETS: DatePreset[] = [
    { key: "24h", label: "24h", unit: "hours", amount: 24 },
    { key: "7d", label: "7d", unit: "days", amount: 7 },
    { key: "30d", label: "30d", unit: "days", amount: 30 },
    { key: "90d", label: "90d", unit: "days", amount: 90 },
];

function fmtDate(d: Date): string {
    return format(d, "yyyy-MM-dd");
}

function fmtDateTime(d: Date): string {
    return format(d, "yyyy-MM-dd'T'HH:mm:ssxxx");
}

function buildWindow(preset: PresetKey): DateWindow {
    const cfg = DATE_PRESETS.find((p) => p.key === preset) ?? DATE_PRESETS[2];
    const to = new Date();

    if (cfg.unit === "hours") {
        const from = subHours(to, cfg.amount);
        return {
            from,
            to,
            queryFrom: fmtDateTime(from),
            queryTo: fmtDateTime(to),
        };
    }

    const from = subDays(to, cfg.amount);
    return {
        from,
        to,
        queryFrom: fmtDate(from),
        queryTo: fmtDate(to),
    };
}

function formatWindow(window: DateWindow, preset: PresetKey): string {
    if (preset === "24h") {
        return `${format(window.from, "MMM d, p")} - ${format(window.to, "MMM d, p")}`;
    }
    return `${format(window.from, "MMM d")} - ${format(window.to, "MMM d, yyyy")}`;
}

function getRangeLabel(preset: PresetKey): string {
    if (preset === "24h") return "Last 24 hours";
    return `Last ${preset.replace("d", " days")}`;
}

export default function App() {
    const [theme, setTheme] = useState<"dark" | "light">(() => {
        const saved = localStorage.getItem("iris-theme");
        return (saved as "dark" | "light") || "dark";
    });

    useEffect(() => {
        document.documentElement.setAttribute("data-theme", theme);
        localStorage.setItem("iris-theme", theme);
    }, [theme]);

    const toggleTheme = () => setTheme(t => t === "dark" ? "light" : "dark");

    const SunIcon = () => (
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <circle cx="12" cy="12" r="5"/>
            <line x1="12" y1="1" x2="12" y2="3"/>
            <line x1="12" y1="21" x2="12" y2="23"/>
            <line x1="4.22" y1="4.22" x2="5.64" y2="5.64"/>
            <line x1="18.36" y1="18.36" x2="19.78" y2="19.78"/>
            <line x1="1" y1="12" x2="3" y2="12"/>
            <line x1="21" y1="12" x2="23" y2="12"/>
            <line x1="4.22" y1="19.78" x2="5.64" y2="18.36"/>
            <line x1="18.36" y1="5.64" x2="19.78" y2="4.22"/>
        </svg>
    );

    const MoonIcon = () => (
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/>
        </svg>
    );
    const [sites, setSites] = useState<SiteStat[]>([]);
    const [sitesLoading, setSitesLoading] = useState(true);
    const [selectedSite, setSelectedSite] = useState<SiteStat | null>(null);

    const [activeTab, setActiveTab] = useState<Tab>("overview");
    const [preset, setPreset] = useState<PresetKey>("30d");
    const [windowRange, setWindowRange] = useState<DateWindow>(() => buildWindow("30d"));

    const [stats, setStats] = useState<StatsResult | null>(null);
    const [pages, setPages] = useState<PageStat[]>([]);
    const [referrers, setReferrers] = useState<ReferrerStat[]>([]);
    const [vitals, setVitals] = useState<VitalStat[]>([]);
    const [devices, setDevices] = useState<DeviceStat[]>([]);
    const [chartData, setChartData] = useState<DayBucket[]>([]);
    const [visitorsData, setVisitorsData] = useState<DayBucket[]>([]);
    const [loading, setLoading] = useState(false);

    useEffect(() => {
        setSitesLoading(true);
        api.sites()
            .then((list) => {
                setSites(list ?? []);
                if (list && list.length > 0) setSelectedSite(list[0]);
            })
            .catch((err) => console.error("Iris: failed to fetch sites", err))
            .finally(() => setSitesLoading(false));
    }, []);

    const fetchAll = useCallback(async (siteId: string, range: DateWindow = windowRange) => {
        if (!siteId) return;
        setLoading(true);
        try {
            const [s, p, r, v, dev, ts, uv] = await Promise.all([
                api.stats(siteId, range.queryFrom, range.queryTo),
                api.pages(siteId, range.queryFrom, range.queryTo),
                api.referrers(siteId, range.queryFrom, range.queryTo),
                api.vitals(siteId, range.queryFrom, range.queryTo),
                api.devices(siteId, range.queryFrom, range.queryTo),
                api.timeseries(siteId, range.queryFrom, range.queryTo),
                api.uniqueVisitorsTimeseries(siteId, range.queryFrom, range.queryTo),
            ]);
            setStats(s);
            setPages(p ?? []);
            setReferrers(r ?? []);
            setVitals(v ?? []);
            setDevices(dev ?? []);
            setChartData(ts ?? []);
            setVisitorsData(uv ?? []);
        } catch (err) {
            console.error("Iris: fetch error", err);
        } finally {
            setLoading(false);
        }
    }, [windowRange]);

    useEffect(() => {
        if (selectedSite) fetchAll(selectedSite.site_id);
    }, [selectedSite, fetchAll]);

    function handlePreset(nextPreset: PresetKey) {
        setPreset(nextPreset);
        setWindowRange(buildWindow(nextPreset));
    }

    function handleSiteChange(e: React.ChangeEvent<HTMLSelectElement>) {
        const site = sites.find((s) => s.site_id === e.target.value) ?? null;
        setSelectedSite(site);

        setStats(null);
        setPages([]);
        setReferrers([]);
        setVitals([]);
        setDevices([]);
        setChartData([]);
    }

    function handleRefresh() {
        if (!selectedSite) return;
        setWindowRange(buildWindow(preset));
    }

    const emptyBuckets = buildEmptyBuckets(windowRange.from, windowRange.to);
    const siteId = selectedSite?.site_id ?? "";

    return (
        <div className="app-layout">
            <aside className="sidebar">
                <div className="sidebar-logo">
                    <div className="sidebar-logo-icon">IR</div>
                    <span className="sidebar-logo-text">Iris</span>
                </div>

                <span className="sidebar-section-label">Menu</span>

                <nav className="sidebar-nav" aria-label="Dashboard tabs">
                    <button className={`sidebar-item ${activeTab === "overview" ? "active" : ""}`} onClick={() => setActiveTab("overview")}>
                        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><rect x="3" y="3" width="7" height="7"/><rect x="14" y="3" width="7" height="7"/><rect x="14" y="14" width="7" height="7"/><rect x="3" y="14" width="7" height="7"/></svg>
                        <span className="sidebar-item-label">Overview</span>
                    </button>
                    <button className={`sidebar-item ${activeTab === "pages" ? "active" : ""}`} onClick={() => setActiveTab("pages")}>
                        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>
                        <span className="sidebar-item-label">Pages</span>
                    </button>
                    <button className={`sidebar-item ${activeTab === "referrers" ? "active" : ""}`} onClick={() => setActiveTab("referrers")}>
                        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/><path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/></svg>
                        <span className="sidebar-item-label">Referrers</span>
                    </button>
                    <button className={`sidebar-item ${activeTab === "vitals" ? "active" : ""}`} onClick={() => setActiveTab("vitals")}>
                        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M22 12h-4l-3 9L9 3l-3 9H2"/></svg>
                        <span className="sidebar-item-label">Web Vitals</span>
                    </button>
                    <button className={`sidebar-item ${activeTab === "devices" ? "active" : ""}`} onClick={() => setActiveTab("devices")}>
                        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><rect x="2" y="3" width="20" height="14" rx="2" ry="2"/><line x1="8" y1="21" x2="16" y2="21"/><line x1="12" y1="17" x2="12" y2="21"/></svg>
                        <span className="sidebar-item-label">Devices</span>
                    </button>
                </nav>

                <div className="sidebar-footnote">
                    <span className="sidebar-footnote-label">Version</span>
                    <span className="sidebar-footnote-value">v0.1.0</span>
                </div>
            </aside>

            <div className="main-area">
                <header className="topbar">
                    <div className="topbar-left">
                        {siteId ? (
                            <div className="topbar-site-stack">
                                <span className="domain-badge">{siteId}</span>
                            </div>
                        ) : (
                            <span className="topbar-title-muted">
                                {sitesLoading ? "Loading sites…" : "No sites found"}
                            </span>
                        )}
                    </div>

                    <div className="topbar-right">
                        {!sitesLoading && sites.length > 0 && (
                            <select
                                className="input"
                                value={selectedSite?.site_id ?? ""}
                                onChange={handleSiteChange}
                                id="site-picker"
                            >
                                {sites.map((s) => (
                                    <option key={s.site_id} value={s.site_id}>
                                        {s.site_id}
                                    </option>
                                ))}
                            </select>
                        )}

                        {selectedSite && (
                            <button
                                className={`btn btn-ghost btn-refresh ${loading ? "is-loading" : ""}`}
                                id="refresh-btn"
                                onClick={handleRefresh}
                                disabled={loading}
                                title="Refresh data"
                            >
                                ↻
                            </button>
                        )}

                        <div className="date-range-group" role="group" aria-label="Date ranges">
                            {DATE_PRESETS.map((p) => (
                                <button
                                    key={p.key}
                                    className={`btn btn-ghost ${preset === p.key ? "active" : ""}`}
                                    onClick={() => handlePreset(p.key)}
                                    id={`preset-${p.key}`}
                                >
                                    {p.label}
                                </button>
                            ))}
                        </div>

                        <button
                            className="theme-toggle"
                            onClick={toggleTheme}
                            title={`Switch to ${theme === "dark" ? "light" : "dark"} mode`}
                        >
                            {theme === "dark" ? <SunIcon /> : <MoonIcon />}
                        </button>
                    </div>
                </header>

                <main className="page-content">
                    {sitesLoading && (
                        <div style={centeredStyle}>
                            <div className="status-glyph">IR</div>
                            <div className="status-title">Loading sites…</div>
                        </div>
                    )}

                    {!sitesLoading && sites.length === 0 && (
                        <div style={centeredStyle}>
                            <div className="status-glyph">00</div>
                            <div className="status-title">No data yet</div>
                            <div className="status-subtitle">
                                Install the Iris SDK on your site to start collecting analytics.
                            </div>
                        </div>
                    )}

                    {!sitesLoading && selectedSite && (
                        <>
                            <section className="overview-hero">
                                <span className="overview-eyebrow">Traffic snapshot</span>
                                <h1 className="overview-title">{getRangeLabel(preset)}</h1>
                                <p className="overview-subtitle">{formatWindow(windowRange, preset)}</p>
                            </section>

                            <StatsCards stats={stats} loading={loading} />

                            {activeTab === "overview" && (
                                <div className="dashboard-grid">
                                    <PageviewsChart
                                        pageviewsData={chartData.length ? chartData : emptyBuckets}
                                        visitorsData={visitorsData.length ? visitorsData : emptyBuckets}
                                        loading={loading}
                                        from={windowRange.from}
                                        to={windowRange.to}
                                    />
                                    <TopPages pages={pages} loading={loading} />
                                    <TopReferrers referrers={referrers} loading={loading} />
                                    <WebVitals vitals={vitals} loading={loading} />
                                    <DeviceBreakdown devices={devices} loading={loading} />
                                </div>
                            )}

                            {activeTab === "pages" && (
                                <div className="single-column-panel">
                                    <TopPages pages={pages} loading={loading} />
                                </div>
                            )}

                            {activeTab === "referrers" && (
                                <div className="single-column-panel">
                                    <TopReferrers referrers={referrers} loading={loading} />
                                </div>
                            )}

                            {activeTab === "vitals" && (
                                <div className="single-column-panel">
                                    <WebVitals vitals={vitals} loading={loading} />
                                </div>
                            )}

                            {activeTab === "devices" && (
                                <div className="single-column-panel">
                                    <DeviceBreakdown devices={devices} loading={loading} />
                                </div>
                            )}
                        </>
                    )}
                </main>
            </div>
        </div>
    );
}

const centeredStyle: React.CSSProperties = {
    display: "flex",
    flexDirection: "column",
    alignItems: "center",
    justifyContent: "center",
    flex: 1,
    gap: 16,
    color: "var(--text-tertiary)",
};
