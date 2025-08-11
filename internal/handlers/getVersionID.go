package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Jaycso/transit-IOMAPI/api"
	"github.com/Jaycso/transit-IOMAPI/internal/tools"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

func getVersionIDByName(w http.ResponseWriter, r *http.Request) {
	timetableName := chi.URLParam(r, "name")

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
