package db

import (
	"context"
	"testing"
	"time"

	"github.com/VatsalP117/iris/pkg/core"
)

func TestApplyRetention_RemovesOnlyExpiredSiteData(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()
	if err := repo.CreateSite(ctx, &core.Site{
		ID: "short-retention", Name: "Short", RetentionDays: 7, Domains: []string{"short.example"},
	}); err != nil {
		t.Fatalf("CreateSite returned error: %v", err)
	}
	now := time.Date(2026, 8, 9, 12, 0, 0, 0, time.UTC)
	for _, event := range []core.Event{
		{ID: "expired", EventName: "$pageview", SiteID: "short-retention", URL: "https://short.example/old", Domain: "short.example", Pathname: "/old", SessionID: "old-session", VisitorID: "old-visitor", Timestamp: now.AddDate(0, 0, -8)},
		{ID: "retained", EventName: "$pageview", SiteID: "short-retention", URL: "https://short.example/new", Domain: "short.example", Pathname: "/new", SessionID: "new-session", VisitorID: "new-visitor", Timestamp: now.AddDate(0, 0, -6)},
	} {
		event := event
		if err := repo.Insert(ctx, &event); err != nil {
			t.Fatalf("Insert returned error: %v", err)
		}
	}

	deleted, err := repo.ApplyRetention(ctx, now)
	if err != nil {
		t.Fatalf("ApplyRetention returned error: %v", err)
	}
	if deleted != 1 {
		t.Fatalf("deleted events = %d, want 1", deleted)
	}
	var remaining string
	if err := repo.db.QueryRow(`SELECT id FROM events WHERE site_id = 'short-retention'`).Scan(&remaining); err != nil {
		t.Fatalf("read retained event: %v", err)
	}
	if remaining != "retained" {
		t.Fatalf("remaining event = %q, want retained", remaining)
	}
}
