package rest

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/transitIOM/projectMercury/internal/adapters/input/middleware"
)

type Error struct {
	Code    int    `json:"code" example:"500"`
	Message string `json:"message" example:"Internal Server Error"`
}

func writeError(ctx context.Context, w http.ResponseWriter, code int, message string) {
	resp := Error{
		Code:    code,
		Message: message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		middleware.GetLogger(ctx).Error("Error writing response", "error", err)
	}
}
