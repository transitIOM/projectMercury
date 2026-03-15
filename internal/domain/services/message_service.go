package services

import (
	"context"
	"log/slog"

	"github.com/transitIOM/projectMercury/internal/domain/models"
)

func (s *TransitService) GetMessages(ctx context.Context) ([]models.Message, string, error) {
	slog.DebugContext(ctx, "Retrieving administrative messages")
	messages, version, err := s.messageRepository.GetMessages(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to retrieve messages", "error", err)
		return nil, "", err
	}
	slog.DebugContext(ctx, "Messages successfully retrieved", "count", len(messages), "version", version)
	return messages, version, nil
}

func (s *TransitService) PostMessage(ctx context.Context, message models.Message) (string, error) {
	slog.InfoContext(ctx, "Posting new administrative message")
	version, err := s.messageRepository.SaveMessage(ctx, message)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to save message", "error", err)
		return "", err
	}
	slog.InfoContext(ctx, "Message successfully saved", "version", version)
	return version, nil
}
