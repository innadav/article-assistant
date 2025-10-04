package strategies

import (
	"context"
	"log"

	"article-assistant/internal/article"
	"article-assistant/internal/planner"
	"article-assistant/internal/prompts"
)

type ComparePositivityStrategy struct {
	BaseStrategy
}

func NewComparePositivityStrategy() *ComparePositivityStrategy {
	s := &ComparePositivityStrategy{}
	s.doExecute = s.comparePositivity
	return s
}

func (s *ComparePositivityStrategy) comparePositivity(ctx context.Context, plan *planner.QueryPlan, articleSvc *article.Service, promptFactory *prompts.Factory) (string, error) {
	log.Println("COMPARE POSITIVITY STRATEGY: Executing...")
	if len(plan.Targets) < 2 {
		return "Please specify two articles to compare.", nil
	}
	if len(plan.Parameters) == 0 {
		return "Please specify the topic for comparison.", nil
	}
	topic := plan.Parameters[0]
	art1, ok1 := articleSvc.GetArticle(plan.Targets[0])
	art2, ok2 := articleSvc.GetArticle(plan.Targets[1])
	if !ok1 || !ok2 {
		return "Could not find one or both articles.", nil
	}
	prompt, _ := promptFactory.CreateComparePositivityPrompt(topic, art1, art2)
	return articleSvc.CallSynthesisLLM(ctx, prompt)
}
