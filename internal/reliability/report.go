package reliability

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func WriteReport(report *Report, directory string) (string, error) {
	if directory == "" {
		directory = filepath.Join("artifacts", "reliability", report.Config.RunID)
	}
	if err := os.MkdirAll(directory, 0755); err != nil {
		return "", err
	}

	jsonPath := filepath.Join(directory, "summary.json")
	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", err
	}
	jsonData = append(jsonData, '\n')
	if err := os.WriteFile(jsonPath, jsonData, 0644); err != nil {
		return "", err
	}

	markdownPath := filepath.Join(directory, "report.md")
	if err := os.WriteFile(markdownPath, []byte(markdownReport(report)), 0644); err != nil {
		return "", err
	}
	return directory, nil
}

func markdownReport(report *Report) string {
	var builder strings.Builder
	status := "PASS"
	if !report.Passed {
		status = "FAIL"
	}

	fmt.Fprintf(&builder, "# Iris Reliability Report: %s\n\n", report.Config.RunID)
	fmt.Fprintf(&builder, "**Verdict:** %s\n\n", status)
	fmt.Fprintf(&builder, "Generated at: %s\n\n", report.GeneratedAt.Format("2006-01-02 15:04:05 UTC"))

	builder.WriteString("## Configuration\n\n")
	builder.WriteString("| Field | Value |\n|---|---:|\n")
	fmt.Fprintf(&builder, "| Target | `%s` |\n", report.Config.TargetURL)
	fmt.Fprintf(&builder, "| Database | `%s` |\n", report.Config.DBPath)
	fmt.Fprintf(&builder, "| Site | `%s` |\n", report.Config.SiteID)
	fmt.Fprintf(&builder, "| Offered rate | %d events/s |\n", report.Config.Rate)
	fmt.Fprintf(&builder, "| Planned events | %d |\n", report.Config.EventCount)
	fmt.Fprintf(&builder, "| Batch size | %d |\n", report.Config.BatchSize)
	fmt.Fprintf(&builder, "| Workers | %d |\n", report.Config.Workers)
	fmt.Fprintf(&builder, "| Concurrent read rate | %d requests/s |\n", report.Config.ReadRate)
	fmt.Fprintf(&builder, "| Read workers | %d |\n\n", report.Config.ReadWorkers)
	if len(report.Config.Stages) > 0 {
		builder.WriteString("### Rate stages\n\n")
		builder.WriteString("| Rate | Duration |\n|---:|---:|\n")
		for _, stage := range report.Config.Stages {
			fmt.Fprintf(&builder, "| %d events/s | %s |\n", stage.Rate, stage.Duration)
		}
		builder.WriteString("\n")
	}

	builder.WriteString("## Environment\n\n")
	builder.WriteString("| Field | Value |\n|---|---|\n")
	fmt.Fprintf(&builder, "| Git revision | `%s` |\n", report.Environment.GitRevision)
	fmt.Fprintf(&builder, "| Git worktree modified | %t |\n", report.Environment.GitModified)
	fmt.Fprintf(&builder, "| Go | `%s` |\n", report.Environment.GoVersion)
	fmt.Fprintf(&builder, "| Platform | `%s/%s` |\n", report.Environment.GOOS, report.Environment.GOARCH)
	fmt.Fprintf(&builder, "| CPUs visible to load generator | %d |\n\n", report.Environment.CPUs)

	builder.WriteString("## Delivery\n\n")
	builder.WriteString("| Measurement | Value |\n|---|---:|\n")
	fmt.Fprintf(&builder, "| Events attempted | %d |\n", report.Load.AttemptedEvents)
	fmt.Fprintf(&builder, "| Events accepted | %d |\n", report.Load.AcceptedEvents)
	fmt.Fprintf(&builder, "| Events rejected | %d |\n", report.Load.RejectedEvents)
	fmt.Fprintf(&builder, "| Request errors | %d |\n", report.Load.RequestErrors)
	fmt.Fprintf(&builder, "| Requests attempted | %d |\n", report.Load.AttemptedRequests)
	fmt.Fprintf(&builder, "| Achieved event rate | %.2f events/s |\n", report.Load.AchievedEventsPerSec)
	fmt.Fprintf(&builder, "| Achieved request rate | %.2f requests/s |\n", report.Load.AchievedRequestsPerSec)
	fmt.Fprintf(&builder, "| Maximum scheduling lag | %.2f ms |\n\n", report.Load.MaxScheduleLagMS)

	if len(report.Load.StatusCodes) > 0 {
		builder.WriteString("### HTTP statuses\n\n")
		builder.WriteString("| Status | Requests |\n|---|---:|\n")
		statuses := make([]int, 0, len(report.Load.StatusCodes))
		for status := range report.Load.StatusCodes {
			statuses = append(statuses, status)
		}
		sort.Ints(statuses)
		for _, status := range statuses {
			fmt.Fprintf(&builder, "| %d | %d |\n", status, report.Load.StatusCodes[status])
		}
		builder.WriteString("\n")
	}

	if report.Config.ReadRate > 0 {
		builder.WriteString("## Concurrent analytics reads\n\n")
		builder.WriteString("| Measurement | Value |\n|---|---:|\n")
		fmt.Fprintf(&builder, "| Requests attempted | %d |\n", report.Load.Reads.AttemptedRequests)
		fmt.Fprintf(&builder, "| Requests successful | %d |\n", report.Load.Reads.SuccessfulRequests)
		fmt.Fprintf(&builder, "| Requests failed | %d |\n", report.Load.Reads.FailedRequests)
		fmt.Fprintf(&builder, "| Achieved rate | %.2f requests/s |\n", report.Load.Reads.AchievedRequestsPerSec)
		fmt.Fprintf(&builder, "| p95 latency | %.2f ms |\n", report.Load.Reads.Latency.P95MS)
		fmt.Fprintf(&builder, "| p99 latency | %.2f ms |\n\n", report.Load.Reads.Latency.P99MS)
	}

	if len(report.Load.Stages) > 0 {
		builder.WriteString("## Stage results\n\n")
		builder.WriteString("| Rate | Result | Accepted / planned | Events/s | p95 | p99 | Max lag |\n")
		builder.WriteString("|---:|---|---:|---:|---:|---:|---:|\n")
		firstFailedRate := 0
		for _, stage := range report.Load.Stages {
			passed := stage.AttemptedEvents == stage.PlannedEvents &&
				stage.AcceptedEvents == stage.AttemptedEvents &&
				stage.RequestErrors == 0
			result := "PASS"
			if !passed {
				result = "FAIL"
				if firstFailedRate == 0 {
					firstFailedRate = stage.Rate
				}
			}
			fmt.Fprintf(
				&builder,
				"| %d | %s | %d / %d | %.2f | %.2f ms | %.2f ms | %.2f ms |\n",
				stage.Rate,
				result,
				stage.AcceptedEvents,
				stage.PlannedEvents,
				stage.AchievedEventsPerSec,
				stage.Latency.P95MS,
				stage.Latency.P99MS,
				stage.MaxScheduleLagMS,
			)
		}
		if firstFailedRate == 0 {
			builder.WriteString("\nFirst failed offered load: none in this profile.\n\n")
		} else {
			fmt.Fprintf(&builder, "\nFirst failed offered load: **%d events/s**.\n\n", firstFailedRate)
		}
	}

	builder.WriteString("## Storage reconciliation\n\n")
	builder.WriteString("| Measurement | Value |\n|---|---:|\n")
	fmt.Fprintf(&builder, "| Stored rows | %d |\n", report.Storage.StoredRows)
	fmt.Fprintf(&builder, "| Unique sequences | %d |\n", report.Storage.UniqueSequences)
	fmt.Fprintf(&builder, "| Missing accepted events | %d |\n", report.Storage.MissingEvents)
	fmt.Fprintf(&builder, "| Duplicate rows | %d |\n", report.Storage.DuplicateRows)
	fmt.Fprintf(&builder, "| Unexpected rows | %d |\n", report.Storage.UnexpectedRows)
	fmt.Fprintf(&builder, "| Field mismatches | %d |\n", report.Storage.FieldMismatches)
	fmt.Fprintf(&builder, "| Database size | %d bytes |\n\n", report.Storage.DatabaseBytes)

	builder.WriteString("## Request latency\n\n")
	builder.WriteString("| Percentile | Milliseconds |\n|---|---:|\n")
	fmt.Fprintf(&builder, "| Average | %.2f |\n", report.Load.Latency.AverageMS)
	fmt.Fprintf(&builder, "| p50 | %.2f |\n", report.Load.Latency.P50MS)
	fmt.Fprintf(&builder, "| p90 | %.2f |\n", report.Load.Latency.P90MS)
	fmt.Fprintf(&builder, "| p95 | %.2f |\n", report.Load.Latency.P95MS)
	fmt.Fprintf(&builder, "| p99 | %.2f |\n", report.Load.Latency.P99MS)
	fmt.Fprintf(&builder, "| Maximum | %.2f |\n\n", report.Load.Latency.MaxMS)

	if report.Resources.Samples > 0 {
		builder.WriteString("## Server resources\n\n")
		builder.WriteString("| Measurement | Value |\n|---|---:|\n")
		fmt.Fprintf(&builder, "| Samples | %d |\n", report.Resources.Samples)
		fmt.Fprintf(&builder, "| Average CPU | %.2f%% |\n", report.Resources.AverageCPUPercent)
		fmt.Fprintf(&builder, "| Peak CPU | %.2f%% |\n", report.Resources.PeakCPUPercent)
		fmt.Fprintf(&builder, "| Average RSS | %d bytes |\n", report.Resources.AverageRSSBytes)
		fmt.Fprintf(&builder, "| Peak RSS | %d bytes |\n", report.Resources.PeakRSSBytes)
		fmt.Fprintf(&builder, "| Database growth | %d bytes |\n", report.Resources.DatabaseGrowthBytes)
		fmt.Fprintf(&builder, "| Peak WAL | %d bytes |\n\n", report.Resources.PeakWALBytes)
	}

	builder.WriteString("## Aggregate checks\n\n")
	builder.WriteString("| Check | Result | Error |\n|---|---|---|\n")
	for _, check := range report.Aggregates {
		result := "PASS"
		if !check.Passed {
			result = "FAIL"
		}
		fmt.Fprintf(&builder, "| %s | %s | %s |\n", check.Name, result, check.Error)
	}

	if len(report.Load.ErrorSamples) > 0 ||
		len(report.Storage.MissingSamples) > 0 ||
		len(report.Storage.DuplicateSamples) > 0 ||
		len(report.Storage.UnexpectedSamples) > 0 ||
		len(report.Storage.FieldMismatchSamples) > 0 {
		builder.WriteString("\n## Diagnostic samples\n\n")
		writeStringSamples(&builder, "Request errors", report.Load.ErrorSamples)
		writeIntSamples(&builder, "Missing sequences", report.Storage.MissingSamples)
		writeIntSamples(&builder, "Duplicate sequences", report.Storage.DuplicateSamples)
		writeIntSamples(&builder, "Unexpected sequences", report.Storage.UnexpectedSamples)
		writeIntSamples(&builder, "Field mismatch sequences", report.Storage.FieldMismatchSamples)
	}

	return builder.String()
}

func writeStringSamples(builder *strings.Builder, label string, samples []string) {
	if len(samples) == 0 {
		return
	}
	fmt.Fprintf(builder, "- %s: `%s`\n", label, strings.Join(samples, "`, `"))
}

func writeIntSamples(builder *strings.Builder, label string, samples []int) {
	if len(samples) == 0 {
		return
	}
	values := make([]string, len(samples))
	for i, sample := range samples {
		values[i] = fmt.Sprintf("%d", sample)
	}
	fmt.Fprintf(builder, "- %s: `%s`\n", label, strings.Join(values, "`, `"))
}
