package repository

import (
	"context"
	"database/sql"
	"fmt"

	"article-assistant/internal/article"

	"github.com/lib/pq" // PostgreSQL driver
)

// ArticleRepository handles database operations for articles.
type ArticleRepository struct {
	db *sql.DB
}

// NewPostgresRepo creates a new repository.
func NewPostgresRepo(dbURL string) (*ArticleRepository, error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}
	// Ping the database to ensure the connection is live
	if err = db.Ping(); err != nil {
		db.Close() // Close the connection if ping fails
		return nil, fmt.Errorf("failed to connect to the database: %w", err)
	}
	return &ArticleRepository{db: db}, nil
}

// Save stores an article in the database.
func (r *ArticleRepository) Save(ctx context.Context, art *article.Article) error {
	query := `
		INSERT INTO articles (url, title, content, excerpt, summary, sentiment, topics, processed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (url) DO UPDATE SET
			title = EXCLUDED.title,
			summary = EXCLUDED.summary,
			sentiment = EXCLUDED.sentiment,
			topics = EXCLUDED.topics;
	`
	_, err := r.db.ExecContext(ctx, query,
		art.URL, art.Title, art.Content, art.Excerpt,
		art.Summary, art.Sentiment, pq.Array(art.Topics), art.ProcessedAt,
	)
	return err
}

// FindByURL retrieves an article by its URL.
func (r *ArticleRepository) FindByURL(ctx context.Context, url string) (*article.Article, error) {
	var art article.Article
	query := `SELECT url, title, content, excerpt, summary, sentiment, topics, processed_at FROM articles WHERE url = $1`
	err := r.db.QueryRowContext(ctx, query, url).Scan(
		&art.URL,
		&art.Title,
		&art.Content,
		&art.Excerpt,
		&art.Summary,
		&art.Sentiment,
		pq.Array(&art.Topics),
		&art.ProcessedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil // Article not found
	}
	if err != nil {
		return nil, fmt.Errorf("error finding article by URL: %w", err)
	}
	return &art, nil
}

// FindAll retrieves all articles.
func (r *ArticleRepository) FindAll(ctx context.Context) ([]*article.Article, error) {
	query := `SELECT url, title, content, excerpt, summary, sentiment, topics, processed_at FROM articles`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error finding all articles: %w", err)
	}
	defer rows.Close()

	var articles []*article.Article
	for rows.Next() {
		var art article.Article
		if err := rows.Scan(
			&art.URL,
			&art.Title,
			&art.Content,
			&art.Excerpt,
			&art.Summary,
			&art.Sentiment,
			pq.Array(&art.Topics),
			&art.ProcessedAt,
		); err != nil {
			return nil, fmt.Errorf("error scanning article: %w", err)
		}
		articles = append(articles, &art)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error with rows: %w", err)
	}

	return articles, nil
}
