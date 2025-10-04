package strategies

import (
	"context"
	"log"

	"article-assistant/internal/article"
	"article-assistant/internal/planner"
	"article-assistant/internal/prompts"
)

type CompareToneStrategy struct {
	BaseStrategy
}

func NewCompareToneStrategy() *CompareToneStrategy {
	s := &CompareToneStrategy{}
	s.doExecute = s.compareTone
	return s
}

func (s *CompareToneStrategy) compareTone(ctx context.Context, plan *planner.QueryPlan, articleSvc *article.Service, promptFactory *prompts.Factory) (string, error) {
	log.Println("COMPARE TONE STRATEGY: Executing...")
	if len(plan.Targets) < 2 {
		return "Please specify two articles to compare.", nil
	}
	art1, ok1 := articleSvc.GetArticle(plan.Targets[0])
	art2, ok2 := articleSvc.GetArticle(plan.Targets[1])
	if !ok1 || !ok2 {
		return "Could not find one or both articles.", nil
	}
	prompt, _ := promptFactory.CreateCompareTonePrompt(art1, art2)
	return articleSvc.CallSynthesisLLM(ctx, prompt)
}
