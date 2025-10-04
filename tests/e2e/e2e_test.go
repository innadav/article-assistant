package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

type ChatRequest struct {
	Query string `json:"query"`
}

type ChatResponse struct {
	Answer       string      `json:"answer"`
	Sources      interface{} `json:"sources"`
	Usage        interface{} `json:"usage"`
	Task         string      `json:"task"`
	ResponseType string      `json:"response_type"`
	Plan         struct {
		Command string                 `json:"command"`
		Args    map[string]interface{} `json:"args"`
	} `json:"plan"`
}

const (
	baseURL = "http://localhost:8080"
	timeout = 60 * time.Second
)

// TestE2EAllQueries tests all 8 supported queries end-to-end
func TestE2EAllQueries(t *testing.T) {
	// Check if server is running
	if !isServerRunning() {
		t.Skip("Server not running, skipping e2e tests")
	}

	// Wait for server to be ready
	waitForServer(t)

	testCases := []struct {
		name           string
		query          string
		expectedTask   string
		expectedAnswer string
		shouldContain  []string
	}{
		{
			name:          "Summary of specific article",
			query:         "Give me a summary of https://techcrunch.com/2025/07/26/astronomer-winks-at-viral-notoriety-with-temporary-spokesperson-gwyneth-paltrow/",
			expectedTask:  "summary",
			shouldContain: []string{"Gwyneth Paltrow", "astronomer"},
		},
		{
			name:          "Extract keywords from article",
			query:         "Extract keywords from https://techcrunch.com/2025/07/26/astronomer-winks-at-viral-notoriety-with-temporary-spokesperson-gwyneth-paltrow/",
			expectedTask:  "keywords_or_topics",
			shouldContain: []string{"keywords", "topics"},
		},
		{
			name:          "Get sentiment of article",
			query:         "What is the sentiment of https://techcrunch.com/2025/07/27/wizard-of-oz-blown-up-by-ai-for-giant-sphere-screen/?",
			expectedTask:  "get_sentiment",
			shouldContain: []string{"sentiment", "positive"},
		},
		{
			name:          "Compare multiple articles",
			query:         "Compare https://techcrunch.com/2025/07/26/astronomer-winks-at-viral-notoriety-with-temporary-spokesperson-gwyneth-paltrow/ and https://techcrunch.com/2025/07/26/allianz-life-says-majority-of-customers-personal-data-stolen-in-cyberattack/",
			expectedTask:  "compare_articles",
			shouldContain: []string{"article", "data breach", "cyberattack", "hackers", "personal data"},
		},
		{
			name:          "Tone differences between sources",
			query:         "What are the key differences in tone between https://techcrunch.com/2025/07/26/astronomer-winks-at-viral-notoriety-with-temporary-spokesperson-gwyneth-paltrow/ and https://techcrunch.com/2025/07/26/allianz-life-says-majority-of-customers-personal-data-stolen-in-cyberattack/",
			expectedTask:  "ton_key_differences",
			shouldContain: []string{"tone", "serious", "urgent"},
		},
		{
			name:          "Articles discussing economic trends",
			query:         "What articles discuss economic trends?",
			expectedTask:  "filter_by_specific_topic",
			shouldContain: []string{"economic trends", "articles"},
		},
		{
			name:          "Most positive article about AI regulation",
			query:         "Which article is more positive about the topic of AI regulation?",
			expectedTask:  "most_positive_article_for_filter",
			shouldContain: []string{"positive", "AI regulation"},
		},
		{
			name:          "Most commonly discussed entities",
			query:         "What are the most commonly discussed entities across the articles?",
			expectedTask:  "get_top_entities",
			shouldContain: []string{"entities", "OpenAI", "Meta", "Intel"},
		},
	}

	results := make(map[string]bool)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			response, err := sendChatRequest(tc.query)
			if err != nil {
				t.Errorf("Failed to send request: %v", err)
				results[tc.name] = false
				return
			}

			// Validate response structure
			if response.Task == "" {
				t.Error("Response task is empty")
				results[tc.name] = false
				return
			}

			if response.Plan.Command == "" {
				t.Error("Response plan command is empty")
				results[tc.name] = false
				return
			}

			// Validate task matches expected
			if response.Task != tc.expectedTask {
				t.Errorf("Task mismatch: got %s, expected %s", response.Task, tc.expectedTask)
				results[tc.name] = false
				return
			}

			// Validate plan command matches task
			if response.Plan.Command != response.Task {
				t.Errorf("Plan command mismatch: got %s, expected %s", response.Plan.Command, response.Task)
				results[tc.name] = false
				return
			}

			// Validate answer contains expected content
			if len(tc.shouldContain) > 0 {
				answerLower := strings.ToLower(response.Answer)
				foundKeywords := 0
				for _, expected := range tc.shouldContain {
					if strings.Contains(answerLower, strings.ToLower(expected)) {
						foundKeywords++
					}
				}
				// Require at least half of the expected keywords to be present
				requiredKeywords := (len(tc.shouldContain) + 1) / 2
				if foundKeywords < requiredKeywords {
					t.Errorf("Answer should contain at least %d of %d expected keywords. Found: %d. Expected: %v. Got: %s",
						requiredKeywords, len(tc.shouldContain), foundKeywords, tc.shouldContain, response.Answer)
					results[tc.name] = false
					return
				}
			}

			// Log success
			t.Logf("✅ %s: %s → %s", tc.name, tc.query, response.Task)
			t.Logf("   Answer: %s", response.Answer)
			results[tc.name] = true
		})
	}

	// Print summary
	t.Log("\n=== E2E Test Summary ===")
	passed := 0
	total := len(testCases)
	for name, success := range results {
		status := "❌ FAIL"
		if success {
			status = "✅ PASS"
			passed++
		}
		t.Logf("%s: %s", status, name)
	}
	t.Logf("\nOverall: %d/%d tests passed (%.1f%%)", passed, total, float64(passed)/float64(total)*100)
}

