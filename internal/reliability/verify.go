package reliability

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/VatsalP117/iris/pkg/core"
	_ "github.com/mattn/go-sqlite3"
)

type storedEvent struct {
	EventName   string
	URL         string
	Domain      string
	Referrer    string
	ScreenWidth int
	SiteID      string
	SessionID   string
	VisitorID   string
	Properties  map[string]any
}

func VerifyStorage(ctx context.Context, dbPath string, manifest []PlannedEvent, accepted map[int]struct{}) (StorageSummary, error) {
	absolutePath, err := filepath.Abs(dbPath)
	if err != nil {
		return StorageSummary{}, err
	}
	dsn := "file:" + filepath.ToSlash(absolutePath) + "?mode=ro&_busy_timeout=5000"
	database, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return StorageSummary{}, err
	}
	defer database.Close()

	if err := database.PingContext(ctx); err != nil {
		return StorageSummary{}, err
	}

	planned := make(map[int]PlannedEvent, len(manifest))
	for _, event := range manifest {
		planned[event.Sequence] = event
	}

	rows, err := database.QueryContext(ctx, `
		SELECT
			CAST(json_extract(properties, '$."$test_seq"') AS INTEGER),
			event_name,
			url,
			domain,
			COALESCE(referrer, ''),
			screen_width,
			site_id,
			session_id,
			visitor_id,
			properties
		FROM events
		WHERE json_extract(properties, '$."$test_run"') = ?
	`, manifestRunID(manifest))
	if err != nil {
		return StorageSummary{}, err
	}
	defer rows.Close()

	storedBySequence := map[int][]storedEvent{}
	summary := StorageSummary{}
	for rows.Next() {
		var sequence int
		var event storedEvent
		var rawProperties string
		if err := rows.Scan(
			&sequence,
			&event.EventName,
			&event.URL,
			&event.Domain,
			&event.Referrer,
			&event.ScreenWidth,
			&event.SiteID,
			&event.SessionID,
			&event.VisitorID,
			&rawProperties,
		); err != nil {
			return StorageSummary{}, err
		}
		if err := json.Unmarshal([]byte(rawProperties), &event.Properties); err != nil {
			return StorageSummary{}, fmt.Errorf("decode properties for sequence %d: %w", sequence, err)
		}
		storedBySequence[sequence] = append(storedBySequence[sequence], event)
		summary.StoredRows++
	}
	if err := rows.Err(); err != nil {
		return StorageSummary{}, err
	}

	summary.UniqueSequences = len(storedBySequence)
	for sequence := range accepted {
		events := storedBySequence[sequence]
		if len(events) == 0 {
			summary.MissingEvents++
			appendIntSample(&summary.MissingSamples, sequence)
			continue
		}
		if len(events) > 1 {
			summary.DuplicateRows += len(events) - 1
			appendIntSample(&summary.DuplicateSamples, sequence)
		}
		expected, ok := planned[sequence]
		if !ok || !storedMatches(expected, events[0]) {
			summary.FieldMismatches++
			appendIntSample(&summary.FieldMismatchSamples, sequence)
		}
	}

	for sequence, events := range storedBySequence {
		if _, ok := accepted[sequence]; ok {
			continue
		}
		summary.UnexpectedRows += len(events)
		appendIntSample(&summary.UnexpectedSamples, sequence)
	}

	sort.Ints(summary.MissingSamples)
	sort.Ints(summary.DuplicateSamples)
	sort.Ints(summary.UnexpectedSamples)
	sort.Ints(summary.FieldMismatchSamples)
	if info, err := os.Stat(absolutePath); err == nil {
		summary.DatabaseBytes = info.Size()
	}
	return summary, nil
}

func storedMatches(planned PlannedEvent, stored storedEvent) bool {
	expected := planned.Event
	if stored.EventName != expected.EventName ||
		stored.URL != expected.URL ||
		stored.Domain != expected.Domain ||
		stored.Referrer != expected.Referrer ||
		stored.ScreenWidth != expected.ScreenWidth ||
		stored.SiteID != expected.SiteID ||
		stored.SessionID != expected.SessionID ||
		stored.VisitorID != expected.VisitorID {
		return false
	}

	return propertyString(stored.Properties, "$test_run") == propertyString(expected.Properties, "$test_run") &&
		propertyInt(stored.Properties, "$test_seq") == planned.Sequence &&
		propertyString(stored.Properties, "$test_flow") == propertyString(expected.Properties, "$test_flow")
}

func propertyString(properties map[string]any, key string) string {
	value, _ := properties[key].(string)
	return value
}

