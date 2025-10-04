package strategies

import (
	"context"
	"fmt"
	"strings"

	"article-chat-system/internal/article"
	"article-chat-system/internal/planner"
	"article-chat-system/internal/prompts"
)

type FindCommonEntitiesStrategy struct{}

func (s *FindCommonEntitiesStrategy) Execute(ctx context.Context, plan *planner.QueryPlan, articleSvc interface{}, promptFactory interface{}) (string, error) {
	svc, ok := articleSvc.(*article.Service)
	if !ok {
		return "", fmt.Errorf("invalid article service type")
	}

	pf, ok := promptFactory.(*prompts.Factory)
	if !ok {
		return "", fmt.Errorf("invalid prompt factory type")
	}

	var contextBuilder strings.Builder
	fmt.Fprintf(&contextBuilder, "Here are the titles of all available articles:\n")
	for _, article := range svc.GetAllArticles() {
		fmt.Fprintf(&contextBuilder, "- %s\n", article.Title)
	}

	prompt := pf.GenerateFindCommonEntitiesPrompt(contextBuilder.String())

	return svc.CallSynthesisLLM(ctx, prompt)
}
