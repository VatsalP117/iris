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

const state = {
    events: [],
    bytes: 0,
    failPosts: 0,
    delayMS: 0,
};

const sdkBuild = await build({
    entryPoints: [sdkPath],
    bundle: true,
    format: "esm",
    platform: "browser",
    write: false,
});
const sdk = sdkBuild.outputFiles[0].contents;
const server = http.createServer(async (request, response) => {
    const requestURL = new URL(request.url ?? "/", "http://127.0.0.1");
    if (requestURL.pathname === "/sdk.js") {
        response.writeHead(200, {
            "Content-Type": "text/javascript",
            "Cache-Control": "no-store",
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
        if (state.delayMS > 0) {
            await wait(state.delayMS);
        }
        if (state.failPosts > 0) {
            state.failPosts -= 1;
            response.writeHead(503);
            response.end("injected failure");
            return;
        }
        try {
            const decoded = JSON.parse(body.toString("utf8"));
            if (Array.isArray(decoded)) {
                state.events.push(...decoded);
            } else {
                state.events.push(decoded);
            }
            response.writeHead(202);
            response.end();
        } catch {
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
        await page.goto(`${baseURL}/fixture?scenario=${encodeURIComponent(name)}`);
        await page.waitForFunction(() => window.__irisReady === true);
        await action({ context, page });
        await waitForRequests();
        results.push({
            name,
            passed: state.events.length === expectedEvents,
            expectedEvents,
            actualEvents: state.events.length,
            eventNames: state.events.map((event) => event.n),
            bytes: state.bytes,
            consoleErrors,
        });
    } catch (error) {
        results.push({
            name,
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
            passed,
            expectedEvents: 2,
            actualEvents: state.events.length,
            sameVisitor: firstEvent?.vid === secondEvent?.vid,
            distinctSessions: firstEvent?.sid !== secondEvent?.sid,
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
            passed: state.events.length === 1,
            expectedEvents: 1,
            actualEvents: state.events.length,
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
            passed: state.events.length === 1000,
            expectedEvents: 1000,
            actualEvents: state.events.length,
            enqueueDurationMS: durationMS,
            bytes: state.bytes,
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
    state.bytes = 0;
    state.failPosts = 0;
    state.delayMS = 0;
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

function browserMarkdown(report) {
    const lines = [
        "# Iris Browser Reliability Report",
        "",
        `**Verdict:** ${report.passed ? "PASS" : "FAIL"}`,
        "",
        `Browser: \`${report.browserExecutable}\``,
        "",
        "| Scenario | Result | Expected | Actual |",
        "|---|---|---:|---:|",
    ];
    for (const scenarioResult of report.scenarios) {
        lines.push(
            `| ${scenarioResult.name} | ${scenarioResult.passed ? "PASS" : "FAIL"} | ` +
                `${scenarioResult.expectedEvents ?? "—"} | ${scenarioResult.actualEvents ?? "—"} |`,
        );
    }
    lines.push("", "## Technical details", "", "```json");
    lines.push(JSON.stringify(report.scenarios, null, 2));
    lines.push("```", "");
    return lines.join("\n");
}
