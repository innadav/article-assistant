package analysis

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"article-assistant/internal/llm"
)

// AnalysisService provides LLM-based analysis for entities, keywords, and topics
type AnalysisService struct {
	LLM llm.Client
}

// NewAnalysisService creates a new analysis service
func NewAnalysisService(llmClient llm.Client) *AnalysisService {
	return &AnalysisService{
		LLM: llmClient,
	}
}

// EntityMatch represents a matched entity with confidence
type EntityMatch struct {
	Entity     string  `json:"entity"`
	Confidence float64 `json:"confidence"`
	Category   string  `json:"category"` // person, organization, location, etc.
}

// KeywordMatch represents a matched keyword with relevance
type KeywordMatch struct {
	Keyword   string  `json:"keyword"`
	Relevance float64 `json:"relevance"`
	Context   string  `json:"context"`
}

// TopicMatch represents a matched topic with score
type TopicMatch struct {
	Topic       string  `json:"topic"`
	Score       float64 `json:"score"`
	Description string  `json:"description"`
}

// AnalysisResult contains the complete LLM analysis
type AnalysisResult struct {
	Entities []EntityMatch  `json:"entities"`
	Keywords []KeywordMatch `json:"keywords"`
	Topics   []TopicMatch   `json:"topics"`
	Summary  string         `json:"summary"`
}

// AnalyzeContent performs comprehensive LLM-based analysis of content
func (s *AnalysisService) AnalyzeContent(ctx context.Context, content string) (*AnalysisResult, error) {
	prompt := fmt.Sprintf(`Analyze the following content and extract:

1. Main entities (people, organizations, locations, etc.) with confidence scores
2. Key topics and keywords with relevance scores
3. Main themes and topics with descriptions
4. A brief summary

Content: %s

Please respond in JSON format with the following structure:
{
  "entities": [{"entity": "name", "confidence": 0.95, "category": "person"}],
  "keywords": [{"keyword": "technology", "relevance": 0.9, "context": "main theme"}],
  "topics": [{"topic": "AI Development", "score": 0.85, "description": "Discussion about AI advancement"}],
  "summary": "Brief summary of the content"
}`, content)

	response, err := s.LLM.GenerateText(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate analysis: %w", err)
	}

	// Parse JSON response
	var result AnalysisResult
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		// If JSON parsing fails, try to extract structured data manually
		return s.parseUnstructuredResponse(response), nil
	}

	return &result, nil
}

// MatchEntities finds matching entities between content and a reference list
func (s *AnalysisService) MatchEntities(ctx context.Context, content string, referenceEntities []string) ([]EntityMatch, error) {
	entitiesStr := strings.Join(referenceEntities, ", ")

	prompt := fmt.Sprintf(`Given the following content and a list of reference entities, identify which entities from the reference list are mentioned or related to the content. Provide confidence scores.

Content: %s

Reference entities: %s

Respond in JSON format:
{"matches": [{"entity": "name", "confidence": 0.95, "category": "organization"}]}`, content, entitiesStr)

	response, err := s.LLM.GenerateText(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to match entities: %w", err)
	}

	var result struct {
		Matches []EntityMatch `json:"matches"`
	}

	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, fmt.Errorf("failed to parse entity matches: %w", err)
	}

	return result.Matches, nil
}

// MatchKeywords finds matching keywords between content and a reference list
func (s *AnalysisService) MatchKeywords(ctx context.Context, content string, referenceKeywords []string) ([]KeywordMatch, error) {
	keywordsStr := strings.Join(referenceKeywords, ", ")

	prompt := fmt.Sprintf(`Given the following content and a list of reference keywords, identify which keywords from the reference list are relevant to the content. Provide relevance scores and context.

Content: %s

Reference keywords: %s

Respond in JSON format:
{"matches": [{"keyword": "technology", "relevance": 0.9, "context": "main theme"}]}`, content, keywordsStr)

	response, err := s.LLM.GenerateText(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to match keywords: %w", err)
	}

	var result struct {
		Matches []KeywordMatch `json:"matches"`
	}

	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, fmt.Errorf("failed to parse keyword matches: %w", err)
	}

	return result.Matches, nil
}

