package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	log "github.com/sirupsen/logrus"
	"github.com/transitIOM/projectMercury/internal/handlers"
)

func main() {
	log.SetReportCaller(true)
	r := chi.NewRouter()
	handlers.Handler(r)

	fmt.Println("Starting transit-IOMAPI service...")

	err := http.ListenAndServe(":8090", r)
	if err != nil {
		log.Error(err)
	}
}