func propertyInt(properties map[string]any, key string) int {
	switch value := properties[key].(type) {
	case int:
		return value
	case float64:
		return int(value)
	case json.Number:
		parsed, _ := value.Int64()
		return int(parsed)
	default:
		return -1
	}
}

func manifestRunID(manifest []PlannedEvent) string {
	if len(manifest) == 0 {
		return ""
	}
	return propertyString(manifest[0].Event.Properties, "$test_run")
}

func appendIntSample(samples *[]int, value int) {
	if len(*samples) < maxDiagnosticSample {
		*samples = append(*samples, value)
	}
}

func VerifyAggregates(ctx context.Context, config Config, manifest []PlannedEvent, accepted map[int]struct{}) []AggregateCheck {
	expectedStats := core.StatsResult{}
	expectedPages := map[string]int{}
	expectedDevices := map[string]int{}
	expectedReferrers := map[string]map[string]struct{}{}
	expectedVitals := map[string][]float64{}
	visitors := map[string]struct{}{}
	sessions := map[string]struct{}{}

	for _, planned := range manifest {
		if _, ok := accepted[planned.Sequence]; !ok {
			continue
		}
		if planned.Event.EventName == "$web_vital" {
			name := propertyString(planned.Event.Properties, "$name")
			if value, ok := propertyFloat(planned.Event.Properties, "$val"); ok {
				expectedVitals[name] = append(expectedVitals[name], value)
			}
		}
		if planned.Event.EventName != "$pageview" {
			continue
		}

		expectedStats.Pageviews++
		visitors[planned.Event.VisitorID] = struct{}{}
		sessions[planned.Event.SessionID] = struct{}{}
		expectedPages[planned.Event.URL]++
		expectedDevices[deviceForWidth(planned.Event.ScreenWidth)]++
		host := normalizeLabReferrer(planned.Event.Referrer)
		if host != "" {
			if expectedReferrers[host] == nil {
				expectedReferrers[host] = map[string]struct{}{}
			}
			expectedReferrers[host][planned.Event.VisitorID] = struct{}{}
		}
	}
	expectedStats.UniqueVisitors = len(visitors)
	expectedStats.Sessions = len(sessions)

	checks := []AggregateCheck{
		verifyJSONAggregate(ctx, config, "stats", "/api/stats", expectedStats),
		verifyJSONAggregate(ctx, config, "pages", "/api/pages", pageStats(expectedPages)),
		verifyJSONAggregate(ctx, config, "referrers", "/api/referrers", referrerStats(expectedReferrers)),
		verifyJSONAggregate(ctx, config, "vitals", "/api/vitals", vitalStats(expectedVitals)),
		verifyJSONAggregate(ctx, config, "devices", "/api/devices", deviceStats(expectedDevices)),
	}
	pageviews, uniqueVisitors, sessionSeries, err := expectedTimeSeries(ctx, config)
	if err != nil {
		errorCheck := AggregateCheck{Name: "timeseries-source", Error: err.Error()}
		return append(checks, errorCheck)
	}
	checks = append(
		checks,
		verifyJSONAggregate(ctx, config, "pageviews-timeseries", "/api/timeseries", pageviews),
		verifyJSONAggregate(ctx, config, "visitors-timeseries", "/api/timeseries/visitors", uniqueVisitors),
		verifyJSONAggregate(ctx, config, "sessions-timeseries", "/api/timeseries/sessions", sessionSeries),
	)
	return append(checks, verifyDateWindows(ctx, config)...)
}

func verifyJSONAggregate(ctx context.Context, config Config, name, path string, expected any) AggregateCheck {
	endpoint := config.TargetURL + path + "?site_id=" + url.QueryEscape(config.SiteID)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return AggregateCheck{Name: name, Expected: expected, Error: err.Error()}
	}
	response, err := (&http.Client{Timeout: config.RequestTimeout}).Do(request)
	if err != nil {
		return AggregateCheck{Name: name, Expected: expected, Error: err.Error()}
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return AggregateCheck{
			Name:     name,
			Expected: expected,
			Error:    fmt.Sprintf("HTTP %d", response.StatusCode),
		}
	}

	var decodeTarget any
	switch expected.(type) {
	case core.StatsResult:
		decodeTarget = &core.StatsResult{}
	case []core.PageStat:
		decodeTarget = &[]core.PageStat{}
	case []core.DeviceStat:
		decodeTarget = &[]core.DeviceStat{}
	case []core.ReferrerStat:
		decodeTarget = &[]core.ReferrerStat{}
	case []core.VitalStat:
		decodeTarget = &[]core.VitalStat{}
	case []core.TimeSeriesBucket:
		decodeTarget = &[]core.TimeSeriesBucket{}
	default:
		return AggregateCheck{Name: name, Expected: expected, Error: "unsupported aggregate type"}
	}

	if err := json.NewDecoder(response.Body).Decode(decodeTarget); err != nil {
		return AggregateCheck{Name: name, Expected: expected, Error: err.Error()}
	}

	actual := canonicalAggregate(decodeTarget)
	check := AggregateCheck{Name: name, Expected: expected, Actual: actual}
	check.Passed = reflect.DeepEqual(expected, check.Actual)
	return check
}

