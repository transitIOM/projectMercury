package main

import (
	"fmt"
	"net/http"

	"github.com/Jaycso/transit-IOMAPI/internal/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/docgen"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetReportCaller(true)
	r := chi.NewRouter()
	handlers.Handler(r)

	docgen.PrintRoutes(r)
	fmt.Println("Starting transit-IOMAPI service...")

	err := http.ListenAndServe(":8090", r)
	if err != nil {
		log.Error(err)
	}
}
