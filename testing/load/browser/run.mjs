import http from "node:http";
import fs from "node:fs/promises";
import path from "node:path";
import process from "node:process";
import { fileURLToPath } from "node:url";

import { build } from "esbuild";
import { chromium } from "playwright";

const scriptDirectory = path.dirname(fileURLToPath(import.meta.url));
const repositoryRoot = path.resolve(scriptDirectory, "../../..");
const sdkPath = path.join(repositoryRoot, "web/src/index.ts");
const fixturesDirectory = path.join(scriptDirectory, "fixtures");
const outputDirectory = path.resolve(
    process.env.IRIS_BROWSER_OUTPUT ??
        path.join(repositoryRoot, "artifacts/reliability/browser"),
);
const allowFailures = process.argv.includes("--allow-failures");
const fixtureHTML = `<!doctype html>
<html>
    <head><meta charset="utf-8"><title>Iris browser fixture</title></head>
    <body>
        <button id="action">Synthetic action</button>
        <a id="away-link" href="/lifecycle-away">Leave fixture</a>
        <script type="module">
            import { Iris } from "/sdk.js";
            window.makeIris = (overrides = {}) => new Iris({
                host: location.origin,
                siteId: "browser-lab",
                autocapture: { pageviews: true },
                ...overrides,
            });
            window.__irisReady = true;
        </script>
    </body>
</html>`;
const delayedFixtureHTML = `<!doctype html>
<html>
    <head><meta charset="utf-8"><title>Iris delayed fixture</title></head>
    <body>
        <script>
            window.__loadIris = async () => {
                const { Iris } = await import("/sdk.js");
                window.makeIris = (overrides = {}) => new Iris({
                    host: location.origin,
                    siteId: "browser-lab",
                    autocapture: { pageviews: true },
                    ...overrides,
                });
                window.__irisReady = true;
            };
        </script>
    </body>
</html>`;
const bfcacheFixtureHTML = fixtureHTML.replace("/sdk.js", "/sdk-cache.js");
const frameworkHTML = (bundle) => `<!doctype html>
<html>
    <head><meta charset="utf-8"><title>Iris framework fixture</title></head>
    <body>
        <div id="root"></div>
        <script type="module" src="${bundle}"></script>
    </body>
</html>`;

const state = {
    events: [],
    eventIds: new Set(),
    bytes: 0,
    failPosts: 0,
    delayMS: 0,
    responsePlan: [],
    requests: [],
    corsEnabled: true,
};

