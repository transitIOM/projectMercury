package api

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type GetVersionIDResponse struct {
	Code    int    `json:"code" example:"200"`
	Version string `json:"version" example:"20231215-143022"`
}

type GetTimetableResponse struct {
	Code        int    `json:"code" example:"200"`
	VersionID   string `json:"versionID" example:"20231215-143022"`
	DownloadURL string `json:"downloadURL" example:"https://example.com/GTFSSchedule.zip"`
}

type PutTimetableResponse struct {
	Code      int    `json:"code" example:"202"`
	VersionID string `json:"versionID" example:"20231215-143022"`
}

type GetMessagesResponse struct {
	Code      int    `json:"code" example:"200"`
	Messages  string `json:"messages" example:"{\"timestamp\": \"2026-1-1T00:00:00.000Z\", \"message\": \"Example message\"}"`
	VersionID string `json:"versionID" example:"5e4b7d12-542f-4ecf-8d95-7fbec7f7e806"`
}

type PutMessageResponse struct {
	Code      int    `json:"code" example:"202"`
	VersionID string `json:"versionID" example:"5e4b7d12-542f-4ecf-8d95-7fbec7f7e806"`
}

type PostReportBody struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Email       string `json:"email"`
	Category    string `json:"category"`
}

type Error struct {
	Code    int    `json:"code" example:"500"`
	Message string `json:"message" example:"Internal Server Error"`
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
	UnauthorizedErrorHandler = func(w http.ResponseWriter, err error) {
		writeError(w, http.StatusUnauthorized, err.Error())
	}
	InternalErrorHandler = func(w http.ResponseWriter) {
		writeError(w, http.StatusInternalServerError, "Internal Server Error")
	}
)
