package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/transitIOM/projectMercury/internal/adapters/input"
	"github.com/transitIOM/projectMercury/internal/adapters/output/cura"
	"github.com/transitIOM/projectMercury/internal/adapters/output/filesystem"
	"github.com/transitIOM/projectMercury/internal/adapters/output/linear"
	"github.com/transitIOM/projectMercury/internal/adapters/output/signalr"
	"github.com/transitIOM/projectMercury/internal/domain/services"
	"github.com/transitIOM/projectMercury/internal/infrastructure/config"
)

func main() {
	// Configure slog for JSON output
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Initialize Output Adapters
	gtfsAdapter := filesystem.NewGTFSAdapter(cfg.GTFSFilePath, cfg.MessagesFilePath)

	linearAdapter, err := linear.NewAdapter()
	if err != nil {
		slog.Warn("Could not initialize Linear adapter", "error", err)
	}

	signalrAdapter := signalr.NewAdapter(ctx, cfg.SignalRExpiry)
	go signalrAdapter.Start(ctx)

	curaAdapter := cura.NewAdapter(cfg.CuraOwner, cfg.CuraRepo, cfg.GTFSFilePath)

	// Initialize Domain Service
	transitService := services.NewTransitService(ctx, gtfsAdapter, curaAdapter, signalrAdapter, linearAdapter, gtfsAdapter)

	// Initialize Input Adapter (Router)
	r := input.NewRouter(transitService)

	srv := &http.Server{
		Addr:    cfg.AppPort,
		Handler: r,
	}

	go func() {
		slog.Info("Starting transit-IOMAPI service...", "port", cfg.AppPort)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("listen error", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}
	time.Sleep(100 * time.Millisecond)
	slog.Info("Server exiting")
}
