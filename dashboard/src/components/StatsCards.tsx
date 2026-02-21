import type { StatsResult } from "../api";

interface Props {
    stats: StatsResult | null;
    loading: boolean;
}

const CARDS = [
    {
        key: "pageviews" as const,
        label: "Pageviews",
        icon: "📄",
        color: "#6c7aff",
        bg: "rgba(108,122,255,0.08)",
    },
    {
        key: "unique_visitors" as const,
        label: "Unique Visitors",
        icon: "👤",
        color: "#34d399",
        bg: "rgba(52,211,153,0.08)",
    },
    {
        key: "sessions" as const,
        label: "Sessions",
        icon: "🔗",
        color: "#fbbf24",
        bg: "rgba(251,191,36,0.08)",
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
                    <div className="stat-card-label">
                        <div
                            className="stat-card-icon"
                            style={{ background: c.bg, color: c.color }}
                        >
                            {c.icon}
                        </div>
                        {c.label}
                    </div>
                    <div className="stat-card-value">
                        {loading ? (
                            <span style={{ fontSize: 16, color: "var(--text-muted)" }}>
                                —
                            </span>
                        ) : (
                            fmt(stats?.[c.key] ?? 0)
                        )}
                    </div>
                </div>
            ))}
        </div>
    );
}
