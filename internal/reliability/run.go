package reliability

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"runtime"
	"runtime/debug"
	"sort"
	"strings"
	"sync"
	"time"
)

type requestJob struct {
	Events     []PlannedEvent
	DueAt      time.Time
	StageIndex int
}

type requestResult struct {
	Sequences   []int
	Status      int
	Response    string
	StageIndex  int
	Latency     time.Duration
	ScheduleLag time.Duration
	Err         error
}

type loadResult struct {
	Summary  LoadSummary
	Accepted map[int]struct{}
}

func Run(ctx context.Context, config Config) (*Report, error) {
	normalized, count, err := normalizeConfig(config)
	if err != nil {
		return nil, err
	}

	manifest := BuildManifest(normalized.RunID, normalized.SiteID, count)
	load, err := executeLoad(ctx, normalized, manifest)
	if err != nil {
		return nil, err
	}

	storage, err := VerifyStorage(ctx, normalized.DBPath, manifest, load.Accepted)
	if err != nil {
		return nil, fmt.Errorf("verify storage: %w", err)
	}

	aggregates := VerifyAggregates(ctx, normalized, manifest, load.Accepted)
	report := &Report{
		GeneratedAt: time.Now().UTC(),
		Config: RunConfig{
			TargetURL:   normalized.TargetURL,
			DBPath:      normalized.DBPath,
			RunID:       normalized.RunID,
			SiteID:      normalized.SiteID,
			Rate:        normalized.Rate,
			Duration:    normalized.Duration.String(),
			EventCount:  count,
			BatchSize:   normalized.BatchSize,
			Workers:     normalized.Workers,
			ReadRate:    normalized.ReadRate,
			ReadWorkers: normalized.ReadWorkers,
			Stages:      stageConfigs(normalized.Stages),
		},
		Environment: currentEnvironment(),
		Load:        load.Summary,
		Storage:     storage,
		Aggregates:  aggregates,
	}
	report.Passed = reportPasses(report)
	return report, nil
}

func normalizeConfig(config Config) (Config, int, error) {
	if config.TargetURL == "" {
		config.TargetURL = "http://127.0.0.1:8080"
	}
	config.TargetURL = strings.TrimRight(config.TargetURL, "/")

	parsed, err := url.Parse(config.TargetURL)
	if err != nil || parsed.Scheme == "" || parsed.Hostname() == "" {
		return Config{}, 0, fmt.Errorf("invalid target URL %q", config.TargetURL)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return Config{}, 0, fmt.Errorf("target URL must use http or https")
	}
	if !config.AllowNonLocal && !isLocalHost(parsed.Hostname()) {
		return Config{}, 0, fmt.Errorf("refusing non-local target %q without --allow-nonlocal", parsed.Hostname())
	}
	if config.DBPath == "" {
		return Config{}, 0, fmt.Errorf("database path is required")
	}
	if config.RunID == "" {
		config.RunID = time.Now().UTC().Format("20060102T150405.000000000Z")
	}
	if !validRunID(config.RunID) {
		return Config{}, 0, fmt.Errorf("run ID must be 1-128 characters using only letters, numbers, dot, underscore, or hyphen")
	}
	if config.SiteID == "" {
		config.SiteID = "iris-lab-" + config.RunID
	}
	if config.Rate <= 0 {
		config.Rate = 10
	}
	if config.Duration <= 0 && config.EventCount <= 0 {
		config.Duration = 30 * time.Second
	}
	if len(config.Stages) > 0 {
		if config.EventCount > 0 {
			return Config{}, 0, fmt.Errorf("event count and staged rates cannot be used together")
		}
		config.Duration = stagesDuration(config.Stages)
		config.Rate = config.Stages[0].Rate
	}
	if config.BatchSize <= 0 {
		config.BatchSize = 1
	}
	if config.BatchSize > 50 {
		return Config{}, 0, fmt.Errorf("batch size %d exceeds the backend limit of 50", config.BatchSize)
	}
	if config.Workers <= 0 {
		config.Workers = 32
	}
	if config.RequestTimeout <= 0 {
		config.RequestTimeout = 5 * time.Second
	}
	if config.ReadRate < 0 {
		return Config{}, 0, fmt.Errorf("read rate cannot be negative")
	}
	if config.ReadWorkers <= 0 {
		config.ReadWorkers = 8
	}

	count, err := config.plannedEventCount()
	if err != nil {
		return Config{}, 0, err
	}
	return config, count, nil
}

