package tools

import (
	gitea2 "code.gitea.io/sdk/gitea"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"

	"ai-agent-go/internal/gitea"
)

type ListFilesTool struct {
	giteaClient *gitea.Client
	prs         map[string]*gitea2.PullRequest
}

func NewListFilesTool(
	giteaClient *gitea.Client,
	prs map[string]*gitea2.PullRequest,
) *ListFilesTool {
	return &ListFilesTool{
		giteaClient: giteaClient,
		prs:         prs,
	}
}

func (t *ListFilesTool) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "list_files",
		Desc: "Gets a list of all files in the modified branch.",
		ParamsOneOf: schema.NewParamsOneOfByParams(
			map[string]*schema.ParameterInfo{
				"pr_code": {
					Type:     schema.String,
					Desc:     "Pull request code",
					Required: true,
				},
			},
		),
	}, nil
}

func (t *ListFilesTool) InvokableRun(_ context.Context, argsJSON string, _ ...tool.Option) (string, error) {
	var params struct {
		Code string `json:"pr_code"`
	}

	if err := json.Unmarshal([]byte(argsJSON), &params); err != nil {
		return "", fmt.Errorf("failed to parse input: %w", err)
	}

	pr, ok := t.prs[strings.TrimSpace(params.Code)]
	if !ok {
		return fmt.Sprintf("PR `%s` not found", params.Code), nil
	}

	files, err := t.giteaClient.ListFiles(
		pr.Head.Repository.Owner.UserName,
		pr.Head.Repository.Name,
		pr.Head.Ref,
	)
	if err != nil {
		return "", fmt.Errorf("failed to get list files: %w", err)
	}

	result := "List files in PR:\n"
	for _, file := range files {
		result += fmt.Sprintf("- %s\n", file)
	}

	return result, nil
}
