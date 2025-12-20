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
// @failure		400	{object}	api.Error					"Invalid timetable name"
// @failure		500	{object}	api.Error					"Internal server error"
// getGTFSScheduleDownloadURL handles GET /schedule and returns a JSON payload containing a download link to the latest GTFS schedule and its version ID.
// It queries the latest schedule URL and version; on success it responds with HTTP 200 and a JSON api.GetTimetableResponse containing Code, DownloadURL, and VersionID and sets Content-Type to application/json.
// If retrieving the URL or encoding the response fails, it logs the error and responds with an internal server error.
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