const [sdkBuild, declarativeBuild, dataRouterBuild] = await Promise.all([
    build({
        entryPoints: [sdkPath],
        bundle: true,
        format: "esm",
        platform: "browser",
        write: false,
    }),
    buildFrameworkFixture("react-router-declarative.tsx"),
    buildFrameworkFixture("react-router-data.tsx"),
]);
const sdk = sdkBuild.outputFiles[0].contents;
const frameworkBundles = {
    "/fixtures/react-router-declarative.js":
        declarativeBuild.outputFiles[0].contents,
    "/fixtures/react-router-data.js": dataRouterBuild.outputFiles[0].contents,
};
const server = http.createServer(async (request, response) => {
    const requestURL = new URL(request.url ?? "/", "http://127.0.0.1");
    if (request.method === "OPTIONS") {
        if (!state.corsEnabled) {
            response.writeHead(403);
            response.end("injected CORS rejection");
            return;
        }
        writeCORSHeaders(response);
        response.writeHead(204);
        response.end();
        return;
    }
    if (requestURL.pathname === "/sdk.js") {
        response.writeHead(200, {
            "Content-Type": "text/javascript",
            "Cache-Control": "no-store",
        });
        response.end(sdk);
        return;
    }
    if (requestURL.pathname === "/sdk-cache.js") {
        response.writeHead(200, {
            "Content-Type": "text/javascript",
            "Cache-Control": "public, max-age=3600",
        });
        response.end(sdk);
        return;
    }
    if (requestURL.pathname === "/fixture") {
        response.writeHead(200, {
            "Content-Type": "text/html",
            "Cache-Control": "no-store",
        });
        response.end(fixtureHTML);
        return;
    }
    if (requestURL.pathname === "/fixture-bfcache") {
        response.writeHead(200, {
            "Content-Type": "text/html",
        });
        response.end(bfcacheFixtureHTML);
        return;
    }
    if (requestURL.pathname === "/fixture-delayed") {
        response.writeHead(200, {
            "Content-Type": "text/html",
            "Cache-Control": "no-store",
        });
        response.end(delayedFixtureHTML);
        return;
    }
    if (requestURL.pathname === "/lifecycle-away") {
        response.writeHead(200, {
            "Content-Type": "text/html",
            "Cache-Control": "public, max-age=60",
        });
        response.end(
            "<!doctype html><html><body><p>Away page</p></body></html>",
        );
        return;
    }
    if (requestURL.pathname === "/framework/react-declarative") {
        response.writeHead(200, {
            "Content-Type": "text/html",
            "Cache-Control": "no-store",
        });
        response.end(
            frameworkHTML("/fixtures/react-router-declarative.js"),
        );
        return;
    }
    if (requestURL.pathname === "/framework/react-data") {
        response.writeHead(200, {
            "Content-Type": "text/html",
            "Cache-Control": "no-store",
        });
        response.end(frameworkHTML("/fixtures/react-router-data.js"));
        return;
    }
    if (frameworkBundles[requestURL.pathname]) {
        response.writeHead(200, {
            "Content-Type": "text/javascript",
            "Cache-Control": "no-store",
        });
        response.end(frameworkBundles[requestURL.pathname]);
        return;
    }
    if (requestURL.pathname === "/favicon.ico") {
        response.writeHead(204);
        response.end();
        return;
    }
    if (
        request.method === "POST" &&
        (requestURL.pathname === "/api/event" ||
            requestURL.pathname === "/api/events")
    ) {
        const body = await readBody(request);
        state.bytes += body.length;
        const behavior = state.responsePlan.shift() ?? "accept";
        const requestRecord = {
            endpoint: requestURL.pathname,
            behavior,
            bytes: body.length,
        };
        state.requests.push(requestRecord);
        if (state.delayMS > 0) {
            await wait(state.delayMS);
        }
        if (state.failPosts > 0 || behavior === "503") {
            if (state.failPosts > 0) {
                state.failPosts -= 1;
            }
            requestRecord.status = 503;
            response.writeHead(503);
            response.end("injected failure");
            return;
        }
        if (behavior === "429") {
            requestRecord.status = 429;
            response.writeHead(429, { "Retry-After": "0" });
            response.end("injected rate limit");
            return;
        }
        if (behavior === "400") {
            requestRecord.status = 400;
            response.writeHead(400);
            response.end("injected permanent rejection");
            return;
        }
        if (behavior.startsWith("status-")) {
            const injectedStatus = Number(behavior.slice("status-".length));
            requestRecord.status = injectedStatus;
            response.writeHead(injectedStatus);
            response.end(`injected HTTP ${injectedStatus}`);
            return;
        }
        if (behavior === "hang-close") {
            await wait(500);
            request.socket.destroy();
            return;
        }
        try {
            const decoded = JSON.parse(body.toString("utf8"));
            const decodedEvents = Array.isArray(decoded) ? decoded : [decoded];
            requestRecord.eventCount = decodedEvents.length;
            for (const event of decodedEvents) {
                if (!event.id || !state.eventIds.has(event.id)) {
                    state.events.push(event);
                    if (event.id) {
                        state.eventIds.add(event.id);
                    }
                }
            }
            requestRecord.status = 202;
            if (behavior === "accept-close") {
                request.socket.destroy();
                return;
            }
            if (behavior === "slow-accept") {
                await wait(750);
            }
            if (state.corsEnabled) {
                writeCORSHeaders(response);
            }
            response.writeHead(202);
            response.end();
        } catch {
            requestRecord.status = 400;
            response.writeHead(400);
            response.end("invalid JSON");
        }
        return;
    }
    response.writeHead(404);
    response.end("not found");
});

await new Promise((resolve) => server.listen(0, "127.0.0.1", resolve));
const address = server.address();
const baseURL = `http://127.0.0.1:${address.port}`;

const executablePath = await findBrowserExecutable();
const browser = await chromium.launch({
    headless: true,
    ignoreDefaultArgs: ["--disable-back-forward-cache"],
    ...(executablePath ? { executablePath } : {}),
});

const results = [];

