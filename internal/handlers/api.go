package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	internalMiddleware "github.com/transitIOM/projectMercury/internal/middleware"
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
			r.Use(internalMiddleware.APIKeyAuth)
			r.Put("/", PutGTFSSchedule)
			r.Get("/admin", func(w http.ResponseWriter, req *http.Request) {
				w.Write([]byte(fmt.Sprintf("protected area.")))
			})
		})
	})

	r.Mount("/api/v1", v1)
}
