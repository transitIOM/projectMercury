package output

import (
	"context"

	"github.com/transitIOM/projectMercury/internal/domain/models"
)

type ReportManager interface {
	CreateIssueFromReport(ctx context.Context, report models.UserReport) error
}
