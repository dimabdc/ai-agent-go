package main

import (
	"context"
	"errors"
	"fmt"
	einopenai "github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/dimabdc/ai-agent-go/internal/gitea"
	"github.com/dimabdc/ai-agent-go/internal/jira"
	"github.com/jessevdk/go-flags"
	"github.com/joho/godotenv"
	"log"
	"time"

	"github.com/dimabdc/ai-agent-go/internal/qareport"
)

// Options - конфигурация QA Agent через CLI и env переменные
type Options struct {
	JiraBaseURL  string `long:"jira-url" description:"Jira API base URL"env:"JIRA_BASE_URL" required:"true"`
	JiraToken    string `long:"jira-token" description:"Jira API token (required)" env:"JIRA_TOKEN" required:"true"`
	GiteaBaseURL string `long:"gitea-url" description:"Gitea API base URL" env:"GITEA_BASE_URL" required:"true"`
	GiteaToken   string `long:"gitea-token" description:"Gitea API token (required)" env:"GITEA_TOKEN" required:"true"`
	OpenAIKey    string `long:"openai-key" description:"OpenAI API key (optional for local models)" env:"OPENAI_API_KEY" required:"true"`
	OpenAIURL    string `long:"openai-url" description:"OpenAI-compatible API base URL (for LM Studio, Ollama, etc.)" env:"OPENAI_BASE_URL" default:"http://localhost:1234/v1"`
	Model        string `long:"model" description:"Chat model name" default:"qwen3.5-35b-a3b" env:"CHAT_MODEL"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	var opts Options
	parser := flags.NewParser(&opts, flags.Default|flags.HelpFlag)

	args, err := parser.Parse()
	if err != nil {
		var flagsErr *flags.Error
		if errors.As(err, &flagsErr) && errors.Is(flagsErr.Type, flags.ErrHelp) {
			return
		}
		log.Fatalf("Failed to parse arguments: %v", err)
	}

	// Поддержка позиционного аргумента для task key
	if args[0] == "" {
		log.Fatal("Task key is required")
	}

	taskKey := args[0]

	fmt.Printf("🔍 QA Report Agent\n")
	fmt.Printf("   Model: %s\n", opts.Model)
	fmt.Printf("   Task: %s\n", taskKey)

	start := time.Now()
	ctx := context.Background()

	modelConfig := &einopenai.ChatModelConfig{
		Model:   opts.Model,
		APIKey:  opts.OpenAIKey,
		BaseURL: opts.OpenAIURL,
	}

	chatModel, err := einopenai.NewChatModel(ctx, modelConfig)
	if err != nil {
		log.Fatalf("failed to create chat model: %s", err)
	}

	jiraClient, err := jira.NewClient(opts.JiraBaseURL, opts.JiraToken)
	if err != nil {
		log.Fatalf("failed to create jira client: %s", err)
	}

	giteaClient, err := gitea.NewClient(ctx, opts.GiteaBaseURL, opts.GiteaToken)
	if err != nil {
		log.Fatalf("failed to create gitea client: %s", err)
	}

	agent, err := qareport.NewQAReportAgent(
		chatModel,
		jiraClient,
		giteaClient,
	)
	if err != nil {
		log.Fatalf("Error creating agent: %s\n", err)
	}

	report, err := agent.GenerateReport(ctx, taskKey)
	if err != nil {
		log.Fatalf("Error generating report: %s\n", err)
	}

	fmt.Printf("   Work time: %s\n", time.Now().Sub(start).String())
	fmt.Printf("   Agent result: %s\n", report)
}
