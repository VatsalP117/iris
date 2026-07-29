package db

import (
	"context"
	"database/sql"
	"math"
	"net/url"
	"sort"
	"strings"

	"github.com/VatsalP117/iris/pkg/core"
)

//
// Helper to build WHERE clauses consistently.
// All query methods accept a site key plus from/to strings (date or datetime, or empty).
// The site key can be the logical site_id or a legacy domain value.
//

const siteMatchClause = "(COALESCE(NULLIF(site_id, ''), domain) = ? OR domain = ?)"
const fromTimeClause = "(? = '' OR datetime(timestamp) >= datetime(?))"
const toTimeClause = "(? = '' OR datetime(timestamp) <= datetime(CASE WHEN length(trim(?)) = 10 THEN ? || ' 23:59:59' ELSE ? END))"

func (r *SqliteRepository) GetStats(ctx context.Context, siteKey, from, to string) (*core.StatsResult, error) {
	query := `
	SELECT
		COUNT(*)                                          AS pageviews,
		COUNT(DISTINCT visitor_id)                        AS unique_visitors,
		COUNT(DISTINCT session_id)                        AS sessions
	FROM events
	WHERE event_name = '$pageview'
	  AND ` + siteMatchClause + `
	  AND ` + fromTimeClause + `
	  AND ` + toTimeClause + `
	`
	row := r.db.QueryRowContext(ctx, query, siteKey, siteKey, from, from, to, to, to, to)

	var res core.StatsResult
	if err := row.Scan(&res.Pageviews, &res.UniqueVisitors, &res.Sessions); err != nil {
		return nil, err
	}
	return &res, nil
}

