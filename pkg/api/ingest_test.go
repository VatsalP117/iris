package api

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/VatsalP117/iris/pkg/db"
)

func TestTrackEvent_DuplicateClientIDIsIdempotent(t *testing.T) {
	repo, err := db.NewSqliteDB(filepath.Join(t.TempDir(), "iris.db"))
	if err != nil {
		t.Fatalf("NewSqliteDB returned error: %v", err)
	}
	t.Cleanup(func() {
		_ = repo.Close()
	})

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
