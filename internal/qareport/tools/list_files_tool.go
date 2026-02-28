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

type ListFilesTool struct {
	giteaClient *gitea.Client
	prs         map[int]*gitea2.PullRequest
}

func NewListFilesTool(
	giteaClient *gitea.Client,
	prs map[int]*gitea2.PullRequest,
) *ListFilesTool {
	return &ListFilesTool{
		giteaClient: giteaClient,
		prs:         prs,
	}
}

func (t *ListFilesTool) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "list_files",
		Desc: "Получает список всех файлов в измененной ветке.",
		ParamsOneOf: schema.NewParamsOneOfByParams(
			map[string]*schema.ParameterInfo{
				"ref": {
					Type:     schema.String,
					Desc:     "Идентификатор Pull Request",
					Required: true,
				},
			},
		),
	}, nil
}

func (t *ListFilesTool) InvokableRun(_ context.Context, argsJSON string, _ ...tool.Option) (string, error) {
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

	files, err := t.giteaClient.ListFiles(
		pr.Head.Repository.Owner.UserName,
		pr.Head.Repository.Name,
		pr.Head.Ref,
	)
	if err != nil {
		return "", fmt.Errorf("не удалось получить список файлов: %w", err)
	}

	result := "Список файлов в PR:\n"
	for _, file := range files {
		result += fmt.Sprintf("- %s\n", file)
	}

	return result, nil
}
