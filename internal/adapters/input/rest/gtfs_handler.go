package rest

import (
	"io"
	"net/http"

	"github.com/transitIOM/projectMercury/internal/adapters/input/middleware"
)

// GetGTFSSchedule serves the GTFS.zip file directly.
// @Summary      Get GTFS static schedule
// @Description  Serves the GTFS.zip file. Supports conditional requests via X-GTFS-Checksum header.
// @Tags         schedule
// @Produce      application/zip
// @Param        X-GTFS-Checksum  header    string  false  "Client's cached GTFS checksum"
// @Success      200              {file}    binary
// @Success      304              {string}  string  "Not Modified"
// @Header       200,304          {string}  X-GTFS-Checksum  "The SHA256 checksum of the served file"
// @Failure      500              {object}  Error
// @Router       /schedule.zip [get]
func (h *Handler) GetGTFSSchedule(w http.ResponseWriter, r *http.Request) {
	checksum, _ := h.service.GetGTFSChecksum(r.Context())

	// Check if client sent a checksum
	clientChecksum := r.Header.Get("X-GTFS-Checksum")
	if clientChecksum != "" && clientChecksum == checksum {
		w.Header().Set("X-GTFS-Checksum", checksum)
		w.WriteHeader(http.StatusNotModified)
		return
	}

	reader, err := h.service.GetGTFS(r.Context())
	if err != nil {
		middleware.GetLogger(r.Context()).Error("Failed to get GTFS schedule", "error", err)
		writeError(r.Context(), w, http.StatusInternalServerError, err.Error())
		return
	}
	defer reader.Close()

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=\"GTFS.zip\"")
	if checksum != "" {
		w.Header().Set("X-GTFS-Checksum", checksum)
	}
	io.Copy(w, reader)
}
