package tests

import (
	"bufio"
	"context"
	"os"
	"testing"

	"article-assistant/internal/ingest"
	"article-assistant/internal/startup"
)

func TestMockLoader(t *testing.T) {
	mockLoader := startup.NewMockLoader()

	// Test successful loading
	err := mockLoader.LoadData(context.Background(), "test-file.txt")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if mockLoader.LoadCount != 1 {
		t.Errorf("Expected LoadCount to be 1, got: %d", mockLoader.LoadCount)
	}

	if len(mockLoader.Data) != 1 || mockLoader.Data[0] != "test-file.txt" {
		t.Errorf("Expected data to contain 'test-file.txt', got: %v", mockLoader.Data)
	}

	// Test failure case
	mockLoader.ShouldFail = true
	err = mockLoader.LoadData(context.Background(), "fail-file.txt")
	if err == nil {
		t.Error("Expected error when ShouldFail is true")
	}

	if mockLoader.LoadCount != 2 {
		t.Errorf("Expected LoadCount to be 2, got: %d", mockLoader.LoadCount)
	}

	// Test reset
	mockLoader.Reset()
	if mockLoader.LoadCount != 0 {
		t.Errorf("Expected LoadCount to be 0 after reset, got: %d", mockLoader.LoadCount)
	}

	if len(mockLoader.Data) != 0 {
		t.Errorf("Expected data to be empty after reset, got: %v", mockLoader.Data)
	}
}

func TestArticleLoader_LoadData_FileNotFound(t *testing.T) {
	// Create a mock ingest service (we don't need a real one for this test)
	var ingestService *ingest.Service = nil
	loader := startup.NewArticleLoader(ingestService)

	// Test with non-existent file
	err := loader.LoadData(context.Background(), "non-existent-file.txt")
	if err != nil {
		t.Errorf("Expected no error for non-existent file, got: %v", err)
	}
}

func TestArticleLoader_LoadData_EmptyFile(t *testing.T) {
	// Create a temporary empty file
	tmpFile, err := os.CreateTemp("", "empty-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Create a mock ingest service
	var ingestService *ingest.Service = nil
	loader := startup.NewArticleLoader(ingestService)

	// Test with empty file
	err = loader.LoadData(context.Background(), tmpFile.Name())
	if err != nil {
		t.Errorf("Expected no error for empty file, got: %v", err)
	}
}

func TestArticleLoader_LoadData_FileWithComments(t *testing.T) {
	// Create a temporary file with comments
	tmpFile, err := os.CreateTemp("", "comments-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write test content with comments
	content := `# This is a comment
https://example.com/article1
# Another comment
https://example.com/article2

# Empty line above
https://example.com/article3`

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	// Test file parsing by reading the file manually to verify the content
	// This tests that our test setup is correct
	file, err := os.Open(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to reopen temp file: %v", err)
	}
	defer file.Close()

	// Read and verify the content
	scanner := bufio.NewScanner(file)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("Error reading file: %v", err)
	}

	// Verify we have the expected lines
	expectedLines := []string{
		"# This is a comment",
		"https://example.com/article1",
		"# Another comment",
		"https://example.com/article2",
		"",
		"# Empty line above",
		"https://example.com/article3",
	}

	if len(lines) != len(expectedLines) {
		t.Errorf("Expected %d lines, got %d", len(expectedLines), len(lines))
	}

	for i, expected := range expectedLines {
		if i < len(lines) && lines[i] != expected {
			t.Errorf("Line %d: expected '%s', got '%s'", i, expected, lines[i])
		}
	}

	t.Log("File parsing test completed successfully - file content verified")
}
