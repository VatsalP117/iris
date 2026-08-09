package reliability_test

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/VatsalP117/iris/internal/reliability"
	"github.com/VatsalP117/iris/pkg/api"
	"github.com/VatsalP117/iris/pkg/core"
	"github.com/VatsalP117/iris/pkg/db"
)

func TestBuildManifestIsDeterministicAndRepresentative(t *testing.T) {
	first := reliability.BuildManifest("run-a", "site-a", 100)
	second := reliability.BuildManifest("run-a", "site-a", 100)

	if len(first) != 100 || len(second) != 100 {
		t.Fatalf("unexpected manifest lengths: %d and %d", len(first), len(second))
	}

	counts := map[string]int{}
	for i := range first {
		if first[i].Sequence != i {
			t.Fatalf("sequence %d stored at index %d", first[i].Sequence, i)
		}
		if first[i].Event.EventName != second[i].Event.EventName ||
			first[i].Event.URL != second[i].Event.URL ||
			first[i].Event.SessionID != second[i].Event.SessionID ||
			first[i].Event.VisitorID != second[i].Event.VisitorID {
			t.Fatalf("manifest differs at sequence %d", i)
		}
		counts[first[i].Event.EventName]++
	}

	expected := map[string]int{
		"$pageview":  60,
		"$click":     20,
		"signup":     10,
		"$web_vital": 10,
	}
	for name, count := range expected {
		if counts[name] != count {
			t.Fatalf("unexpected count for %s: got %d want %d", name, counts[name], count)
		}
	}
}

func TestRunReconcilesSingleAndBatchIngestion(t *testing.T) {
	originalLogWriter := log.Writer()
	log.SetOutput(io.Discard)
	t.Cleanup(func() { log.SetOutput(originalLogWriter) })

	for _, batchSize := range []int{1, 10} {
		t.Run("batch-"+strconv.Itoa(batchSize), func(t *testing.T) {
			dbPath := filepath.Join(t.TempDir(), "iris.db")
			repo, err := db.NewSqliteDB(dbPath)
			if err != nil {
				t.Fatalf("NewSqliteDB returned error: %v", err)
			}
			t.Cleanup(func() { _ = repo.Close() })
			registerLabSite(t, repo, "integration-site-"+strconv.Itoa(batchSize))

			handler := api.NewHandler(repo)
			mux := http.NewServeMux()
			mux.HandleFunc("/api/event", handler.TrackEvent)
			mux.HandleFunc("/api/events", handler.TrackBatchEvents)
			mux.HandleFunc("/api/stats", handler.GetStats)
			mux.HandleFunc("/api/pages", handler.GetPages)
			mux.HandleFunc("/api/referrers", handler.GetReferrers)
			mux.HandleFunc("/api/vitals", handler.GetVitals)
			mux.HandleFunc("/api/devices", handler.GetDevices)
			mux.HandleFunc("/api/timeseries", handler.GetTimeSeries)
			mux.HandleFunc("/api/timeseries/visitors", handler.GetUniqueVisitorsTimeSeries)
			mux.HandleFunc("/api/timeseries/sessions", handler.GetSessionsTimeSeries)

			server := httptest.NewServer(mux)
			t.Cleanup(server.Close)

			report, err := reliability.Run(context.Background(), reliability.Config{
				TargetURL:      server.URL,
				DBPath:         dbPath,
				RunID:          "integration-" + strconv.Itoa(batchSize),
				SiteID:         "integration-site-" + strconv.Itoa(batchSize),
				Rate:           10_000,
				EventCount:     100,
				BatchSize:      batchSize,
				Workers:        8,
				RequestTimeout: 2 * time.Second,
			})
			if err != nil {
				t.Fatalf("Run returned error: %v", err)
			}
			if !report.Passed {
				t.Fatalf("expected report to pass: %+v", report)
			}
			if report.Load.AcceptedEvents != 100 {
				t.Fatalf("accepted events: got %d want 100", report.Load.AcceptedEvents)
			}
			if report.Storage.StoredRows != 100 ||
				report.Storage.MissingEvents != 0 ||
				report.Storage.DuplicateRows != 0 ||
				report.Storage.UnexpectedRows != 0 ||
				report.Storage.FieldMismatches != 0 {
				t.Fatalf("unexpected storage summary: %+v", report.Storage)
			}
			for _, check := range report.Aggregates {
				if !check.Passed {
					t.Fatalf("aggregate %s failed: %+v", check.Name, check)
				}
			}

			reportDir := filepath.Join(t.TempDir(), "report")
			writtenDir, err := reliability.WriteReport(report, reportDir)
			if err != nil {
				t.Fatalf("WriteReport returned error: %v", err)
			}
			if writtenDir != reportDir {
				t.Fatalf("report directory: got %q want %q", writtenDir, reportDir)
			}
			for _, name := range []string{"report.md", "summary.json"} {
				if _, err := os.Stat(filepath.Join(reportDir, name)); err != nil {
					t.Fatalf("expected %s: %v", name, err)
				}
			}
		})
	}
}

