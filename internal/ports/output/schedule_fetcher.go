package output

import (
	"context"
)

type ScheduleFetcher interface {
	FetchLatestSchedule(ctx context.Context) error
}
