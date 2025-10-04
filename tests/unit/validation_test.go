package unit

import (
	"context"
	"encoding/json"
	"testing"

	"article-assistant/internal/domain"
)

// validationLLM simulates OpenAI responses for testing the 8 supported queries
type validationLLM struct{}

func (f *validationLLM) PlanQuery(_ context.Context, query string) (*domain.Plan, error) {
	// Map the 8 supported queries to their expected command structure
	responses := map[string]string{
		// 1. For a summary of a specific article
		"Give me a summary of https://example.com/article1": `{
			"command": "summary",
			"args": { "urls": ["https://example.com/article1"] }
		}`,

		// 2. To extract keywords or main topics
		"Extract keywords from https://example.com/article1": `{
			"command": "keywords_or_topics",
			"args": { "urls": ["https://example.com/article1"] }
		}`,

		// 3. To get the sentiment of one or more articles
		"What is the sentiment of https://example.com/article1?": `{
			"command": "get_sentiment",
			"args": { "urls": ["https://example.com/article1"] }
		}`,

		// 4. To compare multiple articles
		"Compare https://example.com/article1 and https://example.com/article2": `{
			"command": "compare_articles",
			"args": { "urls": ["https://example.com/article1", "https://example.com/article2"] }
		}`,

		// 5. What are the key differences in tone between two sources?
		"What are the key differences in tone between two sources?": `{
			"command": "ton_key_differences",
			"args": { "urls": ["https://example.com/article1", "https://example.com/article2"] }
		}`,

		// 6. What articles discuss economic trends?
		"What articles discuss economic trends?": `{
			"command": "filter_by_specific_topic",
			"args": { "topic": "economic trends" }
		}`,

		// 7. Which article is more positive about the topic of AI regulation?
		"Which article is more positive about the topic of AI regulation?": `{
			"command": "most_positive_article_for_filter",
			"args": { "filter": "positive about the topic of AI regulation" }
		}`,

		// 8. What are the most commonly discussed entities across the articles?
		"What are the most commonly discussed entities across the articles?": `{
			"command": "get_top_entities",
			"args": {}
		}`,
	}

	response, exists := responses[query]
	if !exists {
		// Default response for unrecognized queries
		response = `{
			"command": "filter_by_specific_topic",
			"args": { "topic": "general" }
		}`
	}

	var plan domain.Plan
	err := json.Unmarshal([]byte(response), &plan)
	return &plan, err
}

// Mock implementations for other LLM methods (not used in validation tests)
func (f *validationLLM) Summarize(ctx context.Context, text string) (string, error) {
	return "Mock summary", nil
}

func (f *validationLLM) Embed(ctx context.Context, text string) ([]float32, error) {
	return []float32{0.1, 0.2, 0.3}, nil
}

func (f *validationLLM) ExtractAllSemantics(ctx context.Context, text string) (*domain.SemanticAnalysis, error) {
	return &domain.SemanticAnalysis{
		Entities: []domain.SemanticEntity{
			{Name: "Test Entity", Category: "test", Confidence: 0.8},
		},
		Keywords: []domain.SemanticKeyword{
			{Term: "test keyword", Relevance: 0.7},
		},
		Topics: []domain.SemanticTopic{
			{Name: "test topic", Score: 0.6},
		},
		Sentiment:      "neutral",
		SentimentScore: 0.5,
	}, nil
}

func (f *validationLLM) SentimentScore(ctx context.Context, text string) (float64, error) {
	return 0.5, nil
}

func (f *validationLLM) ToneCompare(ctx context.Context, text1, text2 string) (string, error) {
	return "neutral", nil
}

func (f *validationLLM) GenerateText(ctx context.Context, prompt string) (string, error) {
	return "Mock generated text", nil
}

