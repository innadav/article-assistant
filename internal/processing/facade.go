package processing

import (
	"context"
	"fmt"
	"log"
	"time"

	"article-assistant/internal/article"
	"article-assistant/internal/llm"
	"article-assistant/internal/repository"
)

// ProcessingFacade defines the interface for the article processing facade.
type ProcessingFacade interface {
	AddNewArticle(ctx context.Context, url string) (*article.Article, error)
}

// Facade provides a simplified interface to the article processing subsystem.
type Facade struct {
	fetcher    *Fetcher
	analyzer   *Analyzer
	articleSvc *article.Service
	repo       *repository.ArticleRepository
}

// NewFacade initializes the Facade with all its required subsystem components.
func NewFacade(llmClient llm.Client, articleSvc *article.Service, repo *repository.ArticleRepository) *Facade {
	return &Facade{
		fetcher:    NewFetcher(),
		analyzer:   NewAnalyzer(llmClient),
		articleSvc: articleSvc,
		repo:       repo,
	}
}

// AddNewArticle is the single method that hides the complex processing steps.
func (f *Facade) AddNewArticle(ctx context.Context, url string) (*article.Article, error) {
	log.Printf("FACADE: Starting to process new article from URL: %s", url)

	// It first checks the DB to avoid reprocessing.
	existing, _ := f.repo.FindByURL(ctx, url)
	if existing != nil {
		return nil, fmt.Errorf("article from URL %s already exists", url)
	}

	// 1. Coordinate the Fetcher
	parsedArticle, err := f.fetcher.FetchAndParse(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("fetcher failed: %w", err)
	}

	newArticle := &article.Article{
		URL:         url,
		Title:       parsedArticle.Title,
		Content:     parsedArticle.TextContent,
		Excerpt:     parsedArticle.Excerpt,
		ProcessedAt: time.Now(),
	}

	// 2. Coordinate the Analyzer
	if err := f.analyzer.InitialAnalysis(ctx, newArticle); err != nil {
		log.Printf("WARNING: Initial analysis failed for %s: %v", url, err)
	}

	// It saves the final article to the database.
	if err := f.repo.Save(ctx, newArticle); err != nil {
		return nil, err
	}
	return newArticle, nil
}