try {
    await scenario("initial-pageview", 1, async ({ page }) => {
        await startIris(page);
    });

    await scenario("start-is-idempotent", 1, async ({ page }) => {
        await page.evaluate(() => {
            window.__iris = window.makeIris();
            window.__iris.start();
            window.__iris.start();
        });
    });

    await scenario("push-state-navigation", 2, async ({ page }) => {
        await startIris(page);
        await page.evaluate(() => history.pushState({}, "", "/next"));
    });

    await scenario("same-url-push-state-deduplicated", 1, async ({ page }) => {
        await startIris(page);
        await page.evaluate(() => history.pushState({}, "", location.href));
    });

    await scenario("replace-state-navigation", 2, async ({ page }) => {
        await startIris(page);
        await page.evaluate(() => history.replaceState({}, "", "/replaced"));
    });

    await scenario("hash-navigation", 2, async ({ page }) => {
        await startIris(page);
        await page.evaluate(() => {
            location.hash = "section";
        });
    });

    await scenario("multiple-instances-do-not-double-count", 1, async ({ page }) => {
        await page.evaluate(() => {
            window.__iris = window.makeIris();
            window.__iris2 = window.makeIris();
            window.__iris.start();
            window.__iris2.start();
        });
    });

    await scenario("stop-removes-navigation-tracking", 1, async ({ page }) => {
        await startIris(page);
        await page.evaluate(() => {
            window.__iris.stop();
            history.pushState({}, "", "/after-stop");
        });
    });

    await scenario(
        "rejected-beacon-falls-back-to-fetch",
        1,
        async ({ page }) => {
            await startIris(page);
        },
        {
            initScript: () => {
                Object.defineProperty(Navigator.prototype, "sendBeacon", {
                    configurable: true,
                    value: () => false,
                });
            },
        },
    );

    await scenario("transient-server-failure-retries", 1, async ({ page }) => {
        state.failPosts = 1;
        await startIris(page);
        await wait(500);
    });

    await scenario("pagehide-flushes-batch", 1, async ({ page }) => {
        await page.evaluate(() => {
            window.__iris = window.makeIris({
                batching: {
                    maxSize: 10,
                    flushInterval: 60000,
                    flushOnLeave: true,
                },
            });
            window.__iris.start();
            window.dispatchEvent(new PageTransitionEvent("pagehide"));
        });
    });

    await scenario(
        "storage-unavailable-falls-back-to-memory",
        2,
        async ({ page }) => {
            await startIris(page);
            await page.evaluate(() => history.pushState({}, "", "/memory"));
        },
        {
            initScript: () => {
                Storage.prototype.getItem = () => {
                    throw new Error("injected storage failure");
                };
                Storage.prototype.setItem = () => {
                    throw new Error("injected storage failure");
                };
            },
        },
    );

    await multiTabScenario();
    await offlineScenario();
    await overheadScenario();
    await stateMachineScenario();
    await deliveryChaosScenarios();
    await frameworkScenarios();
    await lifecycleScenarios();
} finally {
    await browser.close();
    server.closeAllConnections();
    await new Promise((resolve) => server.close(resolve));
}

const report = {
    generatedAt: new Date().toISOString(),
    sdkPath,
    browserExecutable: executablePath ?? "playwright-managed",
    passed: results.every((result) => result.passed),
    categories: summarizeCategories(results),
    scenarios: results,
};

await fs.mkdir(outputDirectory, { recursive: true });
await fs.writeFile(
    path.join(outputDirectory, "browser-summary.json"),
    `${JSON.stringify(report, null, 2)}\n`,
);
await fs.writeFile(
    path.join(outputDirectory, "browser-report.md"),
    browserMarkdown(report),
);

console.log(
    `Iris browser oracle: ${report.passed ? "PASS" : "FAIL"} ` +
        `(${results.filter((result) => result.passed).length}/${results.length} scenarios)`,
);
console.log(`Report: ${path.join(outputDirectory, "browser-report.md")}`);

if (!report.passed && !allowFailures) {
    process.exitCode = 1;
}

async function scenario(name, expectedEvents, action, options = {}) {
    console.log(`Browser scenario: ${name}`);
    resetState();
    const context = await browser.newContext();
    if (options.initScript) {
        await context.addInitScript(options.initScript);
    }
    const page = await context.newPage();
    const consoleErrors = [];
    page.on("console", (message) => {
        if (message.type() === "error") {
            consoleErrors.push(message.text());
        }
    });
    page.on("pageerror", (error) => consoleErrors.push(error.message));
    try {
        const fixtureURL =
            options.fixtureURL ??
            `${baseURL}/fixture?scenario=${encodeURIComponent(name)}`;
        await page.goto(fixtureURL);
        await page.waitForFunction(() => window.__irisReady === true);
        await action({ context, page });
        await wait(options.waitMS ?? 250);
        const observation = {
            events: [...state.events],
            requests: [...state.requests],
            bytes: state.bytes,
        };
        const passed = options.validate
            ? options.validate(observation)
            : state.events.length === expectedEvents;
        results.push({
            name,
            category: options.category ?? "sdk-flow",
            passed,
            expectedEvents,
            actualEvents: state.events.length,
            requestAttempts: state.requests.length,
            eventNames: state.events.map((event) => event.n),
            bytes: state.bytes,
            consoleErrors,
        });
    } catch (error) {
        results.push({
            name,
            category: options.category ?? "sdk-flow",
            passed: false,
            expectedEvents,
            actualEvents: state.events.length,
            error: error instanceof Error ? error.message : String(error),
            consoleErrors,
        });
    } finally {
        await context.close();
    }
}

