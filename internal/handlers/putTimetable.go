package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Jaycso/transit-IOMAPI/api"
	"github.com/Jaycso/transit-IOMAPI/internal/tools"
	"github.com/go-chi/chi/v5"
	log "github.com/sirupsen/logrus"
)

func putTimetableByName(w http.ResponseWriter, r *http.Request) {
	timetableName := chi.URLParam(r, "name")
	jsonBody := r.Body

	versionID, err := tools.PutLatestTimetable("timetables", timetableName, jsonBody)
	if err != nil {
		log.Error(err)
		api.InternalErrorHandler(w)
		return
	}

	response := api.PutTimetableResponse{
		VersionID: versionID,
		Code:      http.StatusAccepted,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Error(err)
		api.InternalErrorHandler(w)
	}
}
