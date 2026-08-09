package core

import (
	"context"
	"errors"
	"time"
)

var (
	ErrSiteNotFound     = errors.New("site not found")
	ErrDomainNotAllowed = errors.New("domain not allowed")
)

type Event struct {
	ID            string         `json:"id"           db:"id"`
	EventName     string         `json:"n"            db:"event_name"`
	URL           string         `json:"u"            db:"url"`
	Domain        string         `json:"d"            db:"domain"`
	Referrer      string         `json:"r,omitempty"  db:"referrer"`
	ScreenWidth   int            `json:"w"            db:"screen_width"`
	SiteID        string         `json:"s"            db:"site_id"`
	SessionID     string         `json:"sid"          db:"session_id"`
	VisitorID     string         `json:"vid"          db:"visitor_id"`
	Properties    map[string]any `json:"p,omitempty"  db:"properties"`
	Timestamp     time.Time      `json:"ts,omitempty" db:"timestamp"`
	ReceivedAt    time.Time      `json:"-"             db:"received_at"`
	Pathname      string         `json:"-"             db:"pathname"`
	ReferrerHost  string         `json:"-"             db:"referrer_host"`
	SchemaVersion int            `json:"v,omitempty"   db:"schema_version"`
	SDKVersion    string         `json:"sv,omitempty"  db:"sdk_version"`
}

type Site struct {
	ID            string   `json:"site_id"`
	Name          string   `json:"name"`
	Timezone      string   `json:"timezone"`
	RetentionDays int      `json:"retention_days"`
	Domains       []string `json:"domains"`
}

type StatsResult struct {
	Pageviews      int `json:"pageviews"`
	UniqueVisitors int `json:"unique_visitors"`
	Sessions       int `json:"sessions"`
}

type StatsChange struct {
	Pageviews      float64 `json:"pageviews"`
	UniqueVisitors float64 `json:"unique_visitors"`
	Sessions       float64 `json:"sessions"`
}

type SiteTrendResult struct {
	Current  StatsResult `json:"current"`
	Previous StatsResult `json:"previous"`
	Change   StatsChange `json:"change"`
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
}

type VitalDistribution struct {
	Name             string `json:"name"`
	Total            int    `json:"total"`
	Good             int    `json:"good"`
	NeedsImprovement int    `json:"needs_improvement"`
	Poor             int    `json:"poor"`
}

type PagePerformanceStat struct {
	URL     string   `json:"url"`
	LCP     *float64 `json:"lcp"`
	INP     *float64 `json:"inp"`
	CLS     *float64 `json:"cls"`
	Traffic int      `json:"traffic"`
}

type PerformanceScore struct {
	Score        int            `json:"score"`
	Rating       string         `json:"rating"`
	MetricScores map[string]int `json:"metric_scores"`
	SampleSize   int            `json:"sample_size"`
}

type CustomEventSummary struct {
	TotalEvents    int     `json:"total_events"`
	UniqueUsers    int     `json:"unique_users"`
	ConversionRate float64 `json:"conversion_rate"`
	ChangePercent  float64 `json:"change_percent"`
}

type CustomEventStat struct {
	EventName     string  `json:"event_name"`
	TotalCount    int     `json:"total_count"`
	UniqueUsers   int     `json:"unique_users"`
	ChangePercent float64 `json:"change_percent"`
}

type CustomEventsResult struct {
	Summary CustomEventSummary `json:"summary"`
	Events  []CustomEventStat  `json:"events"`
}

type CustomEventTimeSeriesBucket struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

type DeviceStat struct {
	Device string `json:"device"`
	Count  int    `json:"count"`
}

type SiteStat struct {
	SiteID        string   `json:"site_id"`
	Name          string   `json:"name"`
	Domain        string   `json:"domain"`
	Domains       []string `json:"domains,omitempty"`
	Timezone      string   `json:"timezone"`
	RetentionDays int      `json:"retention_days"`
}

type TimeSeriesBucket struct {
	Date           string `json:"date"` // "YYYY-MM-DD" in UTC
	Pageviews      int    `json:"pageviews"`
	UniqueVisitors int    `json:"uniqueVisitors,omitempty"`
	Sessions       int    `json:"sessions,omitempty"`
}

type EventRepository interface {
	CreateSite(ctx context.Context, site *Site) error
	ValidateSite(ctx context.Context, siteID, domain string) error
	Insert(ctx context.Context, event *Event) error
	InsertBatch(ctx context.Context, events []*Event) error
	GetStats(ctx context.Context, siteKey, from, to string) (*StatsResult, error)
	GetTopPages(ctx context.Context, siteKey, from, to string, limit int) ([]PageStat, error)
	GetTopReferrers(ctx context.Context, siteKey, from, to string, limit int) ([]ReferrerStat, error)
	GetVitals(ctx context.Context, siteKey, from, to string) ([]VitalStat, error)
	GetVitalDistributions(ctx context.Context, siteKey, from, to string) ([]VitalDistribution, error)
	GetPagePerformance(ctx context.Context, siteKey, from, to string, limit int) ([]PagePerformanceStat, error)
	GetPerformanceScore(ctx context.Context, siteKey, from, to string) (*PerformanceScore, error)
	GetCustomEvents(ctx context.Context, siteKey, from, to string) (*CustomEventsResult, error)
	GetCustomEventTimeSeries(ctx context.Context, siteKey, eventName, from, to string) ([]CustomEventTimeSeriesBucket, error)
	GetDevices(ctx context.Context, siteKey, from, to string) ([]DeviceStat, error)
	GetPageviewsTimeSeries(ctx context.Context, siteKey, from, to string) ([]TimeSeriesBucket, error)
	GetUniqueVisitorsTimeSeries(ctx context.Context, siteKey, from, to string) ([]TimeSeriesBucket, error)
	GetSessionsTimeSeries(ctx context.Context, siteKey, from, to string) ([]TimeSeriesBucket, error)
	GetSites(ctx context.Context) ([]SiteStat, error)
	Close() error
}
