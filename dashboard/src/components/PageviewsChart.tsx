import {
    LineChart,
    Line,
    XAxis,
    YAxis,
    CartesianGrid,
    Tooltip,
    ResponsiveContainer,
} from "recharts";
import { format, subDays, eachDayOfInterval, parseISO } from "date-fns";

// This component visualises pageviews over time.
// Since the backend doesn't yet have a time-series endpoint,
// we derive a daily breakdown from the window [from, to] with mock zero-fill
// and let the parent pass pre-aggregated day buckets when available.

export interface DayBucket {
    date: string; // "YYYY-MM-DD"
    pageviews: number;
}

interface Props {
    data: DayBucket[];
    loading: boolean;
    from: Date;
    to: Date;
}

const CustomTooltip = ({ active, payload, label }: any) => {
    if (!active || !payload?.length) return null;
    return (
        <div
            style={{
                background: "var(--bg-elevated)",
                border: "1px solid var(--border)",
                borderRadius: "var(--radius-md)",
                padding: "10px 14px",
                fontSize: 13,
            }}
        >
            <div style={{ color: "var(--text-muted)", marginBottom: 4 }}>
                {format(parseISO(label), "MMM d, yyyy")}
            </div>
            <div style={{ color: "var(--accent)", fontWeight: 600 }}>
                {payload[0].value.toLocaleString()} pageviews
            </div>
        </div>
    );
};

export function PageviewsChart({ data, loading, from, to }: Props) {
    // Zero-fill the entire date range so the chart always shows the full window
    const allDays = eachDayOfInterval({ start: from, end: to });
    const byDate = Object.fromEntries(data.map((d) => [d.date, d.pageviews]));
    const chartData = allDays.map((d) => {
        const key = format(d, "yyyy-MM-dd");
        return { date: key, pageviews: byDate[key] ?? 0 };
    });

    return (
        <div className="card full-width">
            <div className="card-header">
                <span className="card-title">Pageviews Over Time</span>
            </div>
            <div className="card-body" style={{ height: 220 }}>
                {loading ? (
                    <div className="state-center" style={{ height: "100%" }}>
                        <div className="spinner" />
                        Loading…
                    </div>
                ) : (
                    <ResponsiveContainer width="100%" height="100%">
                        <LineChart data={chartData} margin={{ top: 4, right: 8, left: -20, bottom: 0 }}>
                            <CartesianGrid strokeDasharray="3 3" vertical={false} />
                            <XAxis
                                dataKey="date"
                                tick={{ fontSize: 11, fill: "var(--text-muted)" }}
                                tickFormatter={(v) => format(parseISO(v), "MMM d")}
                                axisLine={false}
                                tickLine={false}
                            />
                            <YAxis
                                tick={{ fontSize: 11, fill: "var(--text-muted)" }}
                                axisLine={false}
                                tickLine={false}
                                allowDecimals={false}
                            />
                            <Tooltip content={<CustomTooltip />} />
                            <Line
                                type="monotone"
                                dataKey="pageviews"
                                stroke="var(--accent)"
                                strokeWidth={2}
                                dot={false}
                                activeDot={{ r: 4, fill: "var(--accent)" }}
                            />
                        </LineChart>
                    </ResponsiveContainer>
                )}
            </div>
        </div>
    );
}

// Helper to initialise zero-filled buckets from date range
export function buildEmptyBuckets(from: Date, to: Date): DayBucket[] {
    return eachDayOfInterval({ start: from, end: to }).map((d) => ({
        date: format(d, "yyyy-MM-dd"),
        pageviews: 0,
    }));
}

export { subDays };
