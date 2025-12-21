package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"

	log "github.com/sirupsen/logrus"
	"github.com/transitIOM/projectMercury/api"
	"github.com/transitIOM/projectMercury/internal/tools"
)

func PutGTFSSchedule(w http.ResponseWriter, r *http.Request) {
	file, fileHeader, err := r.FormFile("GTFSSchedule")
	if err != nil {
		api.InternalErrorHandler(w)
		return
	}
	defer func(file multipart.File) {
		err = file.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(file)
	filetype := fileHeader.Header.Get("Content-Type")
	if (filetype != "application/zip") && (filetype != "application/x-zip-compressed") {
		err = fmt.Errorf("unsupported file type: %s. Please upload a GTFS schedule .zip package", filetype)
		api.RequestErrorHandler(w, err)
		return
	}
	if fileHeader.Size == 0 {
		err = errors.New("file size is zero")
		api.RequestErrorHandler(w, err)
		return
	}

	versionID, err := tools.PutLatestGTFSSchedule(file, fileHeader.Size)
	if err != nil {
		log.Error(err)
		api.InternalErrorHandler(w)
		return
	}

	response := api.PutTimetableResponse{
		Code:      http.StatusAccepted,
		VersionID: versionID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(response.Code)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Error(err)
		api.InternalErrorHandler(w)
	}
}
