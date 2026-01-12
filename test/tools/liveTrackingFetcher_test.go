package tools_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/transitIOM/projectMercury/internal/tools"
)

func TestParseLocationString(t *testing.T) {
	tests := []struct {
		name    string
		locStr  string
		want    tools.BusLocation
		wantErr bool
	}{
		{
			name:   "valid location string",
			locStr: "Driver1|Bus1|12:00|Route1|Inbound|54.123|-4.567|2026-01-11T03:55:00Z|1|Extra",
			want: tools.BusLocation{
				DriverNumber:  "Driver1",
				BusID:         "Bus1",
				DepartureTime: "12:00",
				RouteNumber:   "Route1",
				Direction:     "Inbound",
				Latitude:      54.123,
				Longitude:     -4.567,
				Timestamp:     time.Date(2026, 1, 11, 3, 55, 0, 0, time.UTC),
				Unknown1:      1,
				Unknown2:      "Extra",
			},
			wantErr: false,
		},
		{
			name:    "invalid format (too few parts)",
			locStr:  "part1|part2",
			want:    tools.BusLocation{},
			wantErr: true,
		},
		{
			name:    "invalid timestamp",
			locStr:  "D|B|T|R|D|54| -4|invalid-time|1|E",
			want:    tools.BusLocation{},
			wantErr: true,
		},
		{
			name:    "invalid latitude",
			locStr:  "D|B|T|R|D|not-float| -4|2026-01-11T03:55:00Z|1|E",
			want:    tools.BusLocation{},
			wantErr: true,
		},
		{
			name:    "invalid longitude",
			locStr:  "D|B|T|R|D|54|not-float|2026-01-11T03:55:00Z|1|E",
			want:    tools.BusLocation{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tools.ParseLocationString(tt.locStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseLocationString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseLocationString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseBusLocations(t *testing.T) {
	tests := []struct {
		name        string
		responseStr string
		want        []tools.BusLocation
		wantErr     bool
	}{
		{
			name:        "valid SignalR message",
			responseStr: `{"type":1,"target":"updateLocations","arguments":[{"locations":["D1|B1|T1|R1|Dir1|54.1|-4.5|2026-01-11T03:55:00Z|1|E1"]}]}`,
			want: []tools.BusLocation{
				{
					DriverNumber:  "D1",
					BusID:         "B1",
					DepartureTime: "T1",
					RouteNumber:   "R1",
					Direction:     "Dir1",
					Latitude:      54.1,
					Longitude:     -4.5,
					Timestamp:     time.Date(2026, 1, 11, 3, 55, 0, 0, time.UTC),
					Unknown1:      1,
					Unknown2:      "E1",
				},
			},
			wantErr: false,
		},
		{
			name:        "empty arguments",
			responseStr: `{"type":1,"target":"updateLocations","arguments":[]}`,
			want:        []tools.BusLocation{},
			wantErr:     false,
		},
		{
			name:        "wrong message type",
			responseStr: `{"type":2}`,
			want:        []tools.BusLocation{},
			wantErr:     false,
		},
		{
			name:        "invalid JSON",
			responseStr: `invalid json`,
			want:        nil,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tools.ParseBusLocations(tt.responseStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseBusLocations() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseBusLocations() = %v, want %v", got, tt.want)
			}
		})
	}
}
