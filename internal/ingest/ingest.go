package ingest

import (
	"article-assistant/internal/domain"
	"article-assistant/internal/llm"
	"article-assistant/internal/repository"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type Service struct {
	Repo *repository.Repo
	LLM  llm.Client
}

func (s *Service) IngestURL(ctx context.Context, url string) error {
	// Calculate URL hash for caching
	urlHash := calculateURLHash(url)

	// Check if article already exists
	existingArticle, err := s.Repo.GetArticleByURL(ctx, url)
	if err != nil {
		return fmt.Errorf("failed to check existing article: %w", err)
	}

	// If article already exists, skip processing
	if existingArticle != nil {
		log.Printf("📄 Article already processed, skipping: %s", url)
		return nil
	}

	// Fetch content
	contentInfo, err := fetchHTMLWithHeaders(url)
	if err != nil {
		return fmt.Errorf("failed to fetch content: %w", err)
	}

	log.Printf("📄 Processing new article: %s", url)

	// Process the content
	text := StripHTMLBasic(contentInfo.HTML)

	sum, err := s.LLM.Summarize(ctx, text)
	if err != nil {
		return fmt.Errorf("failed to summarize: %w", err)
	}

	emb, err := s.LLM.Embed(ctx, sum)
	if err != nil {
		return fmt.Errorf("failed to embed: %w", err)
	}

	// Extract all semantic data in a single LLM call (faster and cheaper)
	semanticAnalysis, err := s.LLM.ExtractAllSemantics(ctx, sum)
	if err != nil {
		log.Printf("Failed to extract semantic data: %v", err)
		// Fallback to empty data
		semanticAnalysis = &domain.SemanticAnalysis{
			Entities:       []domain.SemanticEntity{},
			Keywords:       []domain.SemanticKeyword{},
			Topics:         []domain.SemanticTopic{},
			Sentiment:      "neutral",
			SentimentScore: 0.5,
		}
	}

	// Convert LLM semantic data to domain types
	entities := make([]domain.SemanticEntity, len(semanticAnalysis.Entities))
	for i, entity := range semanticAnalysis.Entities {
		entities[i] = domain.SemanticEntity{
			Name:       entity.Name,
			Category:   entity.Category,
			Confidence: entity.Confidence,
		}
	}

	keywords := make([]domain.SemanticKeyword, len(semanticAnalysis.Keywords))
	for i, keyword := range semanticAnalysis.Keywords {
		keywords[i] = domain.SemanticKeyword{
			Term:      keyword.Term,
			Relevance: keyword.Relevance,
			Context:   keyword.Context,
		}
	}

	topics := make([]domain.SemanticTopic, len(semanticAnalysis.Topics))
	for i, topic := range semanticAnalysis.Topics {
		topics[i] = domain.SemanticTopic{
			Name:        topic.Name,
			Score:       topic.Score,
			Description: topic.Description,
		}
	}

	// Create article with URL hash
	a := &domain.Article{
		ID:             uuid.New().String(),
		URL:            url,
		Title:          contentInfo.Title,
		Summary:        sum,
		Embedding:      emb,
		Entities:       entities,
		Keywords:       keywords,
		Topics:         topics,
		Sentiment:      semanticAnalysis.Sentiment,
		SentimentScore: semanticAnalysis.SentimentScore,
		URLHash:        urlHash,
	}

	return s.Repo.UpsertArticle(ctx, a)
}

func fetchHTML(url string) (body, title string, err error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "ArticleAssistant/1.0")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	html := string(b)
	t := ExtractBetween(html, "<title>", "</title>")
	return html, strings.TrimSpace(t), nil
}

func StripHTMLBasic(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	for _, tag := range []string{"script", "style"} {
		for {
			start := strings.Index(strings.ToLower(s), "<"+tag)
			if start == -1 {
				break
			}
			end := strings.Index(strings.ToLower(s[start:]), "</"+tag+">")
			if end == -1 {
				break
			}
			s = s[:start] + s[start+end+len(tag)+3:]
		}
	}
	out := ""
	skip := false
	for _, r := range s {
		if r == '<' {
			skip = true
			continue
		}
		if r == '>' {
			skip = false
			continue
		}
		if !skip {
			out += string(r)
		}
	}
	return strings.Join(strings.Fields(out), " ")
}

func ExtractBetween(s, a, b string) string {
	ai := strings.Index(strings.ToLower(s), strings.ToLower(a))
	if ai == -1 {
		return ""
	}
	ai += len(a)
	bi := strings.Index(strings.ToLower(s[ai:]), strings.ToLower(b))
	if bi == -1 {
		return ""
	}
	return s[ai : ai+bi]
}
