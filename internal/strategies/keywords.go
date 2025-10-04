package strategies

import (
	"context"
	"log"
	"strings"

	"article-assistant/internal/article"
	"article-assistant/internal/planner"
	"article-assistant/internal/prompts"
)

type KeywordsStrategy struct {
	BaseStrategy
}

func NewKeywordsStrategy() *KeywordsStrategy {
	s := &KeywordsStrategy{}
	s.doExecute = s.extractKeywords
	return s
}

func (s *KeywordsStrategy) extractKeywords(ctx context.Context, plan *planner.QueryPlan, articleSvc *article.Service, promptFactory *prompts.Factory) (string, error) {
	log.Println("KEYWORDS STRATEGY: Performing specific keyword extraction...")
	if len(plan.Targets) == 0 {
		return "Please specify an article to extract keywords from.", nil
	}
	art, ok := articleSvc.GetArticle(plan.Targets[0])
	if !ok {
		return "I couldn't find the requested article.", nil
	}
	if len(art.Topics) == 0 {
		prompt, err := promptFactory.CreateKeywordsPrompt(art.Title)
		if err != nil {
			return "", err
		}
		keywordsStr, err := articleSvc.CallSynthesisLLM(ctx, prompt)
		if err != nil {
			return "", err
		}
		art.Topics = strings.Split(keywordsStr, ", ")
	}
	return "The main topics are: " + strings.Join(art.Topics, ", "), nil
}
