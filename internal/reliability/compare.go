package reliability

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type MetricComparison struct {
	Name          string  `json:"name"`
	Baseline      float64 `json:"baseline"`
	Candidate     float64 `json:"candidate"`
	Delta         float64 `json:"delta"`
	PercentChange float64 `json:"percent_change"`
	LowerIsBetter bool    `json:"lower_is_better"`
	Regression    bool    `json:"regression"`
}

type ComparisonReport struct {
	GeneratedAt   time.Time          `json:"generated_at"`
	BaselinePath  string             `json:"baseline_path"`
	CandidatePath string             `json:"candidate_path"`
	Comparable    bool               `json:"comparable"`
	Regressions   int                `json:"regressions"`
	Metrics       []MetricComparison `json:"metrics"`
	Passed        bool               `json:"passed"`
	Error         string             `json:"error,omitempty"`
}

func CompareReports(baselinePath, candidatePath string) (*ComparisonReport, error) {
	baseline, err := ReadReport(baselinePath)
	if err != nil {
		return nil, fmt.Errorf("read baseline: %w", err)
	}
	candidate, err := ReadReport(candidatePath)
	if err != nil {
		return nil, fmt.Errorf("read candidate: %w", err)
	}

	comparison := &ComparisonReport{
		GeneratedAt:   time.Now().UTC(),
		BaselinePath:  baselinePath,
		CandidatePath: candidatePath,
		Passed:        true,
	}
	if !comparableConfigs(baseline.Config, candidate.Config) {
		comparison.Passed = false
		comparison.Error = "load configurations differ"
		return comparison, nil
	}
	if !comparableEnvironments(baseline.Environment, candidate.Environment) {
		comparison.Passed = false
		comparison.Error = "execution environments differ"
		return comparison, nil
	}
	comparison.Comparable = true

	comparison.Metrics = []MetricComparison{
		compareMetric("achieved events/s", baseline.Load.AchievedEventsPerSec, candidate.Load.AchievedEventsPerSec, false, 10),
		compareMetric("write p95 ms", baseline.Load.Latency.P95MS, candidate.Load.Latency.P95MS, true, 20),
		compareMetric("write p99 ms", baseline.Load.Latency.P99MS, candidate.Load.Latency.P99MS, true, 20),
		compareMetric("request errors", float64(baseline.Load.RequestErrors), float64(candidate.Load.RequestErrors), true, 0),
		compareMetric("rejected events", float64(baseline.Load.RejectedEvents), float64(candidate.Load.RejectedEvents), true, 0),
		compareMetric("missing accepted", float64(baseline.Storage.MissingEvents), float64(candidate.Storage.MissingEvents), true, 0),
		compareMetric("duplicate rows", float64(baseline.Storage.DuplicateRows), float64(candidate.Storage.DuplicateRows), true, 0),
		compareMetric("read p95 ms", baseline.Load.Reads.Latency.P95MS, candidate.Load.Reads.Latency.P95MS, true, 20),
		compareMetric("peak CPU percent", baseline.Resources.PeakCPUPercent, candidate.Resources.PeakCPUPercent, true, 20),
		compareMetric("peak RSS bytes", float64(baseline.Resources.PeakRSSBytes), float64(candidate.Resources.PeakRSSBytes), true, 20),
		compareMetric("database bytes/event", databaseBytesPerEvent(baseline), databaseBytesPerEvent(candidate), true, 10),
	}
	if len(baseline.Load.Stages) == len(candidate.Load.Stages) {
		for index := range baseline.Load.Stages {
			baselineStage := baseline.Load.Stages[index]
			candidateStage := candidate.Load.Stages[index]
			prefix := fmt.Sprintf("stage %d events/s ", baselineStage.Rate)
			comparison.Metrics = append(
				comparison.Metrics,
				compareMetric(
					prefix+"achieved events/s",
					baselineStage.AchievedEventsPerSec,
					candidateStage.AchievedEventsPerSec,
					false,
					10,
				),
				compareMetric(
					prefix+"p95 ms",
					baselineStage.Latency.P95MS,
					candidateStage.Latency.P95MS,
					true,
					20,
				),
				compareMetric(
					prefix+"rejected events",
					float64(baselineStage.RejectedEvents),
					float64(candidateStage.RejectedEvents),
					true,
					0,
				),
			)
		}
	}
	for _, metric := range comparison.Metrics {
		if metric.Regression {
			comparison.Regressions++
		}
	}
	comparison.Passed = comparison.Regressions == 0 && candidate.Passed
	return comparison, nil
}