async function multiTabScenario() {
    console.log("Browser scenario: multi-tab-identity");
    resetState();
    const context = await browser.newContext();
    const first = await context.newPage();
    const second = await context.newPage();
    try {
        await Promise.all([
            first.goto(`${baseURL}/fixture?scenario=multi-tab-1`),
            second.goto(`${baseURL}/fixture?scenario=multi-tab-2`),
        ]);
        await Promise.all([
            first.waitForFunction(() => window.__irisReady === true),
            second.waitForFunction(() => window.__irisReady === true),
        ]);
        await Promise.all([startIris(first), startIris(second)]);
        await waitForRequests();
        const [firstEvent, secondEvent] = state.events;
        const passed =
            state.events.length === 2 &&
            firstEvent?.vid === secondEvent?.vid &&
            firstEvent?.sid !== secondEvent?.sid;
        results.push({
            name: "multi-tab-identity",
            category: "sdk-flow",
            passed,
            expectedEvents: 2,
            actualEvents: state.events.length,
            sameVisitor: firstEvent?.vid === secondEvent?.vid,
            distinctSessions: firstEvent?.sid !== secondEvent?.sid,
        });
    } catch (error) {
        results.push({
            name: "multi-tab-identity",
            category: "sdk-flow",
            passed: false,
            expectedEvents: 2,
            actualEvents: state.events.length,
            error: error instanceof Error ? error.message : String(error),
        });
    } finally {
        await context.close();
    }
}

async function offlineScenario() {
    console.log("Browser scenario: offline-event-retries-when-online");
    resetState();
    const context = await browser.newContext();
    const page = await context.newPage();
    try {
        await page.goto(`${baseURL}/fixture?scenario=offline-recovery`);
        await page.waitForFunction(() => window.__irisReady === true);
        await context.setOffline(true);
        await startIris(page);
        await wait(200);
        await context.setOffline(false);
        await wait(500);
        results.push({
            name: "offline-event-retries-when-online",
            category: "delivery-chaos",
            passed: state.events.length === 1,
            expectedEvents: 1,
            actualEvents: state.events.length,
        });
    } catch (error) {
        results.push({
            name: "offline-event-retries-when-online",
            category: "delivery-chaos",
            passed: false,
            expectedEvents: 1,
            actualEvents: state.events.length,
            error: error instanceof Error ? error.message : String(error),
        });
    } finally {
        await context.close();
    }
}

async function overheadScenario() {
    console.log("Browser scenario: sdk-main-thread-overhead-1000-events");
    resetState();
    const context = await browser.newContext();
    const page = await context.newPage();
    try {
        await page.goto(`${baseURL}/fixture?scenario=sdk-overhead`);
        await page.waitForFunction(() => window.__irisReady === true);
        const durationMS = await page.evaluate(() => {
            const iris = window.makeIris({
                autocapture: false,
                batching: {
                    maxSize: 10,
                    flushInterval: 60000,
                    flushOnLeave: false,
                },
            });
            const start = performance.now();
            for (let index = 0; index < 1000; index += 1) {
                iris.track("benchmark", { index });
            }
            const duration = performance.now() - start;
            iris.stop();
            return duration;
        });
        await wait(1000);
        results.push({
            name: "sdk-main-thread-overhead-1000-events",
            category: "sdk-flow",
            passed: state.events.length === 1000,
            expectedEvents: 1000,
            actualEvents: state.events.length,
            enqueueDurationMS: durationMS,
            bytes: state.bytes,
        });
    } catch (error) {
        results.push({
            name: "sdk-main-thread-overhead-1000-events",
            category: "sdk-flow",
            passed: false,
            expectedEvents: 1000,
            actualEvents: state.events.length,
            error: error instanceof Error ? error.message : String(error),
        });
    } finally {
        await context.close();
    }
}

