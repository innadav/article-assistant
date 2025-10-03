package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"article-assistant/internal/domain"
)

type Repo struct{ DB *sql.DB }

func NewRepo(db *sql.DB) *Repo { return &Repo{DB: db} }

// ---------- Helpers ----------

// applyURLFilter adds url filtering if urls provided
func applyURLFilter(query string, urls []string, args []interface{}) (string, []interface{}) {
	if len(urls) == 0 {
		return query, args
	}
	placeholders := make([]string, len(urls))
	startIndex := len(args) + 1
	for i, u := range urls {
		placeholders[i] = fmt.Sprintf("$%d", startIndex+i)
		args = append(args, u)
	}
	query += fmt.Sprintf(" AND url IN (%s)", strings.Join(placeholders, ","))
	return query, args
}

// parseJSONFields parses entities/keywords/topics JSON
func parseJSONFields(a *domain.Article, entitiesJSON, keywordsJSON, topicsJSON []byte) {
	if len(entitiesJSON) > 0 {
		_ = json.Unmarshal(entitiesJSON, &a.Entities)
	}
	if len(keywordsJSON) > 0 {
		_ = json.Unmarshal(keywordsJSON, &a.Keywords)
	}
	if len(topicsJSON) > 0 {
		_ = json.Unmarshal(topicsJSON, &a.Topics)
	}
}

// ---------- Core Queries ----------

func (r *Repo) GetSummaryByID(ctx context.Context, id int, urls []string) (string, error) {
	q := "SELECT summary FROM articles WHERE id=$1"
	args := []interface{}{id}
	q, args = applyURLFilter(q, urls, args)

	var s string
	err := r.DB.QueryRowContext(ctx, q, args...).Scan(&s)
	return s, err
}

