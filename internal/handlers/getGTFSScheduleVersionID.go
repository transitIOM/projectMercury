package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	log "github.com/sirupsen/logrus"
	"github.com/transitIOM/projectMercury/api"
	"github.com/transitIOM/projectMercury/internal/tools"
)

// GetGTFSScheduleVersionID godoc
// @Summary      Get latest GTFS schedule version ID
// @Description  Retrieves the version ID of the latest GTFS schedule.
// @Tags         schedule
// @Produce      json
// @Success      200  {object}  api.GetVersionIDResponse
// @Failure      500  {object}  api.Error
// @Router       /schedule/version [get]
func GetGTFSScheduleVersionID(w http.ResponseWriter, r *http.Request) {
	log.Debug("Handling GetGTFSScheduleVersionID request")

	versionID, err := tools.GetLatestGTFSScheduleVersionID()
	if err != nil {
		if errors.Is(err, tools.NoGTFSScheduleFound) {
			response := api.GetVersionIDResponse{
				Code:    http.StatusNoContent,
				Version: "",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(response.Code)
			err = json.NewEncoder(w).Encode(response)
			return
		}
		log.Error(err)
		api.InternalErrorHandler(w)
		return
	}

	log.Debugf("Retrieved GTFS schedule version ID: %s", versionID)
	response := api.GetVersionIDResponse{
		Code:    http.StatusOK,
		Version: versionID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(response.Code)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Error(err)
		api.InternalErrorHandler(w)
		return
	}
}
