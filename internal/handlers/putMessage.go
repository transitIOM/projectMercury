package handlers

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
	"github.com/transitIOM/projectMercury/api"
	"github.com/transitIOM/projectMercury/internal/tools"
)

// PutMessage godoc
// @Summary      Upload a new message
// @Description  Uploads a new message to the message log.
// @Tags         messages
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Param        message  formData  string  true  "Message content"
// @Security     ApiKeyAuth
// @Success      202  {object}  api.PutMessageResponse
// @Failure      400  {object}  api.Error
// @Failure      500  {object}  api.Error
// @Router       /messages/ [put]
func PutMessage(w http.ResponseWriter, r *http.Request) {
	message := r.FormValue("message")

	if message == "" {
		http.Error(w, "message parameter is required", http.StatusBadRequest)
		return
	}

	err := tools.PushMessageToStorage(message)
	if err != nil {
		log.Error(err)
		api.InternalErrorHandler(w)
		return
	}

	versionID, err := tools.GetLatestMessageLogVersionID()
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
