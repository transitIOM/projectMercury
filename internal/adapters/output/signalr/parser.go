package signalr

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/transitIOM/projectMercury/internal/domain/models"
)

type SignalRResponse struct {
	Type      int `json:"type"`
	Arguments []struct {
		Locations []string `json:"locations"`
	} `json:"arguments"`
}

func parseSignalRMessage(response string) ([]models.VehiclePosition, error) {
	messages := strings.Split(response, "\x1e")
	var allPositions []models.VehiclePosition

	for _, msg := range messages {
		msg = strings.TrimSpace(msg)
		if msg == "" {
			continue
		}

		var srResponse SignalRResponse
		if err := json.Unmarshal([]byte(msg), &srResponse); err != nil || srResponse.Type != 1 {
			continue
		}

		for _, arg := range srResponse.Arguments {
			for _, locStr := range arg.Locations {
				pos, err := parseLocationString(locStr)
				if err == nil {
					allPositions = append(allPositions, pos)
				}
			}
		}
	}

	if len(allPositions) == 0 {
		return nil, fmt.Errorf("no valid positions found in message")
	}

	return allPositions, nil
}

func parseLocationString(locStr string) (models.VehiclePosition, error) {
	parts := strings.Split(locStr, "|")
	if len(parts) < 10 {
		return models.VehiclePosition{}, fmt.Errorf("invalid parts")
	}

	lat, _ := strconv.ParseFloat(strings.TrimSpace(parts[5]), 32)
	lon, _ := strconv.ParseFloat(strings.TrimSpace(parts[6]), 32)
	ts, _ := time.Parse(time.RFC3339, strings.TrimSpace(parts[7]))

	direction := 0
	if strings.ToLower(strings.TrimSpace(parts[4])) == "inbound" {
		direction = 1
	}

	return models.VehiclePosition{
		VehicleID:   strings.TrimSpace(parts[1]),
		RouteID:     strings.TrimSpace(parts[3]),
		DirectionID: int32(direction),
		Latitude:    float32(lat),
		Longitude:   float32(lon),
		Timestamp:   ts,
	}, nil
}
