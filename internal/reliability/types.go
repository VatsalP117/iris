package reliability

import (
	"fmt"
	"time"

	"github.com/VatsalP117/iris/pkg/core"
)

const (
	defaultDomain       = "lab.example"
	maxDiagnosticSample = 50
)

type Config struct {
	TargetURL      string
	DBPath         string
	RunID          string
	SiteID         string
	ReportDir      string
	Rate           int
	Duration       time.Duration
	EventCount     int
	BatchSize      int
	Workers        int
	RequestTimeout time.Duration
	ReadRate       int
	ReadWorkers    int
	Stages         []RateStage
	AllowNonLocal  bool
}

type RateStage struct {
	Rate     int
	Duration time.Duration
}

type RateStageConfig struct {
	Rate     int    `json:"rate"`
	Duration string `json:"duration"`
}

type PlannedEvent struct {
	Sequence int
	Event    core.Event
}

type RunConfig struct {
	TargetURL   string            `json:"target_url"`
	DBPath      string            `json:"db_path"`
	RunID       string            `json:"run_id"`
	SiteID      string            `json:"site_id"`
	Rate        int               `json:"rate_events_per_second"`
	Duration    string            `json:"duration"`
	EventCount  int               `json:"event_count"`
	BatchSize   int               `json:"batch_size"`
	Workers     int               `json:"workers"`
	ReadRate    int               `json:"read_requests_per_second"`
	ReadWorkers int               `json:"read_workers"`
	Stages      []RateStageConfig `json:"stages,omitempty"`
}

type LatencySummary struct {
	AverageMS float64 `json:"average_ms"`
	P50MS     float64 `json:"p50_ms"`
	P90MS     float64 `json:"p90_ms"`
	P95MS     float64 `json:"p95_ms"`
	P99MS     float64 `json:"p99_ms"`
	MaxMS     float64 `json:"max_ms"`
}

type LoadSummary struct {
	PlannedEvents          int            `json:"planned_events"`
	AttemptedEvents        int            `json:"attempted_events"`
	AcceptedEvents         int            `json:"accepted_events"`
	RejectedEvents         int            `json:"rejected_events"`
	RequestErrors          int            `json:"request_errors"`
	AttemptedRequests      int            `json:"attempted_requests"`
	AcceptedRequests       int            `json:"accepted_requests"`
	RejectedRequests       int            `json:"rejected_requests"`
	StatusCodes            map[int]int    `json:"status_codes"`
	ErrorSamples           []string       `json:"error_samples,omitempty"`
	ElapsedSeconds         float64        `json:"elapsed_seconds"`
	AchievedEventsPerSec   float64        `json:"achieved_events_per_second"`
	AchievedRequestsPerSec float64        `json:"achieved_requests_per_second"`
	MaxScheduleLagMS       float64        `json:"max_schedule_lag_ms"`
	Latency                LatencySummary `json:"latency"`
	Reads                  ReadSummary    `json:"reads"`
	Stages                 []StageSummary `json:"stages,omitempty"`
}

type StageSummary struct {
	Rate                   int            `json:"rate_events_per_second"`
	Duration               string         `json:"duration"`
	PlannedEvents          int            `json:"planned_events"`
	AttemptedEvents        int            `json:"attempted_events"`
	AcceptedEvents         int            `json:"accepted_events"`
	RejectedEvents         int            `json:"rejected_events"`
	RequestErrors          int            `json:"request_errors"`
	AttemptedRequests      int            `json:"attempted_requests"`
	StatusCodes            map[int]int    `json:"status_codes"`
	AchievedEventsPerSec   float64        `json:"achieved_events_per_second"`
	AchievedRequestsPerSec float64        `json:"achieved_requests_per_second"`
	MaxScheduleLagMS       float64        `json:"max_schedule_lag_ms"`
	Latency                LatencySummary `json:"latency"`
}

type EndpointSummary struct {
	Requests    int            `json:"requests"`
	Errors      int            `json:"errors"`
	StatusCodes map[int]int    `json:"status_codes"`
	Latency     LatencySummary `json:"latency"`
}

