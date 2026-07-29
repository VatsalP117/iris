import assert from "node:assert/strict";
import test from "node:test";

import { compareBrowserReports } from "./compare.mjs";

function report(scenarios) {
    return {
        generatedAt: "2026-07-29T00:00:00Z",
        scenarios,
    };
}

function scenario(name, passed, overrides = {}) {
    return {
        name,
        category: "sdk-flow",
        passed,
        expectedEvents: 1,
        actualEvents: passed ? 1 : 0,
        ...overrides,
    };
}

test("identical reports preserve known failures without failing", () => {
    const baseline = report([
        scenario("passing", true),
        scenario("known-failure", false),
    ]);
    const comparison = compareBrowserReports(baseline, structuredClone(baseline));

    assert.equal(comparison.passed, true);
    assert.deepEqual(comparison.regressions, []);
    assert.deepEqual(comparison.knownFailures, [
        { scenario: "known-failure" },
    ]);
});

test("a previously passing scenario becoming a failure is a regression", () => {
    const baseline = report([scenario("navigation", true)]);
    const candidate = report([scenario("navigation", false)]);
    const comparison = compareBrowserReports(baseline, candidate);

    assert.equal(comparison.passed, false);
    assert.equal(comparison.regressions[0].scenario, "navigation");
});

test("a known failure becoming a pass is an improvement", () => {
    const baseline = report([scenario("retry", false)]);
    const candidate = report([scenario("retry", true)]);
    const comparison = compareBrowserReports(baseline, candidate);

    assert.equal(comparison.passed, true);
    assert.deepEqual(comparison.improvements, [{ scenario: "retry" }]);
});

test("scenario set and contract changes require a baseline update", () => {
    const baseline = report([scenario("existing", true)]);
    const candidate = report([
        scenario("existing", true, { expectedEvents: 2 }),
        scenario("new-scenario", true),
    ]);
    const comparison = compareBrowserReports(baseline, candidate);

    assert.equal(comparison.passed, false);
    assert.equal(comparison.regressions.length, 2);
});

test("harness errors and lost coverage are regressions", () => {
    const baseline = report([
        scenario("bfcache", false, { restoredFromBFCache: true }),
        scenario("state-machine", false, { traces: 16, seed: 42 }),
    ]);
    const candidate = report([
        scenario("bfcache", false, {
            restoredFromBFCache: false,
        }),
        scenario("state-machine", false, {
            traces: 15,
            seed: 42,
            error: "browser crashed",
        }),
    ]);
    const comparison = compareBrowserReports(baseline, candidate);

    assert.equal(comparison.passed, false);
    assert.equal(comparison.regressions.length, 2);
});
