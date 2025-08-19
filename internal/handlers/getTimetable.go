package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Jaycso/transit-IOMAPI/api"
	"github.com/Jaycso/transit-IOMAPI/internal/tools"
	"github.com/go-chi/chi/v5"
	log "github.com/sirupsen/logrus"
)

// @id				getTimetableByName
// @tags			timetable
// @summary		Takes a timetable name and returns the latest JSON timetable data with its corresponding version ID
// @description	Returns the JSON timetable and its version ID
// @produce		json
// @param			name	path		string						true	"The name of the timetable you want (including .json)"	Format(json)	Pattern(name:^[A-Za-z]+$)
// @success		200		{object}	api.GetTimetableResponse	"Returns the latest timetable with version ID"
// @failure		400		{object}	api.Error					"Invalid timetable name"
// @failure		500		{object}	api.Error					"Internal server error"
// @router			/timetable/{name} [get]
func getTimetableByName(w http.ResponseWriter, r *http.Request) {
	timetableName := chi.URLParam(r, "name")
	timetableName = fmt.Sprintf("timetable%s", ".json")

	timetable, versionID, err := tools.GetLatestTimetable("timetables", timetableName)
	if err != nil {
		log.Error(err)
		api.InternalErrorHandler(w)
		return
	}

	response := api.GetTimetableResponse{
		VersionID: versionID,
		File:      timetable,
		Code:      http.StatusOK,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Error(err)
		api.InternalErrorHandler(w)
	}
}
