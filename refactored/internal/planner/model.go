package planner

import (
	"context"
)

// QueryIntent represents the user's goal.
type QueryIntent string

const (
	IntentSummarize          QueryIntent = "SUMMARIZE"
	IntentKeywords           QueryIntent = "KEYWORDS"
	IntentSentiment          QueryIntent = "SENTIMENT"
	IntentCompareTone        QueryIntent = "COMPARE_TONE"
	IntentFindTopic          QueryIntent = "FIND_BY_TOPIC"
	IntentComparePositivity  QueryIntent = "COMPARE_POSITIVITY"
	IntentFindCommonEntities QueryIntent = "FIND_COMMON_ENTITIES"
	IntentUnknown            QueryIntent = "UNKNOWN"
)

// IntentStrategy defines the interface for all query execution strategies
type IntentStrategy interface {
	Execute(ctx context.Context, plan *QueryPlan, articleSvc interface{}, promptFactory interface{}) (string, error)
}

// QueryPlan is the structured representation of a user's request.
// The LLM will be prompted to generate a JSON object matching this structure.
type QueryPlan struct {
	Intent     QueryIntent `json:"intent"`
	Targets    []string    `json:"targets"`    // URLs of articles to act on.
	Parameters []string    `json:"parameters"` // e.g., topics to search for.
	Question   string      `json:"question"`   // The original user question.
}
