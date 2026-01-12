package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/transitIOM/projectMercury/api"
	"github.com/transitIOM/projectMercury/internal/handlers"
	"github.com/transitIOM/projectMercury/test/mocks"
)

func TestGetMessages(t *testing.T) {
	mockSM := new(mocks.ObjectStorageManagerMock)
	logContent := `{"timestamp":"...","message":"hello"}`
	buffer := bytes.NewBufferString(logContent)
	mockSM.On("GetLatestLog").Return(buffer, nil)
	mockSM.On("GetLatestMessageVersionID").Return("m123", nil)

	req := httptest.NewRequest("GET", "/messages/", nil)
	rr := httptest.NewRecorder()

	handler := handlers.GetMessages(mockSM)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp api.GetMessagesResponse
	json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.Equal(t, logContent, resp.Messages)
	assert.Equal(t, "m123", resp.VersionID)
}

func TestGetMessageLogVersionID(t *testing.T) {
	mockSM := new(mocks.ObjectStorageManagerMock)
	mockSM.On("GetLatestMessageVersionID").Return("m123", nil)

	req := httptest.NewRequest("GET", "/messages/version", nil)
	rr := httptest.NewRecorder()

	handler := handlers.GetMessageLogVersionID(mockSM)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp api.GetVersionIDResponse
	json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.Equal(t, "m123", resp.Version)
}
