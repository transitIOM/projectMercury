package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/transitIOM/projectMercury/internal/handlers"
	"github.com/transitIOM/projectMercury/internal/tools"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Warn("Could not load .env file, using environment variables")
	}

	// Set log format
	logFormat := os.Getenv("LOG_FORMAT")
	if logFormat == "text" {
		log.SetFormatter(&log.TextFormatter{
			FullTimestamp: true,
		})
	} else {
		log.SetFormatter(&log.JSONFormatter{})
	}

	// Set log level
	logLevelStr := os.Getenv("LOG_LEVEL")
	if logLevelStr == "" {
		logLevelStr = "info"
	}
	logLevel, err := log.ParseLevel(logLevelStr)
	if err != nil {
		log.Warnf("Invalid LOG_LEVEL '%s', defaulting to 'info'", logLevelStr)
		logLevel = log.InfoLevel
	}
	log.SetLevel(logLevel)
}

func main() {
	log.SetReportCaller(true)

	browserCtx, browserCancel := context.WithCancel(context.Background())
	tools.InitializeMinio()
	tools.InitialiseLinearGraphqlConnection()
	go tools.InitializeBrowser(browserCtx)

	r := chi.NewRouter()
	handlers.Handler(r)

	srv := &http.Server{
		Addr:    ":8090",
		Handler: r,
	}

	go func() {
		log.Info("Starting transit-IOMAPI service...")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be caught, so no need to add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}
	browserCancel()
	time.Sleep(100 * time.Millisecond)
	log.Info("Server exiting")
}
