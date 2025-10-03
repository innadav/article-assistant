package integration

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"article-assistant/internal/domain"
	"article-assistant/internal/executor"
	"article-assistant/internal/llm"
	"article-assistant/internal/repository"

	_ "github.com/lib/pq"
)

func TestCommandIntegration(t *testing.T) {
	// Skip if no API key provided
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("Skipping integration test: OPENAI_API_KEY not set")
	}

	// Test database connection
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5433/article_assistant?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	repo := repository.NewRepo(db)
	llmClient := llm.New(os.Getenv("OPENAI_API_KEY"))
	commandExecutor := executor.NewExecutorWithCommands(repo, llmClient)

	// Test URLs for ingestion
	testURLs := []string{
		"https://techcrunch.com/2025/07/26/astronomer-winks-at-viral-notoriety-with-temporary-spokesperson-gwyneth-paltrow/",
		"https://techcrunch.com/2025/07/26/allianz-life-says-majority-of-customers-personal-data-stolen-in-cyberattack/",
	}

	// Ingest test articles first
	t.Run("IngestTestArticles", func(t *testing.T) {
		for _, url := range testURLs {
			// Check if article already exists
			articles, err := repo.GetArticlesByURLs(context.Background(), []string{url})
			if err != nil {
				t.Errorf("Failed to check existing article: %v", err)
				continue
			}

			if len(articles) == 0 {
				t.Logf("Article not found, would need to ingest: %s", url)
				// Note: In a real test, we would ingest here
				// For now, we'll skip if articles don't exist
			} else {
				t.Logf("Found existing article: %s", articles[0].Title)
			}
		}
	})

	// Test cases for the 8 supported queries
	testCases := []struct {
		name  string
		query string
		urls  []string
	}{
		{
			name:  "Summary of specific article",
			query: "Give me a summary of https://techcrunch.com/2025/07/26/astronomer-winks-at-viral-notoriety-with-temporary-spokesperson-gwyneth-paltrow/",
			urls:  []string{"https://techcrunch.com/2025/07/26/astronomer-winks-at-viral-notoriety-with-temporary-spokesperson-gwyneth-paltrow/"},
		},
		{
			name:  "Extract keywords from article",
			query: "Extract keywords from https://techcrunch.com/2025/07/26/astronomer-winks-at-viral-notoriety-with-temporary-spokesperson-gwyneth-paltrow/",
			urls:  []string{"https://techcrunch.com/2025/07/26/astronomer-winks-at-viral-notoriety-with-temporary-spokesperson-gwyneth-paltrow/"},
		},
		{
			name:  "Get sentiment of article",
			query: "What is the sentiment of https://techcrunch.com/2025/07/26/astronomer-winks-at-viral-notoriety-with-temporary-spokesperson-gwyneth-paltrow/",
			urls:  []string{"https://techcrunch.com/2025/07/26/astronomer-winks-at-viral-notoriety-with-temporary-spokesperson-gwyneth-paltrow/"},
		},
		{
			name:  "Compare multiple articles",
			query: "Compare https://techcrunch.com/2025/07/26/astronomer-winks-at-viral-notoriety-with-temporary-spokesperson-gwyneth-paltrow/ and https://techcrunch.com/2025/07/26/allianz-life-says-majority-of-customers-personal-data-stolen-in-cyberattack/",
			urls:  testURLs,
		},
		{
			name:  "Tone differences between sources",
			query: "What are the key differences in tone between https://techcrunch.com/2025/07/26/astronomer-winks-at-viral-notoriety-with-temporary-spokesperson-gwyneth-paltrow/ and https://techcrunch.com/2025/07/26/allianz-life-says-majority-of-customers-personal-data-stolen-in-cyberattack/",
			urls:  testURLs,
		},
		{
			name:  "Articles discussing economic trends",
			query: "What articles discuss economic trends?",
			urls:  []string{}, // No specific URLs - search all
		},
		{
			name:  "Most positive article about AI regulation",
			query: "Which article is more positive about the topic of AI regulation?",
			urls:  []string{}, // No specific URLs - search all
		},
		{
			name:  "Most commonly discussed entities",
			query: "What are the most commonly discussed entities across the articles?",
			urls:  []string{}, // No specific URLs - analyze all
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Generate plan using LLM
			plan, err := llmClient.PlanQuery(context.Background(), tc.query)
			if err != nil {
				t.Errorf("Failed to generate plan: %v", err)
				return
			}

			// Validate plan structure
			if plan.Command == "" {
				t.Error("Plan command is empty")
				return
			}

			if plan.Args == nil {
				t.Error("Plan args is nil")
				return
			}

			t.Logf("Generated plan: Command=%s, Args=%v", plan.Command, plan.Args)

			// Execute the command
			response, err := commandExecutor.Execute(context.Background(), plan, tc.query, tc.urls)
			if err != nil {
				t.Errorf("Failed to execute command: %v", err)
				return
			}

			// Validate response structure
			if response == nil {
				t.Error("Response is nil")
				return
			}

			if response.Answer == "" {
				t.Error("Response answer is empty")
				return
			}

			if response.Task == "" {
				t.Error("Response task is empty")
				return
			}

			// Validate that the response task matches the plan command
			if response.Task != plan.Command {
				t.Errorf("Response task mismatch: got %s, expected %s", response.Task, plan.Command)
			}

			t.Logf("✅ Command executed successfully: %s → %s", plan.Command, response.Task)
			t.Logf("Response: %s", response.Answer)
		})
	}
}

func TestCommandRegistry(t *testing.T) {
	// Test that all expected commands are registered
	db, err := sql.Open("postgres", "postgres://postgres:postgres@localhost:5433/article_assistant?sslmode=disable")
	if err != nil {
		t.Skip("Database not available for registry test")
	}
	defer db.Close()

	repo := repository.NewRepo(db)
	llmClient := llm.New(os.Getenv("OPENAI_API_KEY"))
	commandExecutor := executor.NewExecutorWithCommands(repo, llmClient)

	expectedCommands := []string{
		"summary",
		"keywords_or_topics",
		"get_sentiment",
		"compare_articles",
		"ton_key_differences",
		"get_list_articles",
		"get_article",
		"get_top_entities",
	}

	// Test each command with a minimal plan
	for _, cmd := range expectedCommands {
		t.Run("Command_"+cmd, func(t *testing.T) {
			plan := &domain.Plan{
				Command: cmd,
				Args:    map[string]interface{}{},
			}

			// Add appropriate args based on command
			switch cmd {
			case "summary", "keywords_or_topics", "get_sentiment", "compare_articles", "ton_key_differences":
				plan.Args["urls"] = []string{"https://example.com/test"}
			case "get_list_articles":
				plan.Args["topic"] = "test topic"
			case "get_article":
				plan.Args["filter"] = "test filter"
			}

			response, err := commandExecutor.Execute(context.Background(), plan, "test query", []string{})

			// We expect some commands to fail due to missing data, but not due to unregistered commands
			if err != nil {
				if err.Error() == "Command not supported: "+cmd {
					t.Errorf("Command %s is not registered", cmd)
				} else {
					t.Logf("Command %s is registered (failed with expected error: %v)", cmd, err)
				}
			} else if response != nil {
				t.Logf("✅ Command %s is registered and executed successfully", cmd)
			}
		})
	}
}
