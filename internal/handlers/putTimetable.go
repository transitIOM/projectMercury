package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	log "github.com/sirupsen/logrus"
	"github.com/transitIOM/projectMercury/api"
	"github.com/transitIOM/projectMercury/internal/tools"
)

func putTimetableByName(w http.ResponseWriter, r *http.Request) {
	timetableName := chi.URLParam(r, "name")
	timetableName = fmt.Sprintf("timetable%s", ".json")
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
