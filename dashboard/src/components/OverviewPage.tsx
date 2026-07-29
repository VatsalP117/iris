import type { DateWindow } from "../App";
import type { DeviceStat, PageStat, ReferrerStat, StatsResult } from "../api";
import { DeviceBreakdown } from "./DeviceBreakdown";
import { Icon, IconName } from "./Icon";
import { DayBucket, PageviewsChart } from "./PageviewsChart";
import { TopPages } from "./TopPages";
import { TopReferrers } from "./TopReferrers";

interface Props {
    stats: StatsResult | null;
    pages: PageStat[];
    referrers: ReferrerStat[];
    devices: DeviceStat[];
    pageviews: DayBucket[];
    visitors: DayBucket[];
    sessions: DayBucket[];
    dateWindow: DateWindow;
    loading: boolean;
}

interface Metric {
    label: string;
    value: string;
    helper: string;
    tone: "positive" | "neutral";
    icon: IconName;
}

function formatMetric(value: number | undefined): string {
    const safeValue = value ?? 0;
    if (safeValue >= 1_000_000) return `${(safeValue / 1_000_000).toFixed(1)}M`;
    if (safeValue >= 1_000) return `${(safeValue / 1_000).toFixed(1)}K`;
    return safeValue.toLocaleString();
}

export function OverviewPage({
    stats,
    pages,
    referrers,
    devices,
    pageviews,
    visitors,
    sessions,
    dateWindow,
    loading,
}: Props) {
    const metrics: Metric[] = [
        {
            label: "Total Pageviews",
            value: formatMetric(stats?.pageviews),
            helper: "All tracked page loads",
            tone: "positive",
            icon: "activity",
        },
        {
            label: "Unique Visitors",
            value: formatMetric(stats?.unique_visitors),
            helper: "Privacy-safe identities",
            tone: "positive",
            icon: "users",
        },
        {
            label: "Sessions",
            value: formatMetric(stats?.sessions),
            helper: "Across the selected period",
            tone: "positive",
            icon: "trend",
        },
        {
            label: "Tracked Pages",
            value: pages.length.toLocaleString(),
            helper: "With recorded activity",
            tone: "neutral",
            icon: "globe",
        },
    ];

    return (
        <div className="page-stack">
            <section className="metric-grid" aria-label="Analytics summary">
                {metrics.map((metric) => (
                    <article className="metric-card" key={metric.label}>
                        <div className="metric-card-heading">
                            <span>{metric.label}</span>
                            <div className="metric-icon"><Icon name={metric.icon} size={16} /></div>
                        </div>
                        <strong>{loading ? "—" : metric.value}</strong>
                        <small className={metric.tone}>{metric.helper}</small>
                    </article>
                ))}
            </section>

            <PageviewsChart
                pageviewsData={pageviews}
                visitorsData={visitors}
                sessionsData={sessions}
                loading={loading}
                from={dateWindow.from}
                to={dateWindow.to}
            />

            <section className="overview-detail-grid">
                <TopPages pages={pages.slice(0, 6)} loading={loading} />
                <TopReferrers referrers={referrers.slice(0, 6)} loading={loading} />
                <DeviceBreakdown devices={devices} loading={loading} />
            </section>

            <footer className="privacy-footer">
                <span>Privacy-first tracking</span>
                <span>Zero personal data collected</span>
                <span>Data stays on your infrastructure</span>
            </footer>
        </div>
    );
}
