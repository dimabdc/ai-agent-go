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

// GetFileContentTool - инструмент для получения содержимого файла
type GetFileContentTool struct {
	client *gitea.Client
	prs    map[int]*gitea2.PullRequest
}

// NewGetFileContentTool создает новый инструмент GetFileContentTool
func NewGetFileContentTool(
	client *gitea.Client,
	prs map[int]*gitea2.PullRequest,
) *GetFileContentTool {
	return &GetFileContentTool{
		client: client,
		prs:    prs,
	}
}

// Info возвращает информацию об инструменте
func (t *GetFileContentTool) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "get_file_content",
		Desc: `Get the full content of a specific file in the repository.
Use this to read file contents for code review context.`,
		ParamsOneOf: schema.NewParamsOneOfByParams(
			map[string]*schema.ParameterInfo{
				"ref": {
					Type:     schema.String,
					Desc:     "Идентификатор Pull Request",
					Required: true,
				},
				"file_path": {
					Type:     schema.String,
					Desc:     "Путь к файлу",
					Required: true,
				},
			},
		),
	}, nil
}

// InvokableRun выполняет инструмент
func (t *GetFileContentTool) InvokableRun(_ context.Context, argsJSON string, _ ...tool.Option) (string, error) {
	var params struct {
		Ref      int    `json:"ref"`
		FilePath string `json:"file_path"`
	}

	if err := json.Unmarshal([]byte(argsJSON), &params); err != nil {
		return "", fmt.Errorf("failed to parse input: %w", err)
	}

	pr, ok := t.prs[params.Ref]
	if !ok {
		return fmt.Sprintf("PR `%d` not found", params.Ref), nil
	}

	content, err := t.client.GetFileContent(
		pr.Head.Repository.Owner.UserName,
		pr.Head.Repository.Name,
		params.FilePath,
		pr.Head.Ref,
	)
	if err != nil {
		return fmt.Sprintf("Error: %s. Do not call tool with file_path: %s", err, params.FilePath), nil
	}

	return content, nil
}
