package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"article-assistant/internal/domain"
	"article-assistant/internal/executor"
	"article-assistant/internal/ingest"
	"article-assistant/internal/llm"
	"article-assistant/internal/repository"
	"article-assistant/internal/startup"

	_ "github.com/lib/pq"
)

func main() {
	// Database connection
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5433/article_assistant?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Initialize components
	repo := repository.NewRepo(db)

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}
	llmClient := llm.New(apiKey)

	ingestService := &ingest.Service{
		Repo: repo,
		LLM:  llmClient,
	}

	// Ingest articles on startup
	articlesFile := "resources/data/startup_articles.txt"
	if err := startup.LoadArticlesOnStartup(ingestService, articlesFile); err != nil {
		log.Printf("⚠️  Startup ingestion failed: %v", err)
		// Continue server startup even if ingestion fails
	}

	// Ingest endpoint
	http.HandleFunc("/ingest", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		if r.Method != "POST" {
			http.Error(w, "Method not allowed", 405)
			return
		}

		var req struct {
			URL string `json:"url"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", 400)
			return
		}

		ctx := context.Background()
		err := ingestService.IngestURL(ctx, req.URL)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to ingest URL: %v", err), 500)
			return
		}

		json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "URL ingested successfully"})
	})

	// Chat endpoint - uses simple LLM planner + executor
	http.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		if r.Method != "POST" {
			http.Error(w, "Method not allowed", 405)
			return
		}

		var req domain.ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", 400)
			return
		}

		ctx := context.Background()

		// Step 1: Create execution plan using LLM
		plan, err := llmClient.PlanQuery(ctx, req.Query)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to create query plan: %v", err), 500)
			return
		}

		// Debug: Log the plan
		log.Printf("Generated plan: %+v", plan)

		// Step 2: Execute the plan
		commandExecutor := executor.NewExecutorWithCommands(repo, llmClient)
		response, err := commandExecutor.Execute(ctx, plan, req.Query, req.URLs)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to execute query plan: %v", err), 500)
			return
		}

		// Add plan to response for debugging
		response.Plan = plan
		log.Printf("Response with plan: %+v", response)

		json.NewEncoder(w).Encode(response)
	})

	// Health check
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	})

	log.Println("🚀 Article Assistant Server with RAG Router")
	log.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
