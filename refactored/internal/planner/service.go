package planner

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"article-chat-system/internal/llm"
	"article-chat-system/internal/prompts"
)

// Service is responsible for creating an execution plan from a user query.
type Service struct {
	client        llm.GenerativeModel
	promptFactory *prompts.Factory
}

func NewService(client llm.GenerativeModel, promptFactory *prompts.Factory) *Service {
	return &Service{client: client, promptFactory: promptFactory}
}

// CreatePlan analyzes a user query and returns a structured QueryPlan.
func (s *Service) CreatePlan(ctx context.Context, query string, availableArticles []*prompts.Article) (*QueryPlan, error) {
	// Convert article.Article to prompts.Article (this conversion logic should now be in the caller)
	// var promptArticles []*prompts.Article
	// for _, art := range availableArticles {
	// 	promptArticles = append(promptArticles, &prompts.Article{
	// 		URL:   art.URL,
	// 		Title: art.Title,
	// 	})
	// }

	prompt := s.promptFactory.GeneratePlannerPrompt(query, availableArticles)

	log.Printf("Generating plan for query: %s", query)

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	resp, err := s.client.GenerateContent(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content from LLM: %w", err)
	}

	// Extract the JSON part of the response.
	jsonResponse := resp.Candidates[0].Content.Parts[0].Text

	var plan QueryPlan
	if err := json.Unmarshal([]byte(jsonResponse), &plan); err != nil {
		return nil, fmt.Errorf("failed to unmarshal plan from LLM response: %w. Response was: %s", err, jsonResponse)
	}

	log.Printf("Successfully created plan. Intent: %s, Targets: %v", plan.Intent, plan.Targets)
	return &plan, nil
}

// buildPlannerPrompt constructs the prompt to send to the LLM for planning.
// This is a critical step in guiding the LLM to produce the correct JSON output.
// Now deprecated, replaced by prompts.Factory.GeneratePlannerPrompt
/*
func (s *Service) buildPlannerPrompt(query string, articles []*Article) string {
	var articleContext strings.Builder
	for _, art := range articles {
		fmt.Fprintf(&articleContext, "- URL: %s, Title: %s\n", art.URL, art.Title)
	}

	// This prompt is engineered to make the LLM act as a JSON-based function caller.
	return fmt.Sprintf(`
		You are an expert system that analyzes user queries and converts them into a structured JSON plan.
		Your task is to determine the user's intent and identify the target articles based on the provided context.

		## Available Intents:
		- "SUMMARIZE": User wants a summary of one or more articles.
		- "KEYWORDS": User wants to extract keywords from one or more articles.
		- "SENTIMENT": User wants the sentiment of one or more articles.
		- "COMPARE_TONE": User wants to compare the tone/sentiment between two articles.
		- "FIND_BY_TOPIC": User is asking for articles that discuss a specific topic.
		- "COMPARE_POSITIVITY": User wants to know which of two or more articles is more positive about a specific topic. The topic will be in the query.
		- "FIND_COMMON_ENTITIES": User wants to know the most commonly discussed people, organizations, or places across all articles.
		- "UNKNOWN": The user's intent cannot be determined.

		## Context: Available Articles
		Here is a list of articles you can work with:
		%s

		## User Query:
		"%s"

		## Instructions:
		1. Analyze the user's query carefully.
		2. Identify the single best 'intent'.
		3. Identify the article URLs that are the 'targets' of the query. Match user references like "the tesla article" to the correct URL from the context.
		4. For "FIND_BY_TOPIC" or "COMPARE_POSITIVITY", populate the 'parameters' field with the specific topic keyword(s) from the query (e.g., "economic trends", "AI regulation").
		5. For "FIND_COMMON_ENTITIES", the 'targets' array can be empty.
		6. Respond with ONLY the JSON object representing the plan. Do not include any other text or markdown formatting.

		Example response for a query "summarize the article about Tesla":
		{
			"intent": "SUMMARIZE",
			"targets": ["https://techcrunch.com/2025/07/26/tesla-vet-says-that-reviewing-real-products-not-mockups-is-the-key-to-staying-innovative/"],
			"parameters": [],
			"question": "summarize the article about Tesla"
		}
	`, articleContext.String(), query)
}
*/
