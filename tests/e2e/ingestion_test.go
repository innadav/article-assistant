package e2e

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"article-assistant/internal/domain"
	"article-assistant/internal/ingest"
	"article-assistant/internal/llm"
	"article-assistant/internal/repository"

	_ "github.com/lib/pq"
)

func TestE2EIngestion(t *testing.T) {
	// Skip if no API key provided
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("Skipping e2e test: OPENAI_API_KEY not set")
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

	// Initialize components
	repo := repository.NewRepo(db)
	apiKey := os.Getenv("OPENAI_API_KEY")
	llmClient := llm.New(apiKey)

	ingestService := &ingest.Service{
		Repo: repo,
		LLM:  llmClient,
	}

	// Test URLs - just a couple for e2e testing
	testURLs := []string{
		"https://techcrunch.com/2025/07/26/astronomer-winks-at-viral-notoriety-with-temporary-spokesperson-gwyneth-paltrow/",
		"https://techcrunch.com/2025/07/26/allianz-life-says-majority-of-customers-personal-data-stolen-in-cyberattack/",
	}

	ctx := context.Background()
	successCount := 0
	errorCount := 0

	t.Log("🚀 Starting E2E ingestion test...")

	for i, url := range testURLs {
		t.Run(fmt.Sprintf("Ingest_URL_%d", i+1), func(t *testing.T) {
			t.Logf("📄 Ingesting: %s", url)

			err := ingestService.IngestURL(ctx, url)
			if err != nil {
				t.Errorf("❌ Failed to ingest %s: %v", url, err)
				errorCount++
			} else {
				t.Logf("✅ Successfully ingested: %s", url)
				successCount++
			}
		})
	}

	// Verify articles were ingested
	t.Run("Verify_Articles_In_Database", func(t *testing.T) {
		for i, url := range testURLs {
			articles, err := repo.GetArticlesByURLs(ctx, []string{url})
			if err != nil {
				t.Errorf("Failed to retrieve article %d: %v", i+1, err)
				continue
			}

			if len(articles) == 0 {
				t.Errorf("Article %d not found in database: %s", i+1, url)
			} else {
				t.Logf("✅ Article %d found in database: %s", i+1, articles[0].Title)
			}
		}
	})

	// Test ingestion via API endpoint
	t.Run("Test_Ingestion_API", func(t *testing.T) {
		// Test URL for API ingestion
		testAPIURL := "https://techcrunch.com/2025/07/27/itch-io-is-the-latest-marketplace-to-crack-down-on-adult-games/"

		// Prepare request
		reqBody := map[string]string{"url": testAPIURL}
		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		// Make API call
		resp, err := http.Post("http://localhost:8080/ingest", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("Failed to call ingestion API: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Ingestion API returned status %d, expected 200", resp.StatusCode)
		}

		// Verify article was ingested
		articles, err := repo.GetArticlesByURLs(ctx, []string{testAPIURL})
		if err != nil {
			t.Errorf("Failed to retrieve API-ingested article: %v", err)
		} else if len(articles) == 0 {
			t.Error("API-ingested article not found in database")
		} else {
			t.Logf("✅ API-ingested article found: %s", articles[0].Title)
		}
	})

	t.Logf("📊 E2E Ingestion Test Summary: ✅ %d success, ❌ %d errors", successCount, errorCount)
}

func TestE2EIngestionWithServer(t *testing.T) {
	// Skip if no API key provided
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("Skipping e2e test: OPENAI_API_KEY not set")
	}

	// Wait for server to be ready
	timeout := 30 * time.Second
	start := time.Now()

	for time.Since(start) < timeout {
		resp, err := http.Get("http://localhost:8080/health")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			break
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(1 * time.Second)
	}

	// Test ingestion via API
	testURL := "https://techcrunch.com/2025/07/25/intel-is-spinning-off-its-network-and-edge-group/"

	reqBody := map[string]string{"url": testURL}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	resp, err := http.Post("http://localhost:8080/ingest", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("Failed to call ingestion API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Ingestion API returned status %d, expected 200", resp.StatusCode)
	}

	// Test chat query to verify article was ingested
	chatReq := domain.ChatRequest{
		Query: fmt.Sprintf("Give me a summary of %s", testURL),
		URLs:  []string{testURL},
	}

	chatJSON, err := json.Marshal(chatReq)
	if err != nil {
		t.Fatalf("Failed to marshal chat request: %v", err)
	}

	chatResp, err := http.Post("http://localhost:8080/chat", "application/json", bytes.NewBuffer(chatJSON))
	if err != nil {
		t.Fatalf("Failed to call chat API: %v", err)
	}
	defer chatResp.Body.Close()

	if chatResp.StatusCode != http.StatusOK {
		t.Errorf("Chat API returned status %d, expected 200", chatResp.StatusCode)
	}

	var chatResponse domain.ChatResponse
	if err := json.NewDecoder(chatResp.Body).Decode(&chatResponse); err != nil {
		t.Fatalf("Failed to decode chat response: %v", err)
	}

	if chatResponse.Answer == "" {
		t.Error("Chat response is empty")
	} else {
		t.Logf("✅ Chat response received: %s", chatResponse.Answer)
	}
}
