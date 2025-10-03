package llm

import (
	"context"
	"encoding/json"
	"math/rand"
	"strings"

	"article-assistant/internal/domain"
)

// MockClient implements the Client interface for testing
type MockClient struct{}

// NewMockClient creates a new mock client
func NewMockClient() *MockClient {
	return &MockClient{}
}

// Embed generates a mock embedding
func (m *MockClient) Embed(ctx context.Context, text string) ([]float32, error) {
	// Generate a deterministic embedding based on text length
	embedding := make([]float32, 1536)
	seed := int64(len(text))
	rand.Seed(seed)

	for i := range embedding {
		embedding[i] = rand.Float32()
	}
	return embedding, nil
}

// Summarize returns a mock summary
func (m *MockClient) Summarize(ctx context.Context, text string) (string, error) {
	// Return a truncated version of the input as mock summary
	if len(text) > 200 {
		return text[:200] + "...", nil
	}
	return text, nil
}

// ExtractAllSemantics returns mock semantic analysis
func (m *MockClient) ExtractAllSemantics(ctx context.Context, text string) (*domain.SemanticAnalysis, error) {
	analysis := &domain.SemanticAnalysis{
		Entities: []domain.SemanticEntity{
			{Name: "Technology", Confidence: 0.9},
			{Name: "Innovation", Confidence: 0.8},
			{Name: "Business", Confidence: 0.7},
		},
		Keywords: []domain.SemanticKeyword{
			{Term: "artificial intelligence", Relevance: 0.9},
			{Term: "machine learning", Relevance: 0.8},
			{Term: "data science", Relevance: 0.7},
		},
		Topics: []domain.SemanticTopic{
			{Name: "Technology", Score: 0.9},
			{Name: "Innovation", Score: 0.8},
			{Name: "Business", Score: 0.7},
		},
		Sentiment:      "positive",
		SentimentScore: 0.7,
		Tone:           "professional",
	}

	// Adjust based on text content
	if strings.Contains(strings.ToLower(text), "negative") {
		analysis.Sentiment = "negative"
		analysis.SentimentScore = 0.3
		analysis.Tone = "critical"
	}

	return analysis, nil
}

// ToneCompare returns mock tone comparison
func (m *MockClient) ToneCompare(ctx context.Context, text1, text2 string) (string, error) {
	return "Text 1 is more analytical while text 2 is more conversational", nil
}

// SentimentScore returns mock sentiment score
func (m *MockClient) SentimentScore(ctx context.Context, text string) (float64, error) {
	// Return a mock sentiment score based on text content
	if strings.Contains(strings.ToLower(text), "positive") {
		return 0.8, nil
	} else if strings.Contains(strings.ToLower(text), "negative") {
		return 0.2, nil
	}
	return 0.5, nil
}

// PlanQuery creates a mock plan based on the query
func (m *MockClient) PlanQuery(ctx context.Context, query string) (*domain.Plan, error) {
	query = strings.ToLower(query)

	// Mock planning logic based on query patterns
	switch {
	case strings.Contains(query, "summary") || strings.Contains(query, "summarize"):
		return &domain.Plan{
			Command: "summary",
			Args:    map[string]interface{}{"urls": []string{"https://example.com/article1"}},
		}, nil

	case strings.Contains(query, "compare") || strings.Contains(query, "comparison"):
		return &domain.Plan{
			Command: "compare_articles",
			Args:    map[string]interface{}{"urls": []string{"https://example.com/article1", "https://example.com/article2"}},
		}, nil

	case strings.Contains(query, "tone") && strings.Contains(query, "differences"):
		return &domain.Plan{
			Command: "ton_key_differences",
			Args:    map[string]interface{}{"urls": []string{"https://example.com/article1", "https://example.com/article2"}},
		}, nil

	case strings.Contains(query, "economic trends"):
		return &domain.Plan{
			Command: "get_list_articles",
			Args:    map[string]interface{}{"topic": "economic trends"},
		}, nil

	case strings.Contains(query, "top entities") || strings.Contains(query, "commonly discussed entities"):
		if strings.Contains(query, "across") || strings.Contains(query, "all articles") {
			return &domain.Plan{
				Command: "get_top_entities",
				Args:    map[string]interface{}{},
			}, nil
		}
		return &domain.Plan{
			Command: "get_top_db_entities",
			Args:    map[string]interface{}{"urls": []string{"https://example.com/article1"}},
		}, nil

	case strings.Contains(query, "positive about") || strings.Contains(query, "more positive"):
		return &domain.Plan{
			Command: "get_article",
			Args:    map[string]interface{}{"filter": "positive about the topic of AI regulation"},
		}, nil

	case strings.Contains(query, "extract keywords") || strings.Contains(query, "keywords about"):
		return &domain.Plan{
			Command: "keywords_or_topics",
			Args:    map[string]interface{}{"topic": "technology"},
		}, nil

	case strings.Contains(query, "sentiment") && strings.Contains(query, "about"):
		return &domain.Plan{
			Command: "get_sentiment",
			Args:    map[string]interface{}{"topic": "technology"},
		}, nil

	case strings.Contains(query, "semantic search") || strings.Contains(query, "vector search"):
		return &domain.Plan{
			Command: "get_list_articles",
			Args:    map[string]interface{}{"topic": "technology"},
		}, nil

	default:
		return &domain.Plan{
			Command: "keywords_or_topics",
			Args:    map[string]interface{}{"topic": "general"},
		}, nil
	}
}

// GenerateText returns mock generated text
func (m *MockClient) GenerateText(ctx context.Context, prompt string) (string, error) {
	return "This is mock generated text based on the prompt: " + prompt, nil
}

// Helper function to create mock JSON data
func createMockJSON() []byte {
	data := map[string]interface{}{
		"entities": []map[string]interface{}{
			{"name": "Technology", "confidence": 0.9},
			{"name": "Innovation", "confidence": 0.8},
		},
		"keywords": []map[string]interface{}{
			{"term": "artificial intelligence", "score": 0.9},
			{"term": "machine learning", "score": 0.8},
		},
		"topics": []map[string]interface{}{
			{"name": "Technology", "relevance": 0.9},
			{"name": "Innovation", "relevance": 0.8},
		},
	}
	b, _ := json.Marshal(data)
	return b
}
