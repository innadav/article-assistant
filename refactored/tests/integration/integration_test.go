// tests/integration/full_flow_test.go
package integration

import (
	"context"
	"errors"
	"strings"
	"testing"

	"article-chat-system/internal/analysis"
	"article-chat-system/internal/article"
	"article-chat-system/internal/llm"
	"article-chat-system/internal/planner"
	"article-chat-system/internal/prompts"
)

// Mock LLM client for integration testing
type mockLLMClient struct {
	MockResponses     []string
	MockError         error
	GenerateCallCount int
}

func (m *mockLLMClient) GenerateContent(ctx context.Context, prompt string) (*llm.OpenAIResponse, error) {
	m.GenerateCallCount++
	if m.MockError != nil {
		return nil, m.MockError
	}
	// Return different responses based on call count
	var response string
	if len(m.MockResponses) > 0 {
		if m.GenerateCallCount <= len(m.MockResponses) {
			response = m.MockResponses[m.GenerateCallCount-1]
		} else {
			response = m.MockResponses[len(m.MockResponses)-1]
		}
	}

	return &llm.OpenAIResponse{
		Candidates: []llm.OpenAICandidate{
			{
				Content: llm.OpenAIContent{
					Parts: []llm.OpenAIPart{
						{Text: response},
					},
				},
			},
		},
	}, nil
}

func TestFullFlow_IngestPlanExecute(t *testing.T) {
	// ARRANGE
	mockLLM := &mockLLMClient{
		MockResponses: []string{
			`{
				"intent": "SUMMARIZE",
				"targets": ["https://techcrunch.com/2025/07/25/sam-altman-warns-theres-no-legal-confidentiality-when-using-chatgpt-as-a-therapist/"],
				"parameters": [],
				"question": "summarize the sam altman article"
			}`,
			"This is a comprehensive summary of Sam Altman's discussion about AI confidentiality and legal issues.",
		},
	}

	analysisSvc := analysis.NewService()
	articleSvc := article.NewService(analysisSvc, mockLLM)

	// Create a dummy prompt loader and factory for the planner.
	mockPromptLoader, err := prompts.NewLoader("v1")
	if err != nil {
		t.Fatalf("Failed to create mock prompt loader: %v", err)
	}
	promptFactory := prompts.NewFactory(prompts.ModelGPT4Turbo, mockPromptLoader)
	plannerSvc := planner.NewService(mockLLM, promptFactory)

	testURL := "https://techcrunch.com/2025/07/25/sam-altman-warns-theres-no-legal-confidentiality-when-using-chatgpt-as-a-therapist/"
	testTitle := "Sam Altman warns there's no legal confidentiality when using ChatGPT as a therapist"
	testContent := "Sam Altman, OpenAI's CEO, discussed the challenges of AI and legal confidentiality on Theo Von's podcast 'This Past Weekend.' Altman highlighted that unlike traditional therapist-client relationships, conversations with ChatGPT currently lack legal confidentiality, posing risks of exposure during legal proceedings."

	// ACT 1: Ingest Article
	t.Run("Step 1: Ingest Article", func(t *testing.T) {
		err := articleSvc.ProcessArticle(context.Background(), testURL, testTitle, testContent)
		if err != nil {
			t.Fatalf("Failed to ingest article: %v", err)
		}

		// Verify article was stored
		article, found := articleSvc.GetArticle(testURL)
		if !found {
			t.Fatal("Article was not found after ingestion")
		}
		if article.Title != testTitle {
			t.Errorf("Expected title '%s', got '%s'", testTitle, article.Title)
		}
	})

	// ACT 2: Create Plan
	t.Run("Step 2: Create Execution Plan", func(t *testing.T) {
		availableArticles := articleSvc.GetAllArticles()
		var promptArticles []*prompts.Article
		for _, art := range availableArticles {
			promptArticles = append(promptArticles, &prompts.Article{
				URL:   art.URL,
				Title: art.Title,
			})
		}

		query := "summarize the sam altman article"
		plan, err := plannerSvc.CreatePlan(context.Background(), query, promptArticles)
		if err != nil {
			t.Fatalf("Failed to create plan: %v", err)
		}

		// Verify plan structure
		if plan.Intent != planner.IntentSummarize {
			t.Errorf("Expected intent SUMMARIZE, got %s", plan.Intent)
		}
		if len(plan.Targets) != 1 {
			t.Errorf("Expected 1 target, got %d", len(plan.Targets))
		}
		if plan.Targets[0] != testURL {
			t.Errorf("Expected target '%s', got '%s'", testURL, plan.Targets[0])
		}
	})

	// ACT 3: Execute Plan
	t.Run("Step 3: Execute Plan and Generate Summary", func(t *testing.T) {
		// The mockLLM already has the responses in MockResponses, so we don't need to reset it.
		// mockLLM.MockResponse = "This is a comprehensive summary of Sam Altman's discussion about AI confidentiality and legal issues."
		// mockLLM.GenerateCallCount = 0

		availableArticles := articleSvc.GetAllArticles()
		var promptArticles []*prompts.Article
		for _, art := range availableArticles {
			promptArticles = append(promptArticles, &prompts.Article{
				URL:   art.URL,
				Title: art.Title,
			})
		}

		query := "summarize the sam altman article"
		plan, err := plannerSvc.CreatePlan(context.Background(), query, promptArticles)
		if err != nil {
			t.Fatalf("Failed to create plan: %v", err)
		}

		summary, err := articleSvc.ExecuteChatPlan(context.Background(), plan, promptFactory)
		if err != nil {
			t.Fatalf("Failed to execute plan: %v", err)
		}

		// Verify summary was generated
		if summary == "" {
			t.Error("Expected non-empty summary")
		}
		if summary != mockLLM.MockResponses[1] {
			t.Errorf("Expected summary '%s', got '%s'", mockLLM.MockResponses[1], summary)
		}

		// Verify LLM was called for summary generation
		if mockLLM.GenerateCallCount != 2 {
			t.Errorf("Expected LLM to be called 2 times, got %d", mockLLM.GenerateCallCount)
		}
	})

	// ACT 4: Verify Database Storage
	t.Run("Step 4: Verify Summary Cached in Database", func(t *testing.T) {
		article, found := articleSvc.GetArticle(testURL)
		if !found {
			t.Fatal("Article not found in database")
		}

		// Verify summary is cached
		if article.Summary == "" {
			t.Error("Summary was not cached in the article")
		}
		if article.Summary != mockLLM.MockResponses[1] {
			t.Errorf("Expected cached summary '%s', got '%s'", mockLLM.MockResponses[1], article.Summary)
		}

		// Verify all article data is intact
		if article.URL != testURL {
			t.Errorf("Expected URL '%s', got '%s'", testURL, article.URL)
		}
		if article.Title != testTitle {
			t.Errorf("Expected title '%s', got '%s'", testTitle, article.Title)
		}
		if article.Content != testContent {
			t.Errorf("Expected content '%s', got '%s'", testContent, article.Content)
		}
	})
}

