package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	log "github.com/sirupsen/logrus"
	_ "github.com/transitIOM/projectMercury/docs"
)

func main() {
	log.SetReportCaller(true)
	r := chi.NewRouter()

	fmt.Println("Starting transit-IOMAPI service...")

	err := http.ListenAndServe(":8090", r)
	if err != nil {
		log.Error(err)
	}
}
