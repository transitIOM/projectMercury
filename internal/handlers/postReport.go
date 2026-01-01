package handlers

import (
	"net/http"

	log "github.com/sirupsen/logrus"
	"github.com/transitIOM/projectMercury/api"
	"github.com/transitIOM/projectMercury/internal/tools"
)

// PostReport godoc
// @Summary      Submit a new report
// @Description  Submits a new report with title, description, email, and category.
// @Tags         report
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Param        title        formData  string  true   "Report title"
// @Param        description  formData  string  true   "Report description"
// @Param        email        formData  string  false  "Reporter email"
// @Param        category     formData  string  false  "Report category"
// @Success      201
// @Failure      400  {object}  api.Error
// @Failure      500  {object}  api.Error
// @Router       /report/ [post]
func PostReport(w http.ResponseWriter, r *http.Request) {
	log.Debug("Handling PostReport request")
	ctx := r.Context()
	props := api.PostReportBody{
		Title:       r.FormValue("title"),
		Description: r.FormValue("description"),
		Email:       r.FormValue("email"),
		Category:    r.FormValue("category"),
	}

	if props.Category == "" {
		log.Debug("Category parameter missing in request, defaulting to 'unknown'")
		props.Category = "unknown"
	}
	tags := []string{"user-report", props.Category}

	if props.Title == "" {
		log.Debug("Title parameter missing in request")
		http.Error(w, "title parameter is required", http.StatusBadRequest)
		return
	}
	if props.Description == "" {
		log.Debug("Description parameter missing in request")
		http.Error(w, "description parameter is required", http.StatusBadRequest)
		return
	}

	log.Debug("Received report")
	err := tools.CreateIssueFromReport(ctx, props.Title, props.Description, props.Description, tags)
	if err != nil {
		log.Error(err)
		api.InternalErrorHandler(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}