func comparableEnvironments(baseline, candidate Environment) bool {
	return baseline.GOOS == candidate.GOOS &&
		baseline.GOARCH == candidate.GOARCH &&
		baseline.CPUs == candidate.CPUs
}

func ReadReport(path string) (*Report, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var report Report
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, err
	}
	return &report, nil
}

func compareMetric(
	name string,
	baseline, candidate float64,
	lowerIsBetter bool,
	tolerancePercent float64,
) MetricComparison {
	comparison := MetricComparison{
		Name:          name,
		Baseline:      baseline,
		Candidate:     candidate,
		Delta:         candidate - baseline,
		LowerIsBetter: lowerIsBetter,
	}
	if baseline != 0 {
		comparison.PercentChange = comparison.Delta / baseline * 100
	} else if candidate != 0 {
		comparison.PercentChange = 100
	}

	if lowerIsBetter {
		if baseline == 0 {
			comparison.Regression = candidate > 0
		} else {
			comparison.Regression = comparison.PercentChange > tolerancePercent
		}
	} else if baseline > 0 {
		comparison.Regression = comparison.PercentChange < -tolerancePercent
	}
	return comparison
}

func comparableConfigs(baseline, candidate RunConfig) bool {
	return baseline.EventCount == candidate.EventCount &&
		baseline.BatchSize == candidate.BatchSize &&
		baseline.Rate == candidate.Rate &&
		baseline.ReadRate == candidate.ReadRate &&
		fmt.Sprint(baseline.Stages) == fmt.Sprint(candidate.Stages)
}

func databaseBytesPerEvent(report *Report) float64 {
	if report.Load.AcceptedEvents == 0 {
		return 0
	}
	return float64(report.Resources.DatabaseGrowthBytes) / float64(report.Load.AcceptedEvents)
}

func WriteComparisonReport(report *ComparisonReport, directory string) error {
	if err := os.MkdirAll(directory, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.WriteFile(filepath.Join(directory, "comparison.json"), data, 0644); err != nil {
		return err
	}

	var builder strings.Builder
	builder.WriteString("# Iris Reliability Comparison\n\n")
	if report.Passed {
		builder.WriteString("**Verdict:** PASS\n\n")
	} else {
		builder.WriteString("**Verdict:** FAIL\n\n")
	}
	fmt.Fprintf(&builder, "- Baseline: `%s`\n", report.BaselinePath)
	fmt.Fprintf(&builder, "- Candidate: `%s`\n", report.CandidatePath)
	fmt.Fprintf(&builder, "- Comparable: %t\n", report.Comparable)
	fmt.Fprintf(&builder, "- Regressions: %d\n\n", report.Regressions)
	if report.Error != "" {
		fmt.Fprintf(&builder, "Error: `%s`\n\n", report.Error)
	}
	builder.WriteString("| Metric | Baseline | Candidate | Change | Regression |\n")
	builder.WriteString("|---|---:|---:|---:|---|\n")
	for _, metric := range report.Metrics {
		fmt.Fprintf(
			&builder,
			"| %s | %.2f | %.2f | %+.2f%% | %t |\n",
			metric.Name,
			metric.Baseline,
			metric.Candidate,
			metric.PercentChange,
			metric.Regression,
		)
	}
	return os.WriteFile(filepath.Join(directory, "comparison.md"), []byte(builder.String()), 0644)
}
