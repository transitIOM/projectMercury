package handlers

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	internaMiddleware "github.com/transitIOM/projectMercury/internal/middleware"
)

func Handler(r *chi.Mux) {
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)
	r.Use(middleware.Timeout(30 * time.Second))

	v1 := chi.NewRouter()

	// GTFS schedule public endpoint v1
	v1.Route("/schedule", func(r chi.Router) {
		r.Use(httprate.LimitByIP(5, time.Minute))

		// public routes
		r.Group(func(r chi.Router) {
			r.Get("/version", GetGTFSScheduleVersionID)
			r.Get("/", GetGTFSScheduleDownloadURL)
		})

		// private routes
		r.Group(func(r chi.Router) {
			r.Use(internaMiddleware.APIKeyAuth)
			r.Put("/", PutGTFSSchedule)
		})
	})

	r.Mount("/api/v1", v1)
}