func isLocalHost(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func validRunID(value string) bool {
	if len(value) == 0 || len(value) > 128 {
		return false
	}
	for _, character := range value {
		if character >= 'a' && character <= 'z' ||
			character >= 'A' && character <= 'Z' ||
			character >= '0' && character <= '9' ||
			character == '.' ||
			character == '_' ||
			character == '-' {
			continue
		}
		return false
	}
	return true
}

func executeLoad(ctx context.Context, config Config, manifest []PlannedEvent) (loadResult, error) {
	readContext, stopReads := context.WithCancel(ctx)
	readResults := make(chan ReadSummary, 1)
	if config.ReadRate > 0 {
		go func() {
			readResults <- executeReadLoad(readContext, config)
		}()
	}

	jobs := make(chan requestJob, config.Workers*8)
	results := make(chan requestResult, config.Workers*8)
	client := &http.Client{Timeout: config.RequestTimeout}

	var workers sync.WaitGroup
	for i := 0; i < config.Workers; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for job := range jobs {
				results <- sendRequest(ctx, client, config.TargetURL, config.BatchSize, job)
			}
		}()
	}

	go func() {
		workers.Wait()
		close(results)
	}()

	startedAt := time.Now()
	go func() {
		defer close(jobs)
		for offset := 0; offset < len(manifest); {
			end := min(offset+config.BatchSize, len(manifest))
			stageIndex, stageEnd := stageForOffset(config, offset, len(manifest))
			if stageEnd > offset && end > stageEnd {
				end = stageEnd
			}
			dueAt := startedAt.Add(eventDueOffset(config, offset))
			if wait := time.Until(dueAt); wait > 0 {
				timer := time.NewTimer(wait)
				select {
				case <-ctx.Done():
					timer.Stop()
					return
				case <-timer.C:
				}
			}
			select {
			case <-ctx.Done():
				return
			case jobs <- requestJob{
				Events:     manifest[offset:end],
				DueAt:      dueAt,
				StageIndex: stageIndex,
			}:
			}
			offset = end
		}
	}()

	summary := LoadSummary{
		PlannedEvents: len(manifest),
		StatusCodes:   map[int]int{},
	}
	accepted := make(map[int]struct{}, len(manifest))
	latencies := make([]time.Duration, 0, (len(manifest)+config.BatchSize-1)/config.BatchSize)
	stageLatencies := make([][]time.Duration, len(config.Stages))
	if len(config.Stages) > 0 {
		summary.Stages = make([]StageSummary, len(config.Stages))
		for index, stage := range config.Stages {
			summary.Stages[index] = StageSummary{
				Rate:          stage.Rate,
				Duration:      stage.Duration.String(),
				PlannedEvents: int(float64(stage.Rate) * stage.Duration.Seconds()),
				StatusCodes:   map[int]int{},
			}
		}
	}

	for result := range results {
		eventCount := len(result.Sequences)
		summary.AttemptedEvents += eventCount
		summary.AttemptedRequests++
		latencies = append(latencies, result.Latency)
		if result.ScheduleLag.Seconds()*1000 > summary.MaxScheduleLagMS {
			summary.MaxScheduleLagMS = result.ScheduleLag.Seconds() * 1000
		}
		var stage *StageSummary
		if result.StageIndex >= 0 && result.StageIndex < len(summary.Stages) {
			stage = &summary.Stages[result.StageIndex]
			stage.AttemptedEvents += eventCount
			stage.AttemptedRequests++
			stageLatencies[result.StageIndex] = append(stageLatencies[result.StageIndex], result.Latency)
			if lagMS := result.ScheduleLag.Seconds() * 1000; lagMS > stage.MaxScheduleLagMS {
				stage.MaxScheduleLagMS = lagMS
			}
		}

		if result.Err != nil {
			summary.RequestErrors++
			summary.RejectedEvents += eventCount
			if stage != nil {
				stage.RequestErrors++
				stage.RejectedEvents += eventCount
			}
			appendErrorSample(&summary.ErrorSamples, result.Err.Error())
			continue
		}

		summary.StatusCodes[result.Status]++
		if stage != nil {
			stage.StatusCodes[result.Status]++
		}
		if result.Status == http.StatusAccepted {
			summary.AcceptedRequests++
			summary.AcceptedEvents += eventCount
			for _, sequence := range result.Sequences {
				accepted[sequence] = struct{}{}
			}
			if stage != nil {
				stage.AcceptedEvents += eventCount
			}
			continue
		}

		summary.RejectedRequests++
		summary.RejectedEvents += eventCount
		if stage != nil {
			stage.RejectedEvents += eventCount
		}
		detail := strings.TrimSpace(result.Response)
		if detail != "" {
			appendErrorSample(
				&summary.ErrorSamples,
				fmt.Sprintf("HTTP %d for %d event(s): %s", result.Status, eventCount, detail),
			)
		} else {
			appendErrorSample(&summary.ErrorSamples, fmt.Sprintf("HTTP %d for %d event(s)", result.Status, eventCount))
		}
	}

	elapsed := time.Since(startedAt)
	summary.ElapsedSeconds = elapsed.Seconds()
	if summary.ElapsedSeconds > 0 {
		summary.AchievedEventsPerSec = float64(summary.AttemptedEvents) / summary.ElapsedSeconds
		summary.AchievedRequestsPerSec = float64(summary.AttemptedRequests) / summary.ElapsedSeconds
	}
	summary.Latency = summarizeLatencies(latencies)
	for index := range summary.Stages {
		stage := &summary.Stages[index]
		duration := config.Stages[index].Duration.Seconds()
		if duration > 0 {
			stage.AchievedEventsPerSec = float64(stage.AttemptedEvents) / duration
			stage.AchievedRequestsPerSec = float64(stage.AttemptedRequests) / duration
		}
		stage.Latency = summarizeLatencies(stageLatencies[index])
	}
	stopReads()
	if config.ReadRate > 0 {
		summary.Reads = <-readResults
	}

	if ctx.Err() != nil {
		return loadResult{}, ctx.Err()
	}
	return loadResult{Summary: summary, Accepted: accepted}, nil
}

