package tools

import (
	"context"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

type LoopBreakTool struct {
	agentName string
}

func NewLoopBreakTool(agentName string) *LoopBreakTool {
	return &LoopBreakTool{
		agentName: agentName,
	}
}

func (t *LoopBreakTool) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name:        "loop_break",
		Desc:        "Прерывает выполнение цикла.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{}),
	}, nil
}

func (t *LoopBreakTool) InvokableRun(ctx context.Context, _ string, _ ...tool.Option) (string, error) {
	_ = adk.SendToolGenAction(ctx, "break", adk.NewBreakLoopAction(t.agentName))

	return "Success", nil
}