func TestRunMeasuresConcurrentAnalyticsReads(t *testing.T) {
	originalLogWriter := log.Writer()
	log.SetOutput(io.Discard)
	t.Cleanup(func() { log.SetOutput(originalLogWriter) })

	dbPath := filepath.Join(t.TempDir(), "iris.db")
	repo, err := db.NewSqliteDB(dbPath)
	if err != nil {
		t.Fatalf("NewSqliteDB returned error: %v", err)
	}
	t.Cleanup(func() { _ = repo.Close() })
	registerLabSite(t, repo, "mixed-read-write")

	handler := api.NewHandler(repo)
	mux := http.NewServeMux()
	mux.HandleFunc("/api/event", handler.TrackEvent)
	mux.HandleFunc("/api/events", handler.TrackBatchEvents)
	mux.HandleFunc("/api/stats", handler.GetStats)
	mux.HandleFunc("/api/pages", handler.GetPages)
	mux.HandleFunc("/api/referrers", handler.GetReferrers)
	mux.HandleFunc("/api/vitals", handler.GetVitals)
	mux.HandleFunc("/api/devices", handler.GetDevices)
	mux.HandleFunc("/api/timeseries", handler.GetTimeSeries)
	mux.HandleFunc("/api/timeseries/visitors", handler.GetUniqueVisitorsTimeSeries)
	mux.HandleFunc("/api/timeseries/sessions", handler.GetSessionsTimeSeries)

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	report, err := reliability.Run(context.Background(), reliability.Config{
		TargetURL:      server.URL,
		DBPath:         dbPath,
		RunID:          "mixed-read-write",
		SiteID:         "mixed-read-write",
		Rate:           200,
		EventCount:     200,
		BatchSize:      10,
		Workers:        8,
		ReadRate:       20,
		ReadWorkers:    4,
		RequestTimeout: 2 * time.Second,
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if !report.Passed {
		t.Fatalf("expected report to pass: %+v", report)
	}
	if report.Load.Reads.AttemptedRequests == 0 {
		t.Fatal("expected concurrent read requests")
	}
	if report.Load.Reads.FailedRequests != 0 {
		t.Fatalf("unexpected read failures: %+v", report.Load.Reads)
	}
}

func TestRunSupportsDeterministicRateStages(t *testing.T) {
	originalLogWriter := log.Writer()
	log.SetOutput(io.Discard)
	t.Cleanup(func() { log.SetOutput(originalLogWriter) })

	dbPath := filepath.Join(t.TempDir(), "iris.db")
	repo, err := db.NewSqliteDB(dbPath)
	if err != nil {
		t.Fatalf("NewSqliteDB returned error: %v", err)
	}
	t.Cleanup(func() { _ = repo.Close() })
	registerLabSite(t, repo, "staged-load")

	handler := api.NewHandler(repo)
	mux := http.NewServeMux()
	mux.HandleFunc("/api/event", handler.TrackEvent)
	mux.HandleFunc("/api/events", handler.TrackBatchEvents)
	mux.HandleFunc("/api/stats", handler.GetStats)
	mux.HandleFunc("/api/pages", handler.GetPages)
	mux.HandleFunc("/api/referrers", handler.GetReferrers)
	mux.HandleFunc("/api/vitals", handler.GetVitals)
	mux.HandleFunc("/api/devices", handler.GetDevices)
	mux.HandleFunc("/api/timeseries", handler.GetTimeSeries)
	mux.HandleFunc("/api/timeseries/visitors", handler.GetUniqueVisitorsTimeSeries)
	mux.HandleFunc("/api/timeseries/sessions", handler.GetSessionsTimeSeries)

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	report, err := reliability.Run(context.Background(), reliability.Config{
		TargetURL: server.URL,
		DBPath:    dbPath,
		RunID:     "staged-load",
		SiteID:    "staged-load",
		BatchSize: 10,
		Workers:   8,
		Stages: []reliability.RateStage{
			{Rate: 500, Duration: 100 * time.Millisecond},
			{Rate: 1000, Duration: 100 * time.Millisecond},
		},
		RequestTimeout: 2 * time.Second,
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if !report.Passed {
		t.Fatalf("expected staged run to pass: %+v", report)
	}
	if report.Load.PlannedEvents != 150 || report.Load.AcceptedEvents != 150 {
		t.Fatalf(
			"staged event counts: planned=%d accepted=%d want 150",
			report.Load.PlannedEvents,
			report.Load.AcceptedEvents,
		)
	}
	if len(report.Config.Stages) != 2 {
		t.Fatalf("reported stages: got %d want 2", len(report.Config.Stages))
	}
	if len(report.Load.Stages) != 2 {
		t.Fatalf("stage summaries: got %d want 2", len(report.Load.Stages))
	}
	if report.Load.Stages[0].AcceptedEvents != 50 ||
		report.Load.Stages[1].AcceptedEvents != 100 {
		t.Fatalf("unexpected stage summaries: %+v", report.Load.Stages)
	}
}

func registerLabSite(t *testing.T, repo *db.SqliteRepository, siteID string) {
	t.Helper()
	if err := repo.CreateSite(context.Background(), &core.Site{
		ID: siteID, Name: siteID, Domains: []string{domainForSiteForTest(siteID)},
	}); err != nil {
		t.Fatalf("CreateSite returned error: %v", err)
	}
}

func domainForSiteForTest(siteID string) string {
	manifest := reliability.BuildManifest("domain-fixture", siteID, 1)
	return manifest[0].Event.Domain
}

func TestCompareReportsDetectsRegressionsAndConfigMismatch(t *testing.T) {
	base := reliability.Report{
		Passed: true,
		Config: reliability.RunConfig{
			Rate:       500,
			EventCount: 1000,
			BatchSize:  10,
		},
		Load: reliability.LoadSummary{
			AcceptedEvents:       1000,
			AchievedEventsPerSec: 500,
			Latency: reliability.LatencySummary{
				P95MS: 10,
				P99MS: 20,
			},
		},
		Resources: reliability.ResourceSummary{
			PeakCPUPercent:      50,
			PeakRSSBytes:        1000,
			DatabaseGrowthBytes: 10_000,
		},
	}
	candidate := base
	candidate.Load.Latency.P95MS = 13
	candidate.Resources.PeakRSSBytes = 1300

	directory := t.TempDir()
	baselinePath := filepath.Join(directory, "baseline.json")
	candidatePath := filepath.Join(directory, "candidate.json")
	writeJSONReport(t, baselinePath, base)
	writeJSONReport(t, candidatePath, candidate)

	comparison, err := reliability.CompareReports(baselinePath, candidatePath)
	if err != nil {
		t.Fatalf("CompareReports returned error: %v", err)
	}
	if comparison.Passed || comparison.Regressions != 2 {
		t.Fatalf("expected two regressions: %+v", comparison)
	}

	candidate.Config.EventCount++
	writeJSONReport(t, candidatePath, candidate)
	comparison, err = reliability.CompareReports(baselinePath, candidatePath)
	if err != nil {
		t.Fatalf("CompareReports returned error: %v", err)
	}
	if comparison.Comparable || comparison.Passed {
		t.Fatalf("expected configuration mismatch: %+v", comparison)
	}

	candidate.Config = base.Config
	candidate.Environment.GOOS = "different-os"
	writeJSONReport(t, candidatePath, candidate)
	comparison, err = reliability.CompareReports(baselinePath, candidatePath)
	if err != nil {
		t.Fatalf("CompareReports returned error: %v", err)
	}
	if comparison.Comparable || comparison.Error != "execution environments differ" {
		t.Fatalf("expected environment mismatch: %+v", comparison)
	}
}

func writeJSONReport(t *testing.T, path string, report reliability.Report) {
	t.Helper()
	data, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("os.WriteFile returned error: %v", err)
	}
}

func TestRunDetectsAcceptedEventsMissingFromStorage(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "iris.db")
	repo, err := db.NewSqliteDB(dbPath)
	if err != nil {
		t.Fatalf("NewSqliteDB returned error: %v", err)
	}
	if err := repo.Close(); err != nil {
		t.Fatalf("Close returned error: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusAccepted)
			return
		}
		http.NotFound(w, r)
	}))
	t.Cleanup(server.Close)

	report, err := reliability.Run(context.Background(), reliability.Config{
		TargetURL:      server.URL,
		DBPath:         dbPath,
		RunID:          "missing-storage",
		SiteID:         "missing-storage-site",
		Rate:           10_000,
		EventCount:     12,
		BatchSize:      1,
		Workers:        2,
		RequestTimeout: time.Second,
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if report.Passed {
		t.Fatal("expected report to fail")
	}
	if report.Storage.MissingEvents != 12 {
		t.Fatalf("missing events: got %d want 12", report.Storage.MissingEvents)
	}
}

func TestRunRejectsNonLocalTargetByDefault(t *testing.T) {
	_, err := reliability.Run(context.Background(), reliability.Config{
		TargetURL:  "https://analytics.example.com",
		DBPath:     "/tmp/unused.db",
		EventCount: 1,
		Rate:       1,
	})
	if err == nil {
		t.Fatal("expected non-local target error")
	}
}
