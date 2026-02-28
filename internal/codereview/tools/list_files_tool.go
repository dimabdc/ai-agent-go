package tools

import (
	gitea2 "code.gitea.io/sdk/gitea"
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"

	"ai-agent-go/internal/gitea"
)

type ListFilesTool struct {
	giteaClient *gitea.Client
	pr          *gitea2.PullRequest
}

func NewListFilesTool(giteaClient *gitea.Client, pr *gitea2.PullRequest) *ListFilesTool {
	return &ListFilesTool{
		giteaClient: giteaClient,
		pr:          pr,
	}
}

func (t *ListFilesTool) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name:        "list_files",
		Desc:        "Получает список всех файлов в измененной ветке из Gitea.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{}),
	}, nil
}

func (t *ListFilesTool) InvokableRun(_ context.Context, _ string, _ ...tool.Option) (string, error) {
	files, err := t.giteaClient.ListFiles(t.pr.Head.Repository.Owner.UserName, t.pr.Head.Repository.Name, t.pr.Head.Ref)
	if err != nil {
		return "", fmt.Errorf("не удалось получить список файлов: %w", err)
	}

	result := fmt.Sprintf("Список файлов в измененной ветке:\n%s", strings.Join(files, "\n"))

	return result, nil
}