func TestFullFlow_MultipleArticles(t *testing.T) {
	// ARRANGE
	mockLLM := &mockLLMClient{
		MockResponses: []string{"Mock summary for multiple articles test"},
	}

	analysisSvc := analysis.NewService()
	articleSvc := article.NewService(analysisSvc, mockLLM)

	testArticles := []struct {
		url     string
		title   string
		content string
	}{
		{
			url:     "https://techcrunch.com/2025/07/25/sam-altman-warns-theres-no-legal-confidentiality-when-using-chatgpt-as-a-therapist/",
			title:   "Sam Altman warns there's no legal confidentiality when using ChatGPT as a therapist",
			content: "Sam Altman discusses AI confidentiality issues...",
		},
		{
			url:     "https://techcrunch.com/2025/07/25/meta-names-shengjia-zhao-as-chief-scientist-of-ai-superintelligence-unit/",
			title:   "Meta names Shengjia Zhao as chief scientist of AI superintelligence unit",
			content: "Meta has appointed Shengjia Zhao as chief scientist...",
		},
		{
			url:     "https://techcrunch.com/2025/07/26/tesla-vet-says-that-reviewing-real-products-not-mockups-is-the-key-to-staying-innovative/",
			title:   "Tesla vet says that reviewing real products, not mockups, is the key to staying innovative",
			content: "A Tesla veteran discusses the importance of real product testing...",
		},
	}

	// ACT: Process multiple articles
	for _, testArticle := range testArticles {
		err := articleSvc.ProcessArticle(context.Background(), testArticle.url, testArticle.title, testArticle.content)
		if err != nil {
			t.Fatalf("Failed to process article %s: %v", testArticle.url, err)
		}
	}

	// ASSERT: Verify all articles were stored
	articles := articleSvc.GetAllArticles()
	if len(articles) != len(testArticles) {
		t.Errorf("Expected %d articles, got %d", len(testArticles), len(articles))
	}

	// Verify each article
	for _, testArticle := range testArticles {
		article, found := articleSvc.GetArticle(testArticle.url)
		if !found {
			t.Errorf("Article with URL '%s' was not found", testArticle.url)
			continue
		}
		if article.Title != testArticle.title {
			t.Errorf("Expected title '%s', got '%s'", testArticle.title, article.Title)
		}
		if article.Content != testArticle.content {
			t.Errorf("Expected content '%s', got '%s'", testArticle.content, article.Content)
		}
	}
}

func TestFullFlow_ErrorHandling(t *testing.T) {
	// ARRANGE
	mockLLM := &mockLLMClient{
		MockResponses: []string{""}, // No response for successful plan creation, but error for execution
		MockError:     errors.New("LLM API error"),
	}

	analysisSvc := analysis.NewService()
	articleSvc := article.NewService(analysisSvc, mockLLM)
	// Create a dummy prompt loader and factory for the planner.
	mockPromptLoader, err := prompts.NewLoader("v1")
	if err != nil {
		t.Fatalf("Failed to create mock prompt loader: %v", err)
	}
	promptFactory := prompts.NewFactory(prompts.ModelGPT4Turbo, mockPromptLoader)
	plannerSvc := planner.NewService(mockLLM, promptFactory)

	testURL := "https://example.com/error-test"
	testTitle := "Error Test Article"
	testContent := "This is a test article for error handling."

	// ACT: Process article
	err = articleSvc.ProcessArticle(context.Background(), testURL, testTitle, testContent)
	if err != nil {
		t.Fatalf("Failed to process article: %v", err)
	}

	// Try to create plan with LLM error
	availableArticles := articleSvc.GetAllArticles()
	var promptArticles []*prompts.Article
	for _, art := range availableArticles {
		promptArticles = append(promptArticles, &prompts.Article{
			URL:   art.URL,
			Title: art.Title,
		})
	}

	query := "summarize the error test article"
	_, err = plannerSvc.CreatePlan(context.Background(), query, promptArticles)

	// ASSERT: Should handle LLM error gracefully
	if err == nil {
		t.Error("Expected error from LLM, got nil")
	}
	if !strings.Contains(err.Error(), "LLM API error") {
		t.Errorf("Expected error to contain 'LLM API error', got '%s'", err.Error())
	}
}
