package rest_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/transitIOM/projectMercury/internal/adapters/input/rest"
	"github.com/transitIOM/projectMercury/internal/domain/models"
)

// Mock Service
type mockService struct {
	mock.Mock
}

func (m *mockService) GetGTFS(ctx context.Context) (io.ReadCloser, error) {
	args := m.Called(ctx)
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func (m *mockService) GetGTFSChecksum(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

func (m *mockService) PostReport(ctx context.Context, report models.UserReport) error {
	args := m.Called(ctx, report)
	return args.Error(0)
}

func (m *mockService) GetMessages(ctx context.Context) ([]models.Message, string, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.Message), args.String(1), args.Error(2)
}

func (m *mockService) PostMessage(ctx context.Context, message models.Message) (string, error) {
	args := m.Called(ctx, message)
	return args.String(0), args.Error(1)
}

func (m *mockService) Subscribe(ctx context.Context, feedType string) (<-chan any, error) {
	args := m.Called(ctx, feedType)
	return args.Get(0).(<-chan any), args.Error(1)
}

func TestHandler(t *testing.T) {
	ms := new(mockService)
	h := rest.NewHandler(ms)
	r := chi.NewRouter()
	h.RegisterRoutes(r)

	t.Run("GetGTFSSchedule - OK", func(t *testing.T) {
		ms.On("GetGTFSChecksum", mock.Anything).Return("checksum123", nil).Once()
		ms.On("GetGTFS", mock.Anything).Return(io.NopCloser(strings.NewReader("fake gtfs")), nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/schedule.zip", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "checksum123", w.Header().Get("X-GTFS-Checksum"))
		assert.Equal(t, "application/zip", w.Header().Get("Content-Type"))
		assert.Equal(t, "fake gtfs", w.Body.String())
	})

	t.Run("GetGTFSSchedule - 304 Not Modified", func(t *testing.T) {
		ms.On("GetGTFSChecksum", mock.Anything).Return("checksum123", nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/schedule.zip", nil)
		req.Header.Set("X-GTFS-Checksum", "checksum123")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotModified, w.Code)
		assert.Equal(t, "checksum123", w.Header().Get("X-GTFS-Checksum"))
	})

	t.Run("GetMessages", func(t *testing.T) {
		msgs := []models.Message{{ID: "1", Content: "Hello"}}
		ms.On("GetMessages", mock.Anything).Return(msgs, "v1", nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/messages", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "v1", w.Header().Get("X-Message-Version"))

		var resp []models.Message
		err := json.NewDecoder(w.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(resp))
		assert.Equal(t, "Hello", resp[0].Content)
	})

	t.Run("PostMessage", func(t *testing.T) {
		msg := models.Message{Content: "New"}
		ms.On("PostMessage", mock.Anything, msg).Return("v2", nil).Once()

		body, _ := json.Marshal(msg)
		req := httptest.NewRequest(http.MethodPost, "/messages", bytes.NewReader(body))
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Equal(t, "v2", w.Header().Get("X-Message-Version"))
	})

	t.Run("PostReport - Error", func(t *testing.T) {
		report := models.UserReport{Title: "Bug"}
		ms.On("PostReport", mock.Anything, report).Return(assert.AnError).Once()

		body, _ := json.Marshal(report)
		req := httptest.NewRequest(http.MethodPost, "/report", bytes.NewReader(body))
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), assert.AnError.Error())
	})

	t.Run("PostMessage - Invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/messages", strings.NewReader("invalid"))
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
