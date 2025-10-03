package executor

import (
	"article-assistant/internal/llm"
	"article-assistant/internal/repository"
)

// NewExecutorWithCommands creates a new executor with all commands registered
func NewExecutorWithCommands(repo *repository.Repo, llmClient *llm.OpenAIClient) *Executor {
	executor := NewExecutor()

	// Register all commands
	executor.Register("summary", &SummaryCommand{Repo: repo})
	executor.Register("keywords_or_topics", &FetchKeywordsOrTopicsCommand{Repo: repo})
	executor.Register("get_sentiment", &FetchSentimentCommand{Repo: repo})
	executor.Register("compare_articles", &CompareCommand{Repo: repo, LLM: llmClient})
	executor.Register("ton_key_differences", &ToneKeyDfferencesCommand{Repo: repo, LLM: llmClient})
	executor.Register("get_article", &FetchMostPositivesByFilter{Repo: repo, LLM: llmClient})
	executor.Register("get_top_entities", &FetchTopEntitiesFromDBCommand{Repo: repo})
	executor.Register("get_list_articles", &FilterFromVectorDBByFilter{Repo: repo, LLM: llmClient})

	return executor
}
