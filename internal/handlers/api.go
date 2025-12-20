package handlers

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
)

// Handler configures the provided root router with global middleware, mounts the API under /api/v1, and registers the /schedule endpoints.
// The global middleware applied are request ID, logging, panic recovery, trailing-slash stripping, and a 30-second request timeout.
// The /api/v1/schedule subtree enforces a per-IP rate limit of 5 requests per minute and exposes:
// GET /version — returns the schedule version ID;
// GET / — returns the GTFS schedule download URL;
// PUT / — accepts a GTFS schedule submission.
func Handler(r *chi.Mux) {
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)
	r.Use(middleware.Timeout(30 * time.Second))

	v1 := chi.NewRouter()

	// dont fuck with this
	v1.Route("/schedule", func(router chi.Router) {
		router.Use(httprate.LimitByIP(5, time.Minute))
		// requires the timetable name to be provided ending in .json

		router.Get("/version", getGTFSScheduleVersionID)
		router.Get("/", getGTFSScheduleDownloadURL)
		router.Put("/", putGTFSSchedule)
	})

	//TODO: Need to test endpoints, finish adding swagger comments for documentation, change logging target to a log file

	r.Mount("/api/v1", v1)
}