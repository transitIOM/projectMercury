package rest

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/transitIOM/projectMercury/internal/ports/input"
)

type Handler struct {
	service input.TransitService
}

func NewHandler(service input.TransitService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/schedule.zip", h.GetGTFSSchedule)
	r.Post("/report", h.PostReport)
	r.Get("/messages", h.GetMessages)
	r.Post("/messages", h.PostMessage)
}

func (h *Handler) sendJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}
