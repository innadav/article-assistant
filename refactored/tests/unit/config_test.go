// tests/unit/config_test.go
package unit

import (
	"os"
	"testing"

	"article-chat-system/internal/config"
)

func TestConfig_DefaultValues(t *testing.T) {
	// Clear environment variables
	os.Unsetenv("INITIAL_ARTICLE_URLS")
	os.Unsetenv("GEMINI_API_KEY")

	cfg := config.New()

	// Test default URLs
	if len(cfg.InitialArticleURLs) == 0 {
		t.Error("Expected default URLs to be set")
	}

	// Check that default URLs contain expected values
	expectedURLs := []string{
		"https://techcrunch.com/2025/07/26/tesla-vet-says-that-reviewing-real-products-not-mockups-is-the-key-to-staying-innovative/",
		"https://techcrunch.com/2025/07/25/sam-altman-warns-theres-no-legal-confidentiality-when-using-chatgpt-as-a-therapist/",
		"https://techcrunch.com/2025/07/25/meta-names-shengjia-zhao-as-chief-scientist-of-ai-superintelligence-unit/",
	}

	for _, expectedURL := range expectedURLs {
		found := false
		for _, actualURL := range cfg.InitialArticleURLs {
			if actualURL == expectedURL {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected URL '%s' not found in default URLs", expectedURL)
		}
	}

	// Test API key (should be empty when not set)
	if cfg.GeminiAPIKey != "" {
		t.Errorf("Expected empty API key, got '%s'", cfg.GeminiAPIKey)
	}
}

func TestConfig_CustomValues(t *testing.T) {
	// Set custom environment variables
	customURLs := "https://example.com/1,https://example.com/2,https://example.com/3"
	customAPIKey := "test-api-key-123"

	os.Setenv("INITIAL_ARTICLE_URLS", customURLs)
	os.Setenv("GEMINI_API_KEY", customAPIKey)
	defer func() {
		os.Unsetenv("INITIAL_ARTICLE_URLS")
		os.Unsetenv("GEMINI_API_KEY")
	}()

	cfg := config.New()

	// Test custom URLs
	if len(cfg.InitialArticleURLs) != 3 {
		t.Errorf("Expected 3 URLs, got %d", len(cfg.InitialArticleURLs))
	}

	expectedURLs := []string{
		"https://example.com/1",
		"https://example.com/2",
		"https://example.com/3",
	}

	for i, expectedURL := range expectedURLs {
		if cfg.InitialArticleURLs[i] != expectedURL {
			t.Errorf("Expected URL '%s', got '%s'", expectedURL, cfg.InitialArticleURLs[i])
		}
	}

	// Test custom API key
	if cfg.GeminiAPIKey != customAPIKey {
		t.Errorf("Expected API key '%s', got '%s'", customAPIKey, cfg.GeminiAPIKey)
	}
}

func TestConfig_EmptyURLs(t *testing.T) {
	// Set empty URLs
	os.Setenv("INITIAL_ARTICLE_URLS", "")
	defer os.Unsetenv("INITIAL_ARTICLE_URLS")

	cfg := config.New()

	// Should fall back to default URLs
	if len(cfg.InitialArticleURLs) == 0 {
		t.Error("Expected default URLs when INITIAL_ARTICLE_URLS is empty")
	}
}

func TestConfig_SingleURL(t *testing.T) {
	// Set single URL (no commas)
	singleURL := "https://example.com/single"
	os.Setenv("INITIAL_ARTICLE_URLS", singleURL)
	defer os.Unsetenv("INITIAL_ARTICLE_URLS")

	cfg := config.New()

	if len(cfg.InitialArticleURLs) != 1 {
		t.Errorf("Expected 1 URL, got %d", len(cfg.InitialArticleURLs))
	}

	if cfg.InitialArticleURLs[0] != singleURL {
		t.Errorf("Expected URL '%s', got '%s'", singleURL, cfg.InitialArticleURLs[0])
	}
}
