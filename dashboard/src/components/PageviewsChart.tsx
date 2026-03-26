import { useState } from "react";
import {
    XAxis,
    YAxis,
    CartesianGrid,
    Tooltip,
    ResponsiveContainer,
    Area,
    AreaChart,
} from "recharts";
import { format, parseISO } from "date-fns";

export interface DayBucket {
    date: string;
    pageviews?: number;
    uniqueVisitors?: number;
    sessions?: number;
}

interface Props {
    pageviewsData: DayBucket[];
    visitorsData: DayBucket[];
    sessionsData: DayBucket[];
    loading: boolean;
    from: Date;
    to: Date;
}

export type MetricKey = "uniqueVisitors" | "pageviews" | "sessions";

const METRICS: { key: MetricKey; label: string; color: string; gradient: string }[] = [
    { key: "uniqueVisitors", label: "Unique visitors", color: "#0070f3", gradient: "url(#gradientUnique)" },
    { key: "pageviews", label: "Pageviews", color: "#0088ff", gradient: "url(#gradientPageviews)" },
    { key: "sessions", label: "Sessions", color: "#0099ff", gradient: "url(#gradientSessions)" },
];

const CustomTooltip = ({ active, payload, label }: any) => {
    if (!active || !payload?.length) return null;
    return (
        <div
            style={{
                background: "var(--bg-secondary)",
                border: "1px solid var(--border)",
                borderRadius: "var(--radius-md)",
                padding: "10px 14px",
                fontSize: 13,
                minWidth: 160,
            }}
        >
            <div style={{ color: "var(--text-tertiary)", marginBottom: 8 }}>
                {format(parseISO(label), "MMM d, yyyy")}
            </div>
            {payload.map((entry: any, index: number) => {
                const metric = METRICS.find((m) => m.key === entry.dataKey);
                return (
                    <div
                        key={index}
                        style={{
                            display: "flex",
                            justifyContent: "space-between",
                            alignItems: "center",
                            gap: 16,
                            marginTop: 4,
                        }}
                    >
                        <span style={{ color: "var(--text-secondary)", display: "flex", alignItems: "center", gap: 6 }}>
                            <span
                                style={{
                                    width: 8,
                                    height: 8,
                                    borderRadius: "50%",
                                    background: entry.color,
                                    display: "inline-block",
                                }}
                            />
                            {metric?.label ?? entry.dataKey}
                        </span>
                        <span style={{ color: "var(--text-primary)", fontWeight: 600 }}>
                            {entry.value.toLocaleString()}
                        </span>
                    </div>
                );
            })}
        </div>
    );
};

const ChevronIcon = () => (
    <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
        <polyline points="6 9 12 15 18 9" />
    </svg>
);

