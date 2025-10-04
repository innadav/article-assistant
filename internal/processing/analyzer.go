package processing

import (
	"context"
	"fmt"
	"strings"

	"article-assistant/internal/article"
	"article-assistant/internal/llm"
)

// Analyzer is responsible for performing the initial, high-level analysis of an article.
type Analyzer struct {
	llmClient llm.Client
}

// NewAnalyzer creates a new analyzer with its required dependencies.
func NewAnalyzer(llmClient llm.Client) *Analyzer {
	return &Analyzer{llmClient: llmClient}
}

// InitialAnalysis generates a summary, sentiment, and keywords for a new article.
func (a *Analyzer) InitialAnalysis(ctx context.Context, art *article.Article) error {
	// 1. Determine the complexity based on the article's word count.
	wordCount := len(strings.Fields(art.Content))
	summarySentences := 2
	keywordCount := 5

	if wordCount > 750 { // Threshold for a more complex article
		summarySentences = 5
		keywordCount = 10
	}

	// 2. Dynamically build the prompt based on the determined complexity.
	prompt := fmt.Sprintf(`
		Analyze the following article content.

		1. Provide a concise summary of %d sentences.
		2. Provide a one-word sentiment (Positive, Negative, or Neutral).
		3. Provide the %d most important keywords and named entities (people, companies, locations), returned as a comma-separated list.

		Format the response exactly as follows:
		SUMMARY: [Your summary here]
		SENTIMENT: [Your one-word sentiment here]
		KEYWORDS: [keyword1, entity1, keyword2, ...]

		--- ARTICLE CONTENT ---
		%s
	`, summarySentences, keywordCount, art.Content)

	// 3. Call the LLM and parse the structured response.
	resp, err := a.llmClient.GenerateContent(ctx, prompt)
	if err != nil {
		return fmt.Errorf("initial analysis LLM call failed: %w", err)
	}

	parseLLMResponse(resp.Text, art)
	return nil
}

// parseLLMResponse is a helper function to robustly parse the structured analysis.
func parseLLMResponse(response string, art *article.Article) {
	lines := strings.Split(response, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "SUMMARY:") {
			art.Summary = strings.TrimSpace(strings.TrimPrefix(line, "SUMMARY:"))
		} else if strings.HasPrefix(line, "SENTIMENT:") {
			art.Sentiment = strings.TrimSpace(strings.TrimPrefix(line, "SENTIMENT:"))
		} else if strings.HasPrefix(line, "KEYWORDS:") {
			keywordsStr := strings.TrimSpace(strings.TrimPrefix(line, "KEYWORDS:"))
			art.Topics = strings.Split(keywordsStr, ", ")
		}
	}
}
