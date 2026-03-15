package rest

import (
	"encoding/json"
	"net/http"

	"github.com/transitIOM/projectMercury/internal/adapters/input/middleware"
	"github.com/transitIOM/projectMercury/internal/domain/models"
)

// PostReport submits a user report or feedback.
// @Summary      Submit user report
// @Description  Submits a user report, feedback, or bug report, which is then forwarded to Linear.
// @Tags         reports
// @Accept       application/json
// @Produce      application/json
// @Param        report  body      models.UserReport  true  "Report content"
// @Success      201     {string}  string             "Created"
// @Failure      400     {object}  Error
// @Failure      500     {object}  Error
// @Router       /report [post]
func (h *Handler) PostReport(w http.ResponseWriter, r *http.Request) {
	var report models.UserReport
	if err := json.NewDecoder(r.Body).Decode(&report); err != nil {
		middleware.GetLogger(r.Context()).Warn("Invalid request body for PostReport", "error", err)
		writeError(r.Context(), w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.service.PostReport(r.Context(), report); err != nil {
		middleware.GetLogger(r.Context()).Error("Failed to post report to Linear", "error", err)
		writeError(r.Context(), w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusCreated)
}
