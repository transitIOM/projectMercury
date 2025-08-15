package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Jaycso/transit-IOMAPI/api"
	"github.com/Jaycso/transit-IOMAPI/internal/tools"
	"github.com/go-chi/chi/v5"
	log "github.com/sirupsen/logrus"
)

func getTimetableByName(w http.ResponseWriter, r *http.Request) {
	timetableName := chi.URLParam(r, "name")

	timetable, versionID, err := tools.GetLatestTimetable("timetables", timetableName)
	if err != nil {
		log.Error(err)
		api.InternalErrorHandler(w)
		return
	}

	response := api.GetTimetableResponse{
		VersionID: versionID,
		Data:      timetable,
		Code:      http.StatusOK,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Error(err)
		api.InternalErrorHandler(w)
	}
}