async function stateMachineScenario() {
    const name = "generated-client-state-machine";
    console.log(`Browser scenario: ${name}`);
    const traces = generatedStateMachineTraces();
    const traceResults = [];
    let expectedEvents = 0;
    let actualEvents = 0;

    for (let traceIndex = 0; traceIndex < traces.length; traceIndex += 1) {
        resetState();
        const context = await browser.newContext();
        const page = await context.newPage();
        const trace = traces[traceIndex];
        let started = false;
        const expectedRecords = [];
        try {
            await page.goto(
                `${baseURL}/fixture?scenario=state-machine-${traceIndex}`,
            );
            await page.waitForFunction(() => window.__irisReady === true);
            await page.evaluate(() => {
                window.__iris = window.makeIris();
            });
            const modeledURL = new URL(page.url());

            for (let step = 0; step < trace.length; step += 1) {
                const action = trace[step];
                if (action === "start") {
                    if (!started) {
                        expectedRecords.push(
                            stateMachineRecord("$pageview", modeledURL),
                        );
                        started = true;
                    }
                } else if (action === "stop") {
                    started = false;
                } else if (action === "track") {
                    expectedRecords.push(
                        stateMachineRecord("state-machine", modeledURL),
                    );
                } else if (action === "push") {
                    modeledURL.pathname = `/state/${traceIndex}/push-${step}`;
                    modeledURL.search = "";
                    modeledURL.hash = "";
                    if (started) {
                        expectedRecords.push(
                            stateMachineRecord("$pageview", modeledURL),
                        );
                    }
                } else if (action === "replace") {
                    modeledURL.pathname = `/state/${traceIndex}/replace-${step}`;
                    modeledURL.search = "";
                    modeledURL.hash = "";
                    if (started) {
                        expectedRecords.push(
                            stateMachineRecord("$pageview", modeledURL),
                        );
                    }
                } else if (action === "hash") {
                    modeledURL.hash = `state-${traceIndex}-${step}`;
                    if (started) {
                        expectedRecords.push(
                            stateMachineRecord("$pageview", modeledURL),
                        );
                    }
                }
                await executeStateAction(page, action, traceIndex, step);
            }
            await waitForRequests();
            const actualRecords = state.events.map((event) =>
                stateMachineRecord(event.n, new URL(event.u)),
            );
            const passed =
                JSON.stringify(actualRecords) ===
                JSON.stringify(expectedRecords);
            traceResults.push({
                trace: traceIndex,
                actions: trace,
                expected: expectedRecords.length,
                actual: actualRecords.length,
                passed,
                expectedRecords,
                actualRecords,
            });
            expectedEvents += expectedRecords.length;
            actualEvents += actualRecords.length;
        } catch (error) {
            traceResults.push({
                trace: traceIndex,
                actions: trace,
                expected: expectedRecords.length,
                actual: state.events.length,
                passed: false,
                error: error instanceof Error ? error.message : String(error),
            });
            expectedEvents += expectedRecords.length;
            actualEvents += state.events.length;
        } finally {
            await context.close();
        }
    }

    results.push({
        name,
        category: "state-machine",
        passed: traceResults.every((trace) => trace.passed),
        expectedEvents,
        actualEvents,
        traces: traceResults.length,
        failedTraceCount: traceResults.filter((trace) => !trace.passed).length,
        failedTraces: traceResults.filter((trace) => !trace.passed).slice(0, 10),
        seed: 0x1a15cafe,
    });
}

function stateMachineRecord(name, eventURL) {
    return {
        name,
        location: `${eventURL.pathname}${eventURL.search}${eventURL.hash}`,
    };
}

async function executeStateAction(page, action, traceIndex, step) {
    await page.evaluate(
        ({ actionName, traceNumber, stepNumber }) => {
            switch (actionName) {
                case "start":
                    window.__iris.start();
                    break;
                case "stop":
                    window.__iris.stop();
                    break;
                case "track":
                    window.__iris.track("state-machine", {
                        trace: traceNumber,
                        step: stepNumber,
                    });
                    break;
                case "push":
                    history.pushState(
                        {},
                        "",
                        `/state/${traceNumber}/push-${stepNumber}`,
                    );
                    break;
                case "push-same":
                    history.pushState({}, "", location.href);
                    break;
                case "replace":
                    history.replaceState(
                        {},
                        "",
                        `/state/${traceNumber}/replace-${stepNumber}`,
                    );
                    break;
                case "hash":
                    location.hash = `state-${traceNumber}-${stepNumber}`;
                    break;
            }
        },
        {
            actionName: action,
            traceNumber: traceIndex,
            stepNumber: step,
        },
    );
    await wait(5);
}

function generatedStateMachineTraces() {
    const traces = [
        ["start", "start", "push-same", "replace", "stop", "stop"],
        ["start", "push", "stop", "push", "start", "hash", "track"],
        ["track", "start", "replace", "push-same", "stop", "track"],
        ["start", "hash", "hash", "stop", "start", "push"],
    ];
    const actions = [
        "start",
        "start",
        "track",
        "push",
        "push-same",
        "replace",
        "hash",
        "stop",
        "stop",
    ];
    let randomState = 0x1a15cafe;
    const random = () => {
        randomState = (1664525 * randomState + 1013904223) >>> 0;
        return randomState / 0x100000000;
    };
    for (let trace = 0; trace < 12; trace += 1) {
        const generated = [];
        for (let step = 0; step < 18; step += 1) {
            generated.push(actions[Math.floor(random() * actions.length)]);
        }
        traces.push(generated);
    }
    return traces;
}

