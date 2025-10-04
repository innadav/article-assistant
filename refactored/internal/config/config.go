package config

import (
	"os"
	"strings"
)

// Config holds application configuration
type Config struct {
	InitialArticleURLs []string
	Port               string
	GeminiAPIKey       string
	LLMProvider        string
	OpenAIModel        string
	PromptVersion      string
}

// New creates a new configuration instance
func New() *Config {
	return &Config{
		InitialArticleURLs: getInitialArticleURLs(),
		Port:               getEnvOrDefault("PORT", "8080"),
		GeminiAPIKey:       getEnvOrDefault("GEMINI_API_KEY", ""),
		LLMProvider:        getEnvOrDefault("LLM_PROVIDER", "openai"), // Default to openai
		OpenAIModel:        getEnvOrDefault("OPENAI_MODEL", "gpt-4-turbo"),
		PromptVersion:      getEnvOrDefault("PROMPT_VERSION", "v1"),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getInitialArticleURLs() []string {
	if urls := os.Getenv("INITIAL_ARTICLE_URLS"); urls != "" {
		if urls == "" {
			return getDefaultURLs()
		}
		// Split by comma and trim spaces
		var result []string
		for _, url := range strings.Split(urls, ",") {
			result = append(result, strings.TrimSpace(url))
		}
		return result
	}
	return getDefaultURLs()
}

func getDefaultURLs() []string {
	return []string{
		"https://techcrunch.com/2025/07/26/tesla-vet-says-that-reviewing-real-products-not-mockups-is-the-key-to-staying-innovative/",
		"https://techcrunch.com/2025/07/25/sam-altman-warns-theres-no-legal-confidentiality-when-using-chatgpt-as-a-therapist/",
		"https://techcrunch.com/2025/07/25/meta-names-shengjia-zhao-as-chief-scientist-of-ai-superintelligence-unit/",
	}
}
