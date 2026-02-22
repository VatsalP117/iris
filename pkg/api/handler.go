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

func CORSMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

type statsQuery struct {
	Domain string
	From   string
	To     string
}

func parseStatsQuery(w http.ResponseWriter, r *http.Request) (statsQuery, bool) {
	q := r.URL.Query()
	domain := q.Get("domain")
	if domain == "" {
		http.Error(w, "domain is required", http.StatusBadRequest)
		return statsQuery{}, false
	}
	return statsQuery{Domain: domain, From: q.Get("from"), To: q.Get("to")}, true
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
		truncateStrings(event.Properties, 200)
	}

	if err := h.Repo.Insert(r.Context(), &event); err != nil {
		log.Printf("[TrackEvent] DB Insert error: %v", err)
		http.Error(w, "Failed to save event", http.StatusInternalServerError)
		return
	}

	log.Printf("[TrackEvent] OK: %s (domain=%s, site=%s, url=%s)", event.EventName, event.Domain, event.SiteID, event.URL)
	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	q, ok := parseStatsQuery(w, r)
	if !ok {
		return
	}
	result, err := h.Repo.GetStats(r.Context(), q.Domain, q.From, q.To)
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
	result, err := h.Repo.GetTopPages(r.Context(), q.Domain, q.From, q.To, 10)
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
	result, err := h.Repo.GetTopReferrers(r.Context(), q.Domain, q.From, q.To, 10)
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
	result, err := h.Repo.GetVitals(r.Context(), q.Domain, q.From, q.To)
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
	result, err := h.Repo.GetDevices(r.Context(), q.Domain, q.From, q.To)
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
	result, err := h.Repo.GetPageviewsTimeSeries(r.Context(), q.Domain, q.From, q.To)
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

func truncateStrings(data any, maxLen int) {
	switch v := data.(type) {
	case map[string]any:
		for key, val := range v {
			if str, ok := val.(string); ok && len(str) > maxLen {
				v[key] = str[:maxLen] + "..."
			} else {
				truncateStrings(val, maxLen)
			}
		}
	case []any:
		for i, val := range v {
			if str, ok := val.(string); ok && len(str) > maxLen {
				v[i] = str[:maxLen] + "..."
			} else {
				truncateStrings(val, maxLen)
			}
		}
	}
}
