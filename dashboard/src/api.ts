// Centralised API client for the Iris analytics backend.
// All endpoints expect ?domain=&from=&to= query params.

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
}

function buildParams(domain: string, from: string, to: string) {
    const p = new URLSearchParams({ domain });
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
    stats: (domain: string, from: string, to: string) =>
        get<StatsResult>(`/api/stats?${buildParams(domain, from, to)}`),

    pages: (domain: string, from: string, to: string) =>
        get<PageStat[]>(`/api/pages?${buildParams(domain, from, to)}`),

    referrers: (domain: string, from: string, to: string) =>
        get<ReferrerStat[]>(`/api/referrers?${buildParams(domain, from, to)}`),

    vitals: (domain: string, from: string, to: string) =>
        get<VitalStat[]>(`/api/vitals?${buildParams(domain, from, to)}`),

    devices: (domain: string, from: string, to: string) =>
        get<DeviceStat[]>(`/api/devices?${buildParams(domain, from, to)}`),

    timeseries: (domain: string, from: string, to: string) =>
        get<{ date: string; pageviews: number }[]>(`/api/timeseries?${buildParams(domain, from, to)}`),

    sites: () =>
        get<SiteStat[]>(`/api/sites`),
};
