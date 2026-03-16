package rest

import (
	"encoding/json"
	"net/http"

	"github.com/transitIOM/projectMercury/internal/adapters/input/middleware"
	"github.com/transitIOM/projectMercury/internal/domain/models"
)

// GetMessages returns the list of system messages.
// @Summary      Get system messages
// @Description  Retrieves all administrative messages stored in the system.
// @Tags         messages
// @Produce      application/json
// @Success      200  {array}   models.Message
// @Header       200  {string}  X-Message-Version  "The version ID of the messages data"
// @Failure      500  {object}  Error
// @Router       /messages [get]
func (h *Handler) GetMessages(w http.ResponseWriter, r *http.Request) {
	messages, version, err := h.service.GetMessages(r.Context())
	if err != nil {
		middleware.GetLogger(r.Context()).Error("Failed to get messages", "error", err)
		writeError(r.Context(), w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("X-Message-Version", version)
	h.sendJSON(w, http.StatusOK, messages)
}

// PostMessage adds a new system message.
// @Summary      Post system message
// @Description  Adds a new administrative message to the system.
// @Tags         messages
// @Accept       application/json
// @Produce      application/json
// @Param        message  body      models.Message  true  "Message object"
// @Success      201      {string}  string          "Created"
// @Header       201      {string}  X-Message-Version  "The updated version ID of the messages data"
// @Failure      400      {object}  Error
// @Failure      500      {object}  Error
// @Router       /messages [post]
func (h *Handler) PostMessage(w http.ResponseWriter, r *http.Request) {
	var msg models.Message
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		middleware.GetLogger(r.Context()).Warn("Invalid request body for PostMessage", "error", err)
		writeError(r.Context(), w, http.StatusBadRequest, "invalid request body")
		return
	}
	version, err := h.service.PostMessage(r.Context(), msg)
	if err != nil {
		middleware.GetLogger(r.Context()).Error("Failed to post message", "error", err)
		writeError(r.Context(), w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("X-Message-Version", version)
	w.WriteHeader(http.StatusCreated)
}
