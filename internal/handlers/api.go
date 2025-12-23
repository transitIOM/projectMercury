package handlers

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	httpSwagger "github.com/swaggo/http-swagger"
	_ "github.com/transitIOM/projectMercury/docs"
	internalMiddleware "github.com/transitIOM/projectMercury/internal/middleware"
)

// @title           Project Mercury API
// @version         0.1.0
// @description     API for Transit IOM GTFS Schedule management.
// @termsOfService  http://swagger.io/terms/

// @contact.name   Jayden Thompson
// @contact.email  admin@transitIOM.com

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8090
// @BasePath  /api/v1

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func Handler(r *chi.Mux) {
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)
	r.Use(middleware.Timeout(30 * time.Second))

	r.Get("/swagger/*", httpSwagger.WrapHandler)

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
			r.Get("/test", func(w http.ResponseWriter, req *http.Request) {
				w.Write([]byte("protected area"))
			})
		})
	})

	r.Mount("/api/v1", v1)
}
