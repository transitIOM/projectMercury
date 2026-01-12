package middleware_test

import (
	"crypto"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/transitIOM/projectMercury/internal/middleware"
)

func TestAPIKeyAuth(t *testing.T) {
	originalHash := middleware.ExpectedHash
	testKey := "test-api-key"
	h := crypto.SHA256.New()
	h.Write([]byte(testKey))
	middleware.ExpectedHash = hex.EncodeToString(h.Sum(nil))
	defer func() { middleware.ExpectedHash = originalHash }()

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handlerToTest := middleware.APIKeyAuth(nextHandler)

	tests := []struct {
		name           string
		apiKeyHeader   string
		wantStatusCode int
	}{
		{
			name:           "valid API key",
			apiKeyHeader:   testKey,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "invalid API key",
			apiKeyHeader:   "wrong-key",
			wantStatusCode: http.StatusUnauthorized,
		},
		{
			name:           "missing API key",
			apiKeyHeader:   "",
			wantStatusCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			if tt.apiKeyHeader != "" {
				req.Header.Set("X-API-Key", tt.apiKeyHeader)
			}
			rr := httptest.NewRecorder()

			handlerToTest.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatusCode {
				t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, tt.wantStatusCode)
			}
		})
	}
}
