package tools

import (
	gitea2 "code.gitea.io/sdk/gitea"
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"

	"github.com/dimabdc/ai-agent-go/internal/gitea"
)

type GetFileContentTool struct {
	giteaClient *gitea.Client
	pr          *gitea2.PullRequest
}

func NewGetFileContentTool(giteaClient *gitea.Client, pr *gitea2.PullRequest) *GetFileContentTool {
	return &GetFileContentTool{
		giteaClient: giteaClient,
		pr:          pr,
	}
}

func (t *GetFileContentTool) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "get_file_content",
		Desc: "Получает содержимое файла из pull request в Gitea. Используй для анализа кода конкретного файла.",
		ParamsOneOf: schema.NewParamsOneOfByParams(
			map[string]*schema.ParameterInfo{
				"file_path": {
					Type:     schema.String,
					Desc:     "Путь к файлу относительно корня репозитория",
					Required: true,
				},
				"reason": {
					Type:     schema.String,
					Desc:     "Причина загрузки файла и список символов, которые нужны из этого файла",
					Required: true,
				},
			},
		),
	}, nil
}

func (t *GetFileContentTool) InvokableRun(_ context.Context, argsJSON string, _ ...tool.Option) (string, error) {
	var params struct {
		FilePath string `json:"file_path"`
	}

	if err := json.Unmarshal([]byte(argsJSON), &params); err != nil {
		return "", fmt.Errorf("ошибка парсинга аргументов: %w", err)
	}

	content, err := t.giteaClient.GetFileContent(
		t.pr.Head.Repository.Owner.UserName,
		t.pr.Head.Repository.Name,
		params.FilePath,
		t.pr.Head.Ref,
	)
	if err != nil {
		return "", fmt.Errorf("не удалось получить содержимое файла: %w", err)
	}

	return content, nil
}
