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

func setCORSHeaders(w http.ResponseWriter, r *http.Request) bool {
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
		return true
	}
	return false
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// POST /api/event
func (h *Handler) TrackEvent(w http.ResponseWriter, r *http.Request) {
	if setCORSHeaders(w, r) {
		return
	}

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

	log.Printf("[TrackEvent] OK Event received: %s (domain=%s, site=%s, url=%s)", event.EventName, event.Domain, event.SiteID, event.URL)

	w.WriteHeader(http.StatusAccepted)
}

// GET /api/stats?domain=&from=&to=
func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	if setCORSHeaders(w, r) {
		return
	}
	q := r.URL.Query()
	domain := q.Get("domain")
	if domain == "" {
		http.Error(w, "domain is required", http.StatusBadRequest)
		return
	}
	result, err := h.Repo.GetStats(r.Context(), domain, q.Get("from"), q.Get("to"))
	if err != nil {
		http.Error(w, "Query failed", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// GET /api/pages?domain=&from=&to=&limit=
func (h *Handler) GetPages(w http.ResponseWriter, r *http.Request) {
	if setCORSHeaders(w, r) {
		return
	}
	q := r.URL.Query()
	domain := q.Get("domain")
	if domain == "" {
		http.Error(w, "domain is required", http.StatusBadRequest)
		return
	}
	limit := 10
	result, err := h.Repo.GetTopPages(r.Context(), domain, q.Get("from"), q.Get("to"), limit)
	if err != nil {
		http.Error(w, "Query failed", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// GET /api/referrers?domain=&from=&to=
func (h *Handler) GetReferrers(w http.ResponseWriter, r *http.Request) {
	if setCORSHeaders(w, r) {
		return
	}
	q := r.URL.Query()
	domain := q.Get("domain")
	if domain == "" {
		http.Error(w, "domain is required", http.StatusBadRequest)
		return
	}
	result, err := h.Repo.GetTopReferrers(r.Context(), domain, q.Get("from"), q.Get("to"), 10)
	if err != nil {
		http.Error(w, "Query failed", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// GET /api/vitals?domain=&from=&to=
func (h *Handler) GetVitals(w http.ResponseWriter, r *http.Request) {
	if setCORSHeaders(w, r) {
		return
	}
	q := r.URL.Query()
	domain := q.Get("domain")
	if domain == "" {
		http.Error(w, "domain is required", http.StatusBadRequest)
		return
	}
	result, err := h.Repo.GetVitals(r.Context(), domain, q.Get("from"), q.Get("to"))
	if err != nil {
		http.Error(w, "Query failed", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// GET /api/devices?domain=&from=&to=
func (h *Handler) GetDevices(w http.ResponseWriter, r *http.Request) {
	if setCORSHeaders(w, r) {
		return
	}
	q := r.URL.Query()
	domain := q.Get("domain")
	if domain == "" {
		http.Error(w, "domain is required", http.StatusBadRequest)
		return
	}
	result, err := h.Repo.GetDevices(r.Context(), domain, q.Get("from"), q.Get("to"))
	if err != nil {
		http.Error(w, "Query failed", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// GET /api/timeseries?domain=&from=&to=
func (h *Handler) GetTimeSeries(w http.ResponseWriter, r *http.Request) {
	if setCORSHeaders(w, r) {
		return
	}
	q := r.URL.Query()
	domain := q.Get("domain")
	if domain == "" {
		http.Error(w, "domain is required", http.StatusBadRequest)
		return
	}
	result, err := h.Repo.GetPageviewsTimeSeries(r.Context(), domain, q.Get("from"), q.Get("to"))
	if err != nil {
		http.Error(w, "Query failed", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// Recursively walks through JSON objects/arrays and truncates long strings
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
