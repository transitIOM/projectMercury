package handlers

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
	"github.com/transitIOM/projectMercury/api"
	"github.com/transitIOM/projectMercury/internal/tools"
)

// @id				getGTFSScheduleDownloadURL
// @tags			GTFS Schedule
// @summary		Get a download link for the latest data
// @description	Returns a link to download GTFSSchedule.zip and the versionID
// @produce		json
// @success		200	{object}	api.GetTimetableResponse	"Returned a download link and versionID"
// @failure		500	{object}	api.Error					"Internal server error"
// @router			/schedule [get]
func getGTFSScheduleDownloadURL(w http.ResponseWriter, r *http.Request) {

	downloadURL, versionID, err := tools.GetLatestGTFSScheduleURL()
	if err != nil {
		log.Error(err)
		api.InternalErrorHandler(w)
		return
	}

	response := api.GetTimetableResponse{
		Code:        http.StatusOK,
		DownloadURL: downloadURL.String(),
		VersionID:   versionID,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Error(err)
		api.InternalErrorHandler(w)
	}
}
