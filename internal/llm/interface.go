package llm

import (
	"article-assistant/internal/domain"
	"context"
)

type Client interface {
	Summarize(ctx context.Context, text string) (string, error)
	SentimentScore(ctx context.Context, text string) (float64, error)
	ToneCompare(ctx context.Context, text1, text2 string) (string, error)
	Embed(ctx context.Context, text string) ([]float32, error)
	GenerateText(ctx context.Context, prompt string) (string, error)
	PlanQuery(ctx context.Context, query string) (*domain.Plan, error)

	// Combined semantic analysis (faster - single API call)
	ExtractAllSemantics(ctx context.Context, text string) (*domain.SemanticAnalysis, error)
}

// Ensure both implementations satisfy the interface
var _ Client = (*OpenAIClient)(nil)
var _ Client = (*MockClient)(nil)
