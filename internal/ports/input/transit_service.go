package input

import (
	"context"
	"io"

	"github.com/transitIOM/projectMercury/internal/domain/models"
)

type TransitService interface {
	GetGTFS(ctx context.Context) (io.ReadCloser, error)
	GetGTFSChecksum(ctx context.Context) (string, error)
	PostReport(ctx context.Context, report models.UserReport) error
	GetMessages(ctx context.Context) ([]models.Message, string, error)
	PostMessage(ctx context.Context, message models.Message) (string, error)
	Subscribe(ctx context.Context, feedType string) (<-chan any, error)
}
