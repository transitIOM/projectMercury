package services

import (
	"context"
	"io"
	"log/slog"
	"time"
)

func (s *TransitService) handleScheduleUpdates(ctx context.Context) {
	// Initial fetch
	slog.InfoContext(ctx, "Starting initial GTFS schedule fetch")
	if err := s.scheduleFetcher.FetchLatestSchedule(ctx); err != nil {
		slog.ErrorContext(ctx, "Initial schedule fetch failed", "error", err)
	} else {
		slog.InfoContext(ctx, "Initial schedule fetch completed successfully")
	}

	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.InfoContext(ctx, "Schedule update handler stopping")
			return
		case <-ticker.C:
			slog.InfoContext(ctx, "Starting daily schedule fetch")
			if err := s.scheduleFetcher.FetchLatestSchedule(ctx); err != nil {
				slog.ErrorContext(ctx, "Daily schedule fetch failed", "error", err)
			} else {
				slog.InfoContext(ctx, "Daily schedule fetch completed successfully")
			}
		}
	}
}

func (s *TransitService) GetGTFSChecksum(ctx context.Context) (string, error) {
	return s.scheduleProvider.GetChecksum(ctx)
}

func (s *TransitService) GetGTFS(ctx context.Context) (io.ReadCloser, error) {
	return s.scheduleProvider.GetGTFSReader(ctx)
}
