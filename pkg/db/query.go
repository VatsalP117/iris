package db

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/VatsalP117/iris/pkg/core"
)

func (r *SqliteRepository) analyticsWindow(
	ctx context.Context,
	siteID, from, to string,
) (string, []any, error) {
	location := time.UTC
	if isDateOnly(from) || isDateOnly(to) {
		var timezone string
		if err := r.db.QueryRowContext(ctx, `
			SELECT COALESCE((SELECT NULLIF(timezone, '') FROM sites WHERE id = ?), 'UTC')
		`, siteID).Scan(&timezone); err != nil {
			return "", nil, fmt.Errorf("get site timezone: %w", err)
		}
		var err error
		location, err = time.LoadLocation(timezone)
		if err != nil {
			return "", nil, fmt.Errorf("load site timezone %q: %w", timezone, err)
		}
	}

	clause := ""
	args := []any{}
	if from != "" {
		value, err := parseAnalyticsTime(from, false, location)
		if err != nil {
			return "", nil, fmt.Errorf("parse from time: %w", err)
		}
		clause += "\n\t  AND occurred_at_us >= ?"
		args = append(args, value.UnixMicro())
	}
	if to != "" {
		value, err := parseAnalyticsTime(to, true, location)
		if err != nil {
			return "", nil, fmt.Errorf("parse to time: %w", err)
		}
		clause += "\n\t  AND occurred_at_us <= ?"
		args = append(args, value.UnixMicro())
	}
	return clause, args, nil
}

func isDateOnly(value string) bool {
	return len(value) == len("2006-01-02")
}

