package ingest

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ContentInfo holds information about fetched content
type ContentInfo struct {
	HTML      string
	Title     string
	FetchedAt time.Time
}

// calculateURLHash computes SHA-256 hash of the URL for caching
func calculateURLHash(url string) string {
	hash := sha256.Sum256([]byte(url))
	return fmt.Sprintf("%x", hash)
}

// fetchHTMLWithHeaders fetches HTML content (simplified version)
func fetchHTMLWithHeaders(url string) (*ContentInfo, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "ArticleAssistant/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	html := string(body)
	title := ExtractBetween(html, "<title>", "</title>")

	contentInfo := &ContentInfo{
		HTML:      html,
		Title:     strings.TrimSpace(title),
		FetchedAt: time.Now(),
	}

	return contentInfo, nil
}

// isArticleAlreadyProcessed checks if an article with the given URL hash already exists
func isArticleAlreadyProcessed(existingURLHash string) bool {
	return existingURLHash != ""
}
