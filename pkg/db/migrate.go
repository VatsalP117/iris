package db

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

type migration struct {
	version int
	name    string
	file    string
}

var migrations = []migration{
	{version: 1, name: "v2_schema", file: "migrations/001_v2_schema.sql"},
	{version: 2, name: "local_day_sets", file: "migrations/002_local_day_sets.sql"},
}

func migrate(ctx context.Context, database *sql.DB) error {
	if _, err := database.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version       INTEGER PRIMARY KEY,
			name          TEXT NOT NULL,
			applied_at_us INTEGER NOT NULL
		)
	`); err != nil {
		return fmt.Errorf("create schema migrations table: %w", err)
	}

	applied := map[int]bool{}
	rows, err := database.QueryContext(ctx, "SELECT version FROM schema_migrations")
	if err != nil {
		return fmt.Errorf("read schema versions: %w", err)
	}
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			rows.Close()
			return fmt.Errorf("scan schema version: %w", err)
		}
		applied[version] = true
	}
	if err := rows.Close(); err != nil {
		return fmt.Errorf("close schema version rows: %w", err)
	}

	for _, item := range migrations {
		if applied[item.version] {
			continue
		}
		if err := applyMigration(ctx, database, item); err != nil {
			return err
		}
	}
	return nil
}

func applyMigration(ctx context.Context, database *sql.DB, item migration) error {
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin migration %d: %w", item.version, err)
	}
	defer tx.Rollback()

	legacy, err := isLegacyEventsTable(ctx, tx)
	if err != nil {
		return fmt.Errorf("inspect migration %d: %w", item.version, err)
	}
	if legacy {
		if _, err := tx.ExecContext(ctx, "ALTER TABLE events RENAME TO legacy_events"); err != nil {
			return fmt.Errorf("rename legacy events: %w", err)
		}
	}

	schema, err := migrationFiles.ReadFile(item.file)
	if err != nil {
		return fmt.Errorf("read migration %d: %w", item.version, err)
	}
	if _, err := tx.ExecContext(ctx, string(schema)); err != nil {
		return fmt.Errorf("apply migration %d: %w", item.version, err)
	}

	if legacy {
		if err := migrateLegacyEvents(ctx, tx); err != nil {
			return fmt.Errorf("migrate legacy events: %w", err)
		}
		if _, err := tx.ExecContext(ctx, "DROP TABLE legacy_events"); err != nil {
			return fmt.Errorf("drop legacy events: %w", err)
		}
	}

	if _, err := tx.ExecContext(
		ctx,
		"INSERT INTO schema_migrations(version, name, applied_at_us) VALUES (?, ?, ?)",
		item.version,
		item.name,
		time.Now().UTC().UnixMicro(),
	); err != nil {
		return fmt.Errorf("record migration %d: %w", item.version, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration %d: %w", item.version, err)
	}
	return nil
}

func isLegacyEventsTable(ctx context.Context, tx *sql.Tx) (bool, error) {
	var exists int
	if err := tx.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = 'events'
	`).Scan(&exists); err != nil || exists == 0 {
		return false, err
	}

	rows, err := tx.QueryContext(ctx, "PRAGMA table_info(events)")
	if err != nil {
		return false, err
	}
	defer rows.Close()
	for rows.Next() {
		var cid, notNull, primaryKey int
		var name, columnType string
		var defaultValue sql.NullString
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			return false, err
		}
		if name == "seq" {
			return false, nil
		}
	}
	return true, rows.Err()
}

type legacyEvent struct {
	id, name, rawURL, domain, referrer, siteID, sessionID, visitorID, properties string
	screenWidth                                                                  int
	timestamp                                                                    string
}

