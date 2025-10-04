package strategies

import (
	"context"
	"fmt"
	"log"
	"strings"

	"article-chat-system/internal/article"
	"article-chat-system/internal/planner"
	"article-chat-system/internal/prompts"
)

type KeywordsStrategy struct{}

func (s *KeywordsStrategy) Execute(ctx context.Context, plan *planner.QueryPlan, articleSvc interface{}, promptFactory interface{}) (string, error) {
	if len(plan.Targets) == 0 {
		return "Please specify an article to extract keywords from.", nil
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

	if len(article.Topics) == 0 {
		log.Printf("Cache miss for keywords on '%s'. Generating now...", article.Title)
		prompt := pf.GenerateKeywordsPrompt(article.Title)
		keywordsStr, err := svc.CallSynthesisLLM(ctx, prompt)
		if err != nil {
			return "", err
		}
		article.Topics = strings.Split(keywordsStr, ", ")
	}

	return "The main topics are: " + strings.Join(article.Topics, ", "), nil
}
