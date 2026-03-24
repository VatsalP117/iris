package db

import (
	"context"
	"fmt"
	"path/filepath"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	"github.com/VatsalP117/iris/pkg/core"
)

var testEventSeq atomic.Int64

func TestGetStatsAndSitesUseSiteIDAcrossDomains(t *testing.T) {
	repo := newTestRepo(t)

	insertEvent(t, repo, core.Event{
		EventName:   "$pageview",
		URL:         "https://example.com/",
		Domain:      "example.com",
		SiteID:      "site-a",
		SessionID:   "s1",
		VisitorID:   "v1",
		ScreenWidth: 390,
	})
	insertEvent(t, repo, core.Event{
		EventName:   "$pageview",
		URL:         "https://www.example.com/pricing",
		Domain:      "www.example.com",
		SiteID:      "site-a",
		SessionID:   "s2",
		VisitorID:   "v2",
		ScreenWidth: 1440,
	})
	insertEvent(t, repo, core.Event{
		EventName:   "$pageview",
		URL:         "https://other.com/",
		Domain:      "other.com",
		SiteID:      "site-b",
		SessionID:   "s3",
		VisitorID:   "v3",
		ScreenWidth: 1280,
	})

	stats, err := repo.GetStats(context.Background(), "site-a", "", "")
	if err != nil {
		t.Fatalf("GetStats returned error: %v", err)
	}
	if stats.Pageviews != 2 || stats.UniqueVisitors != 2 || stats.Sessions != 2 {
		t.Fatalf("unexpected stats for site-a: %+v", stats)
	}

	sites, err := repo.GetSites(context.Background())
	if err != nil {
		t.Fatalf("GetSites returned error: %v", err)
	}

	if len(sites) != 2 {
		t.Fatalf("expected 2 sites, got %d", len(sites))
	}
	if sites[0].SiteID != "site-a" {
		t.Fatalf("expected first site to be site-a, got %+v", sites[0])
	}

	expectedDomains := []string{"example.com", "www.example.com"}
	if !reflect.DeepEqual(sites[0].Domains, expectedDomains) {
		t.Fatalf("unexpected grouped domains: got %v want %v", sites[0].Domains, expectedDomains)
	}
}

func TestGetDevicesCountsOnlyPageviews(t *testing.T) {
	repo := newTestRepo(t)

	insertEvent(t, repo, core.Event{
		EventName:   "$pageview",
		Domain:      "example.com",
		SiteID:      "site-a",
		SessionID:   "s1",
		VisitorID:   "v1",
		ScreenWidth: 390,
	})
	insertEvent(t, repo, core.Event{
		EventName:   "$pageview",
		Domain:      "example.com",
		SiteID:      "site-a",
		SessionID:   "s2",
		VisitorID:   "v2",
		ScreenWidth: 1440,
	})
	insertEvent(t, repo, core.Event{
		EventName:   "$click",
		Domain:      "example.com",
		SiteID:      "site-a",
		SessionID:   "s2",
		VisitorID:   "v2",
		ScreenWidth: 390,
	})
	insertEvent(t, repo, core.Event{
		EventName:   "$web_vital",
		Domain:      "example.com",
		SiteID:      "site-a",
		SessionID:   "s2",
		VisitorID:   "v2",
		ScreenWidth: 1440,
		Properties:  map[string]any{"$name": "LCP", "$val": 2400.0},
	})

	devices, err := repo.GetDevices(context.Background(), "site-a", "", "")
	if err != nil {
		t.Fatalf("GetDevices returned error: %v", err)
	}

	got := map[string]int{}
	for _, device := range devices {
		got[device.Device] = device.Count
	}

	expected := map[string]int{"Mobile": 1, "Desktop": 1}
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("unexpected device counts: got %v want %v", got, expected)
	}
}

func TestGetTopReferrersNormalizesHosts(t *testing.T) {
	repo := newTestRepo(t)

	insertEvent(t, repo, core.Event{
		EventName: "$pageview",
		Domain:    "example.com",
		SiteID:    "site-a",
		SessionID: "s1",
		VisitorID: "v1",
		Referrer:  "https://google.com/search?q=pricing",
	})
	insertEvent(t, repo, core.Event{
		EventName: "$pageview",
		Domain:    "www.example.com",
		SiteID:    "site-a",
		SessionID: "s1",
		VisitorID: "v1",
		Referrer:  "https://www.google.com/maps",
	})
	insertEvent(t, repo, core.Event{
		EventName: "$pageview",
		Domain:    "example.com",
		SiteID:    "site-a",
		SessionID: "s2",
		VisitorID: "v2",
		Referrer:  "https://google.com/search?q=blog",
	})
	insertEvent(t, repo, core.Event{
		EventName: "$pageview",
		Domain:    "example.com",
		SiteID:    "site-a",
		SessionID: "s3",
		VisitorID: "v3",
		Referrer:  "https://www.Bing.com/search?q=iris",
	})

	referrers, err := repo.GetTopReferrers(context.Background(), "site-a", "", "", 10)
	if err != nil {
		t.Fatalf("GetTopReferrers returned error: %v", err)
	}

	if len(referrers) != 2 {
		t.Fatalf("expected 2 normalized referrers, got %d: %+v", len(referrers), referrers)
	}
	if referrers[0].Referrer != "google.com" || referrers[0].Visitors != 2 {
		t.Fatalf("unexpected top referrer row: %+v", referrers[0])
	}
	if referrers[1].Referrer != "bing.com" || referrers[1].Visitors != 1 {
		t.Fatalf("unexpected second referrer row: %+v", referrers[1])
	}
}

