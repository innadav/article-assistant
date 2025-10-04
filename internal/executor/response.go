package executor

import (
	"article-assistant/internal/domain"
	"article-assistant/internal/repository"
	"context"
)

// ResponseGenerator handles common response generation patterns
type ResponseGenerator struct {
	repo *repository.Repo
}

// NewResponseGenerator creates a new response generator
func NewResponseGenerator(repo *repository.Repo) *ResponseGenerator {
	return &ResponseGenerator{repo: repo}
}

// CreateTextResponse creates a text response with sources from articles
func (rg *ResponseGenerator) CreateTextResponse(ctx context.Context, answer, command string, articleURLs []string) (*domain.ChatResponse, error) {
	sources, err := rg.createSourcesFromURLs(ctx, articleURLs)
	if err != nil {
		return nil, err
	}

	return &domain.ChatResponse{
		Answer:       answer,
		Sources:      sources,
		ResponseType: domain.ResponseText,
		Task:         command,
	}, nil
}

// CreateArticleListResponse creates an article list response with sources
func (rg *ResponseGenerator) CreateArticleListResponse(ctx context.Context, answer, command string, articles []domain.Article) (*domain.ChatResponse, error) {
	sources := rg.createSourcesFromArticles(articles)

	return &domain.ChatResponse{
		Answer:       answer,
		Sources:      sources,
		ResponseType: domain.ResponseArticleList,
		Task:         command,
	}, nil
}

// CreateErrorResponse creates an error response without sources
func (rg *ResponseGenerator) CreateErrorResponse(command, message string) *domain.ChatResponse {
	return &domain.ChatResponse{
		Answer:       message,
		ResponseType: domain.ResponseText,
		Task:         command,
	}
}

// createSourcesFromURLs creates sources by fetching articles from URLs
func (rg *ResponseGenerator) createSourcesFromURLs(ctx context.Context, urls []string) ([]domain.Source, error) {
	if len(urls) == 0 {
		return []domain.Source{}, nil
	}

	articles, err := rg.repo.GetArticlesByURLs(ctx, urls)
	if err != nil {
		return nil, err
	}

	return rg.createSourcesFromArticles(articles), nil
}

// createSourcesFromArticles creates sources from article objects
func (rg *ResponseGenerator) createSourcesFromArticles(articles []domain.Article) []domain.Source {
	sources := make([]domain.Source, 0, len(articles))
	for _, article := range articles {
		sources = append(sources, domain.Source{
			ID:    article.ID,
			URL:   article.URL,
			Title: article.Title,
		})
	}
	return sources
}

// CreateSingleArticleResponse creates a response for a single article
func (rg *ResponseGenerator) CreateSingleArticleResponse(ctx context.Context, answer, command string, article *domain.Article) (*domain.ChatResponse, error) {
	var sources []domain.Source
	if article != nil {
		sources = []domain.Source{
			{
				ID:    article.ID,
				URL:   article.URL,
				Title: article.Title,
			},
		}
	}

	return &domain.ChatResponse{
		Answer:       answer,
		Sources:      sources,
		ResponseType: domain.ResponseText,
		Task:         command,
	}, nil
}
