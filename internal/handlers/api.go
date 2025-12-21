package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/go-chi/jwtauth/v5"
)

var tokenAuth *jwtauth.JWTAuth

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
			r.Get("/authenticate", GetAdminToken)
		})

		// private routes
		r.Group(func(r chi.Router) {
			// verify token
			r.Use(jwtauth.Verifier(tokenAuth))
			r.Put("/", PutGTFSSchedule)
			r.Get("/admin", func(w http.ResponseWriter, req *http.Request) {
				_, claims, _ := jwtauth.FromContext(req.Context())
				w.Write([]byte(fmt.Sprintf("protected area. hi %v", claims["user_id"])))
			})
		})
	})

	r.Mount("/api/v1", v1)
}
