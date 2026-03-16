package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
)

type contextKey string

const loggerKey contextKey = "slog_logger"

// SlogLogger is a middleware that injects the slog logger into the request context.
// It also adds the RequestID from Chi's middleware as a field to the logger.
func SlogLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := middleware.GetReqID(r.Context())

			// Create a logger with the request ID
			requestLogger := logger.With(slog.String("request_id", requestID))

			// Update context with the logger
			ctx := context.WithValue(r.Context(), loggerKey, requestLogger)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetLogger retrieves the slog logger from the context.
// If no logger is found, it returns the default slog logger.
func GetLogger(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}
