// internal/article/service_test.go
package unit

import (
	"context"
	"testing"

	"article-chat-system/internal/analysis"
	"article-chat-system/internal/article"
	"article-chat-system/internal/llm"
	"article-chat-system/internal/planner"
	"article-chat-system/internal/prompts"
	"article-chat-system/internal/strategies"
)

// mockLLMClient can be reused for testing the article service as well.
type mockLLMClient struct {
	MockResponse      string
	MockError         error
	GenerateCallCount int // Track how many times the LLM was called.
}

func (m *mockLLMClient) GenerateContent(ctx context.Context, prompt string) (*llm.OpenAIResponse, error) {
	m.GenerateCallCount++
	if m.MockError != nil {
		return nil, m.MockError
	}
	return &llm.OpenAIResponse{
		Candidates: []llm.OpenAICandidate{
			{
				Content: llm.OpenAIContent{
					Parts: []llm.OpenAIPart{
						{Text: m.MockResponse},
					},
				},
			},
		},
	}, nil
}

func TestArticleService_ExecuteChatPlan_Summarize(t *testing.T) {
	// ARRANGE
	mockLLM := &mockLLMClient{
		MockResponse: "This is a mock summary.",
	}
	// The analysis service is not used in this test, so we can pass a nil or empty instance.
	analysisSvc := analysis.NewService()

	articleSvc := article.NewService(analysisSvc, mockLLM)
	// Set the strategy executor to avoid import cycles
	strategyExec := strategies.NewExecutor()
	articleSvc.SetStrategyExecutor(strategyExec)
	// Create a dummy prompt loader and factory for tests
	mockPromptLoader, err := prompts.NewLoader("v1")
	if err != nil {
		t.Fatalf("Failed to create mock prompt loader: %v", err)
	}
	promptFactory := prompts.NewFactory(prompts.ModelGPT4Turbo, mockPromptLoader)

	// Pre-populate the store with an article that has no summary yet.
	sampleArticle := &article.Article{
		URL:     "https://example.com/test",
		Title:   "Test Article",
		Content: "Some content to summarize.",
		Summary: "", // Empty summary, this will cause a cache miss.
	}
	articleSvc.StoreArticleForTest(sampleArticle.URL, sampleArticle)

	summarizePlan := &planner.QueryPlan{
		Intent:  planner.IntentSummarize,
		Targets: []string{"https://example.com/test"},
	}

	// --- TEST CASE 1: Cache Miss ---
	t.Run("Cache Miss - Generates and caches summary", func(t *testing.T) {
		// ACT
		response, err := articleSvc.ExecuteChatPlan(context.Background(), summarizePlan, promptFactory)

		// ASSERT
		if err != nil {
			t.Fatalf("expected no error, but got %v", err)
		}
		if response != mockLLM.MockResponse {
			t.Errorf("expected response '%s', but got '%s'", mockLLM.MockResponse, response)
		}
		if mockLLM.GenerateCallCount != 1 {
			t.Errorf("expected LLM to be called 1 time, but was called %d times", mockLLM.GenerateCallCount)
		}

		// Verify that the summary was cached in the article object.
		cachedArticle, ok := articleSvc.GetArticle(sampleArticle.URL)
		if !ok || cachedArticle.Summary != mockLLM.MockResponse {
			t.Error("summary was not correctly cached in the article store")
		}
	})

	// --- TEST CASE 2: Cache Hit ---
	t.Run("Cache Hit - Returns cached summary without calling LLM", func(t *testing.T) {
		// ARRANGE (The summary is now cached from the previous test case)
		// Reset the call count to ensure we are testing this case in isolation.
		mockLLM.GenerateCallCount = 0

		// ACT
		response, err := articleSvc.ExecuteChatPlan(context.Background(), summarizePlan, promptFactory)

		// ASSERT
		if err != nil {
			t.Fatalf("expected no error, but got %v", err)
		}
		if response != mockLLM.MockResponse {
			t.Errorf("expected response '%s', but got '%s'", mockLLM.MockResponse, response)
		}
		// The key assertion: the LLM should NOT have been called again.
		if mockLLM.GenerateCallCount != 0 {
			t.Errorf("expected LLM to be called 0 times, but was called %d times", mockLLM.GenerateCallCount)
		}
	})
}

