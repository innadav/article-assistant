package strategies

import (
	"context"
	"log"

	"article-assistant/internal/article"
	"article-assistant/internal/planner"
	"article-assistant/internal/prompts"
)

// SummarizeStrategy embeds the BaseStrategy to inherit the common workflow.
type SummarizeStrategy struct {
	BaseStrategy
}

// NewSummarizeStrategy creates a new strategy and wires up its unique logic.
func NewSummarizeStrategy() *SummarizeStrategy {
	s := &SummarizeStrategy{}
	// Assign the unique part of the algorithm to the placeholder in the BaseStrategy.
	s.doExecute = s.summarizeArticle
	return s
}

// summarizeArticle is the specific action for this strategy.
func (s *SummarizeStrategy) summarizeArticle(ctx context.Context, plan *planner.QueryPlan, articleSvc *article.Service, promptFactory *prompts.Factory) (string, error) {
	log.Println("SUMMARIZE STRATEGY: Performing specific summarization logic...")
	if len(plan.Targets) == 0 {
		return "Please specify which article you want to summarize.", nil
	}
	art, ok := articleSvc.GetArticle(plan.Targets[0])
	if !ok {
		return "I couldn't find the requested article.", nil
	}
	if art.Summary == "" {
		prompt, err := promptFactory.CreateSummarizePrompt(art.Content)
		if err != nil {
			return "", err
		}
		summary, err := articleSvc.CallSynthesisLLM(ctx, prompt)
		if err != nil {
			return "", err
		}
		art.Summary = summary
	}
	return art.Summary, nil
}