func canonicalAggregate(value any) any {
	switch typed := value.(type) {
	case *core.StatsResult:
		return *typed
	case *[]core.PageStat:
		result := append([]core.PageStat(nil), (*typed)...)
		sort.Slice(result, func(i, j int) bool {
			if result[i].Pageviews == result[j].Pageviews {
				return result[i].URL < result[j].URL
			}
			return result[i].Pageviews > result[j].Pageviews
		})
		return result
	case *[]core.DeviceStat:
		result := append([]core.DeviceStat(nil), (*typed)...)
		sort.Slice(result, func(i, j int) bool {
			if result[i].Count == result[j].Count {
				return result[i].Device < result[j].Device
			}
			return result[i].Count > result[j].Count
		})
		return result
	case *[]core.ReferrerStat:
		result := append([]core.ReferrerStat(nil), (*typed)...)
		sort.Slice(result, func(i, j int) bool {
			if result[i].Visitors == result[j].Visitors {
				return result[i].Referrer < result[j].Referrer
			}
			return result[i].Visitors > result[j].Visitors
		})
		return result
	case *[]core.VitalStat:
		result := append([]core.VitalStat(nil), (*typed)...)
		sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
		return result
	case *[]core.TimeSeriesBucket:
		result := append([]core.TimeSeriesBucket(nil), (*typed)...)
		sort.Slice(result, func(i, j int) bool { return result[i].Date < result[j].Date })
		return result
	default:
		return value
	}
}

func pageStats(counts map[string]int) []core.PageStat {
	result := make([]core.PageStat, 0, len(counts))
	for page, count := range counts {
		result = append(result, core.PageStat{URL: page, Pageviews: count})
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Pageviews == result[j].Pageviews {
			return result[i].URL < result[j].URL
		}
		return result[i].Pageviews > result[j].Pageviews
	})
	return result
}

func deviceStats(counts map[string]int) []core.DeviceStat {
	result := make([]core.DeviceStat, 0, len(counts))
	for device, count := range counts {
		result = append(result, core.DeviceStat{Device: device, Count: count})
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Count == result[j].Count {
			return result[i].Device < result[j].Device
		}
		return result[i].Count > result[j].Count
	})
	return result
}

func deviceForWidth(width int) string {
	switch {
	case width < 768:
		return "Mobile"
	case width < 1024:
		return "Tablet"
	default:
		return "Desktop"
	}
}

func propertyFloat(properties map[string]any, key string) (float64, bool) {
	switch value := properties[key].(type) {
	case float64:
		return value, true
	case int:
		return float64(value), true
	case json.Number:
		parsed, err := value.Float64()
		return parsed, err == nil
	default:
		return 0, false
	}
}

func normalizeLabReferrer(raw string) string {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Hostname() == "" {
		return ""
	}
	return strings.TrimPrefix(strings.ToLower(parsed.Hostname()), "www.")
}

func referrerStats(values map[string]map[string]struct{}) []core.ReferrerStat {
	result := make([]core.ReferrerStat, 0, len(values))
	for host, visitors := range values {
		result = append(result, core.ReferrerStat{Referrer: host, Visitors: len(visitors)})
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Visitors == result[j].Visitors {
			return result[i].Referrer < result[j].Referrer
		}
		return result[i].Visitors > result[j].Visitors
	})
	return result
}

