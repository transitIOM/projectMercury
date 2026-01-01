package tools

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/hasura/go-graphql-client"
)

var client *graphql.Client

func initialiseLinearGraphqlConnection() {
	linearAPIKey := os.Getenv("LINEAR_API_KEY")

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
	Title       string `json:"title"`
	Description string `json:"description"`
	TeamID      string `json:"teamId"`
	Priority    int    `json:"priority,omitempty"`
}

func createIssueFromReport(ctx context.Context, title, description, teamID string) error {
	var mutation IssueCreateMutation

	variables := map[string]interface{}{
		"input": IssueCreateInput{
			Title:       title,
			Description: description,
			TeamID:      teamID,
			Priority:    2, // 0=None, 1=Urgent, 2=High, 3=Normal, 4=Low
		},
	}

	err := client.Mutate(ctx, &mutation, variables)
	if err != nil {
		return fmt.Errorf("failed to create issue: %w", err)
	}

	if !mutation.IssueCreate.Success {
		return fmt.Errorf("issue creation failed")
	}

	fmt.Printf("Created issue: %s (%s)\n",
		mutation.IssueCreate.Issue.Identifier,
		mutation.IssueCreate.Issue.URL)

	return nil
}
