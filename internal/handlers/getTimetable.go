package handlers

import (
	"errors"
	"github.com/Jaycso/transit-IOMAPI/api"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func getTimetableByName(w http.ResponseWriter, r *http.Request) {
	var err error
	timetableName := chi.URLParam(r, "name")
	if timetableName == "" {
		log.Error("Timetable name is required")
		err = errors.New("timetable name is required")
		api.RequestErrorHandler(w, err)
	}

}
