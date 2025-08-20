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

// @id				getVersionIDByName
// @tags			timetable
// @summary		Takes a timetable name and returns the version ID
// @description	Returns the JSON timetable and its version ID
// @produce		json
// @param			name	path		string						true	"The name of the timetable you want (including .json)"	Format(json)	Pattern(name:^[A-Za-z]+$)
// @success		200		{object}	api.GetVersionIDResponse	"Returns the latest version ID"
// @failure		400		{object}	api.Error					"Invalid timetable name"
// @failure		500		{object}	api.Error					"Internal server error"
// @router			/timetable/version/{name} [get]
func getVersionIDByName(w http.ResponseWriter, r *http.Request) {
	timetableName := chi.URLParam(r, "name")
	timetableName = fmt.Sprintf("timetable%s", ".json")

	versionID, err := tools.GetLatestVersionID("timetables", timetableName)
	if err != nil {
		log.Error(err)
		api.InternalErrorHandler(w)
		return
	}

	response := api.GetVersionIDResponse{
		Version: versionID,
		Code:    http.StatusOK,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Error(err)
		api.InternalErrorHandler(w)
		return
	}
}
