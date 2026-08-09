package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/VatsalP117/iris/pkg/core"
	_ "github.com/mattn/go-sqlite3"
)

type SqliteRepository struct {
	db     *sql.DB
	writer *sql.DB
}

// ConstrainGrowthPages limits this database to its current size plus extraPages.
// It exists for the isolated reliability lab's disk-exhaustion scenario.
func (r *SqliteRepository) ConstrainGrowthPages(ctx context.Context, extraPages int) (int, error) {
	if extraPages < 0 {
		return 0, fmt.Errorf("extra pages must be non-negative")
	}
	var pageCount int
	if err := r.writer.QueryRowContext(ctx, "PRAGMA page_count").Scan(&pageCount); err != nil {
		return 0, err
	}
	var appliedLimit int
	if err := r.writer.QueryRowContext(
		ctx,
		fmt.Sprintf("PRAGMA max_page_count = %d", pageCount+extraPages),
	).Scan(&appliedLimit); err != nil {
		return 0, err
	}
	return appliedLimit, nil
}

func (r *SqliteRepository) ResetGrowthPageLimit(ctx context.Context) error {
	var appliedLimit int
	return r.writer.QueryRowContext(
		ctx,
		"PRAGMA max_page_count = 1073741823",
	).Scan(&appliedLimit)
}

func NewSqliteDB(filepath string) (*SqliteRepository, error) {
	dsn := filepath
	separator := "?"
	if strings.Contains(dsn, "?") {
		separator = "&"
	}
	dsn += separator + "_foreign_keys=on&_journal_mode=WAL&_busy_timeout=5000&_synchronous=NORMAL"

	writer, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	writer.SetMaxOpenConns(1)
	writer.SetMaxIdleConns(1)
	if err := writer.Ping(); err != nil {
		writer.Close()
		return nil, err
	}
	if err := migrate(context.Background(), writer); err != nil {
		writer.Close()
		return nil, err
	}

	reader, err := sql.Open("sqlite3", dsn)
	if err != nil {
		writer.Close()
		return nil, err
	}
	reader.SetMaxOpenConns(4)
	reader.SetMaxIdleConns(4)
	if err := reader.Ping(); err != nil {
		reader.Close()
		writer.Close()
		return nil, err
	}

	return &SqliteRepository{db: reader, writer: writer}, nil
}

func (r *SqliteRepository) Insert(ctx context.Context, e *core.Event) error {
	if err := r.requireSites(ctx, []*core.Event{e}); err != nil {
		return err
	}
	propsJSON, err := json.Marshal(e.Properties)
	if err != nil {
		return fmt.Errorf("encode event properties: %w", err)
	}
	prepareEventTimes(e)

	query := `
	INSERT INTO events (
		id, event_name, site_id, occurred_at_us, received_at_us, timestamp,
		url, domain, pathname, referrer, referrer_host, screen_width,
		session_id, visitor_id, properties, schema_version, sdk_version
	)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(id) DO NOTHING
	`

	_, err = r.writer.ExecContext(ctx, query,
		e.ID,
		e.EventName,
		e.SiteID,
		e.Timestamp.UnixMicro(),
		e.ReceivedAt.UnixMicro(),
		e.Timestamp,
		e.URL,
		e.Domain,
		e.Pathname,
		e.Referrer,
		e.ReferrerHost,
		e.ScreenWidth,
		e.SessionID,
		e.VisitorID,
		string(propsJSON),
		e.SchemaVersion,
		e.SDKVersion,
	)

	return err
}

func (r *SqliteRepository) Close() error {
	readerErr := r.db.Close()
	writerErr := r.writer.Close()
	if readerErr != nil {
		return readerErr
	}
	return writerErr
}

func (r *SqliteRepository) InsertBatch(ctx context.Context, events []*core.Event) error {
	if err := r.requireSites(ctx, events); err != nil {
		return err
	}
	tx, err := r.writer.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
	INSERT INTO events (
		id, event_name, site_id, occurred_at_us, received_at_us, timestamp,
		url, domain, pathname, referrer, referrer_host, screen_width,
		session_id, visitor_id, properties, schema_version, sdk_version
	)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(id) DO NOTHING
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, e := range events {
		propsJSON, err := json.Marshal(e.Properties)
		if err != nil {
			return fmt.Errorf("encode event properties: %w", err)
		}
		prepareEventTimes(e)

		_, err = stmt.ExecContext(ctx,
			e.ID,
			e.EventName,
			e.SiteID,
			e.Timestamp.UnixMicro(),
			e.ReceivedAt.UnixMicro(),
			e.Timestamp,
			e.URL,
			e.Domain,
			e.Pathname,
			e.Referrer,
			e.ReferrerHost,
			e.ScreenWidth,
			e.SessionID,
			e.VisitorID,
			string(propsJSON),
			e.SchemaVersion,
			e.SDKVersion,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *SqliteRepository) requireSites(ctx context.Context, events []*core.Event) error {
	checked := map[string]struct{}{}
	for _, event := range events {
		if event == nil {
			return fmt.Errorf("event is required")
		}
		if _, ok := checked[event.SiteID]; ok {
			continue
		}
		var exists int
		err := r.db.QueryRowContext(ctx, `
			SELECT 1 FROM sites WHERE id = ? AND disabled_at_us IS NULL
		`, event.SiteID).Scan(&exists)
		if err == sql.ErrNoRows {
			return fmt.Errorf("%w: %s", core.ErrSiteNotFound, event.SiteID)
		}
		if err != nil {
			return err
		}
		checked[event.SiteID] = struct{}{}
	}
	return nil
}

func prepareEventTimes(event *core.Event) {
	now := time.Now().UTC()
	if event.Timestamp.IsZero() {
		event.Timestamp = now
	} else {
		event.Timestamp = event.Timestamp.UTC()
	}
	if event.ReceivedAt.IsZero() {
		event.ReceivedAt = now
	} else {
		event.ReceivedAt = event.ReceivedAt.UTC()
	}
	if event.SchemaVersion <= 0 {
		event.SchemaVersion = 1
	}
}