async function deliveryChaosScenarios() {
    const noBeacon = () => {
        Object.defineProperty(Navigator.prototype, "sendBeacon", {
            configurable: true,
            value: undefined,
        });
    };

    await scenario(
        "rate-limit-429-retries",
        1,
        async ({ page }) => {
            state.responsePlan = ["429", "accept"];
            await startIris(page);
        },
        {
            category: "delivery-chaos",
            initScript: noBeacon,
            waitMS: 1000,
            validate: ({ events, requests }) =>
                events.length === 1 && requests.length >= 2,
        },
    );

    for (const status of [408, 502, 504]) {
        await scenario(
            `retryable-${status}-eventually-delivers`,
            1,
            async ({ page }) => {
                state.responsePlan = [`status-${status}`, "accept"];
                await startIris(page);
            },
            {
                category: "delivery-chaos",
                initScript: noBeacon,
                waitMS: 1000,
                validate: ({ events, requests }) =>
                    events.length === 1 && requests.length >= 2,
            },
        );
    }

    await scenario(
        "permanent-400-is-not-retried",
        0,
        async ({ page }) => {
            state.responsePlan = ["400", "accept"];
            await startIris(page);
        },
        {
            category: "delivery-chaos",
            initScript: noBeacon,
            waitMS: 750,
            validate: ({ events, requests }) =>
                events.length === 0 && requests.length === 1,
        },
    );

    await scenario(
        "accepted-response-lost-does-not-duplicate",
        1,
        async ({ page }) => {
            state.responsePlan = ["accept-close", "accept"];
            await startIris(page);
        },
        {
            category: "delivery-chaos",
            initScript: noBeacon,
            waitMS: 1000,
            validate: ({ events }) => events.length === 1,
        },
    );

    await scenario(
        "slow-response-does-not-double-send",
        1,
        async ({ page }) => {
            state.responsePlan = ["slow-accept", "accept"];
            await startIris(page);
        },
        {
            category: "delivery-chaos",
            initScript: noBeacon,
            waitMS: 1200,
            validate: ({ events, requests }) =>
                events.length === 1 && requests.length === 1,
        },
    );

    await scenario(
        "hung-connection-closes-then-retries",
        1,
        async ({ page }) => {
            state.responsePlan = ["hang-close", "accept"];
            await startIris(page);
        },
        {
            category: "delivery-chaos",
            initScript: noBeacon,
            waitMS: 1500,
            validate: ({ events, requests }) =>
                events.length === 1 && requests.length >= 2,
        },
    );

    await scenario(
        "cross-origin-preflight-allows-delivery",
        1,
        async ({ page }) => {
            await page.evaluate((host) => {
                window.__iris = window.makeIris({ host });
                window.__iris.start();
            }, baseURL);
        },
        {
            category: "delivery-chaos",
            initScript: noBeacon,
            fixtureURL: `http://localhost:${address.port}/fixture?scenario=cors-allowed`,
            waitMS: 750,
        },
    );

    await scenario(
        "cross-origin-preflight-rejection-blocks-delivery",
        0,
        async ({ page }) => {
            state.corsEnabled = false;
            await page.evaluate((host) => {
                window.__iris = window.makeIris({ host });
                window.__iris.start();
            }, baseURL);
        },
        {
            category: "delivery-chaos",
            initScript: noBeacon,
            fixtureURL: `http://localhost:${address.port}/fixture?scenario=cors-rejected`,
            waitMS: 750,
        },
    );
}

async function frameworkScenarios() {
    await scenario(
        "react-router-declarative-navigation",
        3,
        async ({ page }) => {
            await page.click("#pricing-link");
            await page.waitForURL("**/framework/react-declarative/pricing");
            await page.click("#same-link");
            await page.click("#docs-button");
            await page.waitForURL("**/framework/react-declarative/docs");
        },
        {
            category: "framework",
            fixtureURL: `${baseURL}/framework/react-declarative`,
            waitMS: 500,
        },
    );

    await scenario(
        "react-router-data-navigation-and-redirect",
        4,
        async ({ page }) => {
            await page.click("#account-link");
            await page.waitForURL("**/framework/react-data/account");
            await page.click("#replace-button");
            await page.waitForURL("**/framework/react-data/settings");
            await page.click("#redirect-link");
            await page.waitForURL("**/framework/react-data/account");
        },
        {
            category: "framework",
            fixtureURL: `${baseURL}/framework/react-data`,
            waitMS: 500,
        },
    );
}

async function lifecycleScenarios() {
    await actualBFCacheScenario();

    await scenario(
        "synthetic-pageshow-without-navigation-does-not-count",
        1,
        async ({ page }) => {
            await startIris(page);
            await page.evaluate(() => {
                window.dispatchEvent(
                    new PageTransitionEvent("pagehide", { persisted: true }),
                );
                window.dispatchEvent(
                    new PageTransitionEvent("pageshow", { persisted: true }),
                );
            });
        },
        { category: "lifecycle" },
    );

    await scenario(
        "chromium-freeze-resume-preserves-navigation-tracking",
        2,
        async ({ context, page }) => {
            await startIris(page);
            const session = await context.newCDPSession(page);
            await session.send("Page.setWebLifecycleState", {
                state: "frozen",
            });
            await wait(100);
            await session.send("Page.setWebLifecycleState", {
                state: "active",
            });
            await page.evaluate(() => {
                history.pushState({}, "", "/after-resume");
            });
            await session.detach();
        },
        { category: "lifecycle" },
    );

    await scenario(
        "hidden-visibility-flushes-batch-once",
        1,
        async ({ page }) => {
            await page.evaluate(() => {
                window.__iris = window.makeIris({
                    autocapture: false,
                    batching: {
                        maxSize: 10,
                        flushInterval: 60000,
                        flushOnLeave: true,
                    },
                });
                window.__iris.track("queued-before-hidden");
                Object.defineProperty(document, "visibilityState", {
                    configurable: true,
                    value: "hidden",
                });
                document.dispatchEvent(new Event("visibilitychange"));
            });
        },
        { category: "lifecycle" },
    );

    await duplicatedTabScenario();
    await earlyNavigationScenario();
    await abruptCloseScenario();
}

