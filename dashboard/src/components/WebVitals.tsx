import type { VitalStat } from "../api";

interface Props {
    vitals: VitalStat[];
    loading: boolean;
}

// Thresholds from web.dev/vitals
const THRESHOLDS: Record<string, [number, number]> = {
    LCP: [2500, 4000],   // ms
    INP: [200, 500],     // ms
    CLS: [0.1, 0.25],   // unitless
};

const UNITS: Record<string, string> = {
    LCP: "ms",
    INP: "ms",
    CLS: "",
};

function getRating(name: string, value: number): "good" | "needs-improvement" | "poor" {
    const thresholds = THRESHOLDS[name];
    if (!thresholds) return "good";
    if (value <= thresholds[0]) return "good";
    if (value <= thresholds[1]) return "needs-improvement";
    return "poor";
}

function fmtValue(name: string, value: number): string {
    if (name === "CLS") return value.toFixed(3);
    return Math.round(value).toLocaleString();
}

const RATING_LABELS: Record<string, string> = {
    "good": "Good",
    "needs-improvement": "Needs Work",
    "poor": "Poor",
};

export function WebVitals({ vitals, loading }: Props) {
    // Ensure CLS, INP, LCP are shown in a fixed order
    const ordered = ["LCP", "INP", "CLS"].map((name) => {
        const found = vitals.find((v) => v.name === name);
        return { name, value: found?.value ?? null };
    });

    return (
        <div className="card">
            <div className="card-header">
                <span className="card-title">Web Vitals</span>
                <span style={{ fontSize: 11, color: "var(--text-muted)" }}>
                    Avg across sessions
                </span>
            </div>
            <div className="card-body">
                {loading ? (
                    <div className="state-center">
                        <div className="spinner" />
                        Loading…
                    </div>
                ) : (
                    <div className="vitals-grid">
                        {ordered.map(({ name, value }) => {
                            const rating = value !== null ? getRating(name, value) : "unknown";
                            const unit = UNITS[name] ?? "";
                            return (
                                <div className="vital-item" key={name}>
                                    <div className="vital-name">{name}</div>
                                    <div className={`vital-value ${rating}`}>
                                        {value !== null ? `${fmtValue(name, value)}${unit}` : "—"}
                                    </div>
                                    {value !== null && (
                                        <div className={`vital-badge ${rating}`}>
                                            {RATING_LABELS[rating]}
                                        </div>
                                    )}
                                </div>
                            );
                        })}
                    </div>
                )}
            </div>
        </div>
    );
}
