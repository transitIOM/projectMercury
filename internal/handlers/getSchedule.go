package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	log "github.com/sirupsen/logrus"
	"github.com/transitIOM/projectMercury/api"
	"github.com/transitIOM/projectMercury/internal/tools"
)

// GetScheduleDownloadURL godoc
// @Summary      Get a download URL for the latest GTFS schedule
// @Description  Generates a short-lived presigned URL to download the latest GTFS schedule zip file, along with its version ID.
// @Tags         schedule
// @Produce      json
// @Success      200  {object}  api.GetTimetableResponse
// @Failure      500  {object}  api.Error
// @Router       /schedule/ [get]
func GetScheduleDownloadURL(sm tools.ObjectStorageManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Debug("Handling GetScheduleDownloadURL request")

		downloadURL, versionID, err := sm.GetLatestURL()
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

		log.Debugf("Retrieved schedule download URL and version ID: %s", versionID)
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
}
