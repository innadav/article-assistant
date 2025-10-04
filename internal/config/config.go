package config

import (
	"os"
)

// Config holds all configuration for the application.
type Config struct {
	InitialArticleURLs []string
	LLMProvider        string
	GoogleAPIKey       string
	PromptVersion      string
	DatabaseURL        string
	Port               string
}

func New() *Config {
	return &Config{
		LLMProvider:   getEnv("LLM_PROVIDER", "google"),
		GoogleAPIKey:  getEnv("GEMINI_API_KEY", ""),
		PromptVersion: getEnv("PROMPT_VERSION", "v1"),
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://user:password@postgres:5432/articledb?sslmode=disable"),
		Port:          getEnv("PORT", "8080"),
		InitialArticleURLs: []string{
			"https://techcrunch.com/2025/07/26/astronomer-winks-at-viral-notoriety-with-temporary-spokesperson-gwyneth-paltrow/",
			"https://techcrunch.com/2025/07/26/allianz-life-says-majority-of-customers-personal-data-stolen-in-cyberattack/",
			// ... add all 20 URLs here
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
