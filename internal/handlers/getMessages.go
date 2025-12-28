package handlers

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
	"github.com/transitIOM/projectMercury/api"
	"github.com/transitIOM/projectMercury/internal/tools"
)

// GetMessages godoc
// @Summary      Get latest messages
// @Description  Retrieves the last 3 lines of the message log and the version ID.
// @Tags         messages
// @Produce      json
// @Success      200  {object}  api.GetMessagesResponse
// @Failure      500  {object}  api.Error
// @Router       /messages/ [get]
func GetMessages(w http.ResponseWriter, r *http.Request) {
	messageCount := 3
	b := tools.GetLastNLines(messageCount)
	v, err := tools.GetLatestMessageLogVersionID()
	if err != nil {
		log.Error(err)
		api.InternalErrorHandler(w)
		return
	}

	response := api.GetMessagesResponse{
		Code:      http.StatusOK,
		Messages:  b.String(),
		VersionID: v,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Error(err)
		api.InternalErrorHandler(w)
		return
	}
}
