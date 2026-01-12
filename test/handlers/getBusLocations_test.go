package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/transitIOM/projectMercury/api"
	"github.com/transitIOM/projectMercury/internal/handlers"
)

func TestGetBusLocations(t *testing.T) {
	req := httptest.NewRequest("GET", "/locations/", nil)
	rr := httptest.NewRecorder()

	handlers.GetBusLocations(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response api.GetBusLocationsResponse
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.Code)
	// tools.GetAllBuses() returns an empty slice by default in tests
	assert.Equal(t, "[]", response.Locations)
}