func parseAnalyticsTime(value string, endOfDay bool, location *time.Location) (time.Time, error) {
	if isDateOnly(value) {
		parsed, err := time.ParseInLocation("2006-01-02", value, location)
		if err != nil {
			return time.Time{}, err
		}
		if endOfDay {
			return parsed.AddDate(0, 0, 1).Add(-time.Microsecond), nil
		}
		return parsed, nil
	}

	for _, layout := range []string{time.RFC3339Nano, "2006-01-02 15:04:05"} {
		parsed, err := time.ParseInLocation(layout, value, time.UTC)
		if err == nil {
			return parsed.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported time %q", value)
}

func (r *SqliteRepository) projectionDayWindow(
	ctx context.Context,
	from, to string,
) (string, []any, bool, error) {
	if (from != "" && !isDateOnly(from)) || (to != "" && !isDateOnly(to)) {
		return "", nil, false, nil
	}
	status, err := r.GetSystemStatus(ctx)
	if err != nil {
		return "", nil, false, err
	}
	if status.EventLastSeq == 0 || status.ProjectionLag != 0 ||
		status.ProjectionVersion != analyticsProjectionVersion {
		return "", nil, false, nil
	}
	clause := ""
	args := []any{}
	if from != "" {
		clause += " AND day >= ?"
		args = append(args, from)
	}
	if to != "" {
		clause += " AND day <= ?"
		args = append(args, to)
	}
	return clause, args, true, nil
}

func (r *SqliteRepository) GetStats(ctx context.Context, siteKey, from, to string) (*core.StatsResult, error) {
	timeClause, timeArgs, err := r.analyticsWindow(ctx, siteKey, from, to)
	if err != nil {
		return nil, err
	}
	query := `
	SELECT
		COUNT(*)                                          AS pageviews,
		COUNT(DISTINCT visitor_id)                        AS unique_visitors,
		COUNT(DISTINCT session_id)                        AS sessions
	FROM events
	WHERE event_name = '$pageview'
	  AND site_id = ?` + timeClause + `
	`
	args := append([]any{siteKey}, timeArgs...)
	row := r.db.QueryRowContext(ctx, query, args...)

	var res core.StatsResult
	if err := row.Scan(&res.Pageviews, &res.UniqueVisitors, &res.Sessions); err != nil {
		return nil, err
	}
	return &res, nil
}

func (r *SqliteRepository) GetTopPages(ctx context.Context, siteKey, from, to string, limit int) ([]core.PageStat, error) {
	if dayClause, dayArgs, ok, err := r.projectionDayWindow(ctx, from, to); err != nil {
		return nil, err
	} else if ok {
		query := `
			SELECT pathname, SUM(pageviews)
			FROM daily_page_metrics
			WHERE site_id = ?` + dayClause + `
			GROUP BY pathname
			ORDER BY SUM(pageviews) DESC, pathname ASC
			LIMIT ?
		`
		args := append([]any{siteKey}, dayArgs...)
		args = append(args, limit)
		rows, queryErr := r.db.QueryContext(ctx, query, args...)
		if queryErr != nil {
			return nil, queryErr
		}
		defer rows.Close()
		var results []core.PageStat
		for rows.Next() {
			var result core.PageStat
			if err := rows.Scan(&result.URL, &result.Pageviews); err != nil {
				return nil, err
			}
			results = append(results, result)
		}
		return results, rows.Err()
	}
	timeClause, timeArgs, err := r.analyticsWindow(ctx, siteKey, from, to)
	if err != nil {
		return nil, err
	}
	query := `
	SELECT pathname, COUNT(*) AS pageviews
	FROM events
	WHERE event_name = '$pageview'
	  AND site_id = ?` + timeClause + `
	GROUP BY pathname
	ORDER BY pageviews DESC
	LIMIT ?
	`
	args := append([]any{siteKey}, timeArgs...)
	args = append(args, limit)
	rows, err := r.db.QueryContext(ctx, query, args...)
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
	timeClause, timeArgs, err := r.analyticsWindow(ctx, siteKey, from, to)
	if err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = -1
	}
	query := `
	SELECT referrer_host, COUNT(DISTINCT NULLIF(visitor_id, '')) AS visitors
	FROM events
	WHERE event_name = '$pageview'
	  AND site_id = ?
	  AND referrer_host != ''` + timeClause + `
	GROUP BY referrer_host
	ORDER BY visitors DESC, referrer_host ASC
	LIMIT ?
	`
	args := append([]any{siteKey}, timeArgs...)
	args = append(args, limit)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []core.ReferrerStat{}
	for rows.Next() {
		var result core.ReferrerStat
		if err := rows.Scan(&result.Referrer, &result.Visitors); err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

func (r *SqliteRepository) GetVitals(ctx context.Context, siteKey, from, to string) ([]core.VitalStat, error) {
	timeClause, timeArgs, err := r.analyticsWindow(ctx, siteKey, from, to)
	if err != nil {
		return nil, err
	}
	query := `
	SELECT
		json_extract(properties, '$.$name') AS name,
		CAST(json_extract(properties, '$.$val') AS REAL) AS value
	FROM events
	WHERE event_name = '$web_vital'
	  AND site_id = ?` + timeClause + `
	`
	args := append([]any{siteKey}, timeArgs...)
	rows, err := r.db.QueryContext(ctx, query, args...)
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
	timeClause, timeArgs, err := r.analyticsWindow(ctx, siteKey, from, to)
	if err != nil {
		return nil, err
	}
	query := `
	SELECT
		json_extract(properties, '$.$name') AS name,
		CAST(json_extract(properties, '$.$val') AS REAL) AS value
	FROM events
	WHERE event_name = '$web_vital'
	  AND site_id = ?` + timeClause + `
	`
	args := append([]any{siteKey}, timeArgs...)
	rows, err := r.db.QueryContext(ctx, query, args...)
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
	timeClause, timeArgs, err := r.analyticsWindow(ctx, siteKey, from, to)
	if err != nil {
		return nil, err
	}
	vitalsQuery := `
	SELECT
		pathname,
		json_extract(properties, '$.$name') AS name,
		CAST(json_extract(properties, '$.$val') AS REAL) AS value
	FROM events
	WHERE event_name = '$web_vital'
	  AND pathname != ''
	  AND site_id = ?` + timeClause + `
	`
	args := append([]any{siteKey}, timeArgs...)
	rows, err := r.db.QueryContext(ctx, vitalsQuery, args...)
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
	SELECT pathname, COUNT(*) AS pageviews
	FROM events
	WHERE event_name = '$pageview'
	  AND site_id = ?` + timeClause + `
	GROUP BY pathname
	`
	trafficRows, err := r.db.QueryContext(ctx, trafficQuery, args...)
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
	timeClause, timeArgs, err := r.analyticsWindow(ctx, siteKey, from, to)
	if err != nil {
		return nil, err
	}
	summaryQuery := `
	SELECT
		COUNT(*) AS total_events,
		COUNT(DISTINCT NULLIF(visitor_id, '')) AS unique_users,
		COUNT(DISTINCT NULLIF(session_id, '')) AS event_sessions
	FROM events
	WHERE event_name != ''
	  AND event_name NOT LIKE '$%'
	  AND site_id = ?` + timeClause + `
	`
	var summary core.CustomEventSummary
	var eventSessions int
	args := append([]any{siteKey}, timeArgs...)
	if err := r.db.QueryRowContext(ctx, summaryQuery, args...).Scan(
		&summary.TotalEvents, &summary.UniqueUsers, &eventSessions,
	); err != nil {
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
	  AND site_id = ?` + timeClause + `
	GROUP BY event_name
	ORDER BY total_count DESC, event_name ASC
	`
	rows, err := r.db.QueryContext(ctx, eventsQuery, args...)
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
	timeClause, timeArgs, err := r.analyticsWindow(ctx, siteKey, from, to)
	if err != nil {
		return nil, err
	}
	query := `
	SELECT
		local_day AS day,
		COUNT(*) AS count
	FROM events
	WHERE event_name = ?
	  AND event_name NOT LIKE '$%'
	  AND site_id = ?` + timeClause + `
	GROUP BY day
	ORDER BY day ASC
	`
	args := append([]any{eventName, siteKey}, timeArgs...)
	rows, err := r.db.QueryContext(ctx, query, args...)
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
	if dayClause, dayArgs, ok, err := r.projectionDayWindow(ctx, from, to); err != nil {
		return nil, err
	} else if ok {
		return r.projectedTimeSeries(ctx, "daily_site_metrics", "SUM(pageviews)", siteKey, dayClause, dayArgs)
	}
	timeClause, timeArgs, err := r.analyticsWindow(ctx, siteKey, from, to)
	if err != nil {
		return nil, err
	}
	query := `
	SELECT
		local_day AS day,
		COUNT(*)                        AS pageviews
	FROM events
	WHERE event_name = '$pageview'
	  AND site_id = ?` + timeClause + `
	GROUP BY day
	ORDER BY day ASC
	`
	args := append([]any{siteKey}, timeArgs...)
	rows, err := r.db.QueryContext(ctx, query, args...)
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
	if dayClause, dayArgs, ok, err := r.projectionDayWindow(ctx, from, to); err != nil {
		return nil, err
	} else if ok {
		return r.projectedTimeSeries(ctx, "daily_visitors", "COUNT(*)", siteKey, dayClause, dayArgs)
	}
	timeClause, timeArgs, err := r.analyticsWindow(ctx, siteKey, from, to)
	if err != nil {
		return nil, err
	}
	query := `
	SELECT
		local_day AS day,
		COUNT(DISTINCT visitor_id)      AS unique_visitors
	FROM events
	WHERE event_name = '$pageview'
	  AND site_id = ?` + timeClause + `
	GROUP BY day
	ORDER BY day ASC
	`
	args := append([]any{siteKey}, timeArgs...)
	rows, err := r.db.QueryContext(ctx, query, args...)
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
	if dayClause, dayArgs, ok, err := r.projectionDayWindow(ctx, from, to); err != nil {
		return nil, err
	} else if ok {
		return r.projectedTimeSeries(ctx, "daily_sessions", "COUNT(*)", siteKey, dayClause, dayArgs)
	}
	timeClause, timeArgs, err := r.analyticsWindow(ctx, siteKey, from, to)
	if err != nil {
		return nil, err
	}
	query := `
	SELECT
		local_day AS day,
		COUNT(DISTINCT session_id)       AS sessions
	FROM events
	WHERE event_name = '$pageview'
	  AND site_id = ?` + timeClause + `
	GROUP BY day
	ORDER BY day ASC
	`
	args := append([]any{siteKey}, timeArgs...)
	rows, err := r.db.QueryContext(ctx, query, args...)
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

func (r *SqliteRepository) projectedTimeSeries(
	ctx context.Context,
	table, aggregate, siteID, dayClause string,
	dayArgs []any,
) ([]core.TimeSeriesBucket, error) {
	query := "SELECT day, " + aggregate + " FROM " + table +
		" WHERE site_id = ?" + dayClause + " GROUP BY day ORDER BY day ASC"
	args := append([]any{siteID}, dayArgs...)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []core.TimeSeriesBucket
	for rows.Next() {
		var bucket core.TimeSeriesBucket
		var value int
		if err := rows.Scan(&bucket.Date, &value); err != nil {
			return nil, err
		}
		switch table {
		case "daily_site_metrics":
			bucket.Pageviews = value
		case "daily_visitors":
			bucket.UniqueVisitors = value
		case "daily_sessions":
			bucket.Sessions = value
		}
		results = append(results, bucket)
	}
	return results, rows.Err()
}

func (r *SqliteRepository) GetDevices(ctx context.Context, siteKey, from, to string) ([]core.DeviceStat, error) {
	timeClause, timeArgs, err := r.analyticsWindow(ctx, siteKey, from, to)
	if err != nil {
		return nil, err
	}
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
	  AND site_id = ?` + timeClause + `
	GROUP BY device
	ORDER BY count DESC
	`
	args := append([]any{siteKey}, timeArgs...)
	rows, err := r.db.QueryContext(ctx, query, args...)
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
		s.id,
		s.name,
		COALESCE(MAX(CASE WHEN d.is_primary = 1 THEN d.hostname END), MIN(d.hostname), ''),
		COALESCE(GROUP_CONCAT(d.hostname), ''),
		s.timezone,
		s.retention_days
	FROM sites s
	LEFT JOIN site_domains d ON d.site_id = s.id
	WHERE s.disabled_at_us IS NULL
	GROUP BY s.id, s.name, s.timezone, s.retention_days
	ORDER BY s.id ASC
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
		if err := rows.Scan(
			&s.SiteID, &s.Name, &s.Domain, &domainsCSV, &s.Timezone, &s.RetentionDays,
		); err != nil {
			return nil, err
		}
		s.Domains = splitDomains(domainsCSV)
		results = append(results, s)
	}
	return results, rows.Err()
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
