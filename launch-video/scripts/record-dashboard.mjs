import { chromium } from "playwright";
import { mkdir, unlink } from "node:fs/promises";
import { fileURLToPath } from "node:url";
import path from "node:path";

import { responseFor } from "./capture-dashboard.mjs";

const scriptDirectory = path.dirname(fileURLToPath(import.meta.url));
const clipDirectory = path.resolve(scriptDirectory, "../public/clips");
const dashboardUrl = process.env.IRIS_DASHBOARD_URL ?? "http://127.0.0.1:5174/";

await mkdir(clipDirectory, { recursive: true });

const browser = await chromium.launch({
    headless: true,
    executablePath: "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
});

async function makePage(name) {
    const context = await browser.newContext({
        viewport: { width: 1920, height: 1080 },
        recordVideo: { dir: clipDirectory, size: { width: 1920, height: 1080 } },
    });
    const page = await context.newPage();
    await page.addInitScript(() => {
        window.addEventListener("DOMContentLoaded", () => {
            const cursor = document.createElement("div");
            cursor.id = "iris-recording-cursor";
            Object.assign(cursor.style, {
                background: "#6558ef",
                border: "3px solid white",
                borderRadius: "50%",
                boxShadow: "0 4px 18px rgba(21,21,18,.28)",
                height: "22px",
                left: "-30px",
                pointerEvents: "none",
                position: "fixed",
                top: "-30px",
                width: "22px",
                zIndex: "999999",
            });
            document.body.appendChild(cursor);
            document.addEventListener("mousemove", (event) => {
                cursor.style.left = `${event.clientX - 11}px`;
                cursor.style.top = `${event.clientY - 11}px`;
            });
            document.addEventListener("mousedown", () => {
                cursor.style.scale = "0.68";
            });
            document.addEventListener("mouseup", () => {
                cursor.style.scale = "1";
            });
        });
    });
    await page.route("**/api/**", async (route) => {
        const body = responseFor(route.request().url());
        if (body === null) return route.continue();
        return route.fulfill({ contentType: "application/json", body: JSON.stringify(body) });
    });
    await page.goto(dashboardUrl, { waitUntil: "networkidle" });
    await page.waitForTimeout(500);
    return { context, page, name, video: page.video() };
}

async function finish({ context, page, name, video }) {
    await page.waitForTimeout(350);
    await page.close();
    const temporaryPath = await video.path();
    await video.saveAs(path.join(clipDirectory, `${name}.webm`));
    await unlink(temporaryPath);
    await context.close();
}

const overview = await makePage("overview-interaction");
await overview.page.mouse.move(1575, 46, { steps: 18 });
await overview.page.locator(".period-select select").click();
await overview.page.locator(".period-select select").selectOption("7d");
await overview.page.waitForTimeout(650);
await overview.page.mouse.move(1180, 520, { steps: 28 });
await overview.page.waitForTimeout(850);
await overview.page.getByRole("button", { name: "Sessions" }).click();
await overview.page.waitForTimeout(650);
await overview.page.mouse.move(860, 610, { steps: 22 });
await overview.page.waitForTimeout(750);
await finish(overview);

const events = await makePage("events-interaction");
await events.page.mouse.move(108, 245, { steps: 22 });
await events.page.getByRole("button", { name: /Custom Events/ }).click();
await events.page.getByText("Event intelligence").waitFor();
await events.page.waitForTimeout(550);
await events.page.mouse.move(930, 650, { steps: 28 });
await events.page.waitForTimeout(650);
await events.page.getByRole("button", { name: /docs_searched/ }).click();
await events.page.waitForTimeout(650);
const search = events.page.getByPlaceholder("Quick search...");
await search.click();
await search.fill("signup");
await events.page.waitForTimeout(850);
await finish(events);

const vitals = await makePage("vitals-interaction");
await vitals.page.mouse.move(105, 295, { steps: 22 });
await vitals.page.getByRole("button", { name: /Web Vitals/ }).click();
await vitals.page.getByText("User Experience Distribution").waitFor();
await vitals.page.waitForTimeout(500);
await vitals.page.mouse.move(1520, 510, { steps: 30 });
await vitals.page.waitForTimeout(700);
await vitals.page.mouse.move(480, 790, { steps: 26 });
await vitals.page.waitForTimeout(600);
await finish(vitals);

await browser.close();
console.log(`Recorded Iris product interactions in ${clipDirectory}`);
