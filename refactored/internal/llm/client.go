package llm

import (
	"context"
	"fmt"
	"log"

	"article-chat-system/internal/config"

	"github.com/sashabaranov/go-openai"
)

// GenerativeModel interface for LLM operations
type GenerativeModel interface {
	GenerateContent(ctx context.Context, prompt string) (*OpenAIResponse, error)
}

// OpenAIClient wraps the OpenAI client for the refactored system
type OpenAIClient struct {
	client *openai.Client
	model  string
}

// Ensure OpenAIClient implements GenerativeModel
var _ GenerativeModel = (*OpenAIClient)(nil)

// NewOpenAIClient initializes and returns a new OpenAI client.
func NewOpenAIClient(cfg *config.Config) *OpenAIClient {
	if cfg.GeminiAPIKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable not set.")
	}

	client := openai.NewClient(cfg.GeminiAPIKey)
	model := cfg.OpenAIModel // Use model from config

	log.Printf("OpenAI client initialized successfully with model: %s.", model)
	return &OpenAIClient{
		client: client,
		model:  model,
	}
}

// NewClientFactory creates and returns a GenerativeModel based on the configuration.
func NewClientFactory(ctx context.Context, cfg *config.Config) (GenerativeModel, error) {
	switch cfg.LLMProvider {
	case "openai":
		return NewOpenAIClient(cfg), nil
	case "gemini":
		// TODO: Implement Gemini client creation if needed
		return nil, fmt.Errorf("Gemini provider not yet implemented")
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", cfg.LLMProvider)
	}
}

// GenerateContent generates content using OpenAI (compatible with Gemini interface)
func (c *OpenAIClient) GenerateContent(ctx context.Context, prompt string) (*OpenAIResponse, error) {
	resp, err := c.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: c.model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			Temperature: 0.1, // Low temperature for consistent JSON output
		},
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create chat completion: %w", err)
	}

	return &OpenAIResponse{
		Candidates: []OpenAICandidate{
			{
				Content: OpenAIContent{
					Parts: []OpenAIPart{
						{Text: resp.Choices[0].Message.Content},
					},
				},
			},
		},
	}, nil
}

// OpenAIResponse mimics the Gemini response structure
type OpenAIResponse struct {
	Candidates []OpenAICandidate `json:"candidates"`
}

type OpenAICandidate struct {
	Content OpenAIContent `json:"content"`
}

type OpenAIContent struct {
	Parts []OpenAIPart `json:"parts"`
}

type OpenAIPart struct {
	Text string `json:"text"`
}
