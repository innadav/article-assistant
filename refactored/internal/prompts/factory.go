package prompts

import (
	"fmt"
	"strings"
)

// LLMModel represents the specific LLM model being used.
type LLMModel string

const (
	ModelGemini15Flash LLMModel = "gemini-1.5-flash-latest"
	ModelGPT4Turbo     LLMModel = "gpt-4-turbo-preview"
)

// Factory provides methods to generate prompts for different LLM tasks.
type Factory struct {
	model        LLMModel
	promptLoader *Loader
}

// NewFactory creates a new PromptFactory with the given LLM model and loader.
func NewFactory(model LLMModel, promptLoader *Loader) *Factory {
	return &Factory{model: model, promptLoader: promptLoader}
}

// GeneratePlannerPrompt constructs the prompt to send to the LLM for planning.
// This is a critical step in guiding the LLM to produce the correct JSON output.
func (f *Factory) GeneratePlannerPrompt(query string, articles []*Article) string {
	var articleContext strings.Builder
	for _, art := range articles {
		fmt.Fprintf(&articleContext, "- URL: %s, Title: %s\n", art.URL, art.Title)
	}

	// This prompt is engineered to make the LLM act as a JSON-based function caller.
	return fmt.Sprintf(f.promptLoader.GetPrompt("planner"), articleContext.String(), query)
}

// GenerateSummaryPrompt constructs the prompt for summarizing an article.
func (f *Factory) GenerateSummaryPrompt(content string) string {
	return fmt.Sprintf(f.promptLoader.GetPrompt("summary"), content)
}

// GenerateKeywordsPrompt constructs the prompt for extracting keywords from an article.
func (f *Factory) GenerateKeywordsPrompt(title string) string {
	return fmt.Sprintf(f.promptLoader.GetPrompt("keywords"), title)
}

// GenerateSentimentPrompt constructs the prompt for analyzing the sentiment of an article.
func (f *Factory) GenerateSentimentPrompt(title string) string {
	return fmt.Sprintf(f.promptLoader.GetPrompt("sentiment"), title)
}

// GenerateCompareTonePrompt constructs the prompt for comparing the tone of two articles.
func (f *Factory) GenerateCompareTonePrompt(title1, title2 string) string {
	return fmt.Sprintf(f.promptLoader.GetPrompt("compare_tone"), title1, title2)
}

// GenerateFindTopicPrompt constructs the prompt for finding articles discussing a specific topic.
func (f *Factory) GenerateFindTopicPrompt(topic, articleListContext string) string {
	return fmt.Sprintf(f.promptLoader.GetPrompt("find_topic"), topic, articleListContext)
}

// GenerateComparePositivityPrompt constructs the prompt for comparing positivity on a specific topic.
func (f *Factory) GenerateComparePositivityPrompt(topic, title1, excerpt1, title2, excerpt2 string) string {
	return fmt.Sprintf(f.promptLoader.GetPrompt("compare_positivity"), topic, title1, excerpt1, title2, excerpt2)
}

// GenerateFindCommonEntitiesPrompt constructs the prompt for finding common entities across articles.
func (f *Factory) GenerateFindCommonEntitiesPrompt(articleTitlesContext string) string {
	return fmt.Sprintf(f.promptLoader.GetPrompt("find_common_entities"), articleTitlesContext)
}
