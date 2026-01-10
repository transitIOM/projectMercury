package handlers

import (
	"net/http"
	"net/mail"
	"slices"

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
// @Param        title        formData  string  true   "Report title" maxlength(300)
// @Param        description  formData  string  true   "Report description" maxlength(4000)
// @Param        email        formData  string  false  "Reporter email" maxlength(50)
// @Param        category     formData  string  false  "Report category" Enums(schedule, realtime, bug)
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

	validTags := []string{"schedule", "realtime", "bug", ""}
	if !slices.Contains(validTags, props.Category) {
		log.Debug("Invalid report category")
		http.Error(w, "invalid report category", http.StatusBadRequest)
		return
	}
	var tags []string
	if props.Category != "" {
		tags = []string{"user-report", props.Category}
	} else {
		tags = []string{"user-report"}
	}

	if props.Title == "" {
		log.Debug("Title parameter missing in request")
		http.Error(w, "title parameter is required", http.StatusBadRequest)
		return
	}
	if len(props.Title) > 300 {
		log.Debug("Title parameter exceeds maximum length of 300 characters")
		http.Error(w, "title parameter exceeds maximum length of 300 characters", http.StatusBadRequest)
		return
	}

	if props.Description == "" {
		log.Debug("Description parameter missing in request")
		http.Error(w, "description parameter is required", http.StatusBadRequest)
		return
	}
	if len(props.Description) > 4000 {
		log.Debug("Description parameter exceeds maximum length of 4000 characters")
		http.Error(w, "description parameter exceeds maximum length of 4000 characters", http.StatusBadRequest)
		return
	}

	if len(props.Email) > 50 {
		log.Debug("Email address exceeds maximum length of 50 characters")
		http.Error(w, "email address exceeds maximum length of 50 characters", http.StatusBadRequest)
		return
	}
	if props.Email != "" {
		if !isEmailValid(props.Email) {
			log.Debug("Invalid email address")
			http.Error(w, "invalid email address", http.StatusBadRequest)
			return
		}
	}

	log.Debug("Received valid report")
	err := tools.CreateIssueFromReport(ctx, props.Title, props.Description, props.Email, tags)
	if err != nil {
		log.Error(err)
		api.InternalErrorHandler(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}

func isEmailValid(email string) bool {
	emailAddress, err := mail.ParseAddress(email)
	return err == nil && emailAddress.Address == email
}
