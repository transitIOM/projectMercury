package handlers

import (
	"net/http"
	"time"

	"github.com/Jaycso/transit-IOMAPI/api"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/httprate"
	log "github.com/sirupsen/logrus"
)

func Handler(r *chi.Mux) {
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
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

	r.Route("latestTimetableVersion", func(router chi.Router) {
		r.Use(httprate.LimitByIP(5, time.Minute))

		r.Get("/latestTimetableVersion", getLatestVersionID)
	})
}
