import { useState, useEffect, useCallback } from "react";
import { format, subDays } from "date-fns";
import { api, StatsResult, PageStat, ReferrerStat, VitalStat, DeviceStat, SiteStat } from "./api";
import { StatsCards } from "./components/StatsCards";
import { TopPages } from "./components/TopPages";
import { TopReferrers } from "./components/TopReferrers";
import { WebVitals } from "./components/WebVitals";
import { DeviceBreakdown } from "./components/DeviceBreakdown";
import { PageviewsChart, buildEmptyBuckets, DayBucket } from "./components/PageviewsChart";

const DATE_PRESETS = [
    { label: "7d", days: 7 },
    { label: "30d", days: 30 },
    { label: "90d", days: 90 },
];

function fmtDate(d: Date): string {
    return format(d, "yyyy-MM-dd");
}

type Tab = "overview" | "pages" | "referrers" | "vitals" | "devices";

export default function App() {
    const [sites, setSites] = useState<SiteStat[]>([]);
    const [sitesLoading, setSitesLoading] = useState(true);
    const [selectedSite, setSelectedSite] = useState<SiteStat | null>(null);

    const [activeTab, setActiveTab] = useState<Tab>("overview");
    const [preset, setPreset] = useState(30);
    const [from, setFrom] = useState<Date>(() => subDays(new Date(), 30));
    const [to] = useState<Date>(new Date());

    const [stats, setStats] = useState<StatsResult | null>(null);
    const [pages, setPages] = useState<PageStat[]>([]);
    const [referrers, setReferrers] = useState<ReferrerStat[]>([]);
    const [vitals, setVitals] = useState<VitalStat[]>([]);
    const [devices, setDevices] = useState<DeviceStat[]>([]);
    const [chartData, setChartData] = useState<DayBucket[]>([]);
    const [loading, setLoading] = useState(false);

    const fromStr = fmtDate(from);
    const toStr = fmtDate(to);

    // Fetch available sites on mount and auto-select first
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

    const fetchAll = useCallback(async (domain: string) => {
        if (!domain) return;
        setLoading(true);
        try {
            const [s, p, r, v, dev, ts] = await Promise.all([
                api.stats(domain, fromStr, toStr),
                api.pages(domain, fromStr, toStr),
                api.referrers(domain, fromStr, toStr),
                api.vitals(domain, fromStr, toStr),
                api.devices(domain, fromStr, toStr),
                api.timeseries(domain, fromStr, toStr),
            ]);
            setStats(s);
            setPages(p ?? []);
            setReferrers(r ?? []);
            setVitals(v ?? []);
            setDevices(dev ?? []);
            setChartData(ts ?? []);
        } catch (err) {
            console.error("Iris: fetch error", err);
        } finally {
            setLoading(false);
        }
    }, [fromStr, toStr]);

    useEffect(() => {
        if (selectedSite) fetchAll(selectedSite.domain);
    }, [selectedSite, fetchAll]);

    function handlePreset(days: number) {
        setPreset(days);
        setFrom(subDays(new Date(), days));
    }

    function handleSiteChange(e: React.ChangeEvent<HTMLSelectElement>) {
        const site = sites.find((s) => s.site_id === e.target.value) ?? null;
        setSelectedSite(site);
        // Reset data so stale values don't flash
        setStats(null);
        setPages([]);
        setReferrers([]);
        setVitals([]);
        setDevices([]);
        setChartData([]);
    }

    const emptyBuckets = buildEmptyBuckets(from, to);
    const domain = selectedSite?.domain ?? "";

    return (
        <div className="app-layout">
            {/* Sidebar */}
            <aside className="sidebar">
                <div className="sidebar-logo">
                    <div className="sidebar-logo-icon">👁</div>
                    <span className="sidebar-logo-text">
                        <span>iris</span>
                    </span>
                </div>

                <span className="sidebar-section-label">Navigation</span>

                <SidebarItem icon="📊" label="Overview" active={activeTab === "overview"} onClick={() => setActiveTab("overview")} />
                <SidebarItem icon="📄" label="Pages" active={activeTab === "pages"} onClick={() => setActiveTab("pages")} />
                <SidebarItem icon="🔗" label="Referrers" active={activeTab === "referrers"} onClick={() => setActiveTab("referrers")} />
                <SidebarItem icon="⚡" label="Web Vitals" active={activeTab === "vitals"} onClick={() => setActiveTab("vitals")} />
                <SidebarItem icon="📱" label="Devices" active={activeTab === "devices"} onClick={() => setActiveTab("devices")} />

                <div style={{ flex: 1 }} />

                <div style={{ padding: "0 16px 8px" }}>
                    <div
                        style={{
                            background: "var(--accent-dim)",
                            border: "1px solid rgba(108,122,255,0.15)",
                            borderRadius: "var(--radius-md)",
                            padding: "12px",
                            fontSize: 12,
                            color: "var(--text-secondary)",
                            lineHeight: 1.5,
                        }}
                    >
                        <span style={{ color: "var(--accent)", fontWeight: 600 }}>iris</span>{" "}
                        v0.1.0 · Self-hosted
                    </div>
                </div>
            </aside>

            {/* Main */}
            <div className="main-area">
                {/* Topbar */}
                <header className="topbar">
                    <div className="topbar-left">
                        {domain ? (
                            <span className="domain-badge">{domain}</span>
                        ) : (
                            <span className="topbar-title" style={{ color: "var(--text-muted)" }}>
                                {sitesLoading ? "Loading sites…" : "No sites found"}
                            </span>
                        )}
                    </div>

                    <div className="topbar-right">
                        {/* Site picker */}
                        {!sitesLoading && sites.length > 0 && (
                            <select
                                className="input"
                                value={selectedSite?.site_id ?? ""}
                                onChange={handleSiteChange}
                                id="site-picker"
                                style={{ minWidth: 200 }}
                            >
                                {sites.map((s) => (
                                    <option key={s.site_id} value={s.site_id}>
                                        {s.domain}
                                    </option>
                                ))}
                            </select>
                        )}

                        {/* Refresh */}
                        {selectedSite && (
                            <button
                                className="btn btn-ghost"
                                id="refresh-btn"
                                onClick={() => fetchAll(selectedSite.domain)}
                                disabled={loading}
                                title="Refresh data"
                                style={{ fontSize: 15, transition: "transform 0.4s", transform: loading ? "rotate(360deg)" : "none" }}
                            >
                                ↻
                            </button>
                        )}

                        {/* Date presets */}
                        <div className="date-range-group">
                            {DATE_PRESETS.map((p) => (
                                <button
                                    key={p.label}
                                    className={`btn btn-ghost ${preset === p.days ? "active" : ""}`}
                                    onClick={() => handlePreset(p.days)}
                                    id={`preset-${p.label}`}
                                >
                                    {p.label}
                                </button>
                            ))}
                        </div>
                    </div>
                </header>

                {/* Page content */}
                <main className="page-content">
                    {/* Empty / loading states */}
                    {sitesLoading && (
                        <div style={centeredStyle}>
                            <div style={{ fontSize: 48 }}>👁</div>
                            <div style={{ fontSize: 18, fontWeight: 600, color: "var(--text-secondary)" }}>
                                Loading sites…
                            </div>
                        </div>
                    )}

                    {!sitesLoading && sites.length === 0 && (
                        <div style={centeredStyle}>
                            <div style={{ fontSize: 48 }}>📭</div>
                            <div style={{ fontSize: 18, fontWeight: 600, color: "var(--text-secondary)" }}>
                                No data yet
                            </div>
                            <div style={{ fontSize: 13, color: "var(--text-muted)" }}>
                                Install the Iris SDK on your site to start collecting analytics.
                            </div>
                        </div>
                    )}

                    {!sitesLoading && selectedSite && (
                        <>
                            <StatsCards stats={stats} loading={loading} />

                            {activeTab === "overview" && (
                                <div className="dashboard-grid">
                                    <PageviewsChart
                                        data={chartData.length ? chartData : emptyBuckets}
                                        loading={loading}
                                        from={from}
                                        to={to}
                                    />
                                    <TopPages pages={pages} loading={loading} />
                                    <TopReferrers referrers={referrers} loading={loading} />
                                    <WebVitals vitals={vitals} loading={loading} />
                                    <DeviceBreakdown devices={devices} loading={loading} />
                                </div>
                            )}

                            {activeTab === "pages" && (
                                <div style={{ marginTop: 24, maxWidth: 800 }}>
                                    <TopPages pages={pages} loading={loading} />
                                </div>
                            )}

                            {activeTab === "referrers" && (
                                <div style={{ marginTop: 24, maxWidth: 800 }}>
                                    <TopReferrers referrers={referrers} loading={loading} />
                                </div>
                            )}

                            {activeTab === "vitals" && (
                                <div style={{ marginTop: 24, maxWidth: 800 }}>
                                    <WebVitals vitals={vitals} loading={loading} />
                                </div>
                            )}

                            {activeTab === "devices" && (
                                <div style={{ marginTop: 24, maxWidth: 800 }}>
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
    color: "var(--text-muted)",
};

function SidebarItem({
    icon,
    label,
    active = false,
    onClick,
}: {
    icon: string;
    label: string;
    active?: boolean;
    onClick?: () => void;
}) {
    return (
        <button
            onClick={onClick}
            style={{
                display: "flex",
                alignItems: "center",
                gap: 10,
                padding: "8px 20px",
                background: active ? "var(--accent-dim)" : "transparent",
                border: "none",
                borderLeft: active ? "2px solid var(--accent)" : "2px solid transparent",
                color: active ? "var(--accent)" : "var(--text-secondary)",
                fontSize: 13,
                fontWeight: active ? 600 : 400,
                cursor: "pointer",
                width: "100%",
                textAlign: "left",
                transition: "all 0.15s",
                fontFamily: "var(--font-sans)",
            }}
        >
            <span style={{ fontSize: 15 }}>{icon}</span>
            {label}
        </button>
    );
}
