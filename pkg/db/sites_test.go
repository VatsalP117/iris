package db

import (
	"context"
	"errors"
	"testing"

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
