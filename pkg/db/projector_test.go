package db

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/VatsalP117/iris/pkg/core"
)

func TestProjectPending_ProjectsOrderedEventsExactlyOnce(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()
	base := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)

	insertProjectionEvent(t, repo, "projection-1", core.Event{
		EventName: "$pageview", SiteID: "site-a", SessionID: "session-a",
		VisitorID: "visitor-a", Pathname: "/exit", ReferrerHost: "search.example",
		Timestamp: base,
	})
	insertProjectionEvent(t, repo, "projection-2", core.Event{
		EventName: "purchase", SiteID: "site-a", SessionID: "session-a",
		VisitorID: "visitor-a", Pathname: "/checkout", Timestamp: base.Add(-30 * time.Minute),
	})
	// This deliberately arrives last with the earliest occurrence time. Session
	// boundaries and entry/exit pages must follow occurrence time, not seq.
	insertProjectionEvent(t, repo, "projection-3", core.Event{
		EventName: "$pageview", SiteID: "site-a", SessionID: "session-a",
		VisitorID: "visitor-a", Pathname: "/entry", ReferrerHost: "first.example",
		Timestamp: base.Add(-2 * time.Hour),
	})
	insertProjectionEvent(t, repo, "projection-4", core.Event{
		EventName: "$pageview", SiteID: "site-a", SessionID: "session-b",
		VisitorID: "visitor-a", Pathname: "/entry", ReferrerHost: "first.example",
		Timestamp: base.Add(-time.Hour),
	})

	count, err := repo.ProjectPending(ctx, 2)
	if err != nil {
		t.Fatalf("ProjectPending returned error: %v", err)
	}
	if count != 2 {
		t.Fatalf("ProjectPending count = %d, want 2", count)
	}
	count, err = repo.ProjectPending(ctx, 10)
	if err != nil {
		t.Fatalf("ProjectPending second batch returned error: %v", err)
	}
	if count != 2 {
		t.Fatalf("ProjectPending second count = %d, want 2", count)
	}

	var startedAt, endedAt int64
	var entry, exit, referrer string
	var pageviews, eventCount, bounce, version int
	if err := repo.db.QueryRowContext(ctx, `
		SELECT started_at_us, ended_at_us, entry_pathname, exit_pathname,
		       referrer_host, pageviews, event_count, is_bounce, projection_version
		FROM sessions WHERE site_id = 'site-a' AND session_id = 'session-a'
	`).Scan(&startedAt, &endedAt, &entry, &exit, &referrer, &pageviews, &eventCount, &bounce, &version); err != nil {
		t.Fatalf("read projected session: %v", err)
	}
	if startedAt != base.Add(-2*time.Hour).UnixMicro() || endedAt != base.UnixMicro() {
		t.Fatalf("session bounds = (%d, %d), want (%d, %d)",
			startedAt, endedAt, base.Add(-2*time.Hour).UnixMicro(), base.UnixMicro())
	}
	if entry != "/entry" || exit != "/exit" || referrer != "first.example" {
		t.Fatalf("session navigation = entry %q, exit %q, referrer %q", entry, exit, referrer)
	}
	if pageviews != 2 || eventCount != 3 || bounce != 0 || version != analyticsProjectionVersion {
		t.Fatalf("unexpected session counts: pageviews=%d events=%d bounce=%d version=%d",
			pageviews, eventCount, bounce, version)
	}

	assertDailySiteMetrics(t, repo, "site-a", "2026-08-04", 3, 1)
	var entryPageviews int
	if err := repo.db.QueryRowContext(ctx, `
		SELECT pageviews FROM daily_page_metrics
		WHERE site_id = 'site-a' AND day = '2026-08-04' AND pathname = '/entry'
	`).Scan(&entryPageviews); err != nil {
		t.Fatalf("read daily page metrics: %v", err)
	}
	if entryPageviews != 2 {
		t.Fatalf("entry pageviews = %d, want 2", entryPageviews)
	}
	var referrerVisitors int
	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM daily_referrer_visitors
		WHERE site_id = 'site-a' AND day = '2026-08-04' AND referrer_host = 'first.example'
	`).Scan(&referrerVisitors); err != nil {
		t.Fatalf("read referrer visitors: %v", err)
	}
	if referrerVisitors != 1 {
		t.Fatalf("distinct referrer visitors = %d, want 1", referrerVisitors)
	}

	count, err = repo.ProjectPending(ctx, 10)
	if err != nil || count != 0 {
		t.Fatalf("idempotent ProjectPending = (%d, %v), want (0, nil)", count, err)
	}
	assertDailySiteMetrics(t, repo, "site-a", "2026-08-04", 3, 1)

	var lastSeq int64
	if err := repo.db.QueryRowContext(ctx, `
		SELECT last_seq FROM projection_checkpoints WHERE name = ?
	`, analyticsProjectionName).Scan(&lastSeq); err != nil {
		t.Fatalf("read projection checkpoint: %v", err)
	}
	if lastSeq != 4 {
		t.Fatalf("last_seq = %d, want 4", lastSeq)
	}
}

func TestProjectPending_UsesSiteTimezoneForDailyMetrics(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()
	if err := repo.CreateSite(ctx, &core.Site{
		ID: "site-west", Name: "West", Timezone: "America/Los_Angeles",
		Domains: []string{"west.example"},
	}); err != nil {
		t.Fatalf("CreateSite returned error: %v", err)
	}
	insertProjectionEvent(t, repo, "timezone-1", core.Event{
		EventName: "$pageview", SiteID: "site-west", SessionID: "session-west",
		VisitorID: "visitor-west", Pathname: "/", Timestamp: time.Date(2026, 8, 5, 1, 0, 0, 0, time.UTC),
	})

	if count, err := repo.ProjectPending(ctx, 10); err != nil || count != 1 {
		t.Fatalf("ProjectPending = (%d, %v), want (1, nil)", count, err)
	}
	assertDailySiteMetrics(t, repo, "site-west", "2026-08-04", 1, 0)

	pageviews, err := repo.GetPageviewsTimeSeries(ctx, "site-west", "2026-08-04", "2026-08-04")
	if err != nil {
		t.Fatalf("GetPageviewsTimeSeries returned error: %v", err)
	}
	visitors, err := repo.GetUniqueVisitorsTimeSeries(ctx, "site-west", "2026-08-04", "2026-08-04")
	if err != nil {
		t.Fatalf("GetUniqueVisitorsTimeSeries returned error: %v", err)
	}
	sessions, err := repo.GetSessionsTimeSeries(ctx, "site-west", "2026-08-04", "2026-08-04")
	if err != nil {
		t.Fatalf("GetSessionsTimeSeries returned error: %v", err)
	}
	if len(pageviews) != 1 || pageviews[0].Date != "2026-08-04" || pageviews[0].Pageviews != 1 {
		t.Fatalf("unexpected projected pageviews: %+v", pageviews)
	}
	if len(visitors) != 1 || visitors[0].UniqueVisitors != 1 {
		t.Fatalf("unexpected projected visitors: %+v", visitors)
	}
	if len(sessions) != 1 || sessions[0].Sessions != 1 {
		t.Fatalf("unexpected projected sessions: %+v", sessions)
	}
}

func TestGetSystemStatus_ReportsProjectionLag(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()
	insertProjectionEvent(t, repo, "status-1", core.Event{
		EventName: "$pageview", SiteID: "site-a", SessionID: "session-status",
		VisitorID: "visitor-status", Pathname: "/", Timestamp: time.Now().UTC(),
	})
	status, err := repo.GetSystemStatus(ctx)
	if err != nil {
		t.Fatalf("GetSystemStatus returned error: %v", err)
	}
	if status.Database != "ok" || status.ProjectionLag != 1 {
		t.Fatalf("unexpected status before projection: %+v", status)
	}
	if _, err := repo.ProjectPending(ctx, 10); err != nil {
		t.Fatalf("ProjectPending returned error: %v", err)
	}
	status, err = repo.GetSystemStatus(ctx)
	if err != nil {
		t.Fatalf("GetSystemStatus after projection returned error: %v", err)
	}
	if status.ProjectionLag != 0 || status.EventLastSeq != status.ProjectionLastSeq {
		t.Fatalf("unexpected status after projection: %+v", status)
	}
}

func TestProjectPending_RollsBackBatchOnFailure(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()
	insertProjectionEvent(t, repo, "rollback-1", core.Event{
		EventName: "$pageview", SiteID: "site-a", SessionID: "session-a",
		VisitorID: "visitor-a", Pathname: "/", Timestamp: time.Date(2026, 8, 5, 1, 0, 0, 0, time.UTC),
	})
	if _, err := repo.db.ExecContext(ctx, `
		CREATE TRIGGER fail_page_projection
		BEFORE INSERT ON daily_page_metrics
		BEGIN
			SELECT RAISE(ABORT, 'forced projection failure');
		END
	`); err != nil {
		t.Fatalf("create failure trigger: %v", err)
	}

	if _, err := repo.ProjectPending(ctx, 10); err == nil {
		t.Fatal("ProjectPending returned nil error for forced projection failure")
	}
	for _, table := range []string{"sessions", "daily_site_metrics", "daily_page_metrics"} {
		var count int
		if err := repo.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+table).Scan(&count); err != nil {
			t.Fatalf("count %s: %v", table, err)
		}
		if count != 0 {
			t.Fatalf("%s contains %d rows after rollback", table, count)
		}
	}
	var checkpoint int64
	if err := repo.db.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(last_seq), 0) FROM projection_checkpoints
	`).Scan(&checkpoint); err != nil {
		t.Fatalf("read checkpoint: %v", err)
	}
	if checkpoint != 0 {
		t.Fatalf("checkpoint = %d after rollback, want 0", checkpoint)
	}
}

