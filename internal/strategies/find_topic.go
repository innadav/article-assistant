package strategies

import (
	"context"
	"log"

	"article-assistant/internal/article"
	"article-assistant/internal/planner"
	"article-assistant/internal/prompts"
)

type FindTopicStrategy struct {
	BaseStrategy
}

func NewFindTopicStrategy() *FindTopicStrategy {
	s := &FindTopicStrategy{}
	s.doExecute = s.findTopic
	return s
}

func (s *FindTopicStrategy) findTopic(ctx context.Context, plan *planner.QueryPlan, articleSvc *article.Service, promptFactory *prompts.Factory) (string, error) {
	log.Println("FIND TOPIC STRATEGY: Executing...")
	if len(plan.Parameters) == 0 {
		return "Please specify a topic to search for.", nil
	}
	topic := plan.Parameters[0]
	allArticles := articleSvc.GetAllArticles()
	prompt, _ := promptFactory.CreateFindTopicPrompt(topic, allArticles)
	return articleSvc.CallSynthesisLLM(ctx, prompt)
}
