package processing

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"article-chat-system/internal/article"
	"article-chat-system/internal/llm"

	"github.com/go-shiori/go-readability"
)

// Facade provides a simplified interface for complex article processing operations.
type Facade struct {
	llmClient  llm.GenerativeModel
	articleSvc *article.Service
}

// NewFacade creates a new processing facade.
func NewFacade(llmClient llm.GenerativeModel, articleSvc *article.Service) *Facade {
	return &Facade{
		llmClient:  llmClient,
		articleSvc: articleSvc,
	}
}

// AddNewArticle fetches, processes, and stores a new article.
func (f *Facade) AddNewArticle(ctx context.Context, articleURL string) (*article.Article, error) {
	// 1. Fetch article content
	resp, err := http.Get(articleURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch article from URL %s: %w", articleURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch article: received status code %d from %s", resp.StatusCode, articleURL)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// 2. Parse article using go-readability
	parsedArticle, err := readability.Parse(strings.NewReader(string(body)), articleURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse article content: %w", err)
	}

	// 3. Check if article already exists
	if _, found := f.articleSvc.GetArticle(articleURL); found {
		return nil, fmt.Errorf("article with URL %s already exists", articleURL)
	}

	// 4. Process and store the article
	newArticle := &article.Article{
		URL:       articleURL,
		Title:     parsedArticle.Title,
		Content:   parsedArticle.TextContent,
		Excerpt:   parsedArticle.Excerpt,
		Keywords:  []string{}, // Placeholder for actual extraction
		Topics:    []string{}, // Placeholder for actual extraction
		Entities:  []string{}, // Placeholder for actual extraction
	}

	err = f.articleSvc.ProcessArticle(ctx, newArticle.URL, newArticle.Title, newArticle.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to process and store article: %w", err)
	}

	log.Printf("Successfully added article: %s (URL: %s)", newArticle.Title, newArticle.URL)
	return newArticle, nil
}