func TestRebuildProjections_ReplacesDerivedState(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()
	base := time.Date(2026, 8, 6, 10, 0, 0, 0, time.UTC)
	insertProjectionEvent(t, repo, "rebuild-1", core.Event{
		EventName: "$pageview", SiteID: "site-a", SessionID: "session-rebuild",
		VisitorID: "visitor-rebuild", Pathname: "/one", Timestamp: base,
	})
	insertProjectionEvent(t, repo, "rebuild-2", core.Event{
		EventName: "$pageview", SiteID: "site-a", SessionID: "session-rebuild",
		VisitorID: "visitor-rebuild", Pathname: "/two", Timestamp: base.Add(time.Minute),
	})
	if _, err := repo.ProjectPending(ctx, 10); err != nil {
		t.Fatalf("ProjectPending returned error: %v", err)
	}
	if _, err := repo.db.ExecContext(ctx, `
		UPDATE daily_site_metrics SET pageviews = 999;
		UPDATE sessions SET pageviews = 999, is_bounce = 1;
	`); err != nil {
		t.Fatalf("corrupt derived state: %v", err)
	}

	if err := repo.RebuildProjections(ctx); err != nil {
		t.Fatalf("RebuildProjections returned error: %v", err)
	}
	assertDailySiteMetrics(t, repo, "site-a", "2026-08-06", 2, 0)
	var pageviews, bounce int
	if err := repo.db.QueryRowContext(ctx, `
		SELECT pageviews, is_bounce FROM sessions
		WHERE site_id = 'site-a' AND session_id = 'session-rebuild'
	`).Scan(&pageviews, &bounce); err != nil {
		t.Fatalf("read rebuilt session: %v", err)
	}
	if pageviews != 2 || bounce != 0 {
		t.Fatalf("rebuilt session = pageviews %d, bounce %d; want 2, 0", pageviews, bounce)
	}
}

