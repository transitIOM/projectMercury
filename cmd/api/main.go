package main

import (
	"fmt"
	"net/http"

	_ "github.com/Jaycso/transit-IOMAPI/docs"
	"github.com/Jaycso/transit-IOMAPI/internal/handlers"
	"github.com/go-chi/chi/v5"
	log "github.com/sirupsen/logrus"
	httpSwagger "github.com/swaggo/http-swagger"
)

//	@title			transitIOMAPI
//	@version		1.0
//	@description	This is an API for the transitIOM application

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
