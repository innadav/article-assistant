package planner

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"article-assistant/internal/article"
	"article-assistant/internal/llm"
	"article-assistant/internal/prompts"
)

// PlannerService defines the interface for the planner's operations.
type PlannerService interface {
	CreatePlan(ctx context.Context, query string) (*QueryPlan, error)
}

// Service translates natural language queries into a structured QueryPlan.
type Service struct {
	llmClient     llm.Client
	promptFactory *prompts.Factory
	articleSvc    *article.Service // Needed to get article context for the prompt
}

func NewService(llmClient llm.Client, promptFactory *prompts.Factory, articleSvc *article.Service) *Service {
	return &Service{
		llmClient:     llmClient,
		promptFactory: promptFactory,
		articleSvc:    articleSvc,
	}
}

// CreatePlan makes an LLM call to generate a structured plan.
func (s *Service) CreatePlan(ctx context.Context, query string) (*QueryPlan, error) {
	// Get available articles to provide context to the planner.
	articles := s.articleSvc.GetAllArticles()
	prompt, err := s.promptFactory.CreatePlannerPrompt(query, articles)
	if err != nil {
		return nil, fmt.Errorf("failed to create planner prompt: %w", err)
	}

	resp, err := s.llmClient.GenerateContent(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("planner LLM call failed: %w", err)
	}

	var plan QueryPlan
	if err := json.Unmarshal([]byte(resp.Text), &plan); err != nil {
		log.Printf("Failed to parse JSON from planner, malformed text: %s", resp.Text)
		// Here you could add retry logic to ask the LLM to fix the JSON.
		return nil, fmt.Errorf("failed to unmarshal plan from LLM response: %w", err)
	}

	log.Printf("Successfully created plan. Intent: %s, Targets: %v", plan.Intent, plan.Targets)
	return &plan, nil
}