func (r *SqliteRepository) GetTopPages(ctx context.Context, siteKey, from, to string, limit int) ([]core.PageStat, error) {
	query := `
	SELECT url, COUNT(*) AS pageviews
	FROM events
	WHERE event_name = '$pageview'
	  AND ` + siteMatchClause + `
	  AND ` + fromTimeClause + `
	  AND ` + toTimeClause + `
	GROUP BY url
	ORDER BY pageviews DESC
	LIMIT ?
	`
	rows, err := r.db.QueryContext(ctx, query, siteKey, siteKey, from, from, to, to, to, to, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []core.PageStat
	for rows.Next() {
		var s core.PageStat
		if err := rows.Scan(&s.URL, &s.Pageviews); err != nil {
			return nil, err
		}
		results = append(results, s)
	}
	return results, rows.Err()
}

func (r *SqliteRepository) GetTopReferrers(ctx context.Context, siteKey, from, to string, limit int) ([]core.ReferrerStat, error) {
	query := `
	SELECT referrer, visitor_id
	FROM events
	WHERE event_name = '$pageview'
	  AND ` + siteMatchClause + `
	  AND referrer != ''
	  AND ` + fromTimeClause + `
	  AND ` + toTimeClause + `
	`
	rows, err := r.db.QueryContext(ctx, query, siteKey, siteKey, from, from, to, to, to, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	visitorsByHost := map[string]map[string]struct{}{}
	for rows.Next() {
		var referrer string
		var visitorID string
		if err := rows.Scan(&referrer, &visitorID); err != nil {
			return nil, err
		}

		host := normalizeReferrer(referrer)
		if host == "" {
			continue
		}
		if visitorsByHost[host] == nil {
			visitorsByHost[host] = map[string]struct{}{}
		}
		visitorsByHost[host][visitorID] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	results := make([]core.ReferrerStat, 0, len(visitorsByHost))
	for host, visitors := range visitorsByHost {
		results = append(results, core.ReferrerStat{
			Referrer: host,
			Visitors: len(visitors),
		})
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].Visitors == results[j].Visitors {
			return results[i].Referrer < results[j].Referrer
		}
		return results[i].Visitors > results[j].Visitors
	})

	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

func (r *SqliteRepository) GetVitals(ctx context.Context, siteKey, from, to string) ([]core.VitalStat, error) {
	query := `
	SELECT
		json_extract(properties, '$.$name') AS name,
		CAST(json_extract(properties, '$.$val') AS REAL) AS value
	FROM events
	WHERE event_name = '$web_vital'
	  AND ` + siteMatchClause + `
	  AND ` + fromTimeClause + `
	  AND ` + toTimeClause + `
	`
	rows, err := r.db.QueryContext(ctx, query, siteKey, siteKey, from, from, to, to, to, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	valuesByMetric := map[string][]float64{}
	for rows.Next() {
		var name sql.NullString
		var value sql.NullFloat64
		if err := rows.Scan(&name, &value); err != nil {
			return nil, err
		}
		if !name.Valid || !value.Valid {
			continue
		}
		valuesByMetric[name.String] = append(valuesByMetric[name.String], value.Float64)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	names := make([]string, 0, len(valuesByMetric))
	for name := range valuesByMetric {
		names = append(names, name)
	}
	sort.Strings(names)

	results := make([]core.VitalStat, 0, len(names))
	for _, name := range names {
		values := valuesByMetric[name]
		sort.Float64s(values)
		results = append(results, core.VitalStat{
			Name:  name,
			Value: percentile75(values),
		})
	}

	return results, nil
}

func (r *SqliteRepository) GetVitalDistributions(ctx context.Context, siteKey, from, to string) ([]core.VitalDistribution, error) {
	query := `
	SELECT
		json_extract(properties, '$.$name') AS name,
		CAST(json_extract(properties, '$.$val') AS REAL) AS value
	FROM events
	WHERE event_name = '$web_vital'
	  AND ` + siteMatchClause + `
	  AND ` + fromTimeClause + `
	  AND ` + toTimeClause + `
	`
	rows, err := r.db.QueryContext(ctx, query, siteKey, siteKey, from, from, to, to, to, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	byName := map[string]*core.VitalDistribution{}
	for rows.Next() {
		var name sql.NullString
		var value sql.NullFloat64
		if err := rows.Scan(&name, &value); err != nil {
			return nil, err
		}
		if !name.Valid || !value.Valid || vitalThresholds[name.String] == [2]float64{} {
			continue
		}

		distribution := byName[name.String]
		if distribution == nil {
			distribution = &core.VitalDistribution{Name: name.String}
			byName[name.String] = distribution
		}
		distribution.Total++
		switch classifyVital(name.String, value.Float64) {
		case "good":
			distribution.Good++
		case "needs-improvement":
			distribution.NeedsImprovement++
		case "poor":
			distribution.Poor++
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	results := make([]core.VitalDistribution, 0, len(byName))
	for _, name := range []string{"LCP", "INP", "CLS"} {
		if distribution := byName[name]; distribution != nil {
			results = append(results, *distribution)
		}
	}
	return results, nil
}

func (r *SqliteRepository) GetPagePerformance(ctx context.Context, siteKey, from, to string, limit int) ([]core.PagePerformanceStat, error) {
	vitalsQuery := `
	SELECT
		url,
		json_extract(properties, '$.$name') AS name,
		CAST(json_extract(properties, '$.$val') AS REAL) AS value
	FROM events
	WHERE event_name = '$web_vital'
	  AND url != ''
	  AND ` + siteMatchClause + `
	  AND ` + fromTimeClause + `
	  AND ` + toTimeClause + `
	`
	rows, err := r.db.QueryContext(ctx, vitalsQuery, siteKey, siteKey, from, from, to, to, to, to)
	if err != nil {
		return nil, err
	}

	valuesByURL := map[string]map[string][]float64{}
	for rows.Next() {
		var pageURL string
		var name sql.NullString
		var value sql.NullFloat64
		if err := rows.Scan(&pageURL, &name, &value); err != nil {
			rows.Close()
			return nil, err
		}
		if !name.Valid || !value.Valid || vitalThresholds[name.String] == [2]float64{} {
			continue
		}
		if valuesByURL[pageURL] == nil {
			valuesByURL[pageURL] = map[string][]float64{}
		}
		valuesByURL[pageURL][name.String] = append(valuesByURL[pageURL][name.String], value.Float64)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()

	trafficQuery := `
	SELECT url, COUNT(*) AS pageviews
	FROM events
	WHERE event_name = '$pageview'
	  AND ` + siteMatchClause + `
	  AND ` + fromTimeClause + `
	  AND ` + toTimeClause + `
	GROUP BY url
	`
	trafficRows, err := r.db.QueryContext(ctx, trafficQuery, siteKey, siteKey, from, from, to, to, to, to)
	if err != nil {
		return nil, err
	}
	defer trafficRows.Close()

	trafficByURL := map[string]int{}
	for trafficRows.Next() {
		var pageURL string
		var pageviews int
		if err := trafficRows.Scan(&pageURL, &pageviews); err != nil {
			return nil, err
		}
		trafficByURL[pageURL] = pageviews
	}
	if err := trafficRows.Err(); err != nil {
		return nil, err
	}

	results := make([]core.PagePerformanceStat, 0, len(valuesByURL))
	for pageURL, metrics := range valuesByURL {
		result := core.PagePerformanceStat{
			URL:     pageURL,
			Traffic: trafficByURL[pageURL],
		}
		for name, values := range metrics {
			sort.Float64s(values)
			value := percentile75(values)
			switch name {
			case "LCP":
				result.LCP = float64Pointer(value)
			case "INP":
				result.INP = float64Pointer(value)
			case "CLS":
				result.CLS = float64Pointer(value)
			}
		}
		results = append(results, result)
	}

	sort.Slice(results, func(i, j int) bool {
		leftSeverity := pagePerformanceSeverity(results[i])
		rightSeverity := pagePerformanceSeverity(results[j])
		if leftSeverity != rightSeverity {
			return leftSeverity > rightSeverity
		}
		if results[i].Traffic != results[j].Traffic {
			return results[i].Traffic > results[j].Traffic
		}
		return results[i].URL < results[j].URL
	})

	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}
	return results, nil
}

func (r *SqliteRepository) GetPerformanceScore(ctx context.Context, siteKey, from, to string) (*core.PerformanceScore, error) {
	vitals, err := r.GetVitals(ctx, siteKey, from, to)
	if err != nil {
		return nil, err
	}
	distributions, err := r.GetVitalDistributions(ctx, siteKey, from, to)
	if err != nil {
		return nil, err
	}

	metricScores := map[string]int{}
	scoreTotal := 0
	for _, vital := range vitals {
		if vitalThresholds[vital.Name] == [2]float64{} {
			continue
		}
		score := vitalMetricScore(vital.Name, vital.Value)
		metricScores[vital.Name] = score
		scoreTotal += score
	}

	sampleSize := 0
	for _, distribution := range distributions {
		sampleSize += distribution.Total
	}

	overall := 0
	if len(metricScores) > 0 {
		overall = int(math.Round(float64(scoreTotal) / float64(len(metricScores))))
	}
	return &core.PerformanceScore{
		Score:        overall,
		Rating:       scoreRating(overall, len(metricScores) > 0),
		MetricScores: metricScores,
		SampleSize:   sampleSize,
	}, nil
}

func (r *SqliteRepository) GetCustomEvents(ctx context.Context, siteKey, from, to string) (*core.CustomEventsResult, error) {
	summaryQuery := `
	SELECT
		COUNT(*) AS total_events,
		COUNT(DISTINCT NULLIF(visitor_id, '')) AS unique_users,
		COUNT(DISTINCT NULLIF(session_id, '')) AS event_sessions
	FROM events
	WHERE event_name != ''
	  AND event_name NOT LIKE '$%'
	  AND ` + siteMatchClause + `
	  AND ` + fromTimeClause + `
	  AND ` + toTimeClause + `
	`
	var summary core.CustomEventSummary
	var eventSessions int
	if err := r.db.QueryRowContext(
		ctx,
		summaryQuery,
		siteKey,
		siteKey,
		from,
		from,
		to,
		to,
		to,
		to,
	).Scan(&summary.TotalEvents, &summary.UniqueUsers, &eventSessions); err != nil {
		return nil, err
	}

	stats, err := r.GetStats(ctx, siteKey, from, to)
	if err != nil {
		return nil, err
	}
	if stats.Sessions > 0 {
		summary.ConversionRate = math.Round((float64(eventSessions)/float64(stats.Sessions))*1000) / 10
	}

	eventsQuery := `
	SELECT
		event_name,
		COUNT(*) AS total_count,
		COUNT(DISTINCT NULLIF(visitor_id, '')) AS unique_users
	FROM events
	WHERE event_name != ''
	  AND event_name NOT LIKE '$%'
	  AND ` + siteMatchClause + `
	  AND ` + fromTimeClause + `
	  AND ` + toTimeClause + `
	GROUP BY event_name
	ORDER BY total_count DESC, event_name ASC
	`
	rows, err := r.db.QueryContext(ctx, eventsQuery, siteKey, siteKey, from, from, to, to, to, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := []core.CustomEventStat{}
	for rows.Next() {
		var event core.CustomEventStat
		if err := rows.Scan(&event.EventName, &event.TotalCount, &event.UniqueUsers); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &core.CustomEventsResult{Summary: summary, Events: events}, nil
}

func (r *SqliteRepository) GetCustomEventTimeSeries(
	ctx context.Context,
	siteKey, eventName, from, to string,
) ([]core.CustomEventTimeSeriesBucket, error) {
	query := `
	SELECT
		strftime('%Y-%m-%d', timestamp) AS day,
		COUNT(*) AS count
	FROM events
	WHERE event_name = ?
	  AND event_name NOT LIKE '$%'
	  AND ` + siteMatchClause + `
	  AND ` + fromTimeClause + `
	  AND ` + toTimeClause + `
	GROUP BY day
	ORDER BY day ASC
	`
	rows, err := r.db.QueryContext(ctx, query, eventName, siteKey, siteKey, from, from, to, to, to, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []core.CustomEventTimeSeriesBucket{}
	for rows.Next() {
		var bucket core.CustomEventTimeSeriesBucket
		if err := rows.Scan(&bucket.Date, &bucket.Count); err != nil {
			return nil, err
		}
		results = append(results, bucket)
	}
	return results, rows.Err()
}

func (r *SqliteRepository) GetPageviewsTimeSeries(ctx context.Context, siteKey, from, to string) ([]core.TimeSeriesBucket, error) {
	query := `
	SELECT
		strftime('%Y-%m-%d', timestamp) AS day,
		COUNT(*)                        AS pageviews
	FROM events
	WHERE event_name = '$pageview'
	  AND ` + siteMatchClause + `
	  AND ` + fromTimeClause + `
	  AND ` + toTimeClause + `
	GROUP BY day
	ORDER BY day ASC
	`
	rows, err := r.db.QueryContext(ctx, query, siteKey, siteKey, from, from, to, to, to, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []core.TimeSeriesBucket
	for rows.Next() {
		var b core.TimeSeriesBucket
		if err := rows.Scan(&b.Date, &b.Pageviews); err != nil {
			return nil, err
		}
		results = append(results, b)
	}
	return results, rows.Err()
}

func (r *SqliteRepository) GetUniqueVisitorsTimeSeries(ctx context.Context, siteKey, from, to string) ([]core.TimeSeriesBucket, error) {
	query := `
	SELECT
		strftime('%Y-%m-%d', timestamp) AS day,
		COUNT(DISTINCT visitor_id)      AS unique_visitors
	FROM events
	WHERE event_name = '$pageview'
	  AND ` + siteMatchClause + `
	  AND ` + fromTimeClause + `
	  AND ` + toTimeClause + `
	GROUP BY day
	ORDER BY day ASC
	`
	rows, err := r.db.QueryContext(ctx, query, siteKey, siteKey, from, from, to, to, to, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []core.TimeSeriesBucket
	for rows.Next() {
		var b core.TimeSeriesBucket
		if err := rows.Scan(&b.Date, &b.UniqueVisitors); err != nil {
			return nil, err
		}
		results = append(results, b)
	}
	return results, rows.Err()
}

func (r *SqliteRepository) GetSessionsTimeSeries(ctx context.Context, siteKey, from, to string) ([]core.TimeSeriesBucket, error) {
	query := `
	SELECT
		strftime('%Y-%m-%d', timestamp) AS day,
		COUNT(DISTINCT session_id)       AS sessions
	FROM events
	WHERE event_name = '$pageview'
	  AND ` + siteMatchClause + `
	  AND ` + fromTimeClause + `
	  AND ` + toTimeClause + `
	GROUP BY day
	ORDER BY day ASC
	`
	rows, err := r.db.QueryContext(ctx, query, siteKey, siteKey, from, from, to, to, to, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []core.TimeSeriesBucket
	for rows.Next() {
		var b core.TimeSeriesBucket
		if err := rows.Scan(&b.Date, &b.Sessions); err != nil {
			return nil, err
		}
		results = append(results, b)
	}
	return results, rows.Err()
}

func (r *SqliteRepository) GetDevices(ctx context.Context, siteKey, from, to string) ([]core.DeviceStat, error) {
	query := `
	SELECT
		CASE
			WHEN screen_width < 768  THEN 'Mobile'
			WHEN screen_width < 1024 THEN 'Tablet'
			ELSE 'Desktop'
		END AS device,
		COUNT(*) AS count
	FROM events
	WHERE event_name = '$pageview'
	  AND ` + siteMatchClause + `
	  AND ` + fromTimeClause + `
	  AND ` + toTimeClause + `
	GROUP BY device
	ORDER BY count DESC
	`
	rows, err := r.db.QueryContext(ctx, query, siteKey, siteKey, from, from, to, to, to, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []core.DeviceStat
	for rows.Next() {
		var s core.DeviceStat
		if err := rows.Scan(&s.Device, &s.Count); err != nil {
			return nil, err
		}
		results = append(results, s)
	}
	return results, rows.Err()
}

func (r *SqliteRepository) GetSites(ctx context.Context) ([]core.SiteStat, error) {
	query := `
	SELECT
		COALESCE(NULLIF(site_id, ''), domain) AS effective_site_id,
		MIN(domain) AS primary_domain,
		GROUP_CONCAT(DISTINCT domain) AS domains_csv
	FROM events
	WHERE domain != ''
	GROUP BY effective_site_id
	ORDER BY effective_site_id ASC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []core.SiteStat{}
	for rows.Next() {
		var s core.SiteStat
		var domainsCSV string
		if err := rows.Scan(&s.SiteID, &s.Domain, &domainsCSV); err != nil {
			return nil, err
		}
		s.Domains = splitDomains(domainsCSV)
		results = append(results, s)
	}
	return results, rows.Err()
}

func normalizeReferrer(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	parsed, err := url.Parse(raw)
	if err == nil && parsed.Hostname() != "" {
		return strings.TrimPrefix(strings.ToLower(parsed.Hostname()), "www.")
	}

	return strings.TrimPrefix(strings.ToLower(raw), "www.")
}

func percentile75(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	index := int(math.Ceil(0.75*float64(len(values)))) - 1
	if index < 0 {
		index = 0
	}
	if index >= len(values) {
		index = len(values) - 1
	}
	return values[index]
}

var vitalThresholds = map[string][2]float64{
	"LCP": {2500, 4000},
	"INP": {200, 500},
	"CLS": {0.1, 0.25},
}

func classifyVital(name string, value float64) string {
	thresholds, ok := vitalThresholds[name]
	if !ok {
		return ""
	}
	if value <= thresholds[0] {
		return "good"
	}
	if value <= thresholds[1] {
		return "needs-improvement"
	}
	return "poor"
}

func pagePerformanceSeverity(page core.PagePerformanceStat) int {
	severity := 0
	for name, value := range map[string]*float64{
		"LCP": page.LCP,
		"INP": page.INP,
		"CLS": page.CLS,
	} {
		if value == nil {
			continue
		}
		switch classifyVital(name, *value) {
		case "poor":
			severity += 2
		case "needs-improvement":
			severity++
		}
	}
	return severity
}

func vitalMetricScore(name string, value float64) int {
	thresholds, ok := vitalThresholds[name]
	if !ok || value < 0 {
		return 0
	}

	good := thresholds[0]
	poor := thresholds[1]
	var score float64
	switch {
	case value <= good:
		score = 100 - 10*(value/good)
	case value <= poor:
		score = 90 - 40*((value-good)/(poor-good))
	default:
		score = 50 - 50*((value-poor)/poor)
	}
	return int(math.Round(math.Max(0, math.Min(100, score))))
}

func scoreRating(score int, hasData bool) string {
	if !hasData {
		return "unknown"
	}
	if score >= 90 {
		return "good"
	}
	if score >= 50 {
		return "needs-improvement"
	}
	return "poor"
}

func float64Pointer(value float64) *float64 {
	return &value
}

func splitDomains(csv string) []string {
	if csv == "" {
		return nil
	}

	parts := strings.Split(csv, ",")
	domains := make([]string, 0, len(parts))
	seen := map[string]struct{}{}
	for _, part := range parts {
		domain := strings.TrimSpace(part)
		if domain == "" {
			continue
		}
		if _, ok := seen[domain]; ok {
			continue
		}
		seen[domain] = struct{}{}
		domains = append(domains, domain)
	}
	sort.Strings(domains)
	return domains
}
