package models

import "time"

type VehiclePosition struct {
	VehicleID   string    `json:"vehicle_id"`
	RouteID     string    `json:"route_id"`
	DirectionID int32     `json:"direction_id"`
	Latitude    float32   `json:"latitude"`
	Longitude   float32   `json:"longitude"`
	Timestamp   time.Time `json:"timestamp"`
	Occupancy   string    `json:"occupancy,omitempty"`
}

type TripUpdate struct {
	TripID    string          `json:"trip_id"`
	RouteID   string          `json:"route_id"`
	StartTime string          `json:"start_time"`
	StartDate string          `json:"start_date"`
	Stops     []StopTimeEvent `json:"stops"`
}

type StopTimeEvent struct {
	StopSequence uint32    `json:"stop_sequence"`
	Arrival      time.Time `json:"arrival,omitempty"`
	Departure    time.Time `json:"departure,omitempty"`
	StopID       string    `json:"stop_id"`
}