func TestProjectPending_RejectsStaleProjectionVersion(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()
	if _, err := repo.db.ExecContext(ctx, `
		INSERT INTO projection_checkpoints(name, last_seq, version, updated_at_us)
		VALUES (?, 0, ?, 0)
	`, analyticsProjectionName, analyticsProjectionVersion+1); err != nil {
		t.Fatalf("insert stale checkpoint: %v", err)
	}

	_, err := repo.ProjectPending(ctx, 10)
	if !errors.Is(err, ErrProjectionVersionMismatch) {
		t.Fatalf("ProjectPending error = %v, want ErrProjectionVersionMismatch", err)
	}
}

func insertProjectionEvent(t *testing.T, repo *SqliteRepository, id string, event core.Event) {
	t.Helper()
	event.ID = id
	if event.URL == "" {
		event.URL = "https://example.com" + event.Pathname
	}
	if event.Domain == "" {
		event.Domain = "example.com"
	}
	if err := repo.Insert(context.Background(), &event); err != nil {
		t.Fatalf("Insert(%s) returned error: %v", id, err)
	}
}

func assertDailySiteMetrics(
	t *testing.T,
	repo *SqliteRepository,
	siteID, day string,
	wantPageviews, wantCustomEvents int,
) {
	t.Helper()
	var pageviews, customEvents int
	if err := repo.db.QueryRowContext(context.Background(), `
		SELECT pageviews, custom_events FROM daily_site_metrics
		WHERE site_id = ? AND day = ?
	`, siteID, day).Scan(&pageviews, &customEvents); err != nil {
		t.Fatalf("read daily site metrics for %s/%s: %v", siteID, day, err)
	}
	if pageviews != wantPageviews || customEvents != wantCustomEvents {
		t.Fatalf("daily site metrics for %s/%s = (%d, %d), want (%d, %d)",
			siteID, day, pageviews, customEvents, wantPageviews, wantCustomEvents)
	}
}

func TestProjectPending_ValidatesBatchSize(t *testing.T) {
	repo := newTestRepo(t)
	for _, batchSize := range []int{0, -1} {
		t.Run(fmt.Sprint(batchSize), func(t *testing.T) {
			if _, err := repo.ProjectPending(context.Background(), batchSize); err == nil {
				t.Fatalf("ProjectPending(%d) returned nil error", batchSize)
			}
		})
	}
}
