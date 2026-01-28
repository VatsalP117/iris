package db

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/VatsalP117/iris/pkg/core"
	_ "github.com/mattn/go-sqlite3"
)

type SqliteRepository struct {
	db *sql.DB
}

func NewSqliteDB(filepath string) (*SqliteRepository, error) {
	db, err := sql.Open("sqlite3", filepath)
	if err != nil {
		return nil, err
	}

	query := `
	CREATE TABLE IF NOT EXISTS events (
		id TEXT PRIMARY KEY,
		event_name TEXT,
		url TEXT,
		domain TEXT,
		referrer TEXT,
		screen_width INTEGER,
		session_id TEXT,
		properties TEXT,
		timestamp DATETIME
	);
	`
	_, err = db.Exec(query)
	if err != nil {
		return nil, err
	}

	return &SqliteRepository{db: db}, nil
}

func (r *SqliteRepository) Insert(ctx context.Context, e *core.Event) error {
	propsJson, err := json.Marshal(e.Properties)
	if err != nil {
		propsJson = []byte("{}")
	}

	query := `
	INSERT INTO events (id, event_name, url, domain, referrer, screen_width, session_id, properties, timestamp)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = r.db.ExecContext(ctx, query,
		e.ID,
		e.EventName,
		e.URL,
		e.Domain,
		e.Referrer,
		e.ScreenWidth,
		e.SessionID,
		string(propsJson), 
		e.Timestamp,
	)

	return err
}

func (r *SqliteRepository) Close() error {
	return r.db.Close()
}