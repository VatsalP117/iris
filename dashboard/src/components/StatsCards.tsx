import type { StatsResult } from "../api";

interface Props {
    stats: StatsResult | null;
    loading: boolean;
}

const CARDS = [
    {
        key: "pageviews" as const,
        label: "Pageviews",
    },
    {
        key: "unique_visitors" as const,
        label: "Unique Visitors",
    },
    {
        key: "sessions" as const,
        label: "Sessions",
    },
];

function fmt(n: number): string {
    if (n >= 1_000_000) return (n / 1_000_000).toFixed(1) + "M";
    if (n >= 1_000) return (n / 1_000).toFixed(1) + "K";
    return n.toLocaleString();
}

export function StatsCards({ stats, loading }: Props) {
    return (
        <div className="stats-grid animate-in animate-in-delay-1">
            {CARDS.map((c, i) => (
                <div className="stat-card" key={c.key} style={{ animationDelay: `${0.05 + i * 0.05}s` }}>
                    <div className="stat-card-label">{c.label}</div>
                    <div className="stat-card-value">
                        {loading ? (
                            <span className="stat-card-pending">\u2014</span>
                        ) : (
                            fmt(stats?.[c.key] ?? 0)
                        )}
                    </div>
                </div>
            ))}
        </div>
    );
}