func vitalStats(values map[string][]float64) []core.VitalStat {
	result := make([]core.VitalStat, 0, len(values))
	for name, samples := range values {
		sorted := append([]float64(nil), samples...)
		sort.Float64s(sorted)
		index := int(math.Ceil(0.75*float64(len(sorted)))) - 1
		if index < 0 {
			index = 0
		}
		result = append(result, core.VitalStat{Name: name, Value: sorted[index]})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result
}

func expectedTimeSeries(
	ctx context.Context,
	config Config,
) ([]core.TimeSeriesBucket, []core.TimeSeriesBucket, []core.TimeSeriesBucket, error) {
	absolutePath, err := filepath.Abs(config.DBPath)
	if err != nil {
		return nil, nil, nil, err
	}
	database, err := sql.Open("sqlite3", "file:"+filepath.ToSlash(absolutePath)+"?mode=ro&_busy_timeout=5000")
	if err != nil {
		return nil, nil, nil, err
	}
	defer database.Close()

	rows, err := database.QueryContext(ctx, `
		SELECT
			strftime('%Y-%m-%d', timestamp),
			COUNT(*),
			COUNT(DISTINCT visitor_id),
			COUNT(DISTINCT session_id)
		FROM events
		WHERE event_name = '$pageview'
		  AND site_id = ?
		  AND json_extract(properties, '$."$test_run"') = ?
		GROUP BY strftime('%Y-%m-%d', timestamp)
		ORDER BY 1
	`, config.SiteID, config.RunID)
	if err != nil {
		return nil, nil, nil, err
	}
	defer rows.Close()

	var pageviews []core.TimeSeriesBucket
	var visitors []core.TimeSeriesBucket
	var sessions []core.TimeSeriesBucket
	for rows.Next() {
		var date string
		var pageviewCount, visitorCount, sessionCount int
		if err := rows.Scan(&date, &pageviewCount, &visitorCount, &sessionCount); err != nil {
			return nil, nil, nil, err
		}
		pageviews = append(pageviews, core.TimeSeriesBucket{Date: date, Pageviews: pageviewCount})
		visitors = append(visitors, core.TimeSeriesBucket{Date: date, UniqueVisitors: visitorCount})
		sessions = append(sessions, core.TimeSeriesBucket{Date: date, Sessions: sessionCount})
	}
	if err := rows.Err(); err != nil {
		return nil, nil, nil, err
	}
	return pageviews, visitors, sessions, nil
}

func verifyDateWindows(ctx context.Context, config Config) []AggregateCheck {
	absolutePath, err := filepath.Abs(config.DBPath)
	if err != nil {
		return []AggregateCheck{{Name: "date-window-fixture", Error: err.Error()}}
	}
	database, err := sql.Open("sqlite3", absolutePath)
	if err != nil {
		return []AggregateCheck{{Name: "date-window-fixture", Error: err.Error()}}
	}
	defer database.Close()

	siteID := config.SiteID + "-date-window"
	timestamps := []string{
		"2026-03-24 00:00:00",
		"2026-03-24 23:59:59",
		"2026-03-25 00:00:00",
	}
	for index, timestamp := range timestamps {
		_, err := database.ExecContext(ctx, `
			INSERT OR REPLACE INTO events (
				id, event_name, url, domain, referrer, screen_width, site_id,
				session_id, visitor_id, properties, timestamp
			) VALUES (?, '$pageview', ?, ?, '', 1440, ?, ?, ?, '{}', ?)
		`,
			fmt.Sprintf("iris-lab-date-%s-%d", config.RunID, index),
			fmt.Sprintf("https://%s/date-%d", defaultDomain, index),
			defaultDomain,
			siteID,
			fmt.Sprintf("date-session-%d", index),
			fmt.Sprintf("date-visitor-%d", index),
			timestamp,
		)
		if err != nil {
			return []AggregateCheck{{Name: "date-window-fixture", Error: err.Error()}}
		}
	}

	dayURL := config.TargetURL + "/api/stats?site_id=" + url.QueryEscape(siteID) +
		"&from=2026-03-24&to=2026-03-24"
	timeURL := config.TargetURL + "/api/stats?site_id=" + url.QueryEscape(siteID) +
		"&from=2026-03-24T12%3A00%3A00Z&to=2026-03-25T00%3A00%3A00Z"
	return []AggregateCheck{
		verifyStatsURL(ctx, config, "date-window-day", dayURL, core.StatsResult{Pageviews: 2, UniqueVisitors: 2, Sessions: 2}),
		verifyStatsURL(ctx, config, "date-window-time", timeURL, core.StatsResult{Pageviews: 2, UniqueVisitors: 2, Sessions: 2}),
	}
}

func verifyStatsURL(
	ctx context.Context,
	config Config,
	name string,
	endpoint string,
	expected core.StatsResult,
) AggregateCheck {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return AggregateCheck{Name: name, Expected: expected, Error: err.Error()}
	}
	response, err := (&http.Client{Timeout: config.RequestTimeout}).Do(request)
	if err != nil {
		return AggregateCheck{Name: name, Expected: expected, Error: err.Error()}
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return AggregateCheck{Name: name, Expected: expected, Error: fmt.Sprintf("HTTP %d", response.StatusCode)}
	}
	var actual core.StatsResult
	if err := json.NewDecoder(response.Body).Decode(&actual); err != nil {
		return AggregateCheck{Name: name, Expected: expected, Error: err.Error()}
	}
	return AggregateCheck{
		Name:     name,
		Expected: expected,
		Actual:   actual,
		Passed:   reflect.DeepEqual(expected, actual),
	}
}
