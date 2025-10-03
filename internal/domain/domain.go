package domain

import "time"

// SemanticEntity represents an extracted entity with metadata
type SemanticEntity struct {
	Name       string  `json:"name"`
	Category   string  `json:"category"` // person, organization, location, technology, etc.
	Confidence float64 `json:"confidence"`
}

// SemanticKeyword represents an extracted keyword with metadata
type SemanticKeyword struct {
	Term      string  `json:"term"`
	Relevance float64 `json:"relevance"`
	Context   string  `json:"context"`
}

// SemanticTopic represents an extracted topic with metadata
type SemanticTopic struct {
	Name        string  `json:"name"`
	Score       float64 `json:"score"`
	Description string  `json:"description"`
}

// SemanticAnalysis contains all semantic data extracted in one call
type SemanticAnalysis struct {
	Entities       []SemanticEntity  `json:"entities"`
	Keywords       []SemanticKeyword `json:"keywords"`
	Topics         []SemanticTopic   `json:"topics"`
	Sentiment      string            `json:"sentiment"`
	SentimentScore float64           `json:"sentiment_score"`
	Tone           string            `json:"tone"`
}

type Article struct {
	ID             string            `json:"id"`
	URL            string            `json:"url"`
	Title          string            `json:"title"`
	Summary        string            `json:"summary"`
	Embedding      []float32         `json:"embedding"`
	Sentiment      string            `json:"sentiment"`
	SentimentScore float64           `json:"sentiment_score"`
	Tone           string            `json:"tone"`
	Entities       []SemanticEntity  `json:"entities"`
	Keywords       []SemanticKeyword `json:"keywords"`
	Topics         []SemanticTopic   `json:"topics"`
	URLHash        string            `json:"url_hash"` // SHA-256 hash of the URL for caching
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

// ChatCache represents a cached chat request/response
type ChatCache struct {
	ID           string      `json:"id"`
	RequestHash  string      `json:"request_hash"`
	RequestJSON  interface{} `json:"request_json"`
	ResponseJSON interface{} `json:"response_json"`
	CreatedAt    time.Time   `json:"created_at"`
	ExpiresAt    time.Time   `json:"expires_at"`
}

type ChatRequest struct {
	Query string `json:"query,omitempty"`
	Task  string `json:"task"` // summary, sentiment, compare, tone, search, more_positive, top_entities
}

type ChatResponse struct {
	Answer       string      `json:"answer"`
	Sources      []Source    `json:"sources"`
	Usage        Usage       `json:"usage"`
	Task         string      `json:"task"`
	ResponseType string      `json:"response_type"`
	Articles     []Article   `json:"articles,omitempty"` // For article list responses
	Data         interface{} `json:"data,omitempty"`     // For structured data responses
	Plan         *Plan       `json:"plan,omitempty"`     // Debug: LLM execution plan
}

type Source struct {
	ID    string `json:"id"`
	URL   string `json:"url"`
	Title string `json:"title"`
}

type Usage struct {
	Tokens int     `json:"tokens"`
	Cost   float64 `json:"cost"`
}

// Plan represents a command-based execution plan from LLM
type Plan struct {
	Command string                 `json:"command"`
	Args    map[string]interface{} `json:"args"`
}

const (
	// Response types
	ResponseText        = "text"         // Single text response
	ResponseArticleList = "article_list" // List of articles with URLs
	ResponseData        = "data"         // Structured data (entities, keywords, etc.)

	// Query types
	QuerySummary      = "summary"       // Single article summary
	QueryKeywords     = "keywords"      // Keywords/topics from articles
	QuerySentiment    = "sentiment"     // Sentiment from DB
	QueryCompare      = "compare"       // LLM comparison of summaries
	QueryTone         = "tone"          // LLM tone comparison
	QuerySearch       = "search"        // Article search by topic
	QueryVectorSearch = "vector_search" // Vector search using embeddings
	QueryMorePositive = "most_positive" // Most positive article
	QueryTopEntities  = "top_entities"  // Top entities across articles
	QueryUnknown      = "unknown"
)
