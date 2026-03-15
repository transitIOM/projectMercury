package sse_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/transitIOM/projectMercury/internal/adapters/input/sse"
	"github.com/transitIOM/projectMercury/internal/domain/models"
)

type mockService struct {
	mock.Mock
}

func (m *mockService) GetGTFS(ctx context.Context) (io.ReadCloser, error) {
	return nil, nil
}
func (m *mockService) GetGTFSChecksum(ctx context.Context) (string, error) {
	return "", nil
}
func (m *mockService) PostReport(ctx context.Context, report models.UserReport) error {
	return nil
}
func (m *mockService) GetMessages(ctx context.Context) ([]models.Message, string, error) {
	return nil, "", nil
}
func (m *mockService) PostMessage(ctx context.Context, message models.Message) (string, error) {
	return "", nil
}
func (m *mockService) Subscribe(ctx context.Context, feedType string) (<-chan any, error) {
	args := m.Called(ctx, feedType)
	return args.Get(0).(<-chan any), args.Error(1)
}

func TestSSEHandler(t *testing.T) {
	ms := new(mockService)
	handler := sse.NewHandler(ms)

	t.Run("Connect and receive message", func(t *testing.T) {
		events := make(chan any, 1)
		ms.On("Subscribe", mock.Anything, "all").Return((<-chan any)(events), nil).Once()

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		req := httptest.NewRequest(http.MethodGet, "/stream", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		// Send an event after a short delay
		go func() {
			time.Sleep(10 * time.Millisecond)
			events <- models.Message{Content: "test message"}
		}()

		handler.ServeHTTP(w, req)

		assert.Equal(t, "text/event-stream", w.Header().Get("Content-Type"))
		assert.Contains(t, w.Body.String(), "event: connected")
		assert.Contains(t, w.Body.String(), "event: message")
		assert.Contains(t, w.Body.String(), "test message")
	})

	t.Run("Path-based filtering - vehicle-positions", func(t *testing.T) {
		events := make(chan any, 1)
		ms.On("Subscribe", mock.Anything, "vehicle-positions").Return((<-chan any)(events), nil).Once()

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		req := httptest.NewRequest(http.MethodGet, "/stream/vehicle-positions", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Contains(t, w.Body.String(), "welcome to vehicle-positions stream")
	})

	t.Run("Subscribe Error", func(t *testing.T) {
		ms.On("Subscribe", mock.Anything, "all").Return((<-chan any)(nil), assert.AnError).Once()

		req := httptest.NewRequest(http.MethodGet, "/stream", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), assert.AnError.Error())
	})
}
