package db

import (
	"context"

	"github.com/VatsalP117/iris/pkg/core"
)

//
// Helper to build WHERE clauses consistently.
// All query methods accept domain, from, to as strings (ISO8601 dates or empty).
//

func (r *SqliteRepository) GetStats(ctx context.Context, domain, from, to string) (*core.StatsResult, error) {
	query := `
	SELECT
		COUNT(*)                                          AS pageviews,
		COUNT(DISTINCT visitor_id)                        AS unique_visitors,
		COUNT(DISTINCT session_id)                        AS sessions
	FROM events
	WHERE event_name = '$pageview'
	  AND domain = ?
	  AND (? = '' OR timestamp >= ?)
	  AND (? = '' OR timestamp <= ?)
	`
	row := r.db.QueryRowContext(ctx, query, domain, from, from, to, to)

	var res core.StatsResult
	if err := row.Scan(&res.Pageviews, &res.UniqueVisitors, &res.Sessions); err != nil {
		return nil, err
	}
	return &res, nil
}

func (r *SqliteRepository) GetTopPages(ctx context.Context, domain, from, to string, limit int) ([]core.PageStat, error) {
	query := `
	SELECT url, COUNT(*) AS pageviews
	FROM events
	WHERE event_name = '$pageview'
	  AND domain = ?
	  AND (? = '' OR timestamp >= ?)
	  AND (? = '' OR timestamp <= ?)
	GROUP BY url
	ORDER BY pageviews DESC
	LIMIT ?
	`
	rows, err := r.db.QueryContext(ctx, query, domain, from, from, to, to, limit)
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

func (r *SqliteRepository) GetTopReferrers(ctx context.Context, domain, from, to string, limit int) ([]core.ReferrerStat, error) {
	query := `
	SELECT referrer, COUNT(DISTINCT visitor_id) AS visitors
	FROM events
	WHERE event_name = '$pageview'
	  AND domain = ?
	  AND referrer != ''
	  AND (? = '' OR timestamp >= ?)
	  AND (? = '' OR timestamp <= ?)
	GROUP BY referrer
	ORDER BY visitors DESC
	LIMIT ?
	`
	rows, err := r.db.QueryContext(ctx, query, domain, from, from, to, to, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []core.ReferrerStat
	for rows.Next() {
		var s core.ReferrerStat
		if err := rows.Scan(&s.Referrer, &s.Visitors); err != nil {
			return nil, err
		}
		results = append(results, s)
	}
	return results, rows.Err()
}

func (r *SqliteRepository) GetVitals(ctx context.Context, domain, from, to string) ([]core.VitalStat, error) {
	// Web vitals are stored as custom events with event_name = '$web_vital'
	// properties JSON contains: $name (CLS/INP/LCP), $val (float)
	query := `
	SELECT
		json_extract(properties, '$.$name') AS name,
		AVG(CAST(json_extract(properties, '$.$val') AS REAL)) AS avg_value
	FROM events
	WHERE event_name = '$web_vital'
	  AND domain = ?
	  AND (? = '' OR timestamp >= ?)
	  AND (? = '' OR timestamp <= ?)
	GROUP BY name
	`
	rows, err := r.db.QueryContext(ctx, query, domain, from, from, to, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []core.VitalStat
	for rows.Next() {
		var s core.VitalStat
		if err := rows.Scan(&s.Name, &s.Value); err != nil {
			return nil, err
		}
		results = append(results, s)
	}
	return results, rows.Err()
}

func (r *SqliteRepository) GetPageviewsTimeSeries(ctx context.Context, domain, from, to string) ([]core.TimeSeriesBucket, error) {
	// Group pageviews by calendar day (UTC) within the requested window.
	query := `
	SELECT
		strftime('%Y-%m-%d', timestamp) AS day,
		COUNT(*)                        AS pageviews
	FROM events
	WHERE event_name = '$pageview'
	  AND domain = ?
	  AND (? = '' OR timestamp >= ?)
	  AND (? = '' OR timestamp <= ?)
	GROUP BY day
	ORDER BY day ASC
	`
	rows, err := r.db.QueryContext(ctx, query, domain, from, from, to, to)
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

func (r *SqliteRepository) GetDevices(ctx context.Context, domain, from, to string) ([]core.DeviceStat, error) {
	// Bucket screen widths into device categories
	query := `
	SELECT
		CASE
			WHEN screen_width < 768  THEN 'Mobile'
			WHEN screen_width < 1024 THEN 'Tablet'
			ELSE 'Desktop'
		END AS device,
		COUNT(*) AS count
	FROM events
	WHERE domain = ?
	  AND (? = '' OR timestamp >= ?)
	  AND (? = '' OR timestamp <= ?)
	GROUP BY device
	ORDER BY count DESC
	`
	rows, err := r.db.QueryContext(ctx, query, domain, from, from, to, to)
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
