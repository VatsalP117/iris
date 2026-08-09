package db

import (
	"context"
	"fmt"
	"time"
)

// ApplyRetention removes raw and projected data outside each site's configured
// retention window. It is safe to run repeatedly.
func (r *SqliteRepository) ApplyRetention(ctx context.Context, now time.Time) (int64, error) {
	now = now.UTC()
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, retention_days FROM sites
	`)
	if err != nil {
		return 0, err
	}
	type policy struct {
		siteID string
		cutoff time.Time
	}
	var policies []policy
	for rows.Next() {
		var siteID string
		var days int
		if err := rows.Scan(&siteID, &days); err != nil {
			rows.Close()
			return 0, err
		}
		policies = append(policies, policy{siteID: siteID, cutoff: now.AddDate(0, 0, -days)})
	}
	if err := rows.Close(); err != nil {
		return 0, err
	}

	tx, err := r.writer.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var deletedEvents int64
	for _, item := range policies {
		boundaryRows, err := tx.QueryContext(ctx, `
			SELECT session_id FROM sessions
			WHERE site_id = ? AND started_at_us < ? AND ended_at_us >= ?
		`, item.siteID, item.cutoff.UnixMicro(), item.cutoff.UnixMicro())
		if err != nil {
			return 0, fmt.Errorf("read boundary sessions for site %s: %w", item.siteID, err)
		}
		var boundarySessions []string
		for boundaryRows.Next() {
			var sessionID string
			if err := boundaryRows.Scan(&sessionID); err != nil {
				boundaryRows.Close()
				return 0, err
			}
			boundarySessions = append(boundarySessions, sessionID)
		}
		if err := boundaryRows.Close(); err != nil {
			return 0, err
		}

		result, err := tx.ExecContext(ctx, `
			DELETE FROM events WHERE site_id = ? AND occurred_at_us < ?
		`, item.siteID, item.cutoff.UnixMicro())
		if err != nil {
			return 0, fmt.Errorf("delete expired events for site %s: %w", item.siteID, err)
		}
		count, err := result.RowsAffected()
		if err != nil {
			return 0, err
		}
		deletedEvents += count

		cutoffDay := item.cutoff.Format("2006-01-02")
		deletions := []struct {
			statement string
			cutoff    any
		}{
			{"DELETE FROM sessions WHERE site_id = ? AND ended_at_us < ?", item.cutoff.UnixMicro()},
			{"DELETE FROM daily_site_metrics WHERE site_id = ? AND day < ?", cutoffDay},
			{"DELETE FROM daily_page_metrics WHERE site_id = ? AND day < ?", cutoffDay},
			{"DELETE FROM daily_referrer_visitors WHERE site_id = ? AND day < ?", cutoffDay},
			{"DELETE FROM daily_visitors WHERE site_id = ? AND day < ?", cutoffDay},
			{"DELETE FROM daily_sessions WHERE site_id = ? AND day < ?", cutoffDay},
		}
		for _, deletion := range deletions {
			if _, err := tx.ExecContext(ctx, deletion.statement, item.siteID, deletion.cutoff); err != nil {
				return 0, fmt.Errorf("delete expired projections for site %s: %w", item.siteID, err)
			}
		}
		var throughSeq int64
		if err := tx.QueryRowContext(ctx, "SELECT COALESCE(MAX(seq), 0) FROM events").Scan(&throughSeq); err != nil {
			return 0, err
		}
		for _, sessionID := range boundarySessions {
			if err := rebuildSession(ctx, tx, item.siteID, sessionID, throughSeq); err != nil {
				return 0, fmt.Errorf("rebuild retained session %s: %w", sessionID, err)
			}
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return deletedEvents, nil
}
