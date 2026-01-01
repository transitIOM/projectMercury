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

	authHeader := fmt.Sprintf("Bearer %s", linearAPIKey)
	client = client.WithRequestModifier(func(req *http.Request) {
		req.Header.Set("Authorization", authHeader)
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

func CreateIssueFromReport(ctx context.Context, title, description, email string, tags []string) error {
	var mutation IssueCreateMutation
	teamID := "DEV"

	variables := map[string]interface{}{
		"input": IssueCreateInput{
			Title:       title,
			Description: description,
			TeamID:      teamID,
			Labels:      tags,
			Priority:    0, // 0=None, 1=Urgent, 2=High, 3=Normal, 4=Low
		},
	}

	err := client.Mutate(ctx, &mutation, variables)
	if err != nil {
		return fmt.Errorf("failed to create issue: %w", err)
	}

	if !mutation.IssueCreate.Success {
		return fmt.Errorf("issue creation failed")
	}

	log.Debug("Created issue: %s (%s)",
		mutation.IssueCreate.Issue.Identifier,
		mutation.IssueCreate.Issue.URL)

	return nil
}
