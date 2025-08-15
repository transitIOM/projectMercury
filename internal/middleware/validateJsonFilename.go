package middleware

import (
	"errors"
	"net/http"
	"regexp"

	"github.com/Jaycso/transit-IOMAPI/api"
	"github.com/go-chi/chi/v5"
	log "github.com/sirupsen/logrus"
)

func ValidateJsonFilename(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		timetableName := chi.URLParam(r, "name")

		if timetableName == "" {
			log.Error("Timetable name is required")
			err := errors.New("timetable name is required")
			api.RequestErrorHandler(w, err)
		}

		regexCheck, err := regexp.MatchString("^[A-Za-z]+\\.json$", timetableName)

		if err != nil {
			log.Error(err)
			return
		}

		if !regexCheck {
			log.Error("Timetable name is invalid")
			err := errors.New("timetable name is invalid")
			api.RequestErrorHandler(w, err)
		}

		next.ServeHTTP(w, r)
	})
}
