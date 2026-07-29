import { useCallback, useEffect, useRef, useState } from "react";
import { format, subDays, subHours } from "date-fns";

import {
    api,
    DeviceStat,
    PagePerformanceStat,
    PageStat,
    PerformanceScore,
    ReferrerStat,
    SiteStat,
    SiteTrendResult,
    StatsResult,
    VitalDistribution,
    VitalStat,
} from "./api";
import { DashboardShell, DashboardView } from "./components/DashboardShell";
import { EventsPage } from "./components/EventsPage";
import { OverviewPage } from "./components/OverviewPage";
import { SitesPage, SiteSummary } from "./components/SitesPage";
import { VitalsPage } from "./components/VitalsPage";
import { buildEmptyBuckets, DayBucket } from "./components/PageviewsChart";

export type PresetKey = "24h" | "7d" | "30d" | "90d";

interface DatePreset {
    key: PresetKey;
    label: string;
    unit: "hours" | "days";
    amount: number;
}

export interface DateWindow {
    from: Date;
    to: Date;
    queryFrom: string;
    queryTo: string;
}

export const DATE_PRESETS: DatePreset[] = [
    { key: "24h", label: "Last 24 Hours", unit: "hours", amount: 24 },
    { key: "7d", label: "Last 7 Days", unit: "days", amount: 7 },
    { key: "30d", label: "Last 30 Days", unit: "days", amount: 30 },
    { key: "90d", label: "Last 90 Days", unit: "days", amount: 90 },
];

const PAGE_TITLES: Record<DashboardView, string> = {
    dashboard: "Overview",
    sites: "Sites",
    events: "Custom Events",
    vitals: "Web Vitals",
};

function buildWindow(preset: PresetKey): DateWindow {
    const config = DATE_PRESETS.find((item) => item.key === preset) ?? DATE_PRESETS[2];
    const to = new Date();
    const from = config.unit === "hours" ? subHours(to, config.amount) : subDays(to, config.amount);

    return {
        from,
        to,
        queryFrom: config.unit === "hours"
            ? format(from, "yyyy-MM-dd'T'HH:mm:ssxxx")
            : format(from, "yyyy-MM-dd"),
        queryTo: config.unit === "hours"
            ? format(to, "yyyy-MM-dd'T'HH:mm:ssxxx")
            : format(to, "yyyy-MM-dd"),
    };
}

