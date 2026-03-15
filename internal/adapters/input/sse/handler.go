package sse

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
	"google.golang.org/protobuf/proto"

	"github.com/transitIOM/projectMercury/internal/adapters/input/middleware"
	"github.com/transitIOM/projectMercury/internal/domain/models"
	"github.com/transitIOM/projectMercury/internal/ports/input"
)

type Handler struct {
	service input.TransitService
}

func NewHandler(service input.TransitService) *Handler {
	return &Handler{service: service}
}

// @Summary      Real-time transit stream (SSE)
// @Description  A persistent Server-Sent Events stream for live transit updates. Supports path-based filtering: /stream/all, /stream/vehicle-positions, /stream/trip-updates, /stream/service-alerts. Data is sent as base64-encoded GTFS-RT protobuf.
// @Tags         realtime
// @Produce      text/event-stream
// @Success      200  {string}  string  "Streaming connection established"
// @Router       /stream [get]
// @Router       /stream/all [get]
// @Router       /stream/vehicle-positions [get]
// @Router       /stream/trip-updates [get]
// @Router       /stream/service-alerts [get]
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Determine feed type from path
	feedType := "all"
	if strings.Contains(r.URL.Path, "vehicle-positions") {
		feedType = "vehicle-positions"
	} else if strings.Contains(r.URL.Path, "trip-updates") {
		feedType = "trip-updates"
	} else if strings.Contains(r.URL.Path, "service-alerts") {
		feedType = "service-alerts"
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Transfer-Encoding", "chunked")

	flusher, ok := w.(http.Flusher)
	if !ok {
		middleware.GetLogger(r.Context()).Error("SSE unsupported by client")
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	events, err := h.service.Subscribe(r.Context(), feedType)
	if err != nil {
		middleware.GetLogger(r.Context()).Error("Failed to subscribe to service", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	middleware.GetLogger(r.Context()).Info("Client connected to SSE stream", "type", feedType)
	fmt.Fprintf(w, "event: connected\ndata: joined %s stream\n", feedType)
	flusher.Flush()

	for {
		select {
		case <-r.Context().Done():
			return
		case event := <-events:
			var eventName string
			var data []byte

			// If it's a FeedMessage (GTFS-RT protobuf)
			if msg, ok := event.(*gtfs.FeedMessage); ok {
				var err error
				protoData, err := proto.Marshal(msg)
				if err != nil {
					middleware.GetLogger(r.Context()).Error("Failed to marshal GTFS-RT protobuf", "error", err)
					continue
				}
				// Encode to base64 for SSE text-based transport
				encoded := base64.StdEncoding.EncodeToString(protoData)
				data = []byte(encoded)

				// Determine event name based on content if possible, or use generic
				eventName = "gtfs-rt"
			} else {
				// Fallback for non-protobuf messages (like system messages)
				var err error
				data, err = json.Marshal(event)
				if err != nil {
					continue
				}

				switch event.(type) {
				case models.VehiclePosition, []models.VehiclePosition:
					eventName = "vehicle-positions"
				case models.TripUpdate, []models.TripUpdate:
					eventName = "trip-updates"
				case models.ServiceAlert, []models.ServiceAlert:
					eventName = "service-alerts"
				default:
					eventName = "message"
				}
			}

			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", eventName, data)
			flusher.Flush()
		}
	}
}
