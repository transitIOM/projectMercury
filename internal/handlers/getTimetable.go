package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	log "github.com/sirupsen/logrus"
	"github.com/transitIOM/projectMercury/api"
	"github.com/transitIOM/projectMercury/internal/tools"
)

// GetGTFSScheduleDownloadURL godoc
// @Summary      Get latest GTFS schedule download URL
// @Description  Retrieves the download URL and version ID of the latest GTFS schedule.
// @Tags         schedule
// @Produce      json
// @Success      200  {object}  api.GetTimetableResponse
// @Failure      500  {object}  api.Error
// @Router       /schedule/ [get]
func GetGTFSScheduleDownloadURL(w http.ResponseWriter, r *http.Request) {
	log.Debug("Handling GetGTFSScheduleDownloadURL request")

	downloadURL, versionID, err := tools.GetLatestGTFSScheduleURL()
	if err != nil {
		if errors.Is(err, tools.NoGTFSScheduleFound) {
			response := api.GetTimetableResponse{
				Code:        http.StatusNoContent,
				DownloadURL: "",
				VersionID:   "",
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

	log.Debugf("Retrieved GTFS schedule download URL and version ID: %s", versionID)
	response := api.GetTimetableResponse{
		Code:        http.StatusOK,
		DownloadURL: downloadURL.String(),
		VersionID:   versionID,
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
