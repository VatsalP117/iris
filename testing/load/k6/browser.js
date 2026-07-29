import { check } from "k6";
import { browser } from "k6/browser";

const targetPage = __ENV.TARGET_PAGE || "http://host.docker.internal:5173";

export const options = {
    scenarios: {
        browser: {
            executor: "constant-vus",
            vus: Number(__ENV.BROWSER_VUS || 5),
            duration: __ENV.DURATION || "30s",
            options: {
                browser: {
                    type: "chromium",
                },
            },
        },
    },
    thresholds: {
        checks: ["rate==1"],
        browser_web_vital_lcp: ["p(95)<2500"],
        browser_web_vital_inp: ["p(95)<200"],
        browser_web_vital_cls: ["p(95)<0.1"],
    },
};

export default async function () {
    const page = await browser.newPage();
    try {
        const response = await page.goto(targetPage);
        check(response, {
            "page loaded": (value) => value?.status() === 200,
        });
        await page.waitForLoadState("networkidle");
    } finally {
        await page.close();
    }
}
