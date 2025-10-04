package unit

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"article-assistant/internal/domain"
)

// fakeLLM simulates OpenAI responses for testing
type fakeLLM struct {
	response string
}

func (f *fakeLLM) PlanQuery(_ context.Context, _ string) (*domain.Plan, error) {
	var plan domain.Plan
	err := json.Unmarshal([]byte(f.response), &plan)
	return &plan, err
}

func TestPlanQuery_CanonicalCases(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		response string
		expected domain.Plan
	}{
		{
			name:  "Summary of article",
			query: "Give me the summary of https://abc.com/a1",
			response: `{
				"command": "summary",
				"args": { "urls": ["https://abc.com/a1"] }
			}`,
			expected: domain.Plan{
				Command: "summary",
				Args: map[string]interface{}{
					"urls": []interface{}{"https://abc.com/a1"},
				},
			},
		},
		{
			name:  "Extract keywords",
			query: "Extract keywords from https://abc.com/a1 and https://abc.com/a2",
			response: `{
				"command": "keywords_or_topics",
				"args": { "urls": ["https://abc.com/a1", "https://abc.com/a2"] }
			}`,
			expected: domain.Plan{
				Command: "keywords_or_topics",
				Args: map[string]interface{}{
					"urls": []interface{}{"https://abc.com/a1", "https://abc.com/a2"},
				},
			},
		},
		{
			name:  "Sentiment",
			query: "What is the sentiment of https://abc.com/a1 and https://abc.com/a2?",
			response: `{
				"command": "get_sentiment",
				"args": { "urls": ["https://abc.com/a1", "https://abc.com/a2"] }
			}`,
			expected: domain.Plan{
				Command: "get_sentiment",
				Args: map[string]interface{}{
					"urls": []interface{}{"https://abc.com/a1", "https://abc.com/a2"},
				},
			},
		},
		{
			name:  "Compare multiple articles",
			query: "Compare https://a.com and https://b.com",
			response: `{
				"command": "compare_articles",
				"args": { "urls": ["https://a.com","https://b.com"] }
			}`,
			expected: domain.Plan{
				Command: "compare_articles",
				Args: map[string]interface{}{
					"urls": []interface{}{"https://a.com", "https://b.com"},
				},
			},
		},
		{
			name:  "Tone difference",
			query: "What are the key differences in tone between https://a.com and https://b.com?",
			response: `{
				"command": "ton_key_differences",
				"args": { "urls": ["https://a.com","https://b.com"] }
			}`,
			expected: domain.Plan{
				Command: "ton_key_differences",
				Args: map[string]interface{}{
					"urls": []interface{}{"https://a.com", "https://b.com"},
				},
			},
		},
		{
			name:  "Articles on economic trends",
			query: "What articles discuss economic trends?",
			response: `{
				"command": "filter_by_specific_topic",
				"args": { "topic": "economic_trends" }
			}`,
			expected: domain.Plan{
				Command: "filter_by_specific_topic",
				Args: map[string]interface{}{
					"topic": "economic_trends",
				},
			},
		},
		{
			name:  "Most positive article",
			query: "Which article is more positive about the topic of AI regulation?",
			response: `{
				"command": "most_positive_article_for_filter",
				"args": { "filter": "positive about the topic of AI regulation" }
			}`,
			expected: domain.Plan{
				Command: "most_positive_article_for_filter",
				Args: map[string]interface{}{
					"filter": "positive about the topic of AI regulation",
				},
			},
		},
		{
			name:  "Top entities across all",
			query: "What are the most commonly discussed entities across all articles?",
			response: `{
				"command": "get_top_entities",
				"args": {}
			}`,
			expected: domain.Plan{
				Command: "get_top_entities",
				Args:    map[string]interface{}{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeLLM{response: tt.response}
			got, err := fake.PlanQuery(context.Background(), tt.query)
			if err != nil {
				t.Fatalf("PlanQuery error = %v", err)
			}
			if got.Command != tt.expected.Command {
				t.Errorf("Command mismatch: got %s, expected %s", got.Command, tt.expected.Command)
			}
			// We only check keys, not deep-equality for maps with slices
			for k, v := range tt.expected.Args {
				if _, ok := got.Args[k]; !ok {
					t.Errorf("Missing arg key %s", k)
				}
				// crude slice/string check
				if fmt.Sprintf("%v", got.Args[k]) != fmt.Sprintf("%v", v) {
					t.Errorf("Arg mismatch for key %s: got %v, expected %v", k, got.Args[k], v)
				}
			}
		})
	}
}
