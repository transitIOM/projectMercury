package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Jaycso/transit-IOMAPI/api"
	"github.com/Jaycso/transit-IOMAPI/internal/minio"
	"github.com/gorilla/schema"
	log "github.com/sirupsen/logrus"
)

func getVersionID(w http.ResponseWriter, r *http.Request) {
	params := api.GetVersionIDParams{}
	decoder := schema.NewDecoder()

	err := decoder.Decode(&params, r.URL.Query())

	if err != nil {
		log.Error(err)
		api.InternalErrorHandler(w)
		return
	}

	versionID, err := minio.GetLatestVersionID("timetables", params.TimetableName)
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
