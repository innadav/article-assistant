package prompts

// Loader loads and manages prompts based on a version.
type Loader struct {
	version string
	prompts map[string]string
}

// NewLoader creates a new prompt loader for a given version.
func NewLoader(version string) (*Loader, error) {
	// In a real application, prompts would be loaded from files or a database
	// based on the version. For now, we'll hardcode them.
	loader := &Loader{
		version: version,
		prompts: make(map[string]string),
	}

	// Hardcoded prompts for demonstration
	// Planner Prompt
	loader.prompts["planner"] = `
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
	`
	// Summary Prompt
	loader.prompts["summary"] = "Please provide a concise, one-paragraph summary of the following article:\n\n---\n\n%s"
	// Keywords Prompt
	loader.prompts["keywords"] = "Extract the 5 most important keywords or topics from the article titled '%s'. Return them as a single, comma-separated list."
	// Sentiment Prompt
	loader.prompts["sentiment"] = "Analyze the sentiment of the article titled '%s'. Respond with a single word: Positive, Negative, or Neutral."
	// Compare Tone Prompt
	loader.prompts["compare_tone"] = "Compare the tone and sentiment of the following two articles. Explain the key differences in a few sentences.\n\n### Article 1: %s\n\n### Article 2: %s"
	// Find Topic Prompt
	loader.prompts["find_topic"] = `
		Based on the following list of articles, which ones discuss the topic of "%s"?
		List the titles of only the relevant articles. If none are relevant, say so.

		## Article List:
		%s
	`
	// Compare Positivity Prompt
	loader.prompts["compare_positivity"] = `
		Analyze the following two articles specifically on the topic of "%s".
		Determine which article is more positive about this topic and explain why in one or two sentences.

		### Article 1: %s
		Content Excerpt: %s

		### Article 2: %s
		Content Excerpt: %s
	`
	// Find Common Entities Prompt
	loader.prompts["find_common_entities"] = `
		Based on the following list of article titles, identify the most commonly discussed entities (people, companies, or major topics).
		List the top 3-5 entities and briefly mention why they are significant across these articles.

		## Article Titles:
		%s
	`

	return loader, nil
}

// GetPrompt retrieves a prompt by its key.
func (l *Loader) GetPrompt(key string) string {
	return l.prompts[key]
}
