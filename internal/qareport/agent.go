package qareport

import (
	gitea2 "code.gitea.io/sdk/gitea"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cloudwego/eino/adk"
	"github.com/dimabdc/ai-agent-go/internal"
	"os"
	"strconv"
	"strings"

	"github.com/dimabdc/ai-agent-go/internal/qareport/tools"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"

	"github.com/dimabdc/ai-agent-go/internal/gitea"
	"github.com/dimabdc/ai-agent-go/internal/jira"
)

// Agent handles QA report generation.
type Agent struct {
	jiraClient  *jira.Client
	giteaClient *gitea.Client
	model       model.ToolCallingChatModel
}

// NewQAReportAgent creates a new QA Report Agent.
func NewQAReportAgent(
	chatModel model.ToolCallingChatModel,
	jiraClient *jira.Client,
	giteaClient *gitea.Client,
) (*Agent, error) {
	return &Agent{
		jiraClient:  jiraClient,
		giteaClient: giteaClient,
		model:       chatModel,
	}, nil
}

// GenerateReport creates a QA test report from a Jira task.
func (a *Agent) GenerateReport(ctx context.Context, taskKey string) (string, error) {
	task, err := a.jiraClient.GetTask(ctx, taskKey)
	if err != nil {
		return "", fmt.Errorf("failed to get task %s: %w", taskKey, err)
	}

	prs := make(map[string]*gitea2.PullRequest, len(task.PRLinks))
	var changes []string
	for num, prLink := range task.PRLinks {
		pr, err := a.getPRInfo(prLink)
		if err != nil {
			continue
		}
		rpCode := fmt.Sprintf("pr-%d", num)
		prs[rpCode] = pr

		changeText := fmt.Sprintf("**PR %d:**\n", num+1)
		changeText += fmt.Sprintf("Code (приватный, для запросов в инструментах): %s\n", rpCode)
		changeText += fmt.Sprintf("Link: %s\n", prLink)
		changes = append(changes, changeText)
	}

	if len(changes) == 0 {
		return "", errors.New("no changes for QA analyze")
	}

	qaAgent, err := adk.NewChatModelAgent(
		ctx, &adk.ChatModelAgentConfig{
			Name:        "qa-report-agent",
			Description: "Анализ задачи и составление отчета",
			Instruction: qaReportPrompt,
			Model:       a.model,
			ToolsConfig: adk.ToolsConfig{
				ToolsNodeConfig: compose.ToolsNodeConfig{
					Tools: []tool.BaseTool{
						tools.NewGetPRDiffTool(a.giteaClient, prs),
						tools.NewGetFileContentTool(a.giteaClient, prs),
						tools.NewListFilesTool(a.giteaClient, prs),
					},
					UnknownToolsHandler: internal.UnknownToolsHandler,
				},
			},
			MaxIterations: 50,
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to create adk agent: %w", err)
	}

	reflectorAgent, err := adk.NewChatModelAgent(
		ctx, &adk.ChatModelAgentConfig{
			Name:        "reflector",
			Description: "validate report",
			Instruction: reflectorPrompt,
			Model:       a.model,
			ToolsConfig: adk.ToolsConfig{
				ToolsNodeConfig: compose.ToolsNodeConfig{
					Tools: []tool.BaseTool{
						tools.NewPostCommentTool(a.jiraClient, taskKey),
					},
					UnknownToolsHandler: internal.UnknownToolsHandler,
				},
			},
			MaxIterations: 2,
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to create adk agent: %w", err)
	}

	// Sequential pipeline
	seq, err := adk.NewSequentialAgent(
		ctx, &adk.SequentialAgentConfig{
			Name: "code_review_pipeline",
			SubAgents: []adk.Agent{
				qaAgent,
				reflectorAgent,
			},
		},
	)
	if err != nil {
		return "", err
	}

	// Создаем входные данные для adk.Agent
	agentInput := &adk.AgentInput{
		Messages: []adk.Message{
			schema.UserMessage(
				fmt.Sprintf(
					"Название задачи: %s\nОписание задачи: %s\n\n"+
						"Проанализируй pull requests:\n\n%s\n\n",
					task.Summary, task.Description, strings.Join(changes, "\n"),
				),
			),
		},
		EnableStreaming: false,
	}

	var logs []adk.Message
	defer func() {
		jsn, _ := json.MarshalIndent(logs, "", "  ")
		_ = os.WriteFile("context.json", jsn, 0644)
	}()

	logs = append(logs, agentInput.Messages...)

	// Запускаем агент через adk.Runner
	stream := seq.Run(ctx, agentInput)

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

// extractPRChanges extracts file list and diff from a PR link.
func (a *Agent) getPRInfo(prLink string) (*gitea2.PullRequest, error) {
	owner, repo, prNumber, err := parsePRURL(prLink)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PR URL: %w", err)
	}

	pr, err := a.giteaClient.GetPullRequestInfo(owner, repo, int64(prNumber))
	if err != nil {
		return nil, fmt.Errorf("failed to load pull request: %w", err)
	}

	return pr, nil
}

// parsePRURL извлекает owner, repo и prNumber из URL
func parsePRURL(url string) (owner, repo string, prNumber int, err error) {
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
			prNumber, err = strconv.Atoi(parts[i+1])
			if err != nil {
				return "", "", 0, fmt.Errorf("invalid PR number: %s", parts[i+1])
			}
			return owner, repo, prNumber, nil
		}
	}

	return "", "", 0, fmt.Errorf("no pull request section found in URL: %s", url)
}
