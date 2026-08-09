package api

import (
	"bytes"
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/VatsalP117/iris/pkg/core"
	"github.com/VatsalP117/iris/pkg/db"
)

func TestTrackEvent_DuplicateClientIDIsIdempotent(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "iris.db")
	repo, err := db.NewSqliteDB(databasePath)
	if err != nil {
		t.Fatalf("NewSqliteDB returned error: %v", err)
	}
	t.Cleanup(func() {
		_ = repo.Close()
	})
	if err := repo.CreateSite(context.Background(), &core.Site{
		ID: "site-a", Name: "Site A", Domains: []string{"example.com"},
	}); err != nil {
		t.Fatalf("CreateSite returned error: %v", err)
	}

	handler := NewHandler(repo)
	body := []byte(`{
		"id": "client-event-1",
		"n": "$pageview",
		"u": "https://example.com/",
		"d": "example.com",
		"w": 1280,
		"s": "site-a",
		"sid": "session-1",
		"vid": "visitor-1"
	}`)

	for attempt := 0; attempt < 2; attempt++ {
		request := httptest.NewRequest(http.MethodPost, "/api/event", bytes.NewReader(body))
		response := httptest.NewRecorder()
		handler.TrackEvent(response, request)
		if response.Code != http.StatusAccepted {
			t.Fatalf("attempt %d returned status %d", attempt+1, response.Code)
		}
	}

	stats, err := repo.GetStats(context.Background(), "site-a", "", "")
	if err != nil {
		t.Fatalf("GetStats returned error: %v", err)
	}
	if stats.Pageviews != 1 {
		t.Fatalf("duplicate event produced %d pageviews, want 1", stats.Pageviews)
	}
}

func TestTrackEvent_RejectsUnknownSiteAndDisallowedDomain(t *testing.T) {
	repo, err := db.NewSqliteDB(filepath.Join(t.TempDir(), "iris.db"))
	if err != nil {
		t.Fatalf("NewSqliteDB returned error: %v", err)
	}
	t.Cleanup(func() { _ = repo.Close() })
	if err := repo.CreateSite(context.Background(), &core.Site{
		ID: "site-a", Name: "Site A", Domains: []string{"example.com"},
	}); err != nil {
		t.Fatalf("CreateSite returned error: %v", err)
	}
	handler := NewHandler(repo)

	tests := []struct {
		name   string
		body   string
		status int
	}{
		{
			name:   "unknown site",
			body:   `{"id":"event-1","n":"$pageview","u":"https://example.com/","s":"missing","sid":"s","vid":"v"}`,
			status: http.StatusNotFound,
		},
		{
			name:   "disallowed domain",
			body:   `{"id":"event-2","n":"$pageview","u":"https://evil.example/","s":"site-a","sid":"s","vid":"v"}`,
			status: http.StatusForbidden,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/api/event", strings.NewReader(test.body))
			response := httptest.NewRecorder()
			handler.TrackEvent(response, request)
			if response.Code != test.status {
				t.Fatalf("status = %d, want %d; body=%s", response.Code, test.status, response.Body.String())
			}
		})
	}
}

func TestTrackEvent_NormalizesURLsAndPreservesOccurrenceTime(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "iris.db")
	repo, err := db.NewSqliteDB(databasePath)
	if err != nil {
		t.Fatalf("NewSqliteDB returned error: %v", err)
	}
	t.Cleanup(func() { _ = repo.Close() })
	if err := repo.CreateSite(context.Background(), &core.Site{
		ID: "site-a", Name: "Site A", Domains: []string{"example.com"},
	}); err != nil {
		t.Fatalf("CreateSite returned error: %v", err)
	}

	handler := NewHandler(repo)
	body := `{
		"id":"event-normalized","n":"$pageview",
		"u":"https://example.com/pricing?token=secret#plans",
		"d":"example.com","r":"https://www.google.com/search?q=iris",
		"w":1440,"s":"site-a","sid":"session-1","vid":"visitor-1",
		"ts":"2026-08-01T12:00:00Z","v":1,"sv":"1.0.0"
	}`
	request := httptest.NewRequest(http.MethodPost, "/api/event", strings.NewReader(body))
	response := httptest.NewRecorder()
	handler.TrackEvent(response, request)
	if response.Code != http.StatusAccepted {
		t.Fatalf("status = %d; body=%s", response.Code, response.Body.String())
	}

	database, err := sql.Open("sqlite3", databasePath)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer database.Close()
	var trackedURL, pathname, referrer, referrerHost string
	var occurredAt, receivedAt int64
	if err := database.QueryRow(`
		SELECT url, pathname, referrer, referrer_host, occurred_at_us, received_at_us
		FROM events WHERE id = 'event-normalized'
	`).Scan(&trackedURL, &pathname, &referrer, &referrerHost, &occurredAt, &receivedAt); err != nil {
		t.Fatalf("read event: %v", err)
	}
	if trackedURL != "https://example.com/pricing" || pathname != "/pricing" {
		t.Fatalf("unexpected tracked URL fields: %q %q", trackedURL, pathname)
	}
	if referrer != "https://www.google.com/search" || referrerHost != "google.com" {
		t.Fatalf("unexpected referrer fields: %q %q", referrer, referrerHost)
	}
	if occurredAt >= receivedAt {
		t.Fatalf("occurrence time %d should precede receive time %d", occurredAt, receivedAt)
	}
}

func TestSites_CreatesRegisteredSite(t *testing.T) {
	repo, err := db.NewSqliteDB(filepath.Join(t.TempDir(), "iris.db"))
	if err != nil {
		t.Fatalf("NewSqliteDB returned error: %v", err)
	}
	t.Cleanup(func() { _ = repo.Close() })
	handler := NewHandler(repo)

	request := httptest.NewRequest(http.MethodPost, "/api/sites", strings.NewReader(
		`{"site_id":"docs","name":"Documentation","timezone":"Asia/Kolkata","retention_days":90,"domains":["docs.example.com"]}`,
	))
	response := httptest.NewRecorder()
	handler.Sites(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d; body=%s", response.Code, response.Body.String())
	}

	sites, err := repo.GetSites(context.Background())
	if err != nil {
		t.Fatalf("GetSites returned error: %v", err)
	}
	if len(sites) != 1 || sites[0].SiteID != "docs" || sites[0].Timezone != "Asia/Kolkata" {
		t.Fatalf("unexpected sites: %+v", sites)
	}
}
