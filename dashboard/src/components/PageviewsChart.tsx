import { useState, useMemo } from "react";
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

const METRICS: { key: MetricKey; label: string; color: string }[] = [
    { key: "uniqueVisitors", label: "Visitors", color: "#c25e00" },
    { key: "pageviews", label: "Pageviews", color: "#1c1917" },
    { key: "sessions", label: "Sessions", color: "#57534e" },
];

const CustomTooltip = ({ active, payload, label }: any) => {
    if (!active || !payload?.length) return null;
    return (
        <div
            style={{
                background: "var(--bg-surface)",
                border: "1px solid var(--border)",
                borderRadius: "var(--radius-md)",
                padding: "12px 16px",
                fontSize: 12,
                minWidth: 170,
                boxShadow: "var(--shadow-lg)",
                fontFamily: "var(--font-sans)",
            }}
        >
            <div style={{
                color: "var(--text-tertiary)",
                marginBottom: 8,
                fontSize: 11,
                textTransform: "uppercase",
                letterSpacing: "0.06em",
                fontFamily: "var(--font-mono)",
            }}>
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
                            gap: 20,
                            marginTop: 4,
                        }}
                    >
                        <span style={{ color: "var(--text-secondary)", display: "flex", alignItems: "center", gap: 8 }}>
                            <span
                                style={{
                                    width: 6,
                                    height: 6,
                                    borderRadius: "50%",
                                    background: entry.color,
                                    display: "inline-block",
                                }}
                            />
                            {metric?.label ?? entry.dataKey}
                        </span>
                        <span style={{
                            color: "var(--text-primary)",
                            fontWeight: 600,
                            fontFamily: "var(--font-mono)",
                            fontSize: 13,
                        }}>
                            {entry.value.toLocaleString()}
                        </span>
                    </div>
                );
            })}
        </div>
    );
};

export function PageviewsChart({ pageviewsData, visitorsData, sessionsData, loading, from, to }: Props) {
    const [selectedMetrics, setSelectedMetrics] = useState<MetricKey[]>(["uniqueVisitors", "pageviews"]);

    const chartData = useMemo(() => {
        const byDatePV = Object.fromEntries(pageviewsData.map((d) => [d.date, d.pageviews ?? 0]));
        const byDateUV = Object.fromEntries(visitorsData.map((d) => [d.date, d.uniqueVisitors ?? 0]));
        const byDateSess = Object.fromEntries(sessionsData.map((d) => [d.date, d.sessions ?? 0]));

        const allDates = Array.from(new Set([
            ...pageviewsData.map((d) => d.date),
            ...visitorsData.map((d) => d.date),
            ...sessionsData.map((d) => d.date),
        ])).sort();

        return allDates.map((date) => ({
            date,
            pageviews: byDatePV[date] ?? 0,
            uniqueVisitors: byDateUV[date] ?? 0,
            sessions: byDateSess[date] ?? 0,
        }));
    }, [pageviewsData, visitorsData, sessionsData]);

    const isShortWindow = to.getTime() - from.getTime() <= 36 * 60 * 60 * 1000;
    const windowLabel = isShortWindow
        ? `${format(from, "MMM d, p")} \u2014 ${format(to, "MMM d, p")}`
        : `${format(from, "MMM d")} \u2014 ${format(to, "MMM d, yyyy")}`;

    function toggleMetric(key: MetricKey) {
        setSelectedMetrics((prev) =>
            prev.includes(key) ? prev.filter((k) => k !== key) : [...prev, key]
        );
    }

    return (
        <div className="card">
            <div className="card-header">
                <div style={{ display: "flex", alignItems: "center", gap: 16 }}>
                    <span className="card-title">Traffic</span>
                    <div className="metric-toggle-group">
                        {METRICS.map((metric) => (
                            <button
                                key={metric.key}
                                className={`metric-toggle ${selectedMetrics.includes(metric.key) ? "active" : ""}`}
                                onClick={() => toggleMetric(metric.key)}
                            >
                                <span
                                    className="metric-toggle-dot"
                                    style={{ background: metric.color }}
                                />
                                {metric.label}
                            </button>
                        ))}
                    </div>
                </div>
                <span className="card-meta">{windowLabel}</span>
            </div>
            <div className="card-body">
                <div className="chart-container">
                    {loading ? (
                        <div className="state-center" style={{ height: "100%" }}>
                            <div className="spinner" />
                        </div>
                    ) : (
                        <ResponsiveContainer width="100%" height="100%">
                            <AreaChart data={chartData} margin={{ top: 10, right: 10, left: -20, bottom: 0 }}>
                                <defs>
                                    <linearGradient id="gradUV" x1="0" y1="0" x2="0" y2="1">
                                        <stop offset="0%" stopColor="#c25e00" stopOpacity={0.12} />
                                        <stop offset="100%" stopColor="#c25e00" stopOpacity={0} />
                                    </linearGradient>
                                    <linearGradient id="gradPV" x1="0" y1="0" x2="0" y2="1">
                                        <stop offset="0%" stopColor="#1c1917" stopOpacity={0.08} />
                                        <stop offset="100%" stopColor="#1c1917" stopOpacity={0} />
                                    </linearGradient>
                                    <linearGradient id="gradSess" x1="0" y1="0" x2="0" y2="1">
                                        <stop offset="0%" stopColor="#57534e" stopOpacity={0.08} />
                                        <stop offset="100%" stopColor="#57534e" stopOpacity={0} />
                                    </linearGradient>
                                </defs>
                                <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="var(--border-light)" />
                                <XAxis
                                    dataKey="date"
                                    tick={{ fontSize: 11, fill: "var(--text-tertiary)", fontFamily: "var(--font-mono)" }}
                                    tickFormatter={(v) => format(parseISO(v), "MMM d")}
                                    axisLine={false}
                                    tickLine={false}
                                />
                                <YAxis
                                    tick={{ fontSize: 11, fill: "var(--text-tertiary)", fontFamily: "var(--font-mono)" }}
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
                                        fill={
                                            metric.key === "uniqueVisitors"
                                                ? "url(#gradUV)"
                                                : metric.key === "pageviews"
                                                    ? "url(#gradPV)"
                                                    : "url(#gradSess)"
                                        }
                                        dot={false}
                                        activeDot={{ r: 4, fill: metric.color, strokeWidth: 0 }}
                                    />
                                ))}
                            </AreaChart>
                        </ResponsiveContainer>
                    )}
                </div>
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
