package strategies

import (
	"context"
	"fmt"

	"article-chat-system/internal/article"
	"article-chat-system/internal/planner"
	"article-chat-system/internal/prompts"
)

type CompareToneStrategy struct{}

func (s *CompareToneStrategy) Execute(ctx context.Context, plan *planner.QueryPlan, articleSvc interface{}, promptFactory interface{}) (string, error) {
	if len(plan.Targets) < 2 {
		return "Please specify two articles to compare.", nil
	}

	svc, ok := articleSvc.(*article.Service)
	if !ok {
		return "", fmt.Errorf("invalid article service type")
	}

	pf, ok := promptFactory.(*prompts.Factory)
	if !ok {
		return "", fmt.Errorf("invalid prompt factory type")
	}

	article1, ok1 := svc.GetArticle(plan.Targets[0])
	article2, ok2 := svc.GetArticle(plan.Targets[1])
	if !ok1 || !ok2 {
		return "I couldn't find one or both of the articles to compare.", nil
	}

	prompt := pf.GenerateCompareTonePrompt(article1.Title, article2.Title)
	return svc.CallSynthesisLLM(ctx, prompt)
}