// MatchTopics finds matching topics between content and a reference list
func (s *AnalysisService) MatchTopics(ctx context.Context, content string, referenceTopics []string) ([]TopicMatch, error) {
	topicsStr := strings.Join(referenceTopics, ", ")

	prompt := fmt.Sprintf(`Given the following content and a list of reference topics, identify which topics from the reference list are relevant to the content. Provide relevance scores and descriptions.

Content: %s

Reference topics: %s

Respond in JSON format:
{"matches": [{"topic": "AI Development", "score": 0.85, "description": "Discussion about AI advancement"}]}`, content, topicsStr)

	response, err := s.LLM.GenerateText(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to match topics: %w", err)
	}

	var result struct {
		Matches []TopicMatch `json:"matches"`
	}

	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, fmt.Errorf("failed to parse topic matches: %w", err)
	}

	return result.Matches, nil
}

// FindSimilarContent finds content similar to the given content based on entities, keywords, and topics
func (s *AnalysisService) FindSimilarContent(ctx context.Context, content string, candidateContents []string) ([]SimilarityMatch, error) {
	// First analyze the source content
	analysis, err := s.AnalyzeContent(ctx, content)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze source content: %w", err)
	}

	var matches []SimilarityMatch

	// Compare with each candidate
	for i, candidate := range candidateContents {
		similarity, err := s.calculateSimilarity(ctx, analysis, candidate)
		if err != nil {
			continue // Skip on error
		}

		matches = append(matches, SimilarityMatch{
			Index:      i,
			Content:    candidate,
			Similarity: similarity,
		})
	}

	return matches, nil
}

// SimilarityMatch represents a content match with similarity score
type SimilarityMatch struct {
	Index      int     `json:"index"`
	Content    string  `json:"content"`
	Similarity float64 `json:"similarity"`
}

// calculateSimilarity calculates similarity between analyzed content and candidate content
func (s *AnalysisService) calculateSimilarity(ctx context.Context, analysis *AnalysisResult, candidate string) (float64, error) {
	// Extract entities, keywords, and topics from analysis
	var entities, keywords, topics []string

	for _, e := range analysis.Entities {
		entities = append(entities, e.Entity)
	}
	for _, k := range analysis.Keywords {
		keywords = append(keywords, k.Keyword)
	}
	for _, t := range analysis.Topics {
		topics = append(topics, t.Topic)
	}

	prompt := fmt.Sprintf(`Calculate the similarity between the candidate content and the reference elements. Consider entity overlap, keyword relevance, and topic alignment.

Candidate content: %s

Reference entities: %s
Reference keywords: %s
Reference topics: %s

Provide a similarity score between 0.0 and 1.0 where 1.0 is identical and 0.0 is completely unrelated.

Respond in JSON format:
{"similarity": 0.75}`, candidate, strings.Join(entities, ", "), strings.Join(keywords, ", "), strings.Join(topics, ", "))

	response, err := s.LLM.GenerateText(ctx, prompt)
	if err != nil {
		return 0.0, fmt.Errorf("failed to calculate similarity: %w", err)
	}

	var result struct {
		Similarity float64 `json:"similarity"`
	}

	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return 0.0, fmt.Errorf("failed to parse similarity: %w", err)
	}

	return result.Similarity, nil
}

// parseUnstructuredResponse attempts to parse unstructured LLM response
func (s *AnalysisService) parseUnstructuredResponse(response string) *AnalysisResult {
	// Fallback parsing for when LLM doesn't return proper JSON
	result := &AnalysisResult{
		Entities: []EntityMatch{},
		Keywords: []KeywordMatch{},
		Topics:   []TopicMatch{},
		Summary:  "Analysis completed but structured data unavailable",
	}

	// Try to extract some basic information
	lines := strings.Split(response, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(strings.ToLower(line), "summary") {
			result.Summary = line
			break
		}
	}

	return result
}
