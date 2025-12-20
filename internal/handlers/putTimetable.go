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

// @id				putGTFSSchedule
// @tags			GTFS Schedule
// @summary		Takes a GTFS Schedule .zip package and uploads it to the object store
// @description	Updates object store with latest GTFS Schedule. Only `.zip` files are allowed.
// @accept			multipart/form-data
// @produce		json
// @param			GTFSSchedule	formData	file						true	"A GTFS schedule package (must be .zip)"
// @success		200				{object}	api.PutTimetableResponse	"File successfully uploaded"
// @failure		400				{object}	api.Error					"Invalid file type"
// @failure		500				{object}	api.Error					"Internal server error"
// putGTFSSchedule handles PUT /schedule requests by accepting a GTFS schedule .zip upload and storing it as the latest schedule.
// It validates the uploaded file's Content-Type (must be "application/zip" or "application/x-zip-compressed") and non-zero size, delegates storage to tools.PutLatestGTFSSchedule, and responds with a JSON payload containing the accepted status code and the stored version ID.
func putGTFSSchedule(w http.ResponseWriter, r *http.Request) {
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
		err = errors.New(fmt.Sprintf("unsupported file type: %s. Please upload a GTFS schedule .zip package.", filetype))
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
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Error(err)
		api.InternalErrorHandler(w)
	}
}