package signalr

import (
	"context"
	"encoding/base64"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/transitIOM/projectMercury/internal/domain/models"
)

type Adapter struct {
	ctx       context.Context
	mu        sync.RWMutex
	positions []models.VehiclePosition
	notifyCh  chan struct{}
	expiry    time.Duration
	timers    map[string]*time.Timer
}

func NewAdapter(ctx context.Context, expiry time.Duration) *Adapter {
	return &Adapter{
		ctx:      ctx,
		notifyCh: make(chan struct{}, 1),
		expiry:   expiry,
		timers:   make(map[string]*time.Timer),
	}
}

func (a *Adapter) GetVehiclePositions(ctx context.Context) ([]models.VehiclePosition, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.positions, nil
}

func (a *Adapter) NotifyChannel() <-chan struct{} {
	return a.notifyCh
}

func (a *Adapter) Start(ctx context.Context) {
	browser := rod.New().MustConnect()
	defer browser.MustClose()

	page := browser.MustPage("https://findmybus.im")

	go page.EachEvent(func(e *proto.NetworkResponseReceived) {
		if strings.Contains(strings.ToLower(e.Response.MIMEType), "application/octet-stream") {
			reqID := e.RequestID
			go func(id proto.NetworkRequestID) {
				result, err := proto.NetworkGetResponseBody{RequestID: id}.Call(page)
				if err != nil {
					return
				}

				var data []byte
				if result.Base64Encoded {
					data, _ = base64.StdEncoding.DecodeString(result.Body)
				} else {
					data = []byte(result.Body)
				}

				a.handleSignalRMessage(string(data))
			}(reqID)
		}
	})()

	page.MustWaitLoad()
	<-ctx.Done()
}

func (a *Adapter) handleSignalRMessage(response string) {
	positions, err := parseSignalRMessage(response)
	if err != nil {
		return
	}

	for _, pos := range positions {
		a.updatePosition(pos)
	}

	select {
	case a.notifyCh <- struct{}{}:
	default:
	}
}

func (a *Adapter) updatePosition(pos models.VehiclePosition) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if timer, ok := a.timers[pos.VehicleID]; ok {
		timer.Stop()
	}

	a.timers[pos.VehicleID] = time.AfterFunc(a.expiry, func() {
		a.removePosition(pos.VehicleID)
	})

	// Update or Add
	found := false
	for i, existing := range a.positions {
		if existing.VehicleID == pos.VehicleID {
			a.positions[i] = pos
			found = true
			break
		}
	}
	if !found {
		a.positions = append(a.positions, pos)
	}
}

func (a *Adapter) removePosition(vehicleID string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	slog.DebugContext(a.ctx, "Bus expired and removed", "vehicle_id", vehicleID)
	newPositions := make([]models.VehiclePosition, 0, len(a.positions))
	for _, p := range a.positions {
		if p.VehicleID != vehicleID {
			newPositions = append(newPositions, p)
		}
	}
	a.positions = newPositions
	delete(a.timers, vehicleID)

	select {
	case a.notifyCh <- struct{}{}:
	default:
	}
}
