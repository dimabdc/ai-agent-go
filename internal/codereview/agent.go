package codereview

import (
	"ai-agent-go/internal"
	"ai-agent-go/internal/jira"
	"context"
	"encoding/json"
	"fmt"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"

	codereviewtools "ai-agent-go/internal/codereview/tools"
	"ai-agent-go/internal/gitea"
)

var jiraTaskRegexp = regexp.MustCompile("#(\\w+-\\d+)")

type Agent struct {
	giteaClient *gitea.Client
	model       model.ToolCallingChatModel
	jiraClient  *jira.Client
}

func NewCodeReviewAgent(
	giteaClient *gitea.Client,
	model model.ToolCallingChatModel,
	jiraClient *jira.Client,
) (*Agent, error) {

	return &Agent{
		giteaClient: giteaClient,
		model:       model,
		jiraClient:  jiraClient,
	}, nil
}

func (e *Agent) Run(ctx context.Context, prURL string) (string, error) {
	owner, repo, prNumber, err := ParsePRURL(prURL)
	if err != nil {
		return "", err
	}

	pr, err := e.giteaClient.GetPullRequestInfo(owner, repo, prNumber)
	if err != nil {
		return "", fmt.Errorf("failed to get PR info: %w", err)
	}

	taskPrompt := ""
	taskKey := jiraTaskRegexp.FindStringSubmatch(pr.Title)
	if len(taskKey) == 2 {
		task, err := e.jiraClient.GetTask(ctx, taskKey[1])
		if err != nil {
			return "", fmt.Errorf("failed to get task %s: %w", taskKey[1], err)
		}

		taskPrompt = fmt.Sprintf("Название задачи: %s\nОписание задачи: %s\n\n", task.Summary, task.Description)
	}

	// 1. Planner (LLM агент)
	planner, err := adk.NewChatModelAgent(
		ctx, &adk.ChatModelAgentConfig{
			Name:        "planner",
			Description: "plan diff analysis",
			Model:       e.model,
			Instruction: plannerPrompt,
			ToolsConfig: adk.ToolsConfig{
				ToolsNodeConfig: compose.ToolsNodeConfig{
					Tools: []tool.BaseTool{
						codereviewtools.NewGetPRDiffTool(e.giteaClient, pr),
					},
					UnknownToolsHandler: internal.UnknownToolsHandler,
				},
			},
		},
	)

	// 2. Explorer (LLM + tools)
	explorer, err := adk.NewChatModelAgent(
		ctx, &adk.ChatModelAgentConfig{
			Name:        "explorer",
			Description: "collect context using tools",
			Model:       e.model,
			Instruction: explorerPrompt,
			ToolsConfig: adk.ToolsConfig{
				ToolsNodeConfig: compose.ToolsNodeConfig{
					Tools: []tool.BaseTool{
						codereviewtools.NewListFilesTool(e.giteaClient, pr),
						codereviewtools.NewGetFileContentTool(e.giteaClient, pr),
					},
					UnknownToolsHandler: internal.UnknownToolsHandler,
				},
			},
		},
	)

	// 3. Reviewer
	reviewer, err := adk.NewChatModelAgent(
		ctx, &adk.ChatModelAgentConfig{
			Name:        "reviewer",
			Description: "analyze diff",
			Model:       e.model,
			Instruction: reviewerPrompt,
		},
	)

	// 4. Reviewer
	loopBreak, err := adk.NewChatModelAgent(
		ctx, &adk.ChatModelAgentConfig{
			Name:        "loop_break",
			Description: "controller for loop break",
			Model:       e.model,
			Instruction: loopBreakPrompt,
			ToolsConfig: adk.ToolsConfig{
				ToolsNodeConfig: compose.ToolsNodeConfig{
					Tools: []tool.BaseTool{
						codereviewtools.NewLoopBreakTool("loop_break"),
					},
					UnknownToolsHandler: internal.UnknownToolsHandler,
				},
			},
		},
	)

	explorerReviewerLoop, err := adk.NewLoopAgent(
		ctx, &adk.LoopAgentConfig{
			Name:        "exploration_reviewer_loop",
			Description: "Review agent loop",
			SubAgents: []adk.Agent{
				explorer, reviewer, loopBreak,
			},
			MaxIterations: 6,
		},
	)

	// 5. Reflector
	reflector, err := adk.NewChatModelAgent(
		ctx, &adk.ChatModelAgentConfig{
			Name:        "reflector",
			Description: "validate report",
			Model:       e.model,
			Instruction: reflectorPrompt,
			ToolsConfig: adk.ToolsConfig{
				ToolsNodeConfig: compose.ToolsNodeConfig{
					Tools: []tool.BaseTool{
						codereviewtools.NewPostCommentTool(e.giteaClient, pr),
					},
					UnknownToolsHandler: internal.UnknownToolsHandler,
				},
			},
			MaxIterations: 2,
		},
	)

	if err != nil {
		return "", err
	}

	plannerRoute, err := adk.SetSubAgents(ctx, planner, []adk.Agent{explorerReviewerLoop})

	// 🔥 Sequential pipeline
	seq, err := adk.NewSequentialAgent(
		ctx, &adk.SequentialAgentConfig{
			Name: "code_review_pipeline",
			SubAgents: []adk.Agent{
				plannerRoute,
				reflector,
			},
		},
	)
	if err != nil {
		return "", err
	}

	runner := adk.NewRunner(
		ctx, adk.RunnerConfig{
			Agent: seq,
		},
	)

	// Создаем входные данные для adk.Agent
	agentInput := []adk.Message{
		schema.UserMessage(
			fmt.Sprintf(
				"%sПроанализируй pull request.\n"+
					"Результат анализа опубликуй комментарием к PR.",
				taskPrompt,
			),
		),
	}

	var logs []adk.Message
	defer func() {
		jsn, _ := json.MarshalIndent(logs, "", "  ")
		_ = os.WriteFile("context.json", jsn, 0644)
	}()

	logs = append(logs, agentInput...)

	// Запускаем агент через adk.Runner
	stream := runner.Run(ctx, agentInput)

	// Собираем все события
	var finalContent string
	for {
		event, ok := stream.Next()
		if !ok {
			break
		}
		if event == nil {
			break
		}

		if event.Err != nil {
			return "", event.Err
		}

		b, _ := json.MarshalIndent(event, "", "    ")
		log.Println(string(b))

		msg, _, err := adk.GetMessage(event)
		if err != nil {
			continue
		}
		if msg == nil {
			continue
		}
		logs = append(logs, msg)
		finalContent = msg.Content
	}

	return finalContent, nil
}

