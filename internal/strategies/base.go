package strategies

import (
	"context"
	"fmt"

	"article-assistant/internal/article"
	"article-assistant/internal/planner"
	"article-assistant/internal/prompts"
)

// IntentStrategy defines the interface that all concrete strategies must implement.
type IntentStrategy interface {
	Execute(ctx context.Context, plan *planner.QueryPlan, articleSvc *article.Service, promptFactory *prompts.Factory) (string, error)
}

// strategyStep defines the function signature for a unique action within the workflow.
type strategyStep func(ctx context.Context, plan *planner.QueryPlan, articleSvc *article.Service, promptFactory *prompts.Factory) (string, error)

// BaseStrategy defines the skeleton of our algorithm using the Template Method pattern.
type BaseStrategy struct {
	// doExecute is the customizable step that each concrete strategy will provide.
	doExecute strategyStep
}

// Execute defines the overall, unchangeable workflow for every strategy.
func (s *BaseStrategy) Execute(ctx context.Context, plan *planner.QueryPlan, articleSvc *article.Service, promptFactory *prompts.Factory) (string, error) {
	// Step 1: Common Input Processing (Shared Logic)
	if err := s.validateInput(plan); err != nil {
		return err.Error(), nil
	}

	// Step 2: Execute the Specific Action (Customizable Step)
	result, err := s.doExecute(ctx, plan, articleSvc, promptFactory)
	if err != nil {
		return "", fmt.Errorf("error during specific execution: %w", err)
	}

	// Step 3: Common Response Processing (Shared Logic)
	formattedResponse := s.formatResponse(result)
	return formattedResponse, nil
}

// validateInput is a helper method shared by all strategies.
func (s *BaseStrategy) validateInput(plan *planner.QueryPlan) error {
	if plan == nil || plan.Intent == "" {
		return fmt.Errorf("invalid plan provided")
	}
	// Add other common validation logic here, e.g., checking for targets if needed.
	return nil
}

// formatResponse is a helper method shared by all strategies.
func (s *BaseStrategy) formatResponse(response string) string {
	return fmt.Sprintf("🤖 Here is your answer:\n\n%s", response)
}