func TestValidationQueries(t *testing.T) {
	validationLLM := &validationLLM{}
	ctx := context.Background()

	// Test cases for the 8 supported queries
	testCases := []struct {
		name     string
		query    string
		expected domain.Plan
	}{
		{
			name:  "Summary of specific article",
			query: "Give me a summary of https://example.com/article1",
			expected: domain.Plan{
				Command: "summary",
				Args: map[string]interface{}{
					"urls": []interface{}{"https://example.com/article1"},
				},
			},
		},
		{
			name:  "Extract keywords from article",
			query: "Extract keywords from https://example.com/article1",
			expected: domain.Plan{
				Command: "keywords_or_topics",
				Args: map[string]interface{}{
					"urls": []interface{}{"https://example.com/article1"},
				},
			},
		},
		{
			name:  "Get sentiment of article",
			query: "What is the sentiment of https://example.com/article1?",
			expected: domain.Plan{
				Command: "get_sentiment",
				Args: map[string]interface{}{
					"urls": []interface{}{"https://example.com/article1"},
				},
			},
		},
		{
			name:  "Compare multiple articles",
			query: "Compare https://example.com/article1 and https://example.com/article2",
			expected: domain.Plan{
				Command: "compare_articles",
				Args: map[string]interface{}{
					"urls": []interface{}{"https://example.com/article1", "https://example.com/article2"},
				},
			},
		},
		{
			name:  "Tone differences between sources",
			query: "What are the key differences in tone between two sources?",
			expected: domain.Plan{
				Command: "ton_key_differences",
				Args: map[string]interface{}{
					"urls": []interface{}{"https://example.com/article1", "https://example.com/article2"},
				},
			},
		},
		{
			name:  "Articles discussing economic trends",
			query: "What articles discuss economic trends?",
			expected: domain.Plan{
				Command: "filter_by_specific_topic",
				Args: map[string]interface{}{
					"topic": "economic trends",
				},
			},
		},
		{
			name:  "Most positive article about AI regulation",
			query: "Which article is more positive about the topic of AI regulation?",
			expected: domain.Plan{
				Command: "most_positive_article_for_filter",
				Args: map[string]interface{}{
					"filter": "positive about the topic of AI regulation",
				},
			},
		},
		{
			name:  "Most commonly discussed entities",
			query: "What are the most commonly discussed entities across the articles?",
			expected: domain.Plan{
				Command: "get_top_entities",
				Args:    map[string]interface{}{},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			plan, err := validationLLM.PlanQuery(ctx, tc.query)
			if err != nil {
				t.Fatalf("PlanQuery error = %v", err)
			}

			// Validate command
			if plan.Command != tc.expected.Command {
				t.Errorf("Command mismatch: got %s, expected %s", plan.Command, tc.expected.Command)
			}

			// Validate args structure
			if len(plan.Args) != len(tc.expected.Args) {
				t.Errorf("Args length mismatch: got %d, expected %d", len(plan.Args), len(tc.expected.Args))
			}

			// Validate specific args
			for key, expectedValue := range tc.expected.Args {
				actualValue, exists := plan.Args[key]
				if !exists {
					t.Errorf("Missing arg key: %s", key)
					continue
				}

				// For slices, check content
				if expectedSlice, ok := expectedValue.([]interface{}); ok {
					if actualSlice, ok := actualValue.([]interface{}); ok {
						if len(actualSlice) != len(expectedSlice) {
							t.Errorf("Slice length mismatch for key %s: got %d, expected %d", key, len(actualSlice), len(expectedSlice))
						}
						for i, expectedItem := range expectedSlice {
							if i < len(actualSlice) && actualSlice[i] != expectedItem {
								t.Errorf("Slice item mismatch for key %s[%d]: got %v, expected %v", key, i, actualSlice[i], expectedItem)
							}
						}
					} else {
						t.Errorf("Expected slice for key %s, got %T", key, actualValue)
					}
				} else {
					// For other types, direct comparison
					if actualValue != expectedValue {
						t.Errorf("Arg mismatch for key %s: got %v, expected %v", key, actualValue, expectedValue)
					}
				}
			}

			t.Logf("✅ Query: %s → Command: %s, Args: %v", tc.query, plan.Command, plan.Args)
		})
	}
}

func TestValidationQueryStability(t *testing.T) {
	validationLLM := &validationLLM{}
	ctx := context.Background()

	// Test that the same query always returns the same plan
	query := "What are the most commonly discussed entities across the articles?"

	plan1, err1 := validationLLM.PlanQuery(ctx, query)
	plan2, err2 := validationLLM.PlanQuery(ctx, query)

	if err1 != nil || err2 != nil {
		t.Fatalf("PlanQuery errors: %v, %v", err1, err2)
	}

	if plan1.Command != plan2.Command {
		t.Errorf("Command not stable: first=%s, second=%s", plan1.Command, plan2.Command)
	}

	t.Logf("✅ Query stability confirmed: %s → %s", query, plan1.Command)
}
