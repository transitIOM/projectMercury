package tools

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/hasura/go-graphql-client"
	log "github.com/sirupsen/logrus"
)

var client *graphql.Client

func InitialiseLinearGraphqlConnection() {
	linearAPIKey := os.Getenv("LINEAR_API_KEY")
	if linearAPIKey == "" {
		log.Fatal("LINEAR_API_KEY environment variable not set")
	}

	client = graphql.NewClient("https://api.linear.app/graphql", nil)
	client = client.WithRequestModifier(func(req *http.Request) {
		req.Header.Set("Authorization", linearAPIKey)
		req.Header.Set("Content-Type", "application/json")
	})
}

type IssueCreateMutation struct {
	IssueCreate struct {
		Success bool
		Issue   struct {
			ID         string `graphql:"id"`
			Number     int    `graphql:"number"`
			Title      string `graphql:"title"`
			Identifier string `graphql:"identifier"`
			URL        string `graphql:"url"`
		} `graphql:"issue"`
	} `graphql:"issueCreate(input: $input)"`
}

type IssueCreateInput struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	TeamID      string   `json:"teamId"`
	Labels      []string `json:"labelIds,omitempty"`
	Priority    int      `json:"priority,omitempty"`
}

var tagUUIDMap = map[string]string{
	"bug":         "9536247d-eaca-4319-8fb0-5667037a828e",
	"user-report": "167f2203-ab68-4ff0-8a82-a12293d49b01",
	"realtime":    "85a1d688-58de-4208-87ee-b6ac4907fdc2",
	"schedule":    "a0e52de7-19a5-4c27-9d4a-898afdb1645f",
}

func CreateIssueFromReport(ctx context.Context, title, description, email string, tags []string) error {
	var mutation IssueCreateMutation
	teamID := "b3a913ce-ea8c-4d0f-8c9a-15630b40ea80"

	description = fmt.Sprintf("Submitted by: %s\n\n%s", email, description)

	var tagIds []string
	for _, tag := range tags {
		if id, ok := tagUUIDMap[tag]; ok {
			tagIds = append(tagIds, id)
		} else {
			log.Warnf("Unknown tag: %s", tag)
		}
	}

	variables := map[string]interface{}{
		"input": IssueCreateInput{
			Title:       title,
			Description: description,
			TeamID:      teamID,
			Labels:      tagIds,
			Priority:    0, // 0=None, 1=Urgent, 2=High, 3=Normal, 4=Low
		},
	}

	err := client.Mutate(ctx, &mutation, variables)
	if err != nil {
		log.WithError(err).Error("Linear mutation failed")
		return fmt.Errorf("failed to create issue: %w", err)
	}

	log.WithFields(log.Fields{
		"mutation": mutation,
	}).Debug("Linear mutation response")

	if !mutation.IssueCreate.Success {
		return fmt.Errorf("issue creation failed")
	}

	log.Infof("Created issue: %s (%s)",
		mutation.IssueCreate.Issue.Identifier,
		mutation.IssueCreate.Issue.URL)

	return nil
}

type LinearReportManager struct{}

func (f *LinearReportManager) CreateIssueFromReport(ctx context.Context, title string, description string, email string, tags []string) error {
	return CreateIssueFromReport(ctx, title, description, email, tags)
}
