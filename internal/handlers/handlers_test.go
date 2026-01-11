package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/transitIOM/projectMercury/api"
	"github.com/transitIOM/projectMercury/internal/tools"
)

func TestGetBusLocations(t *testing.T) {
	req := httptest.NewRequest("GET", "/locations/", nil)
	rr := httptest.NewRecorder()

	GetBusLocations(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}

	var response api.GetBusLocationsResponse
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response.Code != http.StatusOK {
		t.Errorf("response code mismatch: got %v want %v", response.Code, http.StatusOK)
	}
}

func TestGetMessages(t *testing.T) {
	tools.CurrentMessageLog = *bytes.NewBufferString("line1\nline2\nline3\nline4\n")
	tools.CurrentMessageVersionID = "test-version-id"

	req := httptest.NewRequest("GET", "/messages", nil)
	rr := httptest.NewRecorder()

	GetMessages(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}

	var response api.GetMessagesResponse
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response.Code != http.StatusOK {
		t.Errorf("response code mismatch: got %v want %v", response.Code, http.StatusOK)
	}

	expectedMessages := "line2\nline3\nline4\n"
	if response.Messages != expectedMessages {
		t.Errorf("expected messages %q, got %q", expectedMessages, response.Messages)
	}
}
