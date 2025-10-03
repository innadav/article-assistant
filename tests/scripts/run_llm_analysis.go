package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"article-assistant/internal/analysis"
	"article-assistant/internal/llm"
	"article-assistant/internal/repository"

	_ "github.com/lib/pq"
)

// ArticleAnalysisResult combines article info with LLM analysis
type ArticleAnalysisResult struct {
	Article  ArticleInfo              `json:"article"`
	Analysis *analysis.AnalysisResult `json:"analysis"`
	Error    string                   `json:"error,omitempty"`
}

// ArticleInfo contains basic article information
type ArticleInfo struct {
	ID      string    `json:"id"`
	URL     string    `json:"url"`
	Title   string    `json:"title"`
	Summary string    `json:"summary"`
	Created time.Time `json:"created_at"`
}

// BatchAnalysisReport contains all analysis results
type BatchAnalysisReport struct {
	Timestamp     time.Time               `json:"timestamp"`
	TotalArticles int                     `json:"total_articles"`
	Successful    int                     `json:"successful"`
	Failed        int                     `json:"failed"`
	Results       []ArticleAnalysisResult `json:"results"`
}

func main() {
	// Check for API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	// Connect to database
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5433/article_assistant?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Initialize services
	repo := repository.NewRepo(db)
	llmClient := llm.New(apiKey)
	analysisService := analysis.NewAnalysisService(llmClient)

	ctx := context.Background()

	// Get all articles using vector search with a broad query
	log.Println("Fetching all articles from database...")
	// Use vector search with a broad query to get all articles
	embedding, err := llmClient.Embed(ctx, "technology business news")
	if err != nil {
		log.Fatalf("Failed to generate embedding: %v", err)
	}

	articles, err := repo.GetArticlesByVectorSearch(ctx, embedding)
	if err != nil {
		log.Fatalf("Failed to get articles: %v", err)
	}

	log.Printf("Found %d articles to analyze", len(articles))

	// Initialize report
	report := BatchAnalysisReport{
		Timestamp:     time.Now(),
		TotalArticles: len(articles),
		Results:       make([]ArticleAnalysisResult, 0, len(articles)),
	}

	// Analyze each article
	for i, article := range articles {
		log.Printf("Analyzing article %d/%d: %s", i+1, len(articles), article.Title)

		articleInfo := ArticleInfo{
			ID:      article.ID,
			URL:     article.URL,
			Title:   article.Title,
			Summary: article.Summary,
			Created: article.CreatedAt,
		}

		result := ArticleAnalysisResult{
			Article: articleInfo,
		}

		// Run LLM analysis
		analysisResult, err := analysisService.AnalyzeContent(ctx, article.Summary)
		if err != nil {
			log.Printf("Failed to analyze article %s: %v", article.ID, err)
			result.Error = err.Error()
			report.Failed++
		} else {
			result.Analysis = analysisResult
			report.Successful++
			log.Printf("Successfully analyzed: %d entities, %d keywords, %d topics",
				len(analysisResult.Entities), len(analysisResult.Keywords), len(analysisResult.Topics))
		}

		report.Results = append(report.Results, result)

		// Add a small delay to avoid rate limiting
		time.Sleep(500 * time.Millisecond)
	}

	// Create results directory
	if err := os.MkdirAll("results/llm_analysis", 0755); err != nil {
		log.Fatalf("Failed to create results directory: %v", err)
	}

	// Save results to file
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("results/llm_analysis/analysis_report_%s.json", timestamp)

	file, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Failed to create results file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		log.Fatalf("Failed to write results: %v", err)
	}

	// Print summary
	log.Printf("\n=== Analysis Complete ===")
	log.Printf("Total articles: %d", report.TotalArticles)
	log.Printf("Successful: %d", report.Successful)
	log.Printf("Failed: %d", report.Failed)
	log.Printf("Results saved to: %s", filename)

	// Also save a summary file
	summaryFile := "results/llm_analysis/latest_summary.json"
	summary := map[string]interface{}{
		"timestamp":      report.Timestamp,
		"total_articles": report.TotalArticles,
		"successful":     report.Successful,
		"failed":         report.Failed,
		"latest_file":    filename,
	}

	if summaryData, err := json.MarshalIndent(summary, "", "  "); err == nil {
		os.WriteFile(summaryFile, summaryData, 0644)
		log.Printf("Summary saved to: %s", summaryFile)
	}
}
