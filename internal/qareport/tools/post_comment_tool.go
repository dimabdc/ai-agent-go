package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	bfconfluence "github.com/kentaro-m/blackfriday-confluence"
	bf "github.com/russross/blackfriday/v2"

	"ai-agent-go/internal/jira"
)

// PostCommentTool - инструмент для публикации комментария
type PostCommentTool struct {
	client *jira.Client
	task   string
}

// NewPostCommentTool создает новый инструмент PostCommentTool
func NewPostCommentTool(
	client *jira.Client,
	task string,
) *PostCommentTool {
	return &PostCommentTool{
		client: client,
		task:   task,
	}
}

// Info возвращает информацию об инструменте
func (t *PostCommentTool) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "post_comment",
		Desc: `Опубликовать комментарий к задачи. Используй для отправки отчета.`,
		ParamsOneOf: schema.NewParamsOneOfByParams(
			map[string]*schema.ParameterInfo{
				"comment": {
					Type:     schema.String,
					Desc:     "Текст комментария",
					Required: true,
				},
			},
		),
	}, nil
}

// InvokableRun выполняет инструмент
func (t *PostCommentTool) InvokableRun(ctx context.Context, argsJSON string, opts ...tool.Option) (string, error) {
	var params struct {
		Comment string `json:"comment"`
	}

	if err := json.Unmarshal([]byte(argsJSON), &params); err != nil {
		return "", fmt.Errorf("failed to parse input: %w", err)
	}

	renderer := &bfconfluence.Renderer{}
	md := bf.New(bf.WithRenderer(renderer), bf.WithExtensions(bf.CommonExtensions))
	output := renderer.Render(md.Parse([]byte(params.Comment)))

	if err := t.client.PostComment(ctx, t.task, string(output)); err != nil {
		return "", fmt.Errorf("failed to post comment: %w", err)
	}

	return "Comment posted successfully", nil
}
