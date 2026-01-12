package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/simonfrey/jsonl"
	log "github.com/sirupsen/logrus"
	"github.com/transitIOM/projectMercury/api"
	"github.com/transitIOM/projectMercury/internal/tools"
)

// PutMessage godoc
// @Summary      Append a new message to the log
// @Description  Appends a new message entry to the existing message log. Requires API key authentication.
// @Tags         messages
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Param        message  formData  string  true  "Message content"
// @Security     ApiKeyAuth
// @Success      202  {object}  api.PutMessageResponse
// @Failure      400  {object}  api.Error
// @Failure      500  {object}  api.Error
// @Router       /messages/ [put]
func PutMessage(sm tools.ObjectStorageManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Debug("Handling PutMessage request")
		message := r.FormValue("message")

		if message == "" {
			log.Debug("Message parameter missing in request")
			http.Error(w, "message parameter is required", http.StatusBadRequest)
			return
		}

		log.Debugf("Received message to store: %s", message)

		messageObj := tools.NewMessage(message)
		b := bytes.Buffer{}
		writer := jsonl.NewWriter(&b)
		err := writer.Write(messageObj)
		if err != nil {
			log.Error(err)
			api.InternalErrorHandler(w)
			return
		}

		versionID, err := sm.AppendMessage(&b)
		if err != nil {
			log.Error(err)
			api.InternalErrorHandler(w)
			return
		}

		response := api.PutMessageResponse{
			Code:      http.StatusAccepted,
			VersionID: versionID,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(response.Code)
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			log.Error(err)
			api.InternalErrorHandler(w)
		}
	}
}
