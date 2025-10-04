package planner

import (
	"context"

	"article-assistant/internal/article"
	"article-assistant/internal/prompts"
)

// IntentStrategy defines the interface for executing a specific user intent.
// Each strategy class (Summarize, Keywords, etc.) must implement this interface.
type IntentStrategy interface {
	Execute(ctx context.Context, plan *QueryPlan, articleSvc *article.Service, promptFactory *prompts.Factory) (string, error)
}
