package unit

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"article-assistant/internal/domain"
)

// mockLLM simulates an LLM-based planner response
type mockLLM struct{}

func (m *mockLLM) PlanQuery(ctx context.Context, query string) (*domain.Plan, error) {
	switch {
	case query == "For a summary of a specific article.":
		return &domain.Plan{
			Command: "summary",
			Args:    map[string]interface{}{"urls": []string{"https://example.com/article1"}},
		}, nil

	case query == "To extract keywords or main topics.":
		return &domain.Plan{
			Command: "keywords_or_topics",
			Args:    map[string]interface{}{"urls": []string{"https://example.com/article1"}},
		}, nil

	case query == "To get the sentiment of one or more articles.":
		return &domain.Plan{
			Command: "get_sentiment",
			Args:    map[string]interface{}{"urls": []string{"https://example.com/article1"}},
		}, nil

	case query == "To compare multiple articles.":
		return &domain.Plan{
			Command: "compare_articles",
			Args:    map[string]interface{}{"urls": []string{"https://example.com/article1", "https://example.com/article2"}},
		}, nil

	case query == "What are the key differences in tone between two sources?":
		return &domain.Plan{
			Command: "ton_key_differences",
			Args:    map[string]interface{}{"urls": []string{"https://example.com/article1", "https://example.com/article2"}},
		}, nil

	case query == "What articles discuss economic trends?":
		return &domain.Plan{
			Command: "filter_by_specific_topic",
			Args:    map[string]interface{}{"topic": "economic trends"},
		}, nil

	case query == "Which article is more positive about the topic of AI regulation?":
		return &domain.Plan{
			Command: "most_positive_article_for_filter",
			Args:    map[string]interface{}{"filter": "positive about the topic of AI regulation"},
		}, nil

	case query == "What are the most commonly discussed entities across the articles?":
		return &domain.Plan{
			Command: "get_top_entities",
			Args:    map[string]interface{}{},
		}, nil
	}

	return &domain.Plan{Command: "unknown"}, nil
}

func TestMockPlannerScenarios(t *testing.T) {
	mock := &mockLLM{}

	tests := []struct {
		query       string
		wantTask    string
		wantFilters map[string]interface{}
	}{
		{
			query:       "For a summary of a specific article.",
			wantTask:    "summary",
			wantFilters: map[string]interface{}{"urls": []string{"https://example.com/article1"}},
		},
		{
			query:       "To extract keywords or main topics.",
			wantTask:    "keywords_or_topics",
			wantFilters: map[string]interface{}{"urls": []string{"https://example.com/article1"}},
		},
		{
			query:       "To get the sentiment of one or more articles.",
			wantTask:    "get_sentiment",
			wantFilters: map[string]interface{}{"urls": []string{"https://example.com/article1"}},
		},
		{
			query:       "To compare multiple articles.",
			wantTask:    "compare_articles",
			wantFilters: map[string]interface{}{"urls": []string{"https://example.com/article1", "https://example.com/article2"}},
		},
		{
			query:       "What are the key differences in tone between two sources?",
			wantTask:    "ton_key_differences",
			wantFilters: map[string]interface{}{"urls": []string{"https://example.com/article1", "https://example.com/article2"}},
		},
		{
			query:       "What articles discuss economic trends?",
			wantTask:    "filter_by_specific_topic",
			wantFilters: map[string]interface{}{"topic": "economic trends"},
		},
		{
			query:       "Which article is more positive about the topic of AI regulation?",
			wantTask:    "most_positive_article_for_filter",
			wantFilters: map[string]interface{}{"filter": "positive about the topic of AI regulation"},
		},
		{
			query:       "What are the most commonly discussed entities across the articles?",
			wantTask:    "get_top_entities",
			wantFilters: map[string]interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			plan, err := mock.PlanQuery(context.Background(), tt.query)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if plan.Command != tt.wantTask {
				t.Errorf("command mismatch: got %s, want %s", plan.Command, tt.wantTask)
			}

			// Check args match expected filters
			if !reflect.DeepEqual(plan.Args, tt.wantFilters) {
				t.Errorf("args mismatch: got %+v, want %+v", plan.Args, tt.wantFilters)
			}

			// Also check JSON serialization works
			if _, err := json.Marshal(plan); err != nil {
				t.Errorf("plan not serializable: %v", err)
			}
		})
	}
}
