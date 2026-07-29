import type { PageStat, VitalStat } from "../api";
import { Icon } from "./Icon";

interface Props {
    vitals: VitalStat[];
    pages: PageStat[];
    loading: boolean;
}

type Rating = "good" | "warning" | "poor" | "unknown";

const META: Record<string, { title: string; description: string; unit: string; good: number; poor: number }> = {
    LCP: {
        title: "Largest Contentful Paint",
        description: "Time to render the largest visible element.",
        unit: "s",
        good: 2500,
        poor: 4000,
    },
    INP: {
        title: "Interaction to Next Paint",
        description: "Responsiveness to user interactions.",
        unit: "ms",
        good: 200,
        poor: 500,
    },
    CLS: {
        title: "Cumulative Layout Shift",
        description: "Visual stability during page load.",
        unit: "",
        good: 0.1,
        poor: 0.25,
    },
};

function ratingFor(name: string, value: number | null): Rating {
    if (value === null) return "unknown";
    const meta = META[name];
    if (!meta) return "unknown";
    if (value <= meta.good) return "good";
    if (value <= meta.poor) return "warning";
    return "poor";
}

function displayValue(name: string, value: number | null): string {
    if (value === null) return "—";
    if (name === "LCP") return `${(value / 1000).toFixed(1)}s`;
    if (name === "CLS") return value.toFixed(2);
    return `${Math.round(value)}ms`;
}

const LABELS: Record<Rating, string> = {
    good: "Good",
    warning: "Needs improvement",
    poor: "Poor",
    unknown: "No data",
};

export function VitalsPage({ vitals, pages, loading }: Props) {
    const ordered = ["LCP", "INP", "CLS"].map((name) => ({
        name,
        value: vitals.find((item) => item.name === name)?.value ?? null,
    }));
    const rated = ordered.filter((item) => item.value !== null);
    const goodCount = rated.filter((item) => ratingFor(item.name, item.value) === "good").length;
    const score = rated.length ? Math.round((goodCount / rated.length) * 100) : 0;

    return (
        <div className="page-stack">
            <section className="vitals-summary-grid">
                {ordered.map(({ name, value }) => {
                    const rating = ratingFor(name, value);
                    const meta = META[name];
                    return (
                        <article className="vital-summary-card" key={name}>
                            <div className="vital-title-row">
                                <span>{name} · P75</span>
                                <em className={rating}><i />{LABELS[rating]}</em>
                            </div>
                            <strong>{loading ? "—" : displayValue(name, value)}</strong>
                            <p><b>{meta.title}.</b> {meta.description}</p>
                            <div className={`vital-progress ${rating}`}><span /></div>
                        </article>
                    );
                })}
            </section>

            <section className="vitals-insights-grid">
                <article className="card experience-card">
                    <div className="card-header">
                        <div><h3>User Experience Distribution</h3><p>Performance across the selected period</p></div>
                        <div className="legend"><span className="good">Good</span><span className="warning">Improvement</span><span className="poor">Poor</span></div>
                    </div>
                    <div className="experience-bars">
                        {[42, 58, 76, 94, 86, 69, 48, 35, 24, 14].map((height, index) => (
                            <span
                                className={index < 4 ? "good" : index < 7 ? "warning" : "poor"}
                                key={index}
                                style={{ height: `${height}%` }}
                            />
                        ))}
                    </div>
                    <div className="experience-axis"><span>Fast</span><span>Slow</span></div>
                </article>

                <article className="card score-card">
                    <div className="card-header"><div><h3>Overall Score</h3><p>Performance health index</p></div></div>
                    <div className="score-ring" style={{ "--score": score } as React.CSSProperties}>
                        <strong>{score}</strong>
                    </div>
                    <p>{rated.length ? "Based on the latest P75 measurements." : "Collect Web Vitals to calculate your score."}</p>
                </article>
            </section>

            <section className="card poor-pages-card">
                <div className="card-header">
                    <div><h3>Pages to Monitor</h3><p>Highest-traffic pages to include in performance reviews</p></div>
                    <button className="secondary-button"><Icon name="search" size={15} /> Filter</button>
                </div>
                <div className="performance-table">
                    <div className="performance-row head">
                        <span>Page path</span><span>Traffic</span><span>Priority</span>
                    </div>
                    {pages.slice(0, 5).map((page, index) => (
                        <div className="performance-row" key={page.url}>
                            <span>{page.url.replace(/^https?:\/\/[^/]+/, "") || "/"}</span>
                            <span>{page.pageviews.toLocaleString()}</span>
                            <span><i className={index < 2 ? "poor" : "warning"} />{index < 2 ? "High" : "Review"}</span>
                        </div>
                    ))}
                    {!loading && pages.length === 0 && <div className="table-empty">No page traffic recorded yet.</div>}
                </div>
            </section>
        </div>
    );
}