async function actualBFCacheScenario() {
    const name = "actual-bfcache-restore-counts-return-navigation";
    console.log(`Browser scenario: ${name}`);
    resetState();
    const context = await browser.newContext();
    const page = await context.newPage();
    try {
        await page.goto(`${baseURL}/fixture-bfcache?scenario=actual-bfcache`);
        await page.waitForFunction(() => window.__irisReady === true);
        await page.evaluate(() => {
            window.__bfcacheSignals = [];
            window.addEventListener("pagehide", (event) => {
                window.__bfcacheSignals.push({
                    type: "pagehide",
                    persisted: event.persisted,
                });
            });
            window.addEventListener("pageshow", (event) => {
                window.__bfcacheSignals.push({
                    type: "pageshow",
                    persisted: event.persisted,
                });
            });
        });
        await startIris(page);
        await page.click("#away-link");
        await page.waitForURL("**/lifecycle-away");
        await page.evaluate(() => history.back());
        await page.waitForURL("**/fixture-bfcache?scenario=actual-bfcache", {
            waitUntil: "commit",
        });
        await page.waitForFunction(() => window.__irisReady === true);
        await waitForRequests();
        const signals = await page.evaluate(() => window.__bfcacheSignals ?? []);
        const restoredFromBFCache = signals.some(
            (signal) => signal.type === "pageshow" && signal.persisted,
        );
        const navigationDiagnostics = await page.evaluate(() => {
            const navigation = performance.getEntriesByType("navigation")[0];
            return {
                type: navigation?.type,
                notRestoredReasons: navigation?.notRestoredReasons ?? null,
            };
        });
        results.push({
            name,
            category: "lifecycle",
            passed: restoredFromBFCache && state.events.length === 2,
            expectedEvents: 2,
            actualEvents: state.events.length,
            restoredFromBFCache,
            signals,
            navigationDiagnostics,
        });
    } catch (error) {
        results.push({
            name,
            category: "lifecycle",
            passed: false,
            expectedEvents: 2,
            actualEvents: state.events.length,
            error: error instanceof Error ? error.message : String(error),
        });
    } finally {
        await context.close();
    }
}

async function duplicatedTabScenario() {
    const name = "duplicated-tab-keeps-visitor-but-rotates-session";
    console.log(`Browser scenario: ${name}`);
    resetState();
    const context = await browser.newContext();
    const page = await context.newPage();
    try {
        await page.goto(`${baseURL}/fixture?scenario=duplicate-source`);
        await page.waitForFunction(() => window.__irisReady === true);
        await startIris(page);
        const popupPromise = context.waitForEvent("page");
        await page.evaluate(() => window.open(location.href, "_blank"));
        const duplicate = await popupPromise;
        await duplicate.waitForFunction(() => window.__irisReady === true);
        await startIris(duplicate);
        await waitForRequests();
        const [firstEvent, secondEvent] = state.events;
        results.push({
            name,
            category: "lifecycle",
            passed:
                state.events.length === 2 &&
                firstEvent?.vid === secondEvent?.vid &&
                firstEvent?.sid !== secondEvent?.sid,
            expectedEvents: 2,
            actualEvents: state.events.length,
            sameVisitor: firstEvent?.vid === secondEvent?.vid,
            distinctSessions: firstEvent?.sid !== secondEvent?.sid,
        });
    } catch (error) {
        results.push({
            name,
            category: "lifecycle",
            passed: false,
            expectedEvents: 2,
            actualEvents: state.events.length,
            error: error instanceof Error ? error.message : String(error),
        });
    } finally {
        await context.close();
    }
}

async function earlyNavigationScenario() {
    const name = "navigation-before-sdk-load-uses-current-url";
    console.log(`Browser scenario: ${name}`);
    resetState();
    const context = await browser.newContext();
    const page = await context.newPage();
    try {
        await page.goto(`${baseURL}/fixture-delayed?scenario=early-navigation`);
        await page.evaluate(() =>
            history.pushState({}, "", "/early-before-sdk"),
        );
        await page.evaluate(() => window.__loadIris());
        await page.waitForFunction(() => window.__irisReady === true);
        await startIris(page);
        await waitForRequests();
        results.push({
            name,
            category: "lifecycle",
            passed:
                state.events.length === 1 &&
                new URL(state.events[0]?.u).pathname === "/early-before-sdk",
            expectedEvents: 1,
            actualEvents: state.events.length,
            actualURL: state.events[0]?.u,
        });
    } catch (error) {
        results.push({
            name,
            category: "lifecycle",
            passed: false,
            expectedEvents: 1,
            actualEvents: state.events.length,
            error: error instanceof Error ? error.message : String(error),
        });
    } finally {
        await context.close();
    }
}

