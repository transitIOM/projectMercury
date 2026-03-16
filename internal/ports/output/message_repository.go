package output

import (
	"context"

	"github.com/transitIOM/projectMercury/internal/domain/models"
)

type MessageRepository interface {
	GetMessages(ctx context.Context) ([]models.Message, string, error)
	SaveMessage(ctx context.Context, message models.Message) (string, error)
}
