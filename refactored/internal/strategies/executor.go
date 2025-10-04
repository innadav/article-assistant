package strategies

import (
	"context"
	"fmt"

	"article-chat-system/internal/planner"
)

// Executor holds the map of all available strategies.
type Executor struct {
	strategies map[planner.QueryIntent]planner.IntentStrategy
}

// NewExecutor creates and initializes the strategy map.
func NewExecutor() *Executor {
	return &Executor{
		strategies: map[planner.QueryIntent]planner.IntentStrategy{
			planner.IntentSummarize:          &SummarizeStrategy{},
			planner.IntentKeywords:           &KeywordsStrategy{},
			planner.IntentSentiment:          &SentimentStrategy{},
			planner.IntentCompareTone:        &CompareToneStrategy{},
			planner.IntentFindTopic:          &FindTopicStrategy{},
			planner.IntentComparePositivity:  &ComparePositivityStrategy{},
			planner.IntentFindCommonEntities: &FindCommonEntitiesStrategy{},
		},
	}
}

// ExecutePlan finds the correct strategy for the plan's intent and executes it.
func (e *Executor) ExecutePlan(ctx context.Context, plan *planner.QueryPlan, articleSvc interface{}, promptFactory interface{}) (string, error) {
	strategy, ok := e.strategies[plan.Intent]
	if !ok {
		return fmt.Sprintf("I'm sorry, I don't know how to handle the intent: %s", plan.Intent), nil
	}
	// The executor calls the strategy, providing it with all the tools it needs.
	return strategy.Execute(ctx, plan, articleSvc, promptFactory)
}
