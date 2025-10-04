package article

import "time"

// Article represents a processed article with all extracted metadata
type Article struct {
	ID             string    `json:"id"`
	URL            string    `json:"url"`
	Title          string    `json:"title"`
	Summary        string    `json:"summary"`
	Content        string    `json:"content"`
	Excerpt        string    `json:"excerpt"` // Short excerpt for context
	Embedding      []float32 `json:"embedding"`
	Sentiment      string    `json:"sentiment"`
	SentimentScore float64   `json:"sentiment_score"`
	Tone           string    `json:"tone"`
	Entities       []string  `json:"entities"`
	Keywords       []string  `json:"keywords"`
	Topics         []string  `json:"topics"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
