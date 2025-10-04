package strategies

import (
	"context"
	"log"
	"strings"

	"article-assistant/internal/article"
	"article-assistant/internal/planner"
	"article-assistant/internal/prompts"
)

type SentimentStrategy struct {
	BaseStrategy
}

func NewSentimentStrategy() *SentimentStrategy {
	s := &SentimentStrategy{}
	s.doExecute = s.getSentiment
	return s
}

func (s *SentimentStrategy) getSentiment(ctx context.Context, plan *planner.QueryPlan, articleSvc *article.Service, promptFactory *prompts.Factory) (string, error) {
	log.Println("SENTIMENT STRATEGY: Executing...")
	if len(plan.Targets) == 0 {
		return "Please specify an article to analyze.", nil
	}
	art, ok := articleSvc.GetArticle(plan.Targets[0])
	if !ok {
		return "Could not find the requested article.", nil
	}
	if art.Sentiment == "" {
		prompt, _ := promptFactory.CreateSentimentPrompt(art.Title)
		sentiment, err := articleSvc.CallSynthesisLLM(ctx, prompt)
		if err != nil {
			return "", err
		}
		art.Sentiment = strings.TrimSpace(sentiment)
	}
	return "The sentiment is: " + art.Sentiment, nil
}