export function PageviewsChart({ pageviewsData, visitorsData, sessionsData, loading, from, to }: Props) {
    const [selectedMetrics, setSelectedMetrics] = useState<MetricKey[]>(["uniqueVisitors"]);
    const [dropdownOpen, setDropdownOpen] = useState(false);

    const byDatePV = Object.fromEntries(pageviewsData.map((d) => [d.date, d.pageviews ?? 0]));
    const byDateUV = Object.fromEntries(visitorsData.map((d) => [d.date, d.uniqueVisitors ?? 0]));
    const byDateSess = Object.fromEntries(sessionsData.map((d) => [d.date, d.sessions ?? 0]));

    const allDates = Array.from(new Set([
        ...pageviewsData.map((d) => d.date),
        ...visitorsData.map((d) => d.date),
        ...sessionsData.map((d) => d.date),
    ])).sort();

    const chartData = allDates.map((date) => ({
        date,
        pageviews: byDatePV[date] ?? 0,
        uniqueVisitors: byDateUV[date] ?? 0,
        sessions: byDateSess[date] ?? 0,
    }));

    const isShortWindow = to.getTime() - from.getTime() <= 36 * 60 * 60 * 1000;
    const windowLabel = isShortWindow
        ? `${format(from, "MMM d, p")} - ${format(to, "MMM d, p")}`
        : `${format(from, "MMM d")} - ${format(to, "MMM d, yyyy")}`;

    function toggleMetric(key: MetricKey) {
        setSelectedMetrics((prev) =>
            prev.includes(key) ? prev.filter((k) => k !== key) : [...prev, key]
        );
    }

    const selectedLabels = METRICS.filter((m) => selectedMetrics.includes(m.key))
        .map((m) => m.label)
        .join(", ");

    return (
        <div className="card full-width">
            <div className="card-header">
                <div style={{ display: "flex", alignItems: "center", gap: 12 }}>
                    <span className="card-title">Traffic Overview</span>
                    <div className="metric-dropdown">
                        <button
                            className="metric-dropdown-trigger"
                            onClick={() => setDropdownOpen(!dropdownOpen)}
                        >
                            <span>{selectedLabels || "Select metrics"}</span>
                            <ChevronIcon />
                        </button>
                        {dropdownOpen && (
                            <div className="metric-dropdown-menu">
                                {METRICS.map((metric) => (
                                    <label key={metric.key} className="metric-option">
                                        <input
                                            type="checkbox"
                                            checked={selectedMetrics.includes(metric.key)}
                                            onChange={() => toggleMetric(metric.key)}
                                        />
                                        <span
                                            className="metric-color-dot"
                                            style={{ background: metric.color }}
                                        />
                                        <span className="metric-label">{metric.label}</span>
                                    </label>
                                ))}
                            </div>
                        )}
                    </div>
                </div>
                <span className="card-meta">{windowLabel}</span>
            </div>
            <div className="card-body" style={{ height: 280 }}>
                {loading ? (
                    <div className="state-center" style={{ height: "100%" }}>
                        <div className="spinner" />
                        Loading…
                    </div>
                ) : (
                    <ResponsiveContainer width="100%" height="100%">
                        <AreaChart data={chartData} margin={{ top: 8, right: 8, left: -20, bottom: 0 }}>
                            <defs>
                                <linearGradient id="gradientUnique" x1="0" y1="0" x2="0" y2="1">
                                    <stop offset="0%" stopColor="#0070f3" stopOpacity={0.25} />
                                    <stop offset="100%" stopColor="#0070f3" stopOpacity={0} />
                                </linearGradient>
                                <linearGradient id="gradientPageviews" x1="0" y1="0" x2="0" y2="1">
                                    <stop offset="0%" stopColor="#0088ff" stopOpacity={0.2} />
                                    <stop offset="100%" stopColor="#0088ff" stopOpacity={0} />
                                </linearGradient>
                                <linearGradient id="gradientSessions" x1="0" y1="0" x2="0" y2="1">
                                    <stop offset="0%" stopColor="#0099ff" stopOpacity={0.15} />
                                    <stop offset="100%" stopColor="#0099ff" stopOpacity={0} />
                                </linearGradient>
                            </defs>
                            <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="var(--border-light)" />
                            <XAxis
                                dataKey="date"
                                tick={{ fontSize: 11, fill: "var(--text-tertiary)" }}
                                tickFormatter={(v) => format(parseISO(v), "MMM d")}
                                axisLine={false}
                                tickLine={false}
                            />
                            <YAxis
                                tick={{ fontSize: 11, fill: "var(--text-tertiary)" }}
                                axisLine={false}
                                tickLine={false}
                                allowDecimals={false}
                            />
                            <Tooltip content={<CustomTooltip />} />
                            {METRICS.filter((m) => selectedMetrics.includes(m.key)).map((metric) => (
                                <Area
                                    key={metric.key}
                                    type="monotone"
                                    dataKey={metric.key}
                                    stroke={metric.color}
                                    strokeWidth={2}
                                    fill={metric.gradient}
                                    dot={false}
                                    activeDot={{ r: 5, fill: metric.color, strokeWidth: 0 }}
                                />
                            ))}
                        </AreaChart>
                    </ResponsiveContainer>
                )}
            </div>
        </div>
    );
}

export function buildEmptyBuckets(from: Date, to: Date): DayBucket[] {
    const days: DayBucket[] = [];
    const current = new Date(from);
    while (current <= to) {
        days.push({ date: format(current, "yyyy-MM-dd") });
        current.setDate(current.getDate() + 1);
    }
    return days;
}

export { subDays } from "date-fns";
