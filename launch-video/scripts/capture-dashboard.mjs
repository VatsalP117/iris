import { chromium } from "playwright";
import { mkdir } from "node:fs/promises";
import { fileURLToPath } from "node:url";
import path from "node:path";

const scriptDirectory = path.dirname(fileURLToPath(import.meta.url));
const captureDirectory = path.resolve(scriptDirectory, "../public/captures");
const dashboardUrl = process.env.IRIS_DASHBOARD_URL ?? "http://127.0.0.1:5174/";

await mkdir(captureDirectory, { recursive: true });

export const days = Array.from({ length: 30 }, (_, index) => {
    const date = new Date(Date.UTC(2026, 6, 10 + index));
    const wave = Math.sin(index * 0.72) * 180;
    const pageviews = Math.round(620 + index * 26 + wave + (index % 5) * 44);
    const uniqueVisitors = Math.round(pageviews * (0.61 + (index % 3) * 0.02));
    const sessions = Math.round(uniqueVisitors * 1.16);
    return {
        date: date.toISOString().slice(0, 10),
        pageviews,
        uniqueVisitors,
        sessions,
    };
});

export const eventSeries = days.map((day, index) => ({
    date: day.date,
    count: Math.round(32 + index * 1.7 + Math.sin(index * 0.85) * 12),
}));

export const fixtures = {
    sites: [{ site_id: "iris-launch", domain: "iris.sh", domains: ["iris.sh", "docs.iris.sh"] }],
    trend: {
        current: { pageviews: 21642, unique_visitors: 8492, sessions: 9106 },
        previous: { pageviews: 18278, unique_visitors: 7636, sessions: 8248 },
        change: { pageviews: 18.4, unique_visitors: 11.2, sessions: 10.4 },
    },
    pages: [
        { url: "/", pageviews: 6842 },
        { url: "/docs", pageviews: 4914 },
        { url: "/pricing", pageviews: 3821 },
        { url: "/blog/launch", pageviews: 2790 },
        { url: "/docs/self-hosting", pageviews: 1964 },
        { url: "/about", pageviews: 1311 },
    ],
    referrers: [
        { referrer: "Direct", visitors: 3567 },
        { referrer: "google.com", visitors: 2418 },
        { referrer: "github.com", visitors: 1543 },
        { referrer: "news.ycombinator.com", visitors: 723 },
        { referrer: "x.com", visitors: 241 },
    ],
    devices: [
        { device: "Desktop", count: 12120 },
        { device: "Mobile", count: 6974 },
        { device: "Tablet", count: 2548 },
    ],
    vitals: [
        { name: "LCP", value: 1830 },
        { name: "INP", value: 142 },
        { name: "CLS", value: 0.06 },
    ],
    distributions: [
        { name: "LCP", total: 5200, good: 4310, needs_improvement: 703, poor: 187 },
        { name: "INP", total: 4180, good: 3688, needs_improvement: 391, poor: 101 },
        { name: "CLS", total: 5370, good: 4862, needs_improvement: 421, poor: 87 },
    ],
    performancePages: [
        { url: "https://iris.sh/", lcp: 1610, inp: 118, cls: 0.04, traffic: 6842 },
        { url: "https://iris.sh/docs", lcp: 1780, inp: 132, cls: 0.05, traffic: 4914 },
        { url: "https://iris.sh/pricing", lcp: 2120, inp: 151, cls: 0.08, traffic: 3821 },
        { url: "https://iris.sh/blog/launch", lcp: 2860, inp: 218, cls: 0.12, traffic: 2790 },
    ],
    score: { score: 94, rating: "good", metric_scores: { LCP: 92, INP: 96, CLS: 95 }, sample_size: 14750 },
    events: {
        summary: { total_events: 3684, unique_users: 2178, conversion_rate: 18.7, change_percent: 24.8 },
        events: [
            { event_name: "signup_completed", total_count: 1248, unique_users: 1083, change_percent: 31.4 },
            { event_name: "deploy_clicked", total_count: 916, unique_users: 742, change_percent: 22.1 },
            { event_name: "docs_searched", total_count: 684, unique_users: 511, change_percent: 14.8 },
            { event_name: "github_opened", total_count: 502, unique_users: 476, change_percent: 8.6 },
            { event_name: "copy_snippet", total_count: 334, unique_users: 288, change_percent: 19.2 },
        ],
    },
};

export function responseFor(url) {
    const { pathname } = new URL(url);
    if (pathname === "/api/sites") return fixtures.sites;
    if (pathname === "/api/site-trends") return fixtures.trend;
    if (pathname === "/api/pages") return fixtures.pages;
    if (pathname === "/api/referrers") return fixtures.referrers;
    if (pathname === "/api/devices") return fixtures.devices;
    if (pathname === "/api/vitals") return fixtures.vitals;
    if (pathname === "/api/vitals/distribution") return fixtures.distributions;
    if (pathname === "/api/vitals/pages") return fixtures.performancePages;
    if (pathname === "/api/vitals/score") return fixtures.score;
    if (pathname === "/api/custom-events/timeseries") return eventSeries;
    if (pathname === "/api/custom-events") return fixtures.events;
    if (pathname === "/api/timeseries") return days.map(({ date, pageviews }) => ({ date, pageviews }));
    if (pathname === "/api/timeseries/visitors") {
        return days.map(({ date, uniqueVisitors }) => ({ date, uniqueVisitors }));
    }
    if (pathname === "/api/timeseries/sessions") {
        return days.map(({ date, sessions }) => ({ date, sessions }));
    }
    return null;
}

const browser = await chromium.launch({
    headless: true,
    executablePath: "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
});

const page = await browser.newPage({
    viewport: { width: 1920, height: 1080 },
    deviceScaleFactor: 1,
});

await page.route("**/api/**", async (route) => {
    const body = responseFor(route.request().url());
    if (body === null) return route.continue();
    return route.fulfill({
        contentType: "application/json",
        body: JSON.stringify(body),
    });
});

await page.goto(dashboardUrl, { waitUntil: "networkidle" });
await page.screenshot({ path: path.join(captureDirectory, "overview.png") });

await page.getByRole("button", { name: /Custom Events/ }).click();
await page.getByText("Event intelligence").waitFor();
await page.waitForTimeout(400);
await page.screenshot({ path: path.join(captureDirectory, "events.png") });

await page.getByRole("button", { name: /Web Vitals/ }).click();
await page.getByText("User Experience Distribution").waitFor();
await page.waitForTimeout(400);
await page.screenshot({ path: path.join(captureDirectory, "vitals.png") });

console.log(`Captured Iris product screens in ${captureDirectory}`);
await browser.close();
