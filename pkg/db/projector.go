package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	analyticsProjectionName    = "analytics"
	analyticsProjectionVersion = 1
	defaultProjectionBatchSize = 1000
)

var ErrProjectionVersionMismatch = errors.New("projection version mismatch")

type projectionEvent struct {
	seq          int64
	siteID       string
	eventName    string
	occurredAtUS int64
	pathname     string
	referrerHost string
	sessionID    string
	visitorID    string
	localDay     string
}

type projectionSessionKey struct {
	siteID    string
	sessionID string
}

// ProjectPending applies at most batchSize raw events to the rebuildable
// analytics tables. The derived writes and checkpoint advance are committed
// together, so retrying after an error cannot count an event twice.
func (r *SqliteRepository) ProjectPending(ctx context.Context, batchSize int) (int, error) {
	if batchSize <= 0 {
		return 0, fmt.Errorf("projection batch size must be positive")
	}

	tx, err := r.writer.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin projection: %w", err)
	}
	defer tx.Rollback()

	checkpoint, err := projectionCheckpoint(ctx, tx)
	if err != nil {
		return 0, err
	}
	events, err := pendingProjectionEvents(ctx, tx, checkpoint, batchSize)
	if err != nil {
		return 0, err
	}
	if len(events) == 0 {
		if err := tx.Commit(); err != nil {
			return 0, fmt.Errorf("commit empty projection: %w", err)
		}
		return 0, nil
	}

	affectedSessions := make(map[projectionSessionKey]struct{})
	for _, event := range events {
		if err := projectDailyEvent(ctx, tx, event, event.localDay); err != nil {
			return 0, fmt.Errorf("project event %d: %w", event.seq, err)
		}
		if event.sessionID != "" {
			affectedSessions[projectionSessionKey{
				siteID:    event.siteID,
				sessionID: event.sessionID,
			}] = struct{}{}
		}
	}

	lastSeq := events[len(events)-1].seq
	for key := range affectedSessions {
		if err := rebuildSession(ctx, tx, key.siteID, key.sessionID, lastSeq); err != nil {
			return 0, err
		}
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE projection_checkpoints
		SET last_seq = ?, updated_at_us = ?
		WHERE name = ? AND version = ?
	`, lastSeq, time.Now().UTC().UnixMicro(), analyticsProjectionName, analyticsProjectionVersion); err != nil {
		return 0, fmt.Errorf("advance projection checkpoint: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit projection: %w", err)
	}
	return len(events), nil
}

// RebuildProjections discards all derived analytics state and deterministically
// replays the immutable event log using the current projection version.
func (r *SqliteRepository) RebuildProjections(ctx context.Context) error {
	tx, err := r.writer.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin projection reset: %w", err)
	}
	defer tx.Rollback()

	for _, table := range []string{
		"sessions",
		"daily_site_metrics",
		"daily_page_metrics",
		"daily_referrer_visitors",
		"daily_visitors",
		"daily_sessions",
	} {
		if _, err := tx.ExecContext(ctx, "DELETE FROM "+table); err != nil {
			return fmt.Errorf("clear %s: %w", table, err)
		}
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO projection_checkpoints(name, last_seq, version, updated_at_us)
		VALUES (?, 0, ?, ?)
		ON CONFLICT(name) DO UPDATE SET
			last_seq = 0,
			version = excluded.version,
			updated_at_us = excluded.updated_at_us
	`, analyticsProjectionName, analyticsProjectionVersion, time.Now().UTC().UnixMicro()); err != nil {
		return fmt.Errorf("reset projection checkpoint: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit projection reset: %w", err)
	}

	for {
		count, err := r.ProjectPending(ctx, defaultProjectionBatchSize)
		if err != nil {
			return err
		}
		if count < defaultProjectionBatchSize {
			return nil
		}
	}
}

func projectionCheckpoint(ctx context.Context, tx *sql.Tx) (int64, error) {
	now := time.Now().UTC().UnixMicro()
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO projection_checkpoints(name, last_seq, version, updated_at_us)
		VALUES (?, 0, ?, ?)
		ON CONFLICT(name) DO NOTHING
	`, analyticsProjectionName, analyticsProjectionVersion, now); err != nil {
		return 0, fmt.Errorf("initialize projection checkpoint: %w", err)
	}

	var lastSeq int64
	var version int
	if err := tx.QueryRowContext(ctx, `
		SELECT last_seq, version FROM projection_checkpoints WHERE name = ?
	`, analyticsProjectionName).Scan(&lastSeq, &version); err != nil {
		return 0, fmt.Errorf("read projection checkpoint: %w", err)
	}
	if version != analyticsProjectionVersion {
		return 0, fmt.Errorf("%w: database has %d, projector requires %d",
			ErrProjectionVersionMismatch, version, analyticsProjectionVersion)
	}
	return lastSeq, nil
}

func pendingProjectionEvents(
	ctx context.Context,
	tx *sql.Tx,
	checkpoint int64,
	batchSize int,
) ([]projectionEvent, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT e.seq, e.site_id, e.event_name, e.occurred_at_us, e.pathname,
		       e.referrer_host, e.session_id, e.visitor_id, e.local_day
		FROM events e
		WHERE e.seq > ?
		ORDER BY e.seq
		LIMIT ?
	`, checkpoint, batchSize)
	if err != nil {
		return nil, fmt.Errorf("read pending projection events: %w", err)
	}
	defer rows.Close()

	events := make([]projectionEvent, 0, batchSize)
	for rows.Next() {
		var event projectionEvent
		if err := rows.Scan(
			&event.seq,
			&event.siteID,
			&event.eventName,
			&event.occurredAtUS,
			&event.pathname,
			&event.referrerHost,
			&event.sessionID,
			&event.visitorID,
			&event.localDay,
		); err != nil {
			return nil, fmt.Errorf("scan pending projection event: %w", err)
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate pending projection events: %w", err)
	}
	return events, nil
}