export default function App() {
    const [view, setView] = useState<DashboardView>("dashboard");
    const [sites, setSites] = useState<SiteStat[]>([]);
    const [sitesLoading, setSitesLoading] = useState(true);
    const [selectedSite, setSelectedSite] = useState<SiteStat | null>(null);
    const [preset, setPreset] = useState<PresetKey>("30d");
    const [dateWindow, setDateWindow] = useState<DateWindow>(() => buildWindow("30d"));
    const [stats, setStats] = useState<StatsResult | null>(null);
    const [siteTrend, setSiteTrend] = useState<SiteTrendResult | null>(null);
    const [pages, setPages] = useState<PageStat[]>([]);
    const [referrers, setReferrers] = useState<ReferrerStat[]>([]);
    const [vitals, setVitals] = useState<VitalStat[]>([]);
    const [vitalDistributions, setVitalDistributions] = useState<VitalDistribution[]>([]);
    const [pagePerformance, setPagePerformance] = useState<PagePerformanceStat[]>([]);
    const [performanceScore, setPerformanceScore] = useState<PerformanceScore | null>(null);
    const [devices, setDevices] = useState<DeviceStat[]>([]);
    const [pageviews, setPageviews] = useState<DayBucket[]>([]);
    const [visitors, setVisitors] = useState<DayBucket[]>([]);
    const [sessions, setSessions] = useState<DayBucket[]>([]);
    const [siteSummaries, setSiteSummaries] = useState<Record<string, SiteSummary>>({});
    const [loading, setLoading] = useState(false);
    const abortRef = useRef<AbortController | null>(null);

    useEffect(() => {
        api.sites()
            .then((items) => {
                const nextSites = items ?? [];
                setSites(nextSites);
                setSelectedSite(nextSites[0] ?? null);
            })
            .catch((error) => console.error("Iris: failed to fetch sites", error))
            .finally(() => setSitesLoading(false));
    }, []);

    const fetchAnalytics = useCallback(async (siteId: string, range: DateWindow) => {
        abortRef.current?.abort();
        const controller = new AbortController();
        abortRef.current = controller;
        setLoading(true);

        try {
            const [
                nextTrend,
                nextPages,
                nextReferrers,
                nextVitals,
                nextVitalDistributions,
                nextPagePerformance,
                nextPerformanceScore,
                nextDevices,
                nextPageviews,
                nextVisitors,
                nextSessions,
            ] = await Promise.all([
                api.siteTrends(siteId, range.queryFrom, range.queryTo, controller.signal),
                api.pages(siteId, range.queryFrom, range.queryTo, controller.signal),
                api.referrers(siteId, range.queryFrom, range.queryTo, controller.signal),
                api.vitals(siteId, range.queryFrom, range.queryTo, controller.signal),
                api.vitalDistributions(siteId, range.queryFrom, range.queryTo, controller.signal),
                api.pagePerformance(siteId, range.queryFrom, range.queryTo, controller.signal),
                api.performanceScore(siteId, range.queryFrom, range.queryTo, controller.signal),
                api.devices(siteId, range.queryFrom, range.queryTo, controller.signal),
                api.timeseries(siteId, range.queryFrom, range.queryTo, controller.signal),
                api.uniqueVisitorsTimeseries(siteId, range.queryFrom, range.queryTo, controller.signal),
                api.sessionsTimeseries(siteId, range.queryFrom, range.queryTo, controller.signal),
            ]);

            setStats(nextTrend.current);
            setSiteTrend(nextTrend);
            setPages(nextPages ?? []);
            setReferrers(nextReferrers ?? []);
            setVitals(nextVitals ?? []);
            setVitalDistributions(nextVitalDistributions ?? []);
            setPagePerformance(nextPagePerformance ?? []);
            setPerformanceScore(nextPerformanceScore);
            setDevices(nextDevices ?? []);
            setPageviews(nextPageviews ?? []);
            setVisitors(nextVisitors ?? []);
            setSessions(nextSessions ?? []);
        } catch (error) {
            if (error instanceof DOMException && error.name === "AbortError") return;
            console.error("Iris: failed to fetch analytics", error);
        } finally {
            if (!controller.signal.aborted) setLoading(false);
        }
    }, []);

    useEffect(() => {
        if (selectedSite) {
            fetchAnalytics(selectedSite.site_id, dateWindow);
        }
    }, [dateWindow, fetchAnalytics, selectedSite]);

    useEffect(() => {
        if (view !== "sites" || sites.length === 0) return;

        let cancelled = false;
        Promise.all(
            sites.map(async (site) => {
                const summary = await api.siteTrends(site.site_id, dateWindow.queryFrom, dateWindow.queryTo);
                return [site.site_id, summary] as const;
            }),
        )
            .then((entries) => {
                if (!cancelled) setSiteSummaries(Object.fromEntries(entries));
            })
            .catch((error) => console.error("Iris: failed to fetch site summaries", error));

        return () => {
            cancelled = true;
        };
    }, [dateWindow, sites, view]);

    function handlePreset(nextPreset: PresetKey) {
        setPreset(nextPreset);
        setDateWindow(buildWindow(nextPreset));
    }

    function handleSiteChange(siteId: string) {
        const nextSite = sites.find((site) => site.site_id === siteId) ?? null;
        setSelectedSite(nextSite);
    }

    function handleRefresh() {
        setDateWindow(buildWindow(preset));
    }

    function handleViewChange(nextView: DashboardView) {
        setView(nextView);
    }

    const emptyBuckets = buildEmptyBuckets(dateWindow.from, dateWindow.to);
    const hasSites = sites.length > 0;

    return (
        <DashboardShell
            view={view}
            title={PAGE_TITLES[view]}
            sites={sites}
            selectedSiteId={selectedSite?.site_id ?? ""}
            preset={preset}
            loading={loading}
            onNavigate={handleViewChange}
            onSiteChange={handleSiteChange}
            onPresetChange={handlePreset}
            onRefresh={handleRefresh}
        >
            {sitesLoading ? (
                <div className="page-state">
                    <span className="spinner" />
                    <strong>Loading your sites</strong>
                </div>
            ) : !hasSites ? (
                <div className="page-state page-state-empty">
                    <div className="empty-logo">Ir</div>
                    <strong>No analytics data yet</strong>
                    <p>Install the Iris SDK on your site to start collecting privacy-friendly analytics.</p>
                </div>
            ) : selectedSite && view === "dashboard" ? (
                <OverviewPage
                    stats={stats}
                    siteTrend={siteTrend}
                    pages={pages}
                    referrers={referrers}
                    devices={devices}
                    pageviews={pageviews.length ? pageviews : emptyBuckets}
                    visitors={visitors.length ? visitors : emptyBuckets}
                    sessions={sessions.length ? sessions : emptyBuckets}
                    dateWindow={dateWindow}
                    loading={loading}
                />
            ) : view === "sites" ? (
                <SitesPage
                    sites={sites}
                    summaries={siteSummaries}
                    selectedSiteId={selectedSite?.site_id ?? ""}
                    onOpenSite={(siteId) => {
                        handleSiteChange(siteId);
                        setView("dashboard");
                    }}
                />
            ) : selectedSite && view === "events" ? (
                <EventsPage site={selectedSite} dateWindow={dateWindow} />
            ) : selectedSite ? (
                <VitalsPage
                    vitals={vitals}
                    distributions={vitalDistributions}
                    pages={pagePerformance}
                    score={performanceScore}
                    loading={loading}
                />
            ) : null}
        </DashboardShell>
    );
}
