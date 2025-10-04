package strategies

import (
	"context"
	"fmt"
	"strings"

	"article-chat-system/internal/article"
	"article-chat-system/internal/planner"
	"article-chat-system/internal/prompts"
)

type FindTopicStrategy struct{}

func (s *FindTopicStrategy) Execute(ctx context.Context, plan *planner.QueryPlan, articleSvc interface{}, promptFactory interface{}) (string, error) {
	if len(plan.Parameters) == 0 {
		return "What topic are you looking for?", nil
	}

	svc, ok := articleSvc.(*article.Service)
	if !ok {
		return "", fmt.Errorf("invalid article service type")
	}

	pf, ok := promptFactory.(*prompts.Factory)
	if !ok {
		return "", fmt.Errorf("invalid prompt factory type")
	}

	topic := plan.Parameters[0]
	var contextBuilder strings.Builder
	fmt.Fprintf(&contextBuilder, "Here are the available articles:\n")
	for _, article := range svc.GetAllArticles() {
		fmt.Fprintf(&contextBuilder, "- URL: %s, Title: %s, Excerpt: %s\n", article.URL, article.Title, article.Excerpt)
	}

	prompt := pf.GenerateFindTopicPrompt(topic, contextBuilder.String())

	return svc.CallSynthesisLLM(ctx, prompt)
}
