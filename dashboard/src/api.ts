// Centralized API client for the Iris analytics backend.
// All stats endpoints expect ?site_id=&from=&to= query params.

const BASE = "";

export interface StatsResult {
    pageviews: number;
    unique_visitors: number;
    sessions: number;
}

export interface StatsChange {
    pageviews: number;
    unique_visitors: number;
    sessions: number;
}

export interface SiteTrendResult {
    current: StatsResult;
    previous: StatsResult;
    change: StatsChange;
}

export interface PageStat {
    url: string;
    pageviews: number;
}

export interface ReferrerStat {
    referrer: string;
    visitors: number;
}

export interface VitalStat {
    name: string;
    value: number;
}

export interface VitalDistribution {
    name: string;
    total: number;
    good: number;
    needs_improvement: number;
    poor: number;
}

export interface PagePerformanceStat {
    url: string;
    lcp: number | null;
    inp: number | null;
    cls: number | null;
    traffic: number;
}

export interface PerformanceScore {
    score: number;
    rating: "good" | "needs-improvement" | "poor" | "unknown";
    metric_scores: Record<string, number>;
    sample_size: number;
}

export interface CustomEventSummary {
    total_events: number;
    unique_users: number;
    conversion_rate: number;
    change_percent: number;
}

export interface CustomEventStat {
    event_name: string;
    total_count: number;
    unique_users: number;
    change_percent: number;
}

export interface CustomEventsResult {
    summary: CustomEventSummary;
    events: CustomEventStat[];
}

export interface CustomEventTimeSeriesBucket {
    date: string;
    count: number;
}

export interface DeviceStat {
    device: string;
    count: number;
}

export interface SiteStat {
    site_id: string;
    domain: string;
    domains: string[];
}

function buildParams(siteId: string, from: string, to: string) {
    const p = new URLSearchParams({ site_id: siteId });
    if (from) p.set("from", from);
    if (to) p.set("to", to);
    return p.toString();
}

async function get<T>(path: string, signal?: AbortSignal): Promise<T> {
    const res = await fetch(BASE + path, { signal });
    if (!res.ok) throw new Error(`${path} → ${res.status}`);
    return res.json();
}

export const api = {
    stats: (siteId: string, from: string, to: string, signal?: AbortSignal) =>
        get<StatsResult>(`/api/stats?${buildParams(siteId, from, to)}`, signal),

    siteTrends: (siteId: string, from: string, to: string, signal?: AbortSignal) =>
        get<SiteTrendResult>(`/api/site-trends?${buildParams(siteId, from, to)}`, signal),

    pages: (siteId: string, from: string, to: string, signal?: AbortSignal) =>
        get<PageStat[]>(`/api/pages?${buildParams(siteId, from, to)}`, signal),

    referrers: (siteId: string, from: string, to: string, signal?: AbortSignal) =>
        get<ReferrerStat[]>(`/api/referrers?${buildParams(siteId, from, to)}`, signal),

    vitals: (siteId: string, from: string, to: string, signal?: AbortSignal) =>
        get<VitalStat[]>(`/api/vitals?${buildParams(siteId, from, to)}`, signal),

    vitalDistributions: (siteId: string, from: string, to: string, signal?: AbortSignal) =>
        get<VitalDistribution[]>(`/api/vitals/distribution?${buildParams(siteId, from, to)}`, signal),

    pagePerformance: (siteId: string, from: string, to: string, signal?: AbortSignal) =>
        get<PagePerformanceStat[]>(`/api/vitals/pages?${buildParams(siteId, from, to)}`, signal),

    performanceScore: (siteId: string, from: string, to: string, signal?: AbortSignal) =>
        get<PerformanceScore>(`/api/vitals/score?${buildParams(siteId, from, to)}`, signal),

    customEvents: (siteId: string, from: string, to: string, signal?: AbortSignal) =>
        get<CustomEventsResult>(`/api/custom-events?${buildParams(siteId, from, to)}`, signal),

    customEventTimeseries: (siteId: string, eventName: string, from: string, to: string, signal?: AbortSignal) => {
        const params = new URLSearchParams(buildParams(siteId, from, to));
        params.set("event_name", eventName);
        return get<CustomEventTimeSeriesBucket[]>(`/api/custom-events/timeseries?${params}`, signal);
    },

    devices: (siteId: string, from: string, to: string, signal?: AbortSignal) =>
        get<DeviceStat[]>(`/api/devices?${buildParams(siteId, from, to)}`, signal),

    timeseries: (siteId: string, from: string, to: string, signal?: AbortSignal) =>
        get<{ date: string; pageviews: number }[]>(`/api/timeseries?${buildParams(siteId, from, to)}`, signal),

    uniqueVisitorsTimeseries: (siteId: string, from: string, to: string, signal?: AbortSignal) =>
        get<{ date: string; uniqueVisitors: number }[]>(`/api/timeseries/visitors?${buildParams(siteId, from, to)}`, signal),

    sessionsTimeseries: (siteId: string, from: string, to: string, signal?: AbortSignal) =>
        get<{ date: string; sessions: number }[]>(`/api/timeseries/sessions?${buildParams(siteId, from, to)}`, signal),

    sites: () =>
        get<SiteStat[]>(`/api/sites`),
};
