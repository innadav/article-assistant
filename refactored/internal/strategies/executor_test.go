package strategies

import (
	"context"
	"errors"
	"testing"

	"article-chat-system/internal/article"
	"article-chat-system/internal/planner"
	"article-chat-system/internal/prompts"
)

// mockStrategy is a mock implementation of the IntentStrategy interface.
// It allows us to control its behavior for testing purposes.
type mockStrategy struct {
	ExecuteFunc func(ctx context.Context, plan *planner.QueryPlan, articleSvc *article.Service, promptFactory *prompts.Factory) (string, error)
}

func (m *mockStrategy) Execute(ctx context.Context, plan *planner.QueryPlan, articleSvc *article.Service, promptFactory *prompts.Factory) (string, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, plan, articleSvc, promptFactory)
	}
	return "", errors.New("ExecuteFunc not implemented")
}

func TestExecutor_ExecutePlan(t *testing.T) {
	// ARRANGE
	// Create a mock strategy that we can control.
	mockSummarizeStrategy := &mockStrategy{}
	wasCalled := false // Use a flag to track if our mock was called.

	mockSummarizeStrategy.ExecuteFunc = func(ctx context.Context, plan *planner.QueryPlan, articleSvc *article.Service, promptFactory *prompts.Factory) (string, error) {
		wasCalled = true // Set the flag to true when executed.
		return "Mocked summary response", nil
	}

	// Create an executor and manually register our mock strategy.
	executor := &Executor{
		strategies: map[planner.QueryIntent]planner.IntentStrategy{
			planner.IntentSummarize: mockSummarizeStrategy,
		},
	}

	plan := &planner.QueryPlan{
		Intent: planner.IntentSummarize,
	}

	// We can use nil for the dependencies because our mock won't use them.
	// This proves our test is isolated.
	var nilArticleSvc *article.Service
	var nilPromptFactory *prompts.Factory

	// ACT
	response, err := executor.ExecutePlan(context.Background(), plan, nilArticleSvc, nilPromptFactory)

	// ASSERT
	if err != nil {
		t.Fatalf("ExecutePlan() returned an unexpected error: %v", err)
	}
	if !wasCalled {
		t.Error("Expected the SummarizeStrategy's Execute method to be called, but it was not.")
	}
	if response != "Mocked summary response" {
		t.Errorf("Expected response 'Mocked summary response', but got '%s'", response)
	}
}