func sendRequest(ctx context.Context, client *http.Client, targetURL string, batchSize int, job requestJob) requestResult {
	sequences := make([]int, len(job.Events))
	events := make([]any, len(job.Events))
	for i, planned := range job.Events {
		sequences[i] = planned.Sequence
		events[i] = planned.Event
	}

	var payload any = events
	endpoint := "/api/events"
	if batchSize == 1 && len(job.Events) == 1 {
		payload = job.Events[0].Event
		endpoint = "/api/event"
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return requestResult{Sequences: sequences, StageIndex: job.StageIndex, Err: err}
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL+endpoint, bytes.NewReader(body))
	if err != nil {
		return requestResult{Sequences: sequences, StageIndex: job.StageIndex, Err: err}
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", "iris-reliability-lab")

	sentAt := time.Now()
	response, err := client.Do(request)
	latency := time.Since(sentAt)
	scheduleLag := sentAt.Sub(job.DueAt)
	if scheduleLag < 0 {
		scheduleLag = 0
	}
	if err != nil {
		return requestResult{
			Sequences:   sequences,
			StageIndex:  job.StageIndex,
			Latency:     latency,
			ScheduleLag: scheduleLag,
			Err:         err,
		}
	}
	defer response.Body.Close()
	responseBody, _ := io.ReadAll(io.LimitReader(response.Body, 64<<10))
	return requestResult{
		Sequences:   sequences,
		Status:      response.StatusCode,
		Response:    string(responseBody),
		StageIndex:  job.StageIndex,
		Latency:     latency,
		ScheduleLag: scheduleLag,
	}
}

func stageForOffset(config Config, eventOffset, eventCount int) (int, int) {
	if len(config.Stages) == 0 {
		return -1, eventCount
	}
	boundary := 0
	for index, stage := range config.Stages {
		boundary += int(float64(stage.Rate) * stage.Duration.Seconds())
		if eventOffset < boundary {
			return index, boundary
		}
	}
	return len(config.Stages) - 1, eventCount
}

func eventDueOffset(config Config, eventOffset int) time.Duration {
	if len(config.Stages) == 0 {
		return time.Duration(float64(time.Second) * float64(eventOffset) / float64(config.Rate))
	}

	remaining := eventOffset
	var elapsed time.Duration
	for _, stage := range config.Stages {
		stageEvents := int(float64(stage.Rate) * stage.Duration.Seconds())
		if remaining < stageEvents {
			return elapsed + time.Duration(float64(time.Second)*float64(remaining)/float64(stage.Rate))
		}
		remaining -= stageEvents
		elapsed += stage.Duration
	}
	return elapsed
}

func stagesDuration(stages []RateStage) time.Duration {
	var duration time.Duration
	for _, stage := range stages {
		duration += stage.Duration
	}
	return duration
}

func stageConfigs(stages []RateStage) []RateStageConfig {
	configs := make([]RateStageConfig, len(stages))
	for index, stage := range stages {
		configs[index] = RateStageConfig{
			Rate:     stage.Rate,
			Duration: stage.Duration.String(),
		}
	}
	return configs
}

func summarizeLatencies(values []time.Duration) LatencySummary {
	if len(values) == 0 {
		return LatencySummary{}
	}
	sorted := append([]time.Duration(nil), values...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })

	var total time.Duration
	for _, value := range sorted {
		total += value
	}
	toMS := func(value time.Duration) float64 {
		return float64(value) / float64(time.Millisecond)
	}
	percentile := func(p float64) time.Duration {
		index := int(float64(len(sorted)-1) * p)
		return sorted[index]
	}
	return LatencySummary{
		AverageMS: toMS(total / time.Duration(len(sorted))),
		P50MS:     toMS(percentile(0.50)),
		P90MS:     toMS(percentile(0.90)),
		P95MS:     toMS(percentile(0.95)),
		P99MS:     toMS(percentile(0.99)),
		MaxMS:     toMS(sorted[len(sorted)-1]),
	}
}

func appendErrorSample(samples *[]string, value string) {
	if len(*samples) < maxDiagnosticSample {
		*samples = append(*samples, value)
	}
}

func currentEnvironment() Environment {
	environment := Environment{
		GoVersion: runtime.Version(),
		GOOS:      runtime.GOOS,
		GOARCH:    runtime.GOARCH,
		CPUs:      runtime.NumCPU(),
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				environment.GitRevision = setting.Value
			case "vcs.modified":
				environment.GitModified = setting.Value == "true"
			}
		}
	}
	return environment
}

func reportPasses(report *Report) bool {
	if report.Load.AttemptedEvents != report.Load.PlannedEvents ||
		report.Load.AcceptedEvents != report.Load.AttemptedEvents ||
		report.Load.RequestErrors != 0 ||
		report.Storage.MissingEvents != 0 ||
		report.Storage.DuplicateRows != 0 ||
		report.Storage.UnexpectedRows != 0 ||
		report.Storage.FieldMismatches != 0 ||
		report.Load.Reads.FailedRequests != 0 {
		return false
	}
	for _, check := range report.Aggregates {
		if !check.Passed {
			return false
		}
	}
	return true
}
