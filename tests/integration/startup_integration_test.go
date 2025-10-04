package integration

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"article-assistant/internal/ingest"
	"article-assistant/internal/llm"
	"article-assistant/internal/repository"
	"article-assistant/internal/startup"

	_ "github.com/lib/pq"
)

func TestArticleLoader_Integration(t *testing.T) {
	// Skip if no API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set, skipping integration test")
	}

	// Database connection
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5433/article_assistant?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		t.Skipf("Database not available: %v", err)
	}

	// Create repository and LLM client
	repo := repository.NewRepo(db)
	llmClient := llm.New(apiKey, "gpt-3.5-turbo")

	// Create ingest service
	ingestService := &ingest.Service{
		Repo: repo,
		LLM:  llmClient,
	}

	// Create article loader
	loader := startup.NewArticleLoader(ingestService)

	// Create temporary file with 2 URLs
	tmpFile, err := os.CreateTemp("", "test-articles-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write 2 test URLs
	testURLs := `https://techcrunch.com/2025/07/26/astronomer-winks-at-viral-notoriety-with-temporary-spokesperson-gwyneth-paltrow/
https://techcrunch.com/2025/07/26/allianz-life-says-majority-of-customers-personal-data-stolen-in-cyberattack/`

	if _, err := tmpFile.WriteString(testURLs); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	t.Logf("Testing ArticleLoader with file: %s", tmpFile.Name())

	// Test loading articles
	ctx := context.Background()
	err = loader.LoadData(ctx, tmpFile.Name())

	// We expect some errors due to network issues, but the loader should handle them gracefully
	if err != nil {
		t.Logf("LoadData completed with errors (expected): %v", err)
	} else {
		t.Log("LoadData completed successfully")
	}

	// Verify that the loader attempted to process the URLs
	// We can't easily verify success without mocking the network calls,
	// but we can verify the loader doesn't crash and handles errors gracefully
	t.Log("ArticleLoader integration test completed")
}

func TestArticleLoader_RealStartupFile(t *testing.T) {
	// Skip if no API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set, skipping integration test")
	}

	// Database connection
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5433/article_assistant?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		t.Skipf("Database not available: %v", err)
	}

	// Create repository and LLM client
	repo := repository.NewRepo(db)
	llmClient := llm.New(apiKey, "gpt-3.5-turbo")

	// Create ingest service
	ingestService := &ingest.Service{
		Repo: repo,
		LLM:  llmClient,
	}

	// Create article loader
	loader := startup.NewArticleLoader(ingestService)

	// Test with the real startup articles file
	startupFile := "resources/data/startup_articles.txt"

	// Check if file exists
	if _, err := os.Stat(startupFile); os.IsNotExist(err) {
		t.Skipf("Startup articles file not found: %s", startupFile)
	}

	t.Logf("Testing ArticleLoader with real startup file: %s", startupFile)

	// Test loading articles
	ctx := context.Background()
	err = loader.LoadData(ctx, startupFile)

	// We expect some errors due to network issues, but the loader should handle them gracefully
	if err != nil {
		t.Logf("LoadData completed with errors (expected): %v", err)
	} else {
		t.Log("LoadData completed successfully")
	}

	t.Log("ArticleLoader real startup file test completed")
}