func TestGetVitalsUsesP75(t *testing.T) {
	repo := newTestRepo(t)

	for _, value := range []float64{1000, 2000, 3000, 4000} {
		insertEvent(t, repo, core.Event{
			EventName:  "$web_vital",
			Domain:     "example.com",
			SiteID:     "site-a",
			SessionID:  "lcp",
			VisitorID:  "v-lcp",
			Properties: map[string]any{"$name": "LCP", "$val": value},
		})
	}
	for _, value := range []float64{0.05, 0.10, 0.20, 0.30} {
		insertEvent(t, repo, core.Event{
			EventName:  "$web_vital",
			Domain:     "www.example.com",
			SiteID:     "site-a",
			SessionID:  "cls",
			VisitorID:  "v-cls",
			Properties: map[string]any{"$name": "CLS", "$val": value},
		})
	}

	vitals, err := repo.GetVitals(context.Background(), "site-a", "", "")
	if err != nil {
		t.Fatalf("GetVitals returned error: %v", err)
	}

	got := map[string]float64{}
	for _, vital := range vitals {
		got[vital.Name] = vital.Value
	}

	if got["LCP"] != 3000 {
		t.Fatalf("expected LCP p75 to be 3000, got %v", got["LCP"])
	}
	if got["CLS"] != 0.20 {
		t.Fatalf("expected CLS p75 to be 0.20, got %v", got["CLS"])
	}
}

func TestGetStatsSupportsDateTimeAndDateWindows(t *testing.T) {
	repo := newTestRepo(t)

	base := time.Date(2026, 3, 25, 15, 0, 0, 0, time.UTC)
	insertEvent(t, repo, core.Event{
		EventName:   "$pageview",
		Domain:      "example.com",
		SiteID:      "site-a",
		SessionID:   "s-older",
		VisitorID:   "v-older",
		ScreenWidth: 1280,
		Timestamp:   base.Add(-25 * time.Hour),
	})
	insertEvent(t, repo, core.Event{
		EventName:   "$pageview",
		Domain:      "example.com",
		SiteID:      "site-a",
		SessionID:   "s-recent-1",
		VisitorID:   "v-recent-1",
		ScreenWidth: 1440,
		Timestamp:   base.Add(-23 * time.Hour),
	})
	insertEvent(t, repo, core.Event{
		EventName:   "$pageview",
		Domain:      "example.com",
		SiteID:      "site-a",
		SessionID:   "s-recent-2",
		VisitorID:   "v-recent-2",
		ScreenWidth: 390,
		Timestamp:   base.Add(-1 * time.Hour),
	})

	const sqlLayout = "2006-01-02 15:04:05"
	stats24h, err := repo.GetStats(
		context.Background(),
		"site-a",
		base.Add(-24*time.Hour).Format(sqlLayout),
		base.Format(sqlLayout),
	)
	if err != nil {
		t.Fatalf("GetStats(24h) returned error: %v", err)
	}
	if stats24h.Pageviews != 2 || stats24h.UniqueVisitors != 2 || stats24h.Sessions != 2 {
		t.Fatalf("unexpected 24h stats: %+v", stats24h)
	}

	statsDay, err := repo.GetStats(context.Background(), "site-a", "2026-03-24", "2026-03-24")
	if err != nil {
		t.Fatalf("GetStats(day window) returned error: %v", err)
	}
	if statsDay.Pageviews != 2 {
		t.Fatalf("expected 2 pageviews on 2026-03-24, got %+v", statsDay)
	}
}

func newTestRepo(t *testing.T) *SqliteRepository {
	t.Helper()

	repo, err := NewSqliteDB(filepath.Join(t.TempDir(), "iris.db"))
	if err != nil {
		t.Fatalf("NewSqliteDB returned error: %v", err)
	}
	t.Cleanup(func() {
		_ = repo.Close()
	})

	return repo
}

func insertEvent(t *testing.T, repo *SqliteRepository, event core.Event) {
	t.Helper()

	seq := testEventSeq.Add(1)
	event.ID = fmt.Sprintf("event-%d", seq)
	if event.URL == "" {
		event.URL = "https://example.com/"
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Unix(seq, 0).UTC()
	}

	if err := repo.Insert(context.Background(), &event); err != nil {
		t.Fatalf("Insert returned error: %v", err)
	}
}
