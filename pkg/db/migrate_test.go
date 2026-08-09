package db

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func TestNewSqliteDB_CreatesVersionedV2Schema(t *testing.T) {
	repo := newTestRepo(t)

	for _, table := range []string{
		"schema_migrations",
		"sites",
		"site_domains",
		"ingest_keys",
		"events",
		"sessions",
		"daily_site_metrics",
		"daily_page_metrics",
		"daily_referrer_visitors",
		"daily_visitors",
		"daily_sessions",
		"projection_checkpoints",
	} {
		var found string
		if err := repo.db.QueryRow(
			"SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?", table,
		).Scan(&found); err != nil {
			t.Fatalf("expected table %q: %v", table, err)
		}
	}

	var version int
	if err := repo.db.QueryRow("SELECT MAX(version) FROM schema_migrations").Scan(&version); err != nil {
		t.Fatalf("read schema version: %v", err)
	}
	if version != 2 {
		t.Fatalf("schema version = %d, want 2", version)
	}
}

func TestNewSqliteDB_MigratesLegacyEvents(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "legacy.db")
	legacy, err := sql.Open("sqlite3", databasePath)
	if err != nil {
		t.Fatalf("open legacy database: %v", err)
	}
	_, err = legacy.Exec(`
		CREATE TABLE events (
			id TEXT PRIMARY KEY, event_name TEXT, url TEXT, domain TEXT,
			referrer TEXT, screen_width INTEGER, site_id TEXT, session_id TEXT,
			visitor_id TEXT, properties TEXT, timestamp DATETIME
		);
		INSERT INTO events VALUES (
			'legacy-event', '$pageview', 'https://example.com/pricing?secret=1',
			'example.com', 'https://google.com/search?q=iris', 1440, 'site-a',
			'session-a', 'visitor-a', '{}', '2026-08-01 12:00:00'
		);
		INSERT INTO events VALUES (
			'legacy-null-time', '$pageview', 'https://example.com/null?secret=1',
			'example.com', '', 1440, 'site-a', 'session-b', 'visitor-b', '{}', NULL
		);
	`)
	if err != nil {
		t.Fatalf("create legacy schema: %v", err)
	}
	if err := legacy.Close(); err != nil {
		t.Fatalf("close legacy database: %v", err)
	}

	repo, err := NewSqliteDB(databasePath)
	if err != nil {
		t.Fatalf("NewSqliteDB returned error: %v", err)
	}
	t.Cleanup(func() { _ = repo.Close() })

	stats, err := repo.GetStats(context.Background(), "site-a", "", "")
	if err != nil {
		t.Fatalf("GetStats returned error: %v", err)
	}
	if stats.Pageviews != 2 {
		t.Fatalf("migrated pageviews = %d, want 2", stats.Pageviews)
	}

	var trackedURL, pathname, referrer, referrerHost string
	var occurredAt int64
	if err := repo.db.QueryRow(`
		SELECT url, pathname, referrer, referrer_host, occurred_at_us
		FROM events WHERE id = 'legacy-event'
	`).Scan(&trackedURL, &pathname, &referrer, &referrerHost, &occurredAt); err != nil {
		t.Fatalf("read migrated event: %v", err)
	}
	if trackedURL != "https://example.com/pricing" || pathname != "/pricing" ||
		referrer != "https://google.com/search" || referrerHost != "google.com" {
		t.Fatalf("unexpected normalized fields: url=%q pathname=%q referrer=%q host=%q",
			trackedURL, pathname, referrer, referrerHost)
	}
	if occurredAt != time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC).UnixMicro() {
		t.Fatalf("unexpected occurred_at_us: %d", occurredAt)
	}
	if err := repo.db.QueryRow(`
		SELECT occurred_at_us FROM events WHERE id = 'legacy-null-time'
	`).Scan(&occurredAt); err != nil {
		t.Fatalf("read null-time migrated event: %v", err)
	}
	if occurredAt <= 0 {
		t.Fatalf("null legacy timestamp did not receive fallback: %d", occurredAt)
	}
}
