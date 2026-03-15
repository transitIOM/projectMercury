package services

import (
	"context"
	"log/slog"

	"github.com/transitIOM/projectMercury/internal/domain/models"
)

func (s *TransitService) PostReport(ctx context.Context, report models.UserReport) error {
	slog.InfoContext(ctx, "Processing user report", "category", report.Category, "email", report.Email)
	err := s.reportManager.CreateIssueFromReport(ctx, report)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to create issue from report", "error", err)
		return err
	}
	slog.InfoContext(ctx, "User report successfully processed")
	return nil
}
