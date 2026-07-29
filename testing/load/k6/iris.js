import { check } from "k6";
import exec from "k6/execution";
import http from "k6/http";
import { Counter, Rate, Trend } from "k6/metrics";

const target = (__ENV.TARGET || "http://127.0.0.1:8080").replace(/\/$/, "");
const profile = __ENV.PROFILE || "constant";
const eventRate = numberEnv("EVENT_RATE", 100);
const duration = __ENV.DURATION || "30s";
const batchSize = numberEnv("BATCH_SIZE", 1);
const readRate = numberEnv("READ_RATE", 0);
const runID = __ENV.RUN_ID || `k6-${Date.now()}`;
const siteID = __ENV.SITE_ID || `iris-lab-${runID}`;

const acceptedEvents = new Counter("iris_accepted_events");
const rejectedEvents = new Counter("iris_rejected_events");
const ingestionSuccess = new Rate("iris_ingestion_success");
const ingestionLatency = new Trend("iris_ingestion_latency", true);
const analyticsReadSuccess = new Rate("iris_read_success");

export const options = {
    discardResponseBodies: true,
    scenarios: {
        ingestion: ingestionScenario(),
        ...(readRate > 0
            ? {
                  analytics_reads: {
                      executor: "constant-arrival-rate",
                      exec: "analyticsRead",
                      rate: readRate,
                      timeUnit: "1s",
                      duration,
                      preAllocatedVUs: Math.max(4, Math.ceil(readRate / 4)),
                      maxVUs: Math.max(16, readRate * 2),
                  },
              }
            : {}),
    },
    thresholds: {
        iris_ingestion_success: ["rate>0.99"],
        iris_ingestion_latency: ["p(95)<1000"],
        http_req_failed: ["rate<0.01"],
        ...(readRate > 0 ? { iris_read_success: ["rate>0.99"] } : {}),
    },
};

export default function () {
    const iteration = exec.scenario.iterationInTest;
    const events = [];
    for (let index = 0; index < batchSize; index += 1) {
        events.push(buildEvent(iteration * batchSize + index));
    }

    const endpoint = batchSize === 1 ? "/api/event" : "/api/events";
    const payload = batchSize === 1 ? events[0] : events;
    const response = http.post(`${target}${endpoint}`, JSON.stringify(payload), {
        headers: { "Content-Type": "application/json" },
        tags: { operation: "ingest", profile },
    });
    const accepted = response.status === 202;
    ingestionSuccess.add(accepted);
    ingestionLatency.add(response.timings.duration);
    if (accepted) {
        acceptedEvents.add(events.length);
    } else {
        rejectedEvents.add(events.length);
    }
    check(response, { "ingestion accepted": (value) => value.status === 202 });
}

export function analyticsRead() {
    const endpoints = [
        "/api/stats",
        "/api/pages",
        "/api/referrers",
        "/api/vitals",
        "/api/devices",
        "/api/timeseries",
        "/api/timeseries/visitors",
        "/api/timeseries/sessions",
    ];
    const endpoint =
        endpoints[exec.scenario.iterationInTest % endpoints.length];
    const response = http.get(
        `${target}${endpoint}?site_id=${encodeURIComponent(siteID)}`,
        { tags: { operation: "analytics-read", endpoint } },
    );
    analyticsReadSuccess.add(response.status === 200);
}

export function handleSummary(data) {
    const output = __ENV.K6_SUMMARY || "artifacts/reliability/k6-summary.json";
    return {
        stdout: summaryText(data),
        [output]: JSON.stringify(data, null, 2),
    };
}

function ingestionScenario() {
    const requestRate = Math.max(1, Math.ceil(eventRate / batchSize));
    const common = {
        exec: "default",
        timeUnit: "1s",
        preAllocatedVUs: Math.max(8, Math.ceil(requestRate / 4)),
        maxVUs: Math.max(64, requestRate * 2),
    };

    if (profile === "ramp") {
        return {
            ...common,
            executor: "ramping-arrival-rate",
            startRate: Math.max(1, Math.ceil(100 / batchSize)),
            stages: [
                { target: Math.ceil(500 / batchSize), duration: "2m" },
                { target: Math.ceil(1000 / batchSize), duration: "2m" },
                { target: Math.ceil(2000 / batchSize), duration: "2m" },
                { target: Math.ceil(100 / batchSize), duration: "2m" },
            ],
        };
    }
    if (profile === "spike") {
        return {
            ...common,
            executor: "ramping-arrival-rate",
            startRate: Math.max(1, Math.ceil(100 / batchSize)),
            stages: [
                { target: Math.ceil(100 / batchSize), duration: "10s" },
                { target: Math.ceil(2000 / batchSize), duration: "1s" },
                { target: Math.ceil(2000 / batchSize), duration: "30s" },
                { target: Math.ceil(100 / batchSize), duration: "1s" },
                { target: Math.ceil(100 / batchSize), duration: "18s" },
            ],
        };
    }

    return {
        ...common,
        executor: "constant-arrival-rate",
        rate: requestRate,
        duration,
    };
}

function buildEvent(sequence) {
    const pages = ["/", "/pricing", "/docs", "/blog/reliability"];
    const widths = [390, 820, 1440];
    const referrers = [
        "",
        "https://www.google.com/search?q=iris",
        "https://news.ycombinator.com/item?id=1",
    ];
    const kind = sequence % 10;
    let name = "$pageview";
    let flow = "pageview";
    if (kind >= 6 && kind < 8) {
        name = "$click";
        flow = "click";
    } else if (kind === 8) {
        name = "signup";
        flow = "custom";
    } else if (kind === 9) {
        name = "$web_vital";
        flow = "web-vital";
    }

    const properties = {
        $test_run: runID,
        $test_seq: sequence,
        $test_flow: flow,
    };
    if (name === "$web_vital") {
        const metric = Math.floor(sequence / 10) % 3;
        properties.$name = ["LCP", "INP", "CLS"][metric];
        properties.$val = [1800, 180, 0.08][metric];
        properties.$rating = "good";
    }

    return {
        n: name,
        u: `https://lab.example${pages[sequence % pages.length]}`,
        d: "lab.example",
        r: referrers[sequence % referrers.length],
        w: widths[sequence % widths.length],
        s: siteID,
        sid: `session-${String(Math.floor(sequence / 4)).padStart(6, "0")}`,
        vid: `visitor-${String(Math.floor(sequence / 7)).padStart(6, "0")}`,
        p: properties,
    };
}

function numberEnv(name, fallback) {
    const value = Number(__ENV[name] || fallback);
    return Number.isFinite(value) && value > 0 ? value : fallback;
}

function summaryText(data) {
    const metrics = data.metrics;
    const lines = [
        "",
        "Iris k6 summary",
        `  accepted events: ${metrics.iris_accepted_events?.values?.count ?? 0}`,
        `  rejected events: ${metrics.iris_rejected_events?.values?.count ?? 0}`,
        `  HTTP failures: ${(
            (metrics.http_req_failed?.values?.rate ?? 0) * 100
        ).toFixed(2)}%`,
        `  write p95: ${(
            metrics.iris_ingestion_latency?.values?.["p(95)"] ?? 0
        ).toFixed(2)} ms`,
        "",
    ];
    return lines.join("\n");
}
