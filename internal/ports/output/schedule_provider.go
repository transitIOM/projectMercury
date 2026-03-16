package output

import (
	"context"
	"io"
)

type ScheduleProvider interface {
	GetGTFSReader(ctx context.Context) (io.ReadCloser, error)
	GetChecksum(ctx context.Context) (string, error)
}
