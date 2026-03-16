package output

import (
	"context"

	"github.com/transitIOM/projectMercury/internal/domain/models"
)

type RealtimeGatherer interface {
	GetVehiclePositions(ctx context.Context) ([]models.VehiclePosition, error)
	NotifyChannel() <-chan struct{}
}
