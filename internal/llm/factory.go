package llm

import (
	"context"
	"fmt"
	"strings"

	"article-assistant/internal/config"
)

// NewClientFactory reads the config and returns the appropriate LLM client.
func NewClientFactory(ctx context.Context, cfg *config.Config) (Client, error) {
	provider := strings.ToLower(cfg.LLMProvider)
	switch provider {
	case "google", "gemini":
		return newGeminiClient(ctx, cfg.GoogleAPIKey)
	default:
		return nil, fmt.Errorf("unknown or unsupported LLM provider: %s", cfg.LLMProvider)
	}
}
