package input

import (
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
	_ "github.com/transitIOM/projectMercury/docs"
	customMiddleware "github.com/transitIOM/projectMercury/internal/adapters/input/middleware"
	"github.com/transitIOM/projectMercury/internal/adapters/input/rest"
	"github.com/transitIOM/projectMercury/internal/adapters/input/sse"
	"github.com/transitIOM/projectMercury/internal/ports/input"
)

// @title           Project Mercury
// @version         v0.3.0
// @description     The Project Mercury REST API provides comprehensive transit data services for the transitIOM application, including real-time bus locations, GTFS schedules, and messaging systems.
// @termsOfService  coming soon

// @contact.name   Jayden Thompson
// @contact.email  admin@transitIOM.com

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      api.transitiom.com
// @BasePath  /v1

func NewRouter(service input.TransitService) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(customMiddleware.SlogLogger(slog.Default()))
	r.Use(middleware.RealIP)
	//r.Use(middleware.Timeout(30 * time.Second))
	r.Use(middleware.StripSlashes)
	r.Use(middleware.CleanPath)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Heartbeat("/health"))

	v1 := chi.NewRouter()

	v1.Get("/docs/*", httpSwagger.WrapHandler)

	restHandler := rest.NewHandler(service)
	restHandler.RegisterRoutes(v1)

	sseHandler := sse.NewHandler(service)
	v1.Handle("/stream", sseHandler)
	v1.Handle("/stream/all", sseHandler)
	v1.Handle("/stream/vehicle-positions", sseHandler)
	//v1.Handle("/stream/trip-updates", sseHandler)
	//v1.Handle("/stream/service-alerts", sseHandler)

	r.Mount("/v1", v1)

	return r
}
