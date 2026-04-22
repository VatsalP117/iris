// Centralized API client for the Iris analytics backend.
// All stats endpoints expect ?site_id=&from=&to= query params.

const BASE = "";

export interface StatsResult {
    pageviews: number;
    unique_visitors: number;
    sessions: number;
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

    pages: (siteId: string, from: string, to: string, signal?: AbortSignal) =>
        get<PageStat[]>(`/api/pages?${buildParams(siteId, from, to)}`, signal),

    referrers: (siteId: string, from: string, to: string, signal?: AbortSignal) =>
        get<ReferrerStat[]>(`/api/referrers?${buildParams(siteId, from, to)}`, signal),

    vitals: (siteId: string, from: string, to: string, signal?: AbortSignal) =>
        get<VitalStat[]>(`/api/vitals?${buildParams(siteId, from, to)}`, signal),

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