// GetMostPositiveByTopic returns the most positive article on a given topic
func (r *Repo) GetMostPositiveByTopic(ctx context.Context, topic string, urls []string) (*domain.Article, error) {
	q := `
	  SELECT id, url, title, summary, sentiment, sentiment_score, tone, entities, keywords, topics, created_at, updated_at
	  FROM articles
	  WHERE (
	    EXISTS (SELECT 1 FROM jsonb_array_elements(keywords) kw WHERE LOWER(kw->>'term') LIKE LOWER($1))
	    OR EXISTS (SELECT 1 FROM jsonb_array_elements(entities) e WHERE LOWER(e->>'name') LIKE LOWER($1))
	    OR EXISTS (SELECT 1 FROM jsonb_array_elements(topics) t WHERE LOWER(t->>'name') LIKE LOWER($1))
	  )`
	args := []interface{}{"%" + topic + "%"}
	q, args = applyURLFilter(q, urls, args)
	q += " ORDER BY sentiment_score DESC LIMIT 1"

	row := r.DB.QueryRowContext(ctx, q, args...)

	var a domain.Article
	var entitiesJSON, keywordsJSON, topicsJSON []byte
	if err := row.Scan(&a.ID, &a.URL, &a.Title, &a.Summary,
		&a.Sentiment, &a.SentimentScore, &a.Tone,
		&entitiesJSON, &keywordsJSON, &topicsJSON,
		&a.CreatedAt, &a.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	parseJSONFields(&a, entitiesJSON, keywordsJSON, topicsJSON)
	return &a, nil
}

// GetTopEntities returns most commonly discussed entities across all articles
func (r *Repo) GetTopEntities(ctx context.Context, limit int, urls []string) ([]domain.SemanticEntity, error) {
	q := `
	  SELECT elem->>'name' AS entity_name,
	         COUNT(*) AS count,
	         AVG((elem->>'confidence')::float) AS avg_confidence
	  FROM articles, jsonb_array_elements(entities) elem
	  WHERE entities IS NOT NULL`
	args := []interface{}{}
	q, args = applyURLFilter(q, urls, args)
	q += fmt.Sprintf(" GROUP BY elem->>'name' ORDER BY count DESC, avg_confidence DESC LIMIT $%d", len(args)+1)
	args = append(args, limit)

	rows, err := r.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.SemanticEntity
	for rows.Next() {
		var e domain.SemanticEntity
		var count int
		var avg float64
		if err := rows.Scan(&e.Name, &count, &avg); err != nil {
			return nil, err
		}
		e.Confidence = avg
		result = append(result, e)
	}
	return result, nil
}

// GetArticlesByVectorSearch performs semantic search using embeddings
func (r *Repo) GetArticlesByVectorSearch(ctx context.Context, queryEmbedding []float32, limit int, urls []string) ([]domain.Article, error) {
	embeddingStr := "[" + strings.Trim(strings.Join(strings.Fields(fmt.Sprint(queryEmbedding)), ","), "[]") + "]"

	q := `
	  SELECT id, url, title, summary, sentiment, sentiment_score, tone, entities, keywords, topics, created_at, updated_at,
	         1 - (embedding <=> $1::vector) AS similarity
	  FROM articles
	  WHERE embedding IS NOT NULL`
	args := []interface{}{embeddingStr}
	q, args = applyURLFilter(q, urls, args)
	q += fmt.Sprintf(" ORDER BY embedding <=> $1::vector LIMIT $%d", len(args)+1)
	args = append(args, limit)

	rows, err := r.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Article
	for rows.Next() {
		var a domain.Article
		var entitiesJSON, keywordsJSON, topicsJSON []byte
		var sim float64
		if err := rows.Scan(&a.ID, &a.URL, &a.Title, &a.Summary,
			&a.Sentiment, &a.SentimentScore, &a.Tone,
			&entitiesJSON, &keywordsJSON, &topicsJSON,
			&a.CreatedAt, &a.UpdatedAt, &sim); err != nil {
			return nil, err
		}
		parseJSONFields(&a, entitiesJSON, keywordsJSON, topicsJSON)
		out = append(out, a)
	}
	return out, nil
}

// GetArticlesByKeywordsOrEntities queries articles by keywords or entities
func (r *Repo) GetArticlesByKeywordsOrEntities(ctx context.Context, filter string, limit int) ([]domain.Article, error) {
	q := `
	  SELECT id, url, title, summary, sentiment, sentiment_score, tone, entities, keywords, topics, created_at, updated_at
	  FROM articles
	  WHERE 
	    EXISTS (SELECT 1 FROM jsonb_array_elements(keywords) kw WHERE LOWER(kw->>'term') LIKE LOWER($1))
	    OR EXISTS (SELECT 1 FROM jsonb_array_elements(entities) e WHERE LOWER(e->>'name') LIKE LOWER($1))
	    OR EXISTS (SELECT 1 FROM jsonb_array_elements(topics) t WHERE LOWER(t->>'name') LIKE LOWER($1))
	  ORDER BY created_at DESC
	  LIMIT $2`

	rows, err := r.DB.QueryContext(ctx, q, "%"+filter+"%", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var articles []domain.Article
	for rows.Next() {
		var a domain.Article
		var entitiesJSON, keywordsJSON, topicsJSON []byte
		err := rows.Scan(&a.ID, &a.URL, &a.Title, &a.Summary,
			&a.Sentiment, &a.SentimentScore, &a.Tone,
			&entitiesJSON, &keywordsJSON, &topicsJSON,
			&a.CreatedAt, &a.UpdatedAt)
		if err != nil {
			return nil, err
		}
		parseJSONFields(&a, entitiesJSON, keywordsJSON, topicsJSON)
		articles = append(articles, a)
	}

	return articles, nil
}

// GetKeywordsAndTopics retrieves and aggregates keywords and topics from articles
func (r *Repo) GetKeywordsAndTopics(ctx context.Context, urls []string, limit int) ([]domain.SemanticKeyword, []domain.SemanticTopic, error) {
	if len(urls) == 0 {
		return nil, nil, fmt.Errorf("no URLs provided")
	}

	placeholders := make([]string, len(urls))
	args := make([]interface{}, len(urls))
	for i, u := range urls {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = u
	}

	query := fmt.Sprintf(`
		SELECT keywords, topics
		FROM articles
		WHERE url IN (%s)`, strings.Join(placeholders, ","))

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	kwCount := make(map[string]int)
	tpCount := make(map[string]int)

	for rows.Next() {
		var kwJSON, tpJSON []byte
		if err := rows.Scan(&kwJSON, &tpJSON); err != nil {
			return nil, nil, err
		}

		var kws []domain.SemanticKeyword
		var tps []domain.SemanticTopic
		json.Unmarshal(kwJSON, &kws)
		json.Unmarshal(tpJSON, &tps)

		for _, k := range kws {
			kwCount[k.Term]++
		}
		for _, t := range tps {
			tpCount[t.Name]++
		}
	}

	// Convert maps to slices and sort by frequency
	var kwList []domain.SemanticKeyword
	for term, count := range kwCount {
		kwList = append(kwList, domain.SemanticKeyword{Term: term, Relevance: float64(count)})
	}
	sort.Slice(kwList, func(i, j int) bool { return kwList[i].Relevance > kwList[j].Relevance })
	if len(kwList) > limit {
		kwList = kwList[:limit]
	}

	var tpList []domain.SemanticTopic
	for name, count := range tpCount {
		tpList = append(tpList, domain.SemanticTopic{Name: name, Score: float64(count)})
	}
	sort.Slice(tpList, func(i, j int) bool { return tpList[i].Score > tpList[j].Score })
	if len(tpList) > limit {
		tpList = tpList[:limit]
	}

	return kwList, tpList, nil
}

// GetArticlesByURLs retrieves articles by their URLs
func (r *Repo) GetArticlesByURLs(ctx context.Context, urls []string) ([]domain.Article, error) {
	if len(urls) == 0 {
		return nil, fmt.Errorf("no URLs provided")
	}

	placeholders := make([]string, len(urls))
	args := make([]interface{}, len(urls))
	for i, u := range urls {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = u
	}

	query := fmt.Sprintf(`
		SELECT id, url, title, summary, sentiment, sentiment_score, tone, entities, keywords, topics, created_at, updated_at
		FROM articles
		WHERE url IN (%s)`, strings.Join(placeholders, ","))

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var articles []domain.Article
	for rows.Next() {
		var a domain.Article
		var entitiesJSON, keywordsJSON, topicsJSON []byte
		err := rows.Scan(&a.ID, &a.URL, &a.Title, &a.Summary,
			&a.Sentiment, &a.SentimentScore, &a.Tone,
			&entitiesJSON, &keywordsJSON, &topicsJSON,
			&a.CreatedAt, &a.UpdatedAt)
		if err != nil {
			return nil, err
		}
		parseJSONFields(&a, entitiesJSON, keywordsJSON, topicsJSON)
		articles = append(articles, a)
	}

	return articles, nil
}

// ---------- Upsert ----------
func (r *Repo) UpsertArticle(ctx context.Context, article *domain.Article) error {
	query := `INSERT INTO articles (id, url, title, summary, embedding, sentiment, sentiment_score, tone, entities, keywords, topics, created_at, updated_at)
		  VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		  ON CONFLICT (url) DO UPDATE SET 
		    title=EXCLUDED.title, summary=EXCLUDED.summary, embedding=EXCLUDED.embedding,
		    sentiment=EXCLUDED.sentiment, sentiment_score=EXCLUDED.sentiment_score,
		    tone=EXCLUDED.tone, entities=EXCLUDED.entities, keywords=EXCLUDED.keywords,
		    topics=EXCLUDED.topics, updated_at=EXCLUDED.updated_at`

	now := time.Now()
	article.CreatedAt, article.UpdatedAt = now, now

	var embeddingStr string
	if len(article.Embedding) > 0 {
		parts := make([]string, len(article.Embedding))
		for i, v := range article.Embedding {
			parts[i] = fmt.Sprintf("%f", v)
		}
		embeddingStr = "[" + strings.Join(parts, ",") + "]"
	} else {
		embeddingStr = "[]"
	}

	entitiesJSON, err := json.Marshal(article.Entities)
	if err != nil {
		return fmt.Errorf("failed to marshal entities: %w", err)
	}
	keywordsJSON, err := json.Marshal(article.Keywords)
	if err != nil {
		return fmt.Errorf("failed to marshal keywords: %w", err)
	}
	topicsJSON, err := json.Marshal(article.Topics)
	if err != nil {
		return fmt.Errorf("failed to marshal topics: %w", err)
	}

	_, err = r.DB.ExecContext(ctx, query,
		article.ID, article.URL, article.Title, article.Summary,
		embeddingStr, article.Sentiment, article.SentimentScore, article.Tone,
		entitiesJSON, keywordsJSON, topicsJSON,
		article.CreatedAt, article.UpdatedAt,
	)
	return err
}
