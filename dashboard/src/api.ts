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

async function get<T>(path: string): Promise<T> {
    const res = await fetch(BASE + path);
    if (!res.ok) throw new Error(`${path} → ${res.status}`);
    return res.json();
}

export const api = {
    stats: (siteId: string, from: string, to: string) =>
        get<StatsResult>(`/api/stats?${buildParams(siteId, from, to)}`),

    pages: (siteId: string, from: string, to: string) =>
        get<PageStat[]>(`/api/pages?${buildParams(siteId, from, to)}`),

    referrers: (siteId: string, from: string, to: string) =>
        get<ReferrerStat[]>(`/api/referrers?${buildParams(siteId, from, to)}`),

    vitals: (siteId: string, from: string, to: string) =>
        get<VitalStat[]>(`/api/vitals?${buildParams(siteId, from, to)}`),

    devices: (siteId: string, from: string, to: string) =>
        get<DeviceStat[]>(`/api/devices?${buildParams(siteId, from, to)}`),

    timeseries: (siteId: string, from: string, to: string) =>
        get<{ date: string; pageviews: number }[]>(`/api/timeseries?${buildParams(siteId, from, to)}`),

    sites: () =>
        get<SiteStat[]>(`/api/sites`),
};