func projectDailyEvent(ctx context.Context, tx *sql.Tx, event projectionEvent, day string) error {
	pageview := event.eventName == "$pageview"
	customEvent := event.eventName != "" && !strings.HasPrefix(event.eventName, "$")
	if pageview || customEvent {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO daily_site_metrics(site_id, day, pageviews, custom_events)
			VALUES (?, ?, ?, ?)
			ON CONFLICT(site_id, day) DO UPDATE SET
				pageviews = pageviews + excluded.pageviews,
				custom_events = custom_events + excluded.custom_events
		`, event.siteID, day, boolToInt(pageview), boolToInt(customEvent)); err != nil {
			return fmt.Errorf("update daily site metrics: %w", err)
		}
	}
	if !pageview {
		return nil
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO daily_page_metrics(site_id, day, pathname, pageviews)
		VALUES (?, ?, ?, 1)
		ON CONFLICT(site_id, day, pathname) DO UPDATE SET
			pageviews = pageviews + 1
	`, event.siteID, day, event.pathname); err != nil {
		return fmt.Errorf("update daily page metrics: %w", err)
	}
	if event.referrerHost != "" && event.visitorID != "" {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO daily_referrer_visitors(site_id, day, referrer_host, visitor_id)
			VALUES (?, ?, ?, ?)
			ON CONFLICT(site_id, day, referrer_host, visitor_id) DO NOTHING
		`, event.siteID, day, event.referrerHost, event.visitorID); err != nil {
			return fmt.Errorf("update daily referrer visitors: %w", err)
		}
	}
	if event.visitorID != "" {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO daily_visitors(site_id, day, visitor_id)
			VALUES (?, ?, ?)
			ON CONFLICT(site_id, day, visitor_id) DO NOTHING
		`, event.siteID, day, event.visitorID); err != nil {
			return fmt.Errorf("update daily visitors: %w", err)
		}
	}
	if event.sessionID != "" {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO daily_sessions(site_id, day, session_id)
			VALUES (?, ?, ?)
			ON CONFLICT(site_id, day, session_id) DO NOTHING
		`, event.siteID, day, event.sessionID); err != nil {
			return fmt.Errorf("update daily sessions: %w", err)
		}
	}
	return nil
}

func rebuildSession(ctx context.Context, tx *sql.Tx, siteID, sessionID string, throughSeq int64) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO sessions(
			site_id, session_id, visitor_id, started_at_us, ended_at_us,
			entry_pathname, exit_pathname, referrer_host, pageviews,
			event_count, is_bounce, projection_version
		)
		SELECT
			e.site_id,
			e.session_id,
			COALESCE((
				SELECT v.visitor_id FROM events v
				WHERE v.site_id = e.site_id AND v.session_id = e.session_id
				  AND v.seq <= ? AND v.visitor_id != ''
				ORDER BY v.occurred_at_us, v.seq LIMIT 1
			), ''),
			MIN(e.occurred_at_us),
			MAX(e.occurred_at_us),
			COALESCE((
				SELECT p.pathname FROM events p
				WHERE p.site_id = e.site_id AND p.session_id = e.session_id
				  AND p.seq <= ? AND p.event_name = '$pageview'
				ORDER BY p.occurred_at_us, p.seq LIMIT 1
			), '/'),
			COALESCE((
				SELECT p.pathname FROM events p
				WHERE p.site_id = e.site_id AND p.session_id = e.session_id
				  AND p.seq <= ? AND p.event_name = '$pageview'
				ORDER BY p.occurred_at_us DESC, p.seq DESC LIMIT 1
			), '/'),
			COALESCE((
				SELECT ref.referrer_host FROM events ref
				WHERE ref.site_id = e.site_id AND ref.session_id = e.session_id
				  AND ref.seq <= ? AND ref.referrer_host != ''
				ORDER BY ref.occurred_at_us, ref.seq LIMIT 1
			), ''),
			SUM(CASE WHEN e.event_name = '$pageview' THEN 1 ELSE 0 END),
			COUNT(*),
			CASE WHEN SUM(CASE WHEN e.event_name = '$pageview' THEN 1 ELSE 0 END) <= 1
				THEN 1 ELSE 0 END,
			?
		FROM events e
		WHERE e.site_id = ? AND e.session_id = ? AND e.seq <= ?
		GROUP BY e.site_id, e.session_id
		ON CONFLICT(site_id, session_id) DO UPDATE SET
			visitor_id = excluded.visitor_id,
			started_at_us = excluded.started_at_us,
			ended_at_us = excluded.ended_at_us,
			entry_pathname = excluded.entry_pathname,
			exit_pathname = excluded.exit_pathname,
			referrer_host = excluded.referrer_host,
			pageviews = excluded.pageviews,
			event_count = excluded.event_count,
			is_bounce = excluded.is_bounce,
			projection_version = excluded.projection_version
	`, throughSeq, throughSeq, throughSeq, throughSeq, analyticsProjectionVersion,
		siteID, sessionID, throughSeq)
	if err != nil {
		return fmt.Errorf("rebuild session %q for site %q: %w", sessionID, siteID, err)
	}
	return nil
}
