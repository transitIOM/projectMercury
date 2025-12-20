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
// @summary		Returns the latest VersionID
// @description	Get the versionID of the latest GTFS Schedule
// @produce		json
// @success		200	{object}	api.GetVersionIDResponse	"Returned the version ID"
// @failure		500	{object}	api.Error					"Internal server error"
// @router			/schedule/version [get]
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
