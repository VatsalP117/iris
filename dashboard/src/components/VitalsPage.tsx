import type {
    PagePerformanceStat,
    PerformanceScore,
    VitalDistribution,
    VitalStat,
} from "../api";
import { Icon } from "./Icon";
import { EmptyState } from "./EmptyState";

interface Props {
    vitals: VitalStat[];
    distributions: VitalDistribution[];
    pages: PagePerformanceStat[];
    score: PerformanceScore | null;
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

function formatPageMetric(name: string, value: number | null): string {
    if (value === null) return "—";
    return displayValue(name, value);
}

function pageRating(page: PagePerformanceStat): Rating {
    const ratings = [
        ratingFor("LCP", page.lcp),
        ratingFor("INP", page.inp),
        ratingFor("CLS", page.cls),
    ];
    if (ratings.includes("poor")) return "poor";
    if (ratings.includes("warning")) return "warning";
    if (ratings.includes("good")) return "good";
    return "unknown";
}

export function VitalsPage({ vitals, distributions, pages, score, loading }: Props) {
    const ordered = ["LCP", "INP", "CLS"].map((name) => ({
        name,
        value: vitals.find((item) => item.name === name)?.value ?? null,
    }));
    const hasScore = (score?.sample_size ?? 0) > 0;
    const hasVitalData = hasScore || ordered.some((item) => item.value !== null) || pages.length > 0;

    if (!loading && !hasVitalData) {
        return (
            <div className="page-stack">
                <section className="page-heading">
                    <div>
                        <span className="eyebrow">Performance</span>
                        <h2>Real-user vitals</h2>
                        <p>Measure how your pages feel where it matters: on real devices.</p>
                    </div>
                </section>
                <EmptyState
                    eyebrow="No samples yet"
                    title="Performance needs a first signal."
                    description="Enable Web Vitals collection and Iris will turn field measurements into an honest view of speed, responsiveness, and stability."
                    code="autocapture: { pageviews: true, webvitals: true }"
                    steps={[
                        { title: "Enable", description: "Turn on Web Vitals in your client config." },
                        { title: "Collect", description: "Let real visits create field samples." },
                        { title: "Improve", description: "Prioritize the pages with the clearest impact." },
                    ]}
                />
            </div>
        );
    }

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
                        {["LCP", "INP", "CLS"].map((name) => {
                            const distribution = distributions.find((item) => item.name === name);
                            const total = distribution?.total || 1;
                            return (
                                <div className="experience-metric" key={name}>
                                    <div className="experience-bar-group">
                                        <span className="good" style={{ height: `${Math.max(3, ((distribution?.good ?? 0) / total) * 100)}%` }} />
                                        <span className="warning" style={{ height: `${Math.max(3, ((distribution?.needs_improvement ?? 0) / total) * 100)}%` }} />
                                        <span className="poor" style={{ height: `${Math.max(3, ((distribution?.poor ?? 0) / total) * 100)}%` }} />
                                    </div>
                                    <strong>{name}</strong>
                                </div>
                            );
                        })}
                    </div>
                </article>

                <article className="card score-card">
                    <div className="card-header"><div><h3>Overall Score</h3><p>Performance health index</p></div></div>
                    <div className="score-ring" style={{ "--score": score?.score ?? 0 } as React.CSSProperties}>
                        <strong>{score?.score ?? 0}</strong>
                    </div>
                    <p>{hasScore ? `Based on ${score?.sample_size.toLocaleString()} Web Vital samples.` : "Collect Web Vitals to calculate your score."}</p>
                </article>
            </section>

            <section className="card poor-pages-card">
                <div className="card-header">
                    <div><h3>Pages to Monitor</h3><p>Highest-traffic pages to include in performance reviews</p></div>
                    <button className="secondary-button"><Icon name="search" size={15} /> Filter</button>
                </div>
                <div className="performance-table">
                    <div className="performance-row head performance-row-vitals">
                        <span>Page path</span><span>LCP</span><span>INP</span><span>CLS</span><span>Traffic</span>
                    </div>
                    {pages.slice(0, 8).map((page) => {
                        const rating = pageRating(page);
                        return (
                        <div className={`performance-row performance-row-vitals ${rating}`} key={page.url}>
                            <span>{page.url.replace(/^https?:\/\/[^/]+/, "") || "/"}</span>
                            <span>{formatPageMetric("LCP", page.lcp)}</span>
                            <span>{formatPageMetric("INP", page.inp)}</span>
                            <span>{formatPageMetric("CLS", page.cls)}</span>
                            <span>{page.traffic.toLocaleString()}</span>
                        </div>
                        );
                    })}
                    {!loading && pages.length === 0 && <div className="table-empty">No page traffic recorded yet.</div>}
                </div>
            </section>
        </div>
    );
}
