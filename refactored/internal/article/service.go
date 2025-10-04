package article

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"article-chat-system/internal/analysis"
	"article-chat-system/internal/llm"
	"article-chat-system/internal/planner"
	"article-chat-system/internal/prompts"
)

// ArticleManagement defines the methods that strategies need from the article service.
type ArticleManagement interface {
	GetArticle(url string) (*Article, bool)
	GetAllArticles() []*Article
	CallSynthesisLLM(ctx context.Context, prompt string) (string, error)
	StoreArticleForTest(url string, article *Article) // Only for testing purposes
}

// Service handles article processing and retrieval
type Service struct {
	store        sync.Map // In-memory store: URL -> *Article
	analysisSvc  *analysis.Service
	llmClient    llm.GenerativeModel
	strategyExec interface{} // Will be set to *strategies.Executor
}

func NewService(analysisSvc *analysis.Service, llmClient llm.GenerativeModel) *Service {
	return &Service{
		analysisSvc:  analysisSvc,
		llmClient:    llmClient,
		strategyExec: nil, // Will be set later to avoid import cycle
	}
}

// Ensure Service implements ArticleManagement interface
var _ ArticleManagement = (*Service)(nil)

// GetAllArticles retrieves all processed articles from the store.
func (s *Service) GetAllArticles() []*Article {
	var articles []*Article
	s.store.Range(func(key, value interface{}) bool {
		if art, ok := value.(*Article); ok {
			articles = append(articles, art)
		}
		return true // continue iteration
	})
	return articles
}

// GetAllArticlesAsPrompts retrieves all processed articles from the store and converts them to prompts.Article.
func (s *Service) GetAllArticlesAsPrompts() []*prompts.Article {
	var promptArticles []*prompts.Article
	s.store.Range(func(key, value interface{}) bool {
		if art, ok := value.(*Article); ok {
			promptArticles = append(promptArticles, &prompts.Article{
				URL:   art.URL,
				Title: art.Title,
			})
		}
		return true
	})
	return promptArticles
}

// GetArticle retrieves a single article by its URL.
func (s *Service) GetArticle(url string) (*Article, bool) {
	val, ok := s.store.Load(url)
	if !ok {
		return nil, false
	}
	art, ok := val.(*Article)
	return art, ok
}

// CallSynthesisLLM provides a centralized method for LLM synthesis calls
func (s *Service) CallSynthesisLLM(ctx context.Context, prompt string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()

	resp, err := s.llmClient.GenerateContent(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate synthesis content: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "I was unable to generate a response.", nil
	}

	return resp.Candidates[0].Content.Parts[0].Text, nil
}

// ExecuteChatPlan processes a query plan using the strategy pattern
func (s *Service) ExecuteChatPlan(ctx context.Context, plan *planner.QueryPlan, promptFactory interface{}) (string, error) {
	if s.strategyExec == nil {
		return "", fmt.Errorf("strategy executor not initialized")
	}
	// Type assertion to get the ExecutePlan method
	executor, ok := s.strategyExec.(interface {
		ExecutePlan(ctx context.Context, plan *planner.QueryPlan, articleSvc interface{}, promptFactory interface{}) (string, error)
	})
	if !ok {
		return "", fmt.Errorf("strategy executor does not implement ExecutePlan")
	}
	return executor.ExecutePlan(ctx, plan, s, promptFactory)
}

// SetStrategyExecutor sets the strategy executor (used to avoid import cycles)
func (s *Service) SetStrategyExecutor(executor interface{}) {
	s.strategyExec = executor
}

// Test helper methods
func (s *Service) StoreArticleForTest(url string, article *Article) {
	s.store.Store(url, article)
}

// ProcessArticle processes and stores an article
func (s *Service) ProcessArticle(ctx context.Context, url string, title string, content string) error {
	// Generate excerpt (first 200 characters of content)
	excerpt := content
	if len(excerpt) > 200 {
		excerpt = excerpt[:200] + "..."
	}

	article := &Article{
		URL:       url,
		Title:     title,
		Content:   content,
		Excerpt:   excerpt,
		Keywords:  []string{}, // Initialize empty
		Topics:    []string{}, // Initialize empty
		Entities:  []string{}, // Initialize empty
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Process the article (placeholder for actual processing)
	log.Printf("Processing article: %s", title)

	// Store the article
	s.store.Store(url, article)
	return nil
}

// ProcessInitialArticles processes a list of initial articles on startup
func (s *Service) ProcessInitialArticles(ctx context.Context, urls []string) {
	log.Printf("Processing %d initial articles...", len(urls))

	for _, url := range urls {
		// Placeholder for actual article fetching and processing
		s.ProcessArticle(ctx, url, "Sample Title", "Sample content...")
	}

	log.Println("Initial article processing complete")
}
