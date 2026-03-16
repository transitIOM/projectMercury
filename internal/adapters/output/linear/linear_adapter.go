package linear

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/hasura/go-graphql-client"
	"github.com/transitIOM/projectMercury/internal/domain/models"
)

type Adapter struct {
	client *graphql.Client
	teamID string
	tagMap map[string]string
}

func NewAdapter() (*Adapter, error) {
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("LINEAR_API_KEY environment variable not set")
	}

	client := graphql.NewClient("https://api.linear.app/graphql", nil)
	client = client.WithRequestModifier(func(req *http.Request) {
		req.Header.Set("Authorization", apiKey)
		req.Header.Set("Content-Type", "application/json")
	})

	return &Adapter{
		client: client,
		teamID: "b3a913ce-ea8c-4d0f-8c9a-15630b40ea80",
		tagMap: map[string]string{
			"bug":         "9536247d-eaca-4319-8fb0-5667037a828e",
			"user-report": "167f2203-ab68-4ff0-8a82-a12293d49b01",
			"realtime":    "85a1d688-58de-4208-87ee-b6ac4907fdc2",
			"schedule":    "a0e52de7-19a5-4c27-9d4a-898afdb1645f",
		},
	}, nil
}

type IssueCreateMutation struct {
	IssueCreate struct {
		Success bool
		Issue   struct {
			ID         string `graphql:"id"`
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

func (a *Adapter) CreateIssueFromReport(ctx context.Context, report models.UserReport) error {
	var mutation IssueCreateMutation

	description := fmt.Sprintf("Submitted by: %s\n\n%s", report.Email, report.Description)

	var tagIds []string
	if id, ok := a.tagMap[report.Category]; ok {
		tagIds = append(tagIds, id)
	}
	// Always add user-report tag
	if id, ok := a.tagMap["user-report"]; ok && report.Category != "user-report" {
		tagIds = append(tagIds, id)
	}

	variables := map[string]interface{}{
		"input": IssueCreateInput{
			Title:       report.Title,
			Description: description,
			TeamID:      a.teamID,
			Labels:      tagIds,
			Priority:    0,
		},
	}

	err := a.client.Mutate(ctx, &mutation, variables)
	if err != nil {
		slog.ErrorContext(ctx, "failed to create linear issue", "error", err)
		return fmt.Errorf("failed to create linear issue: %w", err)
	}

	if !mutation.IssueCreate.Success {
		slog.ErrorContext(ctx, "linear issue creation failed", "success", mutation.IssueCreate.Success)
		return fmt.Errorf("linear issue creation failed")
	}

	slog.InfoContext(ctx, "Successfully created Linear issue", "id", mutation.IssueCreate.Issue.ID, "identifier", mutation.IssueCreate.Issue.Identifier)
	return nil
}
