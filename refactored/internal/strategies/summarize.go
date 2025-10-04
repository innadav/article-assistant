package strategies

import (
	"context"
	"fmt"
	"log"

	"article-chat-system/internal/article"
	"article-chat-system/internal/planner"
	"article-chat-system/internal/prompts"
)

type SummarizeStrategy struct{}

func (s *SummarizeStrategy) Execute(ctx context.Context, plan *planner.QueryPlan, articleSvc interface{}, promptFactory interface{}) (string, error) {
	if len(plan.Targets) == 0 {
		return "Please specify which article you want to summarize.", nil
	}

	svc, ok := articleSvc.(*article.Service)
	if !ok {
		return "", fmt.Errorf("invalid article service type")
	}

	pf, ok := promptFactory.(*prompts.Factory)
	if !ok {
		return "", fmt.Errorf("invalid prompt factory type")
	}

	article, ok := svc.GetArticle(plan.Targets[0])
	if !ok {
		return "I couldn't find the requested article.", nil
	}

	if article.Summary == "" {
		log.Printf("Cache miss for summary on '%s'. Generating now...", article.Title)
		prompt := pf.GenerateSummaryPrompt(article.Content)
		summary, err := svc.CallSynthesisLLM(ctx, prompt)
		if err != nil {
			return "", err
		}
		article.Summary = summary
	}

	return article.Summary, nil
}
