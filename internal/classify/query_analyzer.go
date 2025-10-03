package classify

import (
	"article-assistant/internal/domain"
	"strings"
)

// QueryAnalysis contains the classification result
type QueryAnalysis struct {
	QueryType    string
	ResponseType string
	NeedsURLs    bool   // Whether this query requires specific URLs
	FilterTopic  string // Extracted topic for filtering (e.g., "AI regulation", "economic trends")
}

// AnalyzeQuery classifies the query and determines response format
func AnalyzeQuery(query string) QueryAnalysis {
	q := strings.ToLower(strings.TrimSpace(query))

	// 1. Summary queries - single article, text response
	if containsAny(q, []string{"summary", "summarize", "what is this article about"}) {
		return QueryAnalysis{
			QueryType:    domain.QuerySummary,
			ResponseType: domain.ResponseText,
			NeedsURLs:    true,
		}
	}

	// 2. Keywords/Topics queries - data response
	if containsAny(q, []string{"keywords", "main topics", "key topics", "extract keywords"}) {
		return QueryAnalysis{
			QueryType:    domain.QueryKeywords,
			ResponseType: domain.ResponseData,
			NeedsURLs:    false, // Can work on all articles or subset
		}
	}

	// 3. Sentiment queries - text response from DB
	if containsAny(q, []string{"sentiment", "how does this feel", "positive or negative"}) {
		return QueryAnalysis{
			QueryType:    domain.QuerySentiment,
			ResponseType: domain.ResponseText,
			NeedsURLs:    false,
		}
	}

	// 4. Article comparison - text response via LLM
	if containsAny(q, []string{"compare", "difference", "contrast", "vs"}) &&
		!containsAny(q, []string{"tone", "positive"}) {
		return QueryAnalysis{
			QueryType:    domain.QueryCompare,
			ResponseType: domain.ResponseText,
			NeedsURLs:    false, // Will use multiple articles
		}
	}

	// 5. Tone comparison - text response via LLM
	if containsAny(q, []string{"tone", "key differences in tone"}) {
		return QueryAnalysis{
			QueryType:    domain.QueryTone,
			ResponseType: domain.ResponseText,
			NeedsURLs:    false,
		}
	}

	// 6. Article search - article list response
	if containsAny(q, []string{"what articles", "which articles", "articles discuss", "articles about"}) {
		topic := extractSearchTopic(q)
		return QueryAnalysis{
			QueryType:    domain.QuerySearch,
			ResponseType: domain.ResponseArticleList,
			NeedsURLs:    false,
			FilterTopic:  topic,
		}
	}

	// 7. Most positive article - article list response (single article)
	if containsAny(q, []string{"which article is more positive", "most positive", "more positive about"}) {
		topic := extractPositiveTopic(q)
		return QueryAnalysis{
			QueryType:    domain.QueryMorePositive,
			ResponseType: domain.ResponseArticleList,
			NeedsURLs:    false,
			FilterTopic:  topic,
		}
	}

	// 8. Top entities - data response
	if containsAny(q, []string{"most commonly discussed entities", "top entities", "common entities"}) {
		return QueryAnalysis{
			QueryType:    domain.QueryTopEntities,
			ResponseType: domain.ResponseData,
			NeedsURLs:    false,
		}
	}

	// Default: unknown
	return QueryAnalysis{
		QueryType:    domain.QueryUnknown,
		ResponseType: domain.ResponseText,
		NeedsURLs:    false,
	}
}

// extractSearchTopic extracts topic from search queries like "What articles discuss economic trends?"
func extractSearchTopic(query string) string {
	q := strings.ToLower(query)

	// Look for "discuss X" pattern
	if strings.Contains(q, "discuss") {
		parts := strings.Split(q, "discuss")
		if len(parts) > 1 {
			topic := strings.TrimSpace(parts[1])
			topic = strings.TrimRight(topic, "?!.")
			return topic
		}
	}

	// Look for "about X" pattern
	if strings.Contains(q, "about") {
		parts := strings.Split(q, "about")
		if len(parts) > 1 {
			topic := strings.TrimSpace(parts[1])
			topic = strings.TrimRight(topic, "?!.")
			return topic
		}
	}

	return ""
}

// extractPositiveTopic extracts topic from positive queries like "Which article is more positive about AI regulation?"
func extractPositiveTopic(query string) string {
	q := strings.ToLower(query)

	// Look for "about X" pattern
	if strings.Contains(q, "about") {
		parts := strings.Split(q, "about")
		if len(parts) > 1 {
			topic := strings.TrimSpace(parts[1])
			topic = strings.TrimRight(topic, "?!.")
			return topic
		}
	}

	// Look for "positive X" or "positive on X"
	if strings.Contains(q, "positive") {
		// Try to extract what comes after "positive"
		parts := strings.Fields(q)
		for i, word := range parts {
			if word == "positive" && i+1 < len(parts) {
				// Skip "about", "on", "regarding"
				start := i + 1
				if start < len(parts) && (parts[start] == "about" || parts[start] == "on" || parts[start] == "regarding") {
					start++
				}
				if start < len(parts) {
					remaining := strings.Join(parts[start:], " ")
					remaining = strings.TrimRight(remaining, "?!.")
					return remaining
				}
			}
		}
	}

	return ""
}

// containsAny checks if the query contains any of the given phrases
func containsAny(query string, phrases []string) bool {
	for _, phrase := range phrases {
		if strings.Contains(query, phrase) {
			return true
		}
	}
	return false
}
