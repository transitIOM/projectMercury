package handlers

import (
	"errors"
	"io"
	"net/http"

	"github.com/Jaycso/transit-IOMAPI/api"
	"github.com/Jaycso/transit-IOMAPI/internal/tools"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

func getTimetableByName(w http.ResponseWriter, r *http.Request) {
	timetableName := chi.URLParam(r, "name")

	if timetableName == "" {
		log.Error("Timetable name is required")
		err := errors.New("timetable name is required")
		api.RequestErrorHandler(w, err)
	}

	timetable, err := tools.GetLatestTimetable("timetables", timetableName)
	if err != nil {
		log.Error(err)
		api.InternalErrorHandler(w)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = io.Copy(w, timetable)
	if err != nil {
		log.Error(err)
		api.InternalErrorHandler(w)
	}
}
