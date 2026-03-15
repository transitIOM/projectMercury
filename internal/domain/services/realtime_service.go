package services

import (
	"context"
	"log/slog"
)

func (s *TransitService) handleRealtimeUpdates(ctx context.Context) {
	notifyCh := s.realtimeGatherer.NotifyChannel()
	slog.InfoContext(ctx, "Realtime update handler started")
	for {
		select {
		case <-ctx.Done():
			slog.InfoContext(ctx, "Realtime update handler stopping")
			return
		case _, ok := <-notifyCh:
			if !ok {
				slog.WarnContext(ctx, "Realtime gatherer notification channel closed")
				return
			}
			positions, err := s.realtimeGatherer.GetVehiclePositions(ctx)
			if err != nil {
				slog.ErrorContext(ctx, "Failed to get vehicle positions from gatherer", "error", err)
				continue
			}

			slog.DebugContext(ctx, "Processing realtime update", "count", len(positions))

			// Pre-generate the GTFS-RT feed snapshot
			feed := BuildGTFSRTFeed(positions, s.tripMapper)

			s.mu.RLock()
			subscriberCount := len(s.subscribers)
			for sub, feedType := range s.subscribers {
				var data any
				switch feedType {
				case "all", "vehicle-positions":
					data = feed
				case "trip-updates", "service-alerts":
					// Empty for now as per requirements
					continue
				default:
					data = positions // Fallback for raw JSON if ever used
				}

				select {
				case sub <- data:
				case <-ctx.Done():
					s.mu.RUnlock()
					return
				default:
					slog.WarnContext(ctx, "Subscriber channel full, skipping update", "feedType", feedType)
				}
			}
			s.mu.RUnlock()
			if subscriberCount > 0 {
				slog.DebugContext(ctx, "Broadcasted realtime update", "subscribers", subscriberCount)
			}
		}
	}
}

func (s *TransitService) Subscribe(ctx context.Context, feedType string) (<-chan any, error) {
	ch := make(chan any, 10)
	s.mu.Lock()
	s.subscribers[ch] = feedType
	slog.InfoContext(ctx, "New subscriber registered", "feedType", feedType, "totalSubscribers", len(s.subscribers))
	s.mu.Unlock()

	go func() {
		<-ctx.Done()
		s.mu.Lock()
		delete(s.subscribers, ch)
		slog.InfoContext(ctx, "Subscriber disconnected", "feedType", feedType, "remainingSubscribers", len(s.subscribers))
		s.mu.Unlock()
		close(ch)
	}()

	return ch, nil
}
