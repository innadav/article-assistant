package strategies

import (
	"context"
	"log"

	"article-assistant/internal/article"
	"article-assistant/internal/planner"
	"article-assistant/internal/prompts"
)

type FindCommonEntitiesStrategy struct {
	BaseStrategy
}

func NewFindCommonEntitiesStrategy() *FindCommonEntitiesStrategy {
	s := &FindCommonEntitiesStrategy{}
	s.doExecute = s.findCommonEntities
	return s
}

func (s *FindCommonEntitiesStrategy) findCommonEntities(ctx context.Context, plan *planner.QueryPlan, articleSvc *article.Service, promptFactory *prompts.Factory) (string, error) {
	log.Println("FIND COMMON ENTITIES STRATEGY: Executing...")
	allArticles := articleSvc.GetAllArticles()
	prompt, _ := promptFactory.CreateFindCommonEntitiesPrompt(allArticles)
	return articleSvc.CallSynthesisLLM(ctx, prompt)
}
