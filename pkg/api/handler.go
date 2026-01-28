package api

import (
	"encoding/json"
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


func (h *Handler) TrackEvent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var event core.Event
	
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	event.ID = uuid.NewString()
	event.Timestamp = time.Now().UTC()

	err := h.Repo.Insert(r.Context(), &event)
	if err != nil {
		http.Error(w, "Failed to save event", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}