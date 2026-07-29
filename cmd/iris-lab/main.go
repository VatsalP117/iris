package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/VatsalP117/iris/internal/reliability"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "iris-lab:", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 2 {
		return errors.New("usage: iris-lab <run|suite|compare|faults> [options]")
	}
	switch os.Args[1] {
	case "run":
		return runLoad(os.Args[2:])
	case "suite":
		return runSuite(os.Args[2:])
	case "faults":
		return runFaults(os.Args[2:])
	case "compare":
		return runCompare(os.Args[2:])
	default:
		return fmt.Errorf("unknown command %q", os.Args[1])
	}
}

func runCompare(arguments []string) error {
	flags := flag.NewFlagSet("compare", flag.ContinueOnError)
	var baseline, candidate, output string
	flags.StringVar(&baseline, "baseline", "", "baseline summary.json")
	flags.StringVar(&candidate, "candidate", "", "candidate summary.json")
	flags.StringVar(&output, "output", "", "comparison artifact directory")
	if err := flags.Parse(arguments); err != nil {
		return err
	}
	if baseline == "" || candidate == "" {
		return errors.New("--baseline and --candidate are required")
	}
	if output == "" {
		output = "artifacts/reliability/comparison-" + time.Now().UTC().Format("20060102T150405Z")
	}
	report, err := reliability.CompareReports(baseline, candidate)
	if err != nil {
		return err
	}
	if err := reliability.WriteComparisonReport(report, output); err != nil {
		return err
	}
	fmt.Printf(
		"Iris Reliability Comparison: passed=%t comparable=%t regressions=%d\n",
		report.Passed,
		report.Comparable,
		report.Regressions,
	)
	if !report.Passed {
		return errors.New("candidate reliability report regressed")
	}
	return nil
}

func runFaults(arguments []string) error {
	flags := flag.NewFlagSet("faults", flag.ContinueOnError)
	var config reliability.FaultSuiteConfig
	flags.StringVar(&config.ServerBinary, "server-bin", "dist/iris-server", "path to the Iris server binary")
	flags.StringVar(&config.OutputDir, "output", "", "fault suite artifact directory")
	flags.StringVar(&config.RunID, "run-id", "", "fault suite run identifier")
	flags.BoolVar(&config.Quick, "quick", false, "use shortened development durations")
	if err := flags.Parse(arguments); err != nil {
		return err
	}

	report, err := reliability.RunFaultSuite(context.Background(), config)
	if err != nil {
		return err
	}
	fmt.Printf("Iris Fault Suite: passed=%t\n", report.Passed)
	for _, scenario := range report.Scenarios {
		fmt.Printf(
			"- %s: passed=%t recovered=%t accepted=%d rejected=%d errors=%d missing=%d\n",
			scenario.Name,
			scenario.Passed,
			scenario.Recovered,
			scenario.AcceptedEvents,
			scenario.RejectedEvents,
			scenario.RequestErrors,
			scenario.MissingAccepted,
		)
	}
	fmt.Printf(
		"- backup: passed=%t integrity=%s rows=%d\n",
		report.Backup.Passed,
		report.Backup.IntegrityCheck,
		report.Backup.StoredRows,
	)
	if !report.Passed {
		return errors.New("one or more fault scenarios failed")
	}
	return nil
}

