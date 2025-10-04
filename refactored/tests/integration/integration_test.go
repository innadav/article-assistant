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
	"article-chat-system/internal/strategies"
)

// mockLLMClient remains the same.
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

// --- Test Setup Helper ---

// testRig holds all the initialized services needed for a test.
type testRig struct {
	articleSvc    *article.Service
	plannerSvc    *planner.Service
	promptFactory *prompts.Factory
	mockLLM       *mockLLMClient
	ctx           context.Context
}

// setupTest initializes all services with a mock LLM.
func setupTest(t *testing.T, llmResponses []string, llmError error) *testRig {
	mockLLM := &mockLLMClient{
		MockResponses: llmResponses,
		MockError:     llmError,
	}

	analysisSvc := analysis.NewService()
	articleSvc := article.NewService(analysisSvc, mockLLM)
	articleSvc.SetStrategyExecutor(strategies.NewExecutor())

	// In a real integration test, you might use a test-specific prompt version.
	// We'll stub this for simplicity as it's not the focus of the test.
	mockPromptLoader, err := prompts.NewLoader("v1")
	if err != nil {
		t.Fatalf("Failed to create mock prompt loader: %v", err)
	}
	promptFactory := prompts.NewFactory(prompts.ModelGPT4Turbo, mockPromptLoader)
	plannerSvc := planner.NewService(mockLLM, promptFactory)

	return &testRig{
		articleSvc:    articleSvc,
		plannerSvc:    plannerSvc,
		promptFactory: promptFactory,
		mockLLM:       mockLLM,
		ctx:           context.Background(),
	}
}

// --- Refactored Tests ---

func TestFullFlow_IngestAndVerify(t *testing.T) {
	// ARRANGE
	rig := setupTest(t, nil, nil)
	testURL := "https://example.com/test-article"
	testTitle := "Test Article Title"
	testContent := "This is the content of the test article."

	// ACT
	err := rig.articleSvc.ProcessArticle(rig.ctx, testURL, testTitle, testContent)

	// ASSERT
	if err != nil {
		t.Fatalf("ProcessArticle() returned an unexpected error: %v", err)
	}

	storedArticle, found := rig.articleSvc.GetArticle(testURL)
	if !found {
		t.Fatal("Article was not found in the service after processing.")
	}
	if storedArticle.Title != testTitle {
		t.Errorf("Expected title '%s', got '%s'", testTitle, storedArticle.Title)
	}
}

func TestFullFlow_PlanAndExecute(t *testing.T) {
	// ARRANGE
	plannerResponse := `{
        "intent": "SUMMARIZE",
        "targets": ["https://example.com/test-article-2"],
        "parameters": []
    }`
	summaryResponse := "This is the mocked summary."

	rig := setupTest(t, []string{plannerResponse, summaryResponse}, nil)
	testURL := "https://example.com/test-article-2"

	// Pre-populate the service with the article needed for the test.
	rig.articleSvc.ProcessArticle(rig.ctx, testURL, "Another Title", "Some other content.")

	// ACT
	// 1. Create the plan.
	plan, err := rig.plannerSvc.CreatePlan(rig.ctx, "summarize the article", rig.articleSvc.GetAllArticlesAsPrompts())
	if err != nil {
		t.Fatalf("CreatePlan() failed: %v", err)
	}

	// 2. Execute the plan.
	finalAnswer, err := rig.articleSvc.ExecuteChatPlan(rig.ctx, plan, rig.promptFactory)
	if err != nil {
		t.Fatalf("ExecuteChatPlan() failed: %v", err)
	}

	// ASSERT
	if finalAnswer != summaryResponse {
		t.Errorf("Expected final answer '%s', got '%s'", summaryResponse, finalAnswer)
	}

	// Verify the summary was cached.
	storedArticle, _ := rig.articleSvc.GetArticle(testURL)
	if storedArticle.Summary != summaryResponse {
		t.Errorf("Summary was not cached correctly in the article object.")
	}

	// Verify the LLM was called twice (once for planning, once for summarizing).
	if rig.mockLLM.GenerateCallCount != 2 {
		t.Errorf("Expected LLM to be called 2 times, but got %d", rig.mockLLM.GenerateCallCount)
	}
}

func TestFullFlow_PlannerErrorHandling(t *testing.T) {
	// ARRANGE
	rig := setupTest(t, nil, errors.New("LLM is unavailable"))

	// ACT
	_, err := rig.plannerSvc.CreatePlan(rig.ctx, "any query", []*prompts.Article{})

	// ASSERT
	if err == nil {
		t.Fatal("Expected an error from CreatePlan, but got nil")
	}
	if !strings.Contains(err.Error(), "LLM is unavailable") {
		t.Errorf("Error message did not contain the expected text. Got: %s", err.Error())
	}
}
