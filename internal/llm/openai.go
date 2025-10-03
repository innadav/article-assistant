package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"article-assistant/internal/domain"

	"github.com/sashabaranov/go-openai"
)

type OpenAIClient struct {
	c *openai.Client
}

func New(apiKey string) *OpenAIClient {
	return &OpenAIClient{
		c: openai.NewClient(apiKey),
	}
}

// getModelLimits returns context and output limits for different models
func getModelLimits(model string) (int, int) {
	switch model {
	case openai.GPT4:
		return 8192, 4096
	case openai.GPT4Turbo:
		return 128000, 4096
	case openai.GPT3Dot5Turbo:
		return 16385, 4096
	case openai.GPT3Dot5Turbo16K:
		return 16385, 4096
	default:
		return 16385, 4096 // Default to GPT-3.5-turbo limits
	}
}

// calculateMaxTokens calculates safe max tokens for output based on input size and model limits
func calculateMaxTokens(inputText string, model string) int {
	contextLimit, outputLimit := getModelLimits(model)

	// Estimate input tokens (rough: ~4 chars per token)
	inputTokens := len(inputText) / 4

	// Add overhead for prompt and formatting (~200 tokens for semantic extraction)
	promptOverhead := 200
	totalInputTokens := inputTokens + promptOverhead

	// Calculate available tokens for output
	availableForOutput := contextLimit - totalInputTokens

	// Take the minimum of desired, available, and model output limit
	maxTokens := availableForOutput
	if availableForOutput < maxTokens {
		maxTokens = availableForOutput
	}
	if outputLimit < maxTokens {
		maxTokens = availableForOutput
	}

	fmt.Printf("Token calc: input=%d, available=%d, desired=%d, final=%d\n",
		inputTokens, availableForOutput, maxTokens)

	return maxTokens
}

func (o *OpenAIClient) Summarize(ctx context.Context, text string) (string, error) {
	model := openai.GPT3Dot5Turbo
	maxTokens := calculateMaxTokens(text, model)
	truncatedText := truncateTextForModel(text, maxTokens)

	resp, err := o.c.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: model,
		Messages: []openai.ChatCompletionMessage{{
			Role:    "user",
			Content: "Summarize this text concisely while preserving key information:\n" + truncatedText,
		}},
		MaxTokens:   maxTokens,
		Temperature: 0,
	})
	if err != nil {
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}

func (o *OpenAIClient) Embed(ctx context.Context, text string) ([]float32, error) {
	resp, err := o.c.CreateEmbeddings(ctx, openai.EmbeddingRequestStrings{
		Input: []string{text},
		Model: openai.SmallEmbedding3,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data[0].Embedding, nil
}

func (o *OpenAIClient) ToneCompare(ctx context.Context, text1, text2 string) (string, error) {
	joined := fmt.Sprintf("%s\n---\n%s", text1, text2)
	model := openai.GPT3Dot5Turbo
	maxTokens := calculateMaxTokens(joined, model) // Tone analysis is more concise

	resp, err := o.c.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: model,
		Messages: []openai.ChatCompletionMessage{{
			Role:    "user",
			Content: "Compare tone across these summaries:\n" + joined,
		}},
		MaxTokens:   maxTokens,
		Temperature: 0, // Consistent tone analysis
	})
	if err != nil {
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}

func (o *OpenAIClient) Compare(ctx context.Context, summaries []string) (string, error) {
	joined := strings.Join(summaries, "\n---\n")
	model := openai.GPT3Dot5Turbo
	maxTokens := calculateMaxTokens(joined, model) // Comparison needs detailed output

	resp, err := o.c.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: model,
		Messages: []openai.ChatCompletionMessage{{
			Role:    "user",
			Content: "Compare these summaries and highlight key differences:\n" + joined,
		}},
		MaxTokens:   maxTokens,
		Temperature: 0, // Consistent comparisons
	})
	if err != nil {
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}

func (o *OpenAIClient) ExtractAllSemantics(ctx context.Context, text string) (*domain.SemanticAnalysis, error) {
	model := openai.GPT3Dot5Turbo
	maxTokens := calculateMaxTokens(text, model) // Conservative ratio for semantic extraction to prevent response overflow

	truncatedText := truncateTextForModel(text, maxTokens) // Truncate for semantic extraction

	prompt := fmt.Sprintf(`Extract entities, keywords, topics, sentiment, and tone from this text. Return JSON in this exact format:
{
  "entities": [{"name": "entity_name", "category": "person|organization|location|technology|other", "confidence": 0.85}],
  "keywords": [{"term": "keyword", "relevance": 0.8, "context": "brief context"}],
  "topics": [{"name": "topic_name", "score": 0.75, "description": "brief description"}],
  "sentiment": "positive|negative|neutral",
  "sentiment_score": 0.75,
  "tone": "professional|casual|analytical|critical|optimistic|pessimistic"
}

Rules:
- Extract 3-7 entities, 5-10 keywords, 2-5 topics
- sentiment_score must be a number between 0.0 and 1.0
- Only include items with confidence/relevance/score >= 0.6
- Sort by score/confidence/relevance (highest first)
- Return valid JSON only`, truncatedText)


	resp, err := o.c.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: model,
		Messages: []openai.ChatCompletionMessage{{
			Role:    "user",
			Content: prompt,
		}},
		MaxTokens:   maxTokens,
		Temperature: 0, // Deterministic results for structured data extraction
	})
	if err != nil {
		return nil, err
	}

	var analysis domain.SemanticAnalysis
	jsonStr := strings.TrimSpace(resp.Choices[0].Message.Content)

	// Debug: log the raw response
	fmt.Printf("LLM Response: %s\n", jsonStr)

	if err := json.Unmarshal([]byte(jsonStr), &analysis); err != nil {
		// Try to clean up the JSON response and parse again
		cleaned := cleanJSONResponse(jsonStr)
		if err := json.Unmarshal([]byte(cleaned), &analysis); err != nil {
			fmt.Printf("Failed to parse JSON response: %v\n", err)
			return createEmptySemanticAnalysis(), nil
		}
	}

	return &analysis, nil
}

