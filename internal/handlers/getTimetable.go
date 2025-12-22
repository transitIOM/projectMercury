package handlers

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
	"github.com/transitIOM/projectMercury/api"
	"github.com/transitIOM/projectMercury/internal/tools"
)

func GetGTFSScheduleDownloadURL(w http.ResponseWriter, r *http.Request) {

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
		return
	}
}