func runLoad(arguments []string) error {
	flags := flag.NewFlagSet("run", flag.ContinueOnError)
	var config reliability.Config
	flags.StringVar(&config.TargetURL, "target", "http://127.0.0.1:8080", "Iris server base URL")
	flags.StringVar(&config.DBPath, "db", "", "path to the dedicated Iris SQLite database")
	flags.StringVar(&config.RunID, "run-id", "", "run identifier (defaults to a UTC timestamp)")
	flags.StringVar(&config.SiteID, "site-id", "", "isolated site ID (defaults to iris-lab-<run-id>)")
	flags.StringVar(&config.ReportDir, "report-dir", "", "report directory")
	flags.IntVar(&config.Rate, "rate", 10, "offered events per second")
	flags.DurationVar(&config.Duration, "duration", 30*time.Second, "load duration")
	flags.IntVar(&config.EventCount, "events", 0, "exact event count (overrides rate × duration)")
	flags.IntVar(&config.BatchSize, "batch-size", 1, "events per HTTP request, from 1 to 50")
	flags.IntVar(&config.Workers, "workers", 32, "concurrent request workers")
	flags.IntVar(&config.ReadRate, "read-rate", 0, "concurrent analytics read requests per second")
	flags.IntVar(&config.ReadWorkers, "read-workers", 8, "concurrent analytics read workers")
	flags.DurationVar(&config.RequestTimeout, "request-timeout", 5*time.Second, "per-request timeout")
	flags.BoolVar(&config.AllowNonLocal, "allow-nonlocal", false, "allow a non-loopback target")
	var stages string
	flags.StringVar(&stages, "stages", "", "comma-separated RATE:DURATION stages, for example 100:30s,500:1m")
	if err := flags.Parse(arguments); err != nil {
		return err
	}
	if stages != "" {
		parsed, err := parseStages(stages)
		if err != nil {
			return err
		}
		config.Stages = parsed
	}

	report, err := reliability.Run(context.Background(), config)
	if err != nil {
		return err
	}
	directory, err := reliability.WriteReport(report, config.ReportDir)
	if err != nil {
		return fmt.Errorf("write report: %w", err)
	}

	verdict := "PASS"
	if !report.Passed {
		verdict = "FAIL"
	}
	fmt.Printf(
		"Iris Reliability Lab: %s\nReport: %s\nAccepted: %d/%d events\nStored: %d rows; missing=%d duplicate=%d unexpected=%d mismatched=%d\n",
		verdict,
		directory,
		report.Load.AcceptedEvents,
		report.Load.PlannedEvents,
		report.Storage.StoredRows,
		report.Storage.MissingEvents,
		report.Storage.DuplicateRows,
		report.Storage.UnexpectedRows,
		report.Storage.FieldMismatches,
	)
	if !report.Passed {
		return errors.New("reliability checks failed")
	}
	return nil
}

func runSuite(arguments []string) error {
	flags := flag.NewFlagSet("suite", flag.ContinueOnError)
	var config reliability.SuiteConfig
	var profiles string
	flags.StringVar(&config.ServerBinary, "server-bin", "dist/iris-server", "path to the Iris server binary")
	flags.StringVar(&config.OutputDir, "output", "", "suite artifact directory")
	flags.StringVar(&config.RunID, "run-id", "", "suite run identifier")
	flags.StringVar(&profiles, "profiles", "smoke", "comma-separated suite profiles")
	flags.BoolVar(&config.Quick, "quick", false, "use shortened development durations")
	flags.BoolVar(&config.CapturePprof, "pprof", true, "capture CPU, heap, and goroutine profiles")
	if err := flags.Parse(arguments); err != nil {
		return err
	}
	for _, profile := range strings.Split(profiles, ",") {
		if trimmed := strings.TrimSpace(profile); trimmed != "" {
			config.Profiles = append(config.Profiles, trimmed)
		}
	}

	report, err := reliability.RunSuite(context.Background(), config)
	if err != nil {
		return err
	}
	fmt.Printf("Iris Reliability Suite: %s\n", map[bool]string{true: "PASS", false: "FAIL"}[report.Passed])
	for _, profile := range report.Profiles {
		fmt.Printf(
			"- %s: passed=%t events=%d rate=%.2f p95=%.2fms\n",
			profile.Name,
			profile.Passed,
			profile.Events,
			profile.EventRate,
			profile.P95MS,
		)
	}
	if !report.Passed {
		return errors.New("one or more reliability profiles failed")
	}
	return nil
}

func parseStages(value string) ([]reliability.RateStage, error) {
	parts := strings.Split(value, ",")
	stages := make([]reliability.RateStage, 0, len(parts))
	for index, part := range parts {
		fields := strings.SplitN(strings.TrimSpace(part), ":", 2)
		if len(fields) != 2 {
			return nil, fmt.Errorf("invalid stage %d %q: expected RATE:DURATION", index, part)
		}
		rate, err := strconv.Atoi(fields[0])
		if err != nil || rate <= 0 {
			return nil, fmt.Errorf("invalid stage %d rate %q", index, fields[0])
		}
		duration, err := time.ParseDuration(fields[1])
		if err != nil || duration <= 0 {
			return nil, fmt.Errorf("invalid stage %d duration %q", index, fields[1])
		}
		stages = append(stages, reliability.RateStage{Rate: rate, Duration: duration})
	}
	return stages, nil
}
