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
