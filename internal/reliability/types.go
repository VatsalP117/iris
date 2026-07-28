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
	AllowNonLocal  bool
}

type PlannedEvent struct {
	Sequence int
	Event    core.Event
}

type RunConfig struct {
	TargetURL  string `json:"target_url"`
	DBPath     string `json:"db_path"`
	RunID      string `json:"run_id"`
	SiteID     string `json:"site_id"`
	Rate       int    `json:"rate_events_per_second"`
	Duration   string `json:"duration"`
	EventCount int    `json:"event_count"`
	BatchSize  int    `json:"batch_size"`
	Workers    int    `json:"workers"`
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
	Passed      bool             `json:"passed"`
}

func (c Config) plannedEventCount() (int, error) {
	if c.EventCount > 0 {
		return c.EventCount, nil
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
