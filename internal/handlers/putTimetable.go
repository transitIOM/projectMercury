package handlers

import (
	"encoding/json"
	"errors"
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
// @param			GTFSSchedule	formData	file						true	"A GTFS schedule .zip package (must be .zip)"
// @success		200				{object}	api.PutTimetableResponse	"File successfully uploaded"
// @failure		400				{object}	api.Error					"Invalid file type"
// @failure		500				{object}	api.Error					"Internal server error"
// @router			/schedule [put]
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
	if fileHeader.Size == 0 {
		err = errors.New("file size is zero")
		api.RequestErrorHandler(w, err)
		return
	}
	filetype := fileHeader.Header.Get("Content-Type")
	if (filetype != "application/zip") && (filetype != "application/x-zip-compressed") {
		err = errors.New("unsupported file type: " + filetype)
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
