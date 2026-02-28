package jira

import (
	"context"
	"fmt"
	"strings"

	"github.com/andygrunwald/go-jira"
)

// Client represents a Jira client using the official go-jira library.
type Client struct {
	client *jira.Client
}

// NewClient creates a new Jira client with the given token and base URL.
func NewClient(baseURL, apiToken string) (*Client, error) {
	if strings.TrimSpace(baseURL) == "" {
		return nil, fmt.Errorf("baseURL cannot be empty")
	}
	if strings.TrimSpace(apiToken) == "" {
		return nil, fmt.Errorf("apiToken cannot be empty")
	}

	bearerTransport := &jira.BearerAuthTransport{
		Token: apiToken,
	}

	jiraClient, err := jira.NewClient(bearerTransport.Client(), baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create jira client: %w", err)
	}

	return &Client{client: jiraClient}, nil
}

// GetTask retrieves a Jira issue by its key (e.g., "PROJ-123").
func (c *Client) GetTask(ctx context.Context, taskKey string) (*Task, error) {
	if strings.TrimSpace(taskKey) == "" {
		return nil, fmt.Errorf("taskKey cannot be empty")
	}

	issue, _, err := c.client.Issue.GetWithContext(ctx, taskKey, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue %s: %w", taskKey, err)
	}

	task := &Task{
		Key:         issue.Key,
		Summary:     issue.Fields.Summary,
		Description: issue.Fields.Description,
		PRLinks:     c.extractPRLinksFromField(issue),
	}

	return task, nil
}

// PostComment adds a comment to the specified Jira issue.
func (c *Client) PostComment(ctx context.Context, taskKey string, body string) error {
	if strings.TrimSpace(taskKey) == "" {
		return fmt.Errorf("taskKey cannot be empty")
	}
	if strings.TrimSpace(body) == "" {
		return fmt.Errorf("body cannot be empty")
	}

	comment := &jira.Comment{
		Body: body,
	}

	_, _, err := c.client.Issue.AddCommentWithContext(ctx, taskKey, comment)
	if err != nil {
		return fmt.Errorf("failed to create comment for issue %s: %w", taskKey, err)
	}

	return nil
}

// extractPRLinksFromField extracts PR links from customfield_20702.
func (c *Client) extractPRLinksFromField(issue *jira.Issue) []string {
	prLinks := make([]string, 0)

	// Extract from customfield_20702
	if val, ok := issue.Fields.Unknowns["customfield_20702"]; ok {
		if strVal, ok := val.(string); ok {
			prLinks = strings.Split(strVal, "\n")
		}
	}

	return prLinks
}
