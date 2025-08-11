package main

import (
	"fmt"
	"net/http"

	"github.com/Jaycso/transit-IOMAPI/internal/handlers"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"

	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

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
