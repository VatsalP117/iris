import { useMemo, useState } from "react";

import type { SiteStat, StatsResult } from "../api";
import { Icon } from "./Icon";

export type SiteSummary = StatsResult;

interface Props {
    sites: SiteStat[];
    summaries: Record<string, SiteSummary>;
    selectedSiteId: string;
    onOpenSite: (siteId: string) => void;
}

function compact(value: number | undefined): string {
    if (!value) return "0";
    if (value >= 1_000_000) return `${(value / 1_000_000).toFixed(1)}M`;
    if (value >= 1_000) return `${(value / 1_000).toFixed(1)}K`;
    return value.toLocaleString();
}

function SparkBars({ seed }: { seed: number }) {
    const bars = Array.from({ length: 12 }, (_, index) => 20 + ((seed * 17 + index * 23) % 64));
    return (
        <div className="spark-bars" aria-hidden="true">
            {bars.map((height, index) => <span key={index} style={{ height: `${height}%` }} />)}
        </div>
    );
}

export function SitesPage({ sites, summaries, selectedSiteId, onOpenSite }: Props) {
    const [query, setQuery] = useState("");
    const filteredSites = useMemo(() => {
        const normalized = query.trim().toLowerCase();
        if (!normalized) return sites;
        return sites.filter((site) =>
            [site.site_id, site.domain, ...site.domains].some((value) => value.toLowerCase().includes(normalized)),
        );
    }, [query, sites]);

    return (
        <div className="page-stack">
            <section className="page-heading">
                <div>
                    <span className="eyebrow">Properties</span>
                    <h2>Managed Domains</h2>
                    <p>You have {sites.length} active {sites.length === 1 ? "domain" : "domains"} recording analytics.</p>
                </div>
                <label className="search-field">
                    <Icon name="search" size={17} />
                    <span className="sr-only">Search sites</span>
                    <input value={query} onChange={(event) => setQuery(event.target.value)} placeholder="Search sites..." />
                </label>
            </section>

            <section className="site-grid" aria-label="Managed sites">
                {filteredSites.map((site, index) => {
                    const summary = summaries[site.site_id];
                    return (
                        <article className={`site-card ${site.site_id === selectedSiteId ? "selected" : ""}`} key={site.site_id}>
                            <div className="site-card-top">
                                <div className="site-identity">
                                    <div className="site-icon"><Icon name="globe" size={20} /></div>
                                    <div>
                                        <h3>{site.domain || site.site_id}</h3>
                                        <span className="status-label"><i />Recording</span>
                                    </div>
                                </div>
                                <button aria-label={`More options for ${site.domain}`} className="icon-button compact">
                                    <Icon name="more" size={18} />
                                </button>
                            </div>

                            <div className="site-card-stats">
                                <div>
                                    <span>Pageviews</span>
                                    <strong>{compact(summary?.pageviews)}</strong>
                                </div>
                                <div>
                                    <span>Unique visitors</span>
                                    <strong>{compact(summary?.unique_visitors)}</strong>
                                </div>
                            </div>

                            <SparkBars seed={index + 1} />

                            <div className="site-card-footer">
                                <span>{site.domains.length || 1} configured {(site.domains.length || 1) === 1 ? "domain" : "domains"}</span>
                                <button onClick={() => onOpenSite(site.site_id)}>
                                    Open analytics <Icon name="external" size={14} />
                                </button>
                            </div>
                        </article>
                    );
                })}

                <article className="site-card add-site-card">
                    <div className="add-site-icon">+</div>
                    <strong>Configure another domain</strong>
                    <span>Initialize data with self-hosting</span>
                </article>
            </section>

            <section className="aggregate-health card">
                <div className="card-header">
                    <div>
                        <h3>Aggregate Health across {sites.length} {sites.length === 1 ? "Site" : "Sites"}</h3>
                        <p>Core Web Vitals status from all active properties</p>
                    </div>
                </div>
                <div className="health-grid">
                    <div className="health-item good"><i /><span>LCP</span><strong>Good</strong></div>
                    <div className="health-item warning"><i /><span>INP</span><strong>Monitor</strong></div>
                    <div className="health-item good"><i /><span>CLS</span><strong>Good</strong></div>
                </div>
            </section>
        </div>
    );
}
