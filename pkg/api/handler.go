package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/VatsalP117/iris/pkg/core"
)

type Handler struct {
	Repo core.EventRepository
}

func NewHandler(repo core.EventRepository) *Handler {
	return &Handler{Repo: repo}
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

type statsQuery struct {
	SiteID string
	From   string
	To     string
}

func parseStatsQuery(w http.ResponseWriter, r *http.Request) (statsQuery, bool) {
	q := r.URL.Query()
	siteID := q.Get("site_id")
	if siteID == "" {
		siteID = q.Get("domain")
	}
	if siteID == "" {
		http.Error(w, "site_id is required", http.StatusBadRequest)
		return statsQuery{}, false
	}
	return statsQuery{SiteID: siteID, From: q.Get("from"), To: q.Get("to")}, true
}

const maxBodyBytes = 1 << 20 // 1 MiB

func (h *Handler) TrackEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	var event core.Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		log.Printf("[TrackEvent] JSON decode error: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.prepareIncomingEvent(r.Context(), &event, time.Now().UTC()); err != nil {
		writeIngestError(w, err)
		return
	}

	if err := h.Repo.Insert(r.Context(), &event); err != nil {
		log.Printf("[TrackEvent] DB Insert error: %v", err)
		http.Error(w, "Failed to save event", http.StatusInternalServerError)
		return
	}

	log.Printf("[TrackEvent] OK: %s (domain=%s, site=%s, url=%s)", event.EventName, event.Domain, event.SiteID, event.URL)
	w.WriteHeader(http.StatusAccepted)
}

const maxBatchSize = 50

func (h *Handler) TrackBatchEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	var events []core.Event
	if err := json.NewDecoder(r.Body).Decode(&events); err != nil {
		log.Printf("[TrackBatchEvents] JSON decode error: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if len(events) == 0 {
		w.WriteHeader(http.StatusAccepted)
		return
	}

	if len(events) > maxBatchSize {
		http.Error(w, "Batch too large", http.StatusRequestEntityTooLarge)
		return
	}

	now := time.Now().UTC()
	ptrs := make([]*core.Event, len(events))
	for i := range events {
		if err := h.prepareIncomingEvent(r.Context(), &events[i], now); err != nil {
			writeIngestError(w, fmt.Errorf("event %d: %w", i, err))
			return
		}
		ptrs[i] = &events[i]
	}

	if err := h.Repo.InsertBatch(r.Context(), ptrs); err != nil {
		log.Printf("[TrackBatchEvents] DB InsertBatch error: %v", err)
		http.Error(w, "Failed to save events", http.StatusInternalServerError)
		return
	}

	log.Printf("[TrackBatchEvents] OK: %d events ingested", len(events))
	w.WriteHeader(http.StatusAccepted)
}

func writeIngestError(w http.ResponseWriter, err error) {
	status := http.StatusBadRequest
	if errors.Is(err, core.ErrSiteNotFound) {
		status = http.StatusNotFound
	} else if errors.Is(err, core.ErrDomainNotAllowed) {
		status = http.StatusForbidden
	}
	http.Error(w, err.Error(), status)
}

