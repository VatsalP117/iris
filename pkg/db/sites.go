package db

import (
	"context"
	"database/sql"
	"fmt"
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

	domains, err := normalizedDomains(site.Domains)
	if err != nil {
		return err
	}
	if len(domains) == 0 {
		return fmt.Errorf("at least one domain is required")
	}
	now := time.Now().UTC().UnixMicro()
	tx, err := r.writer.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var existingTimezone string
	var hasEvents int
	err = tx.QueryRowContext(ctx, `
		SELECT timezone, EXISTS(SELECT 1 FROM events WHERE site_id = sites.id LIMIT 1)
		FROM sites WHERE id = ?
	`, siteID).Scan(&existingTimezone, &hasEvents)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if err == nil && hasEvents == 1 && existingTimezone != timezone {
		return fmt.Errorf("%w: %s uses %s", core.ErrTimezoneImmutable, siteID, existingTimezone)
	}

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
	if _, err := tx.ExecContext(ctx, "DELETE FROM site_domains WHERE site_id = ?", siteID); err != nil {
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

func (r *SqliteRepository) ValidateSite(ctx context.Context, siteID, domain string) error {
	siteID = strings.TrimSpace(siteID)
	domain = strings.TrimSuffix(strings.ToLower(strings.TrimSpace(domain)), ".")
	var exists int
	err := r.db.QueryRowContext(ctx, `
		SELECT 1 FROM sites WHERE id = ? AND disabled_at_us IS NULL
	`, siteID).Scan(&exists)
	if err == sql.ErrNoRows {
		return fmt.Errorf("%w: %s", core.ErrSiteNotFound, siteID)
	}
	if err != nil {
		return err
	}
	err = r.db.QueryRowContext(ctx, `
		SELECT 1 FROM site_domains WHERE site_id = ? AND hostname = ?
	`, siteID, domain).Scan(&exists)
	if err == sql.ErrNoRows {
		return fmt.Errorf("%w: %s", core.ErrDomainNotAllowed, domain)
	}
	return err
}

func normalizedDomains(domains []string) ([]string, error) {
	seen := map[string]struct{}{}
	result := make([]string, 0, len(domains))
	for _, domain := range domains {
		domain = strings.TrimSpace(strings.ToLower(domain))
		domain = strings.TrimSuffix(domain, ".")
		if domain == "" {
			continue
		}
		if err := validateHostname(domain); err != nil {
			return nil, err
		}
		if _, ok := seen[domain]; ok {
			continue
		}
		seen[domain] = struct{}{}
		result = append(result, domain)
	}
	return result, nil
}

func validateHostname(hostname string) error {
	if len(hostname) > 253 || strings.ContainsAny(hostname, " /:@*?#\t\r\n") {
		return fmt.Errorf("invalid hostname %q", hostname)
	}
	for _, label := range strings.Split(hostname, ".") {
		if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return fmt.Errorf("invalid hostname %q", hostname)
		}
		for _, character := range label {
			if (character < 'a' || character > 'z') &&
				(character < '0' || character > '9') && character != '-' {
				return fmt.Errorf("invalid hostname %q", hostname)
			}
		}
	}
	return nil
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