func migrateLegacyEvents(ctx context.Context, tx *sql.Tx) error {
	rows, err := tx.QueryContext(ctx, `
		SELECT id, COALESCE(event_name, ''), COALESCE(url, ''), COALESCE(domain, ''),
		       COALESCE(referrer, ''), COALESCE(screen_width, 0), COALESCE(site_id, ''),
		       COALESCE(session_id, ''), COALESCE(visitor_id, ''), COALESCE(properties, '{}'),
		       COALESCE(datetime(timestamp), CURRENT_TIMESTAMP)
		FROM legacy_events
		ORDER BY timestamp, id
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var events []legacyEvent
	for rows.Next() {
		var event legacyEvent
		if err := rows.Scan(
			&event.id, &event.name, &event.rawURL, &event.domain, &event.referrer,
			&event.screenWidth, &event.siteID, &event.sessionID, &event.visitorID,
			&event.properties, &event.timestamp,
		); err != nil {
			return err
		}
		if event.siteID == "" {
			event.siteID = event.domain
		}
		if event.siteID == "" {
			event.siteID = "legacy-unknown"
		}
		if event.name == "" {
			event.name = "$legacy"
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	siteDomains := map[string]map[string]struct{}{}
	for _, event := range events {
		if siteDomains[event.siteID] == nil {
			siteDomains[event.siteID] = map[string]struct{}{}
		}
		if event.domain != "" {
			siteDomains[event.siteID][strings.ToLower(event.domain)] = struct{}{}
		}
	}
	siteIDs := make([]string, 0, len(siteDomains))
	for siteID := range siteDomains {
		siteIDs = append(siteIDs, siteID)
	}
	sort.Strings(siteIDs)
	now := time.Now().UTC().UnixMicro()
	for _, siteID := range siteIDs {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO sites(id, name, timezone, retention_days, created_at_us)
			VALUES (?, ?, 'UTC', 365, ?)
		`, siteID, siteID, now); err != nil {
			return err
		}
		for domain := range siteDomains[siteID] {
			if _, err := tx.ExecContext(ctx, `
				INSERT OR IGNORE INTO site_domains(site_id, hostname, is_primary, created_at_us)
				VALUES (?, ?, 0, ?)
			`, siteID, domain, now); err != nil {
				return err
			}
		}
	}

	for _, event := range events {
		normalizedURL, pathname, normalizedReferrer, referrerHost := legacyURLParts(
			event.rawURL, event.referrer,
		)
		timestamp, err := time.ParseInLocation("2006-01-02 15:04:05", event.timestamp, time.UTC)
		if err != nil {
			timestamp = time.Now().UTC()
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO events(
				id, event_name, site_id, occurred_at_us, received_at_us, timestamp,
				url, domain, pathname, referrer, referrer_host, screen_width,
				session_id, visitor_id, properties, schema_version, sdk_version
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1, '')
		`,
			event.id, event.name, event.siteID, timestamp.UnixMicro(), timestamp.UnixMicro(), timestamp,
			normalizedURL, event.domain, pathname, normalizedReferrer, referrerHost, event.screenWidth,
			event.sessionID, event.visitorID, event.properties,
		); err != nil {
			return err
		}
	}
	return nil
}

func legacyURLParts(rawURL, rawReferrer string) (string, string, string, string) {
	normalizedURL, pathname, _ := sanitizeLegacyURL(rawURL)
	normalizedReferrer, _, referrerHost := sanitizeLegacyURL(rawReferrer)
	referrerHost = strings.TrimPrefix(referrerHost, "www.")
	return normalizedURL, pathname, normalizedReferrer, referrerHost
}

func sanitizeLegacyURL(raw string) (string, string, string) {
	if raw == "" {
		return "", "/", ""
	}
	parsed, err := url.Parse(raw)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") ||
		parsed.Hostname() == "" || parsed.User != nil {
		return "", "/", ""
	}
	parsed.Scheme = strings.ToLower(parsed.Scheme)
	parsed.Host = strings.ToLower(parsed.Host)
	parsed.RawQuery = ""
	parsed.ForceQuery = false
	parsed.Fragment = ""
	if parsed.Path == "" {
		parsed.Path = "/"
	}
	return parsed.String(), parsed.EscapedPath(), parsed.Hostname()
}
