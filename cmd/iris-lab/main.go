package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
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
	if len(os.Args) < 2 || os.Args[1] != "run" {
		return errors.New("usage: iris-lab run --db /path/to/iris.db [options]")
	}

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
	flags.DurationVar(&config.RequestTimeout, "request-timeout", 5*time.Second, "per-request timeout")
	flags.BoolVar(&config.AllowNonLocal, "allow-nonlocal", false, "allow a non-loopback target")
	if err := flags.Parse(os.Args[2:]); err != nil {
		return err
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
