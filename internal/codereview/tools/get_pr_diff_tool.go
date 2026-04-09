package tools

import (
	gitea2 "code.gitea.io/sdk/gitea"
	"context"
	"fmt"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"

	"github.com/dimabdc/ai-agent-go/internal/gitea"
)

const (
	diffLenLimit = 120_000
)

type GetPRDiffTool struct {
	giteaClient *gitea.Client
	pr          *gitea2.PullRequest
}

func NewGetPRDiffTool(giteaClient *gitea.Client, pr *gitea2.PullRequest) *GetPRDiffTool {
	return &GetPRDiffTool{
		giteaClient: giteaClient,
		pr:          pr,
	}
}

func (t *GetPRDiffTool) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name:        "get_pr_diff",
		Desc:        "Получает дифф pull request из Gitea. Используй для анализа изменений в коде.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{}),
	}, nil
}

func (t *GetPRDiffTool) InvokableRun(ctx context.Context, _ string, _ ...tool.Option) (string, error) {
	diff, err := t.giteaClient.GetPRDiff(t.pr.Head.Repository.Owner.UserName, t.pr.Head.Repository.Name, t.pr.Index)
	if err != nil {
		return "", fmt.Errorf("не удалось получить дифф PR: %w", err)
	}

	if len(diff) > diffLenLimit {
		return "diff to large", nil
	}

	return diff, nil
}
