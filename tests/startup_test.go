package tests

import (
	"context"
	"errors"
	"os"
	"testing"

	"article-assistant/internal/startup"
)

// --- Mock Ingest Service ---
// This mock is needed to test the ArticleLoader.
type mockIngestService struct {
	IngestURLFunc func(ctx context.Context, url string) error
	callCount     int
	ingestedURLs  []string
}

func (m *mockIngestService) IngestURL(ctx context.Context, url string) error {
	m.callCount++
	m.ingestedURLs = append(m.ingestedURLs, url)
	if m.IngestURLFunc != nil {
		return m.IngestURLFunc(ctx, url)
	}
	return nil
}

// --- Mock Loader Tests (Unchanged but included for completeness) ---
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

// --- Article Loader Tests (Corrected and Completed) ---

func TestArticleLoader_LoadData_FileNotFound(t *testing.T) {
	// ARRANGE
	mockIngest := &mockIngestService{}
	loader := startup.NewArticleLoader(mockIngest)

	// ACT
	err := loader.LoadData(context.Background(), "non-existent-file.txt")

	// ASSERT
	// The function should return an error when the file doesn't exist.
	if err == nil {
		t.Error("Expected an error for a non-existent file, but got nil")
	}
}

func TestArticleLoader_LoadData_EmptyFile(t *testing.T) {
	// ARRANGE
	tmpFile, err := os.CreateTemp("", "empty-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	mockIngest := &mockIngestService{}
	loader := startup.NewArticleLoader(mockIngest)

	// ACT
	err = loader.LoadData(context.Background(), tmpFile.Name())

	// ASSERT
	if err != nil {
		t.Errorf("Expected no error for empty file, got: %v", err)
	}
	// Verify that the ingest service was not called for an empty file.
	if mockIngest.callCount != 0 {
		t.Errorf("Expected IngestURL to be called 0 times, but got %d", mockIngest.callCount)
	}
}

func TestArticleLoader_LoadData_Success(t *testing.T) {
	// ARRANGE
	tmpFile, err := os.CreateTemp("", "articles-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	content := `
# This is a comment and should be ignored.
https://example.com/article1

  https://example.com/article2  # This URL has whitespace
# Another comment.

https://example.com/article3
`
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	mockIngest := &mockIngestService{}
	loader := startup.NewArticleLoader(mockIngest)

	// ACT
	err = loader.LoadData(context.Background(), tmpFile.Name())

	// ASSERT
	if err != nil {
		t.Fatalf("LoadData failed unexpectedly: %v", err)
	}

	// Verify that the IngestURL method was called exactly 3 times.
	if mockIngest.callCount != 3 {
		t.Errorf("Expected IngestURL to be called 3 times, but got %d", mockIngest.callCount)
	}

	// Verify that the correct URLs were passed to the ingest service.
	expectedURLs := []string{
		"https://example.com/article1",
		"https://example.com/article2",
		"https://example.com/article3",
	}

	if len(mockIngest.ingestedURLs) != len(expectedURLs) {
		t.Fatalf("Expected %d URLs to be ingested, but got %d", len(expectedURLs), len(mockIngest.ingestedURLs))
	}

	for i, expected := range expectedURLs {
		if mockIngest.ingestedURLs[i] != expected {
			t.Errorf("URL %d: expected '%s', got '%s'", i, expected, mockIngest.ingestedURLs[i])
		}
	}
}

func TestArticleLoader_LoadData_IngestError(t *testing.T) {
	// ARRANGE
	tmpFile, err := os.CreateTemp("", "error-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	content := "https://example.com/article1\nhttps://example.com/fail"
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	// Configure the mock to return an error for a specific URL.
	mockIngest := &mockIngestService{
		IngestURLFunc: func(ctx context.Context, url string) error {
			if url == "https://example.com/fail" {
				return errors.New("ingest failed")
			}
			return nil
		},
	}
	loader := startup.NewArticleLoader(mockIngest)

	// ACT
	err = loader.LoadData(context.Background(), tmpFile.Name())

	// ASSERT
	// The function should return the error from the ingest service.
	if err == nil {
		t.Error("Expected an error from the ingest service, but got nil")
	}
	if err.Error() != "ingest failed" {
		t.Errorf("Expected error 'ingest failed', got: '%v'", err)
	}
}
