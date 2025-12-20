package handlers

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
	"github.com/transitIOM/projectMercury/api"
	"github.com/transitIOM/projectMercury/internal/tools"
)

// @id				getGTFSScheduleVersionID
// @tags			GTFS Schedule
// @summary		Returns the latest  VersionID
// @description	Get the versionID of the latest GTFS Schedule
// @produce		json
// @success		200	{object}	api.GetVersionIDResponse	"Returned the version ID"
// @failure		400	{object}	api.Error					"Invalid timetable name"
// @failure		500	{object}	api.Error					"Internal server error"
// getGTFSScheduleVersionID handles HTTP GET requests and returns the latest GTFS schedule VersionID in JSON.
// On success it writes a 200 response containing api.GetVersionIDResponse with the Version field populated.
// On failure it logs the error and delegates the response to api.InternalErrorHandler.
func getGTFSScheduleVersionID(w http.ResponseWriter, r *http.Request) {

	versionID, err := tools.GetLatestGTFSScheduleVersionID()
	if err != nil {
		log.Error(err)
		api.InternalErrorHandler(w)
		return
	}

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