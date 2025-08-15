package handlers

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
)

func Handler(r *chi.Mux) {
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)
	r.Use(middleware.Timeout(30 * time.Second))

	v1 := chi.NewRouter()

	// dont fuck with this
	v1.Route("/timetable", func(router chi.Router) {
		router.Use(httprate.LimitByIP(5, time.Minute))
		// requires the timetable name to be provided ending in .json

		router.Get("/version/{name:^[A-Za-z]+$}", getVersionIDByName)
		router.Get("/{name:^[A-Za-z]+$}", getTimetableByName)
		router.Put("/{name:^[A-Za-z]+$}", putTimetableByName)
	})

	//TODO: Need to test endpoints, finish adding swagger comments for documentation, change logging target to a log file

	r.Mount("/api/v1", v1)
}
