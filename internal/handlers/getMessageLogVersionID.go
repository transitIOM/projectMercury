package handlers

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
	"github.com/transitIOM/projectMercury/api"
	"github.com/transitIOM/projectMercury/internal/tools"
)

// GetMessageLogVersionID godoc
// @Summary      Get latest message log version ID
// @Description  Retrieves the version ID of the latest message log.
// @Tags         messages
// @Produce      json
// @Success      200  {object}  api.GetVersionIDResponse
// @Failure      500  {object}  api.Error
// @Router       /messages/version [get]
func GetMessageLogVersionID(w http.ResponseWriter, r *http.Request) {
	log.Debug("Handling GetMessageLogVersionID request")

	versionID, err := tools.GetLatestMessageLogVersionID()
	if err != nil {
		log.Error(err)
		api.InternalErrorHandler(w)
		return
	}

	log.Debugf("Retrieved message log version ID: %s", versionID)
	response := api.GetVersionIDResponse{
		Code:    http.StatusOK,
		Version: versionID,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Error(err)
		api.InternalErrorHandler(w)
		return
	}
}
