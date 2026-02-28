package jira

// Task represents a Jira issue with its key, summary, description and custom fields.
type Task struct {
	Key         string   `json:"key"`
	Summary     string   `json:"summary"`
	Description string   `json:"description"`
	PRLinks     []string `json:"links,omitempty"`
}
