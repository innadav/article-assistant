package processing

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/go-shiori/go-readability"
)

// Fetcher is responsible for fetching and parsing the content of an article from a URL.
type Fetcher struct{}

// NewFetcher creates a new Fetcher.
func NewFetcher() *Fetcher {
	return &Fetcher{}
}

// FetchAndParse retrieves the content from a URL and extracts the main article body.
func (f *Fetcher) FetchAndParse(ctx context.Context, rawURL string) (*readability.Article, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for %s: %w", rawURL, err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL %s: %w", rawURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch URL %s: received status code %d", rawURL, resp.StatusCode)
	}

	parsedArticle, err := readability.FromReader(resp.Body, parsedURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse article content from %s: %w", rawURL, err)
	}

	return &parsedArticle, nil
}
