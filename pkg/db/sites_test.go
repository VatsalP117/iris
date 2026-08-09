package db

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/VatsalP117/iris/pkg/core"
)

func TestCreateSite_ReplacesDomainAllowlist(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()
	site := core.Site{
		ID: "site-a", Name: "Updated Site", Timezone: "UTC", RetentionDays: 30,
		Domains: []string{"new.example.com"},
	}
	if err := repo.CreateSite(ctx, &site); err != nil {
		t.Fatalf("CreateSite returned error: %v", err)
	}
	if err := repo.ValidateSite(ctx, "site-a", "new.example.com"); err != nil {
		t.Fatalf("ValidateSite(new domain) returned error: %v", err)
	}
	if err := repo.ValidateSite(ctx, "site-a", "example.com"); !errors.Is(err, core.ErrDomainNotAllowed) {
		t.Fatalf("ValidateSite(old domain) error = %v, want ErrDomainNotAllowed", err)
	}
	sites, err := repo.GetSites(ctx)
	if err != nil {
		t.Fatalf("GetSites returned error: %v", err)
	}
	if sites[0].Domain != "new.example.com" || sites[0].RetentionDays != 30 {
		t.Fatalf("unexpected updated site: %+v", sites[0])
	}
}

func TestCreateSite_PreventsTimezoneChangeAfterIngestion(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()
	event := core.Event{
		ID: "timezone-event", EventName: "$pageview", SiteID: "site-a",
		URL: "https://example.com/", Domain: "example.com", Pathname: "/",
		SessionID: "session", VisitorID: "visitor", Timestamp: time.Now().UTC(),
	}
	if err := repo.Insert(ctx, &event); err != nil {
		t.Fatalf("Insert returned error: %v", err)
	}
	err := repo.CreateSite(ctx, &core.Site{
		ID: "site-a", Name: "Site A", Timezone: "Asia/Kolkata",
		RetentionDays: 365, Domains: []string{"example.com"},
	})
	if !errors.Is(err, core.ErrTimezoneImmutable) {
		t.Fatalf("CreateSite error = %v, want ErrTimezoneImmutable", err)
	}
}

func TestCreateSite_RejectsMalformedDomains(t *testing.T) {
	repo := newTestRepo(t)
	for _, domain := range []string{
		"https://example.com", "example.com:443", "*.example.com", "-bad.example", "bad_.example",
	} {
		t.Run(domain, func(t *testing.T) {
			err := repo.CreateSite(context.Background(), &core.Site{
				ID: "invalid-site", Domains: []string{domain},
			})
			if err == nil {
				t.Fatalf("CreateSite accepted malformed domain %q", domain)
			}
		})
	}
}