func TestArticleService_ExecuteChatPlan_Keywords(t *testing.T) {
	// ARRANGE
	mockLLM := &mockLLMClient{
		MockResponse: "technology, innovation, AI",
	}
	analysisSvc := analysis.NewService()
	articleSvc := article.NewService(analysisSvc, mockLLM)
	// Set the strategy executor to avoid import cycles
	strategyExec := strategies.NewExecutor()
	articleSvc.SetStrategyExecutor(strategyExec)
	// Create a dummy prompt loader and factory for tests
	mockPromptLoader, err := prompts.NewLoader("v1")
	if err != nil {
		t.Fatalf("Failed to create mock prompt loader: %v", err)
	}
	promptFactory := prompts.NewFactory(prompts.ModelGPT4Turbo, mockPromptLoader)

	// Pre-populate the store with an article
	sampleArticle := &article.Article{
		URL:      "https://example.com/tech",
		Title:    "Tech Article",
		Content:  "This article discusses technology and innovation.",
		Keywords: []string{}, // Empty keywords
	}
	articleSvc.StoreArticleForTest(sampleArticle.URL, sampleArticle)

	keywordsPlan := &planner.QueryPlan{
		Intent:  planner.IntentKeywords,
		Targets: []string{"https://example.com/tech"},
	}

	// ACT
	response, err := articleSvc.ExecuteChatPlan(context.Background(), keywordsPlan, promptFactory)

	// ASSERT
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
	if response == "" {
		t.Error("expected non-empty response for keywords")
	}
}

func TestArticleService_ExecuteChatPlan_UnknownIntent(t *testing.T) {
	// ARRANGE
	mockLLM := &mockLLMClient{}
	analysisSvc := analysis.NewService()
	articleSvc := article.NewService(analysisSvc, mockLLM)
	// Set the strategy executor to avoid import cycles
	strategyExec := strategies.NewExecutor()
	articleSvc.SetStrategyExecutor(strategyExec)
	// Create a dummy prompt loader and factory for tests
	mockPromptLoader, err := prompts.NewLoader("v1")
	if err != nil {
		t.Fatalf("Failed to create mock prompt loader: %v", err)
	}
	promptFactory := prompts.NewFactory(prompts.ModelGPT4Turbo, mockPromptLoader)

	unknownPlan := &planner.QueryPlan{
		Intent:  planner.IntentUnknown,
		Targets: []string{},
	}

	// ACT
	response, err := articleSvc.ExecuteChatPlan(context.Background(), unknownPlan, promptFactory)

	// ASSERT
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
	expectedResponse := "I'm sorry, I don't know how to handle the intent: UNKNOWN"
	if response != expectedResponse {
		t.Errorf("expected response '%s', but got '%s'", expectedResponse, response)
	}
}

func TestArticleService_GetAllArticles(t *testing.T) {
	// ARRANGE
	mockLLM := &mockLLMClient{}
	analysisSvc := analysis.NewService()
	articleSvc := article.NewService(analysisSvc, mockLLM)
	// Set the strategy executor to avoid import cycles
	strategyExec := strategies.NewExecutor()
	articleSvc.SetStrategyExecutor(strategyExec)
	// Create a dummy prompt loader and factory for tests
	mockPromptLoader, err := prompts.NewLoader("v1")
	if err != nil {
		t.Fatalf("Failed to create mock prompt loader: %v", err)
	}
	_ = prompts.NewFactory(prompts.ModelGPT4Turbo, mockPromptLoader) // We still need to create the factory for other tests, but we don't need the variable in this specific test.

	// Add some test articles
	article1 := &article.Article{URL: "https://example.com/1", Title: "Article 1"}
	article2 := &article.Article{URL: "https://example.com/2", Title: "Article 2"}

	articleSvc.StoreArticleForTest(article1.URL, article1)
	articleSvc.StoreArticleForTest(article2.URL, article2)

	// ACT
	articles := articleSvc.GetAllArticles()

	// ASSERT
	if len(articles) != 2 {
		t.Errorf("expected 2 articles, but got %d", len(articles))
	}
}

func TestArticleService_GetArticle(t *testing.T) {
	// ARRANGE
	mockLLM := &mockLLMClient{}
	analysisSvc := analysis.NewService()
	articleSvc := article.NewService(analysisSvc, mockLLM)
	// Set the strategy executor to avoid import cycles
	strategyExec := strategies.NewExecutor()
	articleSvc.SetStrategyExecutor(strategyExec)
	// Create a dummy prompt loader and factory for tests
	mockPromptLoader, err := prompts.NewLoader("v1")
	if err != nil {
		t.Fatalf("Failed to create mock prompt loader: %v", err)
	}
	_ = prompts.NewFactory(prompts.ModelGPT4Turbo, mockPromptLoader) // We still need to create the factory for other tests, but we don't need the variable in this specific test.

	testArticle := &article.Article{URL: "https://example.com/test", Title: "Test Article"}
	articleSvc.StoreArticleForTest(testArticle.URL, testArticle)

	// ACT
	article, found := articleSvc.GetArticle("https://example.com/test")

	// ASSERT
	if !found {
		t.Error("expected article to be found")
	}
	if article.Title != "Test Article" {
		t.Errorf("expected title 'Test Article', but got '%s'", article.Title)
	}

	// Test non-existent article
	_, found = articleSvc.GetArticle("https://example.com/nonexistent")
	if found {
		t.Error("expected article not to be found")
	}
}
