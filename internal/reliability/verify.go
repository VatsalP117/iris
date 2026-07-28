package reliability

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"sort"

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
	visitors := map[string]struct{}{}
	sessions := map[string]struct{}{}

	for _, planned := range manifest {
		if _, ok := accepted[planned.Sequence]; !ok || planned.Event.EventName != "$pageview" {
			continue
		}
		expectedStats.Pageviews++
		visitors[planned.Event.VisitorID] = struct{}{}
		sessions[planned.Event.SessionID] = struct{}{}
		expectedPages[planned.Event.URL]++
		expectedDevices[deviceForWidth(planned.Event.ScreenWidth)]++
	}
	expectedStats.UniqueVisitors = len(visitors)
	expectedStats.Sessions = len(sessions)

	return []AggregateCheck{
		verifyJSONAggregate(ctx, config, "stats", "/api/stats", expectedStats),
		verifyJSONAggregate(ctx, config, "pages", "/api/pages", pageStats(expectedPages)),
		verifyJSONAggregate(ctx, config, "devices", "/api/devices", deviceStats(expectedDevices)),
	}
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