type ReadSummary struct {
	AttemptedRequests      int                        `json:"attempted_requests"`
	SuccessfulRequests     int                        `json:"successful_requests"`
	FailedRequests         int                        `json:"failed_requests"`
	AchievedRequestsPerSec float64                    `json:"achieved_requests_per_second"`
	ErrorSamples           []string                   `json:"error_samples,omitempty"`
	Latency                LatencySummary             `json:"latency"`
	Endpoints              map[string]EndpointSummary `json:"endpoints,omitempty"`
}

type StorageSummary struct {
	StoredRows           int   `json:"stored_rows"`
	UniqueSequences      int   `json:"unique_sequences"`
	MissingEvents        int   `json:"missing_events"`
	DuplicateRows        int   `json:"duplicate_rows"`
	UnexpectedRows       int   `json:"unexpected_rows"`
	FieldMismatches      int   `json:"field_mismatches"`
	MissingSamples       []int `json:"missing_samples,omitempty"`
	DuplicateSamples     []int `json:"duplicate_samples,omitempty"`
	UnexpectedSamples    []int `json:"unexpected_samples,omitempty"`
	FieldMismatchSamples []int `json:"field_mismatch_samples,omitempty"`
	DatabaseBytes        int64 `json:"database_bytes"`
}

type AggregateCheck struct {
	Name     string `json:"name"`
	Expected any    `json:"expected,omitempty"`
	Actual   any    `json:"actual,omitempty"`
	Passed   bool   `json:"passed"`
	Error    string `json:"error,omitempty"`
}

type Environment struct {
	GitRevision string `json:"git_revision,omitempty"`
	GitModified bool   `json:"git_modified"`
	GoVersion   string `json:"go_version"`
	GOOS        string `json:"goos"`
	GOARCH      string `json:"goarch"`
	CPUs        int    `json:"cpus"`
}

type Report struct {
	GeneratedAt time.Time        `json:"generated_at"`
	Config      RunConfig        `json:"config"`
	Environment Environment      `json:"environment"`
	Load        LoadSummary      `json:"load"`
	Storage     StorageSummary   `json:"storage"`
	Aggregates  []AggregateCheck `json:"aggregates"`
	Resources   ResourceSummary  `json:"resources"`
	Passed      bool             `json:"passed"`
}

type ResourceSummary struct {
	Samples             int     `json:"samples"`
	AverageCPUPercent   float64 `json:"average_cpu_percent"`
	PeakCPUPercent      float64 `json:"peak_cpu_percent"`
	AverageRSSBytes     int64   `json:"average_rss_bytes"`
	PeakRSSBytes        int64   `json:"peak_rss_bytes"`
	ProcessReadBytes    int64   `json:"process_read_bytes,omitempty"`
	ProcessWriteBytes   int64   `json:"process_write_bytes,omitempty"`
	DatabaseStartBytes  int64   `json:"database_start_bytes"`
	DatabaseEndBytes    int64   `json:"database_end_bytes"`
	DatabaseGrowthBytes int64   `json:"database_growth_bytes"`
	PeakWALBytes        int64   `json:"peak_wal_bytes"`
}

func (c Config) plannedEventCount() (int, error) {
	if c.EventCount > 0 {
		return c.EventCount, nil
	}
	if len(c.Stages) > 0 {
		count := 0
		for index, stage := range c.Stages {
			if stage.Rate <= 0 {
				return 0, fmt.Errorf("stage %d rate must be greater than zero", index)
			}
			if stage.Duration <= 0 {
				return 0, fmt.Errorf("stage %d duration must be greater than zero", index)
			}
			count += int(float64(stage.Rate) * stage.Duration.Seconds())
		}
		if count == 0 {
			return 0, fmt.Errorf("stages produce zero events")
		}
		return count, nil
	}
	if c.Rate <= 0 {
		return 0, fmt.Errorf("rate must be greater than zero")
	}
	if c.Duration <= 0 {
		return 0, fmt.Errorf("duration must be greater than zero")
	}
	count := int(float64(c.Rate) * c.Duration.Seconds())
	if count == 0 {
		return 0, fmt.Errorf("rate and duration produce zero events")
	}
	return count, nil
}
