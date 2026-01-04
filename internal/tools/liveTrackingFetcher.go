package tools

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-rod/rod"
	log "github.com/sirupsen/logrus"
)

var BusLocations struct {
	Data  []BusLocation
	Mutex sync.RWMutex
}

func InitializeBrowser() {
	browser := rod.New().MustConnect()
	defer browser.MustClose()

	page := browser.MustPage()

	router := page.HijackRequests()
	defer router.MustStop()

	router.MustAdd("*", func(ctx *rod.Hijack) {
		ctx.MustLoadResponse()

		contentType := ctx.Response.Headers().Get("Content-Type")
		if strings.Contains(strings.ToLower(contentType), "application/octet-stream") {
			log.Info("Captured locations")
			log.Debugf("URL: %s | Method: %s | Unknown2: %d | Content-Type %s", ctx.Request.URL().String(), ctx.Request.Method(), ctx.Response.Payload().ResponseCode, contentType)

			base64Body := ctx.Response.Body()
			log.Debugf("Body length: %d bytes\n", len(base64Body))

			body, err := base64.StdEncoding.DecodeString(base64Body)
			if err != nil {
				log.Errorf("Failed to decode base64 body: %v", err)
				return
			}
			log.Debugf("Decoded body: %s", string(body))
			BusLocations.Mutex.Lock()
			defer BusLocations.Mutex.Unlock()
			err = updateInMemBusLocations(string(body))
			if err != nil {
				log.Errorf("Failed to parse bus locations: %v", err)
			}
			log.Debugf("Bus locations updated")

		}
	})

	go router.Run()

	page.MustNavigate("https://findmybus.im")

	log.Info("Browser opened, monitoring traffic...")
}

type BusLocation struct {
	DriverNumber  string    `json:"driver_number,omitempty"`
	BusID         string    `json:"bus_id"`
	DepartureTime string    `json:"departure_time"`
	RouteNumber   string    `json:"route_number"`
	Direction     string    `json:"direction"`
	Latitude      float64   `json:"latitude"`
	Longitude     float64   `json:"longitude"`
	Timestamp     time.Time `json:"timestamp"`
	Unknown1      int       `json:"unknown1,omitempty"`
	Unknown2      string    `json:"unknown2,omitempty"`
}

func updateInMemBusLocations(response string) (err error) {
	busLocations, err := parseBusLocations(response)
	if err != nil {
		return err
	}
	if busLocations == nil {
		return fmt.Errorf("no bus locations found")
	}
	BusLocations.Data = busLocations
	return nil
}

type SignalRResponse struct {
	Type      int                `json:"type"`
	Target    string             `json:"target"`
	Arguments []SignalRArguments `json:"arguments"`
}

type SignalRArguments struct {
	Locations []string `json:"locations"`
}

func parseBusLocations(responseStr string) ([]BusLocation, error) {
	var response SignalRResponse

	err := json.Unmarshal([]byte(responseStr), &response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	if len(response.Arguments) == 0 {
		return []BusLocation{}, nil
	}

	locations := response.Arguments[0].Locations
	busLocations := make([]BusLocation, 0, len(locations))

	for _, locStr := range locations {
		busLoc, err := parseLocationString(locStr)
		if err != nil {
			log.Warnf("Warning: failed to parse location: %s, error: %v\n", locStr, err)
			continue
		}
		busLocations = append(busLocations, busLoc)
	}

	return busLocations, nil
}

func parseLocationString(locStr string) (BusLocation, error) {
	parts := strings.Split(locStr, "|")

	if len(parts) < 10 {
		return BusLocation{}, fmt.Errorf("invalid location string format, expected 10 parts, got %d", len(parts))
	}

	timestamp, err := time.Parse(time.RFC3339, parts[7])
	if err != nil {
		return BusLocation{}, fmt.Errorf("failed to parse timestamp: %w", err)
	}

	var lat, lon float64
	_, err = fmt.Sscanf(parts[5], "%f", &lat)
	if err != nil {
		return BusLocation{}, fmt.Errorf("failed to parse latitude: %w", err)
	}

	_, err = fmt.Sscanf(parts[6], "%f", &lon)
	if err != nil {
		return BusLocation{}, fmt.Errorf("failed to parse longitude: %w", err)
	}

	var unknown1 int
	_, err = fmt.Sscanf(parts[8], "%d", &unknown1)
	if err != nil {
		return BusLocation{}, fmt.Errorf("failed to parse unknown1: %w", err)
	}

	return BusLocation{
		DriverNumber:  parts[0],
		BusID:         parts[1],
		DepartureTime: parts[2],
		RouteNumber:   parts[3],
		Direction:     parts[4],
		Latitude:      lat,
		Longitude:     lon,
		Timestamp:     timestamp,
		Unknown1:      unknown1,
		Unknown2:      parts[9],
	}, nil
}
