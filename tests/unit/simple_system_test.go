package unit

import (
	"context"
	"testing"

	"article-assistant/internal/llm"
)

func TestSimplePlannerWithMock(t *testing.T) {
	mockLLM := &llm.MockClient{}
	ctx := context.Background()

	testQueries := []struct {
		name              string
		query             string
		expectedTask      string
		shouldHaveFilters bool
		expectedTopic     string
	}{
		{
			name:              "Top Entities Query",
			query:             "What are the most commonly discussed entities across the articles?",
			expectedTask:      "get_top_entities",
			shouldHaveFilters: false,
			expectedTopic:     "",
		},
		{
			name:              "Positive About AI Query",
			query:             "Which article is more positive about AI regulation?",
			expectedTask:      "most_positive_article_for_filter",
			shouldHaveFilters: true,
			expectedTopic:     "positive about the topic of AI regulation",
		},
		{
			name:              "Keywords Extraction Query",
			query:             "Extract keywords about machine learning",
			expectedTask:      "keywords_or_topics",
			shouldHaveFilters: true,
			expectedTopic:     "technology",
		},
		{
			name:              "Sentiment Analysis Query",
			query:             "Get sentiment of articles about technology",
			expectedTask:      "get_sentiment",
			shouldHaveFilters: true,
			expectedTopic:     "technology",
		},
		{
			name:              "Article Comparison Query",
			query:             "Compare multiple articles about startups",
			expectedTask:      "compare_articles",
			shouldHaveFilters: true,
			expectedTopic:     "", // Don't check specific topic for comparison queries
		},
		{
			name:              "Tone Comparison Query",
			query:             "What are the key differences in tone between two sources?",
			expectedTask:      "ton_key_differences",
			shouldHaveFilters: false,
			expectedTopic:     "",
		},
		{
			name:              "Search Query",
			query:             "What articles discuss economic trends?",
			expectedTask:      "filter_by_specific_topic",
			shouldHaveFilters: true,
			expectedTopic:     "economic trends",
		},
		{
			name:              "Summary Query",
			query:             "For a summary of a specific article",
			expectedTask:      "summary",
			shouldHaveFilters: true,
			expectedTopic:     "", // Don't check specific topic for summary queries
		},
	}

	for _, tc := range testQueries {
		t.Run(tc.name, func(t *testing.T) {
			plan, err := mockLLM.PlanQuery(ctx, tc.query)
			if err != nil {
				t.Fatalf("PlanQuery failed: %v", err)
			}

			if plan.Command != tc.expectedTask {
				t.Errorf("Expected command %s, got %s", tc.expectedTask, plan.Command)
			}

			// Validate plan structure
			if plan.Command == "" {
				t.Errorf("Plan command is empty")
			}

			if plan.Args == nil {
				t.Errorf("Plan args is nil")
			}

			// Test filters
			if tc.shouldHaveFilters {
				if plan.Args == nil {
					t.Errorf("Expected filters to be present for query: %s", tc.query)
				} else {
					// Check if expected topic is in filters
					if tc.expectedTopic != "" {
						found := false
						for key, value := range plan.Args {
							if key == "topic" && value == tc.expectedTopic {
								found = true
								break
							}
							if key == "filter" && value == tc.expectedTopic {
								found = true
								break
							}
							if key == "id" && value == tc.expectedTopic {
								found = true
								break
							}
							if key == "urls" {
								// Handle array case for URLs
								if arr, ok := value.([]interface{}); ok {
									for _, v := range arr {
										if str, ok := v.(string); ok && str == tc.expectedTopic {
											found = true
											break
										}
									}
								}
								// Also check if the expected topic is any URL in the array
								if arr, ok := value.([]interface{}); ok && len(arr) > 0 {
									if str, ok := arr[0].(string); ok {
										// If expected topic matches the first URL, consider it found
										if str == tc.expectedTopic {
											found = true
										}
									}
								}
							}
							if key == "keywords" {
								// Handle array case
								if arr, ok := value.([]string); ok {
									for _, v := range arr {
										if v == tc.expectedTopic {
											found = true
											break
										}
									}
								}
							}
						}
						if !found {
							t.Errorf("Expected topic '%s' not found in args: %v", tc.expectedTopic, plan.Args)
						}
					}
				}
			} else {
				if plan.Args != nil && len(plan.Args) > 0 {
					t.Logf("Note: Args present but not expected: %v", plan.Args)
				}
			}

			t.Logf("✅ Query: %s → Command: %s, Args: %v",
				tc.query, plan.Command, plan.Args)
		})
	}
}

func TestPlanStructure(t *testing.T) {
	mockLLM := &llm.MockClient{}
	ctx := context.Background()

	// Test that plans have the expected structure
	plan, err := mockLLM.PlanQuery(ctx, "What are the most commonly discussed entities?")
	if err != nil {
		t.Fatalf("PlanQuery failed: %v", err)
	}

	// Check required fields
	if plan.Command == "" {
		t.Errorf("Command should not be empty")
	}

	if plan.Args == nil {
		t.Errorf("Args should not be nil")
	}

	t.Logf("✅ Plan structure valid: Command=%s, Args=%v",
		plan.Command, plan.Args)
}

func TestFilterExtraction(t *testing.T) {
	mockLLM := &llm.MockClient{}
	ctx := context.Background()

	testCases := []struct {
		query         string
		expectedTopic string
		expectedKey   string
	}{
		{
			query:         "Which article is more positive about AI regulation?",
			expectedTopic: "positive about the topic of AI regulation",
			expectedKey:   "filter",
		},
		{
			query:         "Extract keywords about machine learning",
			expectedTopic: "technology",
			expectedKey:   "topic",
		},
		{
			query:         "Get sentiment of articles about technology",
			expectedTopic: "technology",
			expectedKey:   "topic",
		},
		{
			query:         "What articles discuss economic trends?",
			expectedTopic: "economic trends",
			expectedKey:   "topic",
		},
	}

	for _, tc := range testCases {
		t.Run("Filter_"+tc.expectedKey, func(t *testing.T) {
			plan, err := mockLLM.PlanQuery(ctx, tc.query)
			if err != nil {
				t.Fatalf("PlanQuery failed: %v", err)
			}

			if plan.Args == nil {
				t.Errorf("Expected args for query: %s", tc.query)
				return
			}

			// Check if the expected topic is in the args
			found := false
			if tc.expectedKey == "topic" {
				if topic, ok := plan.Args["topic"]; ok && topic == tc.expectedTopic {
					found = true
				}
			} else if tc.expectedKey == "filter" {
				if filter, ok := plan.Args["filter"]; ok && filter == tc.expectedTopic {
					found = true
				}
			} else if tc.expectedKey == "keywords" {
				if keywords, ok := plan.Args["keywords"]; ok {
					if arr, ok := keywords.([]string); ok {
						for _, v := range arr {
							if v == tc.expectedTopic {
								found = true
								break
							}
						}
					}
				}
			}

			if !found {
				t.Errorf("Expected %s '%s' not found in args: %v",
					tc.expectedKey, tc.expectedTopic, plan.Args)
			}

			t.Logf("✅ Filter extraction: %s → %s: %s",
				tc.query, tc.expectedKey, tc.expectedTopic)
		})
	}
}