func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	q, ok := parseStatsQuery(w, r)
	if !ok {
		return
	}
	result, err := h.Repo.GetStats(r.Context(), q.SiteID, q.From, q.To)
	if err != nil {
		log.Printf("[GetStats] query error: %v", err)
		http.Error(w, "Query failed", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) GetSiteTrends(w http.ResponseWriter, r *http.Request) {
	q, ok := parseStatsQuery(w, r)
	if !ok {
		return
	}

	current, err := h.Repo.GetStats(r.Context(), q.SiteID, q.From, q.To)
	if err != nil {
		log.Printf("[GetSiteTrends] current-period query error: %v", err)
		http.Error(w, "Query failed", http.StatusInternalServerError)
		return
	}

	result := core.SiteTrendResult{Current: *current}
	previousFrom, previousTo, hasPrevious := previousPeriod(q.From, q.To)
	if hasPrevious {
		previous, queryErr := h.Repo.GetStats(r.Context(), q.SiteID, previousFrom, previousTo)
		if queryErr != nil {
			log.Printf("[GetSiteTrends] previous-period query error: %v", queryErr)
			http.Error(w, "Query failed", http.StatusInternalServerError)
			return
		}
		result.Previous = *previous
		result.Change = core.StatsChange{
			Pageviews:      percentChange(current.Pageviews, previous.Pageviews),
			UniqueVisitors: percentChange(current.UniqueVisitors, previous.UniqueVisitors),
			Sessions:       percentChange(current.Sessions, previous.Sessions),
		}
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) GetPages(w http.ResponseWriter, r *http.Request) {
	q, ok := parseStatsQuery(w, r)
	if !ok {
		return
	}
	result, err := h.Repo.GetTopPages(r.Context(), q.SiteID, q.From, q.To, 10)
	if err != nil {
		log.Printf("[GetPages] query error: %v", err)
		http.Error(w, "Query failed", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) GetReferrers(w http.ResponseWriter, r *http.Request) {
	q, ok := parseStatsQuery(w, r)
	if !ok {
		return
	}
	result, err := h.Repo.GetTopReferrers(r.Context(), q.SiteID, q.From, q.To, 10)
	if err != nil {
		log.Printf("[GetReferrers] query error: %v", err)
		http.Error(w, "Query failed", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) GetVitals(w http.ResponseWriter, r *http.Request) {
	q, ok := parseStatsQuery(w, r)
	if !ok {
		return
	}
	result, err := h.Repo.GetVitals(r.Context(), q.SiteID, q.From, q.To)
	if err != nil {
		log.Printf("[GetVitals] query error: %v", err)
		http.Error(w, "Query failed", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) GetVitalDistributions(w http.ResponseWriter, r *http.Request) {
	q, ok := parseStatsQuery(w, r)
	if !ok {
		return
	}
	result, err := h.Repo.GetVitalDistributions(r.Context(), q.SiteID, q.From, q.To)
	if err != nil {
		log.Printf("[GetVitalDistributions] query error: %v", err)
		http.Error(w, "Query failed", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) GetPagePerformance(w http.ResponseWriter, r *http.Request) {
	q, ok := parseStatsQuery(w, r)
	if !ok {
		return
	}
	result, err := h.Repo.GetPagePerformance(r.Context(), q.SiteID, q.From, q.To, 20)
	if err != nil {
		log.Printf("[GetPagePerformance] query error: %v", err)
		http.Error(w, "Query failed", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) GetPerformanceScore(w http.ResponseWriter, r *http.Request) {
	q, ok := parseStatsQuery(w, r)
	if !ok {
		return
	}
	result, err := h.Repo.GetPerformanceScore(r.Context(), q.SiteID, q.From, q.To)
	if err != nil {
		log.Printf("[GetPerformanceScore] query error: %v", err)
		http.Error(w, "Query failed", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) GetCustomEvents(w http.ResponseWriter, r *http.Request) {
	q, ok := parseStatsQuery(w, r)
	if !ok {
		return
	}

	result, err := h.Repo.GetCustomEvents(r.Context(), q.SiteID, q.From, q.To)
	if err != nil {
		log.Printf("[GetCustomEvents] current-period query error: %v", err)
		http.Error(w, "Query failed", http.StatusInternalServerError)
		return
	}

	previousFrom, previousTo, hasPrevious := previousPeriod(q.From, q.To)
	if hasPrevious {
		previous, queryErr := h.Repo.GetCustomEvents(r.Context(), q.SiteID, previousFrom, previousTo)
		if queryErr != nil {
			log.Printf("[GetCustomEvents] previous-period query error: %v", queryErr)
			http.Error(w, "Query failed", http.StatusInternalServerError)
			return
		}

		result.Summary.ChangePercent = percentChange(result.Summary.TotalEvents, previous.Summary.TotalEvents)
		previousCounts := make(map[string]int, len(previous.Events))
		for _, event := range previous.Events {
			previousCounts[event.EventName] = event.TotalCount
		}
		for i := range result.Events {
			result.Events[i].ChangePercent = percentChange(
				result.Events[i].TotalCount,
				previousCounts[result.Events[i].EventName],
			)
		}
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) GetCustomEventTimeSeries(w http.ResponseWriter, r *http.Request) {
	q, ok := parseStatsQuery(w, r)
	if !ok {
		return
	}
	eventName := strings.TrimSpace(r.URL.Query().Get("event_name"))
	if eventName == "" || strings.HasPrefix(eventName, "$") {
		http.Error(w, "valid custom event_name is required", http.StatusBadRequest)
		return
	}

	result, err := h.Repo.GetCustomEventTimeSeries(r.Context(), q.SiteID, eventName, q.From, q.To)
	if err != nil {
		log.Printf("[GetCustomEventTimeSeries] query error: %v", err)
		http.Error(w, "Query failed", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) GetDevices(w http.ResponseWriter, r *http.Request) {
	q, ok := parseStatsQuery(w, r)
	if !ok {
		return
	}
	result, err := h.Repo.GetDevices(r.Context(), q.SiteID, q.From, q.To)
	if err != nil {
		log.Printf("[GetDevices] query error: %v", err)
		http.Error(w, "Query failed", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) GetTimeSeries(w http.ResponseWriter, r *http.Request) {
	q, ok := parseStatsQuery(w, r)
	if !ok {
		return
	}
	result, err := h.Repo.GetPageviewsTimeSeries(r.Context(), q.SiteID, q.From, q.To)
	if err != nil {
		log.Printf("[GetTimeSeries] query error: %v", err)
		http.Error(w, "Query failed", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) GetUniqueVisitorsTimeSeries(w http.ResponseWriter, r *http.Request) {
	q, ok := parseStatsQuery(w, r)
	if !ok {
		return
	}
	result, err := h.Repo.GetUniqueVisitorsTimeSeries(r.Context(), q.SiteID, q.From, q.To)
	if err != nil {
		log.Printf("[GetUniqueVisitorsTimeSeries] query error: %v", err)
		http.Error(w, "Query failed", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) GetSessionsTimeSeries(w http.ResponseWriter, r *http.Request) {
	q, ok := parseStatsQuery(w, r)
	if !ok {
		return
	}
	result, err := h.Repo.GetSessionsTimeSeries(r.Context(), q.SiteID, q.From, q.To)
	if err != nil {
		log.Printf("[GetSessionsTimeSeries] query error: %v", err)
		http.Error(w, "Query failed", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) ListSites(w http.ResponseWriter, r *http.Request) {
	result, err := h.Repo.GetSites(r.Context())
	if err != nil {
		log.Printf("[ListSites] query error: %v", err)
		http.Error(w, "Query failed", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) Sites(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.ListSites(w, r)
	case http.MethodPost:
		r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
		var site core.Site
		if err := json.NewDecoder(r.Body).Decode(&site); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		if len(site.ID) > maxIdentifierLength || len(site.Name) > 200 || len(site.Domains) > 20 {
			http.Error(w, "Invalid site configuration", http.StatusBadRequest)
			return
		}
		if err := h.Repo.CreateSite(r.Context(), &site); err != nil {
			log.Printf("[Sites] create error: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		writeJSON(w, http.StatusCreated, site)
	default:
		w.Header().Set("Allow", http.MethodGet+", "+http.MethodPost)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) Status(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	status, err := h.Repo.GetSystemStatus(r.Context())
	if err != nil {
		log.Printf("[Status] database error: %v", err)
		http.Error(w, "Unavailable", http.StatusServiceUnavailable)
		return
	}
	writeJSON(w, http.StatusOK, status)
}

func previousPeriod(from, to string) (string, string, bool) {
	start, ok := parsePeriodTime(from, false)
	if !ok {
		return "", "", false
	}
	end, ok := parsePeriodTime(to, true)
	if !ok || !end.After(start) {
		return "", "", false
	}

	duration := end.Sub(start)
	previousEnd := start.Add(-time.Nanosecond)
	previousStart := previousEnd.Add(-duration)
	return previousStart.UTC().Format(time.RFC3339Nano), previousEnd.UTC().Format(time.RFC3339Nano), true
}

func parsePeriodTime(value string, endOfDate bool) (time.Time, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, false
	}

	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02T15:04:05-07:00"} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed, true
		}
	}

	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return time.Time{}, false
	}
	if endOfDate {
		parsed = parsed.Add(24*time.Hour - time.Nanosecond)
	}
	return parsed, true
}

func percentChange(current, previous int) float64 {
	if previous == 0 {
		if current == 0 {
			return 0
		}
		return 100
	}
	change := (float64(current-previous) / float64(previous)) * 100
	return math.Round(change*10) / 10
}

func truncateStrings(data any, maxLen int) any {
	switch v := data.(type) {
	case string:
		if len(v) > maxLen {
			return v[:maxLen] + "..."
		}
		return v
	case map[string]any:
		out := make(map[string]any, len(v))
		for key, val := range v {
			out[key] = truncateStrings(val, maxLen)
		}
		return out
	case []any:
		out := make([]any, len(v))
		for i, val := range v {
			out[i] = truncateStrings(val, maxLen)
		}
		return out
	default:
		return data
	}
}
