package tools

import (
	gitea2 "code.gitea.io/sdk/gitea"
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"

	"ai-agent-go/internal/gitea"
)

const (
	diffLenLimit = 120_000
)

// GetPRDiffTool - инструмент для получения диффа pull request
type GetPRDiffTool struct {
	client *gitea.Client
	prs    map[int]*gitea2.PullRequest
}

// NewGetPRDiffTool создает новый инструмент GetPRDiffTool
func NewGetPRDiffTool(
	client *gitea.Client,
	prs map[int]*gitea2.PullRequest,
) *GetPRDiffTool {
	return &GetPRDiffTool{
		client: client,
		prs:    prs,
	}
}

// Info возвращает информацию об инструменте
func (t *GetPRDiffTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "get_pr_diff",
		Desc: "Get the unified diff content of a pull request from Gitea.",
		ParamsOneOf: schema.NewParamsOneOfByParams(
			map[string]*schema.ParameterInfo{
				"ref": {
					Type:     schema.Integer,
					Desc:     "Идентификатор Pull Request",
					Required: true,
				},
			},
		),
	}, nil
}

// InvokableRun выполняет инструмент
func (t *GetPRDiffTool) InvokableRun(_ context.Context, argsJSON string, _ ...tool.Option) (string, error) {
	var params struct {
		Ref int `json:"ref"`
	}

	if err := json.Unmarshal([]byte(argsJSON), &params); err != nil {
		return "", fmt.Errorf("failed to parse input: %w", err)
	}

	pr, ok := t.prs[params.Ref]
	if !ok {
		return fmt.Sprintf("PR `%d` not found", params.Ref), nil
	}

	diff, err := t.client.GetPRDiff(pr.Head.Repository.Owner.UserName, pr.Head.Repository.Name, pr.Index)
	if err != nil {
		return "", fmt.Errorf("failed to get PR diff: %w", err)
	}

	if len(diff) > diffLenLimit {
		return "diff to large", nil
	}

	return diff, nil
}
