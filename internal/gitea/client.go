package gitea

import (
	"code.gitea.io/sdk/gitea"
	"context"
	"fmt"
)

// Client - клиент для работы с Gitea API
type Client struct {
	gitea *gitea.Client
}

// NewClient создает новый клиент Gitea
func NewClient(ctx context.Context, baseURL, token string) (*Client, error) {
	client, err := gitea.NewClient(baseURL, gitea.SetContext(ctx), gitea.SetToken(token))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gitea client: %w", err)
	}

	return &Client{gitea: client}, nil
}

// GetPRDiff получает дифф pull request
func (c *Client) GetPRDiff(owner, repo string, index int64) (string, error) {
	diff, _, err := c.gitea.GetPullRequestDiff(owner, repo, index, gitea.PullRequestDiffOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get PR diff: %w", err)
	}

	return string(diff), nil
}

// GetPullRequestInfo получает ветку PR
func (c *Client) GetPullRequestInfo(owner, repo string, index int64) (*gitea.PullRequest, error) {
	pr, _, err := c.gitea.GetPullRequest(owner, repo, index)
	if err != nil {
		return nil, fmt.Errorf("failed to get PR info: %w", err)
	}

	return pr, nil
}

// GetFileContent получает содержимое файла в PR
func (c *Client) GetFileContent(owner, repo, filePath string, ref string) (string, error) {
	content, _, err := c.gitea.GetFile(owner, repo, ref, filePath)
	if err != nil {
		return "", fmt.Errorf("failed to get file content: %w", err)
	}

	return string(content), nil
}

// ListFiles получает список файлов в PR
func (c *Client) ListFiles(owner, repo string, ref string) ([]string, error) {
	tree, _, err := c.gitea.GetTrees(
		owner, repo, gitea.ListTreeOptions{
			Ref:       ref,
			Recursive: true,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get PR files: %w", err)
	}

	filePaths := make([]string, 0, len(tree.Entries))
	for _, entre := range tree.Entries {
		if entre.Type != "blob" {
			continue
		}
		filePaths = append(filePaths, entre.Path)
	}

	return filePaths, nil
}

// PostComment публикует комментарий к PR
func (c *Client) PostComment(owner, repo string, index int64, body string) error {
	_, _, err := c.gitea.CreatePullReview(
		owner, repo, index, gitea.CreatePullReviewOptions{
			Body:  body,
			State: gitea.ReviewStateComment,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to post comment: %w", err)
	}

	return nil
}

// ListPRFiles получает список файлов в PR
func (c *Client) ListPRFiles(owner, repo string, index int64) ([]*gitea.ChangedFile, error) {
	files, _, err := c.gitea.ListPullRequestFiles(owner, repo, index, gitea.ListPullRequestFilesOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get PR files: %w", err)
	}

	return files, nil
}
