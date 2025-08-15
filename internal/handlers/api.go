package handlers

import (
	"net/http"
	"time"

	"github.com/Jaycso/transit-IOMAPI/api"
	intmiddleware "github.com/Jaycso/transit-IOMAPI/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	log "github.com/sirupsen/logrus"
)

func Handler(r *chi.Mux) {
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)
	r.Use(middleware.Timeout(30 * time.Second))

	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("pong"))
		if err != nil {
			log.Error(err)
			api.InternalErrorHandler(w)
			return
		}
	})

	apiRouter := chi.NewRouter()

	// dont fuck with this
	apiRouter.Route("timetable", func(router chi.Router) {
		r.Use(httprate.LimitByIP(5, time.Minute))
		// requires the timetable name to be provided ending in .json
		r.Use(intmiddleware.ValidateJsonFilename)

		r.Get("/version/?={name}", getVersionIDByName)
		r.Get("/?={name}", getTimetableByName)
		r.Put("/?={name}", putTimetableByName)
	})

	r.Mount("/api", apiRouter)
}