// ParsePRURL парсит URL pull request и извлекает owner, repo и prNumber
// Формат URL: https://gitea.example.com/owner/repo/pulls/123 или https://gitea.example.com/owner/repo/pull/123
func ParsePRURL(url string) (owner, repo string, prNumber int64, err error) {
	// Ожидаемый формат: https://gitea.example.com/owner/repo/pulls/123
	parts := strings.Split(url, "/")
	if len(parts) < 6 {
		return "", "", 0, fmt.Errorf("invalid PR URL format: %s", url)
	}

	// Находим индекс "pulls" или "pull"
	for i, part := range parts {
		if part == "pulls" || part == "pull" {
			if i+1 >= len(parts) {
				return "", "", 0, fmt.Errorf("missing PR number in URL: %s", url)
			}
			// owner и repo находятся перед "pulls"
			if i < 2 {
				return "", "", 0, fmt.Errorf("invalid PR URL format: %s", url)
			}
			owner = parts[i-2]
			repo = parts[i-1]

			// Парсим номер PR
			prNumber, err = strconv.ParseInt(parts[i+1], 10, 64)
			if err != nil {
				return "", "", 0, fmt.Errorf("invalid PR number: %s", parts[i+1])
			}
			return owner, repo, prNumber, nil
		}
	}

	return "", "", 0, fmt.Errorf("no pull request section found in URL: %s", url)
}
