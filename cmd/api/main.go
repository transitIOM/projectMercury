package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/transitIOM/projectMercury/internal/handlers"
)

func init() {
	//log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)

	if err := godotenv.Load(); err != nil {
		log.Fatal("No .env file found")
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
