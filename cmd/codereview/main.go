package main

import (
	"context"
	"errors"
	"fmt"
	einopenai "github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/dimabdc/ai-agent-go/internal/gitea"
	"github.com/dimabdc/ai-agent-go/internal/jira"
	"log"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/joho/godotenv"

	"github.com/dimabdc/ai-agent-go/internal/codereview"
)

// Options - конфигурация Code Review Agent через CLI и env переменные
type Options struct {
	JiraBaseURL  string `long:"jira-url" description:"Jira API base URL" env:"JIRA_BASE_URL" required:"true"`
	JiraToken    string `long:"jira-token" description:"Jira API token (required)" env:"JIRA_TOKEN" required:"true"`
	GiteaBaseURL string `long:"gitea-url" description:"Gitea API base URL" env:"GITEA_BASE_URL" required:"true"`
	GiteaToken   string `long:"gitea-token" description:"Gitea API token (required)" env:"GITEA_TOKEN" required:"true"`
	OpenAIKey    string `long:"openai-key" description:"OpenAI API key (optional for local models)" env:"OPENAI_API_KEY" required:"true"`
	OpenAIURL    string `long:"openai-url" description:"OpenAI-compatible API base URL (for LM Studio, Ollama, etc.)" env:"OPENAI_BASE_URL" default:"http://localhost:1234/v1"`
	Model        string `long:"model" description:"Chat model name" default:"qwen3.5-35b-a3b" env:"CHAT_MODEL"`
}

func main() {
	_ = godotenv.Load()

	var opts Options
	parser := flags.NewParser(&opts, flags.Default|flags.HelpFlag)

	// Парсим аргументы командной строки
	args, err := parser.Parse()
	if err != nil {
		var flagsErr *flags.Error
		if errors.As(err, &flagsErr) && errors.Is(flagsErr.Type, flags.ErrHelp) {
			return
		}
		log.Fatalf("Failed to parse arguments: %v", err)
	}

	// Поддержка позиционного аргумента для PR URL
	if len(args) < 1 {
		log.Fatal("PR URL is required")
	}

	prurl := args[0]

	fmt.Printf("🔍 Code Review Agent\n")
	fmt.Printf("   Model: %s\n", opts.Model)
	fmt.Printf("   PR: %s\n", prurl)

	start := time.Now()
	ctx := context.Background()

	giteaClient, err := gitea.NewClient(ctx, opts.GiteaBaseURL, opts.GiteaToken)
	if err != nil {
		log.Fatalf("failed to create Gitea client: %w", err)
	}

	modelConfig := &einopenai.ChatModelConfig{
		Model:   opts.Model,
		APIKey:  opts.OpenAIKey,
		BaseURL: opts.OpenAIURL,
	}

	chatModel, err := einopenai.NewChatModel(ctx, modelConfig)
	if err != nil {
		log.Fatalf("failed to create chat model: %w", err)
	}

	jiraClient, err := jira.NewClient(opts.JiraBaseURL, opts.JiraToken)
	if err != nil {
		log.Fatalf("failed to create jira client: %w", err)
	}

	// Запускаем анализ кода
	agent, err := codereview.NewCodeReviewAgent(
		giteaClient,
		chatModel,
		jiraClient,
	)
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	review, err := agent.Run(ctx, prurl)
	if err != nil {
		log.Fatalf("Agent execution failed: %v", err)
	}

	fmt.Printf("   Work time: %s\n", time.Now().Sub(start).String())
	fmt.Printf("   Agent result: %s\n", review)
}
