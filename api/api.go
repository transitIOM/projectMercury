package api

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type GetVersionIDResponse struct {
	Code    int
	Version string
}

type GetTimetableResponse struct {
	Code        int
	VersionID   string
	DownloadURL string
}

type PutTimetableResponse struct {
	Code      int
	VersionID string
}

type Error struct {
	Code    int
	Message string
}

func writeError(w http.ResponseWriter, code int, message string) {
	resp := Error{
		Code:    code,
		Message: message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Errorf("Error writing response: %v", err)
	}
}

var (
	RequestErrorHandler = func(w http.ResponseWriter, err error) {
		writeError(w, http.StatusBadRequest, err.Error())
	}
	InternalErrorHandler = func(w http.ResponseWriter) {
		writeError(w, http.StatusInternalServerError, "Internal Server Error")
	}
)
