package tools

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	log "github.com/sirupsen/logrus"
)

const expiryTime = time.Minute * 2

var BusLocations struct {
	Buses  map[string]*TrackedBus
	Mutex  sync.RWMutex
	expiry time.Duration
}

type TrackedBus struct {
	Location BusLocation
	timer    *time.Timer
}

type BusLocation struct {
	DriverNumber  string    `json:"-"`
	BusID         string    `json:"bus_id"`
	DepartureTime string    `json:"departure_time"`
	RouteNumber   string    `json:"route_number"`
	Direction     string    `json:"direction"`
	Latitude      float64   `json:"latitude"`
	Longitude     float64   `json:"longitude"`
	Timestamp     time.Time `json:"-"`
	Unknown1      int       `json:"-"`
	Unknown2      string    `json:"-"`
}

func init() {
	BusLocations.Buses = make(map[string]*TrackedBus)
	BusLocations.expiry = expiryTime
}

func InitializeBrowser(ctx context.Context) {
	browser := rod.New().MustConnect()
	defer browser.MustClose()

	page := browser.MustPage("https://findmybus.im")

	go page.EachEvent(func(e *proto.NetworkResponseReceived) {
		isRelevant := strings.Contains(strings.ToLower(e.Response.MIMEType), "application/octet-stream")

		if isRelevant {
			reqID := e.RequestID
			url := e.Response.URL

			go func(id proto.NetworkRequestID, u string) {
				result, err := proto.NetworkGetResponseBody{RequestID: id}.Call(page)
				if err != nil {
					return
				}

				log.Debugf("Captured data from: %s", u)

				var data []byte
				if result.Base64Encoded {
					data, err = base64.StdEncoding.DecodeString(result.Body)
					if err != nil {
						log.Errorf("Failed to decode base64 body: %v", err)
						return
					}
				} else {
					data = []byte(result.Body)
				}

				BusLocations.Mutex.Lock()
				defer BusLocations.Mutex.Unlock()

				err = updateInMemBusLocations(string(data))
				if err != nil {
					log.Debugf("Skipping parse (likely not location data or empty frame): %v", err)
				} else {
					log.Debug("Bus locations updated successfully")
				}
			}(reqID, url)
		}
	})()

	page.MustWaitLoad()

	<-ctx.Done()
	log.Info("browser closed gracefully")
}

func updateInMemBusLocations(response string) (err error) {
	messages := strings.Split(response, "\x1e")
	var allLocations []BusLocation
	foundLocations := false

	for _, msg := range messages {
		msg = strings.TrimSpace(msg)
		if msg == "" {
			continue
		}

		busLocations, err := ParseBusLocations(msg)
		if err != nil {
			log.Debugf("Error parsing SignalR message: %v. Message: %s", err, msg)
			continue
		}

		if len(busLocations) > 0 {
			allLocations = append(allLocations, busLocations...)
			foundLocations = true
		}
	}

	if !foundLocations {
		return fmt.Errorf("no bus locations found in any message frame")
	}

	for _, loc := range allLocations {
		if tracked, exists := BusLocations.Buses[loc.BusID]; exists {
			tracked.timer.Stop()
		}

		timer := time.AfterFunc(BusLocations.expiry, func(busID string) func() {
			return func() {
				removeBus(busID)
			}
		}(loc.BusID))

		BusLocations.Buses[loc.BusID] = &TrackedBus{
			Location: loc,
			timer:    timer,
		}
	}

	return nil
}

func removeBus(busID string) {
	BusLocations.Mutex.Lock()
	defer BusLocations.Mutex.Unlock()
	log.Debugf("Bus %s expired and removed", busID)
	delete(BusLocations.Buses, busID)
}

func GetAllBuses() []BusLocation {
	BusLocations.Mutex.RLock()
	defer BusLocations.Mutex.RUnlock()

	locations := make([]BusLocation, 0, len(BusLocations.Buses))
	for _, tracked := range BusLocations.Buses {
		locations = append(locations, tracked.Location)
	}
	return locations
}

type SignalRResponse struct {
	Type      int                `json:"type"`
	Target    string             `json:"target"`
	Arguments []SignalRArguments `json:"arguments"`
}

type SignalRArguments struct {
	Locations []string `json:"locations"`
}

func ParseBusLocations(responseStr string) ([]BusLocation, error) {
	var response SignalRResponse

	err := json.Unmarshal([]byte(responseStr), &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	if response.Type != 1 {
		return []BusLocation{}, nil
	}

	if len(response.Arguments) == 0 || (len(response.Arguments) > 0 && len(response.Arguments[0].Locations) == 0) {
		return []BusLocation{}, nil
	}

	locations := response.Arguments[0].Locations
	busLocations := make([]BusLocation, 0, len(locations))

	for _, locStr := range locations {
		busLoc, err := ParseLocationString(locStr)
		if err != nil {
			log.Warnf("Warning: failed to parse location: %s, error: %v\n", locStr, err)
			continue
		}
		busLocations = append(busLocations, busLoc)
	}

	return busLocations, nil
}

func ParseLocationString(locStr string) (BusLocation, error) {
	parts := strings.Split(locStr, "|")

	if len(parts) < 10 {
		return BusLocation{}, fmt.Errorf("invalid location string format, expected at least 10 parts, got %d", len(parts))
	}

	// Trim whitespace from all parts
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}

	timestamp, err := time.Parse(time.RFC3339, parts[7])
	if err != nil {
		return BusLocation{}, fmt.Errorf("failed to parse timestamp '%s': %w", parts[7], err)
	}

	lat, err := strconv.ParseFloat(parts[5], 64)
	if err != nil {
		return BusLocation{}, fmt.Errorf("failed to parse latitude '%s': %w", parts[5], err)
	}

	lon, err := strconv.ParseFloat(parts[6], 64)
	if err != nil {
		return BusLocation{}, fmt.Errorf("failed to parse longitude '%s': %w", parts[6], err)
	}

	unknown1, err := strconv.Atoi(parts[8])
	if err != nil {
		// If it's not an int, we'll log it and set to 0 rather than failing the whole parse
		// as this field is "Unknown1" and might not be critical.
		log.Debugf("failed to parse unknown1 integer '%s': %v", parts[8], err)
		unknown1 = 0
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
