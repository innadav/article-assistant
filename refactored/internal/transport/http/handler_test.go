package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"article-chat-system/internal/article"
	"article-chat-system/internal/planner"
	"article-chat-system/internal/processing"
	"article-chat-system/internal/prompts"
	"article-chat-system/internal/strategies"
)

// --- Mocks for all Handler Dependencies ---

type mockArticleService struct {
	GetAllArticlesFunc func() []*article.Article
}

func (m *mockArticleService) GetAllArticles() []*article.Article {
	if m.GetAllArticlesFunc != nil {
		return m.GetAllArticlesFunc()
	}
	return nil
}

type mockPlanner struct {
	CreatePlanFunc func(ctx context.Context, query string, availableArticles []*prompts.Article) (*planner.QueryPlan, error)
}

func (m *mockPlanner) CreatePlan(ctx context.Context, query string, availableArticles []*prompts.Article) (*planner.QueryPlan, error) {
	if m.CreatePlanFunc != nil {
		return m.CreatePlanFunc(ctx, query, availableArticles)
	}
	return nil, nil
}

type mockExecutor struct {
	ExecutePlanFunc func(ctx context.Context, plan *planner.QueryPlan, articleSvc *article.Service, promptFactory *prompts.Factory) (string, error)
}

func (m *mockExecutor) ExecutePlan(ctx context.Context, plan *planner.QueryPlan, articleSvc *article.Service, promptFactory *prompts.Factory) (string, error) {
	if m.ExecutePlanFunc != nil {
		return m.ExecutePlanFunc(ctx, plan, articleSvc, promptFactory)
	}
	return "", nil
}

// We don't need to mock the other dependencies for this specific test.
// We can pass nil because the handler doesn't use them directly in the /chat route.

func TestHandler_handleChat_Success(t *testing.T) {
	// ARRANGE
	// 1. Setup the mock planner to return a specific plan.
	mockPlannerSvc := &mockPlanner{
		CreatePlanFunc: func(ctx context.Context, query string, availableArticles []*prompts.Article) (*planner.QueryPlan, error) {
			return &planner.QueryPlan{Intent: planner.IntentSummarize},
			nil
		},
	}

	// 2. Setup the mock executor to return a specific answer.
	mockExecutorSvc := &mockExecutor{
		ExecutePlanFunc: func(ctx context.Context, plan *planner.QueryPlan, articleSvc *article.Service, promptFactory *prompts.Factory) (string, error) {
			// We can even check if it received the correct plan.
			if plan.Intent != planner.IntentSummarize {
				t.Errorf("Executor received wrong intent: got %s, want %s", plan.Intent, planner.IntentSummarize)
			}
			return "Final successful answer", nil
		},
	}

	// 3. Initialize the handler with our mocks.
	h := NewHandler(&mockArticleService{},
		mockPlannerSvc,
		mockExecutorSvc,
		nil, // mockPromptFactory not needed for this test path
		nil, // mockProcessingFacade not needed for this test path
	)

	// 4. Create a fake HTTP request.
	requestBody, _ := json.Marshal(ChatRequest{Query: "test query"})
	req := httptest.NewRequest(http.MethodPost, "/chat", bytes.NewBuffer(requestBody))
	rr := httptest.NewRecorder() // This recorder will capture the HTTP response.

	// ACT
	// Serve the request using our handler.
	h.Routes().ServeHTTP(rr, req)

	// ASSERT
	// 1. Check the HTTP status code.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// 2. Check the response body.
	var responseBody ChatResponse
	if err := json.NewDecoder(rr.Body).Decode(&responseBody); err != nil {
		t.Fatalf("Could not decode response body: %v", err)
	}

	expectedAnswer := "Final successful answer"
	if responseBody.Answer != expectedAnswer {
		t.Errorf("handler returned unexpected body: got '%v' want '%v'", responseBody.Answer, expectedAnswer)
	}
}
