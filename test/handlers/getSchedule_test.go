package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/transitIOM/projectMercury/api"
	"github.com/transitIOM/projectMercury/internal/handlers"
	"github.com/transitIOM/projectMercury/test/mocks"
)

func TestGetScheduleVersionID(t *testing.T) {
	mockSM := new(mocks.ObjectStorageManagerMock)
	versionID := "v123"
	mockSM.On("GetLatestGTFSVersionID").Return(versionID, nil)

	req := httptest.NewRequest("GET", "/schedule/version", nil)
	rr := httptest.NewRecorder()

	handler := handlers.GetScheduleVersionID(mockSM)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp api.GetVersionIDResponse
	json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.Equal(t, versionID, resp.Version)
}

func TestGetScheduleDownloadURL(t *testing.T) {
	mockSM := new(mocks.ObjectStorageManagerMock)
	testURL, _ := url.Parse("https://example.com/gtfs.zip")
	versionID := "v123"
	mockSM.On("GetLatestURL").Return(testURL, versionID, nil)

	req := httptest.NewRequest("GET", "/schedule/", nil)
	rr := httptest.NewRecorder()

	handler := handlers.GetScheduleDownloadURL(mockSM)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp api.GetTimetableResponse
	json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.Equal(t, versionID, resp.VersionID)
	assert.Equal(t, testURL.String(), resp.DownloadURL)
}
