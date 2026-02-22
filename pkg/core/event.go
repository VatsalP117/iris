package core

import (
	"context"
	"time"
)

type Event struct {
	ID          string         `json:"id"           db:"id"`
	EventName   string         `json:"n"            db:"event_name"`
	URL         string         `json:"u"            db:"url"`
	Domain      string         `json:"d"            db:"domain"`
	Referrer    string         `json:"r,omitempty"  db:"referrer"`
	ScreenWidth int            `json:"w"            db:"screen_width"`
	SiteID      string         `json:"s"            db:"site_id"`
	SessionID   string         `json:"sid"          db:"session_id"`
	VisitorID   string         `json:"vid"          db:"visitor_id"`
	Properties  map[string]any `json:"p,omitempty"  db:"properties"`
	Timestamp   time.Time      `json:"ts"           db:"timestamp"`
}

type StatsResult struct {
	Pageviews      int `json:"pageviews"`
	UniqueVisitors int `json:"unique_visitors"`
	Sessions       int `json:"sessions"`
}

type PageStat struct {
	URL       string `json:"url"`
	Pageviews int    `json:"pageviews"`
}

type ReferrerStat struct {
	Referrer string `json:"referrer"`
	Visitors int    `json:"visitors"`
}

type VitalStat struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
	// "good", "needs-improvement", or "poor" — based on last recorded rating
}

type DeviceStat struct {
	Device string `json:"device"`
	Count  int    `json:"count"`
}

type SiteStat struct {
	SiteID string `json:"site_id"`
	Domain string `json:"domain"`
}

type TimeSeriesBucket struct {
	Date      string `json:"date"` // "YYYY-MM-DD" in UTC
	Pageviews int    `json:"pageviews"`
}

type EventRepository interface {
	Insert(ctx context.Context, event *Event) error
	GetStats(ctx context.Context, domain, from, to string) (*StatsResult, error)
	GetTopPages(ctx context.Context, domain, from, to string, limit int) ([]PageStat, error)
	GetTopReferrers(ctx context.Context, domain, from, to string, limit int) ([]ReferrerStat, error)
	GetVitals(ctx context.Context, domain, from, to string) ([]VitalStat, error)
	GetDevices(ctx context.Context, domain, from, to string) ([]DeviceStat, error)
	GetPageviewsTimeSeries(ctx context.Context, domain, from, to string) ([]TimeSeriesBucket, error)
	GetSites(ctx context.Context) ([]SiteStat, error)
	Close() error
}
