package core

import (
	"context"
	"time"
)

type Event struct {
	ID string `json:"id" db:"id"`
	EventName string `json:"n" db:"event_name"`
	URL string `json:"u" db:"url"`
	Domain string `json:"d" db:"domain"`
	Referrer string `json:"r,omitempty" db:"referrer"`
	ScreenWidth int `json:"w" db:"screen_width"`
	SessionID string `json:"s" db:"session_id"`
	Properties map[string]any `json:"p,omitempty" db:"properties"`
	Timestamp time.Time `json:"ts" db:"timestamp"`
}

type EventRepository interface {
	Insert(ctx context.Context, event *Event) error
	Close() error
}