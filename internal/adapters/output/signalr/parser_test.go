package signalr

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/transitIOM/projectMercury/internal/domain/models"
)

func TestAdapter_StateManagement(t *testing.T) {
	ctx := context.Background()
	expiry := 100 * time.Millisecond
	adapter := NewAdapter(ctx, expiry)

	t.Run("Update and Get Positions", func(t *testing.T) {
		pos := models.VehiclePosition{VehicleID: "bus-1", Latitude: 1.0}
		adapter.updatePosition(pos)

		positions, err := adapter.GetVehiclePositions(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(positions))
		assert.Equal(t, "bus-1", positions[0].VehicleID)

		// Update same bus
		pos.Latitude = 2.0
		adapter.updatePosition(pos)
		positions, _ = adapter.GetVehiclePositions(ctx)
		assert.Equal(t, 1, len(positions))
		assert.Equal(t, float32(2.0), positions[0].Latitude)
	})

	t.Run("Expiry", func(t *testing.T) {
		pos := models.VehiclePosition{VehicleID: "bus-expiry", Latitude: 1.0}
		adapter.updatePosition(pos)

		positions, _ := adapter.GetVehiclePositions(ctx)
		found := false
		for _, p := range positions {
			if p.VehicleID == "bus-expiry" {
				found = true
			}
		}
		assert.True(t, found)

		// Wait for expiry
		time.Sleep(expiry + 50*time.Millisecond)

		positions, _ = adapter.GetVehiclePositions(ctx)
		found = false
		for _, p := range positions {
			if p.VehicleID == "bus-expiry" {
				found = true
			}
		}
		assert.False(t, found)
	})

	t.Run("Notify Channel", func(t *testing.T) {
		ch := adapter.NotifyChannel()

		// Trigger update
		mockMsg := `{"type":1,"arguments":[{"locations":["0|notify|X|4|Inbound|54.123|-4.456|2026-03-15T02:44:00Z|0|0|"]}]}` + "\x1e"
		adapter.handleSignalRMessage(mockMsg)

		select {
		case <-ch:
			// Success
		case <-time.After(50 * time.Millisecond):
			t.Fatal("no notification received")
		}
	})
}

func TestParseLocationString(t *testing.T) {
	tests := []struct {
		name     string
		locStr   string
		wantErr  bool
		expected string // VehicleID
	}{
		{
			name:     "Valid Inbound",
			locStr:   "0|123|X|4|Inbound|54.123|-4.456|2026-03-15T02:44:00Z|0|0|",
			wantErr:  false,
			expected: "123",
		},
		{
			name:     "Valid Outbound",
			locStr:   "0|456|Y|1|Outbound|54.789|-4.012|2026-03-15T02:45:00Z|0|0|",
			wantErr:  false,
			expected: "456",
		},
		{
			name:    "Invalid Parts",
			locStr:  "0|123|X|4|Inbound",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pos, err := parseLocationString(tt.locStr)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, pos.VehicleID)
				if tt.name == "Valid Inbound" {
					assert.Equal(t, int32(1), pos.DirectionID)
					assert.Equal(t, float32(54.123), pos.Latitude)
					assert.Equal(t, float32(-4.456), pos.Longitude)
					assert.Equal(t, "4", pos.RouteID)
				} else if tt.name == "Valid Outbound" {
					assert.Equal(t, int32(0), pos.DirectionID)
				}
			}
		})
	}
}

func TestParseSignalRMessage(t *testing.T) {
	// A mock SignalR message with the expected format
	// Type 1 is a message update.
	// We use the record separator \x1e.
	mockMessage := `{"type":1,"arguments":[{"locations":["0|123|X|4|Inbound|54.123|-4.456|2026-03-15T02:44:00Z|0|0|"]}]}` + "\x1e"

	t.Run("Valid Message", func(t *testing.T) {
		positions, err := parseSignalRMessage(mockMessage)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(positions))
		assert.Equal(t, "123", positions[0].VehicleID)
	})

	t.Run("Invalid Message JSON", func(t *testing.T) {
		positions, err := parseSignalRMessage(`{"type":1,"arguments":[{"locations":["0|invalid"]}]}` + "\x1e")
		assert.Error(t, err)
		assert.Nil(t, positions)
	})

	t.Run("Empty Message", func(t *testing.T) {
		positions, err := parseSignalRMessage("\x1e")
		assert.Error(t, err)
		assert.Nil(t, positions)
	})

	t.Run("Multiple Records", func(t *testing.T) {
		multiMessage := `{"type":1,"arguments":[{"locations":["0|123|X|4|Inbound|54.123|-4.456|2026-03-15T02:44:00Z|0|0|"]}]}` + "\x1e" + `{"type":1,"arguments":[{"locations":["0|456|Y|1|Outbound|54.789|-4.012|2026-03-15T02:45:00Z|0|0|"]}]}` + "\x1e"
		positions, err := parseSignalRMessage(multiMessage)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(positions))
	})
}
