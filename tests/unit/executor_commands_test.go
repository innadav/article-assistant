package unit

import (
	"context"
	"strings"
	"testing"

	"article-assistant/internal/domain"
	"article-assistant/internal/executor"
)

// Test that the executor can be created and registered commands work
func TestExecutorCommandPattern(t *testing.T) {
	// Test that executor can be created
	ex := executor.NewExecutor()
	if ex == nil {
		t.Fatal("executor should not be nil")
	}

	// Test that unknown command returns error message
	plan := &domain.Plan{Command: "unknown_command"}
	resp, err := ex.Execute(context.Background(), plan, "test query", []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !strings.Contains(resp.Answer, "Command not supported") {
		t.Errorf("expected error message for unknown command, got: %s", resp.Answer)
	}
}

// Test individual command creation
func TestCommandCreation(t *testing.T) {
	// Test that all command types can be created
	// This tests the command pattern structure

	// Summary command
	summaryCmd := &executor.SummaryCommand{}
	if summaryCmd == nil {
		t.Error("SummaryCommand should be creatable")
	}

	// Keywords command
	keywordsCmd := &executor.FetchKeywordsOrTopicsCommand{}
	if keywordsCmd == nil {
		t.Error("KeywordsOrTopicsCommand should be creatable")
	}

	// Sentiment command
	sentimentCmd := &executor.FetchSentimentCommand{}
	if sentimentCmd == nil {
		t.Error("SentimentCommand should be creatable")
	}

	// Compare command
	compareCmd := &executor.CompareCommand{}
	if compareCmd == nil {
		t.Error("CompareCommand should be creatable")
	}

	// Tone command
	toneCmd := &executor.ToneKeyDfferencesCommand{}
	if toneCmd == nil {
		t.Error("ToneCommand should be creatable")
	}

	// More positive command
	morePositiveCmd := &executor.FetchMostPositivesByFilter{}
	if morePositiveCmd == nil {
		t.Error("MorePositiveCommand should be creatable")
	}

	// Top entities command
	topEntitiesCmd := &executor.FetchTopEntitiesFromDBCommand{}
	if topEntitiesCmd == nil {
		t.Error("TopEntitiesCommand should be creatable")
	}

	// Search command
	searchCmd := &executor.FilterFromVectorDBByFilter{}
	if searchCmd == nil {
		t.Error("SearchCommand should be creatable")
	}
}

// Test command registration
func TestCommandRegistration(t *testing.T) {
	ex := executor.NewExecutor()

	// Register a test command
	testCmd := &executor.SummaryCommand{}
	ex.Register("test_summary", testCmd)

	// Test that registered command can be executed
	plan := &domain.Plan{Command: "test_summary"}
	resp, err := ex.Execute(context.Background(), plan, "test query", []string{})

	// Should not return "Command not supported" error
	if strings.Contains(resp.Answer, "Command not supported") {
		t.Error("registered command should not return 'Command not supported'")
	}

	// Should handle missing ID gracefully
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
