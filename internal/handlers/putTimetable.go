package handlers

import (
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"

	"github.com/Jaycso/transit-IOMAPI/api"
	"github.com/Jaycso/transit-IOMAPI/internal/tools"
	log "github.com/sirupsen/logrus"
)

// @id				putTimetableByName
// @tags			timetable
// @summary		Takes a GTFS .zip file and uploads it to the object store
// @description	Updates object store with latest GTFS data
// @produce		json
// @param			file	body		file						true	"A GTFS .zip package"
// @success		200		{object}	api.PutTimetableResponse	"Returns the latest timetable with version ID"
// @failure		400		{object}	api.Error					"Invalid timetable name"
// @failure		500		{object}	api.Error					"Internal server error"
// @router			/timetable/{name} [put]
func putTimetableByName(w http.ResponseWriter, r *http.Request) {
	file, fileHeader, err := r.FormFile("file")
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
	if fileHeader.Filename == "" {
		err = errors.New("filename is empty")
		api.RequestErrorHandler(w, err)
		return
	}
	filetype := fileHeader.Header.Get("Content-Type")
	if filetype != "application/zip" {
		err = errors.New("unsupported file type: " + filetype)
		api.RequestErrorHandler(w, err)
		return
	}

	versionID, err := tools.PutLatestTimetable("timetables", fileHeader.Filename, file, fileHeader.Size)
	if err != nil {
		log.Error(err)
		api.InternalErrorHandler(w)
		return
	}

	response := api.PutTimetableResponse{
		VersionID: versionID,
		Code:      http.StatusAccepted,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Error(err)
		api.InternalErrorHandler(w)
	}
}
