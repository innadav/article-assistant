package article

import (
	"context"
	"time"

	"article-assistant/internal/llm"
)

// Service provides an abstraction over the article data store.
type Service struct {
	store     *Store
	llmClient llm.Client
}

func NewService(llmClient llm.Client) *Service {
	return &Service{
		store:     NewStore(),
		llmClient: llmClient,
	}
}

func (s *Service) GetArticle(url string) (*Article, bool) {
	return s.store.Get(url)
}

func (s *Service) GetAllArticles() []*Article {
	return s.store.GetAll()
}

func (s *Service) StoreArticle(article *Article) {
	s.store.Set(article.URL, article)
}

func (s *Service) CallSynthesisLLM(ctx context.Context, prompt string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()

	resp, err := s.llmClient.GenerateContent(ctx, prompt)
	if err != nil {
		return "", err
	}
	return resp.Text, nil
}
