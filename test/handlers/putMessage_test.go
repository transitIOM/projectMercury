package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/transitIOM/projectMercury/api"
	"github.com/transitIOM/projectMercury/internal/handlers"
	"github.com/transitIOM/projectMercury/test/mocks"
)

func TestPutMessage(t *testing.T) {
	mockSM := new(mocks.ObjectStorageManagerMock)
	versionID := "m123"
	mockSM.On("AppendMessage", mock.Anything).Return(versionID, nil)

	formData := url.Values{}
	formData.Set("message", "Test updated message")
	req := httptest.NewRequest("PUT", "/messages/", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	handler := handlers.PutMessage(mockSM)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusAccepted, rr.Code)
	var resp api.PutMessageResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, versionID, resp.VersionID)
}
