package db

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/VatsalP117/iris/pkg/core"
)

func (r *SqliteRepository) CreateSite(ctx context.Context, site *core.Site) error {
	if site == nil {
		return fmt.Errorf("site is required")
	}
	siteID := strings.TrimSpace(site.ID)
	if siteID == "" {
		return fmt.Errorf("site id is required")
	}
	name := strings.TrimSpace(site.Name)
	if name == "" {
		name = siteID
	}
	timezone := strings.TrimSpace(site.Timezone)
	if timezone == "" {
		timezone = "UTC"
	}
	if _, err := time.LoadLocation(timezone); err != nil {
		return fmt.Errorf("invalid timezone %q: %w", timezone, err)
	}
	retentionDays := site.RetentionDays
	if retentionDays <= 0 {
		retentionDays = 365
	}

	domains := normalizedDomains(site.Domains)
	now := time.Now().UTC().UnixMicro()
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO sites(id, name, timezone, retention_days, created_at_us)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			timezone = excluded.timezone,
			retention_days = excluded.retention_days
	`, siteID, name, timezone, retentionDays, now); err != nil {
		return err
	}
	for index, domain := range domains {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO site_domains(site_id, hostname, is_primary, created_at_us)
			VALUES (?, ?, ?, ?)
			ON CONFLICT(site_id, hostname) DO UPDATE SET is_primary = excluded.is_primary
		`, siteID, domain, boolToInt(index == 0), now); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func normalizedDomains(domains []string) []string {
	seen := map[string]struct{}{}
	result := make([]string, 0, len(domains))
	for _, domain := range domains {
		domain = strings.TrimSpace(strings.ToLower(domain))
		domain = strings.TrimSuffix(domain, ".")
		if domain == "" {
			continue
		}
		if _, ok := seen[domain]; ok {
			continue
		}
		seen[domain] = struct{}{}
		result = append(result, domain)
	}
	sort.Strings(result)
	return result
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
