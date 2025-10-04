package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"article-assistant/internal/article"
	"article-assistant/internal/planner"
	"article-assistant/internal/prompts"
)

// --- Mocks for Handler Dependencies ---
type mockPlanner struct {
	CreatePlanFunc func(ctx context.Context, query string) (*planner.QueryPlan, error)
}

func (m *mockPlanner) CreatePlan(ctx context.Context, query string) (*planner.QueryPlan, error) {
	return m.CreatePlanFunc(ctx, query)
}

type mockExecutor struct {
	ExecutePlanFunc func(ctx context.Context, plan *planner.QueryPlan, articleSvc *article.Service, promptFactory *prompts.Factory) (string, error)
}

func (m *mockExecutor) ExecutePlan(ctx context.Context, plan *planner.QueryPlan, articleSvc *article.Service, promptFactory *prompts.Factory) (string, error) {
	return m.ExecutePlanFunc(ctx, plan, articleSvc, promptFactory)
}

type mockProcessingFacade struct {
	AddNewArticleFunc func(ctx context.Context, url string) (*article.Article, error)
}

func (m *mockProcessingFacade) AddNewArticle(ctx context.Context, url string) (*article.Article, error) {
	if m.AddNewArticleFunc != nil {
		return m.AddNewArticleFunc(ctx, url)
	}
	return nil, nil
}

func TestHandler_handleChat_Success(t *testing.T) {
	// ARRANGE
	mockPlannerSvc := &mockPlanner{
		CreatePlanFunc: func(ctx context.Context, query string) (*planner.QueryPlan, error) {
			return &planner.QueryPlan{Intent: planner.IntentSummarize}, nil
		},
	}
	mockExecutorSvc := &mockExecutor{
		ExecutePlanFunc: func(ctx context.Context, plan *planner.QueryPlan, articleSvc *article.Service, promptFactory *prompts.Factory) (string, error) {
			if plan.Intent != planner.IntentSummarize {
				t.Errorf("Executor received wrong intent: got %s, want %s", plan.Intent, planner.IntentSummarize)
			}
			return "Final successful answer", nil
		},
	}

	handler := NewHandler(nil, mockPlannerSvc, mockExecutorSvc, nil, nil)

	requestBody, _ := json.Marshal(ChatRequest{Query: "test query"})
	req := httptest.NewRequest(http.MethodPost, "/chat", bytes.NewBuffer(requestBody))
	rr := httptest.NewRecorder()

	// ACT
	handler.Routes().ServeHTTP(rr, req)

	// ASSERT
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var responseBody ChatResponse
	if err := json.NewDecoder(rr.Body).Decode(&responseBody); err != nil {
		t.Fatalf("Could not decode response body: %v", err)
	}
	expectedAnswer := "Final successful answer"
	if responseBody.Answer != expectedAnswer {
		t.Errorf("handler returned unexpected body: got '%v' want '%v'", responseBody.Answer, expectedAnswer)
	}
}

func TestHandler_handleAddArticle_Success(t *testing.T) {
	// ARRANGE
	mockFacade := &mockProcessingFacade{
		AddNewArticleFunc: func(ctx context.Context, url string) (*article.Article, error) {
			return &article.Article{URL: url, Title: "Test Article"}, nil
		},
	}
	handler := NewHandler(nil, nil, nil, nil, mockFacade)

	requestBody, _ := json.Marshal(AddArticleRequest{URL: "http://example.com/new"})
	req := httptest.NewRequest(http.MethodPost, "/articles", bytes.NewBuffer(requestBody))
	rr := httptest.NewRecorder()

	// ACT
	handler.Routes().ServeHTTP(rr, req)

	// ASSERT
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	var responseBody article.Article
	if err := json.NewDecoder(rr.Body).Decode(&responseBody); err != nil {
		t.Fatalf("Could not decode response body: %v", err)
	}
	if responseBody.URL != "http://example.com/new" {
		t.Errorf("handler returned unexpected URL: got %v want %v", responseBody.URL, "http://example.com/new")
	}
}
