package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	log "github.com/sirupsen/logrus"
	httpSwagger "github.com/swaggo/http-swagger"
	_ "github.com/transitIOM/projectMercury/docs"
	"github.com/transitIOM/projectMercury/internal/handlers"
)

//	@title			projectMercury
//	@version		1.0
//	@description	The transitIOM REST API

//	@contact.name	Jayden T
//	@contact.email	support@jaydent.uk

// @basePath	/api/v1
func main() {
	log.SetReportCaller(true)
	r := chi.NewRouter()
	handlers.Handler(r)
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8090/swagger/doc.json"), //The url pointing to API definition
	))

	fmt.Println("Starting transit-IOMAPI service...")

	err := http.ListenAndServe(":8090", r)
	if err != nil {
		log.Error(err)
	}
}
