import type { StatsResult } from "../api";

interface Props {
    stats: StatsResult | null;
    loading: boolean;
}

const CARDS = [
    {
        key: "pageviews" as const,
        label: "Pageviews",
        icon: "PV",
        tone: "accent" as const,
    },
    {
        key: "unique_visitors" as const,
        label: "Unique Visitors",
        icon: "UV",
        tone: "green" as const,
    },
    {
        key: "sessions" as const,
        label: "Sessions",
        icon: "SE",
        tone: "blue" as const,
    },
];

function fmt(n: number): string {
    if (n >= 1_000_000) return (n / 1_000_000).toFixed(1) + "M";
    if (n >= 1_000) return (n / 1_000).toFixed(1) + "K";
    return n.toLocaleString();
}

export function StatsCards({ stats, loading }: Props) {
    return (
        <div className="stats-grid">
            {CARDS.map((c) => (
                <div className="stat-card" key={c.key}>
                    <div className="stat-card-header">
                        <div className="stat-card-label">
                            <div className={`stat-card-icon ${c.tone}`}>
                                {c.icon}
                            </div>
                            {c.label}
                        </div>
                    </div>
                    <div className="stat-card-value">
                        {loading ? (
                            <span className="stat-card-pending" style={{ opacity: 0.3 }}>—</span>
                        ) : (
                            fmt(stats?.[c.key] ?? 0)
                        )}
                    </div>
                </div>
            ))}
        </div>
    );
}
