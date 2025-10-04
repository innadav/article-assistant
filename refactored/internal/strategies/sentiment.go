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

type SentimentStrategy struct{}

func (s *SentimentStrategy) Execute(ctx context.Context, plan *planner.QueryPlan, articleSvc interface{}, promptFactory interface{}) (string, error) {
	if len(plan.Targets) == 0 {
		return "Please specify an article to analyze for sentiment.", nil
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

	if article.Sentiment == "" {
		log.Printf("Cache miss for sentiment on '%s'. Generating now...", article.Title)
		prompt := pf.GenerateSentimentPrompt(article.Title)
		sentiment, err := svc.CallSynthesisLLM(ctx, prompt)
		if err != nil {
			return "", err
		}
		article.Sentiment = strings.TrimSpace(sentiment)
	}

	return fmt.Sprintf("The sentiment of the article '%s' is generally **%s**.", article.Title, article.Sentiment), nil
}
