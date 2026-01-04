package tools

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	log "github.com/sirupsen/logrus"
)

var BusLocations struct {
	Data  []BusLocation
	Mutex sync.RWMutex
}

func InitializeBrowser() {
	browser := rod.New().MustConnect()
	defer browser.MustClose()

	page := browser.MustPage("https://findmybus.im")
	log.Info("connecting to https://findmybus.im")

	// Start listening for network response headers (Passive Monitoring)
	go page.EachEvent(func(e *proto.NetworkResponseReceived) {
		// Check for the specific URL and MIMEType (using capital MIME)
		isRelevant := strings.Contains(strings.ToLower(e.Response.MIMEType), "application/octet-stream")

		if isRelevant {
			// Use RequestID (capital ID)
			reqID := e.RequestID
			url := e.Response.URL

			// Fetch body asynchronously
			go func(id proto.NetworkRequestID, u string) {
				// Call the proto command to get the response body
				result, err := proto.NetworkGetResponseBody{RequestID: id}.Call(page)
				if err != nil {
					// Request might be gone or empty
					return
				}

				log.Debugf("Captured data from: %s", u)

				var data []byte
				// Handle Base64 encoding if the browser flags it
				if result.Base64Encoded {
					data, err = base64.StdEncoding.DecodeString(result.Body)
					if err != nil {
						log.Errorf("Failed to decode base64 body: %v", err)
						return
					}
				} else {
					data = []byte(result.Body)
				}

				// Update logic
				BusLocations.Mutex.Lock()
				defer BusLocations.Mutex.Unlock()

				// The fix: updateInMemBusLocations now handles SignalR message splitting
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

	select {}
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

// FIX: This function now splits the response by the SignalR Record Separator (\x1e)
// and iterates over the messages.
func updateInMemBusLocations(response string) (err error) {
	// Split by the Record Separator (RS) character used by SignalR
	messages := strings.Split(response, "\x1e")
	var allLocations []BusLocation
	foundLocations := false

	for _, msg := range messages {
		msg = strings.TrimSpace(msg)
		if msg == "" {
			continue // Skip empty or purely whitespace segments
		}

		busLocations, err := parseBusLocations(msg)
		if err != nil {
			// Log the error but continue to the next message in the stream
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

	BusLocations.Data = allLocations
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

// FIX: This function now checks for SignalR message Type=1 (Invocation)
func parseBusLocations(responseStr string) ([]BusLocation, error) {
	var response SignalRResponse

	err := json.Unmarshal([]byte(responseStr), &response)
	if err != nil {
		// Use a more specific error message for unmarshalling failures
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// SignalR messages with location data are typically Type 1 (Invocation).
	// Other types (e.g., 6 for ping) should be ignored gracefully.
	if response.Type != 1 {
		return []BusLocation{}, nil
	}

	if len(response.Arguments) == 0 || (len(response.Arguments) > 0 && len(response.Arguments[0].Locations) == 0) {
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
