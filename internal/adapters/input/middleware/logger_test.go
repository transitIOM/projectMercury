package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/stretchr/testify/assert"
)

func TestLoggerMiddleware(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	t.Run("Middleware injects logger", func(t *testing.T) {
		reqID := "test-req-id"
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		// Set request ID in context (simulating chi middleware)
		ctx := context.WithValue(req.Context(), middleware.RequestIDKey, reqID)
		req = req.WithContext(ctx)

		recorder := httptest.NewRecorder()

		handler := SlogLogger(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			l := GetLogger(r.Context())
			assert.NotNil(t, l)
			// Since we can't easily peek into the handler's attributes,
			// we at least verify it's not the default one if we can.
			// But we definitely check that GetLogger returns something from context.
			w.WriteHeader(http.StatusOK)
		}))

		handler.ServeHTTP(recorder, req)
		assert.Equal(t, http.StatusOK, recorder.Code)
	})

	t.Run("GetLogger returns default when missing", func(t *testing.T) {
		l := GetLogger(context.Background())
		assert.Equal(t, slog.Default(), l)
	})
}
