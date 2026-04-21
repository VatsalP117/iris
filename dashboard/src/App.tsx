import { useState, useEffect, useCallback } from "react";
import { format, subDays, subHours } from "date-fns";
import { api, StatsResult, PageStat, ReferrerStat, VitalStat, DeviceStat, SiteStat } from "./api";
import { StatsCards } from "./components/StatsCards";
import { TopPages } from "./components/TopPages";
import { TopReferrers } from "./components/TopReferrers";
import { WebVitals } from "./components/WebVitals";
import { DeviceBreakdown } from "./components/DeviceBreakdown";
import { PageviewsChart, buildEmptyBuckets, DayBucket } from "./components/PageviewsChart";

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
        return `${format(window.from, "MMM d, p")} \u2014 ${format(window.to, "MMM d, p")}`;
    }
    return `${format(window.from, "MMM d")} \u2014 ${format(window.to, "MMM d, yyyy")}`;
}

export default function App() {
    const [sites, setSites] = useState<SiteStat[]>([]);
    const [sitesLoading, setSitesLoading] = useState(true);
    const [selectedSite, setSelectedSite] = useState<SiteStat | null>(null);

    const [preset, setPreset] = useState<PresetKey>("30d");
    const [windowRange, setWindowRange] = useState<DateWindow>(() => buildWindow("30d"));

    const [stats, setStats] = useState<StatsResult | null>(null);
    const [pages, setPages] = useState<PageStat[]>([]);
    const [referrers, setReferrers] = useState<ReferrerStat[]>([]);
    const [vitals, setVitals] = useState<VitalStat[]>([]);
    const [devices, setDevices] = useState<DeviceStat[]>([]);
    const [chartData, setChartData] = useState<DayBucket[]>([]);
    const [visitorsData, setVisitorsData] = useState<DayBucket[]>([]);
    const [sessionsData, setSessionsData] = useState<DayBucket[]>([]);
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
            const [s, p, r, v, dev, ts, uv, sess] = await Promise.all([
                api.stats(siteId, range.queryFrom, range.queryTo),
                api.pages(siteId, range.queryFrom, range.queryTo),
                api.referrers(siteId, range.queryFrom, range.queryTo),
                api.vitals(siteId, range.queryFrom, range.queryTo),
                api.devices(siteId, range.queryFrom, range.queryTo),
                api.timeseries(siteId, range.queryFrom, range.queryTo),
                api.uniqueVisitorsTimeseries(siteId, range.queryFrom, range.queryTo),
                api.sessionsTimeseries(siteId, range.queryFrom, range.queryTo),
            ]);
            setStats(s);
            setPages(p ?? []);
            setReferrers(r ?? []);
            setVitals(v ?? []);
            setDevices(dev ?? []);
            setChartData(ts ?? []);
            setVisitorsData(uv ?? []);
            setSessionsData(sess ?? []);
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
        setSessionsData([]);
    }

    function handleRefresh() {
        if (!selectedSite) return;
        setWindowRange(buildWindow(preset));
    }

    const emptyBuckets = buildEmptyBuckets(windowRange.from, windowRange.to);
    const siteId = selectedSite?.site_id ?? "";

    return (
        <div className="app-layout">
            {/* Top Navigation */}
            <header className="topbar">
                <div className="topbar-inner">
                    <div className="topbar-brand">
                        <div className="brand-mark">Ir</div>
                        <span className="brand-name">Iris</span>
                    </div>

                    <div className="topbar-center">
                        <div className="date-preset-group">
                            {DATE_PRESETS.map((p) => (
                                <button
                                    key={p.key}
                                    className={`date-preset-btn ${preset === p.key ? "active" : ""}`}
                                    onClick={() => handlePreset(p.key)}
                                >
                                    {p.label}
                                </button>
                            ))}
                        </div>
                    </div>

                    <div className="topbar-right">
                        {!sitesLoading && sites.length > 0 && (
                            <select
                                className="select-control"
                                value={selectedSite?.site_id ?? ""}
                                onChange={handleSiteChange}
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
                                className={`icon-btn ${loading ? "is-loading" : ""}`}
                                onClick={handleRefresh}
                                disabled={loading}
                                title="Refresh data"
                            >
                                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                                    <polyline points="23 4 23 10 17 10"/>
                                    <polyline points="1 20 1 14 7 14"/>
                                    <path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"/>
                                </svg>
                            </button>
                        )}
                    </div>
                </div>
            </header>

            {/* Main Content */}
            <main className="main-area">
                <div className="content-wrapper">
                    {sitesLoading && (
                        <div className="empty-state">
                            <div className="empty-state-icon">Ir</div>
                            <div className="empty-state-title">Loading sites&hellip;</div>
                        </div>
                    )}

                    {!sitesLoading && sites.length === 0 && (
                        <div className="empty-state">
                            <div className="empty-state-icon">00</div>
                            <div className="empty-state-title">No data yet</div>
                            <div className="empty-state-subtitle">
                                Install the Iris SDK on your site to start collecting analytics.
                            </div>
                        </div>
                    )}

                    {!sitesLoading && selectedSite && (
                        <>
                            {/* Header Section */}
                            <div className="section-header animate-in">
                                <div className="section-label">Analytics</div>
                                <h1 className="section-title">{siteId}</h1>
                                <p className="section-subtitle">{formatWindow(windowRange, preset)}</p>
                            </div>

                            {/* Stats */}
                            <StatsCards stats={stats} loading={loading} />

                            {/* Chart */}
                            <div className="animate-in animate-in-delay-1" style={{ marginBottom: "var(--space-8)" }}>
                                <PageviewsChart
                                    pageviewsData={chartData.length ? chartData : emptyBuckets}
                                    visitorsData={visitorsData.length ? visitorsData : emptyBuckets}
                                    sessionsData={sessionsData.length ? sessionsData : emptyBuckets}
                                    loading={loading}
                                    from={windowRange.from}
                                    to={windowRange.to}
                                />
                            </div>

                            {/* Bottom Grid */}
                            <div className="dashboard-grid">
                                <div className="dashboard-sidebar">
                                    <div className="animate-in animate-in-delay-2">
                                        <TopPages pages={pages} loading={loading} />
                                    </div>
                                    <div className="animate-in animate-in-delay-3">
                                        <TopReferrers referrers={referrers} loading={loading} />
                                    </div>
                                </div>
                                <div className="dashboard-sidebar">
                                    <div className="animate-in animate-in-delay-4">
                                        <WebVitals vitals={vitals} loading={loading} />
                                    </div>
                                    <div className="animate-in animate-in-delay-5">
                                        <DeviceBreakdown devices={devices} loading={loading} />
                                    </div>
                                </div>
                            </div>
                        </>
                    )}
                </div>
            </main>
        </div>
    );
}
