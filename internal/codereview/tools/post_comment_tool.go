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

type PostCommentTool struct {
	giteaClient *gitea.Client
	pr          *gitea2.PullRequest
}

func NewPostCommentTool(giteaClient *gitea.Client, pr *gitea2.PullRequest) *PostCommentTool {
	return &PostCommentTool{
		giteaClient: giteaClient,
		pr:          pr,
	}
}

func (t *PostCommentTool) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "post_comment",
		Desc: "Публикует комментарий к pull request в Gitea. Используй для отправки отчета.",
		ParamsOneOf: schema.NewParamsOneOfByParams(
			map[string]*schema.ParameterInfo{
				"comment": {
					Type:     schema.String,
					Desc:     "Текст комментария для публикации",
					Required: true,
				},
			},
		),
	}, nil
}

func (t *PostCommentTool) InvokableRun(_ context.Context, argsJSON string, _ ...tool.Option) (string, error) {
	var params struct {
		Comment string `json:"comment"`
	}

	if err := json.Unmarshal([]byte(argsJSON), &params); err != nil {
		return "", fmt.Errorf("ошибка парсинга аргументов: %w", err)
	}

	if err := t.giteaClient.PostComment(
		t.pr.Head.Repository.Owner.UserName,
		t.pr.Head.Repository.Name,
		t.pr.Index,
		params.Comment,
	); err != nil {
		return "", fmt.Errorf("не удалось опубликовать комментарий: %w", err)
	}

	return "Комментарий успешно опубликован", nil
}