func (o *OpenAIClient) SentimentScore(ctx context.Context, text string) (float64, error) {
	model := openai.GPT3Dot5Turbo
	maxTokens := calculateMaxTokens(text, model) // Short response for sentiment score

	resp, err := o.c.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: model,
		Messages: []openai.ChatCompletionMessage{{
			Role:    "user",
			Content: "Rate the sentiment of this text on a scale from 0.0 (very negative) to 1.0 (very positive). Return only the number:\n" + text,
		}},
		MaxTokens:   maxTokens,
		Temperature: 0,
	})
	if err != nil {
		return 0.5, err
	}

	scoreStr := strings.TrimSpace(resp.Choices[0].Message.Content)
	score, err := strconv.ParseFloat(scoreStr, 64)
	if err != nil {
		return 0.5, err
	}

	// Clamp to valid range
	if score < 0.0 {
		score = 0.0
	} else if score > 1.0 {
		score = 1.0
	}

	return score, nil
}

func (o *OpenAIClient) GenerateText(ctx context.Context, prompt string) (string, error) {
	model := openai.GPT3Dot5Turbo
	maxTokens := calculateMaxTokens(prompt, model)

	resp, err := o.c.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: model,
		Messages: []openai.ChatCompletionMessage{{
			Role:    "user",
			Content: prompt,
		}},
		MaxTokens:   maxTokens,
		Temperature: 0.7,
	})
	if err != nil {
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}

func (o *OpenAIClient) PlanQuery(ctx context.Context, query string) (*domain.Plan, error) {
	model := openai.GPT3Dot5Turbo

	prompt := fmt.Sprintf(`You are a query planner for an article assistant. Map user queries to commands with arguments.

Supported commands:
- summary: Get summary of specific articles (requires URLs)
- keywords_or_topics: Extract keywords/topics from articles (requires URLs)  
- get_sentiment: Get sentiment of articles (requires URLs)
- compare_articles: Compare multiple articles (requires URLs)
- ton_key_differences: Analyze tone differences between articles (requires URLs)
- get_list_articles: Find articles by topic/filter (uses filter argument)
- get_article: Find most positive article about a topic (uses filter argument)
- get_top_entities: Get most common entities across all articles (no arguments)

Rules:
1. Extract URLs from query if provided
2. Extract filter/topic from query for search commands
3. Return JSON in this exact format:
{"command": "command_name", "args": {"urls": ["url1"], "filter": "topic"}}

Examples:
- "Summary of https://example.com" → {"command": "summary", "args": {"urls": ["https://example.com"]}}
- "What articles discuss AI?" → {"command": "get_list_articles", "args": {"filter": "AI"}}
- "Most positive about AI regulation" → {"command": "get_article", "args": {"filter": "AI regulation"}}
- "Top entities" → {"command": "get_top_entities", "args": {}}

Query: %s`, query)

	resp, err := o.c.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: model,
		Messages: []openai.ChatCompletionMessage{{
			Role:    "user",
			Content: prompt,
		}},
		MaxTokens:   400,
		Temperature: 0, // Deterministic planning
	})
	if err != nil {
		return nil, err
	}

	var plan domain.Plan
	jsonStr := strings.TrimSpace(resp.Choices[0].Message.Content)

	if err := json.Unmarshal([]byte(jsonStr), &plan); err != nil {
		// Try to clean up the JSON response and parse again
		cleaned := cleanJSONResponse(jsonStr)
		if err := json.Unmarshal([]byte(cleaned), &plan); err != nil {
			return nil, fmt.Errorf("failed to parse plan JSON: %v", err)
		}
	}

	return &plan, nil
}

// truncateTextForModel truncates text to fit within model context limits
func truncateTextForModel(text string, maxInputTokens int) string {
	// Estimate tokens (rough: ~4 chars per token)
	estimatedTokens := len(text) / 4

	if estimatedTokens <= maxInputTokens {
		return text
	}

	// Calculate how many characters we can keep
	maxChars := maxInputTokens * 4
	if len(text) <= maxChars {
		return text
	}

	// Truncate and add ellipsis
	truncated := text[:maxChars-3] + "..."
	fmt.Printf("Truncation: Truncated to %d chars\n", len(truncated))
	return truncated
}

// cleanJSONResponse attempts to clean malformed JSON responses
func cleanJSONResponse(jsonStr string) string {
	// Remove markdown code blocks
	cleaned := strings.TrimPrefix(jsonStr, "```json")
	cleaned = strings.TrimPrefix(cleaned, "```")
	cleaned = strings.TrimSuffix(cleaned, "```")

	// Remove any leading/trailing whitespace
	cleaned = strings.TrimSpace(cleaned)

	return cleaned
}

// createEmptySemanticAnalysis creates an empty semantic analysis as fallback
func createEmptySemanticAnalysis() *domain.SemanticAnalysis {
	return &domain.SemanticAnalysis{
		Entities:       []domain.SemanticEntity{},
		Keywords:       []domain.SemanticKeyword{},
		Topics:         []domain.SemanticTopic{},
		Sentiment:      "neutral",
		SentimentScore: 0.5,
		Tone:           "neutral",
	}
}
