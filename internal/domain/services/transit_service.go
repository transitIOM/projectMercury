package services

import (
	"context"
	"sync"

	"github.com/transitIOM/projectMercury/internal/ports/output"
)

type TransitService struct {
	scheduleProvider  output.ScheduleProvider
	scheduleFetcher   output.ScheduleFetcher
	realtimeGatherer  output.RealtimeGatherer
	reportManager     output.ReportManager
	messageRepository output.MessageRepository

	mu          sync.RWMutex
	subscribers map[chan any]string // map the channel to its feed type
	tripMapper  TripMapper
}

func NewTransitService(
	ctx context.Context,
	sp output.ScheduleProvider,
	sf output.ScheduleFetcher,
	rg output.RealtimeGatherer,
	rm output.ReportManager,
	mr output.MessageRepository,
) *TransitService {
	s := &TransitService{
		scheduleProvider:  sp,
		scheduleFetcher:   sf,
		realtimeGatherer:  rg,
		reportManager:     rm,
		messageRepository: mr,
		subscribers:       make(map[chan any]string),
		tripMapper:        NewTripMapper(sp),
	}

	if rg != nil {
		go s.handleRealtimeUpdates(ctx)
	}

	if sf != nil {
		go s.handleScheduleUpdates(ctx)
	}

	return s
}