async function abruptCloseScenario() {
    const name = "abrupt-page-close-flushes-queued-batch";
    console.log(`Browser scenario: ${name}`);
    resetState();
    const context = await browser.newContext();
    const page = await context.newPage();
    try {
        await page.goto(`${baseURL}/fixture?scenario=abrupt-close`);
        await page.waitForFunction(() => window.__irisReady === true);
        await page.evaluate(() => {
            window.__iris = window.makeIris({
                autocapture: false,
                batching: {
                    maxSize: 10,
                    flushInterval: 60000,
                    flushOnLeave: true,
                },
            });
            window.__iris.track("before-abrupt-close");
        });
        await page.close({ runBeforeUnload: true });
        await wait(500);
        results.push({
            name,
            category: "lifecycle",
            passed: state.events.length === 1,
            expectedEvents: 1,
            actualEvents: state.events.length,
        });
    } catch (error) {
        results.push({
            name,
            category: "lifecycle",
            passed: false,
            expectedEvents: 1,
            actualEvents: state.events.length,
            error: error instanceof Error ? error.message : String(error),
        });
    } finally {
        await context.close();
    }
}

async function startIris(page) {
    await page.evaluate(() => {
        window.__iris = window.makeIris();
        window.__iris.start();
    });
}

function resetState() {
    state.events = [];
    state.eventIds = new Set();
    state.bytes = 0;
    state.failPosts = 0;
    state.delayMS = 0;
    state.responsePlan = [];
    state.requests = [];
    state.corsEnabled = true;
}

async function waitForRequests() {
    await wait(250);
}

function wait(milliseconds) {
    return new Promise((resolve) => setTimeout(resolve, milliseconds));
}

async function readBody(request) {
    const chunks = [];
    for await (const chunk of request) {
        chunks.push(chunk);
    }
    return Buffer.concat(chunks);
}

async function findBrowserExecutable() {
    if (process.env.IRIS_BROWSER_EXECUTABLE) {
        return process.env.IRIS_BROWSER_EXECUTABLE;
    }
    const candidates = [
        "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
        "/Applications/Chromium.app/Contents/MacOS/Chromium",
        "/usr/bin/google-chrome",
        "/usr/bin/chromium",
        "/usr/bin/chromium-browser",
    ];
    for (const candidate of candidates) {
        try {
            await fs.access(candidate);
            return candidate;
        } catch {
            // Try the next browser.
        }
    }
    return null;
}

function buildFrameworkFixture(filename) {
    return build({
        entryPoints: [path.join(fixturesDirectory, filename)],
        bundle: true,
        format: "esm",
        platform: "browser",
        jsx: "automatic",
        nodePaths: [path.join(repositoryRoot, "marketing/node_modules")],
        write: false,
    });
}

function writeCORSHeaders(response) {
    response.setHeader("Access-Control-Allow-Origin", "*");
    response.setHeader("Access-Control-Allow-Methods", "POST, OPTIONS");
    response.setHeader("Access-Control-Allow-Headers", "Content-Type");
}

function summarizeCategories(scenarioResults) {
    const categories = {};
    for (const result of scenarioResults) {
        const category = result.category ?? "sdk-flow";
        categories[category] ??= { passed: 0, failed: 0, total: 0 };
        categories[category].total += 1;
        if (result.passed) {
            categories[category].passed += 1;
        } else {
            categories[category].failed += 1;
        }
    }
    return categories;
}

function browserMarkdown(report) {
    const lines = [
        "# Iris Browser Reliability Report",
        "",
        `**Verdict:** ${report.passed ? "PASS" : "FAIL"}`,
        "",
        `Browser: \`${report.browserExecutable}\``,
        "",
        "## Category summary",
        "",
        "| Category | Passed | Failed | Total |",
        "|---|---:|---:|---:|",
    ];
    for (const [category, summary] of Object.entries(report.categories)) {
        lines.push(
            `| ${category} | ${summary.passed} | ${summary.failed} | ${summary.total} |`,
        );
    }
    lines.push(
        "",
        "## Scenarios",
        "",
        "| Category | Scenario | Result | Expected | Actual |",
        "|---|---|---|---:|---:|",
    );
    for (const scenarioResult of report.scenarios) {
        lines.push(
            `| ${scenarioResult.category ?? "sdk-flow"} | ${scenarioResult.name} | ` +
                `${scenarioResult.passed ? "PASS" : "FAIL"} | ` +
                `${scenarioResult.expectedEvents ?? "—"} | ${scenarioResult.actualEvents ?? "—"} |`,
        );
    }
    lines.push("", "## Technical details", "", "```json");
    lines.push(JSON.stringify(report.scenarios, null, 2));
    lines.push("```", "");
    return lines.join("\n");
}
