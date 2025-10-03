package startup

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"article-assistant/internal/ingest"
)

// LoadOnStartupData defines the interface for loading data on startup
type LoadOnStartupData interface {
	LoadData(ctx context.Context, dataSource string) error
}

// ArticleLoader implements LoadOnStartupData for article ingestion
type ArticleLoader struct {
	ingestService *ingest.Service
}

// NewArticleLoader creates a new ArticleLoader
func NewArticleLoader(ingestService *ingest.Service) *ArticleLoader {
	return &ArticleLoader{
		ingestService: ingestService,
	}
}

// LoadData loads articles from a file in parallel
func (al *ArticleLoader) LoadData(ctx context.Context, articlesFile string) error {
	// Check if file exists
	if _, err := os.Stat(articlesFile); os.IsNotExist(err) {
		log.Printf("⚠️  Articles file not found: %s, skipping startup ingestion", articlesFile)
		return nil
	}

	file, err := os.Open(articlesFile)
	if err != nil {
		return fmt.Errorf("failed to open articles file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// Collect all URLs first
	var urls []string
	for scanner.Scan() {
		url := strings.TrimSpace(scanner.Text())
		if url == "" || strings.HasPrefix(url, "#") {
			continue
		}
		urls = append(urls, url)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading articles file: %w", err)
	}

	if len(urls) == 0 {
		log.Println("📄 No URLs found in articles file, skipping startup ingestion")
		return nil
	}

	log.Printf("📄 Starting parallel article ingestion on startup (%d articles)...", len(urls))

	// Use WaitGroup to wait for all goroutines to complete
	var wg sync.WaitGroup
	var mu sync.Mutex
	successCount := 0
	errorCount := 0

	// Process URLs in parallel (limit to 5 concurrent to avoid overwhelming the API)
	semaphore := make(chan struct{}, 5)

	for _, url := range urls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			log.Printf("📄 Ingesting: %s", url)

			err := al.ingestService.IngestURL(ctx, url)

			mu.Lock()
			if err != nil {
				log.Printf("❌ Failed to ingest %s: %v", url, err)
				errorCount++
			} else {
				log.Printf("✅ Successfully ingested: %s", url)
				successCount++
			}
			mu.Unlock()
		}(url)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	log.Printf("📊 Startup ingestion complete: ✅ %d success, ❌ %d errors", successCount, errorCount)

	if errorCount > 0 {
		return fmt.Errorf("ingestion completed with %d errors out of %d articles", errorCount, len(urls))
	}

	return nil
}

// LoadArticlesOnStartup is a convenience function that loads articles from the default file
func LoadArticlesOnStartup(ingestService *ingest.Service, articlesFile string) error {
	loader := NewArticleLoader(ingestService)
	ctx := context.Background()
	return loader.LoadData(ctx, articlesFile)
}
