package executor

import (
	"context"
	"fmt"
	"article-assistant/internal/article"
	"article-assistant/internal/planner"
	"article-assistant/internal/prompts"
	"article-assistant/internal/strategies"
)

// StrategyExecutor defines the interface for executing query plans.
type StrategyExecutor interface {
	ExecutePlan(ctx context.Context, plan *planner.QueryPlan, articleSvc *article.Service, promptFactory *prompts.Factory) (string, error)
}

// Executor's only responsibility is to hold the map of strategies.
// It no longer holds instances of other services.
type Registry struct {
	strategies map[planner.QueryIntent]strategies.IntentStrategy
}

// NewRegistry is now much simpler. It just creates and populates the strategy map.
func NewRegistry() *Registry {
	return &Registry{
		strategies: map[planner.QueryIntent]strategies.IntentStrategy{
			planner.IntentSummarize: strategies.NewSummarizeStrategy(),
			planner.IntentKeywords:  strategies.NewKeywordsStrategy(),
			// planner.IntentSentiment:          strategies.NewSentimentStrategy(),
			// planner.IntentCompareTone:        strategies.NewCompareToneStrategy(),
			// planner.IntentFindTopic:          strategies.NewFindTopicStrategy(),
			// planner.IntentComparePositive:    NewComparePositivityStrategy(),
			// planner.IntentFindCommonEntities: NewFindCommonEntitiesStrategy(),
		},
	}
}

// ExecutePlan now accepts the dependencies it needs to pass down to the strategies.
// This makes the Executor's role purely about routing.
func (e *Registry) ExecutePlan(
	ctx context.Context,
	plan *planner.QueryPlan,
	articleSvc *article.Service,
	promptFactory *prompts.Factory,
) (string, error) {
	strategy, ok := e.strategies[plan.Intent]
	if !ok {
		return fmt.Sprintf("I'm sorry, I don't know how to handle the intent: %s", plan.Intent), nil
	}
	// It passes the dependencies on to the chosen strategy.
	return strategy.Execute(ctx, plan, articleSvc, promptFactory)
}
