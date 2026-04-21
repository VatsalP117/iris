package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/VatsalP117/iris/pkg/core"
	"github.com/google/uuid"
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

func (h *Handler) TrackEvent(w http.ResponseWriter, r *http.Request) {
	var event core.Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		log.Printf("[TrackEvent] JSON decode error: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	event.ID = uuid.NewString()
	event.Timestamp = time.Now().UTC()

	if event.Properties != nil {
		event.Properties = truncateStrings(event.Properties, 200).(map[string]any)
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
		events[i].ID = uuid.NewString()
		events[i].Timestamp = now
		if events[i].Properties != nil {
			events[i].Properties = truncateStrings(events[i].Properties, 200).(map[string]any)
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

func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	q, ok := parseStatsQuery(w, r)
	if !ok {
		return
	}
	result, err := h.Repo.GetStats(r.Context(), q.SiteID, q.From, q.To)
	if err != nil {
		http.Error(w, "Query failed", http.StatusInternalServerError)
		return
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
		http.Error(w, "Query failed", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) ListSites(w http.ResponseWriter, r *http.Request) {
	result, err := h.Repo.GetSites(r.Context())
	if err != nil {
		http.Error(w, "Query failed", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, result)
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
