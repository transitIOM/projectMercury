package services

import (
	"context"
	"fmt"
	"time"

	"github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
	"github.com/transitIOM/projectMercury/internal/domain/models"
	"github.com/transitIOM/projectMercury/internal/ports/output"
	"google.golang.org/protobuf/proto"
)

// TripMapper provides logic to map real-time vehicle positions to static trips.
type TripMapper interface {
	MapVehicleToTrip(ctx context.Context, pos models.VehiclePosition) (string, error)
}

// SimpleTripMapper implements TripMapper with a placeholder for schedule-based mapping.
type SimpleTripMapper struct {
	scheduleProvider output.ScheduleProvider
}

func NewTripMapper(sp output.ScheduleProvider) *SimpleTripMapper {
	return &SimpleTripMapper{scheduleProvider: sp}
}

func (m *SimpleTripMapper) MapVehicleToTrip(ctx context.Context, pos models.VehiclePosition) (string, error) {
	// TODO: Implement actual lookup in GTFS schedule (trips.txt, stop_times.txt)
	// For now, we return a synthetic trip ID if none exists, or empty
	return "", nil
}

// BuildGTFSRTFeed converts internal models to a GTFS-RT FeedMessage.
func BuildGTFSRTFeed(positions []models.VehiclePosition, mapper TripMapper) *gtfs.FeedMessage {
	version := "2.0"
	timestamp := uint64(time.Now().Unix())
	incrementality := gtfs.FeedHeader_FULL_DATASET

	feed := &gtfs.FeedMessage{
		Header: &gtfs.FeedHeader{
			GtfsRealtimeVersion: &version,
			Incrementality:      &incrementality,
			Timestamp:           &timestamp,
		},
		Entity: make([]*gtfs.FeedEntity, 0, len(positions)),
	}

	for i, pos := range positions {
		id := fmt.Sprintf("v-%s-%d", pos.VehicleID, i)
		entity := &gtfs.FeedEntity{
			Id: &id,
			Vehicle: &gtfs.VehiclePosition{
				Vehicle: &gtfs.VehicleDescriptor{
					Id:    proto.String(pos.VehicleID),
					Label: proto.String(pos.RouteID), // Using RouteID as label for now
				},
				Position: &gtfs.Position{
					Latitude:  proto.Float32(pos.Latitude),
					Longitude: proto.Float32(pos.Longitude),
				},
				Timestamp: proto.Uint64(uint64(pos.Timestamp.Unix())),
			},
		}

		// Try to map to a trip
		if mapper != nil {
			if tripID, err := mapper.MapVehicleToTrip(context.Background(), pos); err == nil && tripID != "" {
				entity.Vehicle.Trip = &gtfs.TripDescriptor{
					TripId:      proto.String(tripID),
					RouteId:     proto.String(pos.RouteID),
					DirectionId: proto.Uint32(uint32(pos.DirectionID)),
				}
			}
		}

		// If no trip ID but we have route info, we can still provide a partial TripDescriptor
		if entity.Vehicle.Trip == nil && pos.RouteID != "" {
			entity.Vehicle.Trip = &gtfs.TripDescriptor{
				RouteId:     proto.String(pos.RouteID),
				DirectionId: proto.Uint32(uint32(pos.DirectionID)),
			}
		}

		feed.Entity = append(feed.Entity, entity)
	}

	return feed
}
