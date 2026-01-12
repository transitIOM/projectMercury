package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/transitIOM/projectMercury/api"
	"github.com/transitIOM/projectMercury/internal/tools"
)

// GetMessages godoc
// @Summary      Get the latest messages from the log
// @Description  Retrieves the most recent messages from the log in JSONL format, along with the current version ID.
// @Tags         messages
// @Produce      json
// @Param        n     query     int    false  "Number of latest messages to retrieve (defaults to 3)"
// @Success      200  {object}  api.GetMessagesResponse
// @Failure      500  {object}  api.Error
// @Router       /messages/ [get]
func GetMessages(sm tools.ObjectStorageManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Debug("Handling GetMessages request")

		var response api.GetMessagesResponse

		messageCount := 3
		if nStr := r.URL.Query().Get("n"); nStr != "" {
			if n, err := strconv.Atoi(nStr); err == nil && n > 0 {
				messageCount = n
			}
		}

		v, err := sm.GetLatestMessageVersionID()
		if err != nil {
			if errors.Is(err, tools.NoMessageLogFound) {
				log.Debug(err)
				response = api.GetMessagesResponse{
					Code:      http.StatusNoContent,
					Messages:  "",
					VersionID: "",
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(response.Code)
				if err = json.NewEncoder(w).Encode(response); err != nil {
					log.Errorf("Failed to encode response: %v", err)
				}
				return
			}
			log.Error(err)
			api.InternalErrorHandler(w)
			return
		}

		log.Debugf("Retrieving latest message log from storage, requesting last %d messages", messageCount)
		b, err := sm.GetLatestLog()
		if err != nil {
			log.Error(err)
			api.InternalErrorHandler(w)
			return
		}

		b = getLastNLines(b, messageCount)

		log.Debugf("Retrieved %d bytes from message log after truncation", b.Len())

		if b.Len() == 0 {
			response = api.GetMessagesResponse{
				Code:      http.StatusNoContent,
				Messages:  "",	
				VersionID: "",
			}
		} else {
			response = api.GetMessagesResponse{
				Code:      http.StatusOK,
				Messages:  b.String(),
				VersionID: v,
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(response.Code)
		if err = json.NewEncoder(w).Encode(response); err != nil {
			log.Errorf("Failed to encode response: %v", err)
		}
	}
}

func getLastNLines(b *bytes.Buffer, n int) *bytes.Buffer {
	if b == nil || b.Len() == 0 {
		return &bytes.Buffer{}
	}

	data := b.Bytes()
	lines := bytes.Split(data, []byte("\n"))

	// Remove trailing empty line if it exists
	if len(lines) > 0 && len(lines[len(lines)-1]) == 0 {
		lines = lines[:len(lines)-1]
	}

	start := 0
	if len(lines) > n {
		start = len(lines) - n
	}

	result := new(bytes.Buffer)
	for i := start; i < len(lines); i++ {
		result.Write(lines[i])
		if i < len(lines)-1 {
			result.WriteByte('\n')
		}
	}

	// Restore trailing newline if original had it and we are returning something
	if result.Len() > 0 && len(data) > 0 && data[len(data)-1] == '\n' {
		result.WriteByte('\n')
	}

	return result
}
