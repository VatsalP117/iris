import fs from "node:fs/promises";
import path from "node:path";
import process from "node:process";
import { pathToFileURL } from "node:url";

export function compareBrowserReports(baseline, candidate) {
    const regressions = [];
    const improvements = [];
    const knownFailures = [];

    const baselineScenarios = scenarioMap(baseline, "baseline");
    const candidateScenarios = scenarioMap(candidate, "candidate");

    for (const name of baselineScenarios.keys()) {
        if (!candidateScenarios.has(name)) {
            regressions.push({
                scenario: name,
                reason: "scenario missing from candidate report",
            });
        }
    }
    for (const name of candidateScenarios.keys()) {
        if (!baselineScenarios.has(name)) {
            regressions.push({
                scenario: name,
                reason: "candidate scenario is not present in the committed baseline",
            });
        }
    }

    for (const [name, baselineScenario] of baselineScenarios) {
        const candidateScenario = candidateScenarios.get(name);
        if (!candidateScenario) {
            continue;
        }
        if (
            (baselineScenario.category ?? "sdk-flow") !==
            (candidateScenario.category ?? "sdk-flow")
        ) {
            regressions.push({
                scenario: name,
                reason: "scenario category changed without a baseline update",
            });
            continue;
        }
        if (baselineScenario.expectedEvents !== candidateScenario.expectedEvents) {
            regressions.push({
                scenario: name,
                reason: "expected event contract changed without a baseline update",
                baseline: baselineScenario.expectedEvents,
                candidate: candidateScenario.expectedEvents,
            });
            continue;
        }
        if (candidateScenario.error) {
            regressions.push({
                scenario: name,
                reason: "scenario ended with a harness or browser error",
                error: candidateScenario.error,
            });
            continue;
        }
        if (baselineScenario.passed && !candidateScenario.passed) {
            regressions.push({
                scenario: name,
                reason: "previously passing scenario now fails",
                expectedEvents: candidateScenario.expectedEvents,
                actualEvents: candidateScenario.actualEvents,
            });
            continue;
        }
        if (!baselineScenario.passed && candidateScenario.passed) {
            improvements.push({ scenario: name });
        } else if (!candidateScenario.passed) {
            knownFailures.push({ scenario: name });
        }

        compareCoverageInvariant(
            name,
            "restoredFromBFCache",
            baselineScenario,
            candidateScenario,
            regressions,
        );
        compareCoverageInvariant(
            name,
            "traces",
            baselineScenario,
            candidateScenario,
            regressions,
        );
        compareCoverageInvariant(
            name,
            "seed",
            baselineScenario,
            candidateScenario,
            regressions,
        );
    }

    return {
        generatedAt: new Date().toISOString(),
        passed: regressions.length === 0,
        baselineGeneratedAt: baseline.generatedAt,
        candidateGeneratedAt: candidate.generatedAt,
        scenarioCount: candidateScenarios.size,
        regressions,
        improvements,
        knownFailures,
    };
}

function scenarioMap(report, label) {
    if (!report || !Array.isArray(report.scenarios)) {
        throw new Error(`${label} report does not contain a scenarios array`);
    }
    const scenarios = new Map();
    for (const scenario of report.scenarios) {
        if (!scenario || typeof scenario.name !== "string" || !scenario.name) {
            throw new Error(`${label} report contains a scenario without a name`);
        }
        if (scenarios.has(scenario.name)) {
            throw new Error(
                `${label} report contains duplicate scenario ${scenario.name}`,
            );
        }
        scenarios.set(scenario.name, scenario);
    }
    return scenarios;
}

function compareCoverageInvariant(
    scenario,
    field,
    baseline,
    candidate,
    regressions,
) {
    if (!(field in baseline)) {
        return;
    }
    if (baseline[field] !== candidate[field]) {
        regressions.push({
            scenario,
            reason: `coverage invariant ${field} changed`,
            baseline: baseline[field],
            candidate: candidate[field],
        });
    }
}

function comparisonMarkdown(comparison) {
    const lines = [
        "# Iris Browser Regression Comparison",
        "",
        `**Verdict:** ${comparison.passed ? "PASS" : "FAIL"}`,
        "",
        `- Scenarios: ${comparison.scenarioCount}`,
        `- Regressions: ${comparison.regressions.length}`,
        `- Improvements: ${comparison.improvements.length}`,
        `- Known failures: ${comparison.knownFailures.length}`,
        "",
    ];
    if (comparison.regressions.length > 0) {
        lines.push("## Regressions", "");
        for (const regression of comparison.regressions) {
            lines.push(`- **${regression.scenario}:** ${regression.reason}`);
        }
        lines.push("");
    }
    if (comparison.improvements.length > 0) {
        lines.push("## Improvements", "");
        for (const improvement of comparison.improvements) {
            lines.push(`- ${improvement.scenario}`);
        }
        lines.push("");
    }
    return lines.join("\n");
}

async function main() {
    const [baselinePath, candidatePath, outputDirectory] = process.argv.slice(2);
    if (!baselinePath || !candidatePath) {
        throw new Error(
            "usage: node compare.mjs BASELINE_JSON CANDIDATE_JSON [OUTPUT_DIRECTORY]",
        );
    }
    const [baseline, candidate] = await Promise.all([
        readJSON(baselinePath),
        readJSON(candidatePath),
    ]);
    const comparison = compareBrowserReports(baseline, candidate);
    const destination = path.resolve(
        outputDirectory ?? path.dirname(candidatePath),
    );
    await fs.mkdir(destination, { recursive: true });
    await fs.writeFile(
        path.join(destination, "browser-comparison.json"),
        `${JSON.stringify(comparison, null, 2)}\n`,
    );
    await fs.writeFile(
        path.join(destination, "browser-comparison.md"),
        comparisonMarkdown(comparison),
    );
    console.log(
        `Iris browser comparison: ${comparison.passed ? "PASS" : "FAIL"} ` +
            `(regressions=${comparison.regressions.length}, ` +
            `improvements=${comparison.improvements.length}, ` +
            `known-failures=${comparison.knownFailures.length})`,
    );
    if (!comparison.passed) {
        process.exitCode = 1;
    }
}

async function readJSON(filename) {
    return JSON.parse(await fs.readFile(filename, "utf8"));
}

if (
    process.argv[1] &&
    import.meta.url === pathToFileURL(path.resolve(process.argv[1])).href
) {
    main().catch((error) => {
        console.error(
            `Iris browser comparison error: ${
                error instanceof Error ? error.message : String(error)
            }`,
        );
        process.exitCode = 1;
    });
}
