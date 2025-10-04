package article

import (
	"sync"
	"time"
)

// Article holds the processed data from a URL.
type Article struct {
	URL         string    `json:"url"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Excerpt     string    `json:"excerpt"`
	Topics      []string  `json:"topics"`
	Sentiment   string    `json:"sentiment"`
	Summary     string    `json:"summary"`
	ProcessedAt time.Time `json:"processed_at"`
}

// Store is an in-memory, concurrent-safe storage for articles.
type Store struct {
	articles sync.Map
}

func NewStore() *Store {
	return &Store{}
}

func (s *Store) Get(url string) (*Article, bool) {
	val, ok := s.articles.Load(url)
	if !ok {
		return nil, false
	}
	return val.(*Article), true
}

func (s *Store) Set(url string, article *Article) {
	s.articles.Store(url, article)
}

func (s *Store) GetAll() []*Article {
	var articles []*Article
	s.articles.Range(func(key, value interface{}) bool {
		articles = append(articles, value.(*Article))
		return true
	})
	return articles
}