// TestE2EServerHealth tests server health endpoint
func TestE2EServerHealth(t *testing.T) {
	if !isServerRunning() {
		t.Skip("Server not running, skipping health test")
	}

	client := &http.Client{Timeout: timeout}
	resp, err := client.Get(baseURL + "/health")
	if err != nil {
		t.Fatalf("Failed to get health endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Health endpoint returned status %d, expected 200", resp.StatusCode)
	}

	var health struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		t.Errorf("Failed to decode health response: %v", err)
	}

	if health.Status != "healthy" {
		t.Errorf("Health status is %s, expected 'healthy'", health.Status)
	}

	t.Log("✅ Server health check passed")
}

// Helper functions

func isServerRunning() bool {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(baseURL + "/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func waitForServer(t *testing.T) {
	maxAttempts := 10
	for i := 0; i < maxAttempts; i++ {
		if isServerRunning() {
			return
		}
		time.Sleep(2 * time.Second)
	}
	t.Fatal("Server not ready after 20 seconds")
}

func sendChatRequest(query string) (*ChatResponse, error) {
	client := &http.Client{Timeout: timeout}

	reqBody := ChatRequest{Query: query}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := client.Post(baseURL+"/chat", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	var response ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &response, nil
}

// TestE2EIngest tests article ingestion
func TestE2EIngest(t *testing.T) {
	if !isServerRunning() {
		t.Skip("Server not running, skipping ingest test")
	}

	testURL := "https://techcrunch.com/2025/07/26/astronomer-winks-at-viral-notoriety-with-temporary-spokesperson-gwyneth-paltrow/"

	client := &http.Client{Timeout: timeout}
	reqBody := map[string]string{"url": testURL}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("Failed to marshal ingest request: %v", err)
	}

	resp, err := client.Post(baseURL+"/ingest", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("Failed to send ingest request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Ingest request failed with status %d", resp.StatusCode)
	}

	var response struct {
		Message string `json:"message"`
		Status  string `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Errorf("Failed to decode ingest response: %v", err)
	}

	if response.Status != "success" {
		t.Errorf("Ingest status is %s, expected 'success'", response.Status)
	}

	t.Log("✅ Article ingestion test passed")
}

// BenchmarkE2EQueries benchmarks the 8 main queries
func BenchmarkE2EQueries(b *testing.B) {
	if !isServerRunning() {
		b.Skip("Server not running, skipping benchmark")
	}

	queries := []string{
		"Give me a summary of https://techcrunch.com/2025/07/26/astronomer-winks-at-viral-notoriety-with-temporary-spokesperson-gwyneth-paltrow/",
		"What is the sentiment of https://techcrunch.com/2025/07/26/astronomer-winks-at-viral-notoriety-with-temporary-spokesperson-gwyneth-paltrow/",
		"What articles discuss economic trends?",
		"Which article is more positive about the topic of AI regulation?",
		"What are the most commonly discussed entities across the articles?",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		query := queries[i%len(queries)]
		_, err := sendChatRequest(query)
		if err != nil {
			b.Errorf("Benchmark query failed: %v", err)
		}
	}
}